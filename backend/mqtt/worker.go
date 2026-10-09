package mqtt

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
)

type State struct {
	mu              sync.RWMutex
	Connected       bool      `json:"connected"`
	CandidateID     string    `json:"candidate_id"`
	LastChallengeID string    `json:"last_challenge_id"`
	LastStatus      string    `json:"last_status"`
	LastError       string    `json:"last_error"`
	LastSeenAt      time.Time `json:"last_seen_at"`
}

var globalState = &State{}

func GetState() map[string]interface{} {
	globalState.mu.RLock()
	defer globalState.mu.RUnlock()
	return map[string]interface{}{
		"connected":         globalState.Connected,
		"candidate_id":      globalState.CandidateID,
		"last_challenge_id": globalState.LastChallengeID,
		"last_status":       globalState.LastStatus,
		"last_error":        globalState.LastError,
		"last_seen_at":      globalState.LastSeenAt,
	}
}

type EventInput struct {
	SourceID      string  `json:"source_id"`
	EventID       string  `json:"event_id"`
	Type          string  `json:"type"`
	Quantity      *int    `json:"quantity"`
	TargetEventID *string `json:"target_event_id"`
	EventTime     string  `json:"event_time"`
}

type EventResult struct {
	EventID string `json:"event_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type Summary struct {
	NetTotal        int `json:"net_total"`
	ProcessedEvents int `json:"processed_events"`
	PendingAck      int `json:"pending_ack"`
	Unresolved      int `json:"unresolved"`
	Duplicates      int `json:"duplicates"`
	Conflicts       int `json:"conflicts"`
}

type Challenge struct {
	ProtocolVersion string       `json:"protocol_version"`
	CandidateID     string       `json:"candidate_id"`
	ChallengeID     string       `json:"challenge_id"`
	Command         string       `json:"command"`
	SentAt          time.Time    `json:"sent_at"`
	ExpiresAt       time.Time    `json:"expires_at"`
	Events          []EventInput `json:"events"`
}

type EventProcessor interface {
	ProcessEvents(events []EventInput, challengeID *string) ([]EventResult, error)
	GetSummary(sourceID string) (Summary, error)
}

type ChallengeStore interface {
	IsDuplicate(challengeID string) bool
	Store(challengeID, candidateID, command string, sentAt, expiresAt time.Time, payload []byte) error
	Complete(challengeID string, response []byte)
	Fail(challengeID, reason string)
}

type Worker struct {
	client         paho.Client
	candidateID    string
	processor      EventProcessor
	challengeStore ChallengeStore
}

func NewWorker(broker string, port int, candidateID string, processor EventProcessor, store ChallengeStore) *Worker {
	w := &Worker{candidateID: candidateID, processor: processor, challengeStore: store}

	suffix := fmt.Sprintf("%04x", rand.Intn(0xFFFF))
	clientID := fmt.Sprintf("fse01-%s-%s", candidateID, suffix)

	willTopic := fmt.Sprintf("fse-01/%s/status", candidateID)
	willPayload := `{"status":"OFFLINE","candidate_id":"` + candidateID + `"}`

	opts := paho.NewClientOptions().
		AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port)).
		SetClientID(clientID).
		SetKeepAlive(30 * time.Second).
		SetWill(willTopic, willPayload, 1, false).
		SetAutoReconnect(true).
		SetOnConnectHandler(w.onConnect).
		SetConnectionLostHandler(w.onDisconnect)

	w.client = paho.NewClient(opts)
	return w
}

func (w *Worker) Start() {
	globalState.mu.Lock()
	globalState.CandidateID = w.candidateID
	globalState.mu.Unlock()

	go func() {
		for {
			if token := w.client.Connect(); token.Wait() && token.Error() != nil {
				globalState.mu.Lock()
				globalState.LastError = token.Error().Error()
				globalState.Connected = false
				globalState.mu.Unlock()
				fmt.Println("MQTT connect error:", token.Error())
				time.Sleep(5 * time.Second)
				continue
			}
			break
		}
	}()
}

func (w *Worker) onConnect(c paho.Client) {
	globalState.mu.Lock()
	globalState.Connected = true
	globalState.LastError = ""
	globalState.mu.Unlock()
	fmt.Println("✅ MQTT connected")

	statusTopic := fmt.Sprintf("fse-01/%s/status", w.candidateID)
	challengeTopic := fmt.Sprintf("fse-01/%s/challenge", w.candidateID)

	w.publish(statusTopic, map[string]string{"status": "ONLINE", "candidate_id": w.candidateID})
	c.Subscribe(challengeTopic, 1, w.handleChallenge)
	go w.heartbeat(statusTopic)
}

func (w *Worker) onDisconnect(_ paho.Client, err error) {
	globalState.mu.Lock()
	globalState.Connected = false
	if err != nil {
		globalState.LastError = err.Error()
	}
	globalState.mu.Unlock()
	fmt.Println("MQTT disconnected:", err)
}

func (w *Worker) heartbeat(statusTopic string) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		if !globalState.Connected {
			return
		}
		w.publish(statusTopic, map[string]string{"status": "HEARTBEAT", "candidate_id": w.candidateID})
	}
}

func (w *Worker) handleChallenge(_ paho.Client, msg paho.Message) {
	var ch Challenge
	if err := json.Unmarshal(msg.Payload(), &ch); err != nil {
		w.publishFailure("", "VALIDATION_ERROR", "invalid JSON payload")
		return
	}

	globalState.mu.Lock()
	globalState.LastChallengeID = ch.ChallengeID
	globalState.LastSeenAt = time.Now()
	globalState.mu.Unlock()

	// Validate protocol
	if ch.ProtocolVersion != "1.0" {
		w.publishFailure(ch.ChallengeID, "UNSUPPORTED_PROTOCOL", "unsupported protocol version")
		return
	}
	if ch.CandidateID != w.candidateID {
		w.publishFailure(ch.ChallengeID, "CANDIDATE_MISMATCH", "candidate_id does not match")
		return
	}
	if ch.Command != "PROCESS_EVENTS" {
		w.publishFailure(ch.ChallengeID, "UNSUPPORTED_PROTOCOL", "unsupported command")
		return
	}
	if time.Now().After(ch.ExpiresAt) {
		w.publishFailure(ch.ChallengeID, "CHALLENGE_EXPIRED", "challenge has expired")
		return
	}

	// Deduplicate challenge
	if w.challengeStore.IsDuplicate(ch.ChallengeID) {
		w.publishFailure(ch.ChallengeID, "CHALLENGE_CONFLICT", "challenge already processed")
		globalState.mu.Lock()
		globalState.LastStatus = "CHALLENGE_CONFLICT"
		globalState.mu.Unlock()
		return
	}

	// Store challenge before processing
	rawPayload, _ := json.Marshal(ch.Events)
	if err := w.challengeStore.Store(ch.ChallengeID, ch.CandidateID, ch.Command, ch.SentAt, ch.ExpiresAt, rawPayload); err != nil {
		w.publishFailure(ch.ChallengeID, "INTERNAL_ERROR", "failed to store challenge")
		return
	}

	cidStr := ch.ChallengeID
	results, err := w.processor.ProcessEvents(ch.Events, &cidStr)
	if err != nil {
		w.challengeStore.Fail(ch.ChallengeID, err.Error())
		w.publishFailure(ch.ChallengeID, "INTERNAL_ERROR", "processing failed")
		return
	}

	stateSum, _ := w.processor.GetSummary("")

	responseTopic := fmt.Sprintf("fse-01/%s/response", w.candidateID)
	resp := map[string]interface{}{
		"protocol_version": "1.0",
		"challenge_id":     ch.ChallengeID,
		"candidate_id":     w.candidateID,
		"status":           "COMPLETED",
		"processed_at":     time.Now().UTC().Format(time.RFC3339),
		"results":          results,
		"state":            stateSum,
	}
	respBytes, _ := json.Marshal(resp)
	w.challengeStore.Complete(ch.ChallengeID, respBytes)
	w.publish(responseTopic, resp)

	globalState.mu.Lock()
	globalState.LastStatus = "COMPLETED"
	globalState.mu.Unlock()
}

func (w *Worker) publishFailure(challengeID, code, reason string) {
	responseTopic := fmt.Sprintf("fse-01/%s/response", w.candidateID)
	w.publish(responseTopic, map[string]string{
		"protocol_version": "1.0",
		"challenge_id":     challengeID,
		"candidate_id":     w.candidateID,
		"status":           "FAILED",
		"error_code":       code,
		"reason":           reason,
	})
	globalState.mu.Lock()
	globalState.LastStatus = "FAILED: " + code
	globalState.LastError = reason
	globalState.mu.Unlock()
}

func (w *Worker) publish(topic string, payload interface{}) {
	data, _ := json.Marshal(payload)
	w.client.Publish(topic, 1, false, data)
}

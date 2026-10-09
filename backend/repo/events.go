package repo

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

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

type EventsRepo struct {
	db *sqlx.DB
}

func NewEventsRepo(db *sqlx.DB) *EventsRepo {
	return &EventsRepo{db: db}
}

func (r *EventsRepo) ProcessEvents(events []EventInput, challengeID *string) ([]EventResult, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	results := make([]EventResult, 0, len(events))
	for _, ev := range events {
		res := r.processSingle(tx, ev, challengeID)
		results = append(results, res)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return results, nil
}

func (r *EventsRepo) processSingle(tx *sql.Tx, ev EventInput, challengeID *string) EventResult {
	rawJSON, _ := json.Marshal(ev)

	// Validate required fields
	if ev.SourceID == "" || ev.EventID == "" {
		reason := "source_id and event_id are required"
		r.recordAttempt(tx, ev, "REJECTED", reason, rawJSON, challengeID)
		return EventResult{EventID: ev.EventID, Status: "REJECTED", Message: reason}
	}
	if ev.Type != "COUNT" && ev.Type != "VOID" {
		reason := "type must be COUNT or VOID"
		r.recordAttempt(tx, ev, "REJECTED", reason, rawJSON, challengeID)
		return EventResult{EventID: ev.EventID, Status: "REJECTED", Message: reason}
	}
	if ev.EventTime == "" {
		reason := "event_time is required"
		r.recordAttempt(tx, ev, "REJECTED", reason, rawJSON, challengeID)
		return EventResult{EventID: ev.EventID, Status: "REJECTED", Message: reason}
	}
	if ev.Type == "COUNT" {
		if ev.Quantity == nil || *ev.Quantity <= 0 {
			reason := "quantity must be a positive integer for COUNT"
			r.recordAttempt(tx, ev, "REJECTED", reason, rawJSON, challengeID)
			return EventResult{EventID: ev.EventID, Status: "REJECTED", Message: reason}
		}
	}
	if ev.Type == "VOID" {
		if ev.TargetEventID == nil || *ev.TargetEventID == "" {
			reason := "target_event_id is required for VOID"
			r.recordAttempt(tx, ev, "REJECTED", reason, rawJSON, challengeID)
			return EventResult{EventID: ev.EventID, Status: "REJECTED", Message: reason}
		}
	}

	// Parse event_time
	eventTime, err := time.Parse(time.RFC3339, ev.EventTime)
	if err != nil {
		reason := "event_time must be ISO 8601 with timezone"
		r.recordAttempt(tx, ev, "REJECTED", reason, rawJSON, challengeID)
		return EventResult{EventID: ev.EventID, Status: "REJECTED", Message: reason}
	}

	// Upsert source
	tx.Exec(`INSERT INTO production_sources (source_id) VALUES ($1) ON CONFLICT DO NOTHING`, ev.SourceID)

	// Check duplicate: already in production_events
	var existingID int64
	err = tx.QueryRow(`SELECT id FROM production_events WHERE source_id=$1 AND event_id=$2`, ev.SourceID, ev.EventID).Scan(&existingID)
	if err == nil {
		// Already exists → DUPLICATE
		r.recordAttempt(tx, ev, "DUPLICATE", "event already processed", rawJSON, challengeID)
		return EventResult{EventID: ev.EventID, Status: "DUPLICATE", Message: "event already processed"}
	}

	if ev.Type == "COUNT" {
		return r.processCount(tx, ev, eventTime, rawJSON, challengeID)
	}
	return r.processVoid(tx, ev, eventTime, rawJSON, challengeID)
}

func (r *EventsRepo) processCount(tx *sql.Tx, ev EventInput, eventTime time.Time, rawJSON []byte, challengeID *string) EventResult {
	// Insert the COUNT event as ACCEPTED
	var newID int64
	err := tx.QueryRow(
		`INSERT INTO production_events (source_id, event_id, type, quantity, event_time, status)
		 VALUES ($1, $2, 'COUNT', $3, $4, 'ACCEPTED') RETURNING id`,
		ev.SourceID, ev.EventID, *ev.Quantity, eventTime,
	).Scan(&newID)
	if err != nil {
		r.recordAttempt(tx, ev, "REJECTED", fmt.Sprintf("db error: %v", err), rawJSON, challengeID)
		return EventResult{EventID: ev.EventID, Status: "REJECTED", Message: "database error"}
	}

	// Collect PENDING_REFERENCE VOIDs waiting for this COUNT — read all before any UPDATE
	var pendingVoidIDs []int64
	rows, qErr := tx.Query(
		`SELECT id FROM production_events
		 WHERE type='VOID' AND status='PENDING_REFERENCE' AND target_event_id=$1 AND source_id=$2
		 ORDER BY received_at ASC`,
		ev.EventID, ev.SourceID,
	)
	if qErr == nil {
		for rows.Next() {
			var voidID int64
			rows.Scan(&voidID)
			pendingVoidIDs = append(pendingVoidIDs, voidID)
		}
		rows.Close()
	}
	// Now apply: first VOID wins, rest become CONFLICT
	for i, voidID := range pendingVoidIDs {
		if i == 0 {
			tx.Exec(`UPDATE production_events SET status='VOIDED' WHERE id=$1`, newID)
			tx.Exec(`UPDATE production_events SET status='ACCEPTED' WHERE id=$1`, voidID)
		} else {
			tx.Exec(`UPDATE production_events SET status='CONFLICT' WHERE id=$1`, voidID)
		}
	}

	r.recordAttempt(tx, ev, "ACCEPTED", "", rawJSON, challengeID)
	return EventResult{EventID: ev.EventID, Status: "ACCEPTED", Message: "event processed"}
}

func (r *EventsRepo) processVoid(tx *sql.Tx, ev EventInput, eventTime time.Time, rawJSON []byte, challengeID *string) EventResult {
	// Find target COUNT with same source_id
	var countID int64
	var countStatus string
	err := tx.QueryRow(
		`SELECT id, status FROM production_events WHERE source_id=$1 AND event_id=$2 AND type='COUNT'`,
		ev.SourceID, *ev.TargetEventID,
	).Scan(&countID, &countStatus)

	if err == sql.ErrNoRows {
		// Target COUNT not found → PENDING_REFERENCE
		_, insertErr := tx.Exec(
			`INSERT INTO production_events (source_id, event_id, type, target_event_id, event_time, status)
			 VALUES ($1, $2, 'VOID', $3, $4, 'PENDING_REFERENCE')`,
			ev.SourceID, ev.EventID, *ev.TargetEventID, eventTime,
		)
		if insertErr != nil {
			r.recordAttempt(tx, ev, "REJECTED", "database error", rawJSON, challengeID)
			return EventResult{EventID: ev.EventID, Status: "REJECTED", Message: "database error"}
		}
		r.recordAttempt(tx, ev, "PENDING_REFERENCE", "", rawJSON, challengeID)
		return EventResult{EventID: ev.EventID, Status: "PENDING_REFERENCE", Message: "waiting for target COUNT"}
	}

	if countStatus == "VOIDED" {
		// COUNT already reversed → CONFLICT
		r.recordAttempt(tx, ev, "CONFLICT", "target COUNT already reversed", rawJSON, challengeID)
		return EventResult{EventID: ev.EventID, Status: "CONFLICT", Message: "target COUNT already reversed"}
	}

	// COUNT found and ACCEPTED → apply VOID
	tx.Exec(`UPDATE production_events SET status='VOIDED' WHERE id=$1`, countID)
	tx.Exec(
		`INSERT INTO production_events (source_id, event_id, type, target_event_id, event_time, status)
		 VALUES ($1, $2, 'VOID', $3, $4, 'ACCEPTED')`,
		ev.SourceID, ev.EventID, *ev.TargetEventID, eventTime,
	)

	r.recordAttempt(tx, ev, "ACCEPTED", "", rawJSON, challengeID)
	return EventResult{EventID: ev.EventID, Status: "ACCEPTED", Message: "COUNT reversed"}
}

func (r *EventsRepo) recordAttempt(tx *sql.Tx, ev EventInput, status, reason string, rawJSON []byte, challengeID *string) {
	tx.Exec(
		`INSERT INTO submission_attempts
		 (source_id, event_id, type, quantity, target_event_id, event_time, status, reason, raw_payload, challenge_id)
		 VALUES ($1,$2,$3,$4,$5,$6::timestamptz,$7,$8,$9,$10)`,
		ev.SourceID, ev.EventID, ev.Type, ev.Quantity, ev.TargetEventID, nullTime(ev.EventTime),
		status, nullStr(reason), rawJSON, challengeID,
	)
}

func nullStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func nullTime(s string) interface{} {
	if s == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil
	}
	return t
}

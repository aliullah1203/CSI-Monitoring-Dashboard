package cmd

import (
	"fmt"
	"os"

	"backend/config"
	"backend/infra"
	"backend/mqtt"
	"backend/repo"
	"backend/rest"
	ackH "backend/rest/handlers/ack"
	eventsH "backend/rest/handlers/events"
	stateH "backend/rest/handlers/state"
	"backend/rest/middlewares"
)

// mqttAdapter bridges repo types to mqtt.EventProcessor interface
type mqttAdapter struct {
	eventsRepo *repo.EventsRepo
	stateRepo  *repo.StateRepo
}

func (a *mqttAdapter) ProcessEvents(inputs []mqtt.EventInput, challengeID *string) ([]mqtt.EventResult, error) {
	// Convert mqtt.EventInput → repo.EventInput
	repoInputs := make([]repo.EventInput, len(inputs))
	for i, in := range inputs {
		repoInputs[i] = repo.EventInput{
			SourceID:      in.SourceID,
			EventID:       in.EventID,
			Type:          in.Type,
			Quantity:      in.Quantity,
			TargetEventID: in.TargetEventID,
			EventTime:     in.EventTime,
		}
	}
	repoResults, err := a.eventsRepo.ProcessEvents(repoInputs, challengeID)
	if err != nil {
		return nil, err
	}
	// Convert repo.EventResult → mqtt.EventResult
	results := make([]mqtt.EventResult, len(repoResults))
	for i, r := range repoResults {
		results[i] = mqtt.EventResult{
			EventID: r.EventID,
			Status:  r.Status,
			Message: r.Message,
		}
	}
	return results, nil
}

func (a *mqttAdapter) GetSummary(sourceID string) (mqtt.Summary, error) {
	s, err := a.stateRepo.GetSummary(sourceID)
	if err != nil {
		return mqtt.Summary{}, err
	}
	return mqtt.Summary{
		NetTotal:        s.NetTotal,
		ProcessedEvents: s.ProcessedEvents,
		PendingAck:      s.PendingAck,
		Unresolved:      s.Unresolved,
		Duplicates:      s.Duplicates,
		Conflicts:       s.Conflicts,
	}, nil
}

func Serve() {
	cnf := config.GetConfig()

	dbCon, err := infra.NewConnection(cnf.ConnectionString)
	if err != nil {
		fmt.Println("Database Connection Error:", err)
		os.Exit(1)
	}

	if err := infra.MigrateDB(dbCon, "./migrations/migrations"); err != nil {
		fmt.Println("Migration Error:", err)
		os.Exit(1)
	}

	mw := middlewares.NewMiddleware(cnf)

	eventsRepo := repo.NewEventsRepo(dbCon)
	ackRepo := repo.NewAckRepo(dbCon)
	stateRepo := repo.NewStateRepo(dbCon)

	challengeRepo := repo.NewChallengeRepo(dbCon)
	adapter := &mqttAdapter{eventsRepo: eventsRepo, stateRepo: stateRepo}
	mqttWorker := mqtt.NewWorker(cnf.MQTTBroker, cnf.MQTTPort, cnf.CandidateID, adapter, challengeRepo)
	mqttWorker.Start()

	eventsHandler := eventsH.NewHandler(eventsRepo, mw)
	ackHandler := ackH.NewHandler(ackRepo, mw)
	stateHandler := stateH.NewHandler(stateRepo, mw)

	server := rest.NewServer(cnf, eventsHandler, ackHandler, stateHandler)
	server.Start()
}

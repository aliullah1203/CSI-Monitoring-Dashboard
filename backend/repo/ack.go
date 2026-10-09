package repo

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
)

type AckRepo struct {
	db *sqlx.DB
}

func NewAckRepo(db *sqlx.DB) *AckRepo {
	return &AckRepo{db: db}
}

type AckResult struct {
	EventID string `json:"event_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

func (r *AckRepo) AckEvents(eventIDs []string) ([]AckResult, error) {
	results := make([]AckResult, 0, len(eventIDs))
	for _, id := range eventIDs {
		res := r.ackOne(id)
		results = append(results, res)
	}
	return results, nil
}

func (r *AckRepo) ackOne(eventID string) AckResult {
	var status string
	var ackedAt *time.Time
	err := r.db.QueryRow(
		`SELECT status, acknowledged_at FROM production_events WHERE event_id=$1 LIMIT 1`,
		eventID,
	).Scan(&status, &ackedAt)

	if err == sql.ErrNoRows {
		return AckResult{EventID: eventID, Status: "NOT_FOUND", Message: "no event with this ID"}
	}
	if err != nil {
		return AckResult{EventID: eventID, Status: "NOT_FOUND", Message: "database error"}
	}
	if ackedAt != nil {
		return AckResult{EventID: eventID, Status: "ALREADY_ACKED", Message: "already acknowledged"}
	}
	if status != "ACCEPTED" {
		return AckResult{EventID: eventID, Status: "NOT_READY", Message: "event is not in accepted state"}
	}

	now := time.Now()
	_, err = r.db.Exec(
		`UPDATE production_events SET acknowledged_at=$1 WHERE event_id=$2 AND acknowledged_at IS NULL`,
		now, eventID,
	)
	if err != nil {
		return AckResult{EventID: eventID, Status: "NOT_READY", Message: "failed to acknowledge"}
	}
	return AckResult{EventID: eventID, Status: "ACKED", Message: "acknowledged"}
}

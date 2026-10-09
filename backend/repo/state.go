package repo

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type StateRepo struct {
	db *sqlx.DB
}

func NewStateRepo(db *sqlx.DB) *StateRepo {
	return &StateRepo{db: db}
}

type Summary struct {
	NetTotal            int `json:"net_total" db:"net_total"`
	ProcessedEvents     int `json:"processed_events" db:"processed_events"`
	PendingAck          int `json:"pending_ack" db:"pending_ack"`
	Unresolved          int `json:"unresolved" db:"unresolved"`
	Duplicates          int `json:"duplicates" db:"duplicates"`
	Conflicts           int `json:"conflicts" db:"conflicts"`
	RejectedSubmissions int `json:"rejected_submissions" db:"rejected_submissions"`
}

type PendingEvent struct {
	SourceID  string     `json:"source_id" db:"source_id"`
	EventID   string     `json:"event_id" db:"event_id"`
	Type      string     `json:"type" db:"type"`
	Quantity  *int       `json:"quantity" db:"quantity"`
	EventTime time.Time  `json:"event_time" db:"event_time"`
	Status    string     `json:"status" db:"status"`
}

type ExceptionEvent struct {
	SourceID  string    `json:"source_id" db:"source_id"`
	EventID   string    `json:"event_id" db:"event_id"`
	Type      *string   `json:"type" db:"type"`
	Status    string    `json:"status" db:"status"`
	Reason    *string   `json:"reason" db:"reason"`
	ReceivedAt time.Time `json:"received_at" db:"received_at"`
}

func (r *StateRepo) GetSummary(sourceID string) (Summary, error) {
	var s Summary
	query := `
		SELECT
			COALESCE(SUM(CASE WHEN pe.type='COUNT' AND pe.status='ACCEPTED' THEN pe.quantity ELSE 0 END), 0) AS net_total,
			COUNT(CASE WHEN pe.status='ACCEPTED' THEN 1 END) AS processed_events,
			COUNT(CASE WHEN pe.status='ACCEPTED' AND pe.acknowledged_at IS NULL THEN 1 END) AS pending_ack,
			COUNT(CASE WHEN pe.status='PENDING_REFERENCE' THEN 1 END) AS unresolved,
			(SELECT COUNT(*) FROM submission_attempts WHERE status='DUPLICATE' AND ($1='' OR source_id=$1)) AS duplicates,
			(SELECT COUNT(*) FROM submission_attempts WHERE status='CONFLICT' AND ($1='' OR source_id=$1)) AS conflicts,
			(SELECT COUNT(*) FROM submission_attempts WHERE status='REJECTED' AND ($1='' OR source_id=$1)) AS rejected_submissions
		FROM production_events pe
		WHERE ($1='' OR pe.source_id=$1)
	`
	err := r.db.Get(&s, query, sourceID)
	return s, err
}

func (r *StateRepo) GetPending(sourceID string) ([]PendingEvent, error) {
	var events []PendingEvent
	query := `
		SELECT source_id, event_id, type, quantity, event_time, status
		FROM production_events
		WHERE status='ACCEPTED' AND acknowledged_at IS NULL
		AND ($1='' OR source_id=$1)
		ORDER BY received_at ASC
	`
	err := r.db.Select(&events, query, sourceID)
	if events == nil {
		events = []PendingEvent{}
	}
	return events, err
}

func (r *StateRepo) GetExceptions(sourceID string) ([]ExceptionEvent, error) {
	var events []ExceptionEvent
	query := `
		SELECT source_id, event_id, type, status, reason, received_at
		FROM submission_attempts
		WHERE status IN ('DUPLICATE','CONFLICT','REJECTED')
		AND ($1='' OR source_id=$1)
		ORDER BY received_at DESC
		LIMIT 100
	`
	err := r.db.Select(&events, query, sourceID)
	if events == nil {
		events = []ExceptionEvent{}
	}
	return events, err
}

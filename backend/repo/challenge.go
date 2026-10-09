package repo

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
)

type ChallengeRepo struct {
	db *sqlx.DB
}

func NewChallengeRepo(db *sqlx.DB) *ChallengeRepo {
	return &ChallengeRepo{db: db}
}

// IsDuplicate returns true if challenge_id already exists
func (r *ChallengeRepo) IsDuplicate(challengeID string) bool {
	var id int64
	err := r.db.QueryRow(`SELECT id FROM mqtt_challenges WHERE challenge_id=$1`, challengeID).Scan(&id)
	return err == nil
}

func (r *ChallengeRepo) Store(challengeID, candidateID, command string, sentAt, expiresAt time.Time, payload []byte) error {
	_, err := r.db.Exec(
		`INSERT INTO mqtt_challenges (challenge_id, candidate_id, command, sent_at, expires_at, events_payload, status)
		 VALUES ($1,$2,$3,$4,$5,$6,'PENDING')`,
		challengeID, candidateID, command, sentAt, expiresAt, payload,
	)
	return err
}

func (r *ChallengeRepo) Complete(challengeID string, response []byte) {
	r.db.Exec(
		`UPDATE mqtt_challenges SET status='COMPLETED', response_payload=$1, processed_at=$2 WHERE challenge_id=$3`,
		response, time.Now(), challengeID,
	)
}

func (r *ChallengeRepo) Fail(challengeID, reason string) {
	r.db.Exec(
		`UPDATE mqtt_challenges SET status='FAILED', processed_at=$1 WHERE challenge_id=$2`,
		time.Now(), challengeID,
	)
	_ = reason
}

func (r *ChallengeRepo) MarkDuplicate(challengeID string) {
	r.db.Exec(`UPDATE mqtt_challenges SET status='DUPLICATE' WHERE challenge_id=$1`, challengeID)
	_ = sql.ErrNoRows // suppress unused import
}

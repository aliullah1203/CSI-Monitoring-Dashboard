-- +migrate Up

CREATE TABLE IF NOT EXISTS production_sources (
    source_id TEXT PRIMARY KEY,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS production_events (
    id BIGSERIAL PRIMARY KEY,
    source_id TEXT NOT NULL,
    event_id TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('COUNT', 'VOID')),
    quantity INTEGER,
    target_event_id TEXT,
    event_time TIMESTAMPTZ NOT NULL,
    status TEXT NOT NULL DEFAULT 'ACCEPTED' CHECK (status IN ('ACCEPTED', 'PENDING_REFERENCE', 'VOIDED', 'CONFLICT')),
    acknowledged_at TIMESTAMPTZ,
    received_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (source_id, event_id)
);

CREATE TABLE IF NOT EXISTS submission_attempts (
    id BIGSERIAL PRIMARY KEY,
    source_id TEXT,
    event_id TEXT,
    type TEXT,
    quantity INTEGER,
    target_event_id TEXT,
    event_time TIMESTAMPTZ,
    status TEXT NOT NULL CHECK (status IN ('ACCEPTED', 'PENDING_REFERENCE', 'DUPLICATE', 'CONFLICT', 'REJECTED')),
    reason TEXT,
    raw_payload JSONB,
    challenge_id TEXT,
    received_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS mqtt_challenges (
    id BIGSERIAL PRIMARY KEY,
    challenge_id TEXT UNIQUE NOT NULL,
    candidate_id TEXT NOT NULL,
    command TEXT NOT NULL,
    sent_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    events_payload JSONB,
    status TEXT NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'COMPLETED', 'FAILED', 'EXPIRED', 'DUPLICATE')),
    response_payload JSONB,
    received_at TIMESTAMPTZ DEFAULT NOW(),
    processed_at TIMESTAMPTZ
);

-- +migrate Down
DROP TABLE IF EXISTS mqtt_challenges;
DROP TABLE IF EXISTS submission_attempts;
DROP TABLE IF EXISTS production_events;
DROP TABLE IF EXISTS production_sources;

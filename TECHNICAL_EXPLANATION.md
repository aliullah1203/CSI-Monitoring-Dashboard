# Technical Explanation — FSE 01

**NorthBridge Garments Production Event Processing Dashboard**
Candidate ID: `CAND-05`

---

## 1. Architecture Overview

The system follows a clean layered architecture with three independent processes communicating through shared database state:

```
                    ┌─────────────────────────────────────┐
                    │         React Dashboard              │
                    │  (polls /api/state/all every 5s)    │
                    └──────────────┬──────────────────────┘
                                   │ HTTP REST
                    ┌──────────────▼──────────────────────┐
                    │         Go HTTP Server               │
                    │    (net/http, no framework)          │
                    │                                      │
                    │  ┌──────────┐  ┌──────────────────┐ │
                    │  │ Handlers │  │  MQTT Worker     │ │
                    │  │ /events  │  │  (goroutine)     │ │
                    │  │ /state   │  │  subscribe →     │ │
                    │  │ /ack     │  │  process →       │ │
                    │  └────┬─────┘  │  publish         │ │
                    │       │        └───────┬──────────┘ │
                    └───────┼────────────────┼────────────┘
                            │                │
                    ┌───────▼────────────────▼────────────┐
                    │           Repository Layer           │
                    │  events.go / state.go / ack.go /    │
                    │  challenge.go                        │
                    └──────────────┬──────────────────────┘
                                   │ pgx/v5
                    ┌──────────────▼──────────────────────┐
                    │    Neon Serverless PostgreSQL        │
                    │  production_events (UNIQUE key)      │
                    │  submission_attempts                  │
                    │  production_sources                   │
                    │  mqtt_challenges                      │
                    └─────────────────────────────────────┘
```

**Design patterns used:**
- **Singleton** — `sync.Once` ensures one DB connection pool across the entire process
- **Repository** — each concern (events, state, ack, challenge) has its own repo struct; no SQL in handlers
- **Dependency Injection** — repos constructed in `cmd/serve.go` and injected into handlers and MQTT worker via interfaces; nothing is global

---

## 2. Event Processing Logic

Every submission — REST or MQTT — runs through `repo/events.go: ProcessEvents()` inside a **single database transaction**. This guarantees atomicity: either all events in a batch commit or none do.

### Production Quantity Validation (Change Request 01)

Before any COUNT event is processed, quantity is validated:
- Valid range: **1 to 500 inclusive**
- If `quantity < 1` or `quantity > 500` → status `REJECTED`, recorded in `submission_attempts`, does **not** increment production totals
- Applies equally to REST API submissions and MQTT challenge events
- Example: `COUNT 450 → ACCEPTED` | `COUNT 501 → REJECTED`

### COUNT event flow

```
Receive COUNT
    │
    ├─ Validate quantity: 1 ≤ qty ≤ 500
    │   └─ Fails → REJECTED, record in submission_attempts, return
    │
    ├─ Insert into production_events (status=ACCEPTED)
    │   UNIQUE(source_id, event_id) → on conflict → DUPLICATE
    │
    └─ Check for any PENDING_REFERENCE VOIDs targeting this event_id
           │
           ├─ Found → UPDATE those VOIDs to ACCEPTED, UPDATE the COUNT to VOIDED
           └─ Not found → COUNT stays ACCEPTED, awaits ACK
```

### VOID event flow

```
Receive VOID(void_event_id=X)
    │
    ├─ Find COUNT with same source_id and event_id=X
    │   ├─ Found (ACCEPTED) → mark COUNT as VOIDED, insert VOID as ACCEPTED
    │   └─ Not found → insert VOID as PENDING_REFERENCE (auto-resolves when X arrives)
    │
    └─ Net effect: VOID always cancels its COUNT; net_total unchanged
```

### Key implementation detail — rows must be closed before UPDATE

The PostgreSQL `pgx` driver does not allow a second query on the same connection while a `sql.Rows` cursor is still open. In the VOID resolution path, all pending VOID IDs are first collected into a `[]int64` slice, `rows.Close()` is called explicitly, and then the UPDATE statements run. Using `defer rows.Close()` would defer until function return — too late.

### DUPLICATE detection

The composite unique index `UNIQUE(source_id, event_id)` on `production_events` is the source of truth. Any second insert with the same pair hits `ON CONFLICT DO NOTHING`, is recorded in `submission_attempts` with status `DUPLICATE`, and returns the duplicate status to the caller. Net total is never incremented.

---

## 3. MQTT Integration

The MQTT worker (`mqtt/worker.go`) runs as a persistent goroutine started in `cmd/serve.go`.

### Connection lifecycle

```
Start()
  │
  ├─ Connect to 152.42.238.142:1883
  │   client_id = fse01-CAND-05-<random hex>
  │   last_will → topic fse-01/CAND-05/status, payload "OFFLINE"
  │
  ├─ On connect:
  │   ├─ Publish "ONLINE" → fse-01/CAND-05/status
  │   └─ Subscribe fse-01/CAND-05/challenge
  │
  ├─ Heartbeat goroutine: every 30s publish "HEARTBEAT" → status topic
  │
  └─ Message handler:
       ├─ Parse JSON challenge payload
       ├─ Validate: protocol_version=1.0, candidate_id=CAND-05,
       │            command=PROCESS_EVENTS, expires_at > now
       ├─ Check challenge_id not already in mqtt_challenges (dedup)
       ├─ Call repo.ProcessEvents() with challenge_id reference
       └─ Publish JSON response → fse-01/CAND-05/response
```

### Challenge deduplication

Each challenge is stored in the `mqtt_challenges` table before processing begins. If the same `challenge_id` arrives again (e.g. broker retry), the worker returns immediately without re-processing. This prevents double-counting production events from network retries.

### In-memory status

`mqtt.GetState()` returns the worker's current status (connected, last challenge, last response, last seen, error list) for the `/api/mqtt/status` endpoint consumed by the dashboard.

---

## 4. Database Schema

```sql
-- Registered production lines / devices
CREATE TABLE production_sources (
  id         SERIAL PRIMARY KEY,
  name       TEXT NOT NULL,
  location   TEXT,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Core event store
CREATE TABLE production_events (
  id              SERIAL PRIMARY KEY,
  source_id       INTEGER NOT NULL REFERENCES production_sources(id),
  event_id        TEXT NOT NULL,
  type            TEXT NOT NULL,           -- COUNT | VOID
  quantity        INTEGER DEFAULT 1,
  void_event_id   TEXT,                    -- set for VOID events
  status          TEXT NOT NULL,           -- ACCEPTED | VOIDED | PENDING_REFERENCE | CONFLICT
  challenge_id    INTEGER,                 -- FK to mqtt_challenges if from MQTT
  ack_at          TIMESTAMPTZ,
  created_at      TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE (source_id, event_id)             -- prevents DUPLICATE counting
);

-- Audit log of every submission attempt
CREATE TABLE submission_attempts (
  id              SERIAL PRIMARY KEY,
  source_id       INTEGER NOT NULL,
  event_id        TEXT NOT NULL,
  type            TEXT NOT NULL,
  status          TEXT NOT NULL,           -- ACCEPTED | DUPLICATE | PENDING_REFERENCE | CONFLICT | REJECTED
  challenge_id    INTEGER,
  created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- MQTT challenge deduplication log
CREATE TABLE mqtt_challenges (
  id           SERIAL PRIMARY KEY,
  challenge_id TEXT NOT NULL UNIQUE,
  status       TEXT NOT NULL DEFAULT 'PENDING',  -- PENDING | COMPLETED | FAILED
  received_at  TIMESTAMPTZ DEFAULT NOW(),
  completed_at TIMESTAMPTZ
);
```

### Key design decisions

| Decision | Reason |
|----------|--------|
| `UNIQUE(source_id, event_id)` composite key | Single DB constraint handles all duplicate detection — no application-level check needed |
| `submission_attempts` separate from `production_events` | Clean audit trail without polluting the event store; duplicates and rejections recorded without a row in events |
| `PENDING_REFERENCE` status | Allows VOID to arrive before COUNT; event is held and auto-resolved when COUNT arrives — no data loss |
| Single `ProcessEvents` transaction | All events in a batch either commit together or roll back — no partial state |
| Neon serverless PostgreSQL | Scales to zero, no server management; connection pooling via built-in pooler endpoint |
| Quantity cap (1–500) at repo layer | Validation lives in `processSingle()` before any DB write; applies to both REST and MQTT with no code duplication |
| `rejected_submissions` from `submission_attempts` | Counted from persistent DB rows (not memory), filtered to `status='REJECTED'` only — excludes DUPLICATE, CONFLICT, PENDING_REFERENCE |

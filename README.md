# Production Event Processing Dashboard — FSE 01

**NorthBridge Garments** | Candidate ID: `CAND-05`

Real-time production event processing system with MQTT device integration. Events arrive via MQTT broker or REST API, are batch-processed in a single DB transaction, and streamed to a live dashboard.

---

## Stack

| Layer    | Technology                                |
|----------|-------------------------------------------|
| Frontend | React 19 + Vite + Tailwind CSS v4         |
| Backend  | Go (net/http, no framework)               |
| Database | Neon Serverless PostgreSQL (pgx/v5 + sql-migrate) |
| MQTT     | Eclipse Paho (paho.mqtt.golang)           |

---

## Repository Structure

```
monitoring/
├── backend/
│   ├── cmd/serve.go          # Wire-up: config → DB → MQTT → server
│   ├── config/config.go      # Env loader
│   ├── infra/
│   │   ├── connection.go     # Singleton DB (sync.Once)
│   │   └── migrate.go        # sql-migrate runner
│   ├── migrations/migrations/
│   │   └── 001_initial.sql   # 4 tables
│   ├── mqtt/worker.go        # MQTT subscribe/publish/heartbeat
│   ├── repo/                 # DB layer (events, state, ack, challenge)
│   ├── rest/
│   │   ├── server.go         # Middleware + route registration
│   │   └── handlers/         # events, state, ack, mqttstatus
│   ├── tests/api_test.go     # 5 integration tests (real DB)
│   └── .env
└── frontend/
    ├── src/
    │   ├── pages/Dashboard.jsx
    │   ├── components/ui/    # StatTiles, PendingTable, ExceptionsTable, EventForm, MqttPanel
    │   ├── hooks/useStateData.js
    │   └── index.css
    └── .env
```

---

## Setup

### Prerequisites
- Go 1.22+
- Node.js 20+
- Access to the Neon database (credentials in `.env`)

### 1 — Backend `.env`

```
# backend/.env
VERSION=1.0
SERVICENAME=production-monitor
HTTPPORT=8080
DB_STRING=postgresql://neondb_owner:<password>@<host>/neondb?sslmode=require&channel_binding=require
CANDIDATE_ID=CAND-05
MQTT_BROKER=152.42.238.142
MQTT_PORT=1883
FRONTEND_URL=http://localhost:5173
```

### 2 — Frontend `.env`

```
# frontend/.env
VITE_API_URL=http://localhost:8080
```

### 3 — Install Backend Dependencies

```bash
cd backend
go mod tidy
```

### 4 — Install Frontend Dependencies

```bash
cd frontend
npm install
```

---

## PostgreSQL Initialization

Migrations run automatically on server start via `sql-migrate`. The migration file lives at:

```
backend/migrations/migrations/001_initial.sql
```

Tables created:

| Table                | Purpose                                      |
|----------------------|----------------------------------------------|
| `production_sources` | Registered devices/lines                     |
| `production_events`  | Events with status (ACCEPTED, VOIDED, etc.)  |
| `submission_attempts`| Every inbound attempt with dedup outcome     |
| `mqtt_challenges`    | MQTT challenge deduplication log             |

Unique constraint on `production_events(source_id, event_id)` prevents double-counting.

---

## Run Commands

### Start Backend

```bash
cd backend
go run main.go
# Server starts on :8080
# Migrations apply automatically
# MQTT worker connects to 152.42.238.142:1883
```

### Start Frontend

```bash
cd frontend
npm run dev
# Opens on http://localhost:5173 (or next available port)
```

---

## Run Tests (5 Automated Integration Tests)

```bash
cd backend
go test ./tests/... -v -timeout 60s
```

| Test | What it verifies |
|------|-----------------|
| `TestCountAddsToNetTotal` | COUNT event increments net_total by 1 |
| `TestDuplicateDoesNotDoubleCount` | Same event_id submitted twice — rejected, net_total stable |
| `TestVoidBeforeCountAutoResolves` | VOID arriving before its COUNT auto-resolves when COUNT arrives |
| `TestRepeatedAckReturnsAlreadyAcked` | Second ACK on same event returns `ALREADY_ACKED` |
| `TestRepeatedMqttChallengeNotReprocessed` | Same MQTT challenge_id not processed twice |

---

## REST API Reference

### Base URL: `http://localhost:8080`

---

### POST `/api/events/process` — Submit Events

Single event:
```bash
curl -X POST http://localhost:8080/api/events/process \
  -H "Content-Type: application/json" \
  -d '{"source_id":1,"event_id":"EVT-001","type":"COUNT","quantity":1,"timestamp":"2026-10-09T10:00:00Z"}'
```

Batch (array):
```bash
curl -X POST http://localhost:8080/api/events/process \
  -H "Content-Type: application/json" \
  -d '[
    {"source_id":1,"event_id":"EVT-002","type":"COUNT","quantity":1,"timestamp":"2026-10-09T10:01:00Z"},
    {"source_id":1,"event_id":"EVT-003","type":"VOID","void_event_id":"EVT-002","timestamp":"2026-10-09T10:02:00Z"}
  ]'
```

Response:
```json
[
  {"event_id":"EVT-002","status":"ACCEPTED"},
  {"event_id":"EVT-003","status":"ACCEPTED"}
]
```

Statuses: `ACCEPTED` | `DUPLICATE` | `PENDING_REFERENCE` | `CONFLICT` | `REJECTED`

**Quantity Validation (Change Request 01):**
- COUNT quantity must be between **1 and 500 inclusive**
- `COUNT 450 → ACCEPTED` | `COUNT 501 → REJECTED`
- Rejected events are recorded in `submission_attempts` and do not increase production totals
- Applies to both REST API submissions and MQTT challenge events

---

### GET `/api/state/all` — Full Dashboard State (single call)

```bash
curl http://localhost:8080/api/state/all
```

With optional source filter:
```bash
curl "http://localhost:8080/api/state/all?source_id=LINE-01"
```

Response:
```json
{
  "summary": {
    "net_total": 42,
    "processed_events": 55,
    "pending_ack": 3,
    "unresolved": 1,
    "duplicates": 2,
    "conflicts": 0,
    "rejected_submissions": 4
  },
  "pending": [...],
  "exceptions": [...]
}
```

---

### GET `/api/state` — Individual Views

```bash
# Summary only
curl "http://localhost:8080/api/state?view=summary"

# Pending ACK list
curl "http://localhost:8080/api/state?view=pending"

# Exceptions (duplicates, conflicts, unresolved)
curl "http://localhost:8080/api/state?view=exceptions"
```

---

### POST `/api/ack` — Acknowledge Events

```bash
curl -X POST http://localhost:8080/api/ack \
  -H "Content-Type: application/json" \
  -d '{"event_ids":[101,102,103]}'
```

Response:
```json
[
  {"event_id":101,"status":"ACKED"},
  {"event_id":102,"status":"ALREADY_ACKED"},
  {"event_id":103,"status":"NOT_FOUND"}
]
```

Statuses: `ACKED` | `ALREADY_ACKED` | `NOT_READY` | `NOT_FOUND`

---

### GET `/api/mqtt/status` — MQTT Worker State

```bash
curl http://localhost:8080/api/mqtt/status
```

Response:
```json
{
  "connected": true,
  "candidate_id": "CAND-05",
  "topic": "fse-01/CAND-05/challenge",
  "last_challenge_id": "ch-abc123",
  "last_response": "ACCEPTED",
  "last_seen": "2026-10-09T10:05:00Z",
  "errors": []
}
```

---

## MQTT Integration

### Broker

```
Host: 152.42.238.142
Port: 1883
Client ID: fse01-CAND-05-<random hex>
```

### Topics

| Direction | Topic | Purpose |
|-----------|-------|---------|
| Subscribe | `fse-01/CAND-05/challenge` | Receive event batches from simulator |
| Publish   | `fse-01/CAND-05/response`  | Send processing results back |
| Last Will | `fse-01/CAND-05/status`    | Publishes `OFFLINE` on disconnect |

### Heartbeat

The worker publishes `HEARTBEAT` to the status topic every 30 seconds while connected.

### Sample MQTT Challenge Payload (inbound)

```json
{
  "protocol_version": "1.0",
  "candidate_id": "CAND-05",
  "command": "PROCESS_EVENTS",
  "challenge_id": "ch-xyz-001",
  "expires_at": "2026-10-09T12:00:00Z",
  "events": [
    {"source_id":1,"event_id":"MQTT-001","type":"COUNT","quantity":1,"timestamp":"2026-10-09T10:00:00Z"},
    {"source_id":1,"event_id":"MQTT-002","type":"COUNT","quantity":1,"timestamp":"2026-10-09T10:01:00Z"}
  ]
}
```

### Sample MQTT Response (outbound)

```json
{
  "challenge_id": "ch-xyz-001",
  "candidate_id": "CAND-05",
  "results": [
    {"event_id":"MQTT-001","status":"ACCEPTED"},
    {"event_id":"MQTT-002","status":"ACCEPTED"}
  ]
}
```

### Validation Rules (worker rejects if violated)

- `protocol_version` must be `"1.0"`
- `candidate_id` must match `CAND-05`
- `command` must be `"PROCESS_EVENTS"`
- `expires_at` must be in the future
- Duplicate `challenge_id` → logged, not reprocessed

# AI Usage Log — FSE 01

**Project:** Production Event Processing Dashboard + MQTT Device Integration
**Candidate:** CAND-05
**AI Assistant:** Claude (Anthropic) via Claude Code CLI

---

## How AI Was Used

This project was built collaboratively with an AI assistant. Below is a summary of the conversation and what the AI helped produce.

---

## Conversation Summary

### 1. Project Understanding

**User:** Shared the FSE 01 project requirement PDF and asked the AI to read it before starting. Mentioned stack: React.js frontend, Go backend, Neon serverless PostgreSQL. Also shared their existing ATMOS project as the reference architecture.

**AI:** Read both the requirement and ATMOS architecture. Confirmed the plan: use ATMOS file structure for both frontend and backend, implement all required features on top.

---

### 2. Architecture Planning

**User:** "now my plan is for frontend in one page name is dashboard i will separate task production supervisor, factory floor operator, support engineering team. flow is one part to another part data will flow right? so now give me a plan for frontend dashboard"

**AI:** Proposed a single Dashboard page with 3 vertical sections:
- **Production Supervisor** — stat tiles (net total, processed, pending ACK, unresolved, duplicates, conflicts) + pending ACK table + exceptions table
- **Factory Floor Operator** — manual event submission form (COUNT/VOID toggle)
- **Support / Engineering** — MQTT status panel (connection info, last challenge, last response)

---

### 3. Backend Implementation

**User:** Shared Neon PostgreSQL connection string. Asked to wire up: `main.go → serve.go → config → DB connection → migrations → repo → handlers → server`. Requested singleton, repository, and dependency injection patterns.

**AI:** Built the complete backend:
- `cmd/serve.go` — full wire-up sequence
- `config/config.go` — env loader
- `infra/connection.go` — `sync.Once` singleton DB pool
- `infra/migrate.go` — sql-migrate runner
- `migrations/001_initial.sql` — 4 tables with composite unique key
- `repo/events.go` — batch processing in single transaction, VOID/COUNT resolution
- `repo/state.go` — summary, pending, exceptions queries
- `repo/ack.go` — acknowledge with status per event
- `repo/challenge.go` — MQTT challenge deduplication
- `mqtt/worker.go` — connect, subscribe, validate, process, publish, heartbeat, last-will
- `rest/server.go` — middleware + route registration
- All handlers under `rest/handlers/`

---

### 4. Bug Fix — VOID Resolution Deadlock

**User:** Test 3 (VOID before COUNT) was failing.

**AI:** Diagnosed root cause: `defer rows.Close()` was keeping the PostgreSQL cursor open when `tx.Exec()` UPDATE statements ran on the same connection — the driver does not allow this. Fixed by collecting all VOID IDs into a `[]int64` slice first, calling `rows.Close()` explicitly, then running all UPDATEs. Test 3 passed after the fix.

---

### 5. Frontend Implementation

**User:** Asked to build the React frontend matching the 3-section dashboard plan.

**AI:** Built the complete frontend:
- `src/pages/Dashboard.jsx` — 3 labelled sections
- `src/components/ui/StatTiles.jsx` — 6 colour-coded metric tiles
- `src/components/ui/PendingTable.jsx` — clickable rows with ACK button
- `src/components/ui/ExceptionsTable.jsx` — status badges per event
- `src/components/ui/EventForm.jsx` — COUNT/VOID toggle form
- `src/components/ui/MqttPanel.jsx` — MQTT info grid
- `src/components/layout/Navbar.jsx` — sticky header with live MQTT indicator
- `src/hooks/useStateData.js` — polls `/api/state/all` + `/api/mqtt/status` every 5s
- `src/index.css` — GitHub dark theme, CSS variables, badge/button/input classes

---

### 6. Performance Optimisation

**User:** Noticed frontend was making 3 separate API calls per poll cycle.

**AI:** Added `/api/state/all` endpoint that returns summary + pending + exceptions in one DB call. Updated `useStateData.js` to use it. Reduced from 3 calls per 5s cycle to 1.

---

### 7. UI Visibility Improvements

**User:** "each box color and background color looking almost similar"
**User:** "box need to more light color for more visible"
**User:** "for 3 part supervisor, factory floor, mqtt connection whole box need to more visible, and all text also"

**AI:** Iteratively improved the theme:
- Raised `--bg-card` from `#0f1623` → `#161b22` → `#21262d`
- Brightened `--text-secondary` from `#6e7681` → `#8b949e` → `#cdd9e5`
- Brightened `--text-primary` to `#f0f6fc`
- Upgraded `--border` from `#21262d` → `#30363d` → `#444c56`
- Increased stat tile number font size to 36px
- Each tile given a distinct saturated colour (blue, green, amber, orange, purple, red)

---

### 8. Automated Tests

**User:** "just create some test as like test1, test2 so that all api including mqtt data sending receiving, frontend stuff, database all are good"

**AI:** Wrote 5 integration tests in `backend/tests/api_test.go` against the real Neon database:

| Test | Result |
|------|--------|
| `TestCountAddsToNetTotal` | PASS |
| `TestDuplicateDoesNotDoubleCount` | PASS |
| `TestVoidBeforeCountAutoResolves` | PASS |
| `TestRepeatedAckReturnsAlreadyAcked` | PASS |
| `TestRepeatedMqttChallengeNotReprocessed` | PASS |

---

### 9. Documentation

**User:** "create readme.md with exact setup, postgresql initialization, run commands, tests, all rest request example, mqtt topics and sample flow. TECHNICAL_EXPLANATION.md just do first 1 to 4."

**AI:** Created:
- `README.md` — full setup guide, schema, run commands, all REST API curl examples, MQTT topic table and sample payloads
- `TECHNICAL_EXPLANATION.md` — architecture diagram, event processing logic, MQTT lifecycle, database schema with design decisions

---

### 10. Final Verification

**User:** "yes some event automatic coming to supervisor pending list"

The MQTT simulator was confirmed to be sending challenge payloads live. Events were appearing in the Supervisor Pending ACK list in real time — confirming end-to-end MQTT → backend → database → frontend flow working correctly.

**User:** "now it's ready for submit?"

**AI:** Confirmed all requirements complete and project ready to submit.

---

## Summary of AI Contributions

| Area | Contribution |
|------|-------------|
| Architecture | Designed the 3-section dashboard layout and Go backend structure |
| Backend | Wrote all Go code (config, DB, migrations, repos, handlers, MQTT worker) |
| Business Logic | Implemented COUNT/VOID/PENDING_REFERENCE/DUPLICATE processing in single TX |
| Bug Fix | Diagnosed and fixed PostgreSQL cursor/transaction deadlock in VOID resolution |
| Frontend | Wrote all React components, hooks, and CSS theme |
| Optimisation | Reduced 3 API calls per cycle to 1 with `/api/state/all` endpoint |
| UI Tuning | Iteratively improved colour contrast and visibility based on user feedback |
| Tests | Wrote 5 integration tests against real Neon DB — all passing |
| Docs | Wrote README.md and TECHNICAL_EXPLANATION.md |

---

*AI assistant: Claude Sonnet 4.6 (Anthropic) — accessed via Claude Code CLI*

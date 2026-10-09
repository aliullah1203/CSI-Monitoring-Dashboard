# Requirement Decisions — FSE 01

**Project:** Production Event Processing Dashboard + MQTT Device Integration
**Candidate:** CAND-05

This document explains the key architectural and implementation decisions made during development, and why each choice was made.

---

## 1. Single Dashboard Page (Three Sections)

**Decision:** One page with three vertical sections — Production Supervisor, Factory Floor Operator, Support/Engineering — instead of multiple pages or tabs.

**Reason:** The requirement described a flow where data produced by Factory Floor feeds directly into the Supervisor view. A single page makes this live data flow visible to all roles simultaneously without navigation. Each section serves a distinct role but shares the same live data source.

---

## 2. Single `/api/state/all` Polling Endpoint

**Decision:** Created one combined endpoint `/api/state/all` returning summary + pending + exceptions in a single DB call, rather than three separate endpoints polled independently.

**Reason:** The original three-endpoint approach made 3 DB calls per 5-second poll cycle (1.5 seconds of DB activity per cycle). A single query per cycle reduces DB load and keeps the dashboard consistent — all three sections always reflect the same snapshot.

---

## 3. Handler-Per-File Architecture (ATMOS Pattern)

**Decision:** Each handler lives in its own file and package under `rest/handlers/<domain>/`. No framework like Gin or Echo — pure `net/http`.

**Reason:** Follows the candidate's existing ATMOS project architecture for consistency and maintainability. Pure `net/http` avoids framework dependency while keeping handlers thin — each one validates input, calls the repo, and returns JSON.

---

## 4. Single DB Transaction for Batch Processing

**Decision:** All events in a `ProcessEvents()` call commit in one transaction, or none do.

**Reason:** The requirement specifies batch processing. A partial commit (some events saved, some not) would corrupt the net_total count. One transaction guarantees all-or-nothing atomicity for every batch — whether from REST or MQTT.

---

## 5. UNIQUE(source_id, event_id) Composite Key for Duplicate Detection

**Decision:** Duplicate detection handled entirely by a PostgreSQL UNIQUE constraint on `production_events(source_id, event_id)` rather than application-level checks.

**Reason:** A DB constraint is the only guarantee that survives concurrent requests. Application-level duplicate checks have a race window — two simultaneous inserts could both pass the check and both succeed. The constraint makes it impossible at the DB level regardless of concurrency.

---

## 6. PENDING_REFERENCE Status for Out-of-Order VOID

**Decision:** When a VOID arrives before its target COUNT, it is stored as `PENDING_REFERENCE` and auto-resolved when the COUNT arrives.

**Reason:** The requirement says VOIDs cancel COUNTs. If we simply reject a VOID whose COUNT hasn't arrived yet, we lose data. `PENDING_REFERENCE` holds the VOID safely and resolves it automatically inside the same transaction when the COUNT is processed — no data loss, no manual intervention.

---

## 7. Repository Pattern with Dependency Injection

**Decision:** `EventsRepo`, `StateRepo`, `AckRepo`, `ChallengeRepo` are separate structs constructed in `cmd/serve.go` and injected into handlers and the MQTT worker via interfaces.

**Reason:** Testability and separation of concerns. The 5 integration tests instantiate the repos directly against the real Neon DB — no mocking needed. Handlers never import the DB package; they receive an interface. Swapping the DB implementation requires changing only the repo, not the handlers.

---

## 8. MQTT Challenge Deduplication via Database

**Decision:** Every MQTT challenge ID is stored in `mqtt_challenges` before processing. Duplicate challenge IDs are rejected before any events are processed.

**Reason:** MQTT brokers can retry message delivery. If the same challenge is delivered twice (network hiccup), we must not double-count the events. Storing the challenge ID in the DB (with a UNIQUE constraint) guarantees idempotency even across server restarts.

---

## 9. Quantity Validation at Repo Layer (Change Request 01)

**Decision:** Quantity validation (1–500) is placed in `repo/events.go: processSingle()`, not in the HTTP handler or the MQTT worker.

**Reason:** The rule must apply equally to both REST submissions and MQTT challenge events. Placing it in the shared repo layer means there is one copy of the business rule, enforced consistently regardless of entry point. Handlers and the MQTT worker do not duplicate the logic.

---

## 10. rejected_submissions Counted from PostgreSQL (Change Request 02)

**Decision:** `rejected_submissions` is counted from `submission_attempts` table rows with `status='REJECTED'`, calculated on every `GetSummary()` call.

**Reason:** The requirement explicitly states "calculate from persistent PostgreSQL data, not hardcoded or memory-only values." The `submission_attempts` table already stores every inbound attempt with its outcome status. No new table or column was needed — just an additional `COUNT` subquery in the existing summary SQL.

---

## 11. Source Filter Applied at API Level (Change Request 03)

**Decision:** The source filter passes `source_id` as a query parameter to `/api/state/all?source_id=X` rather than filtering on the frontend.

**Reason:** Frontend filtering would require fetching all data and discarding most of it. API-level filtering lets the DB do the work with indexed lookups. The `source_id` column is already used in all WHERE clauses across `GetSummary`, `GetPending`, and `GetExceptions`, so the filter cost is minimal.

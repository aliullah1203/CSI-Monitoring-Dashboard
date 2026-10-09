package tests

import (
	"fmt"
	"os"
	"testing"
	"time"

	"backend/infra"
	"backend/repo"

	"github.com/joho/godotenv"
)

func setupRepos(t *testing.T) (*repo.EventsRepo, *repo.StateRepo, *repo.AckRepo, *repo.ChallengeRepo) {
	t.Helper()
	godotenv.Load("../.env")
	db, err := infra.NewConnection(os.Getenv("DB_STRING"))
	if err != nil {
		t.Fatalf("DB connection failed: %v", err)
	}
	if err := infra.MigrateDB(db, "../migrations/migrations"); err != nil {
		t.Fatalf("migration failed: %v", err)
	}
	return repo.NewEventsRepo(db), repo.NewStateRepo(db), repo.NewAckRepo(db), repo.NewChallengeRepo(db)
}

// unique event ID per test run to avoid conflicts
func uid(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

// Test1: COUNT event is accepted and net_total increases
func TestCountAddsToNetTotal(t *testing.T) {
	eventsRepo, stateRepo, _, _ := setupRepos(t)

	before, _ := stateRepo.GetSummary("")

	eventID := uid("T1-EV")
	qty := 7
	results, err := eventsRepo.ProcessEvents([]repo.EventInput{
		{SourceID: "LINE-01", EventID: eventID, Type: "COUNT", Quantity: &qty, EventTime: time.Now().UTC().Format(time.RFC3339)},
	}, nil)
	if err != nil {
		t.Fatalf("ProcessEvents error: %v", err)
	}
	if results[0].Status != "ACCEPTED" {
		t.Fatalf("expected ACCEPTED, got %s", results[0].Status)
	}

	after, _ := stateRepo.GetSummary("")
	if after.NetTotal != before.NetTotal+7 {
		t.Fatalf("net_total: expected %d, got %d", before.NetTotal+7, after.NetTotal)
	}
	t.Logf("✅ Test1 PASS: net_total went from %d to %d", before.NetTotal, after.NetTotal)
}

// Test2: Same event submitted twice → second is DUPLICATE, no double counting
func TestDuplicateDoesNotDoubleCount(t *testing.T) {
	eventsRepo, stateRepo, _, _ := setupRepos(t)

	eventID := uid("T2-EV")
	qty := 5
	input := []repo.EventInput{
		{SourceID: "LINE-01", EventID: eventID, Type: "COUNT", Quantity: &qty, EventTime: time.Now().UTC().Format(time.RFC3339)},
	}

	r1, _ := eventsRepo.ProcessEvents(input, nil)
	if r1[0].Status != "ACCEPTED" {
		t.Fatalf("first submit expected ACCEPTED, got %s", r1[0].Status)
	}

	after1, _ := stateRepo.GetSummary("")

	r2, _ := eventsRepo.ProcessEvents(input, nil)
	if r2[0].Status != "DUPLICATE" {
		t.Fatalf("second submit expected DUPLICATE, got %s", r2[0].Status)
	}

	after2, _ := stateRepo.GetSummary("")
	if after2.NetTotal != after1.NetTotal {
		t.Fatalf("net_total changed on duplicate: before=%d after=%d", after1.NetTotal, after2.NetTotal)
	}
	t.Logf("✅ Test2 PASS: duplicate rejected, net_total stable at %d", after2.NetTotal)
}

// Test3: VOID before COUNT → PENDING_REFERENCE → auto-resolved when COUNT arrives
func TestVoidBeforeCountAutoResolves(t *testing.T) {
	eventsRepo, stateRepo, _, _ := setupRepos(t)

	countID := uid("T3-COUNT")
	voidID := uid("T3-VOID")
	qty := 10

	// Send VOID first (COUNT doesn't exist yet)
	r1, err := eventsRepo.ProcessEvents([]repo.EventInput{
		{SourceID: "LINE-01", EventID: voidID, Type: "VOID", TargetEventID: &countID, EventTime: time.Now().UTC().Format(time.RFC3339)},
	}, nil)
	if err != nil {
		t.Fatalf("VOID error: %v", err)
	}
	if r1[0].Status != "PENDING_REFERENCE" {
		t.Fatalf("expected PENDING_REFERENCE, got %s", r1[0].Status)
	}

	before, _ := stateRepo.GetSummary("")

	// Now send the COUNT — should auto-resolve the VOID
	r2, err := eventsRepo.ProcessEvents([]repo.EventInput{
		{SourceID: "LINE-01", EventID: countID, Type: "COUNT", Quantity: &qty, EventTime: time.Now().UTC().Format(time.RFC3339)},
	}, nil)
	if err != nil {
		t.Fatalf("COUNT error: %v", err)
	}
	if r2[0].Status != "ACCEPTED" {
		t.Fatalf("COUNT expected ACCEPTED, got %s", r2[0].Status)
	}

	after, _ := stateRepo.GetSummary("")
	// COUNT was voided immediately by the pending VOID, so net_total should NOT increase
	if after.NetTotal != before.NetTotal {
		t.Fatalf("voided COUNT should not add to net_total: before=%d after=%d", before.NetTotal, after.NetTotal)
	}
	// Unresolved should not have increased (VOID was resolved)
	if after.Unresolved > before.Unresolved {
		t.Fatalf("unresolved should not increase after resolution: before=%d after=%d", before.Unresolved, after.Unresolved)
	}
	t.Logf("✅ Test3 PASS: VOID resolved automatically, net_total stable at %d", after.NetTotal)
}

// Test4: Acknowledging same event twice → second is ALREADY_ACKED
func TestRepeatedAckReturnsAlreadyAcked(t *testing.T) {
	eventsRepo, _, ackRepo, _ := setupRepos(t)

	eventID := uid("T4-EV")
	qty := 3
	r, _ := eventsRepo.ProcessEvents([]repo.EventInput{
		{SourceID: "LINE-01", EventID: eventID, Type: "COUNT", Quantity: &qty, EventTime: time.Now().UTC().Format(time.RFC3339)},
	}, nil)
	if r[0].Status != "ACCEPTED" {
		t.Fatalf("setup: expected ACCEPTED, got %s", r[0].Status)
	}

	ack1, _ := ackRepo.AckEvents([]string{eventID})
	if ack1[0].Status != "ACKED" {
		t.Fatalf("first ack expected ACKED, got %s", ack1[0].Status)
	}

	ack2, _ := ackRepo.AckEvents([]string{eventID})
	if ack2[0].Status != "ALREADY_ACKED" {
		t.Fatalf("second ack expected ALREADY_ACKED, got %s", ack2[0].Status)
	}
	t.Logf("✅ Test4 PASS: repeated ack → ALREADY_ACKED")
}

// Test5: Same MQTT challenge_id submitted twice → second returns CHALLENGE_CONFLICT, events not reprocessed
func TestRepeatedMqttChallengeNotReprocessed(t *testing.T) {
	eventsRepo, stateRepo, _, challengeRepo := setupRepos(t)

	challengeID := uid("CH")
	eventID := uid("T5-EV")
	qty := 4
	cid := challengeID

	// First challenge: store + process
	if err := challengeRepo.Store(challengeID, "CAND-05", "PROCESS_EVENTS", time.Now(), time.Now().Add(time.Minute), []byte(`[]`)); err != nil {
		t.Fatalf("store challenge: %v", err)
	}
	r1, _ := eventsRepo.ProcessEvents([]repo.EventInput{
		{SourceID: "LINE-01", EventID: eventID, Type: "COUNT", Quantity: &qty, EventTime: time.Now().UTC().Format(time.RFC3339)},
	}, &cid)
	if r1[0].Status != "ACCEPTED" {
		t.Fatalf("first challenge event expected ACCEPTED, got %s", r1[0].Status)
	}
	after1, _ := stateRepo.GetSummary("")

	// Second challenge with same challenge_id: should be detected as duplicate
	isDup := challengeRepo.IsDuplicate(challengeID)
	if !isDup {
		t.Fatal("expected challenge to be detected as duplicate")
	}

	// Events should NOT be reprocessed (same event_id → DUPLICATE)
	r2, _ := eventsRepo.ProcessEvents([]repo.EventInput{
		{SourceID: "LINE-01", EventID: eventID, Type: "COUNT", Quantity: &qty, EventTime: time.Now().UTC().Format(time.RFC3339)},
	}, &cid)
	if r2[0].Status != "DUPLICATE" {
		t.Fatalf("reprocessed event expected DUPLICATE, got %s", r2[0].Status)
	}

	after2, _ := stateRepo.GetSummary("")
	if after2.NetTotal != after1.NetTotal {
		t.Fatalf("net_total changed on challenge replay: before=%d after=%d", after1.NetTotal, after2.NetTotal)
	}
	t.Logf("✅ Test5 PASS: challenge duplicate detected, no double counting")
}

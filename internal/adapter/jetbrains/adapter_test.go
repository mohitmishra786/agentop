package jetbrains

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/agentop-dev/agentop/internal/adapter"
)

func testDataPath(t *testing.T, rel string) string {
	t.Helper()
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "../../../testdata/adapters", rel)
}

func TestParseSession(t *testing.T) {
	a := &Adapter{}
	result, err := a.ParseSession(testDataPath(t, "jetbrains/partition-001.jsonl"))
	if err != nil {
		t.Fatalf("ParseSession failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Events) < 2 {
		t.Fatalf("expected at least 2 events, got %d", len(result.Events))
	}

	if result.Events[0].Type != "user" {
		t.Errorf("expected first event type 'user', got %q", result.Events[0].Type)
	}
	if result.Events[0].CWD != "/test/project" {
		t.Errorf("expected first event CWD '/test/project', got %q", result.Events[0].CWD)
	}

	if result.Events[1].Type != "assistant" {
		t.Errorf("expected second event type 'assistant', got %q", result.Events[1].Type)
	}
}

func TestDiscover(t *testing.T) {
	a := &Adapter{}
	files, err := a.Discover(testDataPath(t, "jetbrains"))
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}
	_ = files
}

func TestAdapterID(t *testing.T) {
	a := &Adapter{}
	if a.ID() != adapter.AgentJetBrains {
		t.Errorf("expected %s, got %s", adapter.AgentJetBrains, a.ID())
	}
}

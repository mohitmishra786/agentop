package copilot

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
	result, err := a.ParseSession(testDataPath(t, "copilot/session.jsonl"))
	if err != nil {
		t.Fatalf("ParseSession failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Events) == 0 {
		t.Fatal("expected at least one event")
	}
}

func TestDiscover(t *testing.T) {
	a := &Adapter{}
	files, err := a.Discover(testDataPath(t, "copilot"))
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}
	_ = files
}

func TestAdapterID(t *testing.T) {
	a := &Adapter{}
	if a.ID() != adapter.AgentCopilot {
		t.Errorf("expected %s, got %s", adapter.AgentCopilot, a.ID())
	}
}

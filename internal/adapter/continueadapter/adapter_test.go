package continueadapter

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
	result, err := a.ParseSession(testDataPath(t, "continue/session.json"))
	if err != nil {
		t.Fatalf("ParseSession failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(result.Events))
	}

	if result.Events[0].Type != "user" {
		t.Errorf("expected first event type 'user', got %q", result.Events[0].Type)
	}
	if result.Events[1].Type != "assistant" {
		t.Errorf("expected second event type 'assistant', got %q", result.Events[1].Type)
	}
	if result.Events[2].Type != "user" {
		t.Errorf("expected third event type 'user', got %q", result.Events[2].Type)
	}

	if result.Events[1].Message == nil {
		t.Fatal("expected non-nil Message on assistant event")
	}
	if result.Events[1].Message.Model != "claude-sonnet-4-6" {
		t.Errorf("expected assistant model 'claude-sonnet-4-6', got %q", result.Events[1].Message.Model)
	}
	if result.Events[1].Message.Usage == nil {
		t.Fatal("expected non-nil Usage on assistant event")
	}
	if result.Events[1].Message.Usage.InputTokens != 8000 {
		t.Errorf("expected InputTokens 8000, got %d", result.Events[1].Message.Usage.InputTokens)
	}
	if result.Events[1].Message.Usage.OutputTokens != 1500 {
		t.Errorf("expected OutputTokens 1500, got %d", result.Events[1].Message.Usage.OutputTokens)
	}
	if result.Events[1].CostUSD != 0.12 {
		t.Errorf("expected CostUSD 0.12, got %f", result.Events[1].CostUSD)
	}

	if result.Meta == nil {
		t.Fatal("expected non-nil Meta")
	}
	if result.Meta.ID != "continue-session-1" {
		t.Errorf("expected meta ID 'continue-session-1', got %q", result.Meta.ID)
	}
	if result.Meta.MessageCount != 3 {
		t.Errorf("expected MessageCount 3, got %d", result.Meta.MessageCount)
	}
}

func TestDiscover(t *testing.T) {
	a := &Adapter{}
	files, err := a.Discover(testDataPath(t, "continue"))
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}
	_ = files
}

func TestAdapterID(t *testing.T) {
	a := &Adapter{}
	if a.ID() != adapter.AgentContinue {
		t.Errorf("expected %s, got %s", adapter.AgentContinue, a.ID())
	}
}

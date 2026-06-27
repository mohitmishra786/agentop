package kiro

import (
	"path/filepath"
	"runtime"
	"strings"
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
	_, err := a.ParseSession(testDataPath(t, "kiro/session.jsonl"))
	if err == nil {
		t.Fatal("expected error for .jsonl file without .json metadata")
	}
	if !strings.Contains(err.Error(), "expected .json metadata file") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestDiscover(t *testing.T) {
	a := &Adapter{}
	files, err := a.Discover(testDataPath(t, "kiro"))
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}
	_ = files
}

func TestAdapterID(t *testing.T) {
	a := &Adapter{}
	if a.ID() != adapter.AgentKiro {
		t.Errorf("expected %s, got %s", adapter.AgentKiro, a.ID())
	}
}

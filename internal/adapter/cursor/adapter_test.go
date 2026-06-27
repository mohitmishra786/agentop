package cursor

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

func TestDiscoverNonExistentDir(t *testing.T) {
	a := &Adapter{}
	files, err := a.Discover(testDataPath(t, "nonexistent"))
	if err != nil {
		t.Fatalf("Discover with non-existent dir should not error: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("expected 0 files for non-existent dir, got %d", len(files))
	}
}

func TestParseSessionFailsWithoutDB(t *testing.T) {
	a := &Adapter{}
	_, err := a.ParseSession(testDataPath(t, "cursor/nonexistent#session"))
	if err == nil {
		t.Fatal("expected error for non-existent session path")
	}
}

func TestAdapterID(t *testing.T) {
	a := &Adapter{}
	if a.ID() != adapter.AgentCursor {
		t.Errorf("expected %s, got %s", adapter.AgentCursor, a.ID())
	}
}

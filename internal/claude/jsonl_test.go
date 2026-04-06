package claude

import (
	"testing"
)

func TestParseSessionGoodCache(t *testing.T) {
	events, err := ParseSession("../../testdata/sessions/good-cache.jsonl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) == 0 {
		t.Fatal("expected events, got none")
	}

	foundUser := false
	foundAssistant := false
	for _, e := range events {
		if e.Type == "user" {
			foundUser = true
			if e.SessionID != "test-good-cache" {
				t.Errorf("expected session ID test-good-cache, got %s", e.SessionID)
			}
		}
		if e.Type == "assistant" {
			foundAssistant = true
			if e.Message == nil {
				t.Error("expected message in assistant event")
			}
			if e.Message.Usage == nil {
				t.Error("expected usage in assistant message")
			}
		}
	}
	if !foundUser {
		t.Error("expected user events")
	}
	if !foundAssistant {
		t.Error("expected assistant events")
	}
}

func TestParseSessionColdStart(t *testing.T) {
	events, err := ParseSession("../../testdata/sessions/cold-start.jsonl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) == 0 {
		t.Fatal("expected events, got none")
	}
}

func TestParseSessionInProgress(t *testing.T) {
	events, err := ParseSession("../../testdata/sessions/in-progress.jsonl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event (truncated line skipped), got %d", len(events))
	}
}

func TestParseSessionNonExistent(t *testing.T) {
	_, err := ParseSession("../../testdata/sessions/nonexistent.jsonl")
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}
}

package aggregator

import (
	"testing"
	"time"

	"github.com/agentop-dev/agentop/internal/claude"
	"github.com/agentop-dev/agentop/internal/pricing"
)

func TestAggregateSession(t *testing.T) {
	events := []claude.RawEvent{
		{
			Type:      "user",
			SessionID: "test-1",
			Timestamp: time.Date(2026, 4, 7, 10, 0, 0, 0, time.UTC),
			CWD:       "/Users/test/project",
		},
		{
			Type:      "assistant",
			SessionID: "test-1",
			Timestamp: time.Date(2026, 4, 7, 10, 0, 5, 0, time.UTC),
			Message: &claude.RawMessage{
				Model: "claude-opus-4-6",
				Usage: &claude.Usage{
					InputTokens:              10000,
					OutputTokens:             1000,
					CacheCreationInputTokens: 5000,
					CacheReadInputTokens:     50000,
				},
			},
			CostUSD: 0.50,
		},
		{
			Type:      "user",
			SessionID: "test-1",
			Timestamp: time.Date(2026, 4, 7, 10, 1, 0, 0, time.UTC),
		},
		{
			Type:      "assistant",
			SessionID: "test-1",
			Timestamp: time.Date(2026, 4, 7, 10, 1, 10, 0, time.UTC),
			Message: &claude.RawMessage{
				Model: "claude-opus-4-6",
				Usage: &claude.Usage{
					InputTokens:              9000,
					OutputTokens:             800,
					CacheCreationInputTokens: 4000,
					CacheReadInputTokens:     60000,
				},
			},
			CostUSD: 0.45,
		},
		{
			Type:      "tool",
			ToolName:  "Read",
			SessionID: "test-1",
			Timestamp: time.Date(2026, 4, 7, 10, 1, 11, 0, time.UTC),
		},
	}

	stats := AggregateSession(events, nil, pricing.DefaultPricer{})

	if stats == nil {
		t.Fatal("expected stats, got nil")
	}
	if stats.MessageCount != 2 {
		t.Errorf("expected 2 messages, got %d", stats.MessageCount)
	}
	if stats.InputTokens != 19000 {
		t.Errorf("expected 19000 input tokens, got %d", stats.InputTokens)
	}
	if stats.CostUSD != 0.95 {
		t.Errorf("expected 0.95 cost, got %f", stats.CostUSD)
	}
	if stats.Model != "claude-opus-4-6" {
		t.Errorf("expected claude-opus-4-6, got %s", stats.Model)
	}
	if stats.ProjectPath != "/Users/test/project" {
		t.Errorf("expected /Users/test/project, got %s", stats.ProjectPath)
	}
	if stats.ToolCalls["Read"] != 1 {
		t.Errorf("expected 1 Read tool call, got %d", stats.ToolCalls["Read"])
	}

	// Cache efficiency: 110000 / (19000 + 9000 + 110000) = 110000 / 138000 ≈ 0.797
	expectedEff := 110000.0 / 138000.0
	if stats.CacheEfficiency < expectedEff-0.01 || stats.CacheEfficiency > expectedEff+0.01 {
		t.Errorf("expected cache efficiency ~%.3f, got %.3f", expectedEff, stats.CacheEfficiency)
	}
}

func TestAggregateSessionEmpty(t *testing.T) {
	stats := AggregateSession(nil, nil, pricing.DefaultPricer{})
	if stats != nil {
		t.Error("expected nil for empty events")
	}
}

func TestAggregateSessionHumanType(t *testing.T) {
	events := []claude.RawEvent{
		{
			Type:      "human",
			SessionID: "test-2",
			Timestamp: time.Date(2026, 4, 7, 10, 0, 0, 0, time.UTC),
		},
		{
			Type:      "assistant",
			SessionID: "test-2",
			Timestamp: time.Date(2026, 4, 7, 10, 0, 5, 0, time.UTC),
			Message: &claude.RawMessage{
				Model: "claude-sonnet-4-6",
				Usage: &claude.Usage{
					InputTokens:  5000,
					OutputTokens: 500,
				},
			},
			CostUSD: 0.02,
		},
	}

	stats := AggregateSession(events, nil, pricing.DefaultPricer{})
	if stats.MessageCount != 1 {
		t.Errorf("expected 1 message, got %d", stats.MessageCount)
	}
}

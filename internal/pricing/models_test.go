package pricing

import (
	"testing"

	"github.com/agentop-dev/agentop/internal/claude"
)

func TestCalculateCostOpus(t *testing.T) {
	usage := claude.Usage{
		InputTokens:              1000000,
		OutputTokens:             500000,
		CacheCreationInputTokens: 200000,
		CacheReadInputTokens:     5000000,
	}

	cost := DefaultPricer{}.Calculate(usage, "claude-opus-4-6")

	// Expected: 15.00 + 37.50 + 3.75 + 7.50 = 63.75
	expected := 63.75
	if cost != expected {
		t.Errorf("expected cost %.2f, got %.2f", expected, cost)
	}
}

func TestCalculateCostSonnet(t *testing.T) {
	usage := claude.Usage{
		InputTokens:  1000000,
		OutputTokens: 500000,
	}

	cost := DefaultPricer{}.Calculate(usage, "claude-sonnet-4-6")

	// Expected: 3.00 + 7.50 = 10.50
	expected := 10.50
	if cost != expected {
		t.Errorf("expected cost %.2f, got %.2f", expected, cost)
	}
}

func TestCalculateCostHaiku(t *testing.T) {
	usage := claude.Usage{
		InputTokens:  1000000,
		OutputTokens: 500000,
	}

	cost := DefaultPricer{}.Calculate(usage, "claude-haiku-4-5")

	// Expected: 0.80 + 2.00 = 2.80
	expected := 2.80
	if cost != expected {
		t.Errorf("expected cost %.2f, got %.2f", expected, cost)
	}
}

func TestGetModelPrefixMatch(t *testing.T) {
	// Prefix match: "claude-opus-4-6-preview" should match "claude-opus-4-6"
	price := Get("claude-opus-4-6-preview")
	if price.Input != 15.00 {
		t.Errorf("expected input 15.00 for opus prefix match, got %.2f", price.Input)
	}
}

func TestGetModelFallback(t *testing.T) {
	// Unknown model should fall back to sonnet
	price := Get("unknown-model-xyz")
	if price.Input != 3.00 {
		t.Errorf("expected input 3.00 for unknown model fallback, got %.2f", price.Input)
	}
}

func TestExactMatch(t *testing.T) {
	price := Get("claude-sonnet-4")
	if price.Input != 3.00 {
		t.Errorf("expected input 3.00 for claude-sonnet-4, got %.2f", price.Input)
	}
}

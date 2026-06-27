package pricing

import (
	"testing"

	"github.com/agentop-dev/agentop/internal/adapter"
)

func TestCalculateCostOpus(t *testing.T) {
	usage := adapter.Usage{
		InputTokens:              1000000,
		OutputTokens:             500000,
		CacheCreationInputTokens: 200000,
		CacheReadInputTokens:     5000000,
	}

	cost := DefaultPricer{}.Calculate(usage, "claude-opus-4-6")

	expected := 63.75
	if cost != expected {
		t.Errorf("expected cost %.2f, got %.2f", expected, cost)
	}
}

func TestCalculateCostSonnet(t *testing.T) {
	usage := adapter.Usage{
		InputTokens:  1000000,
		OutputTokens: 500000,
	}

	cost := DefaultPricer{}.Calculate(usage, "claude-sonnet-4-6")

	expected := 10.50
	if cost != expected {
		t.Errorf("expected cost %.2f, got %.2f", expected, cost)
	}
}

func TestCalculateCostHaiku(t *testing.T) {
	usage := adapter.Usage{
		InputTokens:  1000000,
		OutputTokens: 500000,
	}

	cost := DefaultPricer{}.Calculate(usage, "claude-haiku-4-5")

	expected := 2.80
	if cost != expected {
		t.Errorf("expected cost %.2f, got %.2f", expected, cost)
	}
}

func TestGetModelPrefixMatch(t *testing.T) {
	price := Get("claude-opus-4-6-preview")
	if price.Input != 15.00 {
		t.Errorf("expected input 15.00 for opus prefix match, got %.2f", price.Input)
	}
}

func TestGetModelFallback(t *testing.T) {
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

func TestProviderExactMatchGPT4o(t *testing.T) {
	price := Get("gpt-4o")
	if price.Input != 2.50 {
		t.Errorf("expected input 2.50 for gpt-4o, got %.2f", price.Input)
	}
}

func TestProviderPrefixMatchGPT4(t *testing.T) {
	price := Get("gpt-4o-augmented")
	if price.Input != 2.50 {
		t.Errorf("expected input 2.50 for gpt-4o prefix match, got %.2f", price.Input)
	}
}

func TestProviderFallbackOpenAI(t *testing.T) {
	price := Get("gpt-4o-unknown-suffix")
	if price.Input != 2.50 {
		t.Errorf("expected input 2.50 for gpt-4o prefix match, got %.2f", price.Input)
	}
}

func TestProviderGeminiPrefix(t *testing.T) {
	price := Get("gemini-2.5-flash")
	if price.Input != 0.15 {
		t.Errorf("expected input 0.15 for gemini-2.5-flash, got %.2f", price.Input)
	}
}

func TestProviderO3(t *testing.T) {
	price := Get("o3")
	if price.Input != 10.00 {
		t.Errorf("expected input 10.00 for o3, got %.2f", price.Input)
	}
}

func TestProviderO4Mini(t *testing.T) {
	price := Get("o4-mini")
	if price.Input != 1.10 {
		t.Errorf("expected input 1.10 for o4-mini, got %.2f", price.Input)
	}
}

func TestProviderDeepSeek(t *testing.T) {
	price := Get("deepseek-chat")
	if price.Input != 0.27 {
		t.Errorf("expected input 0.27 for deepseek-chat, got %.2f", price.Input)
	}
}

func TestProviderForModel_Anthropic(t *testing.T) {
	prov := ProviderForModel("claude-sonnet-4-6")
	if prov != "anthropic" {
		t.Errorf("expected 'anthropic', got %q", prov)
	}
}

func TestProviderForModel_OpenAI(t *testing.T) {
	prov := ProviderForModel("gpt-4o")
	if prov != "openai" {
		t.Errorf("expected 'openai', got %q", prov)
	}
}

func TestProviderForModel_Gemini(t *testing.T) {
	prov := ProviderForModel("gemini-2.5-flash")
	if prov != "google" {
		t.Errorf("expected 'google', got %q", prov)
	}
}

func TestProviderForModel_Fallback(t *testing.T) {
	prov := ProviderForModel("totally-unknown-model")
	if prov == "" {
		t.Errorf("expected non-empty provider for unknown model")
	}
}

func TestListProviders(t *testing.T) {
	providers := ListProviders()
	if len(providers) < 3 {
		t.Errorf("expected at least 3 providers, got %d: %v", len(providers), providers)
	}
	foundAnthropic := false
	for _, p := range providers {
		if p == "anthropic" {
			foundAnthropic = true
		}
	}
	if !foundAnthropic {
		t.Errorf("expected 'anthropic' in providers list, got %v", providers)
	}
}

func TestCalculateCostOpenAI(t *testing.T) {
	usage := adapter.Usage{
		InputTokens:  1000000,
		OutputTokens: 200000,
	}
	cost := DefaultPricer{}.Calculate(usage, "gpt-4o")
	expected := 2.50 + 2.00
	if cost != expected {
		t.Errorf("expected cost %.2f, got %.2f", expected, cost)
	}
}

func TestCalculateCostGemini(t *testing.T) {
	usage := adapter.Usage{
		InputTokens:  2000000,
		OutputTokens: 100000,
	}
	cost := DefaultPricer{}.Calculate(usage, "gemini-2.5-pro")
	expected := 2.50 + 1.00
	if cost != expected {
		t.Errorf("expected cost %.2f, got %.2f", expected, cost)
	}
}

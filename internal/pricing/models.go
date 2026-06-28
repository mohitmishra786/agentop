package pricing

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/agentop-dev/agentop/internal/adapter"
)

// ModelPrice holds per-model token pricing in USD per million tokens.
type ModelPrice struct {
	Input       float64 `json:"input"`
	Output      float64 `json:"output"`
	CacheCreate float64 `json:"cacheCreate"`
	CacheRead   float64 `json:"cacheRead"`
}

// ProviderPricing holds pricing data for a single provider.
type ProviderPricing struct {
	Fallback string                `json:"fallback"`
	Models   map[string]ModelPrice `json:"models"`
}

// DB is the root pricing configuration loaded from pricing.json.
type DB struct {
	Version          string                     `json:"version"`
	FallbackModel    string                     `json:"FallbackModel"`
	FallbackProvider string                     `json:"FallbackProvider"`
	Providers        map[string]ProviderPricing `json:"providers"`
}

var pdb *DB

func init() {
	pdb = &DB{}
	if err := json.Unmarshal(embeddedPricing, pdb); err != nil {
		panic("failed to parse embedded pricing: " + err.Error())
	}
}

// Get returns the pricing for the given model, falling back to defaults.
func Get(model string) ModelPrice {
	modelLower := strings.ToLower(model)

	for _, provider := range pdb.Providers {
		if p, ok := provider.Models[modelLower]; ok {
			return p
		}
	}

	for _, provider := range pdb.Providers {
		bestMatch := ""
		for name := range provider.Models {
			nl := strings.ToLower(name)
			if strings.HasPrefix(modelLower, nl) {
				if len(nl) > len(bestMatch) {
					bestMatch = name
				}
			}
		}
		if bestMatch != "" {
			return provider.Models[bestMatch]
		}
	}

	return getFallbackPrice()
}

func getFallbackPrice() ModelPrice {
	fb, ok := pdb.Providers[pdb.FallbackProvider]
	if ok {
		if p, ok := fb.Models[pdb.FallbackModel]; ok {
			return p
		}
		if fb.Fallback != "" {
			if p, ok := fb.Models[fb.Fallback]; ok {
				return p
			}
		}
	}
	for _, provider := range pdb.Providers {
		for _, price := range provider.Models {
			return price
		}
	}
	return ModelPrice{}
}

// ProviderForModel returns the provider name for a given model string.
func ProviderForModel(model string) string {
	modelLower := strings.ToLower(model)

	for providerName, provider := range pdb.Providers {
		for name := range provider.Models {
			if strings.ToLower(name) == modelLower {
				return providerName
			}
		}
	}

	providers := make([]string, 0, len(pdb.Providers))
	for providerName, provider := range pdb.Providers {
		for name := range provider.Models {
			nl := strings.ToLower(name)
			if strings.HasPrefix(modelLower, nl) {
				providers = append(providers, providerName)
				break
			}
		}
	}
	if len(providers) > 0 {
		sort.Strings(providers)
		return providers[0]
	}

	return pdb.FallbackProvider
}

// ListProviders returns all known provider names.
func ListProviders() []string {
	providers := make([]string, 0, len(pdb.Providers))
	for name := range pdb.Providers {
		providers = append(providers, name)
	}
	sort.Strings(providers)
	return providers
}

// GetDB returns the global pricing database.
func GetDB() *DB {
	return pdb
}

// Pricer calculates the cost for a given usage and model.
type Pricer interface {
	Calculate(u adapter.Usage, model string) float64
}

// DefaultPricer implements Pricer using the embedded pricing database.
type DefaultPricer struct{}

// Calculate computes the cost for a given usage and model.
func (DefaultPricer) Calculate(u adapter.Usage, model string) float64 {
	p := Get(model)
	return float64(u.InputTokens)*p.Input/1e6 +
		float64(u.OutputTokens)*p.Output/1e6 +
		float64(u.CacheCreationInputTokens)*p.CacheCreate/1e6 +
		float64(u.CacheReadInputTokens)*p.CacheRead/1e6
}

var _ Pricer = DefaultPricer{}

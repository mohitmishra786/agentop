package pricing

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/agentop-dev/agentop/internal/adapter"
)

type ModelPrice struct {
	Input       float64 `json:"input"`
	Output      float64 `json:"output"`
	CacheCreate float64 `json:"cacheCreate"`
	CacheRead   float64 `json:"cacheRead"`
}

type ProviderPricing struct {
	Fallback string                `json:"fallback"`
	Models   map[string]ModelPrice `json:"models"`
}

type PricingDB struct {
	Version          string                     `json:"version"`
	FallbackModel    string                     `json:"FallbackModel"`
	FallbackProvider string                     `json:"FallbackProvider"`
	Providers        map[string]ProviderPricing `json:"providers"`
}

var db *PricingDB

func init() {
	db = &PricingDB{}
	if err := json.Unmarshal(embeddedPricing, db); err != nil {
		panic("failed to parse embedded pricing: " + err.Error())
	}
}

func Get(model string) ModelPrice {
	modelLower := strings.ToLower(model)

	for _, provider := range db.Providers {
		if p, ok := provider.Models[modelLower]; ok {
			return p
		}
	}

	for _, provider := range db.Providers {
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
	fb, ok := db.Providers[db.FallbackProvider]
	if ok {
		if p, ok := fb.Models[db.FallbackModel]; ok {
			return p
		}
		if fb.Fallback != "" {
			if p, ok := fb.Models[fb.Fallback]; ok {
				return p
			}
		}
	}
	for _, provider := range db.Providers {
		for _, price := range provider.Models {
			return price
		}
	}
	return ModelPrice{}
}

func ProviderForModel(model string) string {
	modelLower := strings.ToLower(model)

	for providerName, provider := range db.Providers {
		for name := range provider.Models {
			if strings.ToLower(name) == modelLower {
				return providerName
			}
		}
	}

	providers := make([]string, 0, len(db.Providers))
	for providerName, provider := range db.Providers {
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

	return db.FallbackProvider
}

func ListProviders() []string {
	providers := make([]string, 0, len(db.Providers))
	for name := range db.Providers {
		providers = append(providers, name)
	}
	sort.Strings(providers)
	return providers
}

func GetDB() *PricingDB {
	return db
}

type Pricer interface {
	Calculate(u adapter.Usage, model string) float64
}

type DefaultPricer struct{}

func (DefaultPricer) Calculate(u adapter.Usage, model string) float64 {
	p := Get(model)
	return float64(u.InputTokens)*p.Input/1e6 +
		float64(u.OutputTokens)*p.Output/1e6 +
		float64(u.CacheCreationInputTokens)*p.CacheCreate/1e6 +
		float64(u.CacheReadInputTokens)*p.CacheRead/1e6
}

var _ Pricer = DefaultPricer{}

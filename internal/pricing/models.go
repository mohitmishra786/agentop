package pricing

import (
	"encoding/json"
	"strings"

	"github.com/agentop-dev/agentop/internal/claude"
)

type ModelPrice struct {
	Input       float64 `json:"input"`
	Output      float64 `json:"output"`
	CacheCreate float64 `json:"cacheCreate"`
	CacheRead   float64 `json:"cacheRead"`
}

type PricingDB struct {
	Version string                `json:"version"`
	Models  map[string]ModelPrice `json:"models"`
}

var db *PricingDB

func init() {
	db = &PricingDB{}
	if err := json.Unmarshal(embeddedPricing, db); err != nil {
		panic("failed to parse embedded pricing: " + err.Error())
	}
}

func Get(model string) ModelPrice {
	if p, ok := db.Models[model]; ok {
		return p
	}

	modelLower := strings.ToLower(model)
	for name, price := range db.Models {
		if strings.HasPrefix(modelLower, strings.ToLower(name)) {
			return price
		}
	}

	return db.Models["claude-sonnet-4-6"]
}

func GetDB() *PricingDB {
	return db
}

type Pricer interface {
	Calculate(u claude.Usage, model string) float64
}

type DefaultPricer struct{}

func (DefaultPricer) Calculate(u claude.Usage, model string) float64 {
	p := Get(model)
	return float64(u.InputTokens)*p.Input/1e6 +
		float64(u.OutputTokens)*p.Output/1e6 +
		float64(u.CacheCreationInputTokens)*p.CacheCreate/1e6 +
		float64(u.CacheReadInputTokens)*p.CacheRead/1e6
}

var _ Pricer = DefaultPricer{}

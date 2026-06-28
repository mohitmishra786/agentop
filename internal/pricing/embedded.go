// Package pricing provides model-specific token pricing for cost calculation.
package pricing

import _ "embed"

//go:embed pricing.json
var embeddedPricing []byte

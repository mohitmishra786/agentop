package opencode

type openCodeSession struct {
	ID        string
	Directory string
	CreatedAt string
}

type openCodeMessage struct {
	ID        string
	SessionID string
	Role      string
	Content   string
	Data      string
	CreatedAt string
}

type messageData struct {
	Tokens     *tokensData `json:"tokens"`
	Cost       float64     `json:"cost"`
	ModelID    string      `json:"modelID"`
	ProviderID string      `json:"providerID"`
}

type tokensData struct {
	Total     int        `json:"total"`
	Input     int        `json:"input"`
	Output    int        `json:"output"`
	Reasoning int        `json:"reasoning"`
	Cache     *cacheData `json:"cache"`
}

type cacheData struct {
	Read  int `json:"read"`
	Write int `json:"write"`
}

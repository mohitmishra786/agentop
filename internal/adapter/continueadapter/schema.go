package continueadapter

type continueSession struct {
	SessionID          string            `json:"sessionId"`
	Title              string            `json:"title,omitempty"`
	WorkspaceDirectory string            `json:"workspaceDirectory,omitempty"`
	Messages           []continueMessage `json:"messages"`
	Usage              *continueUsage    `json:"usage,omitempty"`
}

type continueMessage struct {
	Role    string `json:"role"`
	Content string `json:"content,omitempty"`
	Model   string `json:"model,omitempty"`
}

type continueUsage struct {
	PromptTokens     int     `json:"promptTokens"`
	CompletionTokens int     `json:"completionTokens"`
	CachedTokens     int     `json:"cachedTokens,omitempty"`
	CacheWriteTokens int     `json:"cacheWriteTokens,omitempty"`
	TotalCost        float64 `json:"totalCost,omitempty"`
}

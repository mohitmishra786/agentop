package cursor

type cursorEntry struct {
	Bubbles []cursorBubble `json:"bubbles"`
}

type cursorBubble struct {
	Text       string            `json:"text"`
	Role       string            `json:"role,omitempty"`
	TokenCount *cursorTokenCount `json:"tokenCount,omitempty"`
	ModelInfo  *cursorModelInfo  `json:"modelInfo,omitempty"`
}

type cursorTokenCount struct {
	InputTokens  int `json:"inputTokens"`
	OutputTokens int `json:"outputTokens"`
}

type cursorModelInfo struct {
	ModelName string `json:"modelName"`
}

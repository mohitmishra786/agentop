package codex

import "encoding/json"

type codexEvent struct {
	Timestamp string          `json:"timestamp"`
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
}

type sessionMetaPayload struct {
	SessionID string      `json:"id"`
	CWD       string      `json:"cwd"`
	Source    *sourceInfo `json:"source,omitempty"`
}

type sourceInfo struct {
	Subagent bool `json:"subagent,omitempty"`
}

type turnContextPayload struct {
	Model      string `json:"model"`
	TurnNumber int    `json:"turn_number"`
}

type eventMsgPayload struct {
	Type   string          `json:"type"`
	Info   *tokenInfo      `json:"info,omitempty"`
	Role   string          `json:"role,omitempty"`
	Status string          `json:"status,omitempty"`
	Data   json.RawMessage `json:"data,omitempty"`
}

type tokenInfo struct {
	LastTokenUsage *tokenUsage `json:"last_token_usage,omitempty"`
}

type tokenUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type responseItemPayload struct {
	Type    string          `json:"type"`
	Role    string          `json:"role,omitempty"`
	Status  string          `json:"status,omitempty"`
	CallID  string          `json:"call_id,omitempty"`
	Name    string          `json:"name,omitempty"`
	Content json.RawMessage `json:"content,omitempty"`
}

type configSnapshotPayload struct {
	Model string `json:"model,omitempty"`
}

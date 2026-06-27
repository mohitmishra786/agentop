package copilot

import "encoding/json"

type copilotEvent struct {
	Type      string          `json:"type"`
	Data      json.RawMessage `json:"data"`
	Timestamp string          `json:"timestamp,omitempty"`
}

type sessionStartData struct {
	Model      string `json:"model"`
	CWD        string `json:"cwd"`
	Repository string `json:"repository"`
}

type sessionModelChangeData struct {
	Model         string `json:"model"`
	PreviousModel string `json:"previous_model"`
}

type sessionShutdownData struct {
	ModelMetrics map[string]modelMetric `json:"modelMetrics"`
}

type modelMetric struct {
	Usage modelUsage `json:"usage"`
}

type modelUsage struct {
	InputTokens  int `json:"inputTokens"`
	OutputTokens int `json:"outputTokens"`
}

type toolExecutionData struct {
	ToolName string          `json:"toolName"`
	Args     json.RawMessage `json:"args,omitempty"`
	Result   json.RawMessage `json:"result,omitempty"`
}

type messageData struct {
	Content string `json:"content"`
}

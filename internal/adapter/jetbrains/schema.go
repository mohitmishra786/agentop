package jetbrains

import "encoding/json"

type jetbrainsEvent struct {
	Type      string          `json:"type"`
	Data      json.RawMessage `json:"data"`
	Timestamp string          `json:"timestamp"`
}

type sessionMetaData struct {
	SessionID     string `json:"sessionId"`
	WorkspacePath string `json:"workspacePath"`
	Model         string `json:"model,omitempty"`
	IDE           string `json:"ide,omitempty"`
}

type messageRenderedData struct {
	Message       string `json:"message"`
	Model         string `json:"model,omitempty"`
	WorkspacePath string `json:"workspacePath,omitempty"`
	IDE           string `json:"ide,omitempty"`
}

type assistantMessageData struct {
	Content    string          `json:"content"`
	Model      string          `json:"model,omitempty"`
	MessageID  string          `json:"messageId,omitempty"`
	ToolCallID string          `json:"toolCallId,omitempty"`
	ToolCalls  json.RawMessage `json:"toolCalls,omitempty"`
}

type toolExecutionData struct {
	ToolName   string          `json:"toolName"`
	Args       json.RawMessage `json:"args,omitempty"`
	Result     json.RawMessage `json:"result,omitempty"`
	ToolCallID string          `json:"toolCallId,omitempty"`
}

package kiro

import "encoding/json"

type kiroMeta struct {
	CWD    string `json:"cwd"`
	Title  string `json:"title"`
	Model  string `json:"model"`
	Parent string `json:"parent_session_id"`
	Time   string `json:"timestamp"`
}

type kiroEvent struct {
	Kind      string          `json:"kind"`
	Content   string          `json:"content,omitempty"`
	ToolName  string          `json:"toolName,omitempty"`
	ToolInput json.RawMessage `json:"toolInput,omitempty"`
	Timestamp string          `json:"timestamp,omitempty"`
}

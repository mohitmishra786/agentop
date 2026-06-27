package claude

import (
	"encoding/json"
	"time"
)

type rawEvent struct {
	Type             string          `json:"type"`
	Message          *rawMessage     `json:"message,omitempty"`
	ToolName         string          `json:"toolName,omitempty"`
	ToolInput        json.RawMessage `json:"toolInput,omitempty"`
	ToolResult       json.RawMessage `json:"toolResult,omitempty"`
	Summary          string          `json:"summary,omitempty"`
	SessionID        string          `json:"sessionId"`
	UUID             string          `json:"uuid"`
	ParentUUID       string          `json:"parentUuid,omitempty"`
	Timestamp        time.Time       `json:"timestamp"`
	CostUSD          float64         `json:"costUSD,omitempty"`
	CWD              string          `json:"cwd,omitempty"`
	IsSidechain      bool            `json:"isSidechain,omitempty"`
	MessageID        string          `json:"messageId,omitempty"`
	PromptID         string          `json:"promptId,omitempty"`
	Version          string          `json:"version,omitempty"`
	Entrypoint       string          `json:"entrypoint,omitempty"`
	UserType         string          `json:"userType,omitempty"`
	PermissionMode   string          `json:"permissionMode,omitempty"`
	GitBranch        string          `json:"gitBranch,omitempty"`
	IsSnapshotUpdate bool            `json:"isSnapshotUpdate,omitempty"`
	Snapshot         *fileSnapshot   `json:"snapshot,omitempty"`
	DurationMs       int64           `json:"durationMs,omitempty"`
	Subtype          string          `json:"subtype,omitempty"`
}

type rawMessage struct {
	ID           string       `json:"id"`
	Type         string       `json:"type,omitempty"`
	Role         string       `json:"role"`
	Model        string       `json:"model,omitempty"`
	Usage        *rawUsage    `json:"usage,omitempty"`
	Content      []rawContent `json:"content,omitempty"`
	StopReason   string       `json:"stop_reason,omitempty"`
	StopSequence string       `json:"stop_sequence,omitempty"`
}

type rawUsage struct {
	InputTokens              int            `json:"input_tokens"`
	OutputTokens             int            `json:"output_tokens"`
	CacheCreationInputTokens int            `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int            `json:"cache_read_input_tokens"`
	CacheCreation            *rawCacheTiers `json:"cache_creation,omitempty"`
	ServiceTier              string         `json:"service_tier,omitempty"`
}

type rawCacheTiers struct {
	Ephemeral5m int `json:"ephemeral_5m_input_tokens"`
	Ephemeral1h int `json:"ephemeral_1h_input_tokens"`
}

type rawContent struct {
	Type      string `json:"type"`
	Text      string `json:"text,omitempty"`
	Thinking  string `json:"thinking,omitempty"`
	Signature string `json:"signature,omitempty"`
}

type fileSnapshot struct {
	MessageID          string            `json:"messageId"`
	TrackedFileBackups map[string]string `json:"trackedFileBackups"`
	Timestamp          time.Time         `json:"timestamp"`
}

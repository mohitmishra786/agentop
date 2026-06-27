// Package adapter provides common types and interfaces for AI coding assistant
// session adapters. Each supported agent (Claude Code, Codex CLI, etc.)
// implements the Adapter interface to provide session discovery and parsing.
package adapter

import (
	"encoding/json"
	"time"
)

// AgentID identifies a supported AI coding assistant.
type AgentID string

const (
	AgentClaude    AgentID = "claude"
	AgentCodex     AgentID = "codex"
	AgentKiro      AgentID = "kiro"
	AgentCopilot   AgentID = "copilot"
	AgentCursor    AgentID = "cursor"
	AgentGemini    AgentID = "gemini"
	AgentOpenCode  AgentID = "opencode"
	AgentJetBrains AgentID = "jetbrains"
	AgentContinue  AgentID = "continue"
	AgentWindsurf  AgentID = "windsurf"
	AgentDevin     AgentID = "devin"
)

// TokenSource indicates whether token counts come directly from the agent's data
// or are estimated from other fields (e.g. character count, duration).
type TokenSource int

const (
	TokenExact     TokenSource = iota // exact token counts from API response
	TokenEstimated                    // estimated from other fields
)

func (t TokenSource) String() string {
	switch t {
	case TokenExact:
		return "exact"
	case TokenEstimated:
		return "estimated"
	default:
		return "unknown"
	}
}

// Usage holds per-message token counts returned by the API.
type Usage struct {
	InputTokens              int         `json:"input_tokens"`
	OutputTokens             int         `json:"output_tokens"`
	CacheCreationInputTokens int         `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int         `json:"cache_read_input_tokens"`
	CacheCreation            *CacheTiers `json:"cache_creation,omitempty"`
	ServiceTier              string      `json:"service_tier,omitempty"`
}

// CacheTiers breaks down cache-creation tokens by TTL bucket.
type CacheTiers struct {
	Ephemeral5m int `json:"ephemeral_5m_input_tokens"`
	Ephemeral1h int `json:"ephemeral_1h_input_tokens"`
}

// Content holds a single content block within a message.
type Content struct {
	Type      string `json:"type"`
	Text      string `json:"text,omitempty"`
	Thinking  string `json:"thinking,omitempty"`
	Signature string `json:"signature,omitempty"`
}

// Event is a single event in an AI coding assistant session. It normalises
// the various agent-specific JSONL/JSON formats into a common structure.
type Event struct {
	Type             string          `json:"type"`
	Message          *Message        `json:"message,omitempty"`
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
	Snapshot         *FileSnapshot   `json:"snapshot,omitempty"`
	DurationMs       int64           `json:"durationMs,omitempty"`
	Subtype          string          `json:"subtype,omitempty"`

	AgentID AgentID `json:"agentId,omitempty"`

	TokenSrc TokenSource `json:"tokenSource,omitempty"`
}

// Message holds a single assistant message within an event.
type Message struct {
	ID           string    `json:"id"`
	Type         string    `json:"type,omitempty"`
	Role         string    `json:"role"`
	Model        string    `json:"model,omitempty"`
	Usage        *Usage    `json:"usage,omitempty"`
	Content      []Content `json:"content,omitempty"`
	StopReason   string    `json:"stop_reason,omitempty"`
	StopSequence string    `json:"stop_sequence,omitempty"`
}

// FileSnapshot captures the state of tracked files at a point in time.
type FileSnapshot struct {
	MessageID          string            `json:"messageId"`
	TrackedFileBackups map[string]string `json:"trackedFileBackups"`
	Timestamp          time.Time         `json:"timestamp"`
}

// SessionMeta holds optional metadata for a session.
type SessionMeta struct {
	ID               string    `json:"id"`
	Summary          string    `json:"summary"`
	FirstUserMessage string    `json:"firstUserMessage"`
	MessageCount     int       `json:"messageCount"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
	GitBranch        string    `json:"gitBranch"`
	CWD              string    `json:"cwd"`
	ProjectPath      string    `json:"projectPath"`
}

// SessionFile represents a discovered session file on disk.
type SessionFile struct {
	Path          string
	ProjectHash   string
	SessionID     string
	ModTime       time.Time
	SubagentFiles []string
	AgentID       AgentID
}

// ParseResult wraps the output of Adapter.ParseSession.
type ParseResult struct {
	Events []Event
	Meta   *SessionMeta
}

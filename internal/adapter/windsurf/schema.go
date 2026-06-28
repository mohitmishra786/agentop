// Package windsurf implements an adapter for Windsurf/Devin Desktop session files
// (.pb encrypted protobuf format). Since actual AES-256-GCM decryption requires
// extracting the key from the language_server binary, this adapter uses a
// best-effort estimation approach based on file metadata.
package windsurf

import (
	"encoding/json"
	"time"
)

// StepType represents the type of a CortexTrajectory step.
type StepType int32

const (
	StepTypeUnknown       StepType = 0
	StepTypeUserMessage   StepType = 14
	StepTypeAssistantResp StepType = 15
	StepTypeToolExecution StepType = 21
)

func (s StepType) String() string {
	switch s {
	case StepTypeUserMessage:
		return "user_message"
	case StepTypeAssistantResp:
		return "assistant_response"
	case StepTypeToolExecution:
		return "tool_execution"
	default:
		return "unknown"
	}
}

// CortexTrajectory is the top-level protobuf message stored in .pb files.
type CortexTrajectory struct {
	Steps    []*CortexStep   `json:"steps,omitempty"`
	Metadata *TrajectoryMeta `json:"metadata,omitempty"`
}

// TrajectoryMeta holds session-level metadata.
type TrajectoryMeta struct {
	WorkspacePath string    `json:"workspacePath,omitempty"`
	StartTime     time.Time `json:"startTime,omitempty"`
	EndTime       time.Time `json:"endTime,omitempty"`
	Model         string    `json:"model,omitempty"`
	Provider      string    `json:"provider,omitempty"`
	GitBranch     string    `json:"gitBranch,omitempty"`
	SessionID     string    `json:"sessionId,omitempty"`
}

// CortexStep is a single step within a trajectory.
type CortexStep struct {
	Type       StepType        `json:"type,omitempty"`
	ID         string          `json:"id,omitempty"`
	Content    string          `json:"content,omitempty"`
	Timestamp  time.Time       `json:"timestamp,omitempty"`
	ToolName   string          `json:"toolName,omitempty"`
	ToolInput  json.RawMessage `json:"toolInput,omitempty"`
	ToolResult json.RawMessage `json:"toolResult,omitempty"`
	Model      string          `json:"model,omitempty"`
	Provider   string          `json:"provider,omitempty"`
	IsThinking bool            `json:"isThinking,omitempty"`
}

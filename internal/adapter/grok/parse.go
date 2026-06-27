package grok

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/agentop-dev/agentop/internal/adapter"
)

type summaryData struct {
	Info struct {
		ID  string `json:"id"`
		CWD string `json:"cwd"`
	} `json:"info"`
	SessionSummary   string `json:"session_summary"`
	GeneratedTitle   string `json:"generated_title"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
	CurrentModelID   string `json:"current_model_id"`
	GitRootDir       string `json:"git_root_dir"`
	HeadBranch       string `json:"head_branch"`
}

type signalsData struct {
	TurnCount             int      `json:"turnCount"`
	UserMessageCount      int      `json:"userMessageCount"`
	AssistantMessageCount int      `json:"assistantMessageCount"`
	ContextTokensUsed     int      `json:"contextTokensUsed"`
	ContextWindowTokens   int      `json:"contextWindowTokens"`
	ToolCallCount         int      `json:"toolCallCount"`
	ToolsUsed             []string `json:"toolsUsed"`
	ModelsUsed            []string `json:"modelsUsed"`
	PrimaryModelID        string   `json:"primaryModelId"`
}

func parseSession(sessionPath string) (*adapter.ParseResult, error) {
	summaryPath := filepath.Join(sessionPath, "summary.json")
	signalsPath := filepath.Join(sessionPath, "signals.json")

	var summary summaryData
	if err := readJSON(summaryPath, &summary); err != nil {
		return nil, err
	}

	var signals signalsData
	if err := readJSON(signalsPath, &signals); err != nil {
		signals = signalsData{}
	}

	createdAt := parseTime(summary.CreatedAt)
	updatedAt := parseTime(summary.UpdatedAt)
	if updatedAt.IsZero() {
		updatedAt = createdAt
	}

	model := ""
	if summary.CurrentModelID != "" {
		model = summary.CurrentModelID
	} else if signals.PrimaryModelID != "" {
		model = signals.PrimaryModelID
	}

	turns := signals.TurnCount
	if turns < 1 {
		turns = 1
	}

	totalTokens := signals.ContextTokensUsed
	if totalTokens < 1 {
		totalTokens = 1000
	}

	inputPerTurn := totalTokens / turns
	outputPerTurn := inputPerTurn / 4
	if outputPerTurn < 1 {
		outputPerTurn = 1
	}

	var events []adapter.Event
	eventTime := createdAt
	for i := 0; i < turns; i++ {
		userID := summary.Info.ID + "-user-" + itoa(i)
		events = append(events, adapter.Event{
			Type:      "user",
			SessionID: summary.Info.ID,
			UUID:      userID,
			Timestamp: eventTime,
			Message: &adapter.Message{
				ID:   userID,
				Role: "user",
			},
			AgentID:  adapter.AgentGrok,
			TokenSrc: adapter.TokenEstimated,
		})
		eventTime = eventTime.Add(time.Second)

		asstID := summary.Info.ID + "-asst-" + itoa(i)
		events = append(events, adapter.Event{
			Type:      "assistant",
			SessionID: summary.Info.ID,
			UUID:      asstID,
			Timestamp: eventTime,
			Message: &adapter.Message{
				ID:         asstID,
				Role:       "assistant",
				Model:      model,
				StopReason: "end_turn",
				Usage: &adapter.Usage{
					InputTokens:  inputPerTurn,
					OutputTokens: outputPerTurn,
				},
			},
			AgentID:  adapter.AgentGrok,
			TokenSrc: adapter.TokenEstimated,
		})
		eventTime = eventTime.Add(time.Second)
	}

	title := summary.GeneratedTitle
	if title == "" {
		title = summary.SessionSummary
	}

	meta := &adapter.SessionMeta{
		ID:        summary.Info.ID,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		CWD:       summary.Info.CWD,
		Summary:   title,
	}

	return &adapter.ParseResult{Events: events, Meta: meta}, nil
}

func readJSON(path string, v interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func parseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err == nil {
		return t
	}
	t, err = time.Parse(time.RFC3339Nano, s)
	if err == nil {
		return t
	}
	return time.Time{}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}

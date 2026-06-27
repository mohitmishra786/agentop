package jetbrains

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/agentop-dev/agentop/internal/adapter"
)

const maxLineBytes = 10 * 1024 * 1024

func parseSessionFile(path string) (*adapter.ParseResult, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, maxLineBytes), maxLineBytes)

	var rawEvents []jetbrainsEvent
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		var e jetbrainsEvent
		if err := json.Unmarshal(line, &e); err != nil {
			continue
		}
		rawEvents = append(rawEvents, e)
	}
	if err := scanner.Err(); err != nil {
		if errors.Is(err, bufio.ErrTooLong) {
			return nil, nil
		}
		return nil, err
	}

	var events []adapter.Event
	var meta *adapter.SessionMeta
	var cwd, currentModel, sessionID string
	var firstEventTime time.Time

	dir := filepath.Dir(path)
	sessionID = filepath.Base(dir)

	for _, re := range rawEvents {
		ts := parseTimestamp(re.Timestamp)
		if firstEventTime.IsZero() {
			firstEventTime = ts
		}

		switch re.Type {
		case "session.start":
			var d sessionMetaData
			if json.Unmarshal(re.Data, &d) == nil {
				if d.SessionID != "" {
					sessionID = d.SessionID
				}
				if d.WorkspacePath != "" {
					cwd = d.WorkspacePath
				}
				if d.Model != "" {
					currentModel = d.Model
				}
			}

		case "user.message_rendered":
			var d messageRenderedData
			_ = json.Unmarshal(re.Data, &d)
			if d.Model != "" {
				currentModel = d.Model
			}
			if d.WorkspacePath != "" {
				cwd = d.WorkspacePath
			}
			events = append(events, adapter.Event{
				Type:      "user",
				SessionID: sessionID,
				Timestamp: ts,
				Summary:   d.Message,
				CWD:       cwd,
				Message: &adapter.Message{
					Model: currentModel,
					Usage: &adapter.Usage{
						InputTokens: estimateTokens(d.Message),
					},
				},
			})

		case "assistant.message":
			var d assistantMessageData
			_ = json.Unmarshal(re.Data, &d)
			if d.Model != "" {
				currentModel = d.Model
			} else if d.ToolCallID != "" && currentModel == "" {
				currentModel = inferModelFromCallID(d.ToolCallID)
			}
			events = append(events, adapter.Event{
				Type:      "assistant",
				SessionID: sessionID,
				Timestamp: ts,
				Message: &adapter.Message{
					Model: currentModel,
					Usage: &adapter.Usage{
						OutputTokens: estimateTokens(d.Content),
					},
				},
				Summary: d.Content,
			})

		case "tool.execution_start":
			var d toolExecutionData
			if json.Unmarshal(re.Data, &d) == nil {
				if d.ToolCallID != "" && currentModel == "" {
					currentModel = inferModelFromCallID(d.ToolCallID)
				}
				events = append(events, adapter.Event{
					Type:      "tool",
					SessionID: sessionID,
					Timestamp: ts,
					ToolName:  d.ToolName,
					ToolInput: d.Args,
				})
			}

		case "tool.execution_complete":
			var d toolExecutionData
			if json.Unmarshal(re.Data, &d) == nil {
				events = append(events, adapter.Event{
					Type:       "tool_result",
					SessionID:  sessionID,
					Timestamp:  ts,
					ToolName:   d.ToolName,
					ToolResult: d.Result,
				})
			}
		}
	}

	if sessionID != "" {
		meta = &adapter.SessionMeta{
			ID:        sessionID,
			CWD:       cwd,
			CreatedAt: firstEventTime,
		}
	}

	for i := range events {
		events[i].AgentID = adapter.AgentJetBrains
		events[i].TokenSrc = adapter.TokenEstimated
	}

	return &adapter.ParseResult{Events: events, Meta: meta}, nil
}

func parseTimestamp(s string) time.Time {
	if s == "" {
		return time.Now()
	}
	t, err := time.Parse(time.RFC3339, s)
	if err == nil {
		return t
	}
	t, err = time.Parse(time.RFC3339Nano, s)
	if err == nil {
		return t
	}
	return time.Now()
}

func estimateTokens(s string) int {
	return len(s) / 4
}

func inferModelFromCallID(callID string) string {
	if strings.HasPrefix(callID, "toolu_vrtx_") {
		return "claude-sonnet-4-20250514"
	}
	if strings.HasPrefix(callID, "call_") {
		return "gpt-4o-2025-05-14"
	}
	return ""
}

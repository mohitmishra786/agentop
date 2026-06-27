package copilot

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
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

	var rawEvents []copilotEvent
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		var e copilotEvent
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
	var cwd, currentModel string
	var firstEventTime time.Time

	// Session ID is the UUID directory name containing events.jsonl
	dir := filepath.Dir(path)
	sessionID := filepath.Base(dir)

	for _, re := range rawEvents {
		ts := parseTimestamp(re.Timestamp)
		if firstEventTime.IsZero() {
			firstEventTime = ts
		}

		switch re.Type {
		case "session.start":
			var d sessionStartData
			if json.Unmarshal(re.Data, &d) == nil {
				cwd = d.CWD
				if d.Model != "" {
					currentModel = d.Model
				}
			}

		case "session.model_change":
			var d sessionModelChangeData
			if json.Unmarshal(re.Data, &d) == nil && d.Model != "" {
				currentModel = d.Model
			}

		case "session.shutdown":
			var d sessionShutdownData
			if json.Unmarshal(re.Data, &d) == nil {
				for model, metric := range d.ModelMetrics {
					if metric.Usage.InputTokens > 0 || metric.Usage.OutputTokens > 0 {
						events = append(events, adapter.Event{
							Type:      "assistant",
							SessionID: sessionID,
							Timestamp: ts,
							Message: &adapter.Message{
								Model: model,
								Usage: &adapter.Usage{
									InputTokens:  metric.Usage.InputTokens,
									OutputTokens: metric.Usage.OutputTokens,
								},
							},
						})
					}
				}
			}

		case "user.message":
			var d messageData
			_ = json.Unmarshal(re.Data, &d)
			events = append(events, adapter.Event{
				Type:      "user",
				SessionID: sessionID,
				Timestamp: ts,
				Summary:   d.Content,
				CWD:       cwd,
			})

		case "assistant.message":
			var d messageData
			_ = json.Unmarshal(re.Data, &d)
			events = append(events, adapter.Event{
				Type:      "assistant",
				SessionID: sessionID,
				Timestamp: ts,
				Message: &adapter.Message{
					Model: currentModel,
				},
				Summary: d.Content,
			})

		case "tool.execution_start":
			var d toolExecutionData
			if json.Unmarshal(re.Data, &d) == nil {
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
		events[i].AgentID = adapter.AgentCopilot
		events[i].TokenSrc = adapter.TokenExact
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

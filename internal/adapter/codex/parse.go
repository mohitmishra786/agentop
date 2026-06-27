package codex

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"sort"
	"time"

	"github.com/agentop-dev/agentop/internal/adapter"
)

const maxLineBytes = 10 * 1024 * 1024

func parseSessionFile(path string) (*adapter.ParseResult, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, maxLineBytes), maxLineBytes)

	var rawEvents []codexEvent
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		var e codexEvent
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

	sort.Slice(rawEvents, func(i, j int) bool {
		return rawEvents[i].Timestamp < rawEvents[j].Timestamp
	})

	var events []adapter.Event
	var meta *adapter.SessionMeta
	var sessionID, cwd, currentModel string
	prevTokens := 0

	for _, re := range rawEvents {
		switch re.Type {
		case "session_meta":
			var p sessionMetaPayload
			if json.Unmarshal(re.Payload, &p) == nil {
				sessionID = p.SessionID
				cwd = p.CWD
			}

		case "turn_context":
			var p turnContextPayload
			if json.Unmarshal(re.Payload, &p) == nil && p.Model != "" {
				currentModel = p.Model
			}

		case "config_snapshot":
			var p configSnapshotPayload
			if json.Unmarshal(re.Payload, &p) == nil && p.Model != "" {
				currentModel = p.Model
			}

		case "response_item":
			var p responseItemPayload
			if json.Unmarshal(re.Payload, &p) != nil {
				continue
			}
			ts := parseTimestamp(re.Timestamp)

			switch p.Type {
			case "assistant_message":
				events = append(events, adapter.Event{
					Type:      "assistant",
					SessionID: sessionID,
					Timestamp: ts,
					Message: &adapter.Message{
						Model:   currentModel,
						Role:    p.Role,
					},
				})
			case "function_call":
				events = append(events, adapter.Event{
					Type:      "tool",
					SessionID: sessionID,
					Timestamp: ts,
					ToolName:  p.Name,
				})
			case "function_call_output":
				var content string
				if p.Content != nil {
					json.Unmarshal(p.Content, &content)
				}
				events = append(events, adapter.Event{
					Type:        "tool_result",
					SessionID:   sessionID,
					Timestamp:   ts,
					ToolResult:  p.Content,
					ToolName:    content,
				})
			}

		case "event_msg":
			var p eventMsgPayload
			if json.Unmarshal(re.Payload, &p) != nil {
				continue
			}
			ts := parseTimestamp(re.Timestamp)

			if p.Type == "token_count" && p.Info != nil && p.Info.LastTokenUsage != nil {
				diffInput := p.Info.LastTokenUsage.InputTokens - prevTokens
				if diffInput < 0 {
					diffInput = p.Info.LastTokenUsage.InputTokens
				}
				prevTokens = p.Info.LastTokenUsage.InputTokens

				events = append(events, adapter.Event{
					Type:      "assistant",
					SessionID: sessionID,
					Timestamp: ts,
					Message: &adapter.Message{
						Model: currentModel,
						Usage: &adapter.Usage{
							InputTokens:  diffInput,
							OutputTokens: p.Info.LastTokenUsage.OutputTokens,
						},
					},
				})
			}

			if p.Type == "input_item" || p.Type == "user_message" {
				events = append(events, adapter.Event{
					Type:      "user",
					SessionID: sessionID,
					Timestamp: ts,
				})
			}

		case "input_item":
			ts := parseTimestamp(re.Timestamp)
			events = append(events, adapter.Event{
				Type:      "user",
				SessionID: sessionID,
				Timestamp: ts,
				CWD:       cwd,
			})
		}
	}

	if sessionID != "" {
		meta = &adapter.SessionMeta{
			ID:  sessionID,
			CWD: cwd,
		}
	}

	for i := range events {
		events[i].AgentID = adapter.AgentCodex
		events[i].TokenSrc = adapter.TokenExact
	}

	return &adapter.ParseResult{Events: events, Meta: meta}, nil
}

func parseTimestamp(s string) time.Time {
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

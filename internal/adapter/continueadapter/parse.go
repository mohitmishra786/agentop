package continueadapter

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/agentop-dev/agentop/internal/adapter"
)

func parseSessionFile(path string) (*adapter.ParseResult, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var sess continueSession
	if err := json.NewDecoder(f).Decode(&sess); err != nil {
		return nil, fmt.Errorf("decode continue session: %w", err)
	}

	if sess.SessionID == "" {
		return nil, fmt.Errorf("continue session missing sessionId")
	}

	var events []adapter.Event
	for i, msg := range sess.Messages {
		eventType := msg.Role
		switch eventType {
		case "assistant", "model":
			eventType = "assistant"
		case "user":
			eventType = "user"
		case "tool":
			eventType = "tool"
		default:
			eventType = "user"
		}

		evt := adapter.Event{
			Type:      eventType,
			SessionID: sess.SessionID,
			Timestamp: time.Now(),
			AgentID:   adapter.AgentContinue,
			TokenSrc:  adapter.TokenExact,
			Message: &adapter.Message{
				ID:   fmt.Sprintf("%s-%d", sess.SessionID, i),
				Role: msg.Role,
			},
		}

		if msg.Model != "" {
			evt.Message.Model = msg.Model
		}

		if msg.Content != "" {
			evt.Message.Content = []adapter.Content{
				{Type: "text", Text: msg.Content},
			}
		}

		events = append(events, evt)
	}

	// Attach session-level usage to the last assistant message.
	if sess.Usage != nil {
		usage := &adapter.Usage{
			InputTokens:              sess.Usage.PromptTokens,
			OutputTokens:             sess.Usage.CompletionTokens,
			CacheReadInputTokens:     sess.Usage.CachedTokens,
			CacheCreationInputTokens: sess.Usage.CacheWriteTokens,
		}
		for i := len(events) - 1; i >= 0; i-- {
			if events[i].Type == "assistant" {
				events[i].Message.Usage = usage
				events[i].CostUSD = sess.Usage.TotalCost
				break
			}
		}
	}

	if events == nil {
		events = []adapter.Event{}
	}

	meta := &adapter.SessionMeta{
		ID:           sess.SessionID,
		Summary:      sess.Title,
		MessageCount: len(sess.Messages),
		ProjectPath:  sess.WorkspaceDirectory,
	}

	return &adapter.ParseResult{Events: events, Meta: meta}, nil
}

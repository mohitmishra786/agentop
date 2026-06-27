package cursor

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/agentop-dev/agentop/internal/adapter"
)

const pathSep = "#"

func dbPathFrom(path string) string {
	if i := strings.LastIndex(path, pathSep); i >= 0 {
		return path[:i]
	}
	return path
}

func sessionIDFrom(path string) string {
	if i := strings.LastIndex(path, pathSep); i >= 0 {
		return path[i+len(pathSep):]
	}
	return ""
}

func parseSessionFile(path string) (*adapter.ParseResult, error) {
	dbPath := dbPathFrom(path)
	composerID := sessionIDFrom(path)
	if composerID == "" {
		return nil, fmt.Errorf("invalid session path: %s", path)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var rawValue []byte
	err = db.QueryRow(
		"SELECT value FROM cursorDiskKV WHERE [key] = ?", "composer."+composerID,
	).Scan(&rawValue)
	if err != nil {
		return nil, fmt.Errorf("query composer %s: %w", composerID, err)
	}

	var entry cursorEntry
	if err := json.Unmarshal(rawValue, &entry); err != nil {
		return nil, fmt.Errorf("decode composer %s: %w", composerID, err)
	}

	var events []adapter.Event
	for i, bubble := range entry.Bubbles {
		role := bubble.Role
		if role == "" {
			if i%2 == 0 {
				role = "user"
			} else {
				role = "assistant"
			}
		}
		if role != "user" && role != "assistant" {
			role = "user"
		}

		msg := &adapter.Message{
			ID:   fmt.Sprintf("%s-%d", composerID, i),
			Role: role,
		}

		if bubble.Text != "" {
			msg.Content = []adapter.Content{
				{Type: "text", Text: bubble.Text},
			}
		}

		tokenSrc := adapter.TokenEstimated
		if bubble.TokenCount != nil && (bubble.TokenCount.InputTokens != 0 || bubble.TokenCount.OutputTokens != 0) {
			msg.Usage = &adapter.Usage{
				InputTokens:  bubble.TokenCount.InputTokens,
				OutputTokens: bubble.TokenCount.OutputTokens,
			}
			tokenSrc = adapter.TokenExact
		} else {
			est := len(bubble.Text) / 4
			if est < 1 {
				est = 1
			}
			msg.Usage = &adapter.Usage{
				InputTokens:  est,
				OutputTokens: est,
			}
		}

		if bubble.ModelInfo != nil && bubble.ModelInfo.ModelName != "" {
			if bubble.ModelInfo.ModelName == "default" {
				msg.Model = "cursor-auto"
			} else {
				msg.Model = bubble.ModelInfo.ModelName
			}
		}

		evt := adapter.Event{
			Type:      role,
			SessionID: composerID,
			UUID:      msg.ID,
			Timestamp: time.Now(),
			Message:   msg,
			AgentID:   adapter.AgentCursor,
			TokenSrc:  tokenSrc,
		}

		events = append(events, evt)
	}

	if events == nil {
		events = []adapter.Event{}
	}

	meta := &adapter.SessionMeta{
		ID: composerID,
	}
	if len(events) > 0 {
		meta.UpdatedAt = events[len(events)-1].Timestamp
	}

	return &adapter.ParseResult{Events: events, Meta: meta}, nil
}

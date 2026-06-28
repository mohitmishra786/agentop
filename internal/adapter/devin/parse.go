package devin

import (
	"database/sql"
	"encoding/json"
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

type chatMetadata struct {
	NumTokens *int `json:"num_tokens,omitempty"`
}

type chatMessage struct {
	MessageID string        `json:"message_id"`
	Role      string        `json:"role"`
	Content   string        `json:"content"`
	Metadata  *chatMetadata `json:"metadata,omitempty"`
}

func parseSession(path string) (*adapter.ParseResult, error) {
	dbPath := dbPathFrom(path)
	sid := sessionIDFrom(path)
	if sid == "" {
		return nil, nil
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = db.Close() }()

	var title, model, workingDir string
	var createdAt int64
	err = db.QueryRow(
		"SELECT COALESCE(title, ''), model, created_at, COALESCE(working_directory, '') FROM sessions WHERE id = ?", sid,
	).Scan(&title, &model, &createdAt, &workingDir)
	if err != nil {
		return nil, err
	}

	createdAtTime := time.Unix(createdAt, 0)

	rows, err := db.Query(
		"SELECT chat_message, created_at FROM message_nodes WHERE session_id = ? ORDER BY row_id ASC",
		sid,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var events []adapter.Event
	for rows.Next() {
		var chatMsgStr string
		var msgCreatedAt int64
		if err := rows.Scan(&chatMsgStr, &msgCreatedAt); err != nil {
			continue
		}

		ts := time.Unix(msgCreatedAt, 0)

		var msg chatMessage
		if chatMsgStr != "" {
			_ = json.Unmarshal([]byte(chatMsgStr), &msg)
		}

		evtType := msg.Role
		if evtType == "" {
			evtType = "user"
		}

		role := msg.Role
		if role == "" {
			role = "user"
		}

		var usage *adapter.Usage
		est := estimateTokens(msg.Content)
		if msg.Metadata != nil && msg.Metadata.NumTokens != nil && *msg.Metadata.NumTokens > 0 {
			t := *msg.Metadata.NumTokens
			usage = &adapter.Usage{InputTokens: est, OutputTokens: t}
		} else if est > 0 {
			usage = &adapter.Usage{InputTokens: 0, OutputTokens: est}
		}

		msgObj := &adapter.Message{
			ID:         msg.MessageID,
			Role:       role,
			StopReason: "end_turn",
			Content: []adapter.Content{
				{Type: "text", Text: truncate(msg.Content, 500)},
			},
			Model: model,
			Usage: usage,
		}

		tokenSrc := adapter.TokenEstimated
		if msg.Metadata != nil && msg.Metadata.NumTokens != nil && *msg.Metadata.NumTokens > 0 {
			tokenSrc = adapter.TokenExact
		}

		evt := adapter.Event{
			Type:      evtType,
			SessionID: sid,
			Timestamp: ts,
			Message:   msgObj,
			AgentID:   adapter.AgentDevin,
			TokenSrc:  tokenSrc,
		}

		events = append(events, evt)
	}

	if events == nil {
		events = []adapter.Event{}
	}

	meta := &adapter.SessionMeta{
		ID:          sid,
		Summary:     title,
		CreatedAt:   createdAtTime,
		ProjectPath: workingDir,
	}
	if len(events) > 0 {
		meta.UpdatedAt = events[len(events)-1].Timestamp
	}

	return &adapter.ParseResult{Events: events, Meta: meta}, nil
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) > n {
		return string(runes[:n]) + "..."
	}
	return s
}

func estimateTokens(s string) int {
	return len([]rune(s)) / 4
}

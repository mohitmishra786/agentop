package opencode

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

func parseSessionFile(path string) (*adapter.ParseResult, error) {
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

	var dir string
	var timeCreated int64
	err = db.QueryRow(
		"SELECT directory, time_created FROM session WHERE id = ?", sid,
	).Scan(&dir, &timeCreated)
	if err != nil {
		return nil, err
	}

	createdAtTime := time.UnixMilli(timeCreated)

	rows, err := db.Query(
		"SELECT id, time_created, data FROM message WHERE session_id = ? ORDER BY time_created ASC",
		sid,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var events []adapter.Event
	for rows.Next() {
		var id string
		var msgTime int64
		var dataStr string
		if err := rows.Scan(&id, &msgTime, &dataStr); err != nil {
			continue
		}

		ts := time.UnixMilli(msgTime)

		var data messageData
		if dataStr != "" {
			_ = json.Unmarshal([]byte(dataStr), &data)
		}

		msg := &adapter.Message{
			ID:         id,
			Role:       data.Role,
			StopReason: "end_turn",
		}

		if data.ModelID != "" {
			if data.ProviderID != "" {
				msg.Model = data.ProviderID + "/" + data.ModelID
			} else {
				msg.Model = data.ModelID
			}
		}

		if data.Tokens != nil {
			usage := &adapter.Usage{
				InputTokens:  data.Tokens.Input,
				OutputTokens: data.Tokens.Output,
			}
			if data.Tokens.Cache != nil {
				usage.CacheReadInputTokens = data.Tokens.Cache.Read
				usage.CacheCreationInputTokens = data.Tokens.Cache.Write
			}
			msg.Usage = usage
		}

		evt := adapter.Event{
			Type:      mapRole(data.Role),
			SessionID: sid,
			UUID:      id,
			Timestamp: ts,
			Message:   msg,
			CostUSD:   data.Cost,
			AgentID:   adapter.AgentOpenCode,
			TokenSrc:  adapter.TokenExact,
		}

		events = append(events, evt)
	}

	if events == nil {
		events = []adapter.Event{}
	}

	meta := &adapter.SessionMeta{
		ID:          sid,
		CreatedAt:   createdAtTime,
		ProjectPath: dir,
	}
	if len(events) > 0 {
		meta.UpdatedAt = events[len(events)-1].Timestamp
	}

	return &adapter.ParseResult{Events: events, Meta: meta}, nil
}

func mapRole(role string) string {
	switch role {
	case "user":
		return "user"
	case "assistant":
		return "assistant"
	case "tool":
		return "tool_result"
	default:
		return role
	}
}

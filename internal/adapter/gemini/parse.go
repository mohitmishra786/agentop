package gemini

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

	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".jsonl" {
		return parseJSONL(f)
	}
	return parseJSON(f)
}

func parseJSON(f *os.File) (*adapter.ParseResult, error) {
	var sess geminiSession
	if err := json.NewDecoder(f).Decode(&sess); err != nil {
		return nil, err
	}
	return buildResult(sess.Messages, sess), nil
}

func parseJSONL(f *os.File) (*adapter.ParseResult, error) {
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, maxLineBytes), maxLineBytes)

	var sessID, summary, cwd, branch, created, updated string
	var messages []geminiMessage

	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		var l geminiLine
		if err := json.Unmarshal(line, &l); err != nil {
			continue
		}
		switch l.Type {
		case "metadata":
			sessID = l.SessionID
			summary = l.Summary
			cwd = l.CWD
			branch = l.Branch
			created = l.Created
			updated = l.Updated
		case "message":
			if l.Message != nil {
				messages = append(messages, *l.Message)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		if errors.Is(err, bufio.ErrTooLong) {
			return buildResult(messages, geminiSession{
				ID: sessID, Summary: summary, CWD: cwd,
				Branch: branch, Created: created, Updated: updated,
			}), nil
		}
		return nil, err
	}

	return buildResult(messages, geminiSession{
		ID: sessID, Summary: summary, CWD: cwd,
		Branch: branch, Created: created, Updated: updated,
	}), nil
}

func buildResult(messages []geminiMessage, sess geminiSession) *adapter.ParseResult {
	var events []adapter.Event
	var createdAt, updatedAt time.Time

	createdAt, _ = tryParseTime(sess.Created)
	updatedAt, _ = tryParseTime(sess.Updated)

	for _, msg := range messages {
		ts, _ := tryParseTime(msg.Timestamp)
		if ts.IsZero() {
			ts = time.Now()
		}

		usage := convertTokens(msg.Tokens)

		var content []adapter.Content
		if msg.Content != "" {
			content = append(content, adapter.Content{Type: "text", Text: msg.Content})
		}
		if msg.Tokens != nil && msg.Tokens.Thoughts > 0 {
			content = append(content, adapter.Content{Type: "thinking", Text: ""})
		}

		eventType := msg.Role
		switch eventType {
		case "model", "assistant":
			eventType = "assistant"
		case "user":
			eventType = "user"
		case "tool", "function":
			eventType = "tool"
		}

		events = append(events, adapter.Event{
			Type:      eventType,
			SessionID: sess.ID,
			Timestamp: ts,
			AgentID:   adapter.AgentGemini,
			TokenSrc:  adapter.TokenExact,
			Message: &adapter.Message{
				ID:      msg.ID,
				Role:    msg.Role,
				Model:   msg.Model,
				Content: content,
				Usage:   usage,
			},
		})
	}

	return &adapter.ParseResult{
		Events: events,
		Meta: &adapter.SessionMeta{
			ID:           sess.ID,
			Summary:      sess.Summary,
			MessageCount: len(messages),
			CreatedAt:    createdAt,
			UpdatedAt:    updatedAt,
			CWD:          sess.CWD,
			GitBranch:    sess.Branch,
		},
	}
}

// convertTokens maps Gemini token counts to the shared adapter.Usage.
// Gemini reports cached tokens inside the input total, so we subtract
// cached from input before pricing and store cached separately.
func convertTokens(t *geminiTokens) *adapter.Usage {
	if t == nil {
		return nil
	}
	inputTokens := t.Input - t.Cached
	if inputTokens < 0 {
		inputTokens = 0
	}
	return &adapter.Usage{
		InputTokens:          inputTokens,
		OutputTokens:         t.Output,
		CacheReadInputTokens: t.Cached,
	}
}

func tryParseTime(s string) (time.Time, bool) {
	if s == "" {
		return time.Time{}, false
	}
	t, err := time.Parse(time.RFC3339, s)
	if err == nil {
		return t, true
	}
	t, err = time.Parse(time.RFC3339Nano, s)
	if err == nil {
		return t, true
	}
	return time.Time{}, false
}

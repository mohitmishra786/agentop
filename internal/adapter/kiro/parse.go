package kiro

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

var toolNameMapping = map[string]string{
	"Bash":              "Bash",
	"execute_command":   "Bash",
	"executeCommand":    "Bash",
	"Read":              "Read",
	"read_file":         "Read",
	"readFile":          "Read",
	"Write":             "Write",
	"write_file":        "Write",
	"writeFile":         "Write",
	"Edit":              "Edit",
	"edit_file":         "Edit",
	"editFile":          "Edit",
	"str_replace_editor": "Edit",
	"Glob":              "Glob",
	"glob_files":        "Glob",
	"globFiles":         "Glob",
	"Grep":              "Grep",
	"search_files":      "Grep",
	"searchFiles":       "Grep",
	"grep":              "Grep",
	"WebSearch":         "WebSearch",
	"web_search":        "WebSearch",
	"webSearch":         "WebSearch",
	"WebFetch":          "WebFetch",
	"web_fetch":         "WebFetch",
	"webFetch":          "WebFetch",
	"ListDir":           "ListDir",
	"list_dir":          "ListDir",
	"listDir":           "ListDir",
	"FileSearch":        "FileSearch",
	"file_search":       "FileSearch",
	"fileSearch":        "FileSearch",
	"subagent":          "Subagent",
}

func normalizeToolName(name string) string {
	if name == "" {
		return name
	}
	if canonical, ok := toolNameMapping[name]; ok {
		return canonical
	}
	return name
}

func parseSession(path string) (*adapter.ParseResult, error) {
	if !strings.HasSuffix(path, ".json") {
		return nil, errors.New("kiro: expected .json metadata file")
	}

	meta, err := parseMeta(path)
	if err != nil {
		return nil, err
	}

	jsonlPath := strings.TrimSuffix(path, ".json") + ".jsonl"
	events, err := parseJSONL(jsonlPath, meta)
	if err != nil {
		return nil, err
	}

	sessionID := filepath.Base(strings.TrimSuffix(path, ".json"))

	result := &adapter.ParseResult{
		Events: events,
		Meta: &adapter.SessionMeta{
			ID:          sessionID,
			CWD:         meta.CWD,
			Summary:     meta.Title,
			ProjectPath: meta.CWD,
		},
	}

	for i := range result.Events {
		result.Events[i].AgentID = adapter.AgentKiro
		result.Events[i].TokenSrc = adapter.TokenEstimated
	}

	return result, nil
}

func parseMeta(path string) (*kiroMeta, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m kiroMeta
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func parseJSONL(path string, meta *kiroMeta) ([]adapter.Event, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, maxLineBytes), maxLineBytes)

	var events []adapter.Event
	sessionID := filepath.Base(filepath.Dir(path))

	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}

		var ke kiroEvent
		if err := json.Unmarshal(line, &ke); err != nil {
			continue
		}

		ts := parseTimestamp(ke.Timestamp)

		switch ke.Kind {
		case "Prompt":
			events = append(events, adapter.Event{
				Type:      "user",
				SessionID: sessionID,
				Timestamp: ts,
				CWD:       meta.CWD,
				Message: &adapter.Message{
					Role: "user",
					Usage: &adapter.Usage{
						InputTokens: len(ke.Content) / 4,
					},
				},
			})

		case "AssistantMessage":
			events = append(events, adapter.Event{
				Type:      "assistant",
				SessionID: sessionID,
				Timestamp: ts,
				Message: &adapter.Message{
					Model: meta.Model,
					Role:  "assistant",
					Usage: &adapter.Usage{
						OutputTokens: len(ke.Content) / 4,
					},
				},
				ToolName:  normalizeToolName(ke.ToolName),
				ToolInput: ke.ToolInput,
			})

		case "ToolResults":
			content := ke.Content
			var result json.RawMessage
			if content != "" {
				b, _ := json.Marshal(content)
				result = json.RawMessage(b)
			}

			events = append(events, adapter.Event{
				Type:       "tool_result",
				SessionID:  sessionID,
				Timestamp:  ts,
				ToolName:   normalizeToolName(ke.ToolName),
				ToolResult: result,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		if errors.Is(err, bufio.ErrTooLong) {
			return events, nil
		}
		return nil, err
	}

	return events, nil
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

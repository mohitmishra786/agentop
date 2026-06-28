package windsurf

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/agentop-dev/agentop/internal/adapter"
)

const (
	bytesPerStep = 5120    // avg encrypted step size
	tokenDivisor = 4       // chars per estimated token
	inputRatio   = 0.4     // typical input/output split
	maxReadBytes = 1 << 20 // 1MB — only read header to detect format
)

// parseSession reads a .pb file and returns estimated session data.
// Actual AES-256-GCM decryption of the protobuf is not feasible without the
// hardcoded key from the language_server binary. Instead we use file metadata
// (size, name, mod time) to produce reasonable estimates.
func parseSession(path string) (*adapter.ParseResult, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Detect whether the file looks like a text-based format (e.g. JSONL) or
	// binary (encrypted protobuf). If it contains valid UTF-8 text we try a
	// simple line-based parse; otherwise we estimate from file size.
	events := estimateFromSize(path, fi, data)

	sessionID := sessionIDFromPath(path)

	result := &adapter.ParseResult{
		Events: events,
		Meta: &adapter.SessionMeta{
			ID:           sessionID,
			CreatedAt:    fi.ModTime(),
			UpdatedAt:    fi.ModTime(),
			MessageCount: len(events),
		},
	}

	for i := range result.Events {
		result.Events[i].AgentID = adapter.AgentWindsurf
		result.Events[i].TokenSrc = adapter.TokenEstimated
	}

	return result, nil
}

// estimateFromSize produces placeholder events based on the encrypted file
// size and filename. Since the .pb content is AES-256-GCM encrypted we cannot
// read steps directly; we estimate the number of steps from the file size.
func estimateFromSize(path string, fi os.FileInfo, data []byte) []adapter.Event {
	if len(data) == 0 {
		return nil
	}

	// If the file looks textual (not encrypted protobuf), attempt basic parse.
	if looksTextual(data) {
		events, err := tryTextParse(data, path, fi)
		if err == nil && len(events) > 0 {
			return events
		}
	}

	// Estimate step count from file size (encrypted steps are ~5KB each on average).
	stepCount := int(fi.Size() / bytesPerStep)
	if stepCount < 1 {
		stepCount = 1
	}
	if stepCount > 200 {
		stepCount = 200
	}

	// Estimate total tokens from file size (rough heuristic: encrypted size / 2 ≈
	// content bytes, then / tokenDivisor ≈ tokens).
	estimatedBytes := int(fi.Size())
	if estimatedBytes > 10*1024*1024 {
		estimatedBytes = 10 * 1024 * 1024
	}
	totalTokens := estimatedBytes / 2 / tokenDivisor

	var events []adapter.Event
	baseTime := fi.ModTime().Add(-time.Duration(stepCount) * time.Minute)
	sessionID := sessionIDFromPath(path)

	for i := 0; i < stepCount; i++ {
		ts := baseTime.Add(time.Duration(i) * time.Minute)
		eType := "user"
		var usage *adapter.Usage
		if i == 0 || i%2 == 0 {
			usage = &adapter.Usage{
				InputTokens: int(float64(totalTokens) * inputRatio / float64(stepCount)),
			}
		} else {
			usage = &adapter.Usage{
				OutputTokens: int(float64(totalTokens) * (1 - inputRatio) / float64(stepCount)),
			}
			eType = "assistant"
		}

		events = append(events, adapter.Event{
			Type:      eType,
			SessionID: sessionID,
			Timestamp: ts,
			Message: &adapter.Message{
				Role:  eType,
				Usage: usage,
			},
		})
	}

	return events
}

// looksTextual checks whether the data appears to be a text-based format rather
// than encrypted binary protobuf.
func looksTextual(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	// Check for JSON-like start.
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return false
	}
	if trimmed[0] == '{' || trimmed[0] == '[' {
		return true
	}
	// Check for reasonable ASCII ratio.
	nonASCII := 0
	for _, b := range trimmed {
		if b > 127 {
			nonASCII++
		}
	}
	return float64(nonASCII)/float64(len(trimmed)) < 0.1
}

// tryTextParse attempts to parse a text-based session file (JSONL or JSON).
// This handles the edge case where a session file isn't actually encrypted.
func tryTextParse(data []byte, path string, fi os.FileInfo) ([]adapter.Event, error) {
	sessionID := sessionIDFromPath(path)

	if len(data) > 0 && data[0] == '{' {
		// Maybe it's a single JSON object (protobuf JSON mapping).
		return []adapter.Event{
			{
				Type:      "summary",
				SessionID: sessionID,
				Timestamp: fi.ModTime(),
				Message: &adapter.Message{
					Role: "assistant",
					Usage: &adapter.Usage{
						InputTokens: int(fi.Size()) / tokenDivisor / 2,
					},
				},
			},
		}, nil
	}

	lines := bytes.Split(bytes.TrimSpace(data), []byte{'\n'})
	if len(lines) == 0 {
		return nil, nil
	}

	var events []adapter.Event
	baseTime := fi.ModTime().Add(-time.Duration(len(lines)) * time.Second)

	for i, line := range lines {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		ts := baseTime.Add(time.Duration(i) * time.Second)
		eType := "user"
		if i%2 == 1 {
			eType = "assistant"
		}

		events = append(events, adapter.Event{
			Type:      eType,
			SessionID: sessionID,
			Timestamp: ts,
			Message: &adapter.Message{
				Role: eType,
				Usage: &adapter.Usage{
					InputTokens: len(line) / tokenDivisor,
				},
			},
		})
	}

	return events, nil
}

// sessionIDFromPath extracts a session UUID from the file path.
func sessionIDFromPath(path string) string {
	base := strings.TrimSuffix(filepath.Base(path), ".pb")
	return base
}

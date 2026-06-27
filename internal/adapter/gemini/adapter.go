// Package gemini implements the adapter.Adapter interface for Gemini CLI
// sessions stored under ~/.gemini/tmp/.
package gemini

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/agentop-dev/agentop/internal/adapter"
)

var _ adapter.Adapter = (*Adapter)(nil)

// Adapter implements adapter.Adapter for Gemini CLI session files.
type Adapter struct{}

// ID returns the unique agent identifier for Gemini CLI.
func (a *Adapter) ID() adapter.AgentID { return adapter.AgentGemini }

// Name returns a human-readable display name.
func (a *Adapter) Name() string { return "Gemini CLI" }

// DefaultDir returns the default Gemini CLI data directory (~/.gemini).
func (a *Adapter) DefaultDir() string { return geminiDefaultDir() }

// IsAvailable returns true if ~/.gemini/tmp exists.
func (a *Adapter) IsAvailable() bool { return geminiAvailable() }

// Discover finds all Gemini CLI session files under ~/.gemini/tmp/<hash>/chats/.
func (a *Adapter) Discover(dataDir string) ([]adapter.SessionFile, error) {
	tmpDir := filepath.Join(dataDir, "tmp")

	projectEntries, err := os.ReadDir(tmpDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var files []adapter.SessionFile

	for _, projectEntry := range projectEntries {
		if !projectEntry.IsDir() {
			continue
		}
		projectHash := projectEntry.Name()
		chatsDir := filepath.Join(tmpDir, projectHash, "chats")

		chatEntries, err := os.ReadDir(chatsDir)
		if err != nil {
			continue
		}

		for _, chatEntry := range chatEntries {
			if chatEntry.IsDir() {
				continue
			}
			name := chatEntry.Name()
			if !strings.HasPrefix(name, "session-") {
				continue
			}
			ext := filepath.Ext(name)
			if ext != ".json" && ext != ".jsonl" {
				continue
			}

			info, err := chatEntry.Info()
			if err != nil {
				continue
			}

			sessionID := strings.TrimSuffix(strings.TrimPrefix(name, "session-"), ext)

			files = append(files, adapter.SessionFile{
				Path:        filepath.Join(chatsDir, name),
				ProjectHash: projectHash,
				SessionID:   sessionID,
				ModTime:     info.ModTime(),
				AgentID:     adapter.AgentGemini,
			})
		}
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime.After(files[j].ModTime)
	})

	return files, nil
}

// ParseSession reads a Gemini session file and returns parsed events.
func (a *Adapter) ParseSession(path string) (*adapter.ParseResult, error) {
	return parseSessionFile(path)
}

func geminiDefaultDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".gemini"
	}
	return filepath.Join(home, ".gemini")
}

func geminiAvailable() bool {
	tmpDir := filepath.Join(geminiDefaultDir(), "tmp")
	info, err := os.Stat(tmpDir)
	return err == nil && info.IsDir()
}

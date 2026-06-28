// Package copilot implements the adapter.Adapter interface for Copilot CLI session data.
package copilot

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/agentop-dev/agentop/internal/adapter"
)

var _ adapter.Adapter = (*Adapter)(nil)

// Adapter implements adapter.Adapter for Copilot CLI session files.
type Adapter struct{}

// ID returns the agent identifier.
func (a *Adapter) ID() adapter.AgentID { return adapter.AgentCopilot }

// Name returns the display name of the agent.
func (a *Adapter) Name() string { return "Copilot CLI" }

// DefaultDir returns the default data directory for Copilot CLI.
func (a *Adapter) DefaultDir() string { return copilotDefaultDir() }

// IsAvailable checks whether Copilot CLI session data exists.
func (a *Adapter) IsAvailable() bool { return copilotAvailable() }

// Discover finds Copilot CLI session files in the given data directory.
func (a *Adapter) Discover(dataDir string) ([]adapter.SessionFile, error) {
	var files []adapter.SessionFile

	entries, err := os.ReadDir(dataDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		eventsPath := filepath.Join(dataDir, entry.Name(), "events.jsonl")
		info, err := os.Stat(eventsPath)
		if err != nil {
			continue
		}
		files = append(files, adapter.SessionFile{
			Path:      eventsPath,
			SessionID: entry.Name(),
			ModTime:   info.ModTime(),
			AgentID:   adapter.AgentCopilot,
		})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime.After(files[j].ModTime)
	})

	return files, nil
}

// ParseSession reads and converts a Copilot CLI JSONL session file.
func (a *Adapter) ParseSession(path string) (*adapter.ParseResult, error) {
	return parseSessionFile(path)
}

func copilotDefaultDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".copilot"
	}
	return filepath.Join(home, ".copilot", "session-state")
}

func copilotAvailable() bool {
	info, err := os.Stat(copilotDefaultDir())
	return err == nil && info.IsDir()
}

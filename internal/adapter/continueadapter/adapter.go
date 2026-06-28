// Package continueadapter implements the adapter.Adapter interface for Continue session data.
package continueadapter

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/agentop-dev/agentop/internal/adapter"
)

var _ adapter.Adapter = (*Adapter)(nil)

// Adapter implements adapter.Adapter for Continue session files.
type Adapter struct{}

// ID returns the agent identifier.
func (a *Adapter) ID() adapter.AgentID { return adapter.AgentContinue }

// Name returns the display name of the agent.
func (a *Adapter) Name() string { return "Continue" }

// DefaultDir returns the default data directory for Continue.
func (a *Adapter) DefaultDir() string { return continueDefaultDir() }

// IsAvailable checks whether Continue session data exists.
func (a *Adapter) IsAvailable() bool { return continueAvailable() }

// Discover finds Continue session files in the given data directory.
func (a *Adapter) Discover(dataDir string) ([]adapter.SessionFile, error) {
	sessionsDir := filepath.Join(dataDir, "sessions")
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var files []adapter.SessionFile
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		files = append(files, adapter.SessionFile{
			Path:      filepath.Join(sessionsDir, entry.Name()),
			SessionID: strings.TrimSuffix(entry.Name(), ".json"),
			ModTime:   info.ModTime(),
			AgentID:   adapter.AgentContinue,
		})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime.After(files[j].ModTime)
	})

	return files, nil
}

// ParseSession reads and converts a Continue JSON session file.
func (a *Adapter) ParseSession(path string) (*adapter.ParseResult, error) {
	return parseSessionFile(path)
}

func continueDefaultDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".continue"
	}
	return filepath.Join(home, ".continue")
}

func continueAvailable() bool {
	sessionsDir := filepath.Join(continueDefaultDir(), "sessions")
	info, err := os.Stat(sessionsDir)
	return err == nil && info.IsDir()
}

var _ = time.Time{}

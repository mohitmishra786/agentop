// Package kiro implements the adapter.Adapter interface for Kiro CLI session data.
package kiro

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/agentop-dev/agentop/internal/adapter"
)

var _ adapter.Adapter = (*Adapter)(nil)

// Adapter implements adapter.Adapter for Kiro CLI session files.
type Adapter struct{}

// ID returns the agent identifier.
func (a *Adapter) ID() adapter.AgentID { return adapter.AgentKiro }

// Name returns the display name of the agent.
func (a *Adapter) Name() string { return "Kiro CLI" }

// DefaultDir returns the default data directory for Kiro CLI.
func (a *Adapter) DefaultDir() string { return kiroDefaultDir() }

// IsAvailable checks whether Kiro CLI session data exists.
func (a *Adapter) IsAvailable() bool { return kiroAvailable() }

// Discover finds Kiro CLI session files in the given data directory.
func (a *Adapter) Discover(dataDir string) ([]adapter.SessionFile, error) {
	sessionsDir := filepath.Join(dataDir, "sessions", "cli")
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var files []adapter.SessionFile
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		uuid := entry.Name()
		metaPath := filepath.Join(sessionsDir, uuid, uuid+".json")
		jsonlPath := filepath.Join(sessionsDir, uuid, uuid+".jsonl")

		metaInfo, err := os.Stat(metaPath)
		if err != nil {
			continue
		}
		if _, err := os.Stat(jsonlPath); err != nil {
			continue
		}

		files = append(files, adapter.SessionFile{
			Path:      metaPath,
			SessionID: uuid,
			ModTime:   metaInfo.ModTime(),
			AgentID:   adapter.AgentKiro,
		})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime.After(files[j].ModTime)
	})

	return files, nil
}

// ParseSession reads and converts a Kiro CLI session file.
func (a *Adapter) ParseSession(path string) (*adapter.ParseResult, error) {
	return parseSession(path)
}

func kiroDefaultDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".kiro"
	}
	return filepath.Join(home, ".kiro")
}

func kiroAvailable() bool {
	sessionsDir := filepath.Join(kiroDefaultDir(), "sessions", "cli")
	info, err := os.Stat(sessionsDir)
	return err == nil && info.IsDir()
}

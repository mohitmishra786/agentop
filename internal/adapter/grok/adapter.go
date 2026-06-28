// Package grok implements the adapter.Adapter interface for Grok CLI sessions
// stored under ~/.grok/sessions/.
package grok

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/agentop-dev/agentop/internal/adapter"
)

var _ adapter.Adapter = (*Adapter)(nil)

// Adapter implements adapter.Adapter for Grok CLI sessions.
type Adapter struct{}

// ID returns the agent identifier for Grok CLI.
func (a *Adapter) ID() adapter.AgentID { return adapter.AgentGrok }

// Name returns the human-readable display name.
func (a *Adapter) Name() string { return "Grok CLI" }

// DefaultDir returns the default Grok CLI data directory.
func (a *Adapter) DefaultDir() string { return defaultDir() }

// IsAvailable returns true if the Grok CLI sessions directory exists.
func (a *Adapter) IsAvailable() bool { return available() }

// Discover finds all Grok CLI session directories under ~/.grok/sessions/<project>/<session_id>/.
func (a *Adapter) Discover(dataDir string) ([]adapter.SessionFile, error) {
	sessionsDir := filepath.Join(dataDir, "sessions")
	projectEntries, err := os.ReadDir(sessionsDir)
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
		projectPath := projectEntry.Name()
		projectDir := filepath.Join(sessionsDir, projectPath)

		sessionDirs, err := os.ReadDir(projectDir)
		if err != nil {
			continue
		}
		for _, sessionEntry := range sessionDirs {
			if !sessionEntry.IsDir() {
				continue
			}
			sessionID := sessionEntry.Name()
			sessionPath := filepath.Join(projectDir, sessionID)
			summaryPath := filepath.Join(sessionPath, "summary.json")

			info, err := os.Stat(summaryPath)
			if err != nil {
				continue
			}

			files = append(files, adapter.SessionFile{
				Path:        sessionPath,
				ProjectHash: projectPath,
				SessionID:   sessionID,
				ModTime:     info.ModTime(),
				AgentID:     adapter.AgentGrok,
			})
		}
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime.After(files[j].ModTime)
	})

	return files, nil
}

// ParseSession reads a Grok session directory and returns parsed events.
func (a *Adapter) ParseSession(path string) (*adapter.ParseResult, error) {
	return parseSession(path)
}

func defaultDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".grok"
	}
	return filepath.Join(home, ".grok")
}

func available() bool {
	sessionsDir := filepath.Join(defaultDir(), "sessions")
	info, err := os.Stat(sessionsDir)
	return err == nil && info.IsDir()
}

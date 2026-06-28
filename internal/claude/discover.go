// Package claude provides types and functions for Claude Code session data.
package claude

import (
	"errors"
	"time"

	claudeAdapter "github.com/agentop-dev/agentop/internal/adapter/claude"
)

// ErrClaudeNotFound is returned when the Claude Code data directory doesn't exist.
var ErrClaudeNotFound = errors.New("~/.claude/projects/ not found — is Claude Code installed?")

// SessionFile represents a discovered Claude Code session file on disk.
type SessionFile struct {
	Path          string
	ProjectHash   string
	SessionID     string
	ModTime       time.Time
	SubagentFiles []string
}

var claudeAdapterInstance = &claudeAdapter.Adapter{}

// Discover finds session files in the given Claude Code data directory.
func Discover(claudeDir string) ([]SessionFile, error) {
	files, err := claudeAdapterInstance.Discover(claudeDir)
	if err != nil {
		if errors.Is(err, claudeAdapter.ErrClaudeNotFound) {
			return nil, ErrClaudeNotFound
		}
		return nil, err
	}

	sessions := make([]SessionFile, len(files))
	for i, f := range files {
		sessions[i] = SessionFile{
			Path:          f.Path,
			ProjectHash:   f.ProjectHash,
			SessionID:     f.SessionID,
			ModTime:       f.ModTime,
			SubagentFiles: f.SubagentFiles,
		}
	}
	return sessions, nil
}

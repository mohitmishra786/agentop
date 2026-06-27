package claude

import (
	"errors"
	"time"

	claudeAdapter "github.com/agentop-dev/agentop/internal/adapter/claude"
)

var ErrClaudeNotFound = errors.New("~/.claude/projects/ not found — is Claude Code installed?")

type SessionFile struct {
	Path          string
	ProjectHash   string
	SessionID     string
	ModTime       time.Time
	SubagentFiles []string
}

var claudeAdapterInstance = &claudeAdapter.Adapter{}

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

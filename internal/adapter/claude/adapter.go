// Package claude implements the adapter.Adapter interface for Claude Code
// session data (JSONL files under ~/.claude/projects/).
package claude

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/agentop-dev/agentop/internal/adapter"
)

var _ adapter.Adapter = (*Adapter)(nil)

// Adapter implements adapter.Adapter for Claude Code session files.
type Adapter struct{}

// ID returns the agent identifier.
func (a *Adapter) ID() adapter.AgentID { return adapter.AgentClaude }

// Name returns the display name of the agent.
func (a *Adapter) Name() string { return "Claude Code" }

// DefaultDir returns the default data directory for Claude Code.
func (a *Adapter) DefaultDir() string { return claudeDefaultDir() }

// IsAvailable checks whether Claude Code session data exists.
func (a *Adapter) IsAvailable() bool { return claudeAvailable() }

// Discover finds Claude Code session files in the given data directory.
func (a *Adapter) Discover(dataDir string) ([]adapter.SessionFile, error) {
	files, err := discoverSessionFiles(dataDir)
	if err != nil {
		return nil, err
	}
	adapted := make([]adapter.SessionFile, len(files))
	for i, f := range files {
		adapted[i] = adapter.SessionFile{
			Path:          f.Path,
			ProjectHash:   f.ProjectHash,
			SessionID:     f.SessionID,
			ModTime:       f.ModTime,
			SubagentFiles: f.SubagentFiles,
			AgentID:       adapter.AgentClaude,
		}
	}
	return adapted, nil
}

// ParseSession reads and converts a Claude Code JSONL session file.
func (a *Adapter) ParseSession(path string) (*adapter.ParseResult, error) {
	events, err := parseSessionFile(path)
	if err != nil {
		return nil, err
	}

	adapted := make([]adapter.Event, len(events))
	for i, e := range events {
		ae := adaptEvent(e)
		ae.AgentID = adapter.AgentClaude
		ae.TokenSrc = adapter.TokenExact
		adapted[i] = ae
	}

	return &adapter.ParseResult{Events: adapted}, nil
}

func claudeDefaultDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".claude"
	}
	return filepath.Join(home, ".claude")
}

func claudeAvailable() bool {
	projectsDir := filepath.Join(claudeDefaultDir(), "projects")
	info, err := os.Stat(projectsDir)
	if err != nil || !info.IsDir() {
		return false
	}
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if e.IsDir() {
			sessionDir := filepath.Join(projectsDir, e.Name())
			sessionFiles, _ := os.ReadDir(sessionDir)
			for _, f := range sessionFiles {
				if !f.IsDir() && strings.HasSuffix(f.Name(), ".jsonl") {
					return true
				}
			}
		}
	}
	return false
}

// ErrClaudeNotFound is returned when the Claude Code data directory doesn't exist.
var ErrClaudeNotFound = errors.New("~/.claude/projects/ not found — is Claude Code installed?")

type claudeSessionFile struct {
	Path          string
	ProjectHash   string
	SessionID     string
	ModTime       time.Time
	SubagentFiles []string
}

func discoverSessionFiles(claudeDir string) ([]claudeSessionFile, error) {
	projectsDir := filepath.Join(claudeDir, "projects")

	projectEntries, err := os.ReadDir(projectsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrClaudeNotFound
		}
		return nil, fmt.Errorf("reading projects dir: %w", err)
	}

	var sessions []claudeSessionFile
	for _, projectEntry := range projectEntries {
		if !projectEntry.IsDir() {
			continue
		}
		projectDir := filepath.Join(projectsDir, projectEntry.Name())
		collectSessions(projectDir, projectEntry.Name(), &sessions)
	}

	return sessions, nil
}

func collectSessions(projectDir, projectHash string, sessions *[]claudeSessionFile) {
	entries, err := os.ReadDir(projectDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			dirName := entry.Name()
			if dirName == "tool-results" || dirName == "memory" {
				continue
			}
			if dirName == "subagents" {
				continue
			}
			collectSessions(filepath.Join(projectDir, dirName), projectHash, sessions)
			continue
		}

		if strings.HasSuffix(entry.Name(), ".jsonl") {
			sessionPath := filepath.Join(projectDir, entry.Name())
			sessionID := strings.TrimSuffix(entry.Name(), ".jsonl")
			subFiles := findSubagents(projectDir, sessionID)

			info, _ := entry.Info()

			*sessions = append(*sessions, claudeSessionFile{
				Path:          sessionPath,
				ProjectHash:   projectHash,
				SessionID:     sessionID,
				ModTime:       info.ModTime(),
				SubagentFiles: subFiles,
			})
		}
	}
}

func findSubagents(projectDir, sessionID string) []string {
	subDir := filepath.Join(projectDir, sessionID, "subagents")
	entries, err := os.ReadDir(subDir)
	if err != nil {
		return nil
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".jsonl") {
			files = append(files, filepath.Join(subDir, e.Name()))
		}
	}
	return files
}

func convertUsage(u rawUsage) adapter.Usage {
	au := adapter.Usage{
		InputTokens:              u.InputTokens,
		OutputTokens:             u.OutputTokens,
		CacheCreationInputTokens: u.CacheCreationInputTokens,
		CacheReadInputTokens:     u.CacheReadInputTokens,
		ServiceTier:              u.ServiceTier,
	}
	if u.CacheCreation != nil {
		au.CacheCreation = &adapter.CacheTiers{
			Ephemeral5m: u.CacheCreation.Ephemeral5m,
			Ephemeral1h: u.CacheCreation.Ephemeral1h,
		}
	}
	return au
}

func adaptEvent(e rawEvent) adapter.Event {
	var msg *adapter.Message
	if e.Message != nil {
		msg = &adapter.Message{
			ID:           e.Message.ID,
			Type:         e.Message.Type,
			Role:         e.Message.Role,
			Model:        e.Message.Model,
			StopReason:   e.Message.StopReason,
			StopSequence: e.Message.StopSequence,
		}
		if e.Message.Usage != nil {
			au := convertUsage(*e.Message.Usage)
			msg.Usage = &au
		}
		if len(e.Message.Content) > 0 {
			msg.Content = make([]adapter.Content, len(e.Message.Content))
			for j, c := range e.Message.Content {
				msg.Content[j] = adapter.Content{
					Type:      c.Type,
					Text:      c.Text,
					Thinking:  c.Thinking,
					Signature: c.Signature,
				}
			}
		}
	}

	var snap *adapter.FileSnapshot
	if e.Snapshot != nil {
		snap = &adapter.FileSnapshot{
			MessageID:          e.Snapshot.MessageID,
			TrackedFileBackups: e.Snapshot.TrackedFileBackups,
			Timestamp:          e.Snapshot.Timestamp,
		}
	}

	return adapter.Event{
		Type:             e.Type,
		Message:          msg,
		ToolName:         e.ToolName,
		ToolInput:        e.ToolInput,
		ToolResult:       e.ToolResult,
		Summary:          e.Summary,
		SessionID:        e.SessionID,
		UUID:             e.UUID,
		ParentUUID:       e.ParentUUID,
		Timestamp:        e.Timestamp,
		CostUSD:          e.CostUSD,
		CWD:              e.CWD,
		IsSidechain:      e.IsSidechain,
		MessageID:        e.MessageID,
		PromptID:         e.PromptID,
		Version:          e.Version,
		Entrypoint:       e.Entrypoint,
		UserType:         e.UserType,
		PermissionMode:   e.PermissionMode,
		GitBranch:        e.GitBranch,
		IsSnapshotUpdate: e.IsSnapshotUpdate,
		Snapshot:         snap,
		DurationMs:       e.DurationMs,
		Subtype:          e.Subtype,
	}
}

package claude

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var ErrClaudeNotFound = errors.New("~/.claude/projects/ not found — is Claude Code installed?")

type SessionFile struct {
	Path        string
	ProjectHash string
	SessionID   string
	ModTime     time.Time
}

func Discover(claudeDir string) ([]SessionFile, error) {
	projectsDir := filepath.Join(claudeDir, "projects")

	projectEntries, err := os.ReadDir(projectsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrClaudeNotFound
		}
		return nil, err
	}

	var sessions []SessionFile
	for _, projectEntry := range projectEntries {
		if !projectEntry.IsDir() {
			continue
		}

		projectDir := filepath.Join(projectsDir, projectEntry.Name())
		collectSessions(projectDir, projectEntry.Name(), &sessions)
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].ModTime.After(sessions[j].ModTime)
	})

	return sessions, nil
}

func collectSessions(projectDir, projectHash string, sessions *[]SessionFile) {
	entries, err := os.ReadDir(projectDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			dirName := entry.Name()
			if dirName == "subagents" || dirName == "tool-results" || dirName == "memory" {
				continue
			}
			collectSessions(filepath.Join(projectDir, dirName), projectHash, sessions)
			continue
		}

		if strings.HasSuffix(entry.Name(), ".jsonl") {
			info, _ := entry.Info()
			*sessions = append(*sessions, SessionFile{
				Path:        filepath.Join(projectDir, entry.Name()),
				ProjectHash: projectHash,
				SessionID:   strings.TrimSuffix(entry.Name(), ".jsonl"),
				ModTime:     info.ModTime(),
			})
		}
	}
}

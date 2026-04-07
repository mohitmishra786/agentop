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
	Path          string
	ProjectHash   string
	SessionID     string
	ModTime       time.Time
	SubagentFiles []string
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
			info, _ := entry.Info()
			sessionPath := filepath.Join(projectDir, entry.Name())
			sessionID := strings.TrimSuffix(entry.Name(), ".jsonl")

			subFiles := findSubagents(projectDir, sessionID)

			*sessions = append(*sessions, SessionFile{
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

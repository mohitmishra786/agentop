// Package jetbrains implements the adapter.Adapter interface for JetBrains Copilot session data.
package jetbrains

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/agentop-dev/agentop/internal/adapter"
)

var _ adapter.Adapter = (*Adapter)(nil)

// Adapter implements adapter.Adapter for JetBrains Copilot session files.
type Adapter struct{}

// ID returns the agent identifier.
func (a *Adapter) ID() adapter.AgentID { return adapter.AgentJetBrains }

// Name returns the display name of the agent.
func (a *Adapter) Name() string { return "JetBrains Copilot" }

// DefaultDir returns the default data directory for JetBrains Copilot.
func (a *Adapter) DefaultDir() string { return jetbrainsDefaultDir() }

// IsAvailable checks whether JetBrains Copilot session data exists.
func (a *Adapter) IsAvailable() bool { return jetbrainsAvailable() }

// Discover finds JetBrains Copilot session files in the given data directory.
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
		uuidDir := filepath.Join(dataDir, entry.Name())
		partitions, err := os.ReadDir(uuidDir)
		if err != nil {
			continue
		}
		for _, p := range partitions {
			if p.IsDir() {
				continue
			}
			if !strings.HasPrefix(p.Name(), "partition-") || !strings.HasSuffix(p.Name(), ".jsonl") {
				continue
			}
			info, err := p.Info()
			if err != nil {
				continue
			}
			files = append(files, adapter.SessionFile{
				Path:      filepath.Join(uuidDir, p.Name()),
				SessionID: entry.Name(),
				ModTime:   info.ModTime(),
				AgentID:   adapter.AgentJetBrains,
			})
		}
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime.After(files[j].ModTime)
	})

	return files, nil
}

// ParseSession reads and converts a JetBrains Copilot JSONL session file.
func (a *Adapter) ParseSession(path string) (*adapter.ParseResult, error) {
	return parseSessionFile(path)
}

func jetbrainsDefaultDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".copilot/jb"
	}
	return filepath.Join(home, ".copilot", "jb")
}

func jetbrainsAvailable() bool {
	info, err := os.Stat(jetbrainsDefaultDir())
	return err == nil && info.IsDir()
}

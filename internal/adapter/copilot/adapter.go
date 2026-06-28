package copilot

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/agentop-dev/agentop/internal/adapter"
)

var _ adapter.Adapter = (*Adapter)(nil)

type Adapter struct{}

func (a *Adapter) ID() adapter.AgentID { return adapter.AgentCopilot }
func (a *Adapter) Name() string        { return "Copilot CLI" }
func (a *Adapter) DefaultDir() string  { return copilotDefaultDir() }
func (a *Adapter) IsAvailable() bool   { return copilotAvailable() }

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

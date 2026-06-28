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

type Adapter struct{}

func (a *Adapter) ID() adapter.AgentID { return adapter.AgentContinue }
func (a *Adapter) Name() string        { return "Continue" }
func (a *Adapter) DefaultDir() string  { return continueDefaultDir() }
func (a *Adapter) IsAvailable() bool   { return continueAvailable() }

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

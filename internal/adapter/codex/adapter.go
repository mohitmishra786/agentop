package codex

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

func (a *Adapter) ID() adapter.AgentID { return adapter.AgentCodex }
func (a *Adapter) Name() string        { return "Codex CLI" }
func (a *Adapter) DefaultDir() string  { return codexDefaultDir() }
func (a *Adapter) IsAvailable() bool   { return codexAvailable() }

func (a *Adapter) Discover(dataDir string) ([]adapter.SessionFile, error) {
	var files []adapter.SessionFile

	err := filepath.Walk(dataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(info.Name(), ".jsonl") && strings.HasPrefix(info.Name(), "rollout-") {
			rel, _ := filepath.Rel(dataDir, path)
			parts := strings.Split(rel, string(filepath.Separator))
			hash := ""
			if len(parts) >= 3 {
				hash = strings.Join(parts[:3], "/")
			}
			files = append(files, adapter.SessionFile{
				Path:        path,
				ProjectHash: hash,
				SessionID:   strings.TrimSuffix(info.Name(), ".jsonl"),
				ModTime:     info.ModTime(),
				AgentID:     adapter.AgentCodex,
			})
		}
		return nil
	})

	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime.After(files[j].ModTime)
	})

	return files, err
}

func (a *Adapter) ParseSession(path string) (*adapter.ParseResult, error) {
	return parseSessionFile(path)
}

func codexDefaultDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".codex"
	}
	return filepath.Join(home, ".codex")
}

func codexAvailable() bool {
	sessionsDir := filepath.Join(codexDefaultDir(), "sessions")
	info, err := os.Stat(sessionsDir)
	return err == nil && info.IsDir()
}

var _ = time.Time{}

package windsurf

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/agentop-dev/agentop/internal/adapter"
)

var _ adapter.Adapter = (*Adapter)(nil)

type Adapter struct{}

func (a *Adapter) ID() adapter.AgentID   { return adapter.AgentWindsurf }
func (a *Adapter) Name() string          { return "Windsurf/Devin Desktop" }
func (a *Adapter) DefaultDir() string    { return windsurfDefaultDir() }
func (a *Adapter) IsAvailable() bool     { return windsurfAvailable() }

// Discover walks both the Windsurf and Devin Desktop cascade directories and
// any implicit/ subdirectories for auto-archived sessions.
func (a *Adapter) Discover(dataDir string) ([]adapter.SessionFile, error) {
	var files []adapter.SessionFile

	// Legacy Windsurf path: ~/.codeium/windsurf/cascade/
	windsurfDir := filepath.Join(windsurfDefaultDir(), "windsurf", "cascade")
	sf, err := discoverDir(windsurfDir)
	if err == nil {
		files = append(files, sf...)
	}

	// Devin Desktop path: ~/.devin/desktop/cascade/
	devinDir := filepath.Join(devinDefaultDir(), "desktop", "cascade")
	sf, err = discoverDir(devinDir)
	if err == nil {
		files = append(files, sf...)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime.After(files[j].ModTime)
	})

	return files, nil
}

// ParseSession reads a .pb file and returns estimated events. Actual
// decryption is not feasible without the AES key from the language_server
// binary, so we estimate session data from file metadata.
func (a *Adapter) ParseSession(path string) (*adapter.ParseResult, error) {
	return parseSession(path)
}

// discoverDir walks a single cascade directory (including implicit/ subdir)
// and returns discovered .pb files.
func discoverDir(dir string) ([]adapter.SessionFile, error) {
	var files []adapter.SessionFile

	err := walkPBFiles(dir, func(path string, info os.FileInfo) {
		files = append(files, adapter.SessionFile{
			Path:      path,
			SessionID: sessionIDFromPath(path),
			ModTime:   info.ModTime(),
			AgentID:   adapter.AgentWindsurf,
		})
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}

// walkPBFiles recursively walks a directory collecting .pb files.
func walkPBFiles(dir string, fn func(string, os.FileInfo)) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())

		if entry.IsDir() {
			if entry.Name() == "implicit" {
				_ = walkPBFiles(path, fn)
			}
			continue
		}

		if filepath.Ext(entry.Name()) != ".pb" {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		fn(path, info)
	}

	return nil
}

func windsurfDefaultDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".codeium"
	}
	return filepath.Join(home, ".codeium")
}

func devinDefaultDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".devin"
	}
	return filepath.Join(home, ".devin")
}

func windsurfAvailable() bool {
	dir := filepath.Join(windsurfDefaultDir(), "windsurf", "cascade")
	info, err := os.Stat(dir)
	if err == nil && info.IsDir() {
		return true
	}
	dir = filepath.Join(devinDefaultDir(), "desktop", "cascade")
	info, err = os.Stat(dir)
	return err == nil && info.IsDir()
}

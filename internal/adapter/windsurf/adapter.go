package windsurf

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/agentop-dev/agentop/internal/adapter"
)

var _ adapter.Adapter = (*Adapter)(nil)

// Adapter implements adapter.Adapter for Windsurf/Devin Desktop session files.
type Adapter struct{}

// ID returns the agent identifier.
func (a *Adapter) ID() adapter.AgentID { return adapter.AgentWindsurf }

// Name returns the display name of the agent.
func (a *Adapter) Name() string { return "Windsurf/Devin Desktop" }

// DefaultDir returns the default data directory for Windsurf/Devin Desktop.
func (a *Adapter) DefaultDir() string { return windsurfDefaultDir() }

// IsAvailable checks whether Windsurf/Devin Desktop session data exists.
func (a *Adapter) IsAvailable() bool { return windsurfAvailable() }

// Discover walks the Windsurf and Devin Desktop cascade + implicit directories
// for session .pb files.
func (a *Adapter) Discover(_ string) ([]adapter.SessionFile, error) {
	var files []adapter.SessionFile

	pairs := []struct{ base, sub string }{
		{windsurfDefaultDir(), "windsurf"},
		{devinDefaultDir(), "desktop"},
	}
	for _, p := range pairs {
		for _, subdir := range []string{"cascade", "implicit"} {
			dir := filepath.Join(p.base, p.sub, subdir)
			sf, err := discoverDir(dir)
			if err == nil {
				files = append(files, sf...)
			}
		}
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

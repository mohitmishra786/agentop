package cursor

import (
	"database/sql"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/agentop-dev/agentop/internal/adapter"
	_ "modernc.org/sqlite"
)

var _ adapter.Adapter = (*Adapter)(nil)

type Adapter struct{}

func (a *Adapter) ID() adapter.AgentID { return adapter.AgentCursor }
func (a *Adapter) Name() string        { return "Cursor" }
func (a *Adapter) DefaultDir() string  { return defaultDir() }
func (a *Adapter) IsAvailable() bool   { return available() }

func (a *Adapter) Discover(dataDir string) ([]adapter.SessionFile, error) {
	dbPath := filepath.Join(dataDir, "state.vscdb")
	info, err := os.Stat(dbPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = db.Close() }()

	rows, err := db.Query("SELECT [key] FROM cursorDiskKV WHERE [key] LIKE 'composer.%'")
	if err != nil {
		return nil, nil
	}
	defer func() { _ = rows.Close() }()

	var files []adapter.SessionFile
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			continue
		}
		composerID := strings.TrimPrefix(key, "composer.")
		if composerID == "" {
			continue
		}
		files = append(files, adapter.SessionFile{
			Path:      dbPath + pathSep + composerID,
			SessionID: composerID,
			ModTime:   info.ModTime(),
			AgentID:   adapter.AgentCursor,
		})
	}
	return files, nil
}

func (a *Adapter) ParseSession(path string) (*adapter.ParseResult, error) {
	return parseSessionFile(path)
}

func defaultDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "/tmp"
	}
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "Cursor", "User", "globalStorage")
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData != "" {
			return filepath.Join(appData, "Cursor", "User", "globalStorage")
		}
		return filepath.Join(home, "AppData", "Roaming", "Cursor", "User", "globalStorage")
	default:
		return filepath.Join(home, ".config", "Cursor", "User", "globalStorage")
	}
}

func available() bool {
	dbPath := filepath.Join(defaultDir(), "state.vscdb")
	info, err := os.Stat(dbPath)
	return err == nil && !info.IsDir()
}

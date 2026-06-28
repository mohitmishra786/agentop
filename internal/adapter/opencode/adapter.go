package opencode

import (
	"database/sql"
	"os"
	"path/filepath"
	"time"

	"github.com/agentop-dev/agentop/internal/adapter"
	_ "modernc.org/sqlite"
)

var _ adapter.Adapter = (*Adapter)(nil)

type Adapter struct{}

func (a *Adapter) ID() adapter.AgentID { return adapter.AgentOpenCode }
func (a *Adapter) Name() string        { return "OpenCode" }
func (a *Adapter) DefaultDir() string  { return defaultDir() }
func (a *Adapter) IsAvailable() bool   { return available() }

func (a *Adapter) Discover(dataDir string) ([]adapter.SessionFile, error) {
	dbPath := filepath.Join(dataDir, "opencode.db")

	_, err := os.Stat(dbPath)
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

	rows, err := db.Query("SELECT id, directory, time_created FROM session ORDER BY time_created DESC")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var files []adapter.SessionFile
	for rows.Next() {
		var id, directory string
		var timeCreated int64
		if err := rows.Scan(&id, &directory, &timeCreated); err != nil {
			continue
		}

		encodedPath := dbPath + "#" + id
		modTime := time.UnixMilli(timeCreated)

		files = append(files, adapter.SessionFile{
			Path:        encodedPath,
			ProjectHash: directory,
			SessionID:   id,
			ModTime:     modTime,
			AgentID:     adapter.AgentOpenCode,
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
		return ".local/share/opencode"
	}
	return filepath.Join(home, ".local", "share", "opencode")
}

func available() bool {
	dbPath := filepath.Join(defaultDir(), "opencode.db")
	info, err := os.Stat(dbPath)
	return err == nil && !info.IsDir()
}

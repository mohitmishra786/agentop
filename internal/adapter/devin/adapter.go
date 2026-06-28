// Package devin implements the adapter for Devin Desktop sessions stored in
// ~/.local/share/devin/cli/sessions.db (SQLite).
package devin

import (
	"database/sql"
	"os"
	"path/filepath"
	"time"

	"github.com/agentop-dev/agentop/internal/adapter"
	_ "modernc.org/sqlite" // SQLite driver
)

var _ adapter.Adapter = (*Adapter)(nil)

// Adapter discovers and parses Devin Desktop sessions.
type Adapter struct{}

// ID returns the agent identifier for Devin Desktop.
func (a *Adapter) ID() adapter.AgentID { return adapter.AgentDevin }

// Name returns the human-readable name for Devin Desktop.
func (a *Adapter) Name() string { return "Devin Desktop" }

// DefaultDir returns the default data directory for Devin Desktop.
func (a *Adapter) DefaultDir() string { return defaultDir() }

// IsAvailable checks whether Devin Desktop session data exists.
func (a *Adapter) IsAvailable() bool { return available() }

// Discover finds Devin Desktop sessions in the SQLite database.
func (a *Adapter) Discover(dataDir string) ([]adapter.SessionFile, error) {
	dbPath := filepath.Join(dataDir, "cli", "sessions.db")

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

	rows, err := db.Query("SELECT id, COALESCE(title, ''), created_at, working_directory FROM sessions WHERE hidden = 0 ORDER BY last_activity_at DESC")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var files []adapter.SessionFile
	for rows.Next() {
		var id, title, dir string
		var createdAt int64
		if err := rows.Scan(&id, &title, &createdAt, &dir); err != nil {
			continue
		}

		encodedPath := dbPath + "#" + id
		projectHash := filepath.Base(dir)
		if projectHash == "." || projectHash == "" {
			projectHash = filepath.Base(filepath.Dir(dir))
		}

		files = append(files, adapter.SessionFile{
			Path:        encodedPath,
			ProjectHash: projectHash,
			SessionID:   id,
			ModTime:     time.Unix(createdAt, 0),
			AgentID:     adapter.AgentDevin,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return files, nil
}

// ParseSession reads a single Devin Desktop session from the SQLite database.
func (a *Adapter) ParseSession(path string) (*adapter.ParseResult, error) {
	return parseSession(path)
}

func defaultDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".local/share/devin"
	}
	return filepath.Join(home, ".local", "share", "devin")
}

func available() bool {
	dbPath := filepath.Join(defaultDir(), "cli", "sessions.db")
	info, err := os.Stat(dbPath)
	return err == nil && !info.IsDir()
}

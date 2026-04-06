package claude

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func ReadSessionsIndex(claudeDir, projectHash string) ([]SessionMeta, error) {
	path := filepath.Join(claudeDir, "projects", projectHash, "sessions-index.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var result struct {
		Sessions []SessionMeta `json:"sessions"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result.Sessions, nil
}

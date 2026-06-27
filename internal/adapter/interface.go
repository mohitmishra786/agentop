// Package adapter provides the Adapter interface for AI coding assistant
// session formats and a Registry to manage them.
package adapter

// Adapter is the interface that every AI coding assistant adapter must implement.
// Each adapter knows how to discover session files and parse them into a common
// Event representation.
type Adapter interface {
	// ID returns the unique agent identifier (e.g. "claude", "codex").
	ID() AgentID

	// Name returns a human-readable display name (e.g. "Claude Code").
	Name() string

	// Discover finds all session files for this agent under the given data
	// directory. It returns an error only for hard failures (permissions, I/O);
	// missing directories are not errors — return an empty slice instead.
	Discover(dataDir string) ([]SessionFile, error)

	// DefaultDir returns the default data directory for this agent.
	DefaultDir() string

	// ParseSession reads a session file at the given path and returns the
	// parsed events with optional metadata.
	ParseSession(path string) (*ParseResult, error)

	// IsAvailable returns true if this agent appears to be installed (i.e. its
	// default data directory exists).
	IsAvailable() bool
}

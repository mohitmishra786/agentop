// Package adapter provides a registry of AI coding assistant adapters and
// utilities for discovering installed agents and dispatching to the correct
// adapter.
package adapter

import (
	"fmt"
	"sort"
	"strings"
)

// Registry holds all registered adapters and provides discovery over all of
// them.
type Registry struct {
	adapters map[AgentID]Adapter
	order    []AgentID
}

// NewRegistry creates a new Registry and registers all built-in adapters.
func NewRegistry() *Registry {
	return &Registry{
		adapters: make(map[AgentID]Adapter),
	}
}

// Register adds an adapter to the registry. If an adapter with the same ID
// already exists, it is replaced.
func (r *Registry) Register(a Adapter) {
	id := a.ID()
	if _, exists := r.adapters[id]; !exists {
		r.order = append(r.order, id)
	}
	r.adapters[id] = a
}

// Get returns the adapter for the given agent ID, or nil if not found.
func (r *Registry) Get(id AgentID) Adapter {
	return r.adapters[id]
}

// List returns all registered adapters sorted by ID.
func (r *Registry) List() []Adapter {
	sorted := make([]Adapter, 0, len(r.adapters))
	for _, id := range r.order {
		sorted = append(sorted, r.adapters[id])
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ID() < sorted[j].ID()
	})
	return sorted
}

// DiscoverAll runs Discover on every registered adapter and merges the
// results. Adapters whose data directory doesn't exist return empty results.
func (r *Registry) DiscoverAll() []SessionFile {
	var all []SessionFile
	for _, a := range r.List() {
		files, err := a.Discover(a.DefaultDir())
		if err != nil {
			continue
		}
		all = append(all, files...)
	}
	return all
}

// DiscoverSelected runs Discover only for the named agent IDs. Unknown IDs
// are silently skipped.
func (r *Registry) DiscoverSelected(ids []AgentID) []SessionFile {
	var all []SessionFile
	seen := make(map[AgentID]bool)
	for _, id := range ids {
		if seen[id] {
			continue
		}
		seen[id] = true
		a := r.Get(id)
		if a == nil {
			continue
		}
		files, err := a.Discover(a.DefaultDir())
		if err != nil {
			continue
		}
		all = append(all, files...)
	}
	return all
}

// DetectInstalledAgents checks every registered adapter to see which ones have
// data on disk. Returns the list of available agent IDs.
func (r *Registry) DetectInstalledAgents() []AgentID {
	var available []AgentID
	for _, a := range r.List() {
		if a.IsAvailable() {
			available = append(available, a.ID())
		}
	}
	return available
}

// ListAgentsText returns a formatted string listing all registered agents and
// their availability status.
func (r *Registry) ListAgentsText() string {
	var b strings.Builder
	b.WriteString("Registered agents:\n")
	for _, a := range r.List() {
		status := "not installed"
		if a.IsAvailable() {
			status = "installed"
		}
		fmt.Fprintf(&b, "  %-14s %s  (%s)\n", a.ID(), a.Name(), status)
	}
	return b.String()
}

// ParseAgentFlag parses the --agent flag value. It accepts comma-separated
// agent IDs, or "all" to select all available agents.
func ParseAgentFlag(raw string, r *Registry) []AgentID {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "all" {
		return r.DetectInstalledAgents()
	}
	var ids []AgentID
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			ids = append(ids, AgentID(part))
		}
	}
	return ids
}

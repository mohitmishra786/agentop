package aggregator

import (
	"encoding/json"
	"path/filepath"
	"sort"
	"time"

	"github.com/agentop-dev/agentop/internal/adapter"
	"github.com/agentop-dev/agentop/internal/claude"
	"github.com/agentop-dev/agentop/internal/pricing"
)

// TurnStat holds per-assistant-turn token metrics for cache, context-growth,
// and model-routing analysis.
type TurnStat struct {
	InputTokens  int64
	OutputTokens int64
	CacheCreate  int64
	CacheRead    int64
	Model        string
}

type SessionStats struct {
	ID          string
	ProjectHash string
	ProjectPath string
	ProjectName string
	Summary     string
	GitBranch   string
	Model       string
	AllModels   []string
	Entrypoint  string // "cli", "vscode", "jetbrains", etc.

	StartedAt time.Time
	EndedAt   time.Time
	Duration  time.Duration

	InputTokens       int64
	OutputTokens      int64
	CacheCreateTokens int64
	CacheReadTokens   int64
	TotalTokens       int64

	// Cache tier breakdown (5-min vs 1-hour cache)
	Cache5mTokens int64
	Cache1hTokens int64

	SubagentTokens int64
	SubagentCount  int

	CostUSD           float64
	CostUSDCalculated float64

	MessageCount int
	TurnCount    int
	ToolCalls    map[string]int

	CacheEfficiency float64
	CostPerMessage  float64
	BurnRate        float64
	WasCompacted    bool

	// Per-turn data for cache, context-growth, and model-routing analysis.
	// Ordered by turn index (first assistant response = index 0).
	Turns            []TurnStat
	Turn1CacheCreate int64          // cache_creation tokens on turn 1 (cold-start tax)
	FilesRead        map[string]int // Read tool file_path → invocation count
	// MaxToolResultBytes is the largest tool_result payload seen (bytes).
	// Divide by ~4 for a rough token estimate.
	MaxToolResultBytes int64

	AgentID     adapter.AgentID
	TokenSource adapter.TokenSource
}

func AggregateSession(events []claude.RawEvent, meta *claude.SessionMeta, pricer pricing.Pricer) *SessionStats {
	if len(events) == 0 {
		return nil
	}

	s := &SessionStats{
		ToolCalls: make(map[string]int),
		FilesRead: make(map[string]int),
	}

	s.ID = events[0].SessionID
	s.AgentID = events[0].AgentID

	// Determine overall token source: if any event has estimated tokens, mark as estimated.
	for _, e := range events {
		if e.TokenSrc == adapter.TokenEstimated {
			s.TokenSource = adapter.TokenEstimated
			break
		}
	}

	modelTokens := make(map[string]claude.Usage)

	type assistantEntry struct {
		msgID      string
		model      string
		usage      *claude.Usage
		stopReason string
		costUSD    float64
		turnIndex  int // insertion order for per-turn slice
	}
	assistantMap := make(map[string]assistantEntry)
	var turnCounter int

	var firstTime, lastTime time.Time

	for _, e := range events {
		if e.IsSidechain {
			continue
		}

		if firstTime.IsZero() || e.Timestamp.Before(firstTime) {
			firstTime = e.Timestamp
		}
		if e.Timestamp.After(lastTime) {
			lastTime = e.Timestamp
		}

		switch e.Type {
		case "human", "user":
			if e.UserType == "tool" {
				continue
			}
			s.MessageCount++
			s.TurnCount++
			if s.ProjectPath == "" && e.CWD != "" {
				s.ProjectPath = e.CWD
				s.ProjectName = filepath.Base(e.CWD)
			}
			if s.GitBranch == "" && e.GitBranch != "" {
				s.GitBranch = e.GitBranch
			}
			if s.Entrypoint == "" && e.Entrypoint != "" {
				s.Entrypoint = e.Entrypoint
			}

		case "assistant":
			s.TurnCount++
			if e.Message == nil {
				continue
			}

			msgID := e.Message.ID
			stopReason := e.Message.StopReason
			model := e.Message.Model
			if model == "" {
				model = "unknown"
			}

			existing, exists := assistantMap[msgID]
			if !exists {
				assistantMap[msgID] = assistantEntry{
					msgID:      msgID,
					model:      model,
					usage:      e.Message.Usage,
					stopReason: stopReason,
					costUSD:    e.CostUSD,
					turnIndex:  turnCounter,
				}
				turnCounter++
			} else {
				if stopReason != "" {
					existing.stopReason = stopReason
				}
				if e.Message.Usage != nil {
					existing.usage = e.Message.Usage
				}
				if e.CostUSD > 0 {
					existing.costUSD = e.CostUSD
				}
				if model != "unknown" {
					existing.model = model
				}
				assistantMap[msgID] = existing
			}

		case "tool", "tool_result":
			if e.ToolName != "" {
				s.ToolCalls[e.ToolName]++
			}
			// Parse file path for Read invocations.
			if e.Type == "tool" && e.ToolName == "Read" && len(e.ToolInput) > 0 {
				var inp struct {
					FilePath string `json:"file_path"`
				}
				if json.Unmarshal(e.ToolInput, &inp) == nil && inp.FilePath != "" {
					s.FilesRead[inp.FilePath]++
				}
			}
			// Track largest tool result as a token-bloat proxy.
			if e.Type == "tool_result" {
				if sz := int64(len(e.ToolResult)); sz > s.MaxToolResultBytes {
					s.MaxToolResultBytes = sz
				}
			}

		case "summary":
			if s.Summary == "" {
				s.Summary = e.Summary
			}
		}
	}

	// Sort assistant entries by insertion order so per-turn slice is chronological.
	sortedEntries := make([]assistantEntry, 0, len(assistantMap))
	for _, entry := range assistantMap {
		sortedEntries = append(sortedEntries, entry)
	}
	sort.Slice(sortedEntries, func(i, j int) bool {
		return sortedEntries[i].turnIndex < sortedEntries[j].turnIndex
	})

	for _, entry := range sortedEntries {
		if entry.stopReason == "" {
			continue
		}

		if entry.usage == nil {
			continue
		}

		usage := entry.usage
		model := entry.model

		s.InputTokens += int64(usage.InputTokens)
		s.OutputTokens += int64(usage.OutputTokens)
		s.CacheCreateTokens += int64(usage.CacheCreationInputTokens)
		s.CacheReadTokens += int64(usage.CacheReadInputTokens)
		if usage.CacheCreation != nil {
			s.Cache5mTokens += int64(usage.CacheCreation.Ephemeral5m)
			s.Cache1hTokens += int64(usage.CacheCreation.Ephemeral1h)
		}

		s.CostUSD += entry.costUSD

		existing := modelTokens[model]
		existing.InputTokens += usage.InputTokens
		existing.OutputTokens += usage.OutputTokens
		existing.CacheCreationInputTokens += usage.CacheCreationInputTokens
		existing.CacheReadInputTokens += usage.CacheReadInputTokens
		modelTokens[model] = existing

		// Build chronological per-turn slice.
		s.Turns = append(s.Turns, TurnStat{
			InputTokens:  int64(usage.InputTokens),
			OutputTokens: int64(usage.OutputTokens),
			CacheCreate:  int64(usage.CacheCreationInputTokens),
			CacheRead:    int64(usage.CacheReadInputTokens),
			Model:        model,
		})
	}

	if len(s.Turns) > 0 {
		s.Turn1CacheCreate = s.Turns[0].CacheCreate
	}

	s.TotalTokens = s.InputTokens + s.OutputTokens + s.CacheCreateTokens + s.CacheReadTokens
	s.StartedAt = firstTime
	s.EndedAt = lastTime
	s.Duration = lastTime.Sub(firstTime)

	denom := s.InputTokens + s.CacheCreateTokens + s.CacheReadTokens
	if denom > 0 {
		s.CacheEfficiency = float64(s.CacheReadTokens) / float64(denom)
	}

	for model, usage := range modelTokens {
		s.CostUSDCalculated += pricer.Calculate(usage, model)
	}

	if s.CostUSD <= 0 && s.CostUSDCalculated > 0 {
		s.CostUSD = s.CostUSDCalculated
	}

	models := make([]string, 0, len(modelTokens))
	for model := range modelTokens {
		models = append(models, model)
	}
	sort.Strings(models)

	s.AllModels = models
	maxTokens := 0
	for _, model := range models {
		usage := modelTokens[model]
		total := usage.InputTokens + usage.OutputTokens
		if total > maxTokens {
			maxTokens = total
			s.Model = model
		} else if total == 0 && maxTokens == 0 && s.Model == "" {
			s.Model = model
		}
	}

	if s.MessageCount > 0 {
		s.CostPerMessage = s.CostUSD / float64(s.MessageCount)
	}

	if s.Duration >= time.Minute {
		s.BurnRate = float64(s.TotalTokens) / s.Duration.Minutes()
	}

	if meta != nil {
		if s.Summary == "" {
			s.Summary = meta.Summary
		}
		if s.Summary == "" {
			s.Summary = meta.FirstUserMessage
		}
		if s.GitBranch == "" {
			s.GitBranch = meta.GitBranch
		}
		if s.ProjectPath == "" {
			s.ProjectPath = meta.CWD
		}
	}

	if len(s.Summary) > 80 {
		s.Summary = s.Summary[:77] + "..."
	}

	return s
}

func AggregateSubagents(subagentFiles []string, pricer pricing.Pricer) (int64, int, float64) {
	var totalTokens int64
	var count int
	var totalCost float64

	for _, f := range subagentFiles {
		events, err := claude.ParseSession(f)
		if err != nil {
			continue
		}

		stats := AggregateSession(events, nil, pricer)
		if stats == nil {
			continue
		}

		totalTokens += stats.TotalTokens
		count++
		totalCost += stats.CostUSD
	}

	return totalTokens, count, totalCost
}

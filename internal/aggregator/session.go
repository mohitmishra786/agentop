package aggregator

import (
	"path/filepath"
	"sort"
	"time"

	"github.com/agentop-dev/agentop/internal/claude"
	"github.com/agentop-dev/agentop/internal/pricing"
)

type SessionStats struct {
	ID          string
	ProjectHash string
	ProjectPath string
	ProjectName string
	Summary     string
	GitBranch   string
	Model       string
	AllModels   []string

	StartedAt time.Time
	EndedAt   time.Time
	Duration  time.Duration

	InputTokens       int64
	OutputTokens      int64
	CacheCreateTokens int64
	CacheReadTokens   int64
	TotalTokens       int64

	CostUSD           float64
	CostUSDCalculated float64

	MessageCount int
	TurnCount    int
	ToolCalls    map[string]int

	CacheEfficiency float64
	CostPerMessage  float64
	BurnRate        float64
	WasCompacted    bool
}

func AggregateSession(events []claude.RawEvent, meta *claude.SessionMeta, pricer pricing.Pricer) *SessionStats {
	if len(events) == 0 {
		return nil
	}

	s := &SessionStats{
		ToolCalls: make(map[string]int),
	}

	s.ID = events[0].SessionID

	modelTokens := make(map[string]claude.Usage)

	var firstTime, lastTime time.Time

	for _, e := range events {
		if firstTime.IsZero() || e.Timestamp.Before(firstTime) {
			firstTime = e.Timestamp
		}
		if e.Timestamp.After(lastTime) {
			lastTime = e.Timestamp
		}

		switch e.Type {
		case "human", "user":
			s.MessageCount++
			s.TurnCount++
			if s.ProjectPath == "" && e.CWD != "" {
				s.ProjectPath = e.CWD
				s.ProjectName = filepath.Base(e.CWD)
			}
			if s.GitBranch == "" && e.GitBranch != "" {
				s.GitBranch = e.GitBranch
			}

		case "assistant":
			s.TurnCount++
			if e.Message == nil {
				continue
			}

			usage := e.Message.Usage
			if usage == nil {
				continue
			}

			model := e.Message.Model
			if model == "" {
				model = "unknown"
			}

			s.InputTokens += int64(usage.InputTokens)
			s.OutputTokens += int64(usage.OutputTokens)
			s.CacheCreateTokens += int64(usage.CacheCreationInputTokens)
			s.CacheReadTokens += int64(usage.CacheReadInputTokens)

			s.CostUSD += e.CostUSD

			existing := modelTokens[model]
			existing.InputTokens += usage.InputTokens
			existing.OutputTokens += usage.OutputTokens
			existing.CacheCreationInputTokens += usage.CacheCreationInputTokens
			existing.CacheReadInputTokens += usage.CacheReadInputTokens
			modelTokens[model] = existing

		case "tool", "tool_result":
			if e.ToolName != "" {
				s.ToolCalls[e.ToolName]++
			}

		case "summary":
			if s.Summary == "" {
				s.Summary = e.Summary
			}
		}
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

	maxTokens := 0
	for model, usage := range modelTokens {
		if usage.InputTokens > maxTokens {
			maxTokens = usage.InputTokens
			s.Model = model
		}
		s.AllModels = append(s.AllModels, model)
	}
	sort.Strings(s.AllModels)

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

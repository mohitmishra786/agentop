package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/agentop-dev/agentop/internal/aggregator"
	"github.com/agentop-dev/agentop/internal/pricing"
	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "MCP server (Model Context Protocol)",
	Long:  `Starts a Model Context Protocol server over stdio for AI tools integration.`,
	RunE:  runMCP,
}

type jsonRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type jsonRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  interface{}     `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type toolCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type listSessionsArgs struct {
	Since   string `json:"since"`
	Project string `json:"project,omitempty"`
	Model   string `json:"model,omitempty"`
}

type getStatsArgs struct {
	Since string `json:"since"`
}

type searchSessionsArgs struct {
	Query string `json:"query"`
}

type sessionSummary struct {
	ID          string  `json:"id"`
	Project     string  `json:"project"`
	Summary     string  `json:"summary"`
	Model       string  `json:"model"`
	AgentID     string  `json:"agentId"`
	StartedAt   string  `json:"startedAt"`
	EndedAt     string  `json:"endedAt"`
	TotalTokens int64   `json:"totalTokens"`
	CostUSD     float64 `json:"costUSD"`
	Messages    int     `json:"messages"`
}

type listSessionsResult struct {
	Sessions []sessionSummary `json:"sessions"`
}

type getStatsResult struct {
	TotalSessions       int     `json:"totalSessions"`
	TotalTokens         int64   `json:"totalTokens"`
	InputTokens         int64   `json:"inputTokens"`
	OutputTokens        int64   `json:"outputTokens"`
	CacheCreateTokens   int64   `json:"cacheCreateTokens"`
	CacheReadTokens     int64   `json:"cacheReadTokens"`
	TotalCostUSD        float64 `json:"totalCostUSD"`
	AvgTokensPerSession float64 `json:"avgTokensPerSession"`
}

type agentEntry struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Available bool   `json:"available"`
}

type agentInfoResult struct {
	Agents []agentEntry `json:"agents"`
}

func runMCP(_ *cobra.Command, _ []string) error {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var req jsonRPCRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			continue
		}

		resp := handleMCPRequest(req)
		if resp != nil {
			data, err := json.Marshal(resp)
			if err != nil {
				continue
			}
			_, _ = fmt.Fprintln(os.Stdout, string(data))
		}
	}

	return scanner.Err()
}

func handleMCPRequest(req jsonRPCRequest) *jsonRPCResponse {
	switch req.Method {
	case "initialize":
		return handleInitialize(req)
	case "shutdown":
		return handleShutdown(req)
	case "exit":
		return handleExit(req)
	case "notifications/initialized", "notifications/cancelled", "notifications/progress":
		return nil
	case "resources/list":
		return mcpResult(req, map[string]interface{}{"resources": []interface{}{}})
	case "resources/read":
		return mcpResult(req, map[string]interface{}{"contents": []interface{}{}})
	case "tools/call":
		return handleToolCall(req)
	default:
		return mcpError(req, -32601, "Method not found")
	}
}

func handleInitialize(req jsonRPCRequest) *jsonRPCResponse {
	result := map[string]interface{}{
		"protocolVersion": "2025-03-26",
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{},
		},
		"serverInfo": map[string]interface{}{
			"name":    "agentop",
			"version": Version,
		},
	}
	return mcpResult(req, result)
}

func handleShutdown(req jsonRPCRequest) *jsonRPCResponse {
	return mcpResult(req, map[string]interface{}{})
}

func handleExit(req jsonRPCRequest) *jsonRPCResponse {
	resp := mcpResult(req, map[string]interface{}{})
	go func() {
		time.Sleep(50 * time.Millisecond)
		os.Exit(0)
	}()
	return resp
}

func handleToolCall(req jsonRPCRequest) *jsonRPCResponse {
	var params toolCallParams
	if len(req.Params) == 0 || json.Unmarshal(req.Params, &params) != nil {
		return mcpError(req, -32602, "Invalid tool call params")
	}

	switch params.Name {
	case "list_sessions":
		return handleListSessions(req, params.Arguments)
	case "get_stats":
		return handleGetStats(req, params.Arguments)
	case "search_sessions":
		return handleSearchSessions(req, params.Arguments)
	case "get_agent_info":
		return handleGetAgentInfo(req, params.Arguments)
	default:
		return mcpError(req, -32602, fmt.Sprintf("Unknown tool: %s", params.Name))
	}
}

func handleListSessions(req jsonRPCRequest, rawArgs json.RawMessage) *jsonRPCResponse {
	var args listSessionsArgs
	if len(rawArgs) > 0 {
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return mcpError(req, -32602, "Invalid arguments for list_sessions")
		}
	}
	if args.Since == "" {
		args.Since = "30d"
	}

	sessions, err := loadMCPSessions(args.Since, args.Project, args.Model)
	if err != nil {
		return mcpError(req, -32603, fmt.Sprintf("Failed to load sessions: %v", err))
	}

	var summaries []sessionSummary
	for _, s := range sessions {
		project := s.ProjectName
		if project == "" {
			project = s.ProjectPath
		}
		summaries = append(summaries, sessionSummary{
			ID:          s.ID,
			Project:     project,
			Summary:     s.Summary,
			Model:       s.Model,
			AgentID:     string(s.AgentID),
			StartedAt:   s.StartedAt.Format(time.RFC3339),
			EndedAt:     s.EndedAt.Format(time.RFC3339),
			TotalTokens: s.TotalTokens,
			CostUSD:     s.CostUSD,
			Messages:    s.MessageCount,
		})
	}

	return mcpResult(req, listSessionsResult{Sessions: summaries})
}

func handleGetStats(req jsonRPCRequest, rawArgs json.RawMessage) *jsonRPCResponse {
	var args getStatsArgs
	if len(rawArgs) > 0 {
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return mcpError(req, -32602, "Invalid arguments for get_stats")
		}
	}
	if args.Since == "" {
		args.Since = "30d"
	}

	sessions, err := loadMCPSessions(args.Since, "", "")
	if err != nil {
		return mcpError(req, -32603, fmt.Sprintf("Failed to load sessions: %v", err))
	}

	result := getStatsResult{}
	for _, s := range sessions {
		result.TotalSessions++
		result.TotalTokens += s.TotalTokens
		result.InputTokens += s.InputTokens
		result.OutputTokens += s.OutputTokens
		result.CacheCreateTokens += s.CacheCreateTokens
		result.CacheReadTokens += s.CacheReadTokens
		result.TotalCostUSD += s.CostUSD
	}
	if result.TotalSessions > 0 {
		result.AvgTokensPerSession = float64(result.TotalTokens) / float64(result.TotalSessions)
	}

	return mcpResult(req, result)
}

func handleSearchSessions(req jsonRPCRequest, rawArgs json.RawMessage) *jsonRPCResponse {
	var args searchSessionsArgs
	if len(rawArgs) > 0 {
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return mcpError(req, -32602, "Invalid arguments for search_sessions")
		}
	}
	if args.Query == "" {
		return mcpError(req, -32602, "query is required")
	}

	query := strings.ToLower(args.Query)

	sessions, err := loadMCPSessions("365d", "", "")
	if err != nil {
		return mcpError(req, -32603, fmt.Sprintf("Failed to load sessions: %v", err))
	}

	var matches []sessionSummary
	for _, s := range sessions {
		if s.Summary == "" && s.ProjectName == "" && s.ProjectPath == "" {
			continue
		}
		if strings.Contains(strings.ToLower(s.Summary), query) ||
			strings.Contains(strings.ToLower(s.ProjectName), query) ||
			strings.Contains(strings.ToLower(s.ProjectPath), query) ||
			strings.Contains(strings.ToLower(s.Model), query) ||
			strings.Contains(strings.ToLower(string(s.AgentID)), query) {
			project := s.ProjectName
			if project == "" {
				project = s.ProjectPath
			}
			matches = append(matches, sessionSummary{
				ID:          s.ID,
				Project:     project,
				Summary:     s.Summary,
				Model:       s.Model,
				AgentID:     string(s.AgentID),
				StartedAt:   s.StartedAt.Format(time.RFC3339),
				EndedAt:     s.EndedAt.Format(time.RFC3339),
				TotalTokens: s.TotalTokens,
				CostUSD:     s.CostUSD,
				Messages:    s.MessageCount,
			})
		}
	}

	return mcpResult(req, listSessionsResult{Sessions: matches})
}

func handleGetAgentInfo(req jsonRPCRequest, _ json.RawMessage) *jsonRPCResponse {
	adapters := registry.List()
	var entries []agentEntry
	for _, a := range adapters {
		entries = append(entries, agentEntry{
			ID:        string(a.ID()),
			Name:      a.Name(),
			Available: a.IsAvailable(),
		})
	}
	return mcpResult(req, agentInfoResult{Agents: entries})
}

func loadMCPSessions(sinceStr, projectFilter, modelFilter string) ([]*aggregator.SessionStats, error) {
	agentIDs := resolveAgentIDs()
	files := registry.DiscoverSelected(agentIDs)

	pricer := pricing.DefaultPricer{}
	var sessions []*aggregator.SessionStats

	for _, f := range files {
		ad := registry.Get(f.AgentID)
		if ad == nil {
			continue
		}
		result, err := ad.ParseSession(f.Path)
		if err != nil {
			continue
		}
		stats := aggregator.AggregateSession(result.Events, result.Meta, pricer)
		if stats == nil {
			continue
		}
		stats.ProjectHash = f.ProjectHash
		if stats.ID == "" {
			stats.ID = f.SessionID
		}
		if len(f.SubagentFiles) > 0 {
			subTokens, subCount, subCost := aggregator.AggregateSubagents(f.SubagentFiles, pricer)
			stats.SubagentTokens = subTokens
			stats.SubagentCount = subCount
			stats.CostUSD += subCost
			stats.TotalTokens += subTokens
		}
		sessions = append(sessions, stats)
	}

	startTime, err := parseSince(sinceStr)
	if err != nil {
		startTime = time.Time{}
	}

	var filtered []*aggregator.SessionStats
	for _, s := range sessions {
		if !s.EndedAt.IsZero() && !startTime.IsZero() && s.EndedAt.Before(startTime) {
			continue
		}
		if projectFilter != "" && s.ProjectPath != "" {
			matched := false
			for _, p := range []string{s.ProjectPath, s.ProjectName} {
				if containsFold(p, projectFilter) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}
		if modelFilter != "" && s.Model != "" {
			if !containsFold(s.Model, modelFilter) {
				continue
			}
		}
		filtered = append(filtered, s)
	}

	return filtered, nil
}

func mcpResult(req jsonRPCRequest, result interface{}) *jsonRPCResponse {
	return &jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

func mcpError(req jsonRPCRequest, code int, message string) *jsonRPCResponse {
	return &jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Error: &rpcError{
			Code:    code,
			Message: message,
		},
	}
}

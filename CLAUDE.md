# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Project Is

`agentop` is a universal terminal dashboard for AI coding assistants. It discovers session data from all major AI coding tools (Claude Code, Codex CLI, Cursor, Copilot CLI, Gemini CLI, Kiro, OpenCode, Continue, JetBrains, and Windsurf/Devin Desktop), aggregates token usage, computes costs, and renders an interactive terminal UI. It is a pure Go binary with no runtime dependencies and makes no network calls.

## Commands

```bash
make build        # Build to ./bin/agentop (injects version/commit via ldflags)
make test         # Run all tests with -v
make run          # go run .
make run-watch    # go run . --watch
make lint         # golangci-lint run
make install      # Install to $GOPATH/bin
make clean        # Remove ./bin and ./dist
```

Run a single test:
```bash
go test ./internal/aggregator/... -run TestFunctionName -v
```

## Architecture

Data flows through four distinct layers:

1. **Discovery & Parsing** (`internal/adapter/`): `Discover()` recursively finds session files in agent-specific directories. `ParseSession()` reads and unmarshals events. Each agent has a dedicated adapter (`internal/adapter/<agent-id>/`). Key types are in `types.go`.

2. **Aggregation** (`internal/aggregator/session.go`): `AggregateSession()` deduplicates events by message ID (handles streaming — only the final chunk counts), filters out sidechain events and tool-type users, and computes `SessionStats` (token counts, cost, cache efficiency %, burn rate).

3. **Pricing** (`internal/pricing/`): `DefaultPricer.Calculate()` uses embedded `pricing.json` (rates per million tokens for input/output/cache). Falls back to claude-sonnet-4-6 rates when a model is unknown.

4. **UI** (`internal/ui/`): BubbleTea (`app.go`) handles keyboard input and watch mode. Lipgloss (`styles.go`) handles themes (dark/light/ansi) and model badge colors. `table.go` and `sparkline.go` handle specific visualization components.

**CLI** (`cmd/`): Each subcommand (`today`, `daily`, `monthly`, `blocks`, `session`, `doctor`, `config`) lives in its own file. Global flags (claude-dir, since, until, project, model, json, watch, refresh, theme, layout) are registered in `root.go`.

## All Commands

| Command | Description |
|---------|-------------|
| `agentop` (default) | Show today's sessions (runs `today`) |
| `agentop today` | Today's sessions with token bars and agent badges |
| `agentop daily` | Daily historical report |
| `agentop monthly` | Monthly summaries |
| `agentop session <id>` | Deep-dive into a single session |
| `agentop doctor [id]` | Anomaly detection and optimization insights (multi-agent aware) |
| `agentop config` | Show pricing config and effective rates |
| `agentop blocks` | 5-hour billing window report |
| `agentop budget` | Monthly budget tracking with config file at `~/.config/agentop/budget.json` |
| `agentop mcp` | MCP server over stdio (JSON-RPC 2.0) |

## All Flags

| Flag | Description |
|------|-------------|
| `--agent` | Filter by agent(s): `all`, `claude`, `codex`, etc. (comma-separated for multiple) |
| `--list-agents` | List all registered agents and availability |
| `--since` | Date filter: `today`, `7d`, `30d`, `2026-04-01` |
| `--project` | Filter by project path (partial match) |
| `--model` | Filter by model name (partial match) |
| `--json` | Output as JSON |
| `--watch` / `-w` | Live refresh mode (Bubbletea TUI) |
| `--layout` | Layout style: `default`, `table` |
| `--theme` | Color theme: `dark`, `light`, `ansi` |

## Key Design Details

- **Message deduplication**: Streamed messages produce multiple JSONL lines with the same `messageId`. Only the final chunk (identified by specific fields) should be counted to avoid inflating token totals. See `internal/aggregator/session.go`.
- **Pricing fallback**: When `usage.cost` is missing from the JSONL (older sessions), cost is calculated from token counts using `internal/pricing/`. When present, it reports both the recorded cost and a recalculated cost.
- **Windsurf acquired by Devin (July 2025):** Cognition (Devin) acquired Codeium (Windsurf). In June 2026 Windsurf was relaunched as Devin Desktop. The .pb encrypted protobuf format is still in use under the Devin brand. The adapter should be named `internal/adapter/windsurf/` for backward compatibility.
- **No tests for UI layer** — only `internal/aggregator`, `internal/pricing`, and `internal/adapter/*` have unit tests; testdata lives in `testdata/sessions/` and `testdata/adapters/`.
- **Adding a new model**: Update `internal/pricing/pricing.json` with rates (USD per million tokens).
- **Adding a new command**: Create `cmd/<name>.go`, define a `cobra.Command`, and register it in `cmd/root.go`'s `init()`.
- **Lint config** (`.golangci.yml`): enabled linters are `govet`, `revive`, `staticcheck`, `unused`; formatters are `gofmt` and `goimports`. Tests are excluded from linting.

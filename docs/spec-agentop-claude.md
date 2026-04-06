# agentop — Technical Specification (Claude Code Edition)
## Version 1.0 · April 2026

---

## 1. Overview

`agentop` is a zero-config terminal dashboard for Claude Code sessions. It reads local JSONL transcripts from `~/.claude/projects/`, aggregates token usage, computes costs and cache efficiency, and renders a `duf`-style panel UI in the terminal — one panel per tool, one row per session, visual bars for token composition, color-coded health indicators.

The name is a portmanteau of "agent" + "top" (like `htop`, `nvtop`, `duf`).

**Tagline:** _"See exactly what your Claude sessions are costing you — at a glance."_

---

## 2. Language & Stack

### Choice: Go 1.22+

**Why Go:**
- Single static binary → `brew install`, `apt`, `scoop`, `winget`, `curl | sh` — zero runtime dependencies for users
- The Charm ecosystem (`bubbletea` + `lipgloss` + `bubbles`) is the gold standard for terminal UI in any language, and it's Go-native
- `duf` is in Go — same look, same distribution patterns, proven approach
- Cold start < 30ms (vs 500ms+ for Node/Deno, 200ms+ for Python)
- `goreleaser` automates multi-platform binary releases and Homebrew tap in one config
- JSONL streaming with `bufio.Scanner` is idiomatic, efficient, and handles 100MB files without loading into memory
- `fsnotify` for file watching, `modernc.org/sqlite` for optional caching (pure Go, no CGO)

### Core Dependencies

```
github.com/charmbracelet/bubbletea    v1.x   — TUI event loop and model
github.com/charmbracelet/lipgloss     v1.x   — terminal styling and layout
github.com/charmbracelet/bubbles      v0.x   — progress bar, spinner, viewport, table
github.com/charmbracelet/harmonica    v0.x   — spring animations (optional)
github.com/spf13/cobra                v1.x   — CLI commands and flags
github.com/fsnotify/fsnotify          v1.x   — file system watcher for --watch
modernc.org/sqlite                    v1.x   — pure-Go SQLite for optional index cache
github.com/dustin/go-humanize         v1.x   — human-readable numbers (1.3M, 2h 18m)
github.com/adrg/xdg                   v0.x   — XDG base directories (config, cache paths)
```

### Distribution

```yaml
# .goreleaser.yml produces:
# - macOS arm64 + amd64 (universal binary option)
# - Linux arm64 + amd64
# - Windows amd64
# - .deb, .rpm, .apk packages
# - Homebrew tap formula auto-updated on release
# - checksums.txt + GPG signature

brews:
  - name: agentop
    tap:
      owner: yourname
      name: homebrew-tap
    homepage: https://github.com/yourname/agentop
    description: "duf-style terminal dashboard for Claude Code sessions"
```

**Install methods:**
```sh
# Homebrew (macOS + Linux)
brew install yourname/tap/agentop

# One-liner (all platforms)
curl -sL https://agentop.sh/install | sh

# Go install (Go users)
go install github.com/yourname/agentop@latest

# GitHub releases (direct binary)
# https://github.com/yourname/agentop/releases
```

---

## 3. Project Structure

```
agentop/
├── main.go                      # Entry point, cobra root setup
├── go.mod
├── go.sum
├── .goreleaser.yml
├── Makefile
├── README.md
│
├── cmd/
│   ├── root.go                  # Root command, global flags, shared init
│   ├── today.go                 # Default view (today's sessions)
│   ├── daily.go                 # Daily report (last 7d or --since/--until)
│   ├── monthly.go               # Monthly aggregated report
│   ├── session.go               # Deep-dive single session by ID or index
│   ├── blocks.go                # 5-hour billing window report
│   ├── doctor.go                # Anomaly detection and insights
│   └── config.go                # Show/edit config, pricing cache info
│
├── internal/
│   ├── claude/
│   │   ├── discover.go          # Scan ~/.claude/projects/ for all sessions
│   │   ├── jsonl.go             # JSONL parser with streaming support
│   │   ├── schema.go            # Go structs matching Claude JSONL format
│   │   ├── index.go             # Read sessions-index.json metadata
│   │   └── history.go           # Read ~/.claude/history.jsonl (optional)
│   │
│   ├── aggregator/
│   │   ├── session.go           # Per-session aggregation (tokens, cost, tools)
│   │   ├── daily.go             # Group sessions by calendar day
│   │   ├── blocks.go            # 5-hour UTC billing block calculation
│   │   ├── monthly.go           # Monthly summaries
│   │   └── doctor.go            # Anomaly detection algorithms
│   │
│   ├── pricing/
│   │   ├── models.go            # Pricing structs and lookup
│   │   ├── embedded.go          # go:embed snapshot of pricing JSON
│   │   └── fetcher.go           # Optional fetch from LiteLLM API
│   │
│   ├── ui/
│   │   ├── app.go               # Top-level bubbletea app model
│   │   ├── today.go             # Today view model
│   │   ├── daily.go             # Daily view model
│   │   ├── session_detail.go    # Session drill-down model
│   │   ├── blocks.go            # Blocks view model
│   │   ├── doctor.go            # Doctor view model
│   │   ├── styles.go            # All lipgloss styles, colors, constants
│   │   ├── bars.go              # Token bar rendering functions
│   │   ├── panels.go            # Section panel (border + header) component
│   │   ├── summary.go           # Summary strip component
│   │   ├── table.go             # Grid row rendering helpers
│   │   └── helpers.go           # Truncate, pad, format numbers
│   │
│   ├── watcher/
│   │   └── watcher.go           # fsnotify watcher → message channel
│   │
│   ├── cache/
│   │   └── sqlite.go            # Optional SQLite index for fast startup
│   │
│   └── config/
│       └── config.go            # User config (~/.config/agentop/config.toml)
│
└── assets/
    └── pricing.json             # Embedded model pricing snapshot
```

---

## 4. Data Sources & JSONL Schema

### 4.1 Directory Layout

```
~/.claude/
├── history.jsonl                    # Global prompt history (all sessions, all projects)
└── projects/
    └── <project-hash>/              # SHA256 of absolute project path
        ├── sessions-index.json      # Metadata for all sessions in this project
        └── sessions/
            └── <session-uuid>.jsonl # Full transcript per session
```

The `<project-hash>` is a deterministic hash of the project's absolute path. `agentop` does not need to know the hash function — it scans all subdirectories of `~/.claude/projects/`.

### 4.2 Session JSONL Schema

Each line in a `.jsonl` file is one of several event types:

**Human message:**
```json
{
  "type": "human",
  "message": { "role": "user", "content": "Fix the auth bug" },
  "sessionId": "a1b2c3d4-e5f6-...",
  "uuid": "msg-uuid-1",
  "parentUuid": null,
  "timestamp": "2026-04-07T10:14:23.123Z",
  "cwd": "/Users/mohit/work/acme-app",
  "isSidechain": false
}
```

**Assistant message (contains token counts):**
```json
{
  "type": "assistant",
  "message": {
    "id": "msg_01abc...",
    "role": "assistant",
    "content": [{ "type": "text", "text": "Let me look at the auth module..." }],
    "model": "claude-opus-4-6",
    "stop_reason": "end_turn",
    "usage": {
      "input_tokens": 8500,
      "output_tokens": 1200,
      "cache_creation_input_tokens": 5000,
      "cache_read_input_tokens": 78000
    }
  },
  "costUSD": 0.01234,
  "sessionId": "a1b2c3d4-e5f6-...",
  "uuid": "msg-uuid-2",
  "parentUuid": "msg-uuid-1",
  "timestamp": "2026-04-07T10:14:25.456Z"
}
```

**Tool use (within assistant content array or standalone):**
```json
{
  "type": "tool",
  "toolName": "Read",
  "toolInput": { "file_path": "src/auth.ts" },
  "sessionId": "a1b2c3d4-e5f6-...",
  "uuid": "tool-uuid-1",
  "parentUuid": "msg-uuid-2",
  "timestamp": "2026-04-07T10:14:25.500Z"
}
```

**Tool result:**
```json
{
  "type": "tool_result",
  "toolName": "Read",
  "toolResult": { "output": "export function validateToken..." },
  "sessionId": "a1b2c3d4-e5f6-...",
  "uuid": "tool-result-uuid-1",
  "timestamp": "2026-04-07T10:14:25.600Z"
}
```

**Summary (auto-generated at session end):**
```json
{
  "type": "summary",
  "summary": "Refactored authentication module to use JWT...",
  "sessionId": "a1b2c3d4-e5f6-...",
  "timestamp": "2026-04-07T12:32:11.000Z"
}
```

### 4.3 sessions-index.json Schema

```json
{
  "sessions": [
    {
      "id": "a1b2c3d4-e5f6-...",
      "summary": "Refactored authentication module to use JWT...",
      "firstUserMessage": "Fix the auth bug",
      "messageCount": 47,
      "createdAt": "2026-04-07T10:14:23Z",
      "updatedAt": "2026-04-07T12:32:11Z",
      "gitBranch": "feature/auth-refactor",
      "cwd": "/Users/mohit/work/acme-app",
      "projectPath": "/Users/mohit/work/acme-app"
    }
  ]
}
```

### 4.4 Go Structs (schema.go)

```go
package claude

import "time"

// RawEvent is the top-level JSONL line
type RawEvent struct {
    Type      string          `json:"type"`
    Message   *RawMessage     `json:"message,omitempty"`
    ToolName  string          `json:"toolName,omitempty"`
    ToolInput json.RawMessage `json:"toolInput,omitempty"`
    Summary   string          `json:"summary,omitempty"`
    SessionID string          `json:"sessionId"`
    UUID      string          `json:"uuid"`
    ParentUUID string         `json:"parentUuid,omitempty"`
    Timestamp  time.Time      `json:"timestamp"`
    CostUSD    float64        `json:"costUSD,omitempty"`
    CWD        string         `json:"cwd,omitempty"`
}

type RawMessage struct {
    ID       string     `json:"id"`
    Role     string     `json:"role"`
    Model    string     `json:"model,omitempty"`
    Usage    *Usage     `json:"usage,omitempty"`
    Content  []Content  `json:"content,omitempty"`
}

type Usage struct {
    InputTokens              int `json:"input_tokens"`
    OutputTokens             int `json:"output_tokens"`
    CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
    CacheReadInputTokens     int `json:"cache_read_input_tokens"`
}

type Content struct {
    Type string `json:"type"` // "text", "tool_use", "tool_result", "thinking"
    Text string `json:"text,omitempty"`
}
```

### 4.5 Parser Implementation Notes

```go
// jsonl.go - streaming parser (handles files > 100MB)
func ParseSession(path string) ([]RawEvent, error) {
    f, err := os.Open(path)
    if err != nil { return nil, err }
    defer f.Close()

    var events []RawEvent
    scanner := bufio.NewScanner(f)
    scanner.Buffer(make([]byte, 10*1024*1024), 10*1024*1024) // 10MB line buffer
    for scanner.Scan() {
        line := scanner.Bytes()
        if len(line) == 0 { continue }
        var e RawEvent
        if err := json.Unmarshal(line, &e); err != nil {
            continue // skip malformed lines (tool results can be large)
        }
        events = append(events, e)
    }
    return events, scanner.Err()
}
```

**Edge cases to handle:**
- Lines with invalid JSON (tool results with binary output) — skip silently
- Very large lines (file contents in tool results) — 10MB scanner buffer
- Files currently being written (session in progress) — last line may be truncated, skip
- Empty files — return empty slice, not error
- UTF-8 encoding issues — treat as bytes
- Duplicate `uuid` values across sessions (should not happen but check)
- `costUSD` may be 0.0 for old sessions (before Claude added it) — calculate from tokens

---

## 5. Aggregation Logic

### 5.1 SessionStats (the core computed unit)

```go
type SessionStats struct {
    // Identity
    ID          string
    ProjectPath string
    ProjectHash string    // directory name in ~/.claude/projects/
    Summary     string    // Claude-generated or first user message
    GitBranch   string
    Model       string    // dominant model (most used)
    Models      []string  // all models used (if switched mid-session)

    // Timing
    StartedAt  time.Time
    EndedAt    time.Time
    Duration   time.Duration

    // Token counts (accumulated across all assistant messages)
    InputTokens       int64
    OutputTokens      int64
    CacheCreateTokens int64
    CacheReadTokens   int64
    TotalTokens       int64

    // Cost
    CostUSD           float64 // from costUSD fields (preferred)
    CostUSDCalculated float64 // computed from token counts × pricing

    // Engagement
    MessageCount int // human turns only
    TurnCount    int // total turns (human + assistant)
    ToolCalls    map[string]int // toolName → count
    LinesAdded   int
    LinesRemoved int

    // Derived metrics
    CacheEfficiency float64 // cache_read / (input + cache_create + cache_read)
    CostPerMessage  float64
    TokensPerMinute float64
    WasCompacted    bool    // true if /compact was invoked
    ContextPeakPct  float64 // highest context window fill % seen
}

func (s *SessionStats) HealthColor() string {
    switch {
    case s.CacheEfficiency >= 0.80: return ColorGreen   // "#50c87a"
    case s.CacheEfficiency >= 0.40: return ColorAmber   // "#e5a040"
    default:                         return ColorRed     // "#e05050"
    }
}
```

### 5.2 Cache Efficiency Formula

```
CacheEfficiency = CacheReadTokens / (InputTokens + CacheCreateTokens + CacheReadTokens)

Example (good):    80000 / (500 + 8000 + 80000) = 91.1%  → green
Example (bad):     500 / (8300 + 2400 + 500)    = 4.5%   → red
Example (new):     0 / (1000 + 15000 + 0)       = 0%     → red (expected for brand new session)
```

Note: A 0% efficiency on a session's very first turn is expected (cache is being built). Flag sessions where the *entire* session has low efficiency as anomalous.

### 5.3 5-Hour Billing Block Calculation

```go
// Blocks are fixed UTC boundaries: 00:00, 05:00, 10:00, 15:00, 20:00
func BlockStart(t time.Time) time.Time {
    utc := t.UTC()
    hour := utc.Hour()
    blockHour := (hour / 5) * 5
    return time.Date(utc.Year(), utc.Month(), utc.Day(), blockHour, 0, 0, 0, time.UTC)
}

func BlockEnd(start time.Time) time.Time {
    return start.Add(5 * time.Hour)
}

func IsActiveBlock(start time.Time) bool {
    now := time.Now().UTC()
    return now.After(start) && now.Before(BlockEnd(start))
}

func TimeRemainingInBlock(start time.Time) time.Duration {
    return BlockEnd(start).Sub(time.Now().UTC())
}
```

### 5.4 Anomaly Detection (doctor.go)

```go
type Anomaly struct {
    Severity    string // "warn", "info", "tip"
    SessionID   string
    ProjectName string
    Code        string // machine-readable code
    Title       string
    Detail      string
}

// Rules:
// CACHE_MISS_HIGH:    session CacheEfficiency < 15% AND TotalTokens > 500k
//                     → "Low cache hit rate — possible CLAUDE.md cold start or new session"
// SHORT_SESSION_COST: MessageCount < 5 AND CostUSD > 3.0
//                     → "High cost for a short session — large file reads or Opus on simple tasks?"
// CONTEXT_NEAR_FULL:  ContextPeakPct > 85
//                     → "Context window nearly full — consider /compact or splitting session"
// NO_CACHE_READ:      CacheReadTokens == 0 AND TotalTokens > 2M
//                     → "Zero cache reads on a large session — caching may be broken"
// CACHE_EXPLODE:      CacheCreateTokens > CacheReadTokens * 0.5 (after first 3 turns)
//                     → "Cache creation growing faster than reads — CLAUDE.md may be changing"
// MODEL_MISMATCH:     Model == "claude-opus-4-6" AND MessageCount < 10
//                     → "Opus used for a short session — consider Sonnet for quick tasks"
// EXPENSIVE_TOOLS:    ToolCalls["Bash"] > 50
//                     → "High bash tool usage — consider batching commands"
```

---

## 6. Pricing Engine

### 6.1 Embedded Pricing (assets/pricing.json)

```json
{
  "version": "2026-04-07",
  "models": {
    "claude-opus-4-6": {
      "input":        15.00,
      "output":       75.00,
      "cacheCreate":  18.75,
      "cacheRead":     1.50
    },
    "claude-sonnet-4-6": {
      "input":         3.00,
      "output":       15.00,
      "cacheCreate":   3.75,
      "cacheRead":     0.30
    },
    "claude-haiku-4-5": {
      "input":         0.80,
      "output":        4.00,
      "cacheCreate":   1.00,
      "cacheRead":     0.08
    }
  }
}
```

Units: USD per 1,000,000 tokens.

```go
//go:embed assets/pricing.json
var embeddedPricing []byte

func CalculateCost(usage Usage, model string) float64 {
    p := GetPricing(model) // fallback to sonnet pricing if unknown
    return float64(usage.InputTokens) * p.Input / 1e6 +
           float64(usage.OutputTokens) * p.Output / 1e6 +
           float64(usage.CacheCreationInputTokens) * p.CacheCreate / 1e6 +
           float64(usage.CacheReadInputTokens) * p.CacheRead / 1e6
}
```

### 6.2 Cost Source Priority

1. Use `costUSD` from JSONL if present and > 0 (Anthropic's own calculation)
2. Fall back to calculated cost from token counts × pricing table
3. Flag sessions where both are 0 as "no cost data"

---

## 7. CLI Commands & Flags

### 7.1 Global Flags

```
--claude-dir string   Path to Claude data dir (default: ~/.claude)
--since string        Start date filter: "7d", "30d", "2026-04-01", "today", "yesterday"
--until string        End date filter (same format as --since)
--project string      Filter to sessions in a specific project (partial path match)
--model string        Filter to specific model: "opus", "sonnet", "haiku"
--json                Output raw JSON (machine-readable, no UI)
--compact             Force narrow table layout (for terminals < 100 cols)
--no-color            Disable color output
--watch / -w          Live refresh mode (default interval: 5s)
--refresh int         Refresh interval in seconds for --watch (default: 5)
--debug               Print raw parsed data for debugging
```

### 7.2 Command Reference

```
agentop [today]          Default: today's sessions, duf-style panels
agentop daily            Sessions grouped by day, last 7 days
agentop monthly          Monthly summaries
agentop session [id]     Drill into one session (omit id → picker UI)
agentop blocks           5-hour billing windows
agentop doctor           Anomaly detection report
agentop config           Show effective config and pricing snapshot date
agentop version          Print version and build info
```

### 7.3 Filter Examples

```sh
agentop daily --since 30d                     # last 30 days
agentop today --project acme-app              # only acme-app sessions
agentop daily --model opus                    # only Opus sessions
agentop session abc123                        # specific session
agentop --watch --refresh 10                  # live mode, 10s refresh
agentop blocks --since today                  # today's billing blocks
agentop doctor --since 7d                     # anomalies in last week
agentop --json | jq '.sessions[] | .costUSD'  # pipe to jq
```

---

## 8. UI Specification

### 8.1 Terminal Layout

The UI renders as a sequence of bordered panels to stdout. In `--watch` mode, bubbletea takes over the terminal for live updates. In static mode (default), it prints and exits.

**Minimum terminal width:** 80 columns  
**Optimal terminal width:** 120+ columns  
**Auto-compact below:** 100 columns

### 8.2 Color Palette (styles.go)

```go
const (
    ColorGreen  = "#50c87a"  // cache efficiency ≥ 80%
    ColorAmber  = "#e5a040"  // cache efficiency 40-79%
    ColorRed    = "#e05050"  // cache efficiency < 40%
    ColorBlue   = "#4a90d9"  // input tokens bar segment
    ColorTeal   = "#7ecfb3"  // output tokens bar segment
    ColorPurple = "#b36de0"  // cache read tokens bar segment
    ColorDim    = "#666680"  // secondary text
    ColorBorder = "#3a3a4a"  // panel borders
    ColorHeader = "#9898b8"  // section header text
    ColorBg     = "#1a1a1e"  // background (dark terminal)
)

// Model tag colors
const (
    ColorTagOpus   = "#c090f0" // on bg #3a2050
    ColorTagSonnet = "#70b0e0" // on bg #1a3050
    ColorTagHaiku  = "#90d080" // on bg #1e3520
)
```

### 8.3 Token Bar Rendering (bars.go)

The bar is the visual core of the tool. Given a terminal width, compute the available bar width, then render 4 colored segments proportional to token counts.

```go
type TokenBar struct {
    Input       int64   // blue
    Output      int64   // teal
    CacheCreate int64   // amber
    CacheRead   int64   // purple
    Width       int     // terminal chars available
}

func (b TokenBar) Render() string {
    total := b.Input + b.Output + b.CacheCreate + b.CacheRead
    if total == 0 {
        return lipgloss.NewStyle().
            Background(lipgloss.Color(ColorBorder)).
            Width(b.Width).Render(" ")
    }

    segs := []struct{ count int64; color string }{
        {b.Input,       ColorBlue},
        {b.Output,      ColorTeal},
        {b.CacheCreate, ColorAmber},
        {b.CacheRead,   ColorPurple},
    }

    var parts []string
    remaining := b.Width
    for i, seg := range segs {
        w := int(float64(seg.count) / float64(total) * float64(b.Width))
        if i == len(segs)-1 {
            w = remaining // last segment takes all remaining
        }
        if w <= 0 {
            continue
        }
        parts = append(parts, lipgloss.NewStyle().
            Background(lipgloss.Color(seg.color)).
            Width(w).Render(""))
        remaining -= w
    }
    return strings.Join(parts, "")
}
```

**Below the bar:** A single subtitle line in dim color showing raw token counts:
```
0.3M in · 1.2M out · 0.5M cc · 4.9M cr
```

Where `cc` = cache create, `cr` = cache read.

### 8.4 Panel Structure (panels.go)

Each tool section (Claude Code, Codex CLI) is wrapped in a bordered panel:

```
┌─────────────────────────────────────────────────────────────────┐
│ claude code · 4 sessions · ~/.claude/projects/                  │
├────────────────────┬──────────────────────┬───┬────┬───┬────────┤
│ session            │ token mix             │ ☁ │msgs│ $ │ time  │
├────────────────────┼──────────────────────┼───┼────┼───┼────────┤
│ frontend-redesign  │ ████████░░░░░░░░░░░░ │91%│ 47 │$7 │ 2h18m │
│  [opus] ~/work/... │  0.3M·1.2M·0.5M·4.9M│   │    │   │       │
├────────────────────┼──────────────────────┼───┼────┼───┼────────┤
│ api-refactor  [ops]│ ████████████████████ │3% │  8 │$6 │  22m  │
└────────────────────┴──────────────────────┴───┴────┴───┴────────┘
```

Use lipgloss border styles with the exact color palette from the mockup.

### 8.5 Summary Strip (summary.go)

Five metric cards across the top:

```
╭─ today · apr 7 2026 · 6 sessions across 2 tools ─────────────╮
│  total cost    tokens used   cache efficiency  5h block  burn │
│   $18.42        42.3M            61%            23%     2.1k  │
│  ██████░░░░   ████████░░░   ███████░░░░░      ███░░░░  ████░ │
╰───────────────────────────────────────────────────────────────╯
```

Each card has: label (dim, 10px), value (bold, colored), mini bar (6px high).

### 8.6 Column Grid Layouts

**Full width (≥120 cols):**
```
[session name+model+path: 22%] [bar: 35%] [cache%: 7%] [msgs: 7%] [tools: 7%] [cost: 9%] [duration: 9%]
```

**Compact (80-119 cols):**
```
[session name: 22%] [bar: 40%] [cache%: 8%] [cost: 12%] [duration: 10%]
```

**Narrow (<80 cols):**
Fallback to single-column list with no bars, just numbers.

### 8.7 Watch Mode (--watch)

```go
// Bubbletea model with two message types:
type tickMsg time.Time       // fires every --refresh seconds
type fileChangeMsg string    // fired by fsnotify watcher

func (m Model) Init() tea.Cmd {
    return tea.Batch(
        tickEvery(m.refreshInterval),
        m.watcher.Watch(),
    )
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg.(type) {
    case tickMsg, fileChangeMsg:
        // Re-parse changed files, recompute aggregates, re-render
        return m.reload(), tickEvery(m.refreshInterval)
    case tea.KeyMsg:
        // q/Ctrl+C: quit; r: force refresh; arrows: navigate sessions
    }
}
```

### 8.8 Session Detail View (session.go)

Triggered by `agentop session <id>` or pressing Enter on a session row in watch mode:

```
╭─ session: frontend-redesign ─────────────────────────────────────╮
│ project: ~/work/acme-app · branch: feature/auth                  │
│ model: claude-opus-4-6 · started: 10:14am · duration: 2h 18m    │
├───────────────────────────────────────────────────────────────────┤
│ summary: Refactored authentication module to use JWT tokens...    │
├──────────── token breakdown ──────────────────────────────────────┤
│ input tokens:          300,000  ($0.45)    ████░░░░░░░░░░░░░░░░  │
│ output tokens:       1,200,000  ($9.00)    ████████████░░░░░░░░  │
│ cache create:          500,000  ($0.94)    █████░░░░░░░░░░░░░░░  │
│ cache read:          4,900,000  ($0.74)    █████████████████░░░  │
│ ─────────────────────────────────────────────────────────────    │
│ total:               6,900,000  ($7.20)    cache eff: 91% ✓      │
├──────────── tool usage ───────────────────────────────────────────┤
│ Read: 89  Edit: 34  Bash: 12  Grep: 8  Glob: 5  Write: 3        │
├──────────── turn-by-turn (last 5 turns) ──────────────────────────┤
│ #45 user: "now add tests for the validate function"               │
│     in: 78k  out: 3.4k  cr: 4.9M  cost: $0.14                   │
│ #46 asst: [Read src/auth.test.ts] [Edit src/auth.test.ts]        │
│ #47 user: "looks good, commit it"                                  │
│     in: 82k  out: 0.8k  cr: 4.9M  cost: $0.13                   │
╰───────────────────────────────────────────────────────────────────╯
```

### 8.9 Doctor View Output

```
╭─ anomalies & insights · last 7 days ─────────────────────────────╮
│                                                                   │
│  ⚠  WARN  api-refactor (acme-api)                               │
│           Cache efficiency 3% — cold-start session with large     │
│           context rebuild. 8 messages cost $5.81 ($0.73/msg).   │
│           Tip: resume previous session instead of starting fresh  │
│                                                                   │
│  ⚠  WARN  ci-yaml-fixes (service)                               │
│           Opus used for a 12-minute, 5-message session.          │
│           Sonnet would cost 5× less for routine tasks.           │
│                                                                   │
│  ✓  INFO  frontend-redesign — best cache day this week          │
│           91% efficiency over 2h 18m. Cost/message: $0.15.       │
│                                                                   │
│  ℹ  TIP   CLAUDE.md cache warm-up cost: ~$1.40 today            │
│           First turn of each session pays cache-creation premium. │
│           Run fewer, longer sessions to amortise this cost.       │
╰───────────────────────────────────────────────────────────────────╯
```

---

## 9. Configuration File

Location: `~/.config/agentop/config.toml` (XDG config dir)

```toml
[paths]
claude_dir = "~/.claude"      # Override Claude data directory

[display]
compact = false               # Always use compact mode
no_color = false              # Disable colors
timezone = "local"            # "local", "UTC", or "Asia/Kolkata"
date_format = "Jan 2 15:04"   # Go time format

[refresh]
interval = 5                  # --watch refresh interval in seconds

[pricing]
prefer_calculated = false     # Always calculate from tokens (vs using costUSD)
fetch_on_start = false        # Fetch latest pricing from LiteLLM on each run
pricing_url = "https://raw.githubusercontent.com/BerriAI/litellm/main/model_prices_and_context_window.json"

[thresholds]
cache_efficiency_good = 0.80  # >= this → green
cache_efficiency_warn = 0.40  # >= this → amber; below → red
expensive_session_usd = 3.0   # flag sessions above this cost
short_session_messages = 5    # flag high-cost short sessions

[cache]
enabled = true                # Use SQLite index for fast startup
path = "~/.cache/agentop/index.db"
```

---

## 10. Edge Cases & Robustness

| Scenario | Handling |
|---|---|
| Claude data dir doesn't exist | Print friendly setup message, exit 0 |
| No sessions today | Show empty panels with "No sessions found" |
| Session JSONL currently being written | Skip last line if JSON parse fails |
| costUSD = 0 for old sessions | Calculate from tokens with embedded pricing |
| Unknown model in session | Fall back to Sonnet pricing, flag in output |
| Terminal too narrow (< 60 cols) | Fallback to plain-text list |
| No TTY (piped output) | Auto-select --json mode |
| Permissions error on ~/.claude | Show specific error message |
| Corrupted JSONL (partial write) | Skip corrupt lines, warn if > 10% corrupt |
| Very large session (> 500 messages) | Stream parse, do not load all into memory |
| Sessions with no assistant messages | Show as "0 tokens" but still list the session |
| Multiple models in one session | Show all models, cost computed per-model |
| Time zone handling | All block calculations in UTC, display in local |
| Symlinked ~/.claude dir | Follow symlinks |

---

## 11. Testing Strategy

```
agentop/
└── testdata/
    ├── sessions/
    │   ├── good-cache.jsonl       # High cache efficiency session
    │   ├── cold-start.jsonl       # First session, 0% cache
    │   ├── large-session.jsonl    # 500+ messages
    │   ├── truncated.jsonl        # Last line cut off (in-progress)
    │   ├── mixed-models.jsonl     # Opus + Sonnet in one session
    │   └── no-cost.jsonl          # Old session without costUSD
    └── projects/
        └── example/
            └── sessions-index.json
```

Unit tests for:
- `ParseSession` with each testdata file
- `CalculateCost` for each model
- `BlockStart` / `BlockEnd` across timezone boundaries
- `CacheEfficiency` formula edge cases
- Column width calculation at 80, 100, 120, 160 cols
- Bar rendering at various widths (1, 5, 10, 20, 40 chars)
- Anomaly detection triggers
- Date filter parsing ("7d", "30d", "2026-04-01", "today")

Integration tests:
- Full pipeline: discover → parse → aggregate → render (snapshot testing)
- `--json` output validates as valid JSON
- `--watch` exits cleanly on Ctrl+C

---

## 12. Release & Distribution

```yaml
# .goreleaser.yml (key sections)
builds:
  - binary: agentop
    targets:
      - darwin_amd64
      - darwin_arm64
      - linux_amd64
      - linux_arm64
      - windows_amd64
    ldflags:
      - -s -w
      - -X main.version={{.Version}}
      - -X main.commit={{.Commit}}
      - -X main.date={{.Date}}

archives:
  - format: tar.gz
    format_overrides:
      - goos: windows
        format: zip

brews:
  - name: agentop
    tap:
      owner: yourname
      name: homebrew-tap

nfpms:
  - formats: [deb, rpm, apk]
    maintainer: Your Name <you@example.com>

checksum:
  name_template: checksums.txt

signs:
  - artifacts: checksum
```

**Version info in binary:**
```sh
$ agentop version
agentop v0.1.0 (darwin/arm64)
commit: a1b2c3d
built:  2026-04-07T10:00:00Z
```

---

## 13. Implementation Milestones

| Phase | Features | Estimated |
|---|---|---|
| v0.1 — MVP | Parse JSONL, today view, static output, basic bars | 1 week |
| v0.2 — Full reports | daily, monthly, blocks, session detail | 1 week |
| v0.3 — Live mode | --watch, fsnotify, bubbletea TUI | 3 days |
| v0.4 — Doctor | Anomaly detection, insights panel | 2 days |
| v0.5 — Polish | compact mode, --json, config file, edge cases | 3 days |
| v1.0 — Release | Tests, goreleaser, Homebrew tap, README | 1 week |

# Agentop — Multi-Agent Expansion Research & Implementation Plan

> Research conducted June 2026. Covers all major AI coding assistants, their session data formats, competitive landscape, and a phased implementation plan to make agentop the unified terminal dashboard for every AI coding tool.

---

## Table of Contents

1. [Project Value Assessment](#1-project-value-assessment)
2. [Competitive Landscape](#2-competitive-landscape)
3. [Data Format Deep-Dives](#3-data-format-deep-dives)
4. [Target Architecture](#4-target-architecture)
5. [Phased Implementation Plan](#5-phased-implementation-plan)
6. [Dos and Don'ts](#6-dos-and-donts)
7. [Pricing Architecture](#7-pricing-architecture)
8. [Agent Adapter Guidance](#8-agent-adapter-guidance)
9. [Testing Strategy](#9-testing-strategy)
10. [Risk Assessment](#10-risk-assessment)
11. [Release Roadmap](#11-release-roadmap)

---

## 1. Project Value Assessment

### The Market Need Is Real & Growing

AI coding assistants are burning through API tokens faster than ever. In June 2026, a typical developer might run 4–7 different AI coding tools: Claude Code (CLI), Cursor (IDE), Copilot CLI, Codex CLI, Kiro, Windsurf, Gemini CLI. Each has its own dashboard or no dashboard at all. **There is no unified, local, terminal-native cost dashboard that covers them all** — and that gap is what agentop can fill.

### agentop's Unique Position

| Dimension | agentop (current) | agentop (potential) |
|-----------|-------------------|---------------------|
| Language | Go | Go (still unique) |
| Binary size | ~10MB static binary | Same |
| Dependencies | Zero | Zero (still unique) |
| Agents | 1 (Claude) | 9+ |
| Anomaly detection | Deep | Deeper with multi-agent |
| Release infra | Complete | Same |
| Privacy | No network | Same |
| Setup time | <1 second | Same |

**Go is the moat.** Only agentop can offer a zero-dep, single-binary, cross-platform dashboard in an ecosystem dominated by Node.js and Python tools.

---

## 2. Competitive Landscape

| Tool | Claude | Codex | Cursor | Copilot | Kiro | Windsurf/Devin | Gemini | OpenCode | UI Type | Language |
|------|--------|-------|--------|---------|------|----------|--------|----------|---------|----------|
| **tokentop** | ✅ | ❌ | ✅ | ✅ | ❌ | ✅ | ✅ | ✅ | TUI | Node.js |
| **agenttop** | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | TUI+Web | Python |
| **crux** | ✅ | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | TUI | Rust |
| **ccusage** | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | CLI | Node.js |
| **agenthud** | ✅ | ✅ | ❌ | ❌ | ✅ | ❌ | ❌ | ✅ | TUI | Node.js |
| **codeburn** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ✅ | TUI | Node.js |
| **agent-lens** | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ✅ | ❌ | Web | Node.js |
| **CodeLedger** | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ | Web+Plugin | TypeScript |
| **Cohrint** | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ | CLI | Rust |
| **agentop (yours)** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | TUI | **Go** |

### Key Competitor Analysis

**tokentop** (`tokentop.app`): Most mature multi-agent tool. 7 agents, 11 providers, budget alerts, plugin system, real-time dashboard, 35+ plugins. Weakness: Node.js runtime required, complex npm setup.

**agenttop** (`vicarious11/agenttop`): 5 agents, TUI+Web, AI-powered recommendations via Map-Reduce-Generate pipeline. Activity classification (coding/debugging/testing/exploration). Per-tool breakdown and one-shot rate analysis. Weakness: Python runtime, complex install, still Claude-heavy.

**crux** (`amaljithkuttamath/crux`): 2 agents (Claude + Cursor), strongest per-agent depth. Session health grading (FRESH/OK/AGING/CRITICAL), context growth factor, duplication analysis, cache hit rate, efficiency grade (A-F). MCP server with 5 analysis tools. Weakness: Only 2 agents; Rust ecosystem.

**ccusage** (`ryoppippi/ccusage`): 2 first-class agents (Claude + Codex), ~15 others partial. 5-hour billing window tracking, MCP integration, --instances grouping, ultra-small bundle size. Weakness: No Cursor/Copilot.

**agenthud** (`IAMMARBIT/AgentHUD`): 4 agents. Focus on live session monitoring (active/thinking/waiting states), daily LLM-generated digests, game-style HUD overlay. Weakness: Live-monitor focus vs historical cost analysis.

**codeburn**: 25 agents (broadest coverage). Per-agent reader -> unified TUI via LiteLLM pricing. Weakness: Shallow per-agent detail; complex codebase.

**agent-lens** (`naimjeem/agent-lens`): 8 agents. Browser-based interactive dashboard with animated brand mark, live KPI pills, contribution heatmaps, cache performance dashboards. Weakness: Web UI (not terminal); heavier setup.

**CodeLedger** (`bhvbhushan/codeledger`): 4 agents. Per-skill token attribution, user vs overhead cost split, session category classification, conversational querying via MCP, budget alerts with anomaly detection, local web dashboard. Weakness: Runs as Claude Code plugin (tied to Claude ecosystem).

**Cohrint**: Smart model routing for cost optimization. Routes each request to cheapest model meeting quality bar. OpenTelemetry ingestion, real-time savings tracking. Weakness: Network-dependent (requires Cohrint cloud), not a dashboard tool per se.

### Unique Features to Incorporate from Competitors

| Feature | Source | agentop Equivalent |
|---------|--------|-------------------|
| Session health grading (FRESH/OK/AGING/CRITICAL) | crux | Extend doctor with health grades |
| Context growth factor & duplication analysis | crux | Add to aggregator stats |
| 5-hour billing window tracking | ccusage | Already have `blocks` command |
| Budget alerts with thresholds | tokentop, CodeLedger | Planned (Issue #28) |
| Per-skill/user vs overhead attribution | CodeLedger | Future consideration |
| AI-powered recommendations | agenttop | Future: local LLM analysis |
| Activity classification | agenttop | Future: categorize tool patterns |
| Client billing / invoicing integration | Vibes to Bucks | Future: per-client cost reports |
| Team/org dashboards | Coding IQ, Comet Opik | Support multiple --claude-dir scan |
| MCP server for agent queries | crux, ccusage | Planned (Issue #29) |
| Real-time live session monitoring | agenthud, agenttop | Not in scope (different niche) |
| Model routing optimization | Cohrint | Not in scope (different product)

---

## 3. Data Format Deep-Dives

### 3.1 Claude Code ✅ (Already Supported)

| Property | Value |
|----------|-------|
| **Location** | `~/.claude/projects/<hash>/*.jsonl` |
| **Format** | JSONL, one `RawEvent` per line |
| **Tokens** | Per-message in `message.usage` (input, output, cache_create, cache_read) |
| **Cost** | `costUSD` field in events |
| **Model** | Per-message `message.model` |
| **Project** | `cwd` field in first user event |
| **Dedup** | `message.id` + `stop_reason` — streaming produces multiple lines per message with zero tokens until final chunk |
| **Sub-agents** | Nested `subagents/*.jsonl` in session directories |
| **Edge cases** | Sidechain events (`isSidechain=true`) filtered out; tool-type users (`userType="tool"`) filtered out; empty stop_reason = intermediate streaming chunk |

### 3.2 Codex CLI (OpenAI)

| Property | Value |
|----------|-------|
| **Location** | `~/.codex/sessions/YYYY/MM/DD/rollout-<ISO8601>-<UUID>.jsonl` |
| **Format** | JSONL: `{"timestamp", "type", "payload"}` |
| **Event types** | `session_meta`, `turn_context`, `response_item`, `event_msg`, `input_item`, `config_snapshot` |
| **Tokens** | **Cumulative** via `event_msg` with `payload.type === "token_count"` — must diff consecutive events |
| **Cost** | Not in JSONL; must calculate from tokens × pricing |
| **Model** | Per-turn via `turn_context.model` |
| **Project** | `session_meta.cwd` |
| **Dedup** | No streaming dedup (each message appears once) |
| **Sub-agents** | Separate rollout files in date tree, linked via `session_meta.source.subagent` or `parent_thread_id` |
| **Tool linking** | Flat `call_id`-based (not nested); causality inferred from event ordering |
| **Edge cases** | No official schema documentation; archives excluded from resume; call_id ambiguity in parallel tool calls |

### 3.3 Kiro CLI

| Property | Value |
|----------|-------|
| **Location** | `~/.kiro/sessions/cli/{uuid}.json` + `{uuid}.jsonl` |
| **Format** | 4-file bundle: `.json` (metadata), `.jsonl` (conversation), `.history` (plain-text prompts), `.lock` (active session) |
| **Event kinds** | `Prompt` (user), `AssistantMessage` (model response + tool calls), `ToolResults` (tool outputs) |
| **Tokens** | **Not available** — must estimate via `len(text) / 4` |
| **Cost** | Not available — must calculate from estimated tokens |
| **Model** | From `.json` metadata file |
| **Project** | `cwd` in `.json` metadata |
| **Versions** | v1 (SQLite), v2 (SQLite), v3 (JSONL+JSON) — must support all three |
| **Edge cases** | Kiro IDE uses different VS Code-style `globalStorage` path; tool names need mapping |

### 3.4 GitHub Copilot CLI

| Property | Value |
|----------|-------|
| **Location** | `~/.copilot/session-state/<uuid>/events.jsonl` |
| **Format** | JSONL: `{"type", "timestamp", "data"}` |
| **Event types** | 20+ including `session.start`, `session.task_complete`, `session.shutdown`, `session.model_change`, `assistant.turn_start/end`, `assistant.message`, `user.message`, `tool.execution_start/complete`, `subagent.started/completed` |
| **Tokens** | Cumulative in `session.shutdown.data.modelMetrics.{model}.usage`; per-task in `session.task_complete` |
| **Cost** | Not in events — must calculate from tokens × pricing |
| **Model** | `session.start.data.context.model`; changes via `session.model_change` |
| **Project** | Repository from `session.start.data.context.repository` |
| **Secondary source** | `~/.copilot/session-store.db` (SQLite) with `sessions`, `turns`, `checkpoints`, `session_files`, `session_refs`, `search_index` tables (FTS5-indexed) |
| **Sync** | Data syncs to GitHub by default |
| **Edge cases** | Two formats: new `session-state/` + legacy `history-session-state/`; JetBrains users have separate `~/.copilot/jb/` path |

### 3.5 Cursor IDE

| Property | Value |
|----------|-------|
| **Location** | `~/Library/Application Support/Cursor/User/globalStorage/state.vscdb` (macOS) |
| **Format** | SQLite KV table `cursorDiskKV` with JSON blobs |
| **Data** | Composers (sessions) + bubbles (messages) with token counts, model info, timestamps |
| **Tokens** | Per-bubble `tokenCount.inputTokens` and `tokenCount.outputTokens` |
| **Cache** | **Not surfaced** in local data (always 0) |
| **Model** | `modelInfo.modelName`; "default"/empty falls back to "cursor-auto" |
| **Cost** | Not in local data; estimate or fetch from Cursor's CSV usage export |
| **Project** | Workspace path from bubble data or composer metadata |
| **Edge cases** | Zero-token bug in Cursor v3 (both input+output=0 → fall back to char estimation); `cursorDiskKV` uses `blob` type for some JSON rows; no cache breakdown locally |

### 3.6 Windsurf / Devin Desktop (Cognition)

> **Acquisition update (July 2025):** Cognition (makers of Devin) signed a definitive agreement to acquire Codeium (Windsurf). Windsurf continued operating independently through 2025–2026. In **June 2026**, Windsurf was relaunched as **Devin Desktop** — the IDE with Cascade was rebranded as "Devin Desktop" and Cascade itself was replaced by **"Devin Local"** (rewritten in Rust, ~30% more token efficient). The .pb encrypted protobuf format is still in use but now under the Devin brand.
>
> **Current (June 2026) Devin product line:**
> 1. **Devin Desktop** — formerly Windsurf IDE, the full-IDE experience with Cascade (now "Devin Local")
> 2. **Devin Cloud** — Cloud-hosted autonomous SWE agent ($500/mo+)
> 3. **Devin CLI** — Terminal-based Devin agent (new, post-acquisition)
> 4. **Devin Review** — PR review automation
>
> **Implications for agentop:** The Windsurf adapter should be future-proofed — the .pb format may remain for Devin Desktop local sessions, but Devin CLI sessions may use a different format (TBD, no data directories found yet).

| Property | Value |
|----------|-------|
| **Location** | `~/.codeium/windsurf/cascade/<uuid>.pb` (legacy Windsurf), `~/.devin/desktop/cascade/<uuid>.pb` (Devin Desktop) |
| **Format** | **AES-256-GCM encrypted protobuf** (`CortexTrajectory` message) |
| **Decryption key** | Hardcoded in `language_server` binary, shared across all users; may change with Devin rebranding |
| **Steps** | User messages, AI responses, tool calls, thinking content |
| **Tokens** | **Not available** in local .pb files — must estimate or skip |
| **Cost** | Not available |
| **Provider** | `provider` field per step (e.g., "anthropic") |
| **Auto-cleanup** | Deletes conversations after ~20 sessions; archived in `implicit/` directory |
| **Edge cases** | Key can change with Windsurf/Devin updates; decryption is fragile; AES key extraction from binary required; no local token counts; Devin CLI may use entirely different format |

### 3.7 Gemini CLI

| Property | Value |
|----------|-------|
| **Location** | `~/.gemini/tmp/<project_hash>/chats/session-*.json` or `session-*.jsonl` |
| **Format** | Single JSON (session object with `messages[]`) or JSONL (metadata + message lines) |
| **Tokens** | **Per-message** in `tokens{input, output, cached, thoughts}` — not cumulative |
| **Cost** | Not in data — must calculate from tokens × pricing |
| **Model** | Per-message `model` field |
| **Project** | Derived from directory name (`<project_hash>`) |
| **Dedup** | Message-level `id` — no streaming dedup needed |
| **Edge cases** | Cached tokens reported inside input total (must subtract cached from input before pricing); 30-day auto-cleanup default; thoughts (reasoning) tracked separately |

### 3.8 OpenCode

| Property | Value |
|----------|-------|
| **Location** | `~/.local/share/opencode/opencode.db` (v2.0+) or `~/.local/share/opencode/storage/message/` (legacy) |
| **Format** | SQLite with `sessions` and `messages` tables |
| **Tokens** | Per-message in `messages.data` JSON: `tokens{total, input, output, reasoning, cache{read, write}}` |
| **Cost** | Per-message `cost` field in `messages.data` (most accurate of any agent) |
| **Model** | `messages.data.modelID` + `messages.data.providerID` |
| **Project** | `sessions.directory` field |
| **Dedup** | Message-level `id`; no streaming dedup needed |
| **Edge cases** | Two storage formats (SQLite v2.0+ vs legacy JSON); uses OpenRouter (many providers); cost field is the single source of truth |

### 3.9 JetBrains Copilot

| Property | Value |
|----------|-------|
| **Location** | `~/.copilot/jb/<uuid>/partition-*.jsonl` |
| **Format** | JSONL with JetBrains-specific events |
| **Event types** | `user.message_rendered` (user), `assistant.message` (AI response), `tool.execution_start/complete` (tools) |
| **Tokens** | Limited/absent in local files |
| **Cost** | Not available — must estimate |
| **Model** | Inferred from tool call ID prefixes (`toolu_vrtx_` = Claude, `call_` = GPT) or `data.model` field |
| **Project** | Workspace path from metadata |
| **IDE detection** | Scan `~/Library/Application Support/JetBrains/` for 13+ IDE variants (IntelliJ, PyCharm, WebStorm, GoLand, CLion, etc.) |

### 3.10 Continue CLI (cn)

| Property | Value |
|----------|-------|
| **Location** | `~/.continue/sessions/*.json` |
| **Format** | JSON with `sessionId`, `title`, `workspaceDirectory`, `messages[]`, `usage{}` |
| **Tokens** | Per-session accumulated `usage{promptTokens, completionTokens, cachedTokens, cacheWriteTokens}` |
| **Cost** | Per-session `usage.totalCost` |
| **Model** | Per-message in messages array |
| **Project** | `workspaceDirectory` field |
| **Edge cases** | Simplest format of all agents; supports many providers via proxy architecture |

### 3.11 Devin CLI (Cognition)

> **Note:** Devin CLI is a new product launched in June 2026 after the Windsurf acquisition and rebranding. Its local session data format is still being investigated. This section will be updated when format details are confirmed. Currently, Devin is only supported via the Windsurf/Devin Desktop adapter (.pb files). Devin CLI may use a different format (e.g., JSONL or SQLite).

| Property | Value |
|----------|-------|
| **Location** | TBD — not yet confirmed |
| **Format** | Unknown — likely JSONL or SQLite |
| **Tokens** | Unknown |
| **Cost** | Unknown |
| **Model** | Unknown |
| **Project** | Unknown |
| **Status** | 🔴 Not yet implemented — needs investigation once Devin CLI releases stable data directories |

### 3.12 Summary Table: Agent Data Sources

| Agent | Location | Format | Tokens | Cost | Model | Project |
|-------|----------|--------|--------|------|-------|---------|
| Claude Code | `~/.claude/projects/*/*.jsonl` | JSONL | ✅ per-message | ✅ costUSD field | ✅ per-message | ✅ cwd |
| Codex CLI | `~/.codex/sessions/**/rollout-*.jsonl` | JSONL | ✅ cumulative (needs diff) | ❌ calc from tokens | ✅ per-turn | ✅ cwd |
| Kiro CLI | `~/.kiro/sessions/cli/{uuid}.json/.jsonl` | JSON+JSONL | ❌ must estimate | ❌ calc from tokens | ✅ metadata | ✅ cwd |
| Copilot CLI | `~/.copilot/session-state/*/events.jsonl` | JSONL | ✅ cumulative in shutdown | ❌ calc from tokens | ✅ session.start | ✅ repo |
| Cursor | `state.vscdb` (SQLite) | SQLite+JSON | ✅ per-bubble | ❌ CSV/calc | ✅ per-bubble | ✅ workspace |
| Windsurf/Devin Desktop | `~/.codeium/.../*.pb` / `~/.devin/.../*.pb` | **Encrypted Proto** | ❌ must estimate | ❌ calc | ✅ gRPC | ✅ workspace |
| Gemini CLI | `~/.gemini/tmp/*/chats/session-*.json` | JSON | ✅ per-message | ❌ calc from tokens | ✅ per-message | ✅ project_hash |
| OpenCode | `~/.local/share/opencode/opencode.db` | SQLite | ✅ per-message | ✅ per-message cost | ✅ + provider | ✅ directory |
| JetBrains | `~/.copilot/jb/*/partition-*.jsonl` | JSONL | ❌ limited | ❌ calc | ✅ heuristic | ✅ workspace |
| Continue | `~/.continue/sessions/*.json` | JSON | ✅ per-session | ✅ totalCost | ✅ per-message | ✅ workspace |
| Devin CLI | TBD | Unknown | ❓ | ❓ | ❓ | ❓ |

---

## 4. Target Architecture

### Current Architecture

```
main.go -> cmd/ (cobra commands)
            |
            v
           claude.Discover() -> reads ~/.claude/projects/
           claude.ParseSession() -> JSONL parsing (Claude-specific)
            |
            v
           aggregator.AggregateSession() -> dedup + stats (Claude-specific)
            |
            v
           pricing.DefaultPricer.Calculate() -> Claude-only pricing.json
            |
            v
           ui.RenderToday() / RenderSessionsTable() -> BubbleTea TUI
```

### Target Architecture

```
main.go -> cmd/ (cobra commands)
            |
            v
           registry.Discover() -> scans ALL known agent directories
           registry.DetectAgents() -> auto-detects which agents are installed
            |
            v
           [Agent Adapter Layer - each agent has its own parser]
           +-- claude.Adapter    -> ~/.claude/projects/*/*.jsonl
           +-- codex.Adapter     -> ~/.codex/sessions/**/rollout-*.jsonl
           +-- kiro.Adapter      -> ~/.kiro/sessions/cli/*.{json,jsonl}
           +-- copilot.Adapter   -> ~/.copilot/session-state/*/events.jsonl
           +-- cursor.Adapter    -> ~/Library/.../state.vscdb (SQLite)
           +-- gemini.Adapter    -> ~/.gemini/tmp/*/chats/session-*.json
           +-- opencode.Adapter  -> ~/.local/share/opencode/opencode.db
           +-- windsurf.Adapter  -> ~/.codeium/windsurf/cascade/*.pb (decrypted)
           +-- jetbrains.Adapter -> ~/.copilot/jb/*/partition-*.jsonl
           +-- continue.Adapter  -> ~/.continue/sessions/*.json
           +-- [community adapter interface - anyone can add one]
            |
            v
           Normalized SessionStats (unified data model)
            |
            v
           pricing.Pricer (extended) -> multi-provider pricing
            |
            v
           aggregator (extended -> multi-source merge, source attribution)
            |
            v
           ui (extended) -> source badges, per-agent breakdown, combined views
           doctor (extended) -> multi-agent anomaly detection
```

### New Package Structure

```
internal/
+-- adapter/              NEW: agent adapters (one sub-package per agent)
|   +-- registry.go       NEW: agent discovery + detection
|   +-- types.go          NEW: Adapter interface, ParsedEvent, SessionFile
|   +-- claude/           MOVE: from internal/claude/
|   +-- codex/            NEW
|   +-- kiro/             NEW
|   +-- copilot/          NEW
|   +-- cursor/           NEW
|   +-- gemini/           NEW
|   +-- opencode/         NEW
|   +-- windsurf/         NEW
|   +-- jetbrains/        NEW
|   +-- continue/         NEW
+-- aggregator/           KEEP + extend
|   +-- session.go        EXTEND: handle multi-source, normalize all formats
+-- pricing/              EXTEND: multi-provider pricing
|   +-- models.go
|   +-- pricing.json      EXTEND: add OpenAI, Gemini, Cursor, etc.
|   +-- providers/        NEW: per-provider pricing files
+-- claude/               OBSOLETE: move to adapter/claude/
+-- ui/                   EXTEND: source badges, agent filtering, combined views
```

---

## 5. Phased Implementation Plan

### Phase 0: Foundation (1-2 weeks)

**Goal:** Extract adapter pattern without breaking existing functionality

1. **Create `internal/adapter/types.go`** — Define the Adapter interface:
   - `Adapter` interface with `ID()`, `Name()`, `AgentType()`, `Discover()`, `ParseSession()` methods
   - `SessionFile` struct (Path, AgentID, AgentType, Project, SessionID, ModTime)
   - `ParsedEvent` struct (normalized event for all agents)
   - `AgentType` enum (CLI, IDE)
   - `EventType` enum (user, assistant, tool, summary)

2. **Create `internal/adapter/registry.go`** — Auto-detect installed agents:
   - `DetectInstalledAgents() []Adapter` — checks each known data directory
   - `DiscoverAllSessions(adapters) []SessionFile` — aggregates all discovered files
   - `ParseSessionFromAdapter(adapter, path) []*ParsedEvent` — delegates to correct adapter

3. **Refactor: Move `internal/claude/` -> `internal/adapter/claude/`**
   - Move `discover.go`, `jsonl.go`, `schema.go`, `index.go` to `internal/adapter/claude/`
   - Implement the `Adapter` interface for Claude Code
   - Keep backward compat with aliases or thin wrappers
   - Update all imports in `cmd/`, `internal/aggregator/`, `internal/ui/`

4. **Update CLI:**
   - Add `--agent` flag to root command (`claude`, `codex`, etc. or `all` by default)
   - Add `--list-agents` flag to show detected agents and exit
   - Wire registry into command flow

### Phase 1: CLI Agents (2-3 weeks)

**Goal:** Support all CLI-based agents with normal JSON/JSONL formats

1. **Codex CLI Adapter** (`internal/adapter/codex/`)
   - Parse `rollout-*.jsonl` files from `~/.codex/sessions/YYYY/MM/DD/`
   - Implement cumulative token diff tracking
   - Handle `session_meta` -> project/cwd
   - Handle `turn_context` -> per-turn model
   - Handle `response_item` -> tool tracking via `call_id`
   - Handle sub-agents (separate rollout files with `parent_thread_id`)
   - Add OpenAI model pricing to `pricing.json`

2. **Gemini CLI Adapter** (`internal/adapter/gemini/`)
   - Parse `session-*.json` files from `~/.gemini/tmp/*/chats/`
   - Handle per-message token extraction (not cumulative)
   - Handle cached-in-input subtraction (cached tokens reported inside input total)
   - Handle thoughts/reasoning tokens
   - Add Gemini model pricing

3. **OpenCode Adapter** (`internal/adapter/opencode/`)
   - Add `modernc.org/sqlite` dependency (pure Go SQLite driver)
   - Query `sessions` and `messages` tables
   - Extract per-message tokens, cost, model, provider
   - Handle legacy JSON format as fallback
   - Handle OpenRouter provider mapping

4. **Kiro CLI Adapter** (`internal/adapter/kiro/`)
   - Parse 4-file bundle (`{uuid}.json`, `{uuid}.jsonl`, `{uuid}.history`)
   - Token estimation via `len(text) / 4`
   - Handle all three storage versions (v1 SQLite, v2 SQLite, v3 JSONL)
   - Tool name normalization

5. **Copilot CLI Adapter** (`internal/adapter/copilot/`)
   - Parse `events.jsonl` from `~/.copilot/session-state/<uuid>/`
   - Extract model from `session.start`
   - Extract tokens from `session.shutdown` (cumulative model metrics)
   - Handle `tool.execution_start/complete` for tool tracking
   - Handle `session.model_change` for model switches
   - Optional: parse `session-store.db` SQLite for richer data

### Phase 2: IDE Agents (3-4 weeks)

**Goal:** Support IDE-based agents with more complex data sources

1. **Cursor IDE Adapter** (`internal/adapter/cursor/`)
   - Implement cross-platform path detection
   - Open SQLite, read `cursorDiskKV` table
   - Parse JSON blobs for composers/bubbles
   - Extract `tokenCount`, `modelInfo`, timestamps, workspace paths
   - Handle zero-token bug (fallback to char estimation)
   - Cursor cache tokens are 0 -- mark as estimated

2. **Windsurf Adapter** (`internal/adapter/windsurf/`)
   - Implement AES-256-GCM decryption of `.pb` files
   - Add `google.golang.org/protobuf` dependency
   - Parse decrypted `CortexTrajectory` protobuf
   - Extract steps: user messages, AI responses, tool calls, thinking
   - Token estimation (no token counts in local data)
   - Graceful degradation if key changes

3. **JetBrains Copilot Adapter** (`internal/adapter/jetbrains/`)
   - Parse `partition-*.jsonl` from `~/.copilot/jb/<uuid>/`
   - Handle JetBrains-specific event types
   - Model inference from tool call ID patterns
   - IDE auto-detection (scan for active JetBrains IDEs)
   - Tool name mapping

4. **Continue CLI Adapter** (`internal/adapter/continue/`)
   - Parse JSON files from `~/.continue/sessions/`
   - Simplest format -- extract usage, messages, model directly
   - Handle proxy-based provider mapping

### Phase 3: Unified Aggregation & UI (2-3 weeks)

**Goal:** Merge data from all sources into unified views

1. **Extend SessionStats struct:**
   - Add `AgentID string` field (identifies source agent)
   - Add `AgentName string` field (human-readable)
   - Add `AgentType AgentType` field (CLI vs IDE)
   - Add `TokenSource TokenSource` field (exact vs estimated)

2. **Extend Aggregator:**
   - Accept `[]*ParsedEvent` from any adapter
   - Handle different token availability per agent
   - Track source attribution throughout pipeline
   - Add `Source` field to all intermediate types

3. **Extend Pricing Engine:**
   - Move from single `pricing.json` to multi-provider system
   - Add provider-specific models (OpenAI, Google, Cursor, GitHub)
   - Add model aliasing (e.g., "cursor-auto" -> Cursor Auto pricing)
   - Add estimation fallback when pricing is unknown
   - Support reasoning token pricing (o3, o4-mini)

4. **Extend UI:**
   - Add **source badge** to each session row (colored label like `[claude]` `[codex]`)
   - Add **per-agent summary strip** (total cost/tokens per agent type)
   - Add **agent breakdown panel** in the summary strip
   - Add `d` key in watch mode to cycle through agent filters
   - Color-code agent badges uniquely

### Phase 4: Advanced Features (ongoing)

**Goal:** Differentiate from competitors with unique cross-agent intelligence

1. **Cross-Agent Intelligence:**
   - Combined project view across agents: "You used Claude Code + Cursor on Project X -- combined cost: $Y"
   - Agent migration analysis: "You switched from Codex to Claude mid-project -- cost comparison"
   - Most-expensive-agent ranking

2. **Budget Alerts:**
   - Per-agent and total daily/weekly/monthly budgets
   - Color-coded warning indicators in TUI
   - Optional desktop notifications

3. **MCP Server:**
   - Expose session stats as MCP tools
   - Allow agents to query their own cost mid-session
   - Cross-agent comparison tools

4. **Extended Doctor for Multi-Agent:**
   - "Your Opus usage in Cursor shows same overkill pattern as in Claude Code"
   - "Kiro sessions have zero cache efficiency (no cache tracking)"
   - Cross-agent cache health comparison
   - Agent-specific optimization recommendations

5. **Community Adapter API:**
   - Documented interface for third-party adapters
   - Plugin discovery mechanism
   - Adapter template/example

---

## 6. Dos and Don'ts

### DO

| # | Rule | Why |
|---|------|------|
| 1 | Use the Adapter pattern | Clean separation; anyone can add a new agent. Single most important architectural decision. |
| 2 | Keep Go as the only language | Zero-dependency binary is your biggest moat. |
| 3 | Add `modernc.org/sqlite` | Pure Go SQLite driver (no CGO). Required for Cursor/OpenCode. |
| 4 | Support agent auto-detection | `agentop` should Just Work -- no config needed. |
| 5 | Add `google.golang.org/protobuf` | Pure Go protobuf for Windsurf decryption. |
| 6 | Add agent badges to the UI | Users need to see at a glance which sessions are from which tool. |
| 7 | Mark estimated tokens explicitly | When tokens are estimated (Kiro, Windsurf), show `~` or asterisk in UI. |
| 8 | Release Phase 0 as non-breaking refactor | Test each adapter refactor against existing tests. |
| 9 | Use error-tolerant parsing | Each agent's format changes over time; log warnings, don't crash. |
| 10 | Add `--list-agents` flag first | Users will verify agents are detected before diving in. |
| 11 | Handle multiple agents per project | Show combined cost per project with agent breakdown. |
| 12 | Write separate tests per adapter | Each adapter needs its own test suite with sample data. |

### DON'T

| # | Rule | Why |
|---|------|------|
| 1 | Make network calls | Would lose the privacy advantage. All data must be local. |
| 2 | Add CGO dependencies | Would kill cross-compilation and simple install. |
| 3 | Try to support all 25 agents day one | Start with 6-7 most popular (Claude, Codex, Cursor, Copilot, Kiro, Gemini, OpenCode). |
| 4 | Flatten all agents into one data model too early | Let each adapter return native types; normalize in aggregator. |
| 5 | Prioritize Windsurf decryption | Hardest and most fragile. Save for Phase 2. |
| 6 | Remove old `internal/claude/` package immediately | Keep as compatibility shim until Phase 3. |
| 7 | Estimate tokens without clear indicator | Users need to know which numbers are exact vs estimated. |
| 8 | Make UI more complex than necessary | Add a source column; don't redesign the entire layout. |
| 9 | Add real-time streaming monitoring (top for agents) | That's agenttop/agenthud's niche. agentop = historical cost analysis. |
| 10 | Depend on external pricing API | Keep pricing embedded. Opt-in refresh is acceptable. |
| 11 | Fail hard on Windsurf format changes | Make it best-effort: skip with warning if decryption fails. |
| 12 | Skip error handling per agent format | Each agent handles streaming differently; has different edge cases. |

---

## 7. Pricing Architecture

### Current: Single pricing.json (Claude-only)

```json
{
  "version": "2026-04-07",
  "models": {
    "claude-opus-4-6": {"input": 15.00, "output": 75.00, "cacheCreate": 18.75, "cacheRead": 1.50},
    "claude-sonnet-4-6": {"input": 3.00, "output": 15.00, "cacheCreate": 3.75, "cacheRead": 0.30},
    "claude-haiku-4-5": {"input": 0.80, "output": 4.00, "cacheCreate": 1.00, "cacheRead": 0.08}
  }
}
```

### Target: Multi-Provider Pricing

```json
{
  "version": "2026-07-01",
  "providers": {
    "anthropic": {
      "display": "Anthropic",
      "models": {
        "claude-opus-4-6": {"input": 15.00, "output": 75.00, "cacheCreate": 18.75, "cacheRead": 1.50},
        "claude-sonnet-4-6": {"input": 3.00, "output": 15.00, "cacheCreate": 3.75, "cacheRead": 0.30},
        "claude-haiku-4-5": {"input": 0.80, "output": 4.00, "cacheCreate": 1.00, "cacheRead": 0.08}
      }
    },
    "openai": {
      "display": "OpenAI",
      "models": {
        "gpt-4.1": {"input": 2.00, "output": 8.00},
        "gpt-4.1-mini": {"input": 0.40, "output": 1.60},
        "gpt-4.1-nano": {"input": 0.10, "output": 0.40},
        "gpt-5": {"input": 10.00, "output": 40.00},
        "o3": {"input": 10.00, "output": 40.00, "reasoning": 40.00},
        "o4-mini": {"input": 1.10, "output": 4.40, "reasoning": 4.40}
      }
    },
    "google": {
      "display": "Google",
      "models": {
        "gemini-2.5-pro": {"input": 1.25, "output": 5.00, "cacheRead": 0.125},
        "gemini-2.5-flash": {"input": 0.15, "output": 0.60, "cacheRead": 0.015}
      }
    },
    "cursor": {
      "display": "Cursor",
      "models": {
        "cursor-auto": {"input": 1.25, "output": 6.00, "cacheRead": 0.25}
      }
    }
  },
  "fallback": {
    "provider": "openai",
    "model": "gpt-4.1"
  }
}
```

### Pricing Fallback Chain

1. Exact model match -> use that pricing
2. Prefix match (e.g., `gpt-5-preview` -> `gpt-5`)
3. Unknown model -> provider fallback (e.g., unknown OpenAI model -> `gpt-4.1` rates)
4. Unknown provider -> global fallback (`claude-sonnet-4-6` rates, current behavior)

---

## 8. Agent Adapter Guidance

### Adapter Interface

```go
// internal/adapter/types.go

type AgentType int
const (
    AgentCLI AgentType = iota
    AgentIDE
)

type EventType int
const (
    EventUser EventType = iota
    EventAssistant
    EventTool
    EventSummary
)

type TokenSource int
const (
    TokenExact TokenSource = iota
    TokenEstimated
)

type Usage struct {
    InputTokens   int64
    OutputTokens  int64
    CacheCreate   int64
    CacheRead     int64
    Reasoning     int64
    TokenSource   TokenSource
}

type ParsedEvent struct {
    SessionID    string
    AgentID      string
    Timestamp    time.Time
    Type         EventType
    Model        string
    Usage        *Usage
    CostUSD      float64
    CostSource   string // "exact" or "calculated"
    ToolName     string
    ToolInput    json.RawMessage
    ToolResult   json.RawMessage
    MessageID    string
    StopReason   string
    ProjectPath  string
    GitBranch    string
    IsSidechain  bool
    Raw          json.RawMessage // original line for debugging
}

type SessionFile struct {
    Path      string
    AgentID   string
    AgentType AgentType
    Project   string
    SessionID string
    ModTime   time.Time
    Size      int64
}

type Adapter interface {
    ID() string
    Name() string
    AgentType() AgentType
    Discover(baseDir string) ([]SessionFile, error)
    ParseSession(path string) ([]*ParsedEvent, error)
}
```

### Codex CLI Adapter -- Key Implementation Details

- **Discovery:** Walk `~/.codex/sessions/YYYY/MM/DD/*.jsonl` (or `CODEX_SESSIONS_DIR`)
- **Parsing:** Track cumulative token counts per message; subtract previous from current
- **Token events:** `type=="event_msg" && payload.type=="token_count"` with `payload.info.last_token_usage`
- **Model:** From `turn_context` event (per-turn, not per-message)
- **Sub-agent detection:** `session_meta.source` is object-with-`subagent` -> child session
- **Tool linking:** `call_id` on `response_item.type=function_call` and `response_item.type=function_call_output`

### Kiro CLI Adapter -- Key Implementation Details

- **Discovery:** Walk `~/.kiro/sessions/cli/*.json` (metadata) + corresponding `*.jsonl`
- **Metadata parsing:** Read `.json` for `cwd`, `title`, `model`, `parent_session_id`
- **Conversation parsing:** Read `.jsonl` -- kind `Prompt` (user), `AssistantMessage` (response), `ToolResults` (tool outputs)
- **Token estimation:** `len(text) / 4` for all text content
- **Version detection:** Check `~/.kiro/sessions/cli/` (v3 JSONL) first, fall back to SQLite
- **Tool name mapping:** Kiro uses different names than Claude -- needs normalization

### Copilot CLI Adapter -- Key Implementation Details

- **Discovery:** Walk `~/.copilot/session-state/<uuid>/events.jsonl`
- **Session start:** `session.start` -> model, repository, cwd
- **Token extraction:** `session.shutdown.data.modelMetrics.{model}.usage.inputTokens/outputTokens`
- **Tool tracking:** `tool.execution_start/complete` -- name, arguments, success/failure
- **Model changes:** `session.model_change` -- track model mid-session
- **Secondary source:** `~/.copilot/session-store.db` has structured data (sessions, turns, files tables)
- **Sync awareness:** Note in UI if data may be partial due to cloud sync

### Cursor Adapter -- Key Implementation Details

- **Discovery:** Platform-specific path detection
  - macOS: `~/Library/Application Support/Cursor/User/globalStorage/state.vscdb`
  - Linux: `~/.config/Cursor/User/globalStorage/state.vscdb`
  - Windows: `%APPDATA%/Cursor/User/globalStorage/state.vscdb`
- **SQLite:** Use `modernc.org/sqlite` (pure Go, no CGO)
- **KV parsing:** Read `cursorDiskKV` table; decode blob/JSON values
- **Composer data:** Sessions as composers with per-bubble messages
- **Token extraction:** `tokenCount.inputTokens`, `tokenCount.outputTokens` per bubble
- **Zero-token fallback:** If both are 0, estimate via `len(text) / 4`
- **Model:** `modelInfo.modelName`; "default" -> "cursor-auto"
- **No cache data:** Cache tokens are always 0 locally

### Windsurf Adapter -- Key Implementation Details

- **Discovery:** Walk `~/.codeium/windsurf/cascade/*.pb` and `implicit/*.pb`
- **Decryption:** AES-256-GCM with hardcoded key from `language_server` binary
- **Key extraction:** Parse Go binary for the hardcoded key (documented technique)
- **Protobuf schema:** `CortexTrajectory` message with repeated steps
- **Step types:** 14=user message, 15=AI response, 21=tool execution
- **No token data:** All token counts must be estimated or marked as unavailable
- **Error handling:** If decryption fails, log warning and skip Windsurf sessions
- **Key rotation monitoring:** Track Windsurf version changes that may update the key

### Gemini CLI Adapter -- Key Implementation Details

- **Discovery:** Walk `~/.gemini/tmp/*/chats/session-*.json` (JSON) and `session-*.jsonl`
- **Parsing:** Single JSON with `sessionId`, `startTime`, `messages[]` array
- **Token extraction:** Per-message `tokens{input, output, cached, thoughts}`
- **Cached-in-input:** `cached` is reported inside `input` -- subtract cached from input for pricing
- **Reasoning:** `thoughts` field -> reasoning tokens for Gemini models
- **Model:** Per-message `model` field
- **No cost:** Always calculate from tokens x pricing

### OpenCode Adapter -- Key Implementation Details

- **Discovery:** Check `~/.local/share/opencode/opencode.db` (SQLite v2.0+)
- **SQLite tables:**
  - `sessions`: `id`, `directory`, `title`, `time_created`, `time_updated`
  - `messages`: `id`, `session_id`, `role`, `data` (JSON), `time_created`
- **Data extraction:** Parse `messages.data` JSON for `modelID`, `providerID`, `cost`, `tokens{input, output, reasoning, cache{read, write}}`
- **Direct cost:** `messages.data.cost` is the most accurate cost of any agent -- use directly
- **Provider awareness:** OpenCode uses OpenRouter -- `providerID` identifies the upstream provider
- **Legacy fallback:** If SQLite not found, check `~/.local/share/opencode/storage/message/` for JSON files

### JetBrains Adapter -- Key Implementation Details

- **Discovery:** Walk `~/.copilot/jb/<uuid>/partition-*.jsonl`; detect active IDEs via `~/Library/Application Support/JetBrains/`
- **Event types:** `user.message_rendered` (user input), `assistant.message` (AI response), `tool.execution_start/complete`
- **Model inference:**
  - `toolu_vrtx_` prefix in toolCallId -> Claude
  - `call_` prefix -> GPT
  - Check `data.model` field if available (100x weight)
- **IDE detection:**
  - macOS: `~/Library/Application Support/JetBrains/` -- list subdirectories (IntelliJIdea*, PyCharm*, WebStorm*, etc.)
  - Linux: `~/.config/JetBrains/`
  - Windows: `%APPDATA%/JetBrains/`

### Continue Adapter -- Key Implementation Details

- **Discovery:** Walk `~/.continue/sessions/*.json`
- **Parsing:** Direct JSON unmarshal into `{sessionId, title, workspaceDirectory, messages[], usage{}}`
- **Token extraction:** `usage.promptTokens`, `usage.completionTokens`, `usage.cachedTokens`, `usage.cacheWriteTokens`
- **Cost:** `usage.totalCost` -- direct from session data
- **Model:** Per-message in messages array

---

## 9. Testing Strategy

| Agent | Test Data Needed | Edge Cases |
|-------|-----------------|------------|
| Codex | Sample rollout JSONL file | Cumulative token diff, call_id linking, sub-agent rollouts |
| Kiro | Sample {uuid}.json + .jsonl pair | Three storage versions, missing .json file, empty session |
| Copilot | Sample events.jsonl | Legacy format migration, sync status detection, model change events |
| Cursor | Sample state.vscdb (anonymized) | Zero-token bug, cursorDiskKV blob vs text, workspace path extraction |
| Gemini | Sample session-*.json | Per-message tokens, thoughts field, cached-in-input subtraction |
| OpenCode | Sample opencode.db | SQLite vs legacy JSON, per-message cost extraction, OpenRouter model format |
| Windsurf | Sample .pb files | Key rotation, decryption failure, missing token fields |

### Testdata Directory Structure

```
testdata/
+-- sessions/
|   +-- good-cache.jsonl          (existing, Claude)
|   +-- cold-start.jsonl          (existing, Claude)
|   +-- in-progress.jsonl         (existing, Claude)
|   +-- codex-rollout.jsonl       (NEW)
|   +-- kiro-session/             (NEW - directory with 4 files)
|   |   +-- uuid.json
|   |   +-- uuid.jsonl
|   |   +-- uuid.history
|   |   +-- uuid.lock
|   +-- copilot-events.jsonl      (NEW)
|   +-- gemini-session.json       (NEW)
|   +-- opencode.db               (NEW)
|   +-- windsurf-trajectory.pb    (NEW)
+-- adapters/
    +-- codex_test.go
    +-- kiro_test.go
    +-- copilot_test.go
    +-- cursor_test.go
    +-- gemini_test.go
    +-- opencode_test.go
    +-- windsurf_test.go
```

---

## 10. Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| Windsurf encryption key changes | High -- adapter breaks | Make Windsurf best-effort; detect key failure gracefully; document extraction method |
| Cursor SQLite format changes | Medium -- parser breaks | Error-tolerant parsing; schema version detection; fallback to char estimation |
| Codex CLI cumulative token diff errors | Medium -- token counts wrong | Validate: per-event tokens should be >= 0; if negative, log warning and skip |
| Copilot stops writing local files | Low -- data only in cloud | Monitor GitHub changes; add cloud sync integration as opt-in |
| Kiro adds v4 format | Medium -- missing sessions | Version detection; support rolling upgrades |
| Competitor adds all agents in Go | Low -- unlikely (ecosystem is JS/Python/Rust-heavy) | First-mover in Go space; focus on Go-specific advantages |
| Community expects all agents day one | Medium -- reputation risk | Clear agent support matrix in README; transparent roadmap |

---

## 11. Release Roadmap

| Release | Timeline | Features |
|---------|----------|----------|
| **v0.2.0** | Month 1 | Phase 0: Adapter pattern, Codex CLI support, Codex pricing |
| **v0.3.0** | Month 2 | Gemini CLI, OpenCode adapters; per-agent filtering; --agent flag |
| **v0.4.0** | Month 3 | Kiro, Copilot CLI adapters; estimation indicators; source badges in UI |
| **v0.5.0** | Month 4-5 | Cursor adapter (SQLite); per-agent summary strip; combined view |
| **v0.6.0** | Month 6 | Windsurf (decrypt, no tokens), JetBrains, Continue adapters |
| **v0.7.0** | Month 7 | Budget alerts, MCP server, extended doctor for multi-agent |
| **v1.0.0** | Month 8 | All features stable, documented adapter API for community plugins |

---

## 12. Final Coverage Matrix (as of implementation)

### Agent Support

| Agent | ID | Status | Format | Token Source | Cache Data | Cost Data |
|-------|----|--------|--------|-------------|------------|-----------|
| Claude Code | `claude` | ✅ Done | JSONL (files) | Exact | ✅ Yes | ✅ Yes |
| Codex CLI | `codex` | ✅ Done | JSONL (events) | Exact (cumulative) | ❌ No | ❌ Calc'd |
| Gemini CLI | `gemini` | ✅ Done | JSON/JSONL | Exact (cached in input) | Partial | ❌ Calc'd |
| OpenCode | `opencode` | ✅ Done | SQLite | Exact | ❌ No | ✅ Yes |
| Kiro CLI | `kiro` | ✅ Done | JSONL | Estimated | ❌ No | ❌ Calc'd |
| Copilot CLI | `copilot` | ✅ Done | JSONL | Exact (shutdown) | ❌ No | ❌ Calc'd |
| Cursor IDE | `cursor` | ✅ Done | SQLite (KV) | Exact/Estimated per bubble | ❌ No | ❌ Calc'd |
| Continue | `continue` | ✅ Done | JSON | Exact | ✅ Yes | ✅ Yes |
| JetBrains | `jetbrains` | ✅ Done | JSONL | Estimated | ❌ No | ❌ Calc'd |
| Windsurf/Devin | `windsurf` | ✅ Done | Encrypted .pb | Estimated (best-effort) | ❌ No | ❌ Calc'd |

### CLI Commands

| Command | Status | Description |
|---------|--------|-------------|
| `today` | ✅ Done | Daily session view with token bars and agent badges |
| `daily` | ✅ Done | Per-day historical summary |
| `monthly` | ✅ Done | Monthly aggregate report |
| `session` | ✅ Done | Deep-dive into a single session |
| `doctor` | ✅ Done | Anomaly detection + budget integration |
| `config` | ✅ Done | Pricing and configuration display |
| `blocks` | ✅ Done | 5-hour billing window report |
| `budget` | ✅ New | Monthly budget tracking with progress bar |
| `mcp` | ✅ New | Model Context Protocol stdio server |

### UI Features

| Feature | Status | Description |
|---------|--------|-------------|
| Agent badge colors | ✅ Done | Per-agent colored tags (10 agents) |
| Token source indicators | ✅ Done | `~` suffix for estimated tokens |
| Combined multi-agent view | ✅ Done | Per-agent breakdown strip in today |
| Model tag colors | ✅ Done | Opus/Sonnet/Haiku/GPT badges with versions |
| Cache efficiency bar | ✅ Done | Color-coded cache metrics |
| Watch mode | ✅ Done | Bubbletea TUI with keyboard navigation |
| Theme support | ✅ Done | Dark/light/ANSI themes |
| JSON output | ✅ Done | All commands support --json |
| Agent filtering | ✅ Done | `--agent claude,codex` flag |

### Adapter Architecture

| Component | Status |
|-----------|--------|
| `Adapter` interface | ✅ Done |
| `Registry` with auto-detection | ✅ Done |
| `Discover()` - file discovery | ✅ Done (10 agents) |
| `ParseSession()` - session parsing | ✅ Done (10 agents) |
| `IsAvailable()` - agent detection | ✅ Done |
| `DetectInstalledAgents()` | ✅ Done |
| `--list-agents` flag | ✅ Done |
| `--agent` filter flag | ✅ Done |

### Pricing Support

| Provider | Models | Status |
|----------|--------|--------|
| Anthropic | claude-opus/sonnet/haiku 4.x/3.x | ✅ Done |
| OpenAI | gpt-4o/gpt-4.1/o3/o4-mini/gpt-4.5 | ✅ Done |
| Google | gemini-2.5-flash/pro/2.0-flash | ✅ Done |
| Mistral | Large/Small | ✅ Done |
| Meta | Llama 4 | ✅ Done |
| DeepSeek | R1/V3 | ✅ Done |
| Cohere | Command R+ | ✅ Done |
| Fallback | Unknown models → claude-sonnet-4-6 rates | ✅ Done |

### Infrastructure

| Component | Status |
|-----------|--------|
| Build (Makefile) | ✅ Done |
| Lint (golangci-lint) | ✅ Done |
| Unit tests (43 passing) | ✅ Done |
| Go build (zero deps) | ✅ Done |
| Cross-platform (macOS/Linux/Windows) | ✅ Pure Go |
| Git log | ✅ Committed per issue |
| Research document | ✅ Updated |

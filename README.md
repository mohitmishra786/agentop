# agentop

*"What gets measured gets managed."* — Peter Drucker

A terminal dashboard for AI coding assistant sessions. Token usage, cost, and cache efficiency at a glance.

[![Go Version](https://img.shields.io/badge/go-1.23+-00ADD8?style=flat-square)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue?style=flat-square)](LICENSE)
[![Twitter Follow](https://img.shields.io/twitter/follow/chessMan786?style=flat-square)](https://x.com/chessMan786)

---

## What It Does

Reads your AI assistant's session data and shows you what it is actually doing with your tokens.

![agentop screenshot](assets/agentop.png)

### Default View

```
  total cost          tokens              cache eff           sessions
$38.74              61.5M               93%                 10

╭─────────────────────────────────────────────────────────────────────────────────────╮
│ claude code · 2 sessions  +8 empty   █ in  █ out  █ cc  █ cr                        │
├──────────────────────────┬──────────┬────────────────────────────────────┬────────┬──────────┤
│ SESSION                  │ MODEL    │ TOKENS                             │ CACHE  │ COST     │
├──────────────────────────┼──────────┼────────────────────────────────────┼────────┼──────────┤
│ agentop                  │  sonnet  │ [██████████████████████████]       │    93% │    $8.05 │
│ 89ab341f …/agentop       │          │ 2k in  155k out  721k cc  10.1M cr │        │    9h36m │
│ ↳ 3 subagents            │          │                                    │        │          │
├──────────────────────────┼──────────┼────────────────────────────────────┼────────┼──────────┤
│ build-distributed-s...   │  sonnet  │ [██████████████████████████]       │    93% │   $30.68 │
│ aa0a95ca …uted-systems   │          │ 5k in  257k out  3.4M cc  46.9M cr │        │      N/A │
│ ⎇ fix/task-1-1-weak...   │          │                                    │        │          │
╰──────────────────────────┴──────────┴────────────────────────────────────┴────────┴──────────╯
```

### Table View

```
$ agentop --layout table

┌──────────────────────────────────────────────────────────────────────────────────────────────┐
│ 2 sessions  (+8 empty hidden)                                                                │
├──────────────────────┬──────────┬──────────────────────┬───────┬──────────────────────────────┬────────┬───────┤
│ SESSION              │ MODEL    │ TOKENS               │ CACHE │ IN / OUT / CC / CR           │   COST │  TIME │
├──────────────────────┼──────────┼──────────────────────┼───────┼──────────────────────────────┼────────┼───────┤
│ agentop              │  sonnet  │ [██████████████████] │   93% │ 2k in  155k out  721k cc  10M │  $8.05 │ 9h36m │
├──────────────────────┼──────────┼──────────────────────┼───────┼──────────────────────────────┼────────┼───────┤
│ build-distributed... │  sonnet  │ [██████████████████] │   93% │ 5k in  257k out  3.4M  46.9M │ $30.68 │   N/A │
└──────────────────────┴──────────┴──────────────────────┴───────┴──────────────────────────────┴────────┴───────┘
```

## Features

- **duf-style grid layout**: Clean column-separated table with session, model, token bar, cache efficiency, and cost
- **Color-coded token bar**: Amber=input, coral=output, yellow=cache-create, teal=cache-read — matches the breakdown
- **Empty session filtering**: Zero-token sessions are hidden with a count shown in the title
- **Multiple Layouts**: Default panel view or compact table layout with `--layout table`
- **Interactive Watch Mode**: Keyboard navigation (↑/↓), sort cycling (s), help (?)
- **Color Themes**: Auto-detects terminal theme (dark/light), supports ansi mode
- **Smart Filtering**: By date, project, model
- **Anomaly Detection**: Flags cold starts, high-cost sessions, cache inefficiencies
- **Multi-OS Support**: Linux, FreeBSD, OpenBSD, macOS, Windows — wide architecture coverage

## Install

### Homebrew (macOS / Linux)

```bash
brew tap mohitmishra786/tap
brew install agentop
```

### Prebuilt Binaries

**[Download v0.1.1](https://github.com/mohitmishra786/agentop/releases/tag/v0.1.1)**

Available for:
- **Linux**: x86_64, arm64, i386, armv6, armv7, ppc64le — `.deb`, `.rpm`, `.apk` packages included
- **FreeBSD / OpenBSD**: x86_64, arm64, i386
- **macOS**: x86_64 (Intel), arm64 (Apple Silicon)
- **Windows**: x86_64, i386, arm64

### Go Install

```bash
go install github.com/mohitmishra786/agentop@latest
```

### Build from Source

```bash
git clone https://github.com/mohitmishra786/agentop.git
cd agentop
make build
./bin/agentop
```

### Package Managers (Coming Soon)

| Platform | Manager | Status |
|----------|---------|--------|
| macOS | **Homebrew** | Available — `brew install mohitmishra786/tap/agentop` |
| Linux | Arch (AUR) | In progress |
| Linux | Nix / nixpkgs | In progress |
| Linux | Debian/Ubuntu (apt) | In progress |
| Windows | Scoop | In progress |
| Windows | Chocolatey | Planned |
| BSD | FreeBSD (pkg) | Planned |
| Android | Termux | Planned |

Want to help? Packaging agentop for your favorite package manager is a great contribution — see [CONTRIBUTING.md](CONTRIBUTING.md).

## Usage

```bash
# Today's sessions (default layout)
agentop

# Table layout with all columns
agentop --layout table

# Last 7 days
agentop --since 7d

# Filter by project
agentop --project myapp

# Filter by model
agentop --model sonnet

# Interactive watch mode with live refresh
agentop --watch

# JSON output for scripting
agentop --json

# Anomaly detection (poor cache, high cost, model mismatches)
agentop doctor

# Deep-dive a specific session
agentop session <session-id>

# Daily / monthly breakdowns
agentop daily
agentop monthly

# 5-hour billing windows
agentop blocks

# Show pricing table and config
agentop config
```

## How It Works

Three things happen when you run `agentop`.

**Discovery**: finds every JSONL session file under `~/.claude/projects/`, including subagent sessions nested inside session directories.

**Parsing + deduplication**: AI assistants stream responses, so the same message appears across multiple JSONL lines with zero token counts until the final chunk. The parser collects all chunks per message ID and only counts the final one. Sidechain events and tool-type users are filtered out.

**Cost calculation**: uses the embedded pricing table (`internal/pricing/pricing.json`). For sessions that do not report `costUSD`, it falls back to token-based calculation. Sessions with no cost data show `~` instead of `$0.00`.

## Why This Exists

AI coding assistant sessions cost money. Real money. A single long session can burn through $30+ in tokens — most of it in cache reads you did not know were happening.

The assistant UI shows you a chat log. It does not show you token breakdowns, cache efficiency, or which sessions are burning cash on cold-start cache creation.

**agentop** makes the invisible visible. The anomaly detector (`agentop doctor`) flags sessions with poor cache efficiency, high cost for few messages, and model mismatches.

## Roadmap

- [x] Claude Code support
- [x] duf-style grid UI with column separators
- [x] Homebrew distribution
- [ ] AUR / Nix / Scoop packages
- [ ] Codex CLI support
- [ ] OpenCode support
- [ ] Budget alerts
- [ ] Team/organization dashboards

## Contributing

```bash
git clone https://github.com/mohitmishra786/agentop.git
cd agentop
make test
```

```
cmd/              — CLI commands (cobra)
internal/claude/  — JSONL parsing and session discovery
internal/aggregator/ — Token aggregation and stats
internal/pricing/ — Model pricing engine (update pricing.json to add models)
internal/ui/      — Terminal UI (bubbletea/lipgloss/go-pretty)
testdata/         — Sample session files
```

To add a new model: update `internal/pricing/pricing.json`.  
To add a new command: create `cmd/<name>.go` and register it in `cmd/root.go`.

## Security

- Reads local session data only. Nothing leaves your machine.
- No network calls. No telemetry. No analytics.
- Single Go binary, no runtime dependencies.

---

MIT License

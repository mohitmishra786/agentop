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
│ agentop                  │  sonnet 4.6  │ [██████████████████████████]       │    93% │    $8.05 │
│ 89ab341f …/agentop       │              │ 2k in  155k out  721k cc  10.1M cr │        │    9h36m │
│ ↳ 3 subagents            │              │                                    │        │          │
├──────────────────────────┼──────────────┼────────────────────────────────────┼────────┼──────────┤
│ build-distributed-s...   │  sonnet 4.6  │ [██████████████████████████]       │    93% │   $30.68 │
│ aa0a95ca …uted-systems   │              │ 5k in  257k out  3.4M cc  46.9M cr │        │      N/A │
│ ⎇ fix/task-1-1-weak...   │              │                                    │        │          │
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
│ agentop              │  sonnet 4.6  │ [██████████████████] │   93% │ 2k in  155k out  721k cc  10M │  $8.05 │ 9h36m │
├──────────────────────┼──────────────┼──────────────────────┼───────┼──────────────────────────────┼────────┼───────┤
│ build-distributed... │  sonnet 4.6  │ [██████████████████] │   93% │ 5k in  257k out  3.4M  46.9M │ $30.68 │   N/A │
└──────────────────────┴──────────┴──────────────────────┴───────┴──────────────────────────────┴────────┴───────┘
```

## Features

- **Clean-style grid layout**: Clean column-separated table with session, model, token bar, cache efficiency, and cost
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

### Arch Linux (AUR)

```bash
yay -S agentop-bin
# or: paru -S agentop-bin
```

### Scoop (Windows)

```powershell
scoop bucket add mohitmishra786 https://github.com/mohitmishra786/scoop-bucket
scoop install mohitmishra786/agentop
```

### Fedora / RHEL / EPEL 9 (COPR)

```bash
dnf copr enable chessman/agentop
dnf install agentop
```

> Supports Fedora (latest), RHEL 9, and EPEL 9. The COPR project builds from source on Fedora and uses prebuilt binaries on RHEL/EPEL 9 (where Go 1.25 is unavailable).

### NixOS / Nix

```bash
# Once merged into nixpkgs:
nix-env -iA nixpkgs.agentop
# or in flakes:
nix profile install nixpkgs#agentop
```

> PR under review: [NixOS/nixpkgs#509351](https://github.com/NixOS/nixpkgs/pull/509351)

### Alpine Linux

> APKBUILD under review: [alpine/aports!100647](https://gitlab.alpinelinux.org/alpine/aports/-/merge_requests/100647)

### Go Install

```bash
go install github.com/mohitmishra786/agentop@latest
```

### Prebuilt Binaries

**[Download v0.1.2](https://github.com/mohitmishra786/agentop/releases/tag/v0.1.2)**

| Platform | Format |
|----------|--------|
| Linux (x86_64, arm64, i386, armv6/7, ppc64le) | `.tar.gz`, `.deb`, `.rpm`, `.apk` |
| macOS (Intel + Apple Silicon) | `.tar.gz` |
| Windows (x86_64, arm64, i386) | `.zip` |
| FreeBSD / OpenBSD | `.tar.gz` |

### Build from Source

```bash
git clone https://github.com/mohitmishra786/agentop.git
cd agentop
make build
./bin/agentop
```

### Package Manager Summary

| Platform | Manager | Status | Install |
|----------|---------|--------|---------|
| macOS / Linux | Homebrew | Live | `brew install mohitmishra786/tap/agentop` |
| Linux | Arch AUR | Live | `yay -S agentop-bin` |
| Windows | Scoop | Live | see above |
| Linux | Fedora / RHEL 9 / EPEL 9 COPR | Live | `dnf copr enable chessman/agentop` |
| Linux | Nix | PR open | [NixOS/nixpkgs#509351](https://github.com/NixOS/nixpkgs/pull/509351) |
| Linux | Alpine (apk) | MR open | [aports!100647](https://gitlab.alpinelinux.org/alpine/aports/-/merge_requests/100647) |
| Linux | Debian/Ubuntu | `.deb` on releases page | — |
| BSD | FreeBSD/OpenBSD | `.tar.gz` on releases page | — |
| Windows | Chocolatey | Planned | — |
| Android | Termux | Planned | — |

Want to help package agentop for your favourite distro? See [CONTRIBUTING.md](CONTRIBUTING.md).

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

# Per-session deep-dive with full anomaly detail
agentop doctor <session-id>

# Deep-dive a specific session's raw data
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

**agentop** makes the invisible visible. The anomaly detector (`agentop doctor`) flags sessions with poor cache efficiency, context bloat, model mismatches, and subagent cost — and `agentop doctor <session-id>` gives a full per-session breakdown.

## Roadmap

- [x] Claude Code support
- [x] Clean-style grid UI with column separators
- [x] Homebrew distribution
- [x] AUR package (`agentop-bin`)
- [x] Scoop bucket (Windows)
- [x] Fedora / RHEL 9 COPR (`chessman/agentop`)
- [x] Per-session doctor view (`agentop doctor <session-id>`)
- [x] Full model version in badge (sonnet 4.6, opus 4.7, etc.)
- [x] Windows path detection
- [ ] Nix / Alpine packages (PRs open)
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

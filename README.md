# agentop

*"What gets measured gets managed."* — Peter Drucker

A terminal dashboard for AI coding assistant sessions. Token usage, cost, and cache efficiency at a glance.

[![Go Version](https://img.shields.io/badge/go-1.25+-00ADD8?style=flat-square)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue?style=flat-square)](LICENSE)
[![Twitter Follow](https://img.shields.io/twitter/follow/chessMan786?style=flat-square)](https://x.com/chessMan786)

---

## What It Does

Reads your AI assistant's session data and shows you what it is actually doing with your tokens.

```
$ agentop

 total cost   tokens   cache eff
   $96.97     248.1M      98%

╭──────────────────────────────────────────────────────────────────────╮
│  claude code · 6 sessions                                            │
│ 02ac7e59  glm                        98%  $33.62  19m                │
│ 1.3M in · 355k out · 0 cc · 80.8M cr                                 │
│ ──────────────────────────────────────────────────────────────────── │
│ 5264f18b  glm                        98%  $43.41  2m                 │
│ 2.3M in · 262k out · 0 cc · 108.7M cr                                │
╰──────────────────────────────────────────────────────────────────────╯
```

## Install

### Homebrew (macOS/Linux)

```bash
brew tap mohitmishra786/homebrew-tap
brew install agentop
```

### Go install

```bash
go install github.com/agentop-dev/agentop@latest
```

### Build from source

```bash
git clone https://github.com/mohitmishra786/agentop.git
cd agentop
make build
```

### Binaries

Download prebuilt binaries for your platform from the [Releases](https://github.com/mohitmishra786/agentop/releases) page:

- **Linux**: x86_64, arm64, i386, arm, ppc64le
- **FreeBSD**: x86_64
- **OpenBSD**: x86_64
- **macOS**: x86_64, arm64 (Apple Silicon)
- **Windows**: x86_64, i386

## Usage

```bash
# Today's sessions
agentop

# Last 7 days
agentop --since 7d

# Filter by project
agentop --project myapp

# Filter by model
agentop --model sonnet

# Live watch mode
agentop --watch

# JSON output
agentop --json

# Anomaly detection
agentop doctor

# Deep-dive a session
agentop session <session-id>

# Daily breakdown
agentop daily

# Monthly breakdown
agentop monthly

# 5-hour billing windows
agentop blocks

# Show config and pricing
agentop config
```

## How It Works

Three things happen when you run `agentop`.

First, it discovers every JSONL session file — including subagent sessions. It handles the real directory structure where files live directly in project hash directories.

Second, it parses each session file and deduplicates by message ID. AI assistants stream responses, so the same message appears across multiple lines with zero token counts until the final chunk. The parser collects all chunks per message and only counts the final one. Sidechain events and tool-type users are filtered out.

Third, it calculates cost using the embedded pricing table. For providers that do not report `costUSD`, it falls back to token-based calculation. Sessions with no cost data show `~` instead of `$0.00`.

## Why This Exists

AI coding assistant sessions cost money. Real money. A single long session can burn through $40+ in tokens and most of it goes to cache reads that you did not know were happening.

The assistant UI shows you a chat log. It does not show you token breakdowns. It does not show you cache efficiency. It does not tell you that your first session of the day is paying a premium for cold-start cache creation.

**agentop** makes the invisible visible. You see which sessions are efficient, which ones are burning cash, and which models you are actually using. The anomaly detector (`agentop doctor`) flags sessions with poor cache efficiency, high cost for few messages, and model mismatches.

## Philosophy

*"The first principle is that you must not fool yourself — and you are the easiest person to fool."* — Richard Feynman

Most developers run AI coding assistants for weeks without checking what it costs. The token counts are buried in JSONL files. The cache efficiency is invisible. The billing windows are opaque.

This tool exists because transparency is the first step toward optimization. You cannot improve what you cannot measure.

## Roadmap

- [x] Claude Code support
- [ ] Codex CLI support
- [ ] OpenCode support
- [ ] Custom provider plugins
- [ ] Budget alerts
- [ ] Team/organization dashboards

## Contributing

Contributions are welcome. Here is how to get started:

```bash
git clone https://github.com/mohitmishra786/agentop.git
cd agentop
make test
```

The codebase is organized by layer:

```
cmd/              — CLI commands (cobra)
internal/claude/  — JSONL parsing and session discovery
internal/aggregator/ — Token aggregation and stats
internal/pricing/ — Model pricing engine
internal/ui/      — Terminal UI (bubbletea/lipgloss)
testdata/         — Sample session files
```

To add support for a new model, update `internal/pricing/pricing.json`. To add a new command, create a file in `cmd/` and register it in `cmd/root.go`. To add a new provider, create a new package under `internal/` and wire it into the aggregator.

Run tests before submitting a PR:

```bash
make test
```

## Security

- This tool reads your local session data. Nothing leaves your machine.
- No network calls are made. No telemetry. No analytics.
- The binary is a single Go binary with no runtime dependencies.
- If you find a security issue, please open a private issue or email the maintainer.

---

MIT License

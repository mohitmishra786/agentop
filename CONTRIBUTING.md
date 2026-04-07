# Contributing

## Getting Started

```bash
git clone https://github.com/mohitmishra786/agentop.git
cd agentop
make test
```

## Project Structure

```
cmd/              — CLI commands (cobra)
internal/claude/  — JSONL parsing and session discovery
internal/aggregator/ — Token aggregation and stats
internal/pricing/ — Model pricing engine
internal/ui/      — Terminal UI (bubbletea/lipgloss)
testdata/         — Sample session files
```

## Adding a Model

Update both pricing files:
- `assets/pricing.json`
- `internal/pricing/pricing.json`

Add a model tag style in `internal/ui/styles.go`.

## Adding a Command

1. Create `cmd/<name>.go`
2. Register in `cmd/root.go` `init()`

## Testing

```bash
make test
```

Add test data to `testdata/sessions/` if testing parsing logic.

## Pull Requests

- Keep changes focused and small
- Run `make test` before submitting
- Follow existing code style (no comments, short functions)

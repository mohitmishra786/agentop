# Security Policy

## Supported Versions

Only the latest release is supported.

## Reporting a Vulnerability

Please open a private issue or email the maintainer. Do not disclose the vulnerability publicly until it has been addressed.

## Scope

- Local data handling (JSONL parsing, file discovery)
- No network calls are made by this tool
- All data stays on your machine

## What We Take Seriously

- Path traversal in file discovery
- JSONL parsing edge cases that could cause panics
- Any issue that could expose session data

## What Is Out of Scope

- Upstream dependencies (report to their maintainers)
- The Claude Code application itself

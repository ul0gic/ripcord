RIPCORD — Research Indexed Platform for Channel Observation, Retrieval & Discovery

[![Go Version](https://img.shields.io/badge/go-1.24+-00ADD8?logo=go&logoColor=white)](#) [![Platform](https://img.shields.io/badge/platform-linux%20%7C%20macOS-blue)](#) [![Status](https://img.shields.io/badge/status-alpha-orange)](#)

Ripcord is a no-database, no-frills research tool that lets you sweep a Discord channel and drop the results as JSON or Markdown directly into the directory you ran the command from. Tuned filters (time spans, keywords, user/bot controls) let you dig through history quickly without tripping Discord’s rate limits.

---

## Features at a Glance
- **Token aware:** Works with either `--token`, `DISCORD_TOKEN`, or the built-in `set-token` subcommand that injects credentials into `~/.bashrc`.
- **Flexible filters:** `--days`/`--hours` (required) or `--range`, plus repeatable `--keyword`, `--max`, and bot exclusion toggles.
- **Portable output:** `--format json|markdown|both` and custom filename prefixes; files land in the current working directory.
- **Rate-limit friendly:** Adjustable `--rate` and `--batch-size`, automatic retry/backoff, and detailed scrape stats.
- **Zero infrastructure:** Pure CLI workflow—no database, queues, or external storage required.

---

## Install & Token Setup
```bash
go install github.com/ul0gic/ripcord@latest
ripcord set-token "$DISCORD_TOKEN"
source ~/.bashrc    # reload so future shells inherit the token
```
`go install` drops a single binary into `$(go env GOPATH)/bin` (typically `~/go/bin` on Linux/macOS). The `set-token` command writes a block to your `~/.bashrc`, so Unix shells auto-load the token after `source ~/.bashrc` or starting a new terminal. Use Discord’s “Developer Mode → Copy Channel ID” before scraping.

---

## Quick Start
```bash
ripcord \
  --channel 123456789012345678 \
  --days 3 \
  --keyword breach \
  --keyword poc \
  --format both
```
Result: `discord_123456789012345678_<timestamp>.json` and `.md` written into the current directory, along with a console summary.

---

## CLI Reference
```
╔════════════════════════════════════════════════════════════╗
║ RIPCORD — Discord channel intelligence tool                ║
╚════════════════════════════════════════════════════════════╝
Usage
  ripcord --channel <id> [flags]        Scrape a channel and export history
  ripcord set-token <discord_token>     Store token in ~/.bashrc once
Tokens
  --token <value>    Provide token explicitly (optional when DISCORD_TOKEN exists)
Core Flags
  --channel <id>     REQUIRED channel ID
  --days/--hours     Relative history (required) ·  --range start,end (UTC) absolute window
  --keyword <text>   Repeat for OR matches ·  --include-bots to retain bot posts
Output
  --format json|markdown|both  ·  --output <prefix>  ·  --max <n>  ·  --quiet
Performance
  --batch-size <1-100>          ·  --rate <req/s>
Notes
  • Tokens come from --token, DISCORD_TOKEN, or set-token
  • Filters are client-side; stay within Discord’s ToS and rate limits
  • Multiple --keyword flags create OR-style matching
```

---

## Project Layout
```
ripcord/
├─ main.go          # Entrypoint; routes set-token vs scrape, assembles export summary
├─ cli.go           # Flag parsing, runConfig, fancy usage output
├─ help.go          # ASCII usage banner template
├─ client.go        # Discord API client, pagination, rate limiting, keyword filters
├─ export.go        # JSON + Markdown writers and path helpers
├─ token.go         # Set-token implementation, ~/.bashrc manipulation
├─ types.go         # Shared data structures for messages, exports, stats
├─ constants.go     # API base URL, user agent, batch size caps
└─ go.mod           # Minimal module definition (no external deps yet)
```
Every file is intentionally flat to keep the repo approachable—ideal for quick hacks or contributions.

---

## Notes & Etiquette
- Operate within Discord’s Terms of Service and only scrape content you are authorized to access.
- User tokens can expire; rerun `ripcord set-token <new-token>` if you hit 401s.
- Bump `--rate` gently—Discord enforces global ceilings.
- Markdown exports are designed for human review; JSON retains the normalized schema for tooling.

Have ideas or want to add another output format? Crack open the relevant file (see the project layout table) and go wild.

---

## Contributing & Local Development
- Clone the repo, run `go build ./...`, and keep changes scoped.
- Follow `gofmt` + `staticcheck` before committing: `go fmt ./...` and `staticcheck ./...` (install via `go install honnef.co/go/tools/cmd/staticcheck@latest`).
- Use short commit formats like `feat: add --range flag`; describe CLI behavior and attach sample logs/outputs in PR descriptions.

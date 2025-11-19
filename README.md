[![Go Version](https://img.shields.io/badge/go-1.24+-00ADD8?logo=go&logoColor=white)](#) [![Platform](https://img.shields.io/badge/platform-linux%20%7C%20macOS-blue)](#) [![Status](https://img.shields.io/badge/status-alpha-orange)](#)

<h1 align="center">RIPCORD</h1>
<p align="center">Research Indexed Platform for Channel Observation, Retrieval & Discovery</p>

Ripcord is a no-database, no-frills research tool that lets you sweep a Discord channel and drop the results as JSON or Markdown directly into the directory you ran the command from. Tuned filters (time spans, keywords, user/bot controls) let you dig through history quickly.

---

## Features at a Glance

| Feature | Details |
|---------|---------|
| Token Aware | Works with either `--token`, `DISCORD_TOKEN`, or the built-in `set-token` subcommand that injects credentials into `~/.bashrc`. |
| Flexible Filters | Use `--hours` (1-24) for short runs, `--days` for longer spans, or `--range`, plus repeatable `--keyword`, `--max`, and bot exclusion toggles. |
| Portable Output | `--format json|markdown|both` and custom filename prefixes; both formats land in the current working directory. |
| Zero Infrastructure | Pure CLI workflow—no database, queues, or external storage required. |

---

## Install & Token Setup

1. **Install the binary**

   ```bash
   go install github.com/ul0gic/ripcord@latest
   ```

   This drops the binary in the following locations:

   | OS / Shell | Path after `go install` |
   |-----------|-------------------------|
   | Linux, macOS | `~/go/bin` |
   | Other Unix (custom GOPATH) | `$(go env GOPATH)/bin` |

2. **Store your Discord token**

   ```bash
   ripcord set-token "$DISCORD_TOKEN"
   ```

   This adds the token export to these shell config files:

   | Shell | Config Touched |
   |-------|----------------|
   | bash  | `~/.bashrc`    |
   | zsh   | `~/.zshrc`     |
   | fish  | `~/.config/fish/config.fish` |

   After running `set-token`, reload the shell:

   ```bash
   source ~/.bashrc   # or ~/.zshrc, etc.
   ```

3. **Prepare the channel**

   Enable Discord’s Developer Mode and use **Copy Channel ID** on the channel you want to scrape. Pass that value to `--channel` when running Ripcord.

## CLI Reference

| Category | Flags / Description |
|----------|---------------------|
| Usage | `ripcord --channel <id> [flags]`  Scrape a channel and export history |
| Token | `ripcord set-token <token>`  Writes the shell export so tokens persist |
| Required | `--channel <id>` |
| Relative Window | `--hours <1-24>` for short runs or `--days <n>` for longer spans (at least one required) |
| Range | `--range start,end` (RFC3339 UTC timestamps) |
| Content Filters | Repeat `--keyword foo`; add `--include-bots` to keep bot posts |
| Output | `--format json|markdown|both` · `--output <prefix>` · `--max <n>` · `--quiet` |
| Notes | Tokens are sourced from `--token`, `DISCORD_TOKEN`, or `set-token`. Stay within Discord ToS. |

### CLI Examples

| Goal | Example |
|------|---------|
| Scrape Last 24h | `ripcord --channel 12345 --days 1`
| Scrape Last 12h | `ripcord --channel 12345 --hours 12`
| Range | `ripcord --channel 12345 --range "2025-01-01T00:00:00Z,2025-01-02T00:00:00Z"`
| Keyword Filter | `ripcord --channel 12345 --days 2 --keyword breach --keyword poc`
| Markdown Export | `ripcord --channel 12345 --days 1 --format markdown`
| JSON Export | `ripcord --channel 12345 --days 1 --format json`

---

## Project Layout
```
ripcord/
├─ main.go          # Entrypoint; routes set-token vs scrape, assembles export summary
├─ cli.go           # Flag parsing, runConfig, fancy usage output
├─ help.go          # ASCII usage banner template
├─ client.go        # Discord API client, pagination, keyword filters
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
- Markdown exports are designed for human review; JSON retains the normalized schema for tooling.

Have ideas or want to add another output format? Crack open the relevant file (see the project layout table) and go wild.

---

## Contributing & Local Development
- Clone the repo, run `go build ./...`, and keep changes scoped.
- Follow `gofmt` + `staticcheck` before committing: `go fmt ./...` and `staticcheck ./...` (install via `go install honnef.co/go/tools/cmd/staticcheck@latest`).
- Use short commit formats like `feat: add --range flag`; describe CLI behavior and attach sample logs/outputs in PR descriptions.

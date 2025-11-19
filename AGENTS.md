# Repository Guidelines

## Project Structure & Module Organization
- `main.go` bootstraps the CLI, handles the `set-token` subcommand, and writes exports.
- `cli.go` defines flag parsing (`--channel`, `--days/--hours`, `--range`, etc.) and assembles `runConfig`.
- `client.go` owns Discord API interactions, pagination, throttling, and progress logging.
- `export.go` serializes output to JSON/Markdown; `token.go` manages the `~/.bashrc` token block; `types.go` holds shared structs.
- The repo is intentionally flat—no nested packages. Test assets are not yet present.

## Build, Test, and Development Commands
- `go fmt ./...` – format all Go files.
- `staticcheck ./...` – lint; requires `staticcheck` in `$HOME/go/bin`.
- `go build ./...` – compile the CLI locally.
- `go run . …` – run without installing (e.g., `go run . --channel <id> --days 1`).

## Coding Style & Naming Conventions
- Go 1.24+ with default `gofmt` formatting (tabs for indentation, 120-char lines OK).
- Functions and types use UpperCamelCase; locals use lowerCamelCase.
- Logging is plain `fmt.Printf` (progress) unless `--quiet` is set.

## Testing Guidelines
- No automated tests yet. When adding them, use Go’s `testing` package under `_test.go` files.
- Prefer table-driven tests covering CLI parsing and `DiscordClient` behavior.
- Future test runs should be `go test ./...` (ensure commands do not hit live Discord APIs).

## Commit & Pull Request Guidelines
- Commit messages follow a short `<type>: <summary>` style (e.g., `docs: restyle header`, `feat: add --range flag`).
- Keep changes scoped; run `go fmt`, `staticcheck`, and `go build` before committing.
- PRs should describe the change, include reproduction steps for CLI behavior, and attach sample output snippets when relevant (e.g., JSON diff, command logs).

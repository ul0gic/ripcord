package main

const usageText = `
╔════════════════════════════════════════════════════════════╗
║ RIPCORD — Discord channel intelligence tool                ║
╚════════════════════════════════════════════════════════════╝

Usage
  %s --channel <id> [flags]        Scrape a channel and export history
  %s set-token <discord_token>     Store token in ~/.discord.env (mode 0600)

Tokens
  Resolution order: --token → $DISCORD_TOKEN → $DISCORD_AUTH_TOKEN → ~/.discord.env
  --token <value>                  Provide token explicitly (overrides env + file)

Core Flags
  --channel <id>                   REQUIRED. Channel ID to scrape
  --days <n>                       Relative days window (required if --hours absent)
  --hours <n>                      Relative hours window (required if --days absent)
  --range start,end                Absolute RFC3339 window, e.g. 2025-01-01T00:00:00Z,2025-01-02T00:00:00Z
  --keyword <text>                 Case-insensitive substring filter (repeatable, OR-matched)
  --user <name|id>                 Filter by username, display name, or user ID (repeatable)

Output
  --format json|markdown|both      Export format (default json; "md" accepted as alias)
  --output <prefix>                Filename prefix (default discord_<channel>_<ts>)
  --max <n>                        Stop after N messages (0 = unlimited)
  --quiet                          Suppress progress output (errors still print)

Examples
  # Pull last seven days of history into JSON
  %s --channel 123 --days 7

  # Filter by keywords and export Markdown
  %s --channel 123 --keyword breach --keyword poc --format markdown

  # Set token once and reuse automatically
  %s set-token $DISCORD_TOKEN

Notes
  • set-token writes ~/.discord.env (mode 0600) — no shell sourcing required.
  • Bot messages are always skipped automatically.
  • Output files land in the current working directory.

`

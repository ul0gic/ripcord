package main

const usageText = `
╔════════════════════════════════════════════════════════════╗
║ RIPCORD — Discord channel intelligence tool                ║
╚════════════════════════════════════════════════════════════╝

Usage
  %s --channel <id> [flags]        Scrape a channel and export history
  %s set-token <discord_token>     Store token in ~/.bashrc once

Tokens
  --token <value>                  Provide token explicitly (optional when
                                   DISCORD_TOKEN or set-token is used)

Core Flags
  --channel <id>                   REQUIRED. Channel ID to scrape
  --days <n>                       Relative days window (required if hours absent)
  --hours <n>                      Relative hours window (required if days absent)
  --range start,end                Absolute RFC3339 window (UTC)
  --keyword <text>                 Case-insensitive filter; repeat flag
  --include-bots                   Include bot-authored messages

Output
  --format json|markdown|both      Export format (default json)
  --output <prefix>                Filename prefix (default discord_<channel>_<ts>)
  --max <n>                        Stop after N messages (0=all)
  --quiet                          Only print errors

Examples
  # Pull last seven days of history into JSON
  %s --channel 123 --days 7

  # Filter by keywords and export Markdown
  %s --channel 123 --keyword breach --keyword poc --format markdown

  # Set token once and reuse automatically
  %s set-token $DISCORD_TOKEN

Notes
  • Tokens can come from --token, DISCORD_TOKEN, or set-token.
  • Use multiple --keyword flags for OR-style keyword matching.

`

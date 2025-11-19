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
  --guild <id>                     Guild/server ID (enables jump links)
  --months-back <n>                Months of history (default 3)
  --since <RFC3339>                Absolute start timestamp
  --until <RFC3339>                Absolute end timestamp
  --keyword <text>                 Case-insensitive filter; repeat flag
  --include-bots                   Include bot-authored messages

Output
  --format json|markdown|both      Export format (default json)
  --output <prefix>                Filename prefix (default discord_<channel>_<ts>)
  --max <n>                        Stop after N messages (0=all)
  --quiet                          Only print errors

Performance
  --batch-size <1-100>             Discord page size (default 100)
  --rate <req/s>                   Max Discord requests/sec (default 4)

Examples
  # Pull three months of history into JSON
  %s --channel 123 --guild 999 --months-back 3

  # Filter by keywords and export Markdown
  %s --channel 123 --keyword breach --keyword poc --format markdown

  # Set token once and reuse automatically
  %s set-token $DISCORD_TOKEN

Notes
  • Tokens can come from --token, DISCORD_TOKEN, or set-token.
  • All filters are applied client-side; respect Discord rate limits.
  • Use multiple --keyword flags for OR-style keyword matching.

`

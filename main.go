package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "set-token" {
		var token string
		if len(os.Args) >= 3 {
			token = strings.TrimSpace(os.Args[2])
		}
		if token == "" {
			fmt.Fprintln(os.Stderr, "usage: ripcord set-token <discord_token>")
			os.Exit(1)
		}
		if err := setTokenInBashrc(token); err != nil {
			fmt.Fprintln(os.Stderr, "failed to set token:", err)
			os.Exit(1)
		}
		fmt.Println("✅ Token stored in ~/.bashrc. Run 'source ~/.bashrc' or open a new shell to apply.")
		return
	}

	cfg, err := parseConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	client := NewDiscordClient(cfg.Token)
	messages, stats, err := client.ScrapeChannel(cfg.Options)
	if err != nil {
		fmt.Fprintln(os.Stderr, "scrape failed:", err)
		os.Exit(1)
	}

	if len(messages) == 0 && !cfg.Quiet {
		fmt.Println("no messages matched the provided filters")
	}

	reverseMessages(messages)

	export := Export{
		ChannelID:    cfg.Options.ChannelID,
		ExportedAt:   time.Now().UTC(),
		MessageCount: len(messages),
		Messages:     messages,
		Filters: FilterSummary{
			Since:       cfg.Options.Since,
			Until:       cfg.Options.Until,
			Keywords:    cfg.Options.Keywords,
			Limit:       cfg.Options.MaxMessages,
			IncludeBots: cfg.Options.IncludeBots,
		},
		Stats: stats,
	}

	outputs, err := writeOutputs(export, cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "write failed:", err)
		os.Exit(1)
	}

	if !cfg.Quiet {
		fmt.Printf("wrote %d messages to %s\n", len(messages), strings.Join(outputs, ", "))
	}
}

func reverseMessages(messages []Message) {
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
}

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func writeOutputs(export Export, cfg *runConfig) ([]string, error) {
	var written []string
	switch cfg.Format {
	case "json":
		path := cfg.OutputPrefix
		if !strings.HasSuffix(strings.ToLower(path), ".json") {
			path = path + ".json"
		}
		if err := writeJSON(path, export); err != nil {
			return nil, err
		}
		written = append(written, path)
	case "markdown":
		path := ensureExtension(cfg.OutputPrefix, ".md")
		if err := writeMarkdown(path, export); err != nil {
			return nil, err
		}
		written = append(written, path)
	case "both":
		jsonPath := ensureExtension(cfg.OutputPrefix, ".json")
		mdPath := ensureExtension(cfg.OutputPrefix, ".md")
		if err := writeJSON(jsonPath, export); err != nil {
			return nil, err
		}
		if err := writeMarkdown(mdPath, export); err != nil {
			return nil, err
		}
		written = append(written, jsonPath, mdPath)
	}

	return written, nil
}

func ensureExtension(prefix, ext string) string {
	cleaned := prefix
	if strings.HasSuffix(strings.ToLower(cleaned), strings.ToLower(ext)) {
		return cleaned
	}
	return cleaned + ext
}

func writeJSON(path string, export Export) error {
	file, err := os.Create(filepath.Clean(path))
	if err != nil {
		return err
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	return enc.Encode(export)
}

func writeMarkdown(path string, export Export) error {
	var b strings.Builder
	fmt.Fprintf(&b, "# Discord export for channel %s\n\n", export.ChannelID)
	fmt.Fprintf(&b, "- Exported at: %s\n", export.ExportedAt.Format(time.RFC3339))
	fmt.Fprintf(&b, "- Messages: %d\n", export.MessageCount)
	if export.GuildID != "" {
		fmt.Fprintf(&b, "- Guild: %s\n", export.GuildID)
	}
	if export.Filters.Since != nil {
		fmt.Fprintf(&b, "- Since: %s\n", export.Filters.Since.Format(time.RFC3339))
	}
	if export.Filters.Until != nil {
		fmt.Fprintf(&b, "- Until: %s\n", export.Filters.Until.Format(time.RFC3339))
	}
	if len(export.Filters.Keywords) > 0 {
		fmt.Fprintf(&b, "- Keywords: %s\n", strings.Join(export.Filters.Keywords, ", "))
	}
	if export.Filters.Limit > 0 {
		fmt.Fprintf(&b, "- Limit: %d\n", export.Filters.Limit)
	}
	fmt.Fprintf(&b, "- Include bots: %t\n", export.Filters.IncludeBots)
	if export.Stats.Requests > 0 {
		fmt.Fprintf(&b, "- API requests: %d\n", export.Stats.Requests)
	}
	if export.Stats.RateLimitHits > 0 {
		fmt.Fprintf(&b, "- Rate limit waits: %d\n", export.Stats.RateLimitHits)
	}

	for _, msg := range export.Messages {
		fmt.Fprintf(&b, "\n## %s — %s\n\n", msg.Timestamp.Format("2006-01-02 15:04:05 MST"), describeAuthor(msg.Author))
		if msg.Content != "" {
			fmt.Fprintf(&b, "%s\n\n", msg.Content)
		}
		if len(msg.Attachments) > 0 {
			fmt.Fprintf(&b, "**Attachments:**\n")
			for _, att := range msg.Attachments {
				fmt.Fprintf(&b, "- [%s](%s)\n", att.Filename, att.URL)
			}
			fmt.Fprintf(&b, "\n")
		}
		if len(msg.Reactions) > 0 {
			parts := make([]string, 0, len(msg.Reactions))
			for _, react := range msg.Reactions {
				parts = append(parts, fmt.Sprintf("%s ×%d", react.Emoji, react.Count))
			}
			fmt.Fprintf(&b, "**Reactions:** %s\n\n", strings.Join(parts, ", "))
		}
		if msg.JumpURL != "" {
			fmt.Fprintf(&b, "[Jump to message](%s)\n\n", msg.JumpURL)
		}
	}

	return os.WriteFile(filepath.Clean(path), []byte(b.String()), 0o644)
}

func describeAuthor(author Author) string {
	if author.DisplayName != "" && author.DisplayName != author.Username {
		return fmt.Sprintf("%s (%s)", author.DisplayName, author.Username)
	}
	return author.Username
}

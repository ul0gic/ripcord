package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func writeOutputs(export *Export, cfg *runConfig) ([]string, error) {
	var written []string
	switch cfg.Format {
	case "json":
		path := cfg.OutputPrefix
		if !strings.HasSuffix(strings.ToLower(path), ".json") {
			path += ".json"
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

func writeJSON(path string, export *Export) (err error) {
	file, err := os.Create(filepath.Clean(path))
	if err != nil {
		return err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	return enc.Encode(export)
}

func writeMarkdown(path string, export *Export) error {
	var b strings.Builder
	fmt.Fprintf(&b, "# Discord export for channel %s\n\n", export.ChannelID)
	fmt.Fprintf(&b, "- Exported at: %s\n", export.ExportedAt.Format(time.RFC3339))
	fmt.Fprintf(&b, "- Messages: %d\n", export.MessageCount)
	if export.Filters.Since != nil {
		fmt.Fprintf(&b, "- Since: %s\n", export.Filters.Since.Format(time.RFC3339))
	}
	if export.Filters.Until != nil {
		fmt.Fprintf(&b, "- Until: %s\n", export.Filters.Until.Format(time.RFC3339))
	}
	if len(export.Filters.Keywords) > 0 {
		fmt.Fprintf(&b, "- Keywords: %s\n", strings.Join(export.Filters.Keywords, ", "))
	}
	if len(export.Filters.Users) > 0 {
		fmt.Fprintf(&b, "- Users: %s\n", strings.Join(export.Filters.Users, ", "))
	}
	if export.Filters.Limit > 0 {
		fmt.Fprintf(&b, "- Limit: %d\n", export.Filters.Limit)
	}
	if export.Stats.Requests > 0 {
		fmt.Fprintf(&b, "- API requests: %d\n", export.Stats.Requests)
	}
	if export.Stats.RateLimitHits > 0 {
		fmt.Fprintf(&b, "- Rate limit waits: %d\n", export.Stats.RateLimitHits)
	}

	for i := range export.Messages {
		msg := &export.Messages[i]
		fmt.Fprintf(&b, "\n## %s — %s\n\n", msg.Timestamp.Format("2006-01-02 15:04:05 MST"), describeAuthor(&msg.Author))
		if msg.Content != "" {
			fmt.Fprintf(&b, "%s\n\n", msg.Content)
		}
		if len(msg.Attachments) > 0 {
			fmt.Fprintln(&b, "**Attachments:**")
			for j := range msg.Attachments {
				att := &msg.Attachments[j]
				fmt.Fprintf(&b, "- [%s](%s)\n", att.Filename, att.URL)
			}
			fmt.Fprintln(&b)
		}
		if len(msg.Reactions) > 0 {
			parts := make([]string, 0, len(msg.Reactions))
			for j := range msg.Reactions {
				react := &msg.Reactions[j]
				parts = append(parts, fmt.Sprintf("%s ×%d", react.Emoji, react.Count))
			}
			fmt.Fprintf(&b, "**Reactions:** %s\n\n", strings.Join(parts, ", "))
		}
	}

	return os.WriteFile(filepath.Clean(path), []byte(b.String()), 0o600)
}

func describeAuthor(author *Author) string {
	if author.DisplayName != "" && author.DisplayName != author.Username {
		return fmt.Sprintf("%s (%s)", author.DisplayName, author.Username)
	}
	return author.Username
}

package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type multiValue []string

type runConfig struct {
	Token        string
	OutputPrefix string
	Format       string
	Quiet        bool
	Options      scrapeOptions
}

type scrapeOptions struct {
	ChannelID   string
	Keywords    []string
	Users       []string
	MaxMessages int
	Since       *time.Time
	Until       *time.Time
	Quiet       bool
}

func (m *multiValue) String() string {
	return strings.Join(*m, ",")
}

func (m *multiValue) Set(value string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed != "" {
		*m = append(*m, trimmed)
	}
	return nil
}

func parseConfig() (*runConfig, error) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flag.CommandLine.Usage = func() {
		printUsage(flag.CommandLine.Output(), os.Args[0])
	}

	token := flag.String("token", "", "Discord bot/user token (or set DISCORD_TOKEN)")
	channel := flag.String("channel", "", "Channel ID to scrape (required)")
	daysBack := flag.Int("days", 0, "Relative days window (required if --hours absent)")
	hoursBack := flag.Int("hours", 0, "Relative hours window (required if --days absent)")
	rangeStr := flag.String("range", "", "Absolute window start,end (RFC3339)")
	maxMessages := flag.Int("max", 0, "Stop after collecting this many messages (0 = unlimited)")
	var users multiValue
	flag.Var(&users, "user", "Filter by username or ID (repeatable)")
	format := flag.String("format", "json", "Output format: json, markdown, or both")
	output := flag.String("output", "", "Output filename prefix (default discord_<channel>_<timestamp>)")
	quiet := flag.Bool("quiet", false, "Only print errors")

	var keywords multiValue
	flag.Var(&keywords, "keyword", "Case-insensitive keyword filter (repeatable)")

	flag.Parse()

	resolvedToken := resolveToken(*token)
	if resolvedToken == "" {
		return nil, errors.New("missing Discord token (pass --token, set DISCORD_TOKEN, or run `ripcord set-token`)")
	}

	if *channel == "" {
		return nil, errors.New("--channel is required")
	}

	fmtChoice, err := normalizeFormat(*format)
	if err != nil {
		return nil, err
	}

	since, until, err := resolveTimeWindow(*rangeStr, *daysBack, *hoursBack)
	if err != nil {
		return nil, err
	}

	cfg := &runConfig{
		Token:        resolvedToken,
		OutputPrefix: resolveOutputPrefix(*output, *channel),
		Format:       fmtChoice,
		Quiet:        *quiet,
		Options: scrapeOptions{
			ChannelID:   *channel,
			Keywords:    normalizeStringList(keywords),
			Users:       normalizeStringList(users),
			MaxMessages: *maxMessages,
			Since:       since,
			Until:       until,
			Quiet:       *quiet,
		},
	}

	return cfg, nil
}

func resolveToken(flagValue string) string {
	if t := strings.TrimSpace(flagValue); t != "" {
		return t
	}
	if t := strings.TrimSpace(os.Getenv("DISCORD_TOKEN")); t != "" {
		return t
	}
	if t := strings.TrimSpace(os.Getenv("DISCORD_AUTH_TOKEN")); t != "" {
		return t
	}
	return readTokenFromEnvFile()
}

func normalizeFormat(format string) (string, error) {
	choice := strings.ToLower(strings.TrimSpace(format))
	switch choice {
	case "json", "markdown", "both":
		return choice, nil
	case "md":
		return "markdown", nil
	}
	return "", errors.New("format must be one of json, markdown, or both")
}

func resolveTimeWindow(rangeStr string, daysBack, hoursBack int) (since, until *time.Time, err error) {
	if strings.TrimSpace(rangeStr) != "" {
		start, end, perr := parseRange(rangeStr)
		if perr != nil {
			return nil, nil, fmt.Errorf("invalid --range value: %w", perr)
		}
		return &start, &end, nil
	}
	if daysBack > 0 || hoursBack > 0 {
		totalHours := (daysBack * 24) + hoursBack
		if totalHours <= 0 {
			return nil, nil, errors.New("days/hours window must be positive")
		}
		cutoff := time.Now().UTC().Add(-time.Duration(totalHours) * time.Hour)
		return &cutoff, nil, nil
	}
	return nil, nil, errors.New("specify --range or a --days/--hours window")
}

func resolveOutputPrefix(flagValue, channel string) string {
	prefix := strings.TrimSpace(flagValue)
	if prefix != "" {
		return prefix
	}
	timestamp := time.Now().UTC().Format("20060102T150405Z")
	return fmt.Sprintf("discord_%s_%s", channel, timestamp)
}

func normalizeStringList(values []string) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		if trimmed := strings.ToLower(strings.TrimSpace(v)); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func parseTimestamp(value string) (time.Time, error) {
	v := strings.TrimSpace(value)
	if v == "" {
		return time.Time{}, errors.New("empty time value")
	}
	parsed, err := time.Parse(time.RFC3339, v)
	if err == nil {
		return parsed.UTC(), nil
	}
	parsed, err = time.Parse(time.RFC3339Nano, v)
	if err == nil {
		return parsed.UTC(), nil
	}
	return time.Time{}, err
}

func parseRange(value string) (start, end time.Time, err error) {
	v := strings.TrimSpace(value)
	parts := strings.Split(v, ",")
	if len(parts) != 2 {
		parts = strings.Split(v, "..")
		if len(parts) != 2 {
			return time.Time{}, time.Time{}, errors.New("range must be start,end")
		}
	}
	start, err = parseTimestamp(parts[0])
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start: %w", err)
	}
	end, err = parseTimestamp(parts[1])
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid end: %w", err)
	}
	if start.After(end) {
		return time.Time{}, time.Time{}, errors.New("range start must be before end")
	}
	return start, end, nil
}

func printUsage(w io.Writer, bin string) {
	if _, err := fmt.Fprintf(w, usageText, bin, bin, bin, bin, bin); err != nil {
		fmt.Fprintln(os.Stderr, "warning: failed to write usage:", err)
	}
}

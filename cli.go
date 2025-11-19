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
	GuildID     string
	Keywords    []string
	IncludeBots bool
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
	guild := flag.String("guild", "", "Guild/server ID for jump links (optional)")
	daysBack := flag.Int("days", 0, "Relative days window (required if --hours absent)")
	hoursBack := flag.Int("hours", 0, "Relative hours window (required if --days absent)")
	rangeStr := flag.String("range", "", "Absolute window start,end (RFC3339)")
	maxMessages := flag.Int("max", 0, "Stop after collecting this many messages (0 = unlimited)")
	includeBots := flag.Bool("include-bots", false, "Include messages from bot accounts")
	format := flag.String("format", "json", "Output format: json, markdown, or both")
	output := flag.String("output", "", "Output filename prefix (default discord_<channel>_<timestamp>)")
	quiet := flag.Bool("quiet", false, "Only print errors")

	var keywords multiValue
	flag.Var(&keywords, "keyword", "Case-insensitive keyword filter (repeatable)")

	flag.Parse()

	resolvedToken := strings.TrimSpace(*token)
	if resolvedToken == "" {
		resolvedToken = strings.TrimSpace(os.Getenv("DISCORD_TOKEN"))
	}
	if resolvedToken == "" {
		resolvedToken = strings.TrimSpace(os.Getenv("DISCORD_AUTH_TOKEN"))
	}

	if resolvedToken == "" {
		return nil, errors.New("missing Discord token (pass --token or set DISCORD_TOKEN)")
	}

	if *channel == "" {
		return nil, errors.New("--channel is required")
	}

	fmtChoice := strings.ToLower(strings.TrimSpace(*format))
	switch fmtChoice {
	case "json", "markdown", "md", "both":
	default:
		return nil, errors.New("format must be one of json, markdown, or both")
	}
	if fmtChoice == "md" {
		fmtChoice = "markdown"
	}

	var since *time.Time
	var until *time.Time
	if strings.TrimSpace(*rangeStr) != "" {
		start, end, err := parseRange(*rangeStr)
		if err != nil {
			return nil, fmt.Errorf("invalid --range value: %w", err)
		}
		since = &start
		until = &end
	} else if *daysBack > 0 || *hoursBack > 0 {
		totalHours := (*daysBack * 24) + *hoursBack
		if totalHours <= 0 {
			return nil, errors.New("days/hours window must be positive")
		}
		dur := time.Duration(totalHours) * time.Hour
		cutoff := time.Now().UTC().Add(-dur)
		since = &cutoff
	} else {
		return nil, errors.New("specify --range or a --days/--hours window")
	}

	keywordList := make([]string, 0, len(keywords))
	for _, kw := range keywords {
		if trimmed := strings.ToLower(strings.TrimSpace(kw)); trimmed != "" {
			keywordList = append(keywordList, trimmed)
		}
	}

	prefix := strings.TrimSpace(*output)
	if prefix == "" {
		timestamp := time.Now().UTC().Format("20060102T150405Z")
		prefix = fmt.Sprintf("discord_%s_%s", *channel, timestamp)
	}

	cfg := &runConfig{
		Token:        resolvedToken,
		OutputPrefix: prefix,
		Format:       fmtChoice,
		Quiet:        *quiet,
		Options: scrapeOptions{
			ChannelID:   *channel,
			GuildID:     strings.TrimSpace(*guild),
			Keywords:    keywordList,
			IncludeBots: *includeBots,
			MaxMessages: *maxMessages,
			Since:       since,
			Until:       until,
			Quiet:       *quiet,
		},
	}

	return cfg, nil
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

func parseRange(value string) (time.Time, time.Time, error) {
	v := strings.TrimSpace(value)
	parts := strings.Split(v, ",")
	if len(parts) != 2 {
		parts = strings.Split(v, "..")
		if len(parts) != 2 {
			return time.Time{}, time.Time{}, errors.New("range must be start,end")
		}
	}
	start, err := parseTimestamp(parts[0])
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start: %w", err)
	}
	end, err := parseTimestamp(parts[1])
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid end: %w", err)
	}
	if start.After(end) {
		return time.Time{}, time.Time{}, errors.New("range start must be before end")
	}
	return start, end, nil
}

func printUsage(w io.Writer, bin string) {
	fmt.Fprintf(w, usageText, bin, bin, bin, bin, bin)
}

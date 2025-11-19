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
	RateLimit    float64
	Quiet        bool
	Options      scrapeOptions
}

type scrapeOptions struct {
	ChannelID   string
	GuildID     string
	Keywords    []string
	IncludeBots bool
	MaxMessages int
	BatchSize   int
	Since       *time.Time
	Until       *time.Time
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
	monthsBack := flag.Int("months-back", 3, "How many months of history to pull (ignored if --since provided)")
	sinceStr := flag.String("since", "", "Only include messages on/after this RFC3339 timestamp")
	untilStr := flag.String("until", "", "Only include messages on/before this RFC3339 timestamp")
	batchSize := flag.Int("batch-size", 100, "Messages per request (1-100)")
	maxMessages := flag.Int("max", 0, "Stop after collecting this many messages (0 = unlimited)")
	rateLimit := flag.Float64("rate", 4.0, "Max API requests per second")
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

	if *batchSize < 1 || *batchSize > maxBatchSize {
		return nil, fmt.Errorf("batch-size must be between 1 and %d", maxBatchSize)
	}

	if *rateLimit <= 0 {
		return nil, errors.New("rate must be greater than zero")
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
	if strings.TrimSpace(*sinceStr) != "" {
		parsed, err := parseTimestamp(*sinceStr)
		if err != nil {
			return nil, fmt.Errorf("invalid --since value: %w", err)
		}
		since = &parsed
	} else if *monthsBack > 0 {
		cutoff := time.Now().UTC().AddDate(0, -*monthsBack, 0)
		since = &cutoff
	}

	var until *time.Time
	if strings.TrimSpace(*untilStr) != "" {
		parsed, err := parseTimestamp(*untilStr)
		if err != nil {
			return nil, fmt.Errorf("invalid --until value: %w", err)
		}
		until = &parsed
	}

	if since != nil && until != nil && since.After(*until) {
		return nil, errors.New("--since must be before --until")
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
		RateLimit:    *rateLimit,
		Quiet:        *quiet,
		Options: scrapeOptions{
			ChannelID:   *channel,
			GuildID:     strings.TrimSpace(*guild),
			Keywords:    keywordList,
			IncludeBots: *includeBots,
			MaxMessages: *maxMessages,
			BatchSize:   *batchSize,
			Since:       since,
			Until:       until,
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

func printUsage(w io.Writer, bin string) {
	fmt.Fprintf(w, usageText, bin, bin, bin, bin, bin)
}

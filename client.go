package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type DiscordClient struct {
	token       string
	httpClient  *http.Client
	minInterval time.Duration
	lastRequest time.Time
}

func NewDiscordClient(token string) *DiscordClient {
	base := time.Second
	interval := time.Duration(float64(base) / defaultRateLimit)
	return &DiscordClient{
		token:       token,
		httpClient:  &http.Client{Timeout: 15 * time.Second},
		minInterval: interval,
	}
}

func (c *DiscordClient) ScrapeChannel(opts *scrapeOptions) ([]Message, Stats, error) {
	var results []Message
	var stats Stats
	var before string
	keywords := normalizeFilters(opts.Keywords)
	users := normalizeFilters(opts.Users)

	for {
		batch, metrics, err := c.fetchBatch(opts.ChannelID, before, maxBatchSize)
		stats.Requests += metrics.requests
		stats.RateLimitHits += metrics.rateLimitHits

		if err != nil {
			if errors.Is(err, errNoMoreMessages) {
				break
			}
			return nil, stats, err
		}

		if len(batch) == 0 {
			break
		}

		var stop bool
		results, stop = collectBatch(batch, results, opts, users, keywords)
		if opts.MaxMessages > 0 && len(results) >= opts.MaxMessages {
			return results[:opts.MaxMessages], stats, nil
		}
		if stop {
			break
		}

		before = batch[len(batch)-1].ID
		if !opts.Quiet {
			fmt.Printf("pulled %d messages so far\n", len(results))
		}
	}

	return results, stats, nil
}

// collectBatch filters one page of API messages and appends keepers to results.
// Returns updated results and stop=true when a message older than opts.Since is
// reached (which means pagination should halt).
func collectBatch(batch []apiMessage, results []Message, opts *scrapeOptions, users, keywords []string) ([]Message, bool) {
	for i := range batch {
		raw := &batch[i]
		if raw.Author.Bot {
			continue
		}

		msgTime, ok := parseMessageTime(raw.Timestamp)
		if !ok {
			continue
		}

		if opts.Until != nil && msgTime.After(opts.Until.UTC()) {
			continue
		}
		if opts.Since != nil && msgTime.Before(opts.Since.UTC()) {
			return results, true
		}

		normalized := normalizeMessage(raw, msgTime)
		if !messagePassesFilters(&normalized, users, keywords) {
			continue
		}

		results = append(results, normalized)
	}
	return results, false
}

func parseMessageTime(timestamp string) (time.Time, bool) {
	if t, err := time.Parse(time.RFC3339Nano, timestamp); err == nil {
		return t.UTC(), true
	}
	if t, err := time.Parse(time.RFC3339, timestamp); err == nil {
		return t.UTC(), true
	}
	return time.Time{}, false
}

func messagePassesFilters(msg *Message, users, keywords []string) bool {
	if len(users) > 0 && !matchesUsers(&msg.Author, users) {
		return false
	}
	if len(keywords) > 0 && !matchesKeywords(msg.Content, keywords) {
		return false
	}
	if msg.Content == "" && len(msg.Attachments) == 0 && msg.EmbedCount == 0 {
		return false
	}
	return true
}

var errNoMoreMessages = errors.New("no more messages")

func (c *DiscordClient) fetchBatch(channelID, before string, limit int) ([]apiMessage, batchMetrics, error) {
	var metrics batchMetrics
	endpoint := fmt.Sprintf("%s/channels/%s/messages", apiBase, channelID)

	params := url.Values{}
	params.Set("limit", strconv.Itoa(limit))
	if before != "" {
		params.Set("before", before)
	}

	var lastErr error
	for attempt := 0; attempt < 5; attempt++ {
		c.throttle()

		req, err := http.NewRequest(http.MethodGet, endpoint+"?"+params.Encode(), http.NoBody)
		if err != nil {
			return nil, metrics, err
		}
		req.Header.Set("Authorization", c.token)
		req.Header.Set("User-Agent", userAgent)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(backoffDuration(attempt))
			continue
		}

		metrics.requests++

		body, readErr := io.ReadAll(resp.Body)
		closeErr := resp.Body.Close()
		if combined := errors.Join(readErr, closeErr); combined != nil {
			lastErr = combined
			time.Sleep(backoffDuration(attempt))
			continue
		}

		if resp.StatusCode == http.StatusOK {
			var messages []apiMessage
			if err := json.Unmarshal(body, &messages); err != nil {
				return nil, metrics, err
			}
			if len(messages) == 0 {
				return nil, metrics, errNoMoreMessages
			}
			return messages, metrics, nil
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			metrics.rateLimitHits++
			retryAfter := parseRetryAfter(body)
			time.Sleep(retryAfter)
			continue
		}

		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("discord api error %d", resp.StatusCode)
			time.Sleep(backoffDuration(attempt))
			continue
		}

		return nil, metrics, fmt.Errorf("discord api returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	if lastErr != nil {
		return nil, metrics, lastErr
	}
	return nil, metrics, errors.New("maximum retries exceeded")
}

func (c *DiscordClient) throttle() {
	if c.minInterval <= 0 {
		return
	}
	since := time.Since(c.lastRequest)
	if since < c.minInterval {
		time.Sleep(c.minInterval - since)
	}
	c.lastRequest = time.Now()
}

func parseRetryAfter(body []byte) time.Duration {
	var payload struct {
		RetryAfter float64 `json:"retry_after"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return 3 * time.Second
	}
	if payload.RetryAfter <= 0 {
		return 3 * time.Second
	}
	return time.Duration(payload.RetryAfter * float64(time.Second))
}

func backoffDuration(attempt int) time.Duration {
	if attempt <= 0 {
		return 500 * time.Millisecond
	}
	d := time.Duration(1<<attempt) * time.Second
	if d > 30*time.Second {
		d = 30 * time.Second
	}
	return d
}

func normalizeMessage(raw *apiMessage, timestamp time.Time) Message {
	msg := Message{
		ID:        raw.ID,
		ChannelID: raw.ChannelID,
		Author: Author{
			ID:          raw.Author.ID,
			Username:    chooseName(raw.Author.Username, raw.Author.GlobalName),
			DisplayName: chooseDisplayName(&raw.Author),
			Bot:         raw.Author.Bot,
		},
		Content:    raw.Content,
		Timestamp:  timestamp,
		Type:       raw.Type,
		EmbedCount: len(raw.Embeds),
	}

	if raw.EditedTimestamp != nil {
		if t, err := time.Parse(time.RFC3339Nano, *raw.EditedTimestamp); err == nil {
			parsed := t.UTC()
			msg.EditedTimestamp = &parsed
		}
	}

	if len(raw.Mentions) > 0 {
		mentions := make([]string, 0, len(raw.Mentions))
		for i := range raw.Mentions {
			mentions = append(mentions, raw.Mentions[i].ID)
		}
		msg.MentionUserIDs = mentions
	}

	if len(raw.MentionRoles) > 0 {
		msg.MentionRoleIDs = append([]string{}, raw.MentionRoles...)
	}

	if len(raw.Attachments) > 0 {
		attachments := make([]Attachment, 0, len(raw.Attachments))
		for i := range raw.Attachments {
			attachments = append(attachments, Attachment(raw.Attachments[i]))
		}
		msg.Attachments = attachments
	}

	if len(raw.Reactions) > 0 {
		reactions := make([]Reaction, 0, len(raw.Reactions))
		for i := range raw.Reactions {
			react := &raw.Reactions[i]
			emoji := react.Emoji.Name
			if emoji == "" && react.Emoji.ID != "" {
				emoji = fmt.Sprintf(":%s:", react.Emoji.ID)
			}
			reactions = append(reactions, Reaction{
				Emoji: emoji,
				Count: react.Count,
			})
		}
		msg.Reactions = reactions
	}

	if raw.ReferencedMessage != nil {
		msg.ReplyTo = &ReplyReference{
			MessageID: raw.ReferencedMessage.ID,
			AuthorID:  raw.ReferencedMessage.Author.ID,
		}
	}

	return msg
}

func chooseName(username, global string) string {
	if global != "" {
		return global
	}
	return username
}

func chooseDisplayName(author *apiAuthor) string {
	if author.DisplayName != "" {
		return author.DisplayName
	}
	if author.GlobalName != "" {
		return author.GlobalName
	}
	return ""
}

func normalizeFilters(values []string) []string {
	lower := make([]string, 0, len(values))
	seen := make(map[string]struct{})
	for _, kw := range values {
		trimmed := strings.ToLower(strings.TrimSpace(kw))
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		lower = append(lower, trimmed)
	}
	return lower
}

func matchesKeywords(content string, keywords []string) bool {
	if len(keywords) == 0 {
		return true
	}
	text := strings.ToLower(content)
	for _, kw := range keywords {
		if strings.Contains(text, kw) {
			return true
		}
	}
	return false
}

func matchesUsers(author *Author, users []string) bool {
	if len(users) == 0 {
		return true
	}
	username := strings.ToLower(author.Username)
	display := strings.ToLower(author.DisplayName)
	id := strings.ToLower(author.ID)
	for _, u := range users {
		if u == "" {
			continue
		}
		if username == u || (display != "" && display == u) || id == u {
			return true
		}
	}
	return false
}

package main

import (
	"encoding/json"
	"time"
)

type Export struct {
	ChannelID    string        `json:"channel_id"`
	ExportedAt   time.Time     `json:"exported_at"`
	MessageCount int           `json:"message_count"`
	Messages     []Message     `json:"messages"`
	Filters      FilterSummary `json:"filters"`
	Stats        Stats         `json:"stats"`
}

type FilterSummary struct {
	Since       *time.Time `json:"since,omitempty"`
	Until       *time.Time `json:"until,omitempty"`
	Keywords    []string   `json:"keywords,omitempty"`
	Limit       int        `json:"limit,omitempty"`
	IncludeBots bool       `json:"include_bots"`
}

type Stats struct {
	Requests      int `json:"api_requests"`
	RateLimitHits int `json:"rate_limit_hits"`
}

type Message struct {
	ID              string          `json:"id"`
	ChannelID       string          `json:"channel_id"`
	Author          Author          `json:"author"`
	Content         string          `json:"content"`
	Timestamp       time.Time       `json:"timestamp"`
	EditedTimestamp *time.Time      `json:"edited_timestamp,omitempty"`
	MentionUserIDs  []string        `json:"mention_user_ids,omitempty"`
	MentionRoleIDs  []string        `json:"mention_role_ids,omitempty"`
	Attachments     []Attachment    `json:"attachments,omitempty"`
	Reactions       []Reaction      `json:"reactions,omitempty"`
	ReplyTo         *ReplyReference `json:"reply_to,omitempty"`
	Type            int             `json:"type"`
	EmbedCount      int             `json:"embed_count,omitempty"`
}

type Author struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name,omitempty"`
	Bot         bool   `json:"bot"`
}

type Attachment struct {
	ID          string `json:"id"`
	Filename    string `json:"filename"`
	URL         string `json:"url"`
	ContentType string `json:"content_type,omitempty"`
	Size        int64  `json:"size_bytes,omitempty"`
}

type Reaction struct {
	Emoji string `json:"emoji"`
	Count int    `json:"count"`
}

type ReplyReference struct {
	MessageID string `json:"message_id"`
	AuthorID  string `json:"author_id,omitempty"`
}

type apiMessage struct {
	ID                string            `json:"id"`
	ChannelID         string            `json:"channel_id"`
	Content           string            `json:"content"`
	Timestamp         string            `json:"timestamp"`
	EditedTimestamp   *string           `json:"edited_timestamp"`
	Author            apiAuthor         `json:"author"`
	Mentions          []apiUser         `json:"mentions"`
	MentionRoles      []string          `json:"mention_roles"`
	Attachments       []apiAttachment   `json:"attachments"`
	Reactions         []apiReaction     `json:"reactions"`
	Embeds            []json.RawMessage `json:"embeds"`
	Type              int               `json:"type"`
	ReferencedMessage *apiRefMessage    `json:"referenced_message"`
}

type apiAuthor struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	GlobalName  string `json:"global_name"`
	DisplayName string `json:"display_name"`
	Bot         bool   `json:"bot"`
}

type apiUser struct {
	ID         string `json:"id"`
	Username   string `json:"username"`
	GlobalName string `json:"global_name"`
}

type apiAttachment struct {
	ID          string `json:"id"`
	Filename    string `json:"filename"`
	URL         string `json:"url"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
}

type apiReaction struct {
	Count int      `json:"count"`
	Emoji apiEmoji `json:"emoji"`
}

type apiEmoji struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type apiRefMessage struct {
	ID     string    `json:"id"`
	Author apiAuthor `json:"author"`
}

type batchMetrics struct {
	requests      int
	rateLimitHits int
}

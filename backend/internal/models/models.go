package models

import "time"

type SourceType string

const (
	SourceReddit     SourceType = "reddit"
	SourceHackerNews SourceType = "hackernews"
	SourceRSS        SourceType = "rss"
	SourceSECEDGAR   SourceType = "secedgar"
)

type Source struct {
	ID        int64      `json:"id"`
	Type      SourceType `json:"type"`
	Name      string     `json:"name"`
	URL       string     `json:"url"`
	Interval  int        `json:"interval"` // poll interval in seconds
	Enabled   bool       `json:"enabled"`
	CreatedAt time.Time  `json:"created_at"`
}

type Topic struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Keywords  string    `json:"keywords"` // comma-separated, or newline-separated
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
}

type Item struct {
	ID          string     `json:"id"`
	SourceID    int64      `json:"source_id"`
	SourceType  SourceType `json:"source_type"`
	SourceName  string     `json:"source_name"`
	Title       string     `json:"title"`
	Body        string     `json:"body"`
	URL         string     `json:"url"`
	Author      string     `json:"author"`
	PublishedAt time.Time  `json:"published_at"`
	FetchedAt   time.Time  `json:"fetched_at"`
	Metadata    string     `json:"metadata"` // JSON blob for source-specific data
}

type ItemTopic struct {
	ItemID  string `json:"item_id"`
	TopicID int64  `json:"topic_id"`
}

type WatchItem struct {
	ID              int64     `json:"id"`
	Name            string    `json:"name"`
	SearchTerm      string    `json:"search_term"`
	Threshold       float64   `json:"threshold"`
	Floor           float64   `json:"floor"`
	Category        string    `json:"category"`
	Active          bool      `json:"active"`
	EbayPrice       *float64  `json:"ebay_price"`
	SlickdealsPrice *float64  `json:"slickdeals_price"`
	RedditPrice     *float64  `json:"reddit_price"`
	LastChecked     *string   `json:"last_checked"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type SaleAlert struct {
	ID        int64     `json:"id"`
	ItemID    int64     `json:"item_id"`
	ItemName  string    `json:"item_name"`
	Price     float64   `json:"price"`
	Title     string    `json:"title"`
	DealURL   string    `json:"deal_url"`
	Source    string    `json:"source"`
	Sent      bool      `json:"sent"`
	Dismissed bool      `json:"dismissed"`
	CreatedAt time.Time `json:"created_at"`
}

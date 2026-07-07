package adapters

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/matthewaquino/timeline/internal/models"
)

type RSSAdapter struct {
	client *http.Client
}

func NewRSSAdapter() *RSSAdapter {
	return &RSSAdapter{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (a *RSSAdapter) Fetch(source models.Source) ([]models.Item, error) {
	req, err := http.NewRequest("GET", source.URL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "timeline-aggregator/1.0")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("rss fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		retryAfter := resp.Header.Get("Retry-After")
		if retryAfter != "" {
			return nil, fmt.Errorf("rss returned status %d (retry-after: %ss)", resp.StatusCode, retryAfter)
		}
		return nil, fmt.Errorf("rss returned status %d", resp.StatusCode)
	}

	fp := NewFeedParser()
	feed, err := fp.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("rss parse: %w", err)
	}

	now := time.Now()
	items := make([]models.Item, 0, len(feed.Items))
	for _, entry := range feed.Items {
		pubTime := now
		if entry.PublishedParsed != nil {
			pubTime = *entry.PublishedParsed
		} else if entry.UpdatedParsed != nil {
			pubTime = *entry.UpdatedParsed
		}

		itemID := fmt.Sprintf("rss-%d-%s", source.ID, hashString(entry.Link))

		metadata, _ := json.Marshal(map[string]interface{}{
			"categories": entry.Categories,
		})

		author := ""
		if entry.Author != nil {
			author = entry.Author.Name
		}

		items = append(items, models.Item{
			ID:          itemID,
			SourceID:    source.ID,
			SourceType:  models.SourceRSS,
			SourceName:  source.Name,
			Title:       entry.Title,
			Body:        truncate(stripHTML(entry.Description), 500),
			URL:         entry.Link,
			Author:      author,
			PublishedAt: pubTime,
			FetchedAt:   now,
			Metadata:    string(metadata),
		})
	}
	return items, nil
}

func hashString(s string) string {
	h := 0
	for _, c := range s {
		h = h*31 + int(c)
	}
	if h < 0 {
		h = -h
	}
	return fmt.Sprintf("%x", h)
}

func stripHTML(s string) string {
	inTag := false
	inComment := false
	var result []byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		if inComment {
			if c == '-' && i+2 < len(s) && s[i+1] == '-' && s[i+2] == '>' {
				inComment = false
				i += 2
			}
			continue
		}
		if c == '<' {
			if i+3 < len(s) && s[i:i+4] == "<!--" {
				inComment = true
				i += 3
				continue
			}
			inTag = true
			continue
		}
		if inTag && c == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result = append(result, c)
		}
	}
	clean := string(result)
	clean = strings.Join(strings.Fields(clean), " ")
	return clean
}

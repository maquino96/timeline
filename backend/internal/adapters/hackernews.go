package adapters

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/maquino96/timeline/internal/models"
)

const hnBaseURL = "https://hacker-news.firebaseio.com/v0"

type HackerNewsAdapter struct {
	client *http.Client
}

func NewHackerNewsAdapter() *HackerNewsAdapter {
	return &HackerNewsAdapter{
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

type hnItem struct {
	ID          int64  `json:"id"`
	Type        string `json:"type"`
	Title       string `json:"title"`
	Text        string `json:"text"`
	URL         string `json:"url"`
	By          string `json:"by"`
	Time        int64  `json:"time"`
	Score       int    `json:"score"`
	Descendants int    `json:"descendants"`
}

func (a *HackerNewsAdapter) Fetch(source models.Source) ([]models.Item, error) {
	listType := source.URL
	if listType == "" || listType == "top" {
		listType = "topstories"
	}

	idsURL := fmt.Sprintf("%s/%s.json", hnBaseURL, listType)
	storyIDs, err := a.fetchIDs(idsURL)
	if err != nil {
		return nil, err
	}

	if len(storyIDs) > 50 {
		storyIDs = storyIDs[:50]
	}

	items := make([]models.Item, 0, len(storyIDs))
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, 10)

	now := time.Now()
	for _, id := range storyIDs {
		wg.Add(1)
		go func(id int64) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			item, err := a.fetchItem(id)
			if err != nil {
				return
			}
			if item == nil || item.Type != "story" || item.Title == "" {
				return
			}

			itemID := fmt.Sprintf("hn-%d", item.ID)
			pubTime := time.Unix(item.Time, 0)
			storyURL := item.URL
			if storyURL == "" {
				storyURL = fmt.Sprintf("https://news.ycombinator.com/item?id=%d", item.ID)
			}

			threadURL := fmt.Sprintf("https://news.ycombinator.com/item?id=%d", item.ID)
			metadata, _ := json.Marshal(map[string]interface{}{
				"score":       item.Score,
				"descendants": item.Descendants,
				"thread_url":  threadURL,
			})

			mu.Lock()
			items = append(items, models.Item{
				ID:          itemID,
				SourceID:    source.ID,
				SourceType:  models.SourceHackerNews,
				SourceName:  source.Name,
				Title:       item.Title,
				Body:        truncate(item.Text, 500),
				URL:         storyURL,
				Author:      item.By,
				PublishedAt: pubTime,
				FetchedAt:   now,
				Metadata:    string(metadata),
			})
			mu.Unlock()
		}(id)
	}
	wg.Wait()

	return items, nil
}

func (a *HackerNewsAdapter) fetchIDs(url string) ([]int64, error) {
	resp, err := a.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("hn fetch ids: %w", err)
	}
	defer resp.Body.Close()

	var ids []int64
	if err := json.NewDecoder(resp.Body).Decode(&ids); err != nil {
		return nil, fmt.Errorf("hn decode ids: %w", err)
	}
	return ids, nil
}

func (a *HackerNewsAdapter) fetchItem(id int64) (*hnItem, error) {
	url := fmt.Sprintf("%s/item/%d.json", hnBaseURL, id)
	resp, err := a.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var item hnItem
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, err
	}
	return &item, nil
}

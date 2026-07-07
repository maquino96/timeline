package adapters

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/matthewaquino/timeline/internal/models"
)

type RedditAdapter struct {
	client      *http.Client
	accessToken string
	tokenExpiry time.Time
	mu          sync.Mutex
}

func NewRedditAdapter() *RedditAdapter {
	return &RedditAdapter{
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

type redditResponse struct {
	Data struct {
		Children []struct {
			Data redditPost `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

type redditPost struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Selftext    string  `json:"selftext"`
	URL         string  `json:"url"`
	Author      string  `json:"author"`
	Created     float64 `json:"created_utc"`
	Permalink   string  `json:"permalink"`
	Subreddit   string  `json:"subreddit"`
	Score       int     `json:"score"`
	NumComments int     `json:"num_comments"`
}

func (a *RedditAdapter) ensureAuth() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.accessToken != "" && time.Now().Before(a.tokenExpiry) {
		return nil
	}

	clientID := os.Getenv("REDDIT_CLIENT_ID")
	clientSecret := os.Getenv("REDDIT_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		return fmt.Errorf("REDDIT_CLIENT_ID and REDDIT_CLIENT_SECRET env vars required for Reddit API")
	}

	req, err := http.NewRequest("POST", "https://www.reddit.com/api/v1/access_token",
		strings.NewReader("grant_type=https://oauth.reddit.com/grants/installed_client"),
	)
	if err != nil {
		return err
	}
	req.SetBasicAuth(clientID, clientSecret)
	req.Header.Set("User-Agent", "timeline-aggregator/1.0")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("reddit auth: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		Error       string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("reddit auth decode: %w", err)
	}
	if result.Error != "" {
		return fmt.Errorf("reddit auth error: %s", result.Error)
	}

	a.accessToken = result.AccessToken
	a.tokenExpiry = time.Now().Add(time.Duration(result.ExpiresIn) * time.Second).Add(-60 * time.Second)
	return nil
}

func (a *RedditAdapter) Fetch(source models.Source) ([]models.Item, error) {
	if err := a.ensureAuth(); err != nil {
		return nil, err
	}

	subreddit := extractSubreddit(source.URL)
	apiURL := fmt.Sprintf("https://oauth.reddit.com/r/%s/new?limit=25", subreddit)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "timeline-aggregator/1.0")
	req.Header.Set("Authorization", "Bearer "+a.accessToken)

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("reddit fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		a.mu.Lock()
		a.accessToken = ""
		a.mu.Unlock()
		return nil, fmt.Errorf("reddit auth expired, will retry")
	}
	if resp.StatusCode != http.StatusOK {
		retryAfter := resp.Header.Get("Retry-After")
		if retryAfter != "" {
			return nil, fmt.Errorf("reddit returned status %d (retry-after: %ss)", resp.StatusCode, retryAfter)
		}
		return nil, fmt.Errorf("reddit returned status %d", resp.StatusCode)
	}

	var data redditResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("reddit decode: %w", err)
	}

	now := time.Now()
	items := make([]models.Item, 0, len(data.Data.Children))
	for _, child := range data.Data.Children {
		post := child.Data
		itemID := fmt.Sprintf("reddit-%s", post.ID)
		pubTime := time.Unix(int64(post.Created), 0)
		permalink := "https://www.reddit.com" + post.Permalink

		metadata, _ := json.Marshal(map[string]interface{}{
			"subreddit":    post.Subreddit,
			"score":        post.Score,
			"num_comments": post.NumComments,
		})

		items = append(items, models.Item{
			ID:          itemID,
			SourceID:    source.ID,
			SourceType:  models.SourceReddit,
			SourceName:  source.Name,
			Title:       post.Title,
			Body:        truncate(post.Selftext, 500),
			URL:         permalink,
			Author:      post.Author,
			PublishedAt: pubTime,
			FetchedAt:   now,
			Metadata:    string(metadata),
		})
	}
	return items, nil
}

func extractSubreddit(url string) string {
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "www.reddit.com")
	url = strings.TrimPrefix(url, "reddit.com")
	url = strings.TrimPrefix(url, "/r/")
	url = strings.TrimPrefix(url, "r/")
	url = strings.TrimSuffix(url, "/")
	return strings.TrimSpace(url)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

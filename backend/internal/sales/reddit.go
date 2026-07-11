package sales

import (
	"encoding/xml"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"
)

type redditAtom struct {
	Entries []struct {
		Title string `xml:"title"`
		Link  struct {
			Href string `xml:"href,attr"`
		} `xml:"link"`
		Content string `xml:"content"`
	} `xml:"entry"`
}

var reHref = regexp.MustCompile(`(?is)<a\s+href="([^"]+)"[^>]*>(.*?)</a>`)

// FetchReddit searches r/buildapcsales for the term and returns the lowest
// priced deal at or above the floor.
func FetchReddit(searchTerm string, floor float64) (float64, string, bool) {
	req, err := http.NewRequest("GET", "https://www.reddit.com/r/buildapcsales/search.rss", nil)
	if err != nil {
		return 0, "", false
	}
	q := req.URL.Query()
	q.Set("q", searchTerm)
	q.Set("sort", "new")
	q.Set("restrict_sr", "on")
	req.URL.RawQuery = q.Encode()
	req.Header.Set("User-Agent", "timeline-sale-watcher/1.0")

	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, "", false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, "", false
	}

	var feed redditAtom
	if err := xml.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return 0, "", false
	}

	type cand struct {
		price float64
		url   string
	}
	var candidates []cand
	for _, e := range feed.Entries {
		price, ok := extractPrice(e.Title)
		if !ok || price < floor {
			continue
		}
		dealURL := extractRedditDealURL(e.Content)
		if dealURL == "" {
			dealURL = e.Link.Href
		}
		candidates = append(candidates, cand{price, dealURL})
	}
	if len(candidates) == 0 {
		return 0, "", false
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].price < candidates[j].price })
	return candidates[0].price, candidates[0].url, true
}

func extractRedditDealURL(html string) string {
	for _, m := range reHref.FindAllStringSubmatch(html, -1) {
		href, text := m[1], m[2]
		if strings.Contains(text, "[link]") && href != "" && !strings.HasPrefix(href, "https://www.reddit.com") {
			return href
		}
	}
	// fall back to any [link] anchor even if reddit-hosted
	for _, m := range reHref.FindAllStringSubmatch(html, -1) {
		href, text := m[1], m[2]
		if strings.Contains(text, "[link]") && href != "" {
			return href
		}
	}
	return ""
}

var httpClient = &http.Client{Timeout: 15 * time.Second}

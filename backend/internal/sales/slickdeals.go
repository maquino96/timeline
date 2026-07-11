package sales

import (
	"encoding/xml"
	"net/http"
	"sort"
)

type slickRSS struct {
	Channel struct {
		Items []struct {
			Title       string `xml:"title"`
			Link        string `xml:"link"`
			Description string `xml:"description"`
		} `xml:"item"`
	} `xml:"channel"`
}

// FetchSlickdeals searches the Slickdeals RSS feed and returns the lowest priced
// deal at or above the floor.
func FetchSlickdeals(searchTerm string, floor float64) (float64, string, bool) {
	req, err := http.NewRequest("GET", "https://slickdeals.net/newsearch.php", nil)
	if err != nil {
		return 0, "", false
	}
	q := req.URL.Query()
	q.Set("rss", "1")
	q.Set("q", searchTerm)
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

	var feed slickRSS
	if err := xml.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return 0, "", false
	}

	type cand struct {
		price float64
		url   string
	}
	var candidates []cand
	for _, item := range feed.Channel.Items {
		best, ok := bestOf(item.Title, item.Description)
		if !ok || best < floor {
			continue
		}
		candidates = append(candidates, cand{best, item.Link})
	}
	if len(candidates) == 0 {
		return 0, "", false
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].price < candidates[j].price })
	return candidates[0].price, candidates[0].url, true
}

func bestOf(title, description string) (float64, bool) {
	tp, tok := extractPrice(title)
	dp, dok := extractPrice(description)
	switch {
	case tok && dok:
		if tp <= dp {
			return tp, true
		}
		return dp, true
	case tok:
		return tp, true
	case dok:
		return dp, true
	}
	return 0, false
}

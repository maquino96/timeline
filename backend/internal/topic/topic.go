package topic

import (
	"strings"

	"github.com/matthewaquino/timeline/internal/models"
)

type Engine struct{}

func New() *Engine {
	return &Engine{}
}

func (e *Engine) Match(item *models.Item, topics []models.Topic) []int64 {
	if len(topics) == 0 {
		return nil
	}

	searchText := strings.ToLower(item.Title + " " + item.Body)
	var matches []int64

	for _, topic := range topics {
		if !topic.Enabled {
			continue
		}
		if topicMatches(searchText, topic.Keywords) {
			matches = append(matches, topic.ID)
		}
	}

	return matches
}

func topicMatches(searchText, keywords string) bool {
	for _, kw := range strings.Split(keywords, ",") {
		kw = strings.TrimSpace(kw)
		if kw == "" {
			continue
		}
		if strings.Contains(searchText, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}

package adapters

import (
	"fmt"

	"github.com/matthewaquino/timeline/internal/models"
)

type Registry struct {
	adapters map[models.SourceType]Adapter
}

func NewRegistry() *Registry {
	return &Registry{
		adapters: map[models.SourceType]Adapter{
			models.SourceReddit:     NewRedditAdapter(),
			models.SourceHackerNews: NewHackerNewsAdapter(),
			models.SourceRSS:        NewRSSAdapter(),
			models.SourceSECEDGAR:   NewSECEDGARAdapter(),
		},
	}
}

func (r *Registry) Fetch(source models.Source) ([]models.Item, error) {
	adapter, ok := r.adapters[source.Type]
	if !ok {
		return nil, fmt.Errorf("no adapter for source type %q", source.Type)
	}
	return adapter.Fetch(source)
}

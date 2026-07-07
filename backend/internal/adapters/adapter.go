package adapters

import "github.com/matthewaquino/timeline/internal/models"

type Adapter interface {
	Fetch(source models.Source) ([]models.Item, error)
}

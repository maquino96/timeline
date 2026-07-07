package adapters

import "github.com/maquino96/timeline/internal/models"

type Adapter interface {
	Fetch(source models.Source) ([]models.Item, error)
}

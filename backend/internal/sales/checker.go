package sales

import (
	"fmt"
	"log"

	"github.com/maquino96/timeline/internal/models"
	"github.com/maquino96/timeline/internal/store"
)

var sourceNames = map[string]string{
	"reddit":     "Reddit (r/buildapcsales)",
	"ebay":       "eBay",
	"slickdeals": "Slickdeals",
}

type fetcher func(searchTerm string, floor float64) (float64, string, bool)

// CheckAll runs a price check across all active watched items, stores the
// latest per-source prices, and emits an alert (in-app + email) when the best
// price at or above the item's floor drops below its threshold and no alert
// has fired for that item in the last 24h.
func CheckAll(s *store.Store) {
	items, err := s.GetWatchItems(true)
	if err != nil {
		log.Printf("sales check: get items: %v", err)
		return
	}
	if len(items) == 0 {
		return
	}

	handlers := map[string]fetcher{
		"reddit":     FetchReddit,
		"ebay":       FetchEbay,
		"slickdeals": FetchSlickdeals,
	}

	for _, item := range items {
		type result struct {
			price float64
			url   string
		}
		results := map[string]result{}

		for key, fetch := range handlers {
			price, dealURL, ok := fetch(item.SearchTerm, item.Floor)
			if !ok {
				continue
			}
			results[key] = result{price, dealURL}
			if err := s.UpdateWatchItemPrice(item.ID, key, price); err != nil {
				log.Printf("sales check: update price %s/%s: %v", item.Name, key, err)
			}
		}

		if len(results) == 0 {
			continue
		}

		bestSource := ""
		var best result
		for key, r := range results {
			if bestSource == "" || r.price < best.price {
				bestSource = key
				best = r
			}
		}

		if best.price >= item.Threshold {
			continue
		}

		recent, err := s.HasRecentSaleAlert(item.ID, 24)
		if err != nil {
			log.Printf("sales check: recent alert %s: %v", item.Name, err)
			continue
		}
		if recent {
			continue
		}

		title := fmt.Sprintf("%s — $%.2f on %s", item.Name, best.price, sourceNames[bestSource])
		sent := SendPriceAlert(item.Name, best.price, item.Threshold, title, best.url, sourceNames[bestSource])
		alert := &models.SaleAlert{
			ItemID:  item.ID,
			Price:   best.price,
			Title:   title,
			DealURL: best.url,
			Source:  bestSource,
			Sent:    sent,
		}
		if err := s.AddSaleAlert(alert); err != nil {
			log.Printf("sales check: add alert %s: %v", item.Name, err)
			continue
		}
		log.Printf("sales check: alert for %s ($%.2f on %s, emailed=%v)", item.Name, best.price, bestSource, sent)
	}
}

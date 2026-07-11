package sales

import (
	"encoding/json"
	"net/http"
	"os"
	"sort"
	"strconv"
)

const ebayEndpoint = "https://svcs.ebay.com/services/search/FindingService/v1"

// FetchEbay searches eBay fixed-price listings and returns the lowest priced
// listing at or above the floor. Returns false when EBAY_APP_ID is unset.
func FetchEbay(searchTerm string, floor float64) (float64, string, bool) {
	appID := os.Getenv("EBAY_APP_ID")
	if appID == "" {
		return 0, "", false
	}

	req, err := http.NewRequest("GET", ebayEndpoint, nil)
	if err != nil {
		return 0, "", false
	}
	q := req.URL.Query()
	q.Set("OPERATION-NAME", "findItemsByKeywords")
	q.Set("SERVICE-VERSION", "1.0.0")
	q.Set("SECURITY-APPNAME", appID)
	q.Set("RESPONSE-DATA-FORMAT", "JSON")
	q.Set("keywords", searchTerm)
	q.Set("itemFilter(0).name", "ListingType")
	q.Set("itemFilter(0).value(0)", "FixedPrice")
	q.Set("paginationInput.entriesPerPage", "10")
	q.Set("sortOrder", "PricePlusShippingLowest")
	req.URL.RawQuery = q.Encode()

	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, "", false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, "", false
	}

	var data struct {
		Resp []struct {
			SearchResult []struct {
				Item []struct {
					ViewItemURL   []string `json:"viewItemURL"`
					SellingStatus []struct {
						CurrentPrice []struct {
							Value string `json:"__value__"`
						} `json:"currentPrice"`
					} `json:"sellingStatus"`
				} `json:"item"`
			} `json:"searchResult"`
		} `json:"findItemsByKeywordsResponse"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, "", false
	}
	if len(data.Resp) == 0 || len(data.Resp[0].SearchResult) == 0 {
		return 0, "", false
	}

	type cand struct {
		price float64
		url   string
	}
	var candidates []cand
	for _, item := range data.Resp[0].SearchResult[0].Item {
		if len(item.SellingStatus) == 0 || len(item.SellingStatus[0].CurrentPrice) == 0 {
			continue
		}
		price, err := strconv.ParseFloat(item.SellingStatus[0].CurrentPrice[0].Value, 64)
		if err != nil || price < floor {
			continue
		}
		url := ""
		if len(item.ViewItemURL) > 0 {
			url = item.ViewItemURL[0]
		}
		candidates = append(candidates, cand{price, url})
	}
	if len(candidates) == 0 {
		return 0, "", false
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].price < candidates[j].price })
	return candidates[0].price, candidates[0].url, true
}

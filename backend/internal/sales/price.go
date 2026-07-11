package sales

import (
	"regexp"
	"strconv"
)

var (
	reAmount   = regexp.MustCompile(`\$(\d+(?:\.\d{1,2})?)`)
	reShipping = regexp.MustCompile(`(?i)\+\s*\$(\d+(?:\.\d{1,2})?)\s*(?:shipping|s/h|delivery)`)
	reAfter    = regexp.MustCompile(`(?i)(?:after|save)\s+\$(\d+(?:\.\d{1,2})?)`)
	reOff      = regexp.MustCompile(`(?i)\$(\d+(?:\.\d{1,2})?)\s+off`)
)

func extractAmounts(text string) []float64 {
	var out []float64
	for _, m := range reAmount.FindAllStringSubmatch(text, -1) {
		if v, err := strconv.ParseFloat(m[1], 64); err == nil && v >= 1.0 {
			out = append(out, v)
		}
	}
	return out
}

func extractShipping(text string) (float64, bool) {
	m := reShipping.FindStringSubmatch(text)
	if m == nil {
		return 0, false
	}
	v, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

func extractDiscounts(text string) map[float64]bool {
	out := map[float64]bool{}
	for _, m := range reAfter.FindAllStringSubmatch(text, -1) {
		if v, err := strconv.ParseFloat(m[1], 64); err == nil {
			out[v] = true
		}
	}
	for _, m := range reOff.FindAllStringSubmatch(text, -1) {
		if v, err := strconv.ParseFloat(m[1], 64); err == nil {
			out[v] = true
		}
	}
	return out
}

// extractPrice returns the best (lowest sale price + shipping) found in text,
// mirroring the original Python price-parsing logic. ok is false when no
// plausible price is present.
func extractPrice(text string) (price float64, ok bool) {
	amounts := extractAmounts(text)
	if len(amounts) == 0 {
		return 0, false
	}

	shipping, hasShipping := extractShipping(text)
	excluded := extractDiscounts(text)
	if hasShipping {
		excluded[shipping] = true
	}

	best := 0.0
	found := false
	for _, a := range amounts {
		if excluded[a] {
			continue
		}
		if !found || a < best {
			best = a
			found = true
		}
	}
	if !found {
		return 0, false
	}

	if hasShipping {
		best += shipping
	}
	return best, true
}

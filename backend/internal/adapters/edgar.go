package adapters

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/matthewaquino/timeline/internal/models"
)

const secBaseURL = "https://www.sec.gov"
const fidelityQFRBase = "https://fundresearch.fidelity.com/mutual-funds/analysis/%s?documentType=QFR"

var cusipRe = regexp.MustCompile(`\((\d{9})\)`)

var secFormTypes = []string{"NPORT-P", "N-CSR", "N-CSRS"}

var formLabels = map[string]string{
	"NPORT-P": "Quarterly Holdings",
	"N-CSR":   "Annual Report",
	"N-CSRS":  "Semi-Annual Report",
}

type SECEDGARAdapter struct {
	client *http.Client
}

func NewSECEDGARAdapter() *SECEDGARAdapter {
	return &SECEDGARAdapter{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

const atomNS = "http://www.w3.org/2005/Atom"

type secFilingEntry struct {
	AccessionNum string
	FilingDate   string
	FilingType   string
	FormName     string
	FilingHref   string
	Size         string
}

func (a *SECEDGARAdapter) Fetch(source models.Source) ([]models.Item, error) {
	cik := strings.TrimSpace(source.URL)
	if cik == "" {
		return nil, fmt.Errorf("secedgar: CIK required in URL field for source %q", source.Name)
	}

	cusip, fundName := parseSECSourceName(source.Name)

	now := time.Now()
	var items []models.Item

	for _, formType := range secFormTypes {
		entries, err := a.fetchEntries(cik, formType)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if len(items) >= 4 {
				break
			}

			if !a.filingMatches(entry.FilingHref, fundName, cusip) {
				continue
			}

			filingDate, err := time.Parse("2006-01-02", entry.FilingDate)
			if err != nil {
				filingDate = now
			}

			label := formLabels[entry.FilingType]
			if label == "" {
				label = entry.FilingType
			}

			itemURL := entry.FilingHref
			if cusip != "" {
				itemURL = fmt.Sprintf(fidelityQFRBase, cusip)
			}

			item := models.Item{
				ID:          fmt.Sprintf("edgar-%s", entry.AccessionNum),
				SourceID:    source.ID,
				SourceType:  models.SourceSECEDGAR,
				SourceName:  source.Name,
				Title:       fmt.Sprintf("%s Filed: %s", label, entry.FilingDate),
				Body:        entry.FormName,
				URL:         itemURL,
				Author:      "SEC EDGAR",
				PublishedAt: filingDate,
				FetchedAt:   now,
				Metadata:    fmt.Sprintf(`{"form_type":"%s","accession_number":"%s","size":"%s","filing_label":"%s"}`, entry.FilingType, entry.AccessionNum, entry.Size, label),
			}

			items = append(items, item)
		}

		if len(items) >= 4 {
			break
		}
	}

	return items, nil
}

func parseSECSourceName(name string) (cusip string, fundName string) {
	matches := cusipRe.FindStringSubmatch(name)
	if len(matches) == 2 {
		cusip = matches[1]
		name = strings.TrimSpace(strings.Split(name, "(")[0])
	}
	if idx := strings.Index(name, " - "); idx != -1 {
		fundName = strings.TrimSpace(name[idx+3:])
	} else {
		fundName = strings.TrimSpace(name)
	}
	return
}

func (a *SECEDGARAdapter) fetchEntries(cik, formType string) ([]secFilingEntry, error) {
	url := fmt.Sprintf("%s/cgi-bin/browse-edgar?action=getcompany&CIK=%s&type=%s&dateb=&owner=include&count=10&output=atom", secBaseURL, cik, formType)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "timeline-aggregator/1.0 (matthewaquino96@gmail.com)")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("secedgar fetch %s: %w", formType, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("secedgar fetch %s: status %d", formType, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("secedgar read %s: %w", formType, err)
	}

	entries, err := parseSECAtomFeed(body)
	if err != nil {
		return nil, fmt.Errorf("secedgar parse %s: %w", formType, err)
	}

	return entries, nil
}

var (
	secAccessionRe = regexp.MustCompile(`<accession-number>([^<]+)</accession-number>`)
	secFilingDateRe = regexp.MustCompile(`<filing-date>([^<]+)</filing-date>`)
	secFilingTypeRe = regexp.MustCompile(`<filing-type>([^<]+)</filing-type>`)
	secFormNameRe  = regexp.MustCompile(`<form-name>([^<]+)</form-name>`)
	secFilingHrefRe = regexp.MustCompile(`<filing-href>([^<]+)</filing-href>`)
	secSizeRe      = regexp.MustCompile(`<size>([^<]+)</size>`)
	secEntryRe     = regexp.MustCompile(`(?s)<entry>(.*?)</entry>`)
)

func parseSECAtomFeed(data []byte) ([]secFilingEntry, error) {
	content := string(data)
	matches := secEntryRe.FindAllStringSubmatch(content, -1)

	var entries []secFilingEntry
	for _, match := range matches {
		entryXML := match[1]

		filingType := firstMatch(secFilingTypeRe, entryXML)
		filingHref := firstMatch(secFilingHrefRe, entryXML)
		if filingType == "" || filingHref == "" {
			continue
		}

		entries = append(entries, secFilingEntry{
			AccessionNum: firstMatch(secAccessionRe, entryXML),
			FilingDate:   firstMatch(secFilingDateRe, entryXML),
			FilingType:   filingType,
			FormName:     firstMatch(secFormNameRe, entryXML),
			FilingHref:   filingHref,
			Size:         firstMatch(secSizeRe, entryXML),
		})
	}

	return entries, nil
}

func firstMatch(re *regexp.Regexp, s string) string {
	m := re.FindStringSubmatch(s)
	if len(m) >= 2 {
		return m[1]
	}
	return ""
}

func (a *SECEDGARAdapter) filingMatches(filingHref, fundName, cusip string) bool {
	req, err := http.NewRequest("GET", filingHref, nil)
	if err != nil {
		return false
	}
	req.Header.Set("User-Agent", "timeline-aggregator/1.0 (matthewaquino96@gmail.com)")

	resp, err := a.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false
	}

	html := string(body)

	if cusip != "" && strings.Contains(html, cusip) {
		return true
	}

	if fundName != "" && strings.Contains(html, fundName) {
		return true
	}

	return false
}

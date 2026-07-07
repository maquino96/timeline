package adapters

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"time"
)

type Feed struct {
	Items []FeedItem
}

type FeedItem struct {
	Title           string
	Description     string
	Link            string
	Author          *FeedAuthor
	Categories      []string
	PublishedParsed *time.Time
	UpdatedParsed   *time.Time
}

type FeedAuthor struct {
	Name string
}

type FeedParser struct{}

func NewFeedParser() *FeedParser {
	return &FeedParser{}
}

func (p *FeedParser) Parse(r io.Reader) (*Feed, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	content := string(data)

	if strings.Contains(content, "<rss") || strings.Contains(content, "<rdf:RDF") {
		return p.parseRSS(data)
	}
	if strings.Contains(content, "<feed") {
		return p.parseAtom(data)
	}
	return p.parseRSS(data)
}

type rssFeed struct {
	XMLName xml.Name  `xml:"rss"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Items []rssItem `xml:"item"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Link        string `xml:"link"`
	Author      string `xml:"author"`
	PubDate     string `xml:"pubDate"`
	Category    []string `xml:"category"`
	Creator     string `xml:"http://purl.org/dc/elements/1.1/ creator"`
}

type rdfFeed struct {
	XMLName xml.Name   `xml:"RDF"`
	Items   []rdfItem  `xml:"item"`
}

type rdfItem struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Link        string `xml:"link"`
	Creator     string `xml:"http://purl.org/dc/elements/1.1/ creator"`
	Date        string `xml:"http://purl.org/dc/elements/1.1/ date"`
}

type atomFeed struct {
	XMLName xml.Name    `xml:"feed"`
	Entries []atomEntry `xml:"entry"`
}

type atomEntry struct {
	Title     string `xml:"title"`
	Summary   string `xml:"summary"`
	Content   string `xml:"content"`
	Link      atomLink  `xml:"link"`
	Author    atomAuthor `xml:"author"`
	Published string `xml:"published"`
	Updated   string `xml:"updated"`
	Category  []atomCategory `xml:"category"`
}

type atomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
}

type atomAuthor struct {
	Name string `xml:"name"`
}

type atomCategory struct {
	Term string `xml:"term,attr"`
}

func (p *FeedParser) parseRSS(data []byte) (*Feed, error) {
	var feed Feed

	if strings.Contains(string(data), "<rss") {
		var rss rssFeed
		if err := xml.Unmarshal(data, &rss); err != nil {
			return nil, err
		}
		for _, item := range rss.Channel.Items {
			fi := FeedItem{
				Title:       item.Title,
				Description: item.Description,
				Link:        item.Link,
				Categories:  item.Category,
			}
			author := item.Author
			if author == "" {
				author = item.Creator
			}
			if author != "" {
				fi.Author = &FeedAuthor{Name: author}
			}
			if t, err := parseTime(item.PubDate); err == nil {
				fi.PublishedParsed = &t
			}
			feed.Items = append(feed.Items, fi)
		}
		return &feed, nil
	}

	var rdf rdfFeed
	if err := xml.Unmarshal(data, &rdf); err != nil {
		return nil, err
	}
	for _, item := range rdf.Items {
		fi := FeedItem{
			Title:       item.Title,
			Description: item.Description,
			Link:        item.Link,
		}
		if item.Creator != "" {
			fi.Author = &FeedAuthor{Name: item.Creator}
		}
		if t, err := parseTime(item.Date); err == nil {
			fi.PublishedParsed = &t
		}
		feed.Items = append(feed.Items, fi)
	}
	return &feed, nil
}

func (p *FeedParser) parseAtom(data []byte) (*Feed, error) {
	var af atomFeed
	if err := xml.Unmarshal(data, &af); err != nil {
		return nil, err
	}

	var feed Feed
	for _, entry := range af.Entries {
		fi := FeedItem{
			Title: entry.Title,
		}
		if entry.Summary != "" {
			fi.Description = entry.Summary
		} else {
			fi.Description = entry.Content
		}

		if entry.Link.Href != "" {
			fi.Link = entry.Link.Href
		} else {
			for _, l := range []atomLink{entry.Link} {
				if l.Href != "" {
					fi.Link = l.Href
					break
				}
			}
		}

		if entry.Author.Name != "" {
			fi.Author = &FeedAuthor{Name: entry.Author.Name}
		}

		for _, cat := range entry.Category {
			fi.Categories = append(fi.Categories, cat.Term)
		}

		if t, err := parseTime(entry.Published); err == nil {
			fi.PublishedParsed = &t
		}
		if t, err := parseTime(entry.Updated); err == nil {
			fi.UpdatedParsed = &t
		}

		feed.Items = append(feed.Items, fi)
	}
	return &feed, nil
}

func parseTime(s string) (time.Time, error) {
	formats := []string{
		time.RFC1123Z, time.RFC1123,
		time.RFC3339, time.RFC3339Nano,
		"2006-01-02T15:04:05-07:00",
		"2006-01-02T15:04:05Z",
		"2006-01-02",
		"Mon, 02 Jan 2006 15:04:05 MST",
		"Mon, 02 Jan 2006 15:04:05 -0700",
		"Mon, 2 Jan 2006 15:04:05 -0700",
		"02 Jan 2006 15:04:05 -0700",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse time: %s", s)
}

package scraper

import (
	"fmt"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

const userAgent = "Mozilla/5.0 (compatible; scraper-mcp-server/1.0)"

// ScrapeStatic scrapes a single URL using Colly and returns matched values.
func ScrapeStatic(url, selector, attribute string) ([]string, error) {
	if attribute == "" {
		attribute = "text"
	}

	var results []string
	var scrapeErr error

	c := colly.NewCollector(
		colly.UserAgent(userAgent),
	)
	c.SetRequestTimeout(30 * time.Second)

	c.OnHTML(selector, func(e *colly.HTMLElement) {
		var val string
		if attribute == "text" {
			val = strings.TrimSpace(e.Text)
		} else {
			val = strings.TrimSpace(e.Attr(attribute))
		}
		if val != "" {
			results = append(results, val)
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		scrapeErr = fmt.Errorf("request to %s failed (status %d): %w", r.Request.URL, r.StatusCode, err)
	})

	if err := c.Visit(url); err != nil {
		return nil, fmt.Errorf("failed to visit %s: %w", url, err)
	}

	if scrapeErr != nil {
		return nil, scrapeErr
	}

	return results, nil
}

// ScrapeMultiple scrapes multiple URLs with the same selector using parallel Colly workers.
func ScrapeMultiple(urls []string, selector, attribute string) (map[string][]string, error) {
	if attribute == "" {
		attribute = "text"
	}

	results := make(map[string][]string, len(urls))
	errs := make(map[string]string, len(urls))

	c := colly.NewCollector(
		colly.UserAgent(userAgent),
		colly.Async(true),
	)
	c.SetRequestTimeout(30 * time.Second)
	if err := c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 5,
	}); err != nil {
		return nil, fmt.Errorf("failed to set parallelism: %w", err)
	}

	c.OnHTML(selector, func(e *colly.HTMLElement) {
		reqURL := e.Request.URL.String()
		var val string
		if attribute == "text" {
			val = strings.TrimSpace(e.Text)
		} else {
			val = strings.TrimSpace(e.Attr(attribute))
		}
		if val != "" {
			results[reqURL] = append(results[reqURL], val)
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		errs[r.Request.URL.String()] = fmt.Sprintf("status %d: %s", r.StatusCode, err.Error())
	})

	for _, u := range urls {
		if err := c.Visit(u); err != nil {
			errs[u] = err.Error()
		}
	}
	c.Wait()

	// Ensure every requested URL has an entry (empty slice if no matches).
	for _, u := range urls {
		if _, ok := results[u]; !ok {
			if msg, failed := errs[u]; failed {
				results[u] = []string{"error: " + msg}
			} else {
				results[u] = []string{}
			}
		}
	}

	return results, nil
}

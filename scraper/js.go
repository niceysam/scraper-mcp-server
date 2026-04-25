package scraper

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

// ScrapeJS scrapes a JavaScript-rendered page using headless Chrome.
func ScrapeJS(url, selector, attribute, waitFor string, timeoutSeconds int) ([]string, error) {
	if attribute == "" {
		attribute = "text"
	}
	if timeoutSeconds <= 0 {
		timeoutSeconds = 30
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", "new"),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.UserAgent(userAgent),
	)

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelAlloc()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancelTimeout := context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
	defer cancelTimeout()

	// Navigate to the page and optionally wait for a specific element.
	tasks := chromedp.Tasks{
		chromedp.Navigate(url),
	}
	if waitFor != "" {
		tasks = append(tasks, chromedp.WaitVisible(waitFor, chromedp.ByQuery))
	} else {
		tasks = append(tasks, chromedp.WaitReady("body", chromedp.ByQuery))
	}

	if err := chromedp.Run(ctx, tasks); err != nil {
		return nil, fmt.Errorf("chromedp navigation failed: %w", err)
	}

	// Extract values via JS evaluation to handle all attribute types uniformly.
	var jsExpr string
	if attribute == "text" {
		jsExpr = fmt.Sprintf(`
			(function() {
				var els = document.querySelectorAll(%q);
				var out = [];
				for (var i = 0; i < els.length; i++) {
					var t = (els[i].innerText || els[i].textContent || "").trim();
					if (t) out.push(t);
				}
				return out;
			})()
		`, selector)
	} else {
		jsExpr = fmt.Sprintf(`
			(function() {
				var els = document.querySelectorAll(%q);
				var out = [];
				for (var i = 0; i < els.length; i++) {
					var v = (els[i].getAttribute(%q) || "").trim();
					if (v) out.push(v);
				}
				return out;
			})()
		`, selector, attribute)
	}

	var raw []interface{}
	if err := chromedp.Run(ctx, chromedp.Evaluate(jsExpr, &raw)); err != nil {
		return nil, fmt.Errorf("chromedp evaluate failed: %w", err)
	}

	results := make([]string, 0, len(raw))
	for _, v := range raw {
		if s, ok := v.(string); ok {
			s = strings.TrimSpace(s)
			if s != "" {
				results = append(results, s)
			}
		}
	}

	return results, nil
}

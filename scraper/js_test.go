package scraper

import (
	"strings"
	"testing"
)

func TestScrapeJS_InvalidURL(t *testing.T) {
	_, err := ScrapeJS("http://127.0.0.1:1/nope", "body", "", "", 5)
	if err == nil {
		t.Fatal("expected error for unreachable URL, got nil")
	}
	if !strings.Contains(err.Error(), "chromedp") {
		t.Errorf("expected chromedp error, got: %v", err)
	}
}

func TestScrapeJS_DefaultAttribute(t *testing.T) {
	_, err := ScrapeJS("http://127.0.0.1:1/nope", "body", "", "", 5)
	if err == nil {
		t.Fatal("expected error for unreachable URL, got nil")
	}
}

func TestScrapeJS_DefaultTimeout(t *testing.T) {
	_, err := ScrapeJS("http://127.0.0.1:1/nope", "body", "", "", 0)
	if err == nil {
		t.Fatal("expected error for unreachable URL, got nil")
	}
}

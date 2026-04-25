package scraper

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestServer(pages map[string]string) *httptest.Server {
	mux := http.NewServeMux()
	for path, body := range pages {
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(body))
		})
	}
	return httptest.NewServer(mux)
}

func TestScrapeStatic_Text(t *testing.T) {
	ts := newTestServer(map[string]string{
		"/": `<html><body>
			<h1>Title</h1>
			<p class="item">First</p>
			<p class="item">Second</p>
			<p class="item">Third</p>
		</body></html>`,
	})
	defer ts.Close()

	results, err := ScrapeStatic(ts.URL+"/", "p.item", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d: %v", len(results), results)
	}
	expected := []string{"First", "Second", "Third"}
	for i, want := range expected {
		if results[i] != want {
			t.Errorf("results[%d] = %q, want %q", i, results[i], want)
		}
	}
}

func TestScrapeStatic_Attribute(t *testing.T) {
	ts := newTestServer(map[string]string{
		"/": `<html><body>
			<a class="link" href="/page1">Page 1</a>
			<a class="link" href="/page2">Page 2</a>
		</body></html>`,
	})
	defer ts.Close()

	results, err := ScrapeStatic(ts.URL+"/", "a.link", "href")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d: %v", len(results), results)
	}
	if results[0] != "/page1" || results[1] != "/page2" {
		t.Errorf("unexpected results: %v", results)
	}
}

func TestScrapeStatic_NoMatches(t *testing.T) {
	ts := newTestServer(map[string]string{
		"/": `<html><body><p>No matching selector</p></body></html>`,
	})
	defer ts.Close()

	results, err := ScrapeStatic(ts.URL+"/", "div.nonexistent", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d: %v", len(results), results)
	}
}

func TestScrapeStatic_EmptyTextSkipped(t *testing.T) {
	ts := newTestServer(map[string]string{
		"/": `<html><body>
			<span class="v">Hello</span>
			<span class="v">   </span>
			<span class="v">World</span>
		</body></html>`,
	})
	defer ts.Close()

	results, err := ScrapeStatic(ts.URL+"/", "span.v", "text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results (whitespace-only skipped), got %d: %v", len(results), results)
	}
}

func TestScrapeStatic_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer ts.Close()

	_, err := ScrapeStatic(ts.URL, "p", "")
	if err == nil {
		t.Fatal("expected error for 500 response, got nil")
	}
}

func TestScrapeStatic_InvalidURL(t *testing.T) {
	_, err := ScrapeStatic("http://127.0.0.1:1/nope", "p", "")
	if err == nil {
		t.Fatal("expected error for unreachable URL, got nil")
	}
}

func TestScrapeMultiple_Basic(t *testing.T) {
	ts := newTestServer(map[string]string{
		"/a": `<html><body><h1>Page A</h1></body></html>`,
		"/b": `<html><body><h1>Page B</h1></body></html>`,
	})
	defer ts.Close()

	urls := []string{ts.URL + "/a", ts.URL + "/b"}
	results, err := ScrapeMultiple(urls, "h1", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 URL entries, got %d", len(results))
	}
	for _, u := range urls {
		vals, ok := results[u]
		if !ok {
			t.Errorf("missing entry for %s", u)
			continue
		}
		if len(vals) != 1 {
			t.Errorf("expected 1 result for %s, got %d: %v", u, len(vals), vals)
		}
	}
}

func TestScrapeMultiple_MissingSelector(t *testing.T) {
	ts := newTestServer(map[string]string{
		"/": `<html><body><p>Only paragraph</p></body></html>`,
	})
	defer ts.Close()

	urls := []string{ts.URL + "/"}
	results, err := ScrapeMultiple(urls, "div.missing", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	vals := results[urls[0]]
	if len(vals) != 0 {
		t.Errorf("expected empty slice for no matches, got %v", vals)
	}
}

func TestScrapeMultiple_ErrorEntry(t *testing.T) {
	urls := []string{"http://127.0.0.1:1/nope"}
	results, err := ScrapeMultiple(urls, "h1", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	vals, ok := results[urls[0]]
	if !ok {
		t.Fatal("expected entry for failed URL")
	}
	if len(vals) != 1 || vals[0][:6] != "error:" {
		t.Errorf("expected error entry, got %v", vals)
	}
}

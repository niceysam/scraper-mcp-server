# scraper-mcp-server

A lightweight MCP server for web scraping — handles both static HTML and JavaScript-rendered pages through a single, consistent interface.

---

## Why I built this

Most AI assistants can browse the web, but they often struggle with pages that require JavaScript to render content. When I tried to fetch data from sites like AWS blogs or dashboards, I kept getting back empty results — because the actual content only appears *after* JavaScript runs.

I already had Python scripts using `requests` + `BeautifulSoup` for static pages, but JS-rendered pages meant spinning up Playwright separately, writing boilerplate every time, and context-switching between tools.

So I built this MCP server to handle it all in one place. You tell it what URL and CSS selector you want — it figures out whether to use a lightweight HTTP fetch or headless Chrome, and gives you back the data.

---

## What it does

Four tools, each for a different use case:

| Tool | Engine | When to use |
|---|---|---|
| `scrape_static` | Colly (HTTP) | Fast static HTML pages |
| `scrape_js` | chromedp (headless Chrome) | JS-rendered SPAs, dashboards |
| `scrape_multiple` | Colly parallel | Same selector across many URLs |
| `scrape_crawl` | Colly recursive | Follow links to a given depth |

All tools use the same interface: give a URL and a CSS selector, get back an array of matched values.

---

## Requirements

- Go 1.21+
- Chrome or Chromium installed on the host (only required for `scrape_js`)

---

## Installation

```bash
go install github.com/niceysam/scraper-mcp-server@latest
```

Or download a pre-built binary from [Releases](https://github.com/niceysam/scraper-mcp-server/releases).

---

## MCP configuration

### Claude Desktop

Add to `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) or `%APPDATA%\Claude\claude_desktop_config.json` (Windows):

```json
{
  "mcpServers": {
    "scraper": {
      "command": "/Users/YOU/go/bin/scraper-mcp-server"
    }
  }
}
```

### OpenClaw mcporter

```json
{
  "mcpServers": {
    "scraper": {
      "command": "/path/to/scraper-mcp-server"
    }
  }
}
```

---

## Tools

### `scrape_static`

Fetches raw HTML via HTTP and extracts values using a CSS selector. No JavaScript execution — fast and lightweight.

**Parameters**

| Name | Type | Default | Description |
|---|---|---|---|
| `url` | string | required | Target URL |
| `selector` | string | required | CSS selector |
| `attribute` | string | `"text"` | `"text"` for inner text, or any attribute name (`"href"`, `"src"`, ...) |

**Example**

```json
{
  "url": "https://news.ycombinator.com",
  "selector": ".titleline > a",
  "attribute": "text"
}
```

---

### `scrape_js`

Launches headless Chrome, waits for JavaScript to execute, then extracts values. Use this for any page that loads content dynamically.

**Parameters**

| Name | Type | Default | Description |
|---|---|---|---|
| `url` | string | required | Target URL |
| `selector` | string | required | CSS selector |
| `attribute` | string | `"text"` | `"text"` or any attribute name |
| `wait_for` | string | — | CSS selector to wait for before extracting |
| `timeout_seconds` | number | `30` | Total timeout |

**Example**

```json
{
  "url": "https://aws.amazon.com/ko/blogs/tech/",
  "selector": "article",
  "timeout_seconds": 25
}
```

---

### `scrape_multiple`

Scrapes multiple URLs concurrently (5 parallel workers) with the same selector. Returns a map of URL → matched values.

**Parameters**

| Name | Type | Default | Description |
|---|---|---|---|
| `urls` | string[] | required | List of URLs |
| `selector` | string | required | CSS selector |
| `attribute` | string | `"text"` | `"text"` or any attribute name |

---

### `scrape_crawl`

Starts at a URL and recursively follows links to a specified depth, collecting matched values from every page visited.

**Parameters**

| Name | Type | Default | Description |
|---|---|---|---|
| `url` | string | required | Starting URL |
| `selector` | string | required | CSS selector |
| `attribute` | string | `"text"` | `"text"` or any attribute name |
| `depth` | number | `2` | How many link levels to follow |
| `max_pages` | number | `20` | Maximum pages to visit |
| `same_domain_only` | boolean | `true` | Restrict crawl to the same domain |
| `timeout_seconds` | number | `60` | Total timeout |

**Example** — crawl an AWS blog section two levels deep:

```json
{
  "url": "https://aws.amazon.com/ko/blogs/tech/",
  "selector": "article p",
  "depth": 2,
  "max_pages": 15
}
```

---

## Notes

- All requests send `User-Agent: Mozilla/5.0 (compatible; scraper-mcp-server/1.0)`
- `scrape_js` requires Chrome or Chromium to be available on the host
- Empty and whitespace-only matches are dropped from results
- For RSS feeds, use a dedicated RSS parser instead — this tool is for HTML pages

---

## License

MIT

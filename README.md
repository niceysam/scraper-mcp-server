# scraper-mcp-server

A Model Context Protocol (MCP) server that exposes web-scraping capabilities to any MCP-compatible AI client (Claude Desktop, OpenClaw mcporter, etc.).

Three tools are provided:

| Tool | Engine | Use when |
|---|---|---|
| `scrape_static` | Colly (HTTP) | Page content is in the initial HTML response |
| `scrape_js` | chromedp (headless Chrome) | Page requires JavaScript to render content |
| `scrape_multiple` | Colly parallel | Same selector across many URLs at once |

---

## Installation

**Prerequisites:** Go 1.21+ and (for `scrape_js`) Google Chrome or Chromium installed on the machine running the server.

```bash
go install github.com/niceysam/scraper-mcp-server@latest
```

The binary will be placed at `$(go env GOPATH)/bin/scraper-mcp-server`.

---

## MCP configuration

### Claude Desktop (`~/Library/Application Support/Claude/claude_desktop_config.json`)

```json
{
  "mcpServers": {
    "scraper": {
      "command": "/Users/YOU/go/bin/scraper-mcp-server"
    }
  }
}
```

### OpenClaw mcporter (`mcporter.json`)

```json
{
  "servers": [
    {
      "name": "scraper",
      "transport": "stdio",
      "command": ["scraper-mcp-server"]
    }
  ]
}
```

---

## Tools

### `scrape_static`

Fetches the raw HTML with Colly and returns values matched by a CSS selector.
No JavaScript is executed — fast and lightweight.

**Parameters**

| Name | Type | Required | Description |
|---|---|---|---|
| `url` | string | yes | URL to scrape |
| `selector` | string | yes | CSS selector |
| `attribute` | string | no | `"text"` (default) or any HTML attribute (`"href"`, `"src"`, `"data-id"`, …) |

**Example — extract all link hrefs from a page**

```json
{
  "tool": "scrape_static",
  "arguments": {
    "url": "https://example.com",
    "selector": "a",
    "attribute": "href"
  }
}
```

**Response**

```json
["https://www.iana.org/domains/reserved"]
```

---

### `scrape_js`

Launches headless Chrome, navigates to the URL, optionally waits for an element, then extracts values.
Use this for single-page apps or pages that load content via JavaScript.

**Parameters**

| Name | Type | Required | Description |
|---|---|---|---|
| `url` | string | yes | URL to scrape |
| `selector` | string | yes | CSS selector |
| `attribute` | string | no | `"text"` (default) or any HTML attribute |
| `wait_for` | string | no | CSS selector to wait for before extracting (ensures dynamic content is loaded) |
| `timeout_seconds` | number | no | Max wait time in seconds (default `30`) |

**Example — scrape a React-rendered product list**

```json
{
  "tool": "scrape_js",
  "arguments": {
    "url": "https://example-spa.com/products",
    "selector": ".product-title",
    "wait_for": ".product-list",
    "timeout_seconds": 15
  }
}
```

**Response**

```json
["Widget A", "Widget B", "Widget C"]
```

---

### `scrape_multiple`

Scrapes many URLs concurrently (up to 5 parallel workers) with the same selector.
Returns a map of URL → matched values.

**Parameters**

| Name | Type | Required | Description |
|---|---|---|---|
| `urls` | string[] | yes | List of URLs to scrape |
| `selector` | string | yes | CSS selector applied to every URL |
| `attribute` | string | no | `"text"` (default) or any HTML attribute |

**Example — grab the `<h1>` from several pages**

```json
{
  "tool": "scrape_multiple",
  "arguments": {
    "urls": [
      "https://example.com",
      "https://example.org",
      "https://example.net"
    ],
    "selector": "h1"
  }
}
```

**Response**

```json
{
  "https://example.com": ["Example Domain"],
  "https://example.org": ["Example Domain"],
  "https://example.net": ["Example Domain"]
}
```

---

## Notes

- All requests are sent with `User-Agent: Mozilla/5.0 (compatible; scraper-mcp-server/1.0)`.
- `scrape_static` and `scrape_multiple` use a 30-second HTTP timeout per request.
- `scrape_js` requires Chrome/Chromium to be installed on the host.
- Empty or whitespace-only matches are silently dropped from results.

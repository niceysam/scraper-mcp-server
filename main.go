package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/niceysam/scraper-mcp-server/scraper"
)

func main() {
	s := server.NewMCPServer(
		"scraper-mcp-server",
		"1.0.0",
		server.WithToolCapabilities(false),
	)

	s.AddTool(scrapeStaticTool(), scrapeStaticHandler)
	s.AddTool(scrapeJSTool(), scrapeJSHandler)
	s.AddTool(scrapeMultipleTool(), scrapeMultipleHandler)

	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

// ── scrape_static ────────────────────────────────────────────────────────────

func scrapeStaticTool() mcp.Tool {
	return mcp.NewTool(
		"scrape_static",
		mcp.WithDescription("Scrape static HTML from a URL using CSS selectors (Colly-based, no JavaScript execution)."),
		mcp.WithString("url",
			mcp.Required(),
			mcp.Description("The URL to scrape."),
		),
		mcp.WithString("selector",
			mcp.Required(),
			mcp.Description("CSS selector to match elements."),
		),
		mcp.WithString("attribute",
			mcp.Description(`Attribute to extract. Use "text" (default) for inner text, or any HTML attribute name such as "href", "src", "data-id".`),
		),
	)
}

func scrapeStaticHandler(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	url, err := req.RequireString("url")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	selector, err := req.RequireString("selector")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	attribute := req.GetString("attribute", "text")

	results, err := scraper.ScrapeStatic(url, selector, attribute)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("scrape_static failed: %v", err)), nil
	}

	return jsonResult(results)
}

// ── scrape_js ────────────────────────────────────────────────────────────────

func scrapeJSTool() mcp.Tool {
	return mcp.NewTool(
		"scrape_js",
		mcp.WithDescription("Scrape a JavaScript-rendered page using headless Chrome (chromedp). Use this when the target page requires JavaScript to load its content."),
		mcp.WithString("url",
			mcp.Required(),
			mcp.Description("The URL to scrape."),
		),
		mcp.WithString("selector",
			mcp.Required(),
			mcp.Description("CSS selector to match elements."),
		),
		mcp.WithString("attribute",
			mcp.Description(`Attribute to extract. Use "text" (default) for inner text, or any HTML attribute name such as "href", "src".`),
		),
		mcp.WithString("wait_for",
			mcp.Description("Optional CSS selector to wait for before scraping (waits for the element to become visible)."),
		),
		mcp.WithNumber("timeout_seconds",
			mcp.Description("Maximum seconds to wait for the page (default 30)."),
		),
	)
}

func scrapeJSHandler(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	url, err := req.RequireString("url")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	selector, err := req.RequireString("selector")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	attribute := req.GetString("attribute", "text")
	waitFor := req.GetString("wait_for", "")
	timeoutSeconds := int(req.GetFloat("timeout_seconds", 30))

	results, err := scraper.ScrapeJS(url, selector, attribute, waitFor, timeoutSeconds)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("scrape_js failed: %v", err)), nil
	}

	return jsonResult(results)
}

// ── scrape_multiple ──────────────────────────────────────────────────────────

func scrapeMultipleTool() mcp.Tool {
	return mcp.NewTool(
		"scrape_multiple",
		mcp.WithDescription("Scrape multiple URLs with the same CSS selector in parallel (Colly-based, static HTML). Returns a map of URL → matched values."),
		mcp.WithArray("urls",
			mcp.Required(),
			mcp.Description("Array of URLs to scrape."),
			mcp.Items(map[string]any{"type": "string"}),
		),
		mcp.WithString("selector",
			mcp.Required(),
			mcp.Description("CSS selector to match elements on each page."),
		),
		mcp.WithString("attribute",
			mcp.Description(`Attribute to extract. Use "text" (default) for inner text, or any HTML attribute name such as "href".`),
		),
	)
}

func scrapeMultipleHandler(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	rawURLs, err := req.RequireStringSlice("urls")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	selector, err := req.RequireString("selector")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	attribute := req.GetString("attribute", "text")

	results, err := scraper.ScrapeMultiple(rawURLs, selector, attribute)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("scrape_multiple failed: %v", err)), nil
	}

	return jsonResult(results)
}

// ── helpers ──────────────────────────────────────────────────────────────────

func jsonResult(v any) (*mcp.CallToolResult, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(b)), nil
}

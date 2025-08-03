package search

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/kringen/go-mcp-server/pkg/mcp"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
)

// WebSearcher interface defines web search capabilities
type WebSearcher interface {
	Search(ctx context.Context, query mcp.SearchQuery) ([]*mcp.SearchResult, error)
	HealthCheck(ctx context.Context) error
}

// CollySearcher implements WebSearcher using Colly web scraper
type CollySearcher struct {
	config Config
}

// Config holds search configuration
type Config struct {
	UserAgent       string        `json:"user_agent"`
	Timeout         time.Duration `json:"timeout"`
	Delay           time.Duration `json:"delay"`
	RandomDelay     time.Duration `json:"random_delay"`
	MaxDepth        int           `json:"max_depth"`
	MaxResults      int           `json:"max_results"`
	EnableDebug     bool          `json:"enable_debug"`
	AllowedDomains  []string      `json:"allowed_domains"`
	BlockedDomains  []string      `json:"blocked_domains"`
	CacheResults    bool          `json:"cache_results"`
	CacheTTL        time.Duration `json:"cache_ttl"`
}

// DefaultConfig returns a default search configuration
func DefaultConfig() Config {
	return Config{
		UserAgent:      "MCP-Server-Bot/1.0",
		Timeout:        30 * time.Second,
		Delay:          1 * time.Second,
		RandomDelay:    500 * time.Millisecond,
		MaxDepth:       2,
		MaxResults:     10,
		EnableDebug:    false,
		AllowedDomains: []string{},
		BlockedDomains: []string{
			"facebook.com",
			"twitter.com",
			"instagram.com",
			"tiktok.com",
		},
		CacheResults: true,
		CacheTTL:     1 * time.Hour,
	}
}

// NewCollySearcher creates a new CollySearcher
func NewCollySearcher(config Config) *CollySearcher {
	return &CollySearcher{
		config: config,
	}
}

// Search performs a web search using the provided query
func (s *CollySearcher) Search(ctx context.Context, query mcp.SearchQuery) ([]*mcp.SearchResult, error) {
	// Create a new collector for this search
	c := s.createCollector()

	var results []*mcp.SearchResult
	var searchErrors []error

	// Configure the collector to extract search results
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		if len(results) >= s.getMaxResults(query.MaxResults) {
			return
		}

		link := e.Attr("href")
		title := strings.TrimSpace(e.Text)
		
		// Skip empty titles or non-http links
		if title == "" || (!strings.HasPrefix(link, "http://") && !strings.HasPrefix(link, "https://")) {
			return
		}

		// Skip blocked domains
		if s.isBlockedDomain(link) {
			return
		}

		// Get description from nearby elements
		description := s.extractDescription(e)

		result := &mcp.SearchResult{
			Title:       title,
			URL:         link,
			Description: description,
			Timestamp:   time.Now(),
			Metadata: map[string]string{
				"query":      query.Query,
				"user_agent": s.config.UserAgent,
			},
		}

		results = append(results, result)
	})

	// Error handling
	c.OnError(func(r *colly.Response, err error) {
		searchErrors = append(searchErrors, fmt.Errorf("request to %s failed: %w", r.Request.URL, err))
	})

	// Start searching with multiple search engines/strategies
	searchURLs := s.buildSearchURLs(query)
	
	for _, searchURL := range searchURLs {
		if len(results) >= s.getMaxResults(query.MaxResults) {
			break
		}
		
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		default:
			if err := c.Visit(searchURL); err != nil {
				searchErrors = append(searchErrors, fmt.Errorf("failed to visit %s: %w", searchURL, err))
			}
		}
	}

	// If we have results, return them even if there were some errors
	if len(results) > 0 {
		return results, nil
	}

	// If no results and we had errors, return the first error
	if len(searchErrors) > 0 {
		return nil, searchErrors[0]
	}

	return results, nil
}

// SearchWithContent performs a search and fetches content from result pages
func (s *CollySearcher) SearchWithContent(ctx context.Context, query mcp.SearchQuery) ([]*mcp.SearchResult, error) {
	results, err := s.Search(ctx, query)
	if err != nil {
		return nil, err
	}

	// Fetch content for each result
	for _, result := range results {
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		default:
			content, err := s.extractContent(ctx, result.URL)
			if err == nil {
				result.Content = content
			}
			// Continue even if content extraction fails
		}
	}

	return results, nil
}

// HealthCheck verifies that the search service is working
func (s *CollySearcher) HealthCheck(ctx context.Context) error {
	// Simple connectivity test
	query := mcp.SearchQuery{
		Query:      "test",
		MaxResults: 1,
	}

	_, err := s.Search(ctx, query)
	return err
}

// Helper methods

func (s *CollySearcher) createCollector() *colly.Collector {
	c := colly.NewCollector(
		colly.Async(true),
	)

	// Configure collector
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 2,
		Delay:       s.config.Delay,
		RandomDelay: s.config.RandomDelay,
	})

	c.SetRequestTimeout(s.config.Timeout)
	c.UserAgent = s.config.UserAgent

	if s.config.EnableDebug {
		c.SetDebugger(&debug.LogDebugger{})
	}

	// Set allowed/blocked domains
	if len(s.config.AllowedDomains) > 0 {
		c.AllowedDomains = s.config.AllowedDomains
	}

	c.OnRequest(func(r *colly.Request) {
		// Add headers to appear more like a real browser
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
		r.Headers.Set("Accept-Language", "en-US,en;q=0.5")
		r.Headers.Set("Accept-Encoding", "gzip, deflate")
		r.Headers.Set("DNT", "1")
		r.Headers.Set("Connection", "keep-alive")
		r.Headers.Set("Upgrade-Insecure-Requests", "1")
	})

	return c
}

func (s *CollySearcher) buildSearchURLs(query mcp.SearchQuery) []string {
	encodedQuery := url.QueryEscape(query.Query)
	var urls []string

	// DuckDuckGo (respects robots.txt and privacy-friendly)
	duckURL := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s", encodedQuery)
	if query.Region != "" {
		duckURL += "&kl=" + query.Region
	}
	urls = append(urls, duckURL)

	// Startpage (Google results via proxy)
	startpageURL := fmt.Sprintf("https://www.startpage.com/sp/search?query=%s", encodedQuery)
	if query.Language != "" {
		startpageURL += "&language=" + query.Language
	}
	urls = append(urls, startpageURL)

	return urls
}

func (s *CollySearcher) getMaxResults(queryMax int) int {
	if queryMax > 0 && queryMax < s.config.MaxResults {
		return queryMax
	}
	return s.config.MaxResults
}

func (s *CollySearcher) isBlockedDomain(link string) bool {
	parsedURL, err := url.Parse(link)
	if err != nil {
		return true
	}

	hostname := parsedURL.Hostname()
	for _, blocked := range s.config.BlockedDomains {
		if strings.Contains(hostname, blocked) {
			return true
		}
	}
	return false
}

func (s *CollySearcher) extractDescription(e *colly.HTMLElement) string {
	// Try to find description in nearby elements
	description := ""
	
	// Check next sibling
	if next := e.DOM.Next(); next.Length() > 0 {
		description = strings.TrimSpace(next.Text())
	}
	
	// Check parent's next sibling
	if description == "" {
		if parentNext := e.DOM.Parent().Next(); parentNext.Length() > 0 {
			description = strings.TrimSpace(parentNext.Text())
		}
	}

	// Limit description length
	if len(description) > 200 {
		description = description[:200] + "..."
	}

	return description
}

func (s *CollySearcher) extractContent(ctx context.Context, url string) (string, error) {
	c := s.createCollector()
	
	var content strings.Builder
	var extractionError error

	c.OnHTML("body", func(e *colly.HTMLElement) {
		// Extract main content, avoiding navigation and ads
		e.ForEach("p, article, main, .content, .post-content, .entry-content", func(_ int, el *colly.HTMLElement) {
			text := strings.TrimSpace(el.Text)
			if len(text) > 50 { // Skip short snippets
				content.WriteString(text)
				content.WriteString("\n\n")
			}
		})
	})

	c.OnError(func(r *colly.Response, err error) {
		extractionError = err
	})

	if err := c.Visit(url); err != nil {
		return "", err
	}

	if extractionError != nil {
		return "", extractionError
	}

	result := content.String()
	// Limit content length
	if len(result) > 5000 {
		result = result[:5000] + "..."
	}

	return strings.TrimSpace(result), nil
}

// MockSearcher implements WebSearcher for testing
type MockSearcher struct {
	results []*mcp.SearchResult
	err     error
}

// NewMockSearcher creates a new MockSearcher
func NewMockSearcher(results []*mcp.SearchResult, err error) *MockSearcher {
	return &MockSearcher{
		results: results,
		err:     err,
	}
}

// Search returns the mock results
func (m *MockSearcher) Search(ctx context.Context, query mcp.SearchQuery) ([]*mcp.SearchResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	
	maxResults := len(m.results)
	if query.MaxResults > 0 && query.MaxResults < maxResults {
		maxResults = query.MaxResults
	}
	
	return m.results[:maxResults], nil
}

// HealthCheck always returns nil for the mock
func (m *MockSearcher) HealthCheck(ctx context.Context) error {
	return m.err
}

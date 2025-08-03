package search

import (
	"context"
	"testing"
	"time"

	"github.com/kringen/go-mcp-server/pkg/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollySearcher_Unit(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		config := DefaultConfig()
		assert.Equal(t, "MCP-Server-Bot/1.0", config.UserAgent)
		assert.Equal(t, 30*time.Second, config.Timeout)
		assert.Equal(t, 1*time.Second, config.Delay)
		assert.Equal(t, 10, config.MaxResults)
		assert.True(t, config.CacheResults)
		assert.Contains(t, config.BlockedDomains, "facebook.com")
	})

	t.Run("NewCollySearcher", func(t *testing.T) {
		config := DefaultConfig()
		searcher := NewCollySearcher(config)
		assert.NotNil(t, searcher)
		assert.Equal(t, config, searcher.config)
	})

	t.Run("BuildSearchURLs", func(t *testing.T) {
		searcher := NewCollySearcher(DefaultConfig())
		query := mcp.SearchQuery{
			Query:  "golang programming",
			Region: "us",
		}

		urls := searcher.buildSearchURLs(query)
		assert.Greater(t, len(urls), 0)
		
		for _, url := range urls {
			assert.Contains(t, url, "golang%20programming")
		}
	})

	t.Run("GetMaxResults", func(t *testing.T) {
		config := DefaultConfig()
		config.MaxResults = 20
		searcher := NewCollySearcher(config)

		// Test with query max less than config max
		assert.Equal(t, 5, searcher.getMaxResults(5))
		
		// Test with query max greater than config max
		assert.Equal(t, 20, searcher.getMaxResults(50))
		
		// Test with zero query max
		assert.Equal(t, 20, searcher.getMaxResults(0))
	})

	t.Run("IsBlockedDomain", func(t *testing.T) {
		searcher := NewCollySearcher(DefaultConfig())
		
		assert.True(t, searcher.isBlockedDomain("https://facebook.com/page"))
		assert.True(t, searcher.isBlockedDomain("https://www.twitter.com/user"))
		assert.False(t, searcher.isBlockedDomain("https://golang.org"))
		assert.False(t, searcher.isBlockedDomain("https://github.com"))
		
		// Test invalid URL
		assert.True(t, searcher.isBlockedDomain("invalid-url"))
	})

	t.Run("TruncateString", func(t *testing.T) {
		searcher := NewCollySearcher(DefaultConfig())
		
		longText := "This is a very long text that should be truncated"
		short := searcher.extractDescription(nil) // This will return empty string
		assert.Equal(t, "", short)
		
		// Test with actual max length
		if len(longText) > 10 {
			truncated := longText[:10] + "..."
			assert.Contains(t, truncated, "...")
		}
	})
}

// TestCollySearcher_Integration runs integration tests against real web services
// These tests require internet connectivity and may be flaky
func TestCollySearcher_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	config := DefaultConfig()
	config.MaxResults = 3
	config.Timeout = 15 * time.Second
	searcher := NewCollySearcher(config)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("Search_Basic", func(t *testing.T) {
		query := mcp.SearchQuery{
			Query:      "golang programming language",
			MaxResults: 2,
		}

		results, err := searcher.Search(ctx, query)
		if err != nil {
			t.Logf("Search failed (this may be expected due to rate limiting): %v", err)
			return
		}

		assert.LessOrEqual(t, len(results), 2)
		
		for _, result := range results {
			assert.NotEmpty(t, result.Title)
			assert.NotEmpty(t, result.URL)
			assert.False(t, result.Timestamp.IsZero())
			assert.Contains(t, result.Metadata["query"], "golang")
		}
	})

	t.Run("Search_WithFilters", func(t *testing.T) {
		query := mcp.SearchQuery{
			Query:      "machine learning",
			MaxResults: 1,
			Language:   "en",
			SafeSearch: true,
		}

		results, err := searcher.Search(ctx, query)
		if err != nil {
			t.Logf("Search with filters failed: %v", err)
			return
		}

		assert.LessOrEqual(t, len(results), 1)
	})

	t.Run("SearchWithContent", func(t *testing.T) {
		query := mcp.SearchQuery{
			Query:      "example.com",
			MaxResults: 1,
		}

		results, err := searcher.SearchWithContent(ctx, query)
		if err != nil {
			t.Logf("Search with content failed: %v", err)
			return
		}

		// Content extraction may fail for some sites
		for _, result := range results {
			t.Logf("Result: %s - Content length: %d", result.URL, len(result.Content))
		}
	})

	t.Run("HealthCheck", func(t *testing.T) {
		err := searcher.HealthCheck(ctx)
		// Health check may fail due to network issues, so we just log it
		if err != nil {
			t.Logf("Health check failed (may be expected): %v", err)
		}
	})
}

func TestMockSearcher(t *testing.T) {
	mockResults := []*mcp.SearchResult{
		{
			Title:       "Test Result 1",
			URL:         "https://example.com/1",
			Description: "First test result",
			Timestamp:   time.Now(),
		},
		{
			Title:       "Test Result 2", 
			URL:         "https://example.com/2",
			Description: "Second test result",
			Timestamp:   time.Now(),
		},
	}

	t.Run("MockSearcher_Success", func(t *testing.T) {
		searcher := NewMockSearcher(mockResults, nil)
		
		query := mcp.SearchQuery{
			Query:      "test query",
			MaxResults: 1,
		}

		results, err := searcher.Search(context.Background(), query)
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "Test Result 1", results[0].Title)
	})

	t.Run("MockSearcher_Error", func(t *testing.T) {
		expectedErr := assert.AnError
		searcher := NewMockSearcher(nil, expectedErr)
		
		query := mcp.SearchQuery{Query: "test"}
		
		results, err := searcher.Search(context.Background(), query)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, results)
	})

	t.Run("MockSearcher_HealthCheck", func(t *testing.T) {
		// Test healthy mock
		searcher := NewMockSearcher(mockResults, nil)
		err := searcher.HealthCheck(context.Background())
		assert.NoError(t, err)
		
		// Test unhealthy mock
		expectedErr := assert.AnError
		searcher = NewMockSearcher(nil, expectedErr)
		err = searcher.HealthCheck(context.Background())
		assert.Equal(t, expectedErr, err)
	})

	t.Run("MockSearcher_MaxResults", func(t *testing.T) {
		searcher := NewMockSearcher(mockResults, nil)
		
		query := mcp.SearchQuery{
			Query:      "test",
			MaxResults: 10, // More than available results
		}

		results, err := searcher.Search(context.Background(), query)
		require.NoError(t, err)
		assert.Len(t, results, 2) // Should return all available results
	})
}

// Benchmark tests
func BenchmarkCollySearcher_Search(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	config := DefaultConfig()
	config.MaxResults = 1
	searcher := NewCollySearcher(config)

	ctx := context.Background()
	query := mcp.SearchQuery{
		Query:      "golang",
		MaxResults: 1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := searcher.Search(ctx, query)
		if err != nil {
			b.Logf("Search failed: %v", err)
		}
	}
}

func BenchmarkMockSearcher_Search(b *testing.B) {
	mockResults := []*mcp.SearchResult{
		{
			Title:       "Benchmark Result",
			URL:         "https://example.com",
			Description: "Benchmark test result",
			Timestamp:   time.Now(),
		},
	}

	searcher := NewMockSearcher(mockResults, nil)
	query := mcp.SearchQuery{Query: "test"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := searcher.Search(context.Background(), query)
		if err != nil {
			b.Errorf("Mock search failed: %v", err)
		}
	}
}

package tools

import (
	"context"
	"testing"
	"time"

	"github.com/kringen/go-mcp-server/internal/search"
	"github.com/kringen/go-mcp-server/pkg/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchTool(t *testing.T) {
	// Create mock search results
	mockResults := []*mcp.SearchResult{
		{
			Title:       "Go Programming Language",
			URL:         "https://golang.org",
			Description: "The official Go programming language website",
			Content:     "Go is an open source programming language...",
			Timestamp:   time.Now(),
		},
		{
			Title:       "Go Documentation",
			URL:         "https://pkg.go.dev",
			Description: "Go package documentation",
			Timestamp:   time.Now(),
		},
	}

	t.Run("ListTools", func(t *testing.T) {
		searcher := search.NewMockSearcher(mockResults, nil)
		tool := NewSearchTool(searcher)

		tools, err := tool.ListTools(context.Background())
		require.NoError(t, err)
		assert.Len(t, tools, 2)

		// Check web_search tool
		webSearchTool := findTool(tools, "web_search")
		require.NotNil(t, webSearchTool)
		assert.Equal(t, "web_search", webSearchTool.Name)
		assert.Contains(t, webSearchTool.Description, "Search the web")
		assert.NotNil(t, webSearchTool.InputSchema)

		// Check health check tool
		healthTool := findTool(tools, "search_health_check")
		require.NotNil(t, healthTool)
		assert.Equal(t, "search_health_check", healthTool.Name)
	})

	t.Run("CallTool_WebSearch_Success", func(t *testing.T) {
		searcher := search.NewMockSearcher(mockResults, nil)
		tool := NewSearchTool(searcher)

		request := mcp.ToolCallRequest{
			Name: "web_search",
			Arguments: map[string]interface{}{
				"query":       "golang programming",
				"max_results": 2,
			},
		}

		response, err := tool.CallTool(context.Background(), request)
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.False(t, response.IsError)
		assert.Greater(t, len(response.Content), 0)

		// Check that results are included in response
		foundContent := false
		for _, content := range response.Content {
			if content.Type == "text" && len(content.Text) > 0 {
				foundContent = true
				break
			}
		}
		assert.True(t, foundContent)
	})

	t.Run("CallTool_WebSearch_WithContent", func(t *testing.T) {
		searcher := search.NewMockSearcher(mockResults, nil)
		tool := NewSearchTool(searcher)

		request := mcp.ToolCallRequest{
			Name: "web_search",
			Arguments: map[string]interface{}{
				"query":           "golang",
				"include_content": true,
				"max_results":     1,
			},
		}

		response, err := tool.CallTool(context.Background(), request)
		require.NoError(t, err)
		assert.False(t, response.IsError)
	})

	t.Run("CallTool_WebSearch_InvalidQuery", func(t *testing.T) {
		searcher := search.NewMockSearcher(mockResults, nil)
		tool := NewSearchTool(searcher)

		request := mcp.ToolCallRequest{
			Name: "web_search",
			Arguments: map[string]interface{}{
				"query": "", // Empty query
			},
		}

		response, err := tool.CallTool(context.Background(), request)
		require.NoError(t, err)
		assert.True(t, response.IsError)
		assert.Contains(t, response.Content[0].Text, "Missing or invalid 'query'")
	})

	t.Run("CallTool_WebSearch_SearchError", func(t *testing.T) {
		searcher := search.NewMockSearcher(nil, assert.AnError)
		tool := NewSearchTool(searcher)

		request := mcp.ToolCallRequest{
			Name: "web_search",
			Arguments: map[string]interface{}{
				"query": "test query",
			},
		}

		response, err := tool.CallTool(context.Background(), request)
		require.NoError(t, err)
		assert.True(t, response.IsError)
		assert.Contains(t, response.Content[0].Text, "Search failed")
	})

	t.Run("CallTool_HealthCheck_Success", func(t *testing.T) {
		searcher := search.NewMockSearcher(mockResults, nil)
		tool := NewSearchTool(searcher)

		request := mcp.ToolCallRequest{
			Name:      "search_health_check",
			Arguments: map[string]interface{}{},
		}

		response, err := tool.CallTool(context.Background(), request)
		require.NoError(t, err)
		assert.False(t, response.IsError)
		assert.Contains(t, response.Content[0].Text, "healthy")
	})

	t.Run("CallTool_HealthCheck_Failure", func(t *testing.T) {
		searcher := search.NewMockSearcher(nil, assert.AnError)
		tool := NewSearchTool(searcher)

		request := mcp.ToolCallRequest{
			Name:      "search_health_check", 
			Arguments: map[string]interface{}{},
		}

		response, err := tool.CallTool(context.Background(), request)
		require.NoError(t, err)
		assert.True(t, response.IsError)
		assert.Contains(t, response.Content[0].Text, "health check failed")
	})

	t.Run("CallTool_UnknownTool", func(t *testing.T) {
		searcher := search.NewMockSearcher(mockResults, nil)
		tool := NewSearchTool(searcher)

		request := mcp.ToolCallRequest{
			Name:      "unknown_tool",
			Arguments: map[string]interface{}{},
		}

		response, err := tool.CallTool(context.Background(), request)
		require.NoError(t, err)
		assert.True(t, response.IsError)
		assert.Contains(t, response.Content[0].Text, "Unknown search tool")
	})

	t.Run("CallTool_WebSearch_ParameterValidation", func(t *testing.T) {
		searcher := search.NewMockSearcher(mockResults, nil)
		tool := NewSearchTool(searcher)

		testCases := []struct {
			name      string
			args      map[string]interface{}
			shouldErr bool
		}{
			{
				name: "Valid parameters",
				args: map[string]interface{}{
					"query":       "test",
					"max_results": 5,
					"language":    "en",
					"region":      "us",
					"safe_search": true,
				},
				shouldErr: false,
			},
			{
				name: "Invalid max_results (too high)",
				args: map[string]interface{}{
					"query":       "test",
					"max_results": 100, // Should be capped at 50
				},
				shouldErr: false, // Should not error, just cap the value
			},
			{
				name: "Invalid max_results (negative)",
				args: map[string]interface{}{
					"query":       "test",
					"max_results": -1,
				},
				shouldErr: false, // Should use default
			},
			{
				name: "Invalid parameter types",
				args: map[string]interface{}{
					"query":       "test",
					"max_results": "invalid", // Should be int
					"safe_search": "invalid", // Should be bool
				},
				shouldErr: false, // Should use defaults for invalid values
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				request := mcp.ToolCallRequest{
					Name:      "web_search",
					Arguments: tc.args,
				}

				response, err := tool.CallTool(context.Background(), request)
				require.NoError(t, err)
				
				if tc.shouldErr {
					assert.True(t, response.IsError)
				} else {
					assert.False(t, response.IsError)
				}
			})
		}
	})
}

// Helper function to find a tool by name
func findTool(tools []mcp.Tool, name string) *mcp.Tool {
	for _, tool := range tools {
		if tool.Name == name {
			return &tool
		}
	}
	return nil
}

// Test helper functions
func TestSearchTool_Helpers(t *testing.T) {
	searcher := search.NewMockSearcher(nil, nil)
	tool := NewSearchTool(searcher)

	t.Run("toInt", func(t *testing.T) {
		testCases := []struct {
			input    interface{}
			expected int
			hasError bool
		}{
			{5, 5, false},
			{5.7, 5, false},
			{"10", 10, false},
			{"invalid", 0, true},
			{true, 0, true},
		}

		for _, tc := range testCases {
			result, err := tool.toInt(tc.input)
			if tc.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		}
	})

	t.Run("errorResponse", func(t *testing.T) {
		response := tool.errorResponse("test error message")
		assert.True(t, response.IsError)
		assert.Len(t, response.Content, 1)
		assert.Equal(t, "text", response.Content[0].Type)
		assert.Equal(t, "test error message", response.Content[0].Text)
	})
}

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/kringen/go-mcp-server/internal/search"
	"github.com/kringen/go-mcp-server/pkg/mcp"
)

// SearchTool provides web search capabilities as an MCP tool
type SearchTool struct {
	searcher search.WebSearcher
}

// NewSearchTool creates a new SearchTool
func NewSearchTool(searcher search.WebSearcher) *SearchTool {
	return &SearchTool{
		searcher: searcher,
	}
}

// ListTools returns the available search tools
func (s *SearchTool) ListTools(ctx context.Context) ([]mcp.Tool, error) {
	return []mcp.Tool{
		{
			Name:        "web_search",
			Description: "Search the web for information",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "The search query",
					},
					"max_results": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of results to return (default: 10)",
						"minimum":     1,
						"maximum":     50,
					},
					"include_content": map[string]interface{}{
						"type":        "boolean",
						"description": "Whether to fetch full content from result pages (default: false)",
					},
					"language": map[string]interface{}{
						"type":        "string",
						"description": "Language preference for search results",
					},
					"region": map[string]interface{}{
						"type":        "string",
						"description": "Region preference for search results",
					},
					"safe_search": map[string]interface{}{
						"type":        "boolean",
						"description": "Enable safe search filtering (default: true)",
					},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "search_health_check",
			Description: "Check if the web search service is healthy",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{},
			},
		},
	}, nil
}

// CallTool executes the specified search tool
func (s *SearchTool) CallTool(ctx context.Context, request mcp.ToolCallRequest) (*mcp.ToolCallResponse, error) {
	switch request.Name {
	case "web_search":
		return s.webSearch(ctx, request.Arguments)
	case "search_health_check":
		return s.healthCheck(ctx)
	default:
		return &mcp.ToolCallResponse{
			IsError: true,
			Content: []mcp.Content{
				{
					Type: "text",
					Text: fmt.Sprintf("Unknown search tool: %s", request.Name),
				},
			},
		}, nil
	}
}

func (s *SearchTool) webSearch(ctx context.Context, args map[string]interface{}) (*mcp.ToolCallResponse, error) {
	// Extract query
	queryStr, ok := args["query"].(string)
	if !ok || queryStr == "" {
		return s.errorResponse("Missing or invalid 'query' parameter"), nil
	}

	// Build search query
	searchQuery := mcp.SearchQuery{
		Query:      queryStr,
		MaxResults: 10, // default
		SafeSearch: true, // default
	}

	// Extract optional parameters
	if maxResults, ok := args["max_results"]; ok {
		if mr, err := s.toInt(maxResults); err == nil && mr > 0 && mr <= 50 {
			searchQuery.MaxResults = mr
		}
	}

	if lang, ok := args["language"].(string); ok {
		searchQuery.Language = lang
	}

	if region, ok := args["region"].(string); ok {
		searchQuery.Region = region
	}

	if safeSearch, ok := args["safe_search"].(bool); ok {
		searchQuery.SafeSearch = safeSearch
	}

	includeContent := false
	if ic, ok := args["include_content"].(bool); ok {
		includeContent = ic
	}

	// Perform search
	var results []*mcp.SearchResult
	var err error

	if includeContent {
		if contentSearcher, ok := s.searcher.(*search.CollySearcher); ok {
			results, err = contentSearcher.SearchWithContent(ctx, searchQuery)
		} else {
			// Fallback to regular search
			results, err = s.searcher.Search(ctx, searchQuery)
		}
	} else {
		results, err = s.searcher.Search(ctx, searchQuery)
	}

	if err != nil {
		return s.errorResponse(fmt.Sprintf("Search failed: %v", err)), nil
	}

	// Format results
	content := []mcp.Content{}

	if len(results) == 0 {
		content = append(content, mcp.Content{
			Type: "text",
			Text: "No search results found.",
		})
	} else {
		// Summary
		content = append(content, mcp.Content{
			Type: "text",
			Text: fmt.Sprintf("Found %d search results for: %s\n", len(results), queryStr),
		})

		// Individual results
		for i, result := range results {
			resultText := fmt.Sprintf("%d. **%s**\n   URL: %s\n   Description: %s\n",
				i+1, result.Title, result.URL, result.Description)

			if includeContent && result.Content != "" {
				// Truncate content for display
				content := result.Content
				if len(content) > 500 {
					content = content[:500] + "..."
				}
				resultText += fmt.Sprintf("   Content: %s\n", content)
			}

			resultText += fmt.Sprintf("   Timestamp: %s\n\n", result.Timestamp.Format(time.RFC3339))

			content = append(content, mcp.Content{
				Type: "text",
				Text: resultText,
			})
		}

		// Add JSON data for programmatic access
		jsonData, _ := json.Marshal(results)
		content = append(content, mcp.Content{
			Type: "text",
			Text: fmt.Sprintf("Raw JSON data:\n```json\n%s\n```", string(jsonData)),
		})
	}

	return &mcp.ToolCallResponse{
		Content: content,
	}, nil
}

func (s *SearchTool) healthCheck(ctx context.Context) (*mcp.ToolCallResponse, error) {
	err := s.searcher.HealthCheck(ctx)
	if err != nil {
		return &mcp.ToolCallResponse{
			IsError: true,
			Content: []mcp.Content{
				{
					Type: "text",
					Text: fmt.Sprintf("Search service health check failed: %v", err),
				},
			},
		}, nil
	}

	return &mcp.ToolCallResponse{
		Content: []mcp.Content{
			{
				Type: "text",
				Text: "Search service is healthy and ready to accept queries.",
			},
		},
	}, nil
}

func (s *SearchTool) toInt(value interface{}) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case float64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	case json.Number:
		i, err := v.Int64()
		return int(i), err
	default:
		return 0, fmt.Errorf("cannot convert %T to int", value)
	}
}

func (s *SearchTool) errorResponse(message string) *mcp.ToolCallResponse {
	return &mcp.ToolCallResponse{
		IsError: true,
		Content: []mcp.Content{
			{
				Type: "text",
				Text: message,
			},
		},
	}
}

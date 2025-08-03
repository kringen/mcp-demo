// +build integration

package tests

import (
	"context"
	"encoding/json"
	"net/url"
	"testing"
	"time"

	"github.com/kringen/go-mcp-server/internal/database"
	"github.com/kringen/go-mcp-server/internal/search"
	"github.com/kringen/go-mcp-server/internal/server"
	"github.com/kringen/go-mcp-server/pkg/mcp"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMCPServerIntegration tests the complete MCP server integration
func TestMCPServerIntegration(t *testing.T) {
	// Skip if no MongoDB is available
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	// Setup test context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Initialize test database
	dbConfig := database.Config{
		URI:            "mongodb://admin:password@localhost:27017",
		Database:       "mcp_server_integration_test",
		ConnectTimeout: 5 * time.Second,
		QueryTimeout:   10 * time.Second,
	}

	db, err := database.NewMongoDB(dbConfig)
	require.NoError(t, err, "Failed to connect to test database")
	defer func() {
		// Cleanup test database
		db.Close(context.Background())
	}()

	// Initialize search service
	searchConfig := search.DefaultConfig()
	searchConfig.MaxResults = 2
	searcher := search.NewCollySearcher(searchConfig)

	// Create and start MCP server
	serverConfig := server.DefaultConfig()
	serverConfig.Port = 8081 // Use different port for testing
	mcpServer := server.NewServer(serverConfig, db, searcher)

	// Start server
	serverCtx, serverCancel := context.WithCancel(ctx)
	defer serverCancel()

	go func() {
		err := mcpServer.Start(serverCtx, "localhost:8081")
		if err != nil {
			t.Logf("Server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(2 * time.Second)

	// Test MCP protocol over WebSocket
	t.Run("MCPProtocolFlow", func(t *testing.T) {
		// Connect to WebSocket
		u := url.URL{Scheme: "ws", Host: "localhost:8081", Path: "/mcp"}
		conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		require.NoError(t, err, "Failed to connect to WebSocket")
		defer conn.Close()

		// Test initialize
		initRequest := mcp.Message{
			JSONRPC: "2.0",
			ID:      1,
			Method:  "initialize",
			Params: map[string]interface{}{
				"protocolVersion": mcp.ProtocolVersion,
				"capabilities":    map[string]interface{}{},
				"clientInfo": map[string]interface{}{
					"name":    "integration-test",
					"version": "1.0.0",
				},
			},
		}

		err = conn.WriteJSON(&initRequest)
		require.NoError(t, err)

		var initResponse mcp.Message
		err = conn.ReadJSON(&initResponse)
		require.NoError(t, err)
		assert.Equal(t, "2.0", initResponse.JSONRPC)
		assert.Equal(t, 1, initResponse.ID)
		assert.NotNil(t, initResponse.Result)

		// Test list tools
		listToolsRequest := mcp.Message{
			JSONRPC: "2.0",
			ID:      2,
			Method:  "tools/list",
			Params:  map[string]interface{}{},
		}

		err = conn.WriteJSON(&listToolsRequest)
		require.NoError(t, err)

		var listToolsResponse mcp.Message
		err = conn.ReadJSON(&listToolsResponse)
		require.NoError(t, err)
		assert.Equal(t, 2, listToolsResponse.ID)
		assert.NotNil(t, listToolsResponse.Result)

		// Verify tools are present
		resultData, err := json.Marshal(listToolsResponse.Result)
		require.NoError(t, err)

		var toolsResult struct {
			Tools []mcp.Tool `json:"tools"`
		}
		err = json.Unmarshal(resultData, &toolsResult)
		require.NoError(t, err)
		assert.Greater(t, len(toolsResult.Tools), 0)

		// Find expected tools
		toolNames := make(map[string]bool)
		for _, tool := range toolsResult.Tools {
			toolNames[tool.Name] = true
		}

		expectedTools := []string{"add", "web_search", "db_create_document"}
		for _, expectedTool := range expectedTools {
			assert.True(t, toolNames[expectedTool], "Expected tool %s not found", expectedTool)
		}

		// Test math tool
		mathToolRequest := mcp.Message{
			JSONRPC: "2.0",
			ID:      3,
			Method:  "tools/call",
			Params: map[string]interface{}{
				"name": "add",
				"arguments": map[string]interface{}{
					"a": 5,
					"b": 3,
				},
			},
		}

		err = conn.WriteJSON(&mathToolRequest)
		require.NoError(t, err)

		var mathToolResponse mcp.Message
		err = conn.ReadJSON(&mathToolResponse)
		require.NoError(t, err)
		assert.Equal(t, 3, mathToolResponse.ID)
		assert.NotNil(t, mathToolResponse.Result)

		// Test database tool
		dbToolRequest := mcp.Message{
			JSONRPC: "2.0",
			ID:      4,
			Method:  "tools/call",
			Params: map[string]interface{}{
				"name": "db_create_document",
				"arguments": map[string]interface{}{
					"collection": "test_docs",
					"title":      "Integration Test Document",
					"content":    "This is a test document created during integration testing",
					"tags":       []string{"test", "integration"},
				},
			},
		}

		err = conn.WriteJSON(&dbToolRequest)
		require.NoError(t, err)

		var dbToolResponse mcp.Message
		err = conn.ReadJSON(&dbToolResponse)
		require.NoError(t, err)
		assert.Equal(t, 4, dbToolResponse.ID)
		assert.NotNil(t, dbToolResponse.Result)

		// Test health check tool
		healthRequest := mcp.Message{
			JSONRPC: "2.0",
			ID:      5,
			Method:  "tools/call",
			Params: map[string]interface{}{
				"name":      "db_health_check",
				"arguments": map[string]interface{}{},
			},
		}

		err = conn.WriteJSON(&healthRequest)
		require.NoError(t, err)

		var healthResponse mcp.Message
		err = conn.ReadJSON(&healthResponse)
		require.NoError(t, err)
		assert.Equal(t, 5, healthResponse.ID)
		assert.NotNil(t, healthResponse.Result)
	})

	// Test HTTP health endpoint
	t.Run("HTTPHealthEndpoint", func(t *testing.T) {
		// This would require an HTTP client test
		// For now, we'll skip this as it requires additional setup
		t.Skip("HTTP endpoint testing requires additional setup")
	})

	// Test error handling
	t.Run("ErrorHandling", func(t *testing.T) {
		u := url.URL{Scheme: "ws", Host: "localhost:8081", Path: "/mcp"}
		conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		require.NoError(t, err)
		defer conn.Close()

		// Test invalid method
		invalidRequest := mcp.Message{
			JSONRPC: "2.0",
			ID:      1,
			Method:  "invalid/method",
			Params:  map[string]interface{}{},
		}

		err = conn.WriteJSON(&invalidRequest)
		require.NoError(t, err)

		var errorResponse mcp.Message
		err = conn.ReadJSON(&errorResponse)
		require.NoError(t, err)
		assert.Equal(t, 1, errorResponse.ID)
		assert.NotNil(t, errorResponse.Error)
		assert.Equal(t, mcp.ErrorCodeMethodNotFound, errorResponse.Error.Code)

		// Test invalid tool
		invalidToolRequest := mcp.Message{
			JSONRPC: "2.0",
			ID:      2,
			Method:  "tools/call",
			Params: map[string]interface{}{
				"name":      "nonexistent_tool",
				"arguments": map[string]interface{}{},
			},
		}

		err = conn.WriteJSON(&invalidToolRequest)
		require.NoError(t, err)

		var toolErrorResponse mcp.Message
		err = conn.ReadJSON(&toolErrorResponse)
		require.NoError(t, err)
		assert.Equal(t, 2, toolErrorResponse.ID)
		assert.NotNil(t, toolErrorResponse.Error)
	})
}

// TestDatabaseIntegration tests database operations end-to-end
func TestDatabaseIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx := context.Background()

	// Setup test database
	dbConfig := database.Config{
		URI:            "mongodb://admin:password@localhost:27017",
		Database:       "mcp_server_db_integration_test",
		ConnectTimeout: 5 * time.Second,
		QueryTimeout:   10 * time.Second,
	}

	db, err := database.NewMongoDB(dbConfig)
	require.NoError(t, err)
	defer db.Close(ctx)

	collection := "integration_test_docs"

	t.Run("DocumentLifecycle", func(t *testing.T) {
		// Create document
		doc := &mcp.Document{
			Title:   "Integration Test Doc",
			Content: "This is a test document for integration testing",
			Tags:    []string{"integration", "test", "go"},
			Metadata: map[string]interface{}{
				"author":      "integration-test",
				"environment": "test",
			},
		}

		err := db.CreateDocument(ctx, collection, doc)
		require.NoError(t, err)
		assert.NotEmpty(t, doc.ID)

		// Retrieve document
		retrieved, err := db.GetDocument(ctx, collection, doc.ID)
		require.NoError(t, err)
		assert.Equal(t, doc.Title, retrieved.Title)
		assert.Equal(t, doc.Content, retrieved.Content)
		assert.Equal(t, doc.Tags, retrieved.Tags)

		// Update document
		retrieved.Title = "Updated Integration Test Doc"
		retrieved.Content = "This document has been updated"
		retrieved.Tags = append(retrieved.Tags, "updated")

		err = db.UpdateDocument(ctx, collection, retrieved)
		require.NoError(t, err)

		// Verify update
		updated, err := db.GetDocument(ctx, collection, doc.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Integration Test Doc", updated.Title)
		assert.Contains(t, updated.Tags, "updated")

		// Query documents
		query := mcp.DatabaseQuery{
			Collection: collection,
			Filter: map[string]interface{}{
				"tags": "integration",
			},
			Limit: 10,
		}

		results, err := db.QueryDocuments(ctx, query)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 1)

		// Count documents
		count, err := db.CountDocuments(ctx, collection, map[string]interface{}{
			"tags": "integration",
		})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(1))

		// Delete document
		err = db.DeleteDocument(ctx, collection, doc.ID)
		require.NoError(t, err)

		// Verify deletion
		_, err = db.GetDocument(ctx, collection, doc.ID)
		assert.Error(t, err)
	})
}

// TestSearchIntegration tests web search functionality
func TestSearchIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	config := search.DefaultConfig()
	config.MaxResults = 2
	config.Timeout = 10 * time.Second
	searcher := search.NewCollySearcher(config)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	t.Run("BasicSearch", func(t *testing.T) {
		query := mcp.SearchQuery{
			Query:      "golang programming",
			MaxResults: 1,
		}

		results, err := searcher.Search(ctx, query)
		if err != nil {
			t.Logf("Search failed (this may be expected due to rate limiting): %v", err)
			t.Skip("Skipping search test due to service unavailability")
		}

		assert.LessOrEqual(t, len(results), 1)
		
		if len(results) > 0 {
			result := results[0]
			assert.NotEmpty(t, result.Title)
			assert.NotEmpty(t, result.URL)
			assert.False(t, result.Timestamp.IsZero())
		}
	})

	t.Run("HealthCheck", func(t *testing.T) {
		err := searcher.HealthCheck(ctx)
		if err != nil {
			t.Logf("Search health check failed (may be expected): %v", err)
		}
		// Don't fail the test if health check fails due to network issues
	})
}

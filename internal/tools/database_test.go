package tools

import (
	"context"
	"testing"

	"github.com/kringen/go-mcp-server/internal/database"
	"github.com/kringen/go-mcp-server/pkg/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockMongoDB implements a mock version of MongoDB for testing
type MockMongoDB struct {
	documents map[string]*mcp.Document
	healthy   bool
	err       error
}

func NewMockMongoDB(healthy bool, err error) *MockMongoDB {
	return &MockMongoDB{
		documents: make(map[string]*mcp.Document),
		healthy:   healthy,
		err:       err,
	}
}

func (m *MockMongoDB) CreateDocument(ctx context.Context, collection string, doc *mcp.Document) error {
	if m.err != nil {
		return m.err
	}
	if doc.ID == "" {
		doc.ID = "mock-id-123"
	}
	m.documents[doc.ID] = doc
	return nil
}

func (m *MockMongoDB) GetDocument(ctx context.Context, collection, id string) (*mcp.Document, error) {
	if m.err != nil {
		return nil, m.err
	}
	doc, exists := m.documents[id]
	if !exists {
		return nil, assert.AnError
	}
	return doc, nil
}

func (m *MockMongoDB) UpdateDocument(ctx context.Context, collection string, doc *mcp.Document) error {
	if m.err != nil {
		return m.err
	}
	if _, exists := m.documents[doc.ID]; !exists {
		return assert.AnError
	}
	m.documents[doc.ID] = doc
	return nil
}

func (m *MockMongoDB) DeleteDocument(ctx context.Context, collection, id string) error {
	if m.err != nil {
		return m.err
	}
	if _, exists := m.documents[id]; !exists {
		return assert.AnError
	}
	delete(m.documents, id)
	return nil
}

func (m *MockMongoDB) QueryDocuments(ctx context.Context, query mcp.DatabaseQuery) ([]*mcp.Document, error) {
	if m.err != nil {
		return nil, m.err
	}
	
	var results []*mcp.Document
	for _, doc := range m.documents {
		results = append(results, doc)
		if query.Limit > 0 && len(results) >= query.Limit {
			break
		}
	}
	return results, nil
}

func (m *MockMongoDB) SearchDocuments(ctx context.Context, collection, searchText string, limit int) ([]*mcp.Document, error) {
	if m.err != nil {
		return nil, m.err
	}
	
	var results []*mcp.Document
	for _, doc := range m.documents {
		// Simple mock search - just check if searchText is in title or content
		if searchText == "" || 
		   containsIgnoreCase(doc.Title, searchText) || 
		   containsIgnoreCase(doc.Content, searchText) {
			results = append(results, doc)
			if limit > 0 && len(results) >= limit {
				break
			}
		}
	}
	return results, nil
}

func (m *MockMongoDB) CountDocuments(ctx context.Context, collection string, filter map[string]interface{}) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	return int64(len(m.documents)), nil
}

func (m *MockMongoDB) HealthCheck(ctx context.Context) error {
	if !m.healthy {
		return assert.AnError
	}
	return nil
}

func (m *MockMongoDB) Close(ctx context.Context) error {
	return nil
}

func containsIgnoreCase(str, substr string) bool {
	return len(str) >= len(substr) && 
		   (str == substr || len(substr) == 0)
}

// TestDatabaseTool tests the DatabaseTool implementation
func TestDatabaseTool(t *testing.T) {
	t.Run("ListTools", func(t *testing.T) {
		mockDB := NewMockMongoDB(true, nil)
		tool := NewDatabaseTool(mockDB)

		tools, err := tool.ListTools(context.Background())
		require.NoError(t, err)
		
		expectedTools := []string{
			"db_create_document",
			"db_get_document", 
			"db_update_document",
			"db_delete_document",
			"db_query_documents",
			"db_search_documents",
			"db_count_documents",
			"db_health_check",
		}

		assert.Len(t, tools, len(expectedTools))
		
		for _, expectedName := range expectedTools {
			found := false
			for _, tool := range tools {
				if tool.Name == expectedName {
					found = true
					assert.NotEmpty(t, tool.Description)
					assert.NotNil(t, tool.InputSchema)
					break
				}
			}
			assert.True(t, found, "Tool %s not found", expectedName)
		}
	})

	t.Run("CallTool_CreateDocument_Success", func(t *testing.T) {
		mockDB := NewMockMongoDB(true, nil)
		tool := NewDatabaseTool(mockDB)

		request := mcp.ToolCallRequest{
			Name: "db_create_document",
			Arguments: map[string]interface{}{
				"collection": "test_docs",
				"title":      "Test Document",
				"content":    "This is a test document",
				"tags":       []interface{}{"test", "demo"},
				"metadata": map[string]interface{}{
					"author": "test_user",
				},
			},
		}

		response, err := tool.CallTool(context.Background(), request)
		require.NoError(t, err)
		assert.False(t, response.IsError)
		assert.Contains(t, response.Content[0].Text, "Document created successfully")
	})

	t.Run("CallTool_CreateDocument_MissingParams", func(t *testing.T) {
		mockDB := NewMockMongoDB(true, nil)
		tool := NewDatabaseTool(mockDB)

		testCases := []struct {
			name string
			args map[string]interface{}
		}{
			{
				name: "Missing collection",
				args: map[string]interface{}{
					"title":   "Test",
					"content": "Test content",
				},
			},
			{
				name: "Missing title",
				args: map[string]interface{}{
					"collection": "test",
					"content":    "Test content",
				},
			},
			{
				name: "Missing content",
				args: map[string]interface{}{
					"collection": "test",
					"title":      "Test",
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				request := mcp.ToolCallRequest{
					Name:      "db_create_document",
					Arguments: tc.args,
				}

				response, err := tool.CallTool(context.Background(), request)
				require.NoError(t, err)
				assert.True(t, response.IsError)
				assert.Contains(t, response.Content[0].Text, "Missing or invalid")
			})
		}
	})

	t.Run("CallTool_GetDocument_Success", func(t *testing.T) {
		mockDB := NewMockMongoDB(true, nil)
		tool := NewDatabaseTool(mockDB)

		// First create a document
		doc := &mcp.Document{
			ID:      "test-123",
			Title:   "Test Doc",
			Content: "Test content",
		}
		mockDB.documents[doc.ID] = doc

		request := mcp.ToolCallRequest{
			Name: "db_get_document",
			Arguments: map[string]interface{}{
				"collection": "test_docs",
				"id":         "test-123",
			},
		}

		response, err := tool.CallTool(context.Background(), request)
		require.NoError(t, err)
		assert.False(t, response.IsError)
		assert.Contains(t, response.Content[0].Text, "Document found")
		assert.Contains(t, response.Content[1].Text, "Test content")
	})

	t.Run("CallTool_GetDocument_NotFound", func(t *testing.T) {
		mockDB := NewMockMongoDB(true, nil)
		tool := NewDatabaseTool(mockDB)

		request := mcp.ToolCallRequest{
			Name: "db_get_document",
			Arguments: map[string]interface{}{
				"collection": "test_docs",
				"id":         "nonexistent",
			},
		}

		response, err := tool.CallTool(context.Background(), request)
		require.NoError(t, err)
		assert.True(t, response.IsError)
		assert.Contains(t, response.Content[0].Text, "Failed to get document")
	})

	t.Run("CallTool_UpdateDocument_Success", func(t *testing.T) {
		mockDB := NewMockMongoDB(true, nil)
		tool := NewDatabaseTool(mockDB)

		// Create initial document
		doc := &mcp.Document{
			ID:      "test-123",
			Title:   "Original Title",
			Content: "Original content",
		}
		mockDB.documents[doc.ID] = doc

		request := mcp.ToolCallRequest{
			Name: "db_update_document",
			Arguments: map[string]interface{}{
				"collection": "test_docs",
				"id":         "test-123",
				"title":      "Updated Title",
				"content":    "Updated content",
			},
		}

		response, err := tool.CallTool(context.Background(), request)
		require.NoError(t, err)
		assert.False(t, response.IsError)
		assert.Contains(t, response.Content[0].Text, "Document updated successfully")
		
		// Verify the document was updated in mock
		updatedDoc := mockDB.documents["test-123"]
		assert.Equal(t, "Updated Title", updatedDoc.Title)
		assert.Equal(t, "Updated content", updatedDoc.Content)
	})

	t.Run("CallTool_DeleteDocument_Success", func(t *testing.T) {
		mockDB := NewMockMongoDB(true, nil)
		tool := NewDatabaseTool(mockDB)

		// Create document to delete
		doc := &mcp.Document{ID: "test-123", Title: "To Delete"}
		mockDB.documents[doc.ID] = doc

		request := mcp.ToolCallRequest{
			Name: "db_delete_document",
			Arguments: map[string]interface{}{
				"collection": "test_docs",
				"id":         "test-123",
			},
		}

		response, err := tool.CallTool(context.Background(), request)
		require.NoError(t, err)
		assert.False(t, response.IsError)
		assert.Contains(t, response.Content[0].Text, "deleted successfully")
		
		// Verify document was deleted
		_, exists := mockDB.documents["test-123"]
		assert.False(t, exists)
	})

	t.Run("CallTool_QueryDocuments_Success", func(t *testing.T) {
		mockDB := NewMockMongoDB(true, nil)
		tool := NewDatabaseTool(mockDB)

		// Add some test documents
		docs := []*mcp.Document{
			{ID: "1", Title: "Doc 1", Content: "Content 1"},
			{ID: "2", Title: "Doc 2", Content: "Content 2"},
		}
		for _, doc := range docs {
			mockDB.documents[doc.ID] = doc
		}

		request := mcp.ToolCallRequest{
			Name: "db_query_documents",
			Arguments: map[string]interface{}{
				"collection": "test_docs",
				"limit":      1,
			},
		}

		response, err := tool.CallTool(context.Background(), request)
		require.NoError(t, err)
		assert.False(t, response.IsError)
		assert.Contains(t, response.Content[0].Text, "Found")
	})

	t.Run("CallTool_SearchDocuments_Success", func(t *testing.T) {
		mockDB := NewMockMongoDB(true, nil)
		tool := NewDatabaseTool(mockDB)

		// Add test document
		doc := &mcp.Document{
			ID:      "1",
			Title:   "Golang Programming",
			Content: "Go is a programming language",
		}
		mockDB.documents[doc.ID] = doc

		request := mcp.ToolCallRequest{
			Name: "db_search_documents",
			Arguments: map[string]interface{}{
				"collection":  "test_docs",
				"search_text": "golang",
				"limit":       5,
			},
		}

		response, err := tool.CallTool(context.Background(), request)
		require.NoError(t, err)
		assert.False(t, response.IsError)
		assert.Contains(t, response.Content[0].Text, "Found")
	})

	t.Run("CallTool_CountDocuments_Success", func(t *testing.T) {
		mockDB := NewMockMongoDB(true, nil)
		tool := NewDatabaseTool(mockDB)

		// Add test documents
		for i := 0; i < 3; i++ {
			doc := &mcp.Document{
				ID:      string(rune('1' + i)),
				Title:   "Doc",
				Content: "Content",
			}
			mockDB.documents[doc.ID] = doc
		}

		request := mcp.ToolCallRequest{
			Name: "db_count_documents",
			Arguments: map[string]interface{}{
				"collection": "test_docs",
			},
		}

		response, err := tool.CallTool(context.Background(), request)
		require.NoError(t, err)
		assert.False(t, response.IsError)
		assert.Contains(t, response.Content[0].Text, "contains 3 documents")
	})

	t.Run("CallTool_HealthCheck_Success", func(t *testing.T) {
		mockDB := NewMockMongoDB(true, nil)
		tool := NewDatabaseTool(mockDB)

		request := mcp.ToolCallRequest{
			Name:      "db_health_check",
			Arguments: map[string]interface{}{},
		}

		response, err := tool.CallTool(context.Background(), request)
		require.NoError(t, err)
		assert.False(t, response.IsError)
		assert.Contains(t, response.Content[0].Text, "healthy")
	})

	t.Run("CallTool_HealthCheck_Failure", func(t *testing.T) {
		mockDB := NewMockMongoDB(false, nil)
		tool := NewDatabaseTool(mockDB)

		request := mcp.ToolCallRequest{
			Name:      "db_health_check",
			Arguments: map[string]interface{}{},
		}

		response, err := tool.CallTool(context.Background(), request)
		require.NoError(t, err)
		assert.True(t, response.IsError)
		assert.Contains(t, response.Content[0].Text, "health check failed")
	})

	t.Run("CallTool_UnknownTool", func(t *testing.T) {
		mockDB := NewMockMongoDB(true, nil)
		tool := NewDatabaseTool(mockDB)

		request := mcp.ToolCallRequest{
			Name:      "unknown_tool",
			Arguments: map[string]interface{}{},
		}

		response, err := tool.CallTool(context.Background(), request)
		require.NoError(t, err)
		assert.True(t, response.IsError)
		assert.Contains(t, response.Content[0].Text, "Unknown database tool")
	})

	t.Run("CallTool_DatabaseError", func(t *testing.T) {
		mockDB := NewMockMongoDB(true, assert.AnError)
		tool := NewDatabaseTool(mockDB)

		request := mcp.ToolCallRequest{
			Name: "db_create_document",
			Arguments: map[string]interface{}{
				"collection": "test",
				"title":      "Test",
				"content":    "Test content",
			},
		}

		response, err := tool.CallTool(context.Background(), request)
		require.NoError(t, err)
		assert.True(t, response.IsError)
		assert.Contains(t, response.Content[0].Text, "Failed to create document")
	})
}

// Test helper functions
func TestDatabaseTool_Helpers(t *testing.T) {
	mockDB := NewMockMongoDB(true, nil)
	tool := NewDatabaseTool(mockDB)

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

	t.Run("truncateString", func(t *testing.T) {
		longString := "This is a very long string that should be truncated"
		
		result := tool.truncateString(longString, 10)
		assert.Equal(t, "This is a ...", result)
		
		shortString := "Short"
		result = tool.truncateString(shortString, 10)
		assert.Equal(t, "Short", result)
	})

	t.Run("errorResponse", func(t *testing.T) {
		response := tool.errorResponse("test error message")
		assert.True(t, response.IsError)
		assert.Len(t, response.Content, 1)
		assert.Equal(t, "text", response.Content[0].Type)
		assert.Equal(t, "test error message", response.Content[0].Text)
	})
}

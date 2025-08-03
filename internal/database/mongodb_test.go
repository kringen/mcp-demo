package database

import (
	"context"
	"testing"
	"time"

	"github.com/kringen/go-mcp-server/pkg/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMongoDB_Integration runs integration tests against a real MongoDB instance
// Run with: go test -tags=integration
func TestMongoDB_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	// Test configuration
	config := Config{
		URI:            "mongodb://admin:password@localhost:27017",
		Database:       "mcp_server_test",
		ConnectTimeout: 5 * time.Second,
		QueryTimeout:   10 * time.Second,
	}

	// Connect to MongoDB
	db, err := NewMongoDB(config)
	require.NoError(t, err, "Failed to connect to MongoDB")
	defer func() {
		err := db.Close(context.Background())
		assert.NoError(t, err)
	}()

	ctx := context.Background()

	// Test health check
	t.Run("HealthCheck", func(t *testing.T) {
		err := db.HealthCheck(ctx)
		assert.NoError(t, err)
	})

	// Test document operations
	t.Run("DocumentCRUD", func(t *testing.T) {
		collection := "test_documents"

		// Create document
		doc := &mcp.Document{
			Title:   "Test Document",
			Content: "This is a test document for integration testing.",
			Tags:    []string{"test", "integration"},
			Metadata: map[string]interface{}{
				"author": "test_user",
				"type":   "integration_test",
			},
		}

		err := db.CreateDocument(ctx, collection, doc)
		require.NoError(t, err)
		assert.NotEmpty(t, doc.ID)
		assert.False(t, doc.CreatedAt.IsZero())
		assert.False(t, doc.UpdatedAt.IsZero())
		assert.Equal(t, 1, doc.Version)

		// Get document
		retrieved, err := db.GetDocument(ctx, collection, doc.ID)
		require.NoError(t, err)
		assert.Equal(t, doc.ID, retrieved.ID)
		assert.Equal(t, doc.Title, retrieved.Title)
		assert.Equal(t, doc.Content, retrieved.Content)
		assert.Equal(t, doc.Tags, retrieved.Tags)
		assert.Equal(t, doc.Version, retrieved.Version)

		// Update document
		retrieved.Title = "Updated Test Document"
		retrieved.Content = "This document has been updated."
		retrieved.Tags = append(retrieved.Tags, "updated")

		err = db.UpdateDocument(ctx, collection, retrieved)
		require.NoError(t, err)
		assert.Equal(t, 2, retrieved.Version)

		// Verify update
		updated, err := db.GetDocument(ctx, collection, doc.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Test Document", updated.Title)
		assert.Equal(t, "This document has been updated.", updated.Content)
		assert.Contains(t, updated.Tags, "updated")
		assert.Equal(t, 2, updated.Version)

		// Delete document
		err = db.DeleteDocument(ctx, collection, doc.ID)
		require.NoError(t, err)

		// Verify deletion
		_, err = db.GetDocument(ctx, collection, doc.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "document not found")
	})

	// Test query operations
	t.Run("QueryOperations", func(t *testing.T) {
		collection := "test_query_documents"

		// Create test documents
		docs := []*mcp.Document{
			{
				Title:   "Document 1",
				Content: "First test document",
				Tags:    []string{"first", "test"},
			},
			{
				Title:   "Document 2", 
				Content: "Second test document",
				Tags:    []string{"second", "test"},
			},
			{
				Title:   "Document 3",
				Content: "Third test document",
				Tags:    []string{"third", "test"},
			},
		}

		for _, doc := range docs {
			err := db.CreateDocument(ctx, collection, doc)
			require.NoError(t, err)
		}

		// Test query all
		query := mcp.DatabaseQuery{
			Collection: collection,
			Limit:      10,
		}
		results, err := db.QueryDocuments(ctx, query)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 3)

		// Test query with filter
		query.Filter = map[string]interface{}{
			"tags": "first",
		}
		results, err = db.QueryDocuments(ctx, query)
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "Document 1", results[0].Title)

		// Test query with limit
		query.Filter = map[string]interface{}{
			"tags": "test",
		}
		query.Limit = 2
		results, err = db.QueryDocuments(ctx, query)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(results), 2)

		// Test count
		count, err := db.CountDocuments(ctx, collection, map[string]interface{}{
			"tags": "test",
		})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(3))

		// Clean up
		for _, doc := range docs {
			_ = db.DeleteDocument(ctx, collection, doc.ID)
		}
	})

	// Test index creation
	t.Run("IndexCreation", func(t *testing.T) {
		err := db.CreateIndexes(ctx)
		assert.NoError(t, err)
	})
}

// TestMongoDB_Unit contains unit tests that don't require a database
func TestMongoDB_Unit(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		config := DefaultConfig()
		assert.Equal(t, "mongodb://admin:password@localhost:27017", config.URI)
		assert.Equal(t, "mcp_server", config.Database)
		assert.Equal(t, 10*time.Second, config.ConnectTimeout)
		assert.Equal(t, 30*time.Second, config.QueryTimeout)
	})

	t.Run("NewMongoDB_InvalidURI", func(t *testing.T) {
		config := Config{
			URI:            "invalid-uri",
			Database:       "test",
			ConnectTimeout: 1 * time.Second,
			QueryTimeout:   1 * time.Second,
		}

		_, err := NewMongoDB(config)
		assert.Error(t, err)
	})
}

// Benchmark tests
func BenchmarkMongoDB_CreateDocument(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	config := DefaultConfig()
	config.Database = "mcp_server_benchmark"
	
	db, err := NewMongoDB(config)
	if err != nil {
		b.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer db.Close(context.Background())

	ctx := context.Background()
	collection := "benchmark_documents"

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			doc := &mcp.Document{
				Title:   "Benchmark Document",
				Content: "This is a benchmark test document.",
				Tags:    []string{"benchmark", "test"},
			}
			
			err := db.CreateDocument(ctx, collection, doc)
			if err != nil {
				b.Errorf("Failed to create document: %v", err)
			}
		}
	})
}

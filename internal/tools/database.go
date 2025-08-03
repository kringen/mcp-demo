package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/kringen/go-mcp-server/internal/database"
	"github.com/kringen/go-mcp-server/pkg/mcp"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// DatabaseTool provides database operations as MCP tools
type DatabaseTool struct {
	db *database.MongoDB
}

// NewDatabaseTool creates a new DatabaseTool
func NewDatabaseTool(db *database.MongoDB) *DatabaseTool {
	return &DatabaseTool{
		db: db,
	}
}

// ListTools returns the available database tools
func (d *DatabaseTool) ListTools(ctx context.Context) ([]mcp.Tool, error) {
	return []mcp.Tool{
		{
			Name:        "db_create_document",
			Description: "Create a new document in the database",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"collection": map[string]interface{}{
						"type":        "string",
						"description": "Collection name",
					},
					"title": map[string]interface{}{
						"type":        "string",
						"description": "Document title",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "Document content",
					},
					"tags": map[string]interface{}{
						"type":        "array",
						"description": "Document tags",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
					"metadata": map[string]interface{}{
						"type":        "object",
						"description": "Additional metadata",
					},
				},
				"required": []string{"collection", "title", "content"},
			},
		},
		{
			Name:        "db_get_document",
			Description: "Get a document by ID",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"collection": map[string]interface{}{
						"type":        "string",
						"description": "Collection name",
					},
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Document ID",
					},
				},
				"required": []string{"collection", "id"},
			},
		},
		{
			Name:        "db_update_document",
			Description: "Update an existing document",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"collection": map[string]interface{}{
						"type":        "string",
						"description": "Collection name",
					},
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Document ID",
					},
					"title": map[string]interface{}{
						"type":        "string",
						"description": "Document title",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "Document content",
					},
					"tags": map[string]interface{}{
						"type":        "array",
						"description": "Document tags",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
					"metadata": map[string]interface{}{
						"type":        "object",
						"description": "Additional metadata",
					},
				},
				"required": []string{"collection", "id"},
			},
		},
		{
			Name:        "db_delete_document",
			Description: "Delete a document by ID",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"collection": map[string]interface{}{
						"type":        "string",
						"description": "Collection name",
					},
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Document ID",
					},
				},
				"required": []string{"collection", "id"},
			},
		},
		{
			Name:        "db_query_documents",
			Description: "Query documents in a collection",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"collection": map[string]interface{}{
						"type":        "string",
						"description": "Collection name",
					},
					"filter": map[string]interface{}{
						"type":        "object",
						"description": "MongoDB filter query",
					},
					"sort": map[string]interface{}{
						"type":        "object",
						"description": "Sort specification",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of documents to return",
						"minimum":     1,
						"maximum":     100,
					},
					"skip": map[string]interface{}{
						"type":        "integer",
						"description": "Number of documents to skip",
						"minimum":     0,
					},
				},
				"required": []string{"collection"},
			},
		},
		{
			Name:        "db_search_documents",
			Description: "Search documents using text search",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"collection": map[string]interface{}{
						"type":        "string",
						"description": "Collection name",
					},
					"search_text": map[string]interface{}{
						"type":        "string",
						"description": "Text to search for",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of results",
						"minimum":     1,
						"maximum":     50,
					},
				},
				"required": []string{"collection", "search_text"},
			},
		},
		{
			Name:        "db_count_documents",
			Description: "Count documents matching a filter",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"collection": map[string]interface{}{
						"type":        "string",
						"description": "Collection name",
					},
					"filter": map[string]interface{}{
						"type":        "object",
						"description": "MongoDB filter query",
					},
				},
				"required": []string{"collection"},
			},
		},
		{
			Name:        "db_health_check",
			Description: "Check database health",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{},
			},
		},
	}, nil
}

// CallTool executes the specified database tool
func (d *DatabaseTool) CallTool(ctx context.Context, request mcp.ToolCallRequest) (*mcp.ToolCallResponse, error) {
	switch request.Name {
	case "db_create_document":
		return d.createDocument(ctx, request.Arguments)
	case "db_get_document":
		return d.getDocument(ctx, request.Arguments)
	case "db_update_document":
		return d.updateDocument(ctx, request.Arguments)
	case "db_delete_document":
		return d.deleteDocument(ctx, request.Arguments)
	case "db_query_documents":
		return d.queryDocuments(ctx, request.Arguments)
	case "db_search_documents":
		return d.searchDocuments(ctx, request.Arguments)
	case "db_count_documents":
		return d.countDocuments(ctx, request.Arguments)
	case "db_health_check":
		return d.healthCheck(ctx)
	default:
		return &mcp.ToolCallResponse{
			IsError: true,
			Content: []mcp.Content{
				{
					Type: "text",
					Text: fmt.Sprintf("Unknown database tool: %s", request.Name),
				},
			},
		}, nil
	}
}

func (d *DatabaseTool) createDocument(ctx context.Context, args map[string]interface{}) (*mcp.ToolCallResponse, error) {
	collection, ok := args["collection"].(string)
	if !ok || collection == "" {
		return d.errorResponse("Missing or invalid 'collection' parameter"), nil
	}

	title, ok := args["title"].(string)
	if !ok || title == "" {
		return d.errorResponse("Missing or invalid 'title' parameter"), nil
	}

	content, ok := args["content"].(string)
	if !ok || content == "" {
		return d.errorResponse("Missing or invalid 'content' parameter"), nil
	}

	doc := &mcp.Document{
		ID:      bson.NewObjectID().Hex(),
		Title:   title,
		Content: content,
	}

	// Extract optional tags
	if tagsInterface, ok := args["tags"]; ok {
		if tagsSlice, ok := tagsInterface.([]interface{}); ok {
			tags := make([]string, len(tagsSlice))
			for i, tag := range tagsSlice {
				if tagStr, ok := tag.(string); ok {
					tags[i] = tagStr
				}
			}
			doc.Tags = tags
		}
	}

	// Extract optional metadata
	if metadata, ok := args["metadata"].(map[string]interface{}); ok {
		doc.Metadata = metadata
	}

	err := d.db.CreateDocument(ctx, collection, doc)
	if err != nil {
		return d.errorResponse(fmt.Sprintf("Failed to create document: %v", err)), nil
	}

	return &mcp.ToolCallResponse{
		Content: []mcp.Content{
			{
				Type: "text",
				Text: fmt.Sprintf("Document created successfully with ID: %s", doc.ID),
			},
			{
				Type: "text",
				Text: fmt.Sprintf("Document details:\n- Title: %s\n- Collection: %s\n- Created: %s",
					doc.Title, collection, doc.CreatedAt.Format(time.RFC3339)),
			},
		},
	}, nil
}

func (d *DatabaseTool) getDocument(ctx context.Context, args map[string]interface{}) (*mcp.ToolCallResponse, error) {
	collection, ok := args["collection"].(string)
	if !ok || collection == "" {
		return d.errorResponse("Missing or invalid 'collection' parameter"), nil
	}

	id, ok := args["id"].(string)
	if !ok || id == "" {
		return d.errorResponse("Missing or invalid 'id' parameter"), nil
	}

	doc, err := d.db.GetDocument(ctx, collection, id)
	if err != nil {
		return d.errorResponse(fmt.Sprintf("Failed to get document: %v", err)), nil
	}

	// Format document for display
	content := []mcp.Content{
		{
			Type: "text",
			Text: fmt.Sprintf("Document found:\n- ID: %s\n- Title: %s\n- Created: %s\n- Updated: %s\n- Version: %d",
				doc.ID, doc.Title, doc.CreatedAt.Format(time.RFC3339), doc.UpdatedAt.Format(time.RFC3339), doc.Version),
		},
		{
			Type: "text",
			Text: fmt.Sprintf("Content:\n%s", doc.Content),
		},
	}

	if len(doc.Tags) > 0 {
		content = append(content, mcp.Content{
			Type: "text",
			Text: fmt.Sprintf("Tags: %v", doc.Tags),
		})
	}

	// Add JSON data
	jsonData, _ := json.Marshal(doc)
	content = append(content, mcp.Content{
		Type: "text",
		Text: fmt.Sprintf("Raw JSON:\n```json\n%s\n```", string(jsonData)),
	})

	return &mcp.ToolCallResponse{
		Content: content,
	}, nil
}

func (d *DatabaseTool) updateDocument(ctx context.Context, args map[string]interface{}) (*mcp.ToolCallResponse, error) {
	collection, ok := args["collection"].(string)
	if !ok || collection == "" {
		return d.errorResponse("Missing or invalid 'collection' parameter"), nil
	}

	id, ok := args["id"].(string)
	if !ok || id == "" {
		return d.errorResponse("Missing or invalid 'id' parameter"), nil
	}

	// Get existing document
	doc, err := d.db.GetDocument(ctx, collection, id)
	if err != nil {
		return d.errorResponse(fmt.Sprintf("Failed to find document: %v", err)), nil
	}

	// Update fields if provided
	if title, ok := args["title"].(string); ok && title != "" {
		doc.Title = title
	}

	if content, ok := args["content"].(string); ok && content != "" {
		doc.Content = content
	}

	if tagsInterface, ok := args["tags"]; ok {
		if tagsSlice, ok := tagsInterface.([]interface{}); ok {
			tags := make([]string, len(tagsSlice))
			for i, tag := range tagsSlice {
				if tagStr, ok := tag.(string); ok {
					tags[i] = tagStr
				}
			}
			doc.Tags = tags
		}
	}

	if metadata, ok := args["metadata"].(map[string]interface{}); ok {
		doc.Metadata = metadata
	}

	err = d.db.UpdateDocument(ctx, collection, doc)
	if err != nil {
		return d.errorResponse(fmt.Sprintf("Failed to update document: %v", err)), nil
	}

	return &mcp.ToolCallResponse{
		Content: []mcp.Content{
			{
				Type: "text",
				Text: fmt.Sprintf("Document updated successfully. New version: %d", doc.Version),
			},
		},
	}, nil
}

func (d *DatabaseTool) deleteDocument(ctx context.Context, args map[string]interface{}) (*mcp.ToolCallResponse, error) {
	collection, ok := args["collection"].(string)
	if !ok || collection == "" {
		return d.errorResponse("Missing or invalid 'collection' parameter"), nil
	}

	id, ok := args["id"].(string)
	if !ok || id == "" {
		return d.errorResponse("Missing or invalid 'id' parameter"), nil
	}

	err := d.db.DeleteDocument(ctx, collection, id)
	if err != nil {
		return d.errorResponse(fmt.Sprintf("Failed to delete document: %v", err)), nil
	}

	return &mcp.ToolCallResponse{
		Content: []mcp.Content{
			{
				Type: "text",
				Text: fmt.Sprintf("Document with ID %s deleted successfully from collection %s", id, collection),
			},
		},
	}, nil
}

func (d *DatabaseTool) queryDocuments(ctx context.Context, args map[string]interface{}) (*mcp.ToolCallResponse, error) {
	collection, ok := args["collection"].(string)
	if !ok || collection == "" {
		return d.errorResponse("Missing or invalid 'collection' parameter"), nil
	}

	query := mcp.DatabaseQuery{
		Collection: collection,
		Limit:      10, // default
	}

	if filter, ok := args["filter"].(map[string]interface{}); ok {
		query.Filter = filter
	}

	if sort, ok := args["sort"].(map[string]interface{}); ok {
		query.Sort = sort
	}

	if limit, ok := args["limit"]; ok {
		if l, err := d.toInt(limit); err == nil && l > 0 && l <= 100 {
			query.Limit = l
		}
	}

	if skip, ok := args["skip"]; ok {
		if s, err := d.toInt(skip); err == nil && s >= 0 {
			query.Skip = s
		}
	}

	docs, err := d.db.QueryDocuments(ctx, query)
	if err != nil {
		return d.errorResponse(fmt.Sprintf("Query failed: %v", err)), nil
	}

	content := []mcp.Content{
		{
			Type: "text",
			Text: fmt.Sprintf("Found %d documents in collection '%s'", len(docs), collection),
		},
	}

	for i, doc := range docs {
		content = append(content, mcp.Content{
			Type: "text",
			Text: fmt.Sprintf("%d. **%s** (ID: %s)\n   Created: %s\n   Content preview: %s...",
				i+1, doc.Title, doc.ID, doc.CreatedAt.Format(time.RFC3339),
				d.truncateString(doc.Content, 100)),
		})
	}

	// Add JSON data
	jsonData, _ := json.Marshal(docs)
	content = append(content, mcp.Content{
		Type: "text",
		Text: fmt.Sprintf("Raw JSON:\n```json\n%s\n```", string(jsonData)),
	})

	return &mcp.ToolCallResponse{
		Content: content,
	}, nil
}

func (d *DatabaseTool) searchDocuments(ctx context.Context, args map[string]interface{}) (*mcp.ToolCallResponse, error) {
	collection, ok := args["collection"].(string)
	if !ok || collection == "" {
		return d.errorResponse("Missing or invalid 'collection' parameter"), nil
	}

	searchText, ok := args["search_text"].(string)
	if !ok || searchText == "" {
		return d.errorResponse("Missing or invalid 'search_text' parameter"), nil
	}

	limit := 10 // default
	if l, ok := args["limit"]; ok {
		if parsed, err := d.toInt(l); err == nil && parsed > 0 && parsed <= 50 {
			limit = parsed
		}
	}

	docs, err := d.db.SearchDocuments(ctx, collection, searchText, limit)
	if err != nil {
		return d.errorResponse(fmt.Sprintf("Search failed: %v", err)), nil
	}

	content := []mcp.Content{
		{
			Type: "text",
			Text: fmt.Sprintf("Found %d documents matching '%s' in collection '%s'", len(docs), searchText, collection),
		},
	}

	for i, doc := range docs {
		content = append(content, mcp.Content{
			Type: "text",
			Text: fmt.Sprintf("%d. **%s** (ID: %s)\n   Created: %s\n   Content preview: %s...",
				i+1, doc.Title, doc.ID, doc.CreatedAt.Format(time.RFC3339),
				d.truncateString(doc.Content, 100)),
		})
	}

	return &mcp.ToolCallResponse{
		Content: content,
	}, nil
}

func (d *DatabaseTool) countDocuments(ctx context.Context, args map[string]interface{}) (*mcp.ToolCallResponse, error) {
	collection, ok := args["collection"].(string)
	if !ok || collection == "" {
		return d.errorResponse("Missing or invalid 'collection' parameter"), nil
	}

	filter := make(map[string]interface{})
	if f, ok := args["filter"].(map[string]interface{}); ok {
		filter = f
	}

	count, err := d.db.CountDocuments(ctx, collection, filter)
	if err != nil {
		return d.errorResponse(fmt.Sprintf("Count failed: %v", err)), nil
	}

	return &mcp.ToolCallResponse{
		Content: []mcp.Content{
			{
				Type: "text",
				Text: fmt.Sprintf("Collection '%s' contains %d documents matching the filter", collection, count),
			},
		},
	}, nil
}

func (d *DatabaseTool) healthCheck(ctx context.Context) (*mcp.ToolCallResponse, error) {
	err := d.db.HealthCheck(ctx)
	if err != nil {
		return &mcp.ToolCallResponse{
			IsError: true,
			Content: []mcp.Content{
				{
					Type: "text",
					Text: fmt.Sprintf("Database health check failed: %v", err),
				},
			},
		}, nil
	}

	return &mcp.ToolCallResponse{
		Content: []mcp.Content{
			{
				Type: "text",
				Text: "Database is healthy and ready to accept operations.",
			},
		},
	}, nil
}

// Helper methods

func (d *DatabaseTool) toInt(value interface{}) (int, error) {
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

func (d *DatabaseTool) truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func (d *DatabaseTool) errorResponse(message string) *mcp.ToolCallResponse {
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

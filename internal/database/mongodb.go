package database

import (
	"context"
	"fmt"
	"time"

	"github.com/kringen/go-mcp-server/pkg/mcp"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// MongoDB implements database operations
type MongoDB struct {
	client   *mongo.Client
	database *mongo.Database
	config   Config
}

// Config holds MongoDB configuration
type Config struct {
	URI            string        `json:"uri"`
	Database       string        `json:"database"`
	ConnectTimeout time.Duration `json:"connect_timeout"`
	QueryTimeout   time.Duration `json:"query_timeout"`
}

// DefaultConfig returns a default MongoDB configuration
func DefaultConfig() Config {
	return Config{
		URI:            "mongodb://admin:password@localhost:27017",
		Database:       "mcp_server",
		ConnectTimeout: 10 * time.Second,
		QueryTimeout:   30 * time.Second,
	}
}

// NewMongoDB creates a new MongoDB client
func NewMongoDB(config Config) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), config.ConnectTimeout)
	defer cancel()

	clientOptions := options.Client().ApplyURI(config.URI)
	client, err := mongo.Connect(clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Test the connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	database := client.Database(config.Database)

	return &MongoDB{
		client:   client,
		database: database,
		config:   config,
	}, nil
}

// Close closes the MongoDB connection
func (m *MongoDB) Close(ctx context.Context) error {
	return m.client.Disconnect(ctx)
}

// CreateDocument creates a new document in the specified collection
func (m *MongoDB) CreateDocument(ctx context.Context, collection string, doc *mcp.Document) error {
	ctx, cancel := context.WithTimeout(ctx, m.config.QueryTimeout)
	defer cancel()

	if doc.ID == "" {
		doc.ID = bson.NewObjectID().Hex()
	}
	doc.CreatedAt = time.Now()
	doc.UpdatedAt = time.Now()
	doc.Version = 1

	coll := m.database.Collection(collection)
	_, err := coll.InsertOne(ctx, doc)
	if err != nil {
		return fmt.Errorf("failed to create document: %w", err)
	}

	return nil
}

// GetDocument retrieves a document by ID
func (m *MongoDB) GetDocument(ctx context.Context, collection, id string) (*mcp.Document, error) {
	ctx, cancel := context.WithTimeout(ctx, m.config.QueryTimeout)
	defer cancel()

	coll := m.database.Collection(collection)
	
	var doc mcp.Document
	err := coll.FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("document not found")
		}
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	return &doc, nil
}

// UpdateDocument updates an existing document
func (m *MongoDB) UpdateDocument(ctx context.Context, collection string, doc *mcp.Document) error {
	ctx, cancel := context.WithTimeout(ctx, m.config.QueryTimeout)
	defer cancel()

	doc.UpdatedAt = time.Now()
	doc.Version++

	coll := m.database.Collection(collection)
	update := bson.M{
		"$set": bson.M{
			"title":      doc.Title,
			"content":    doc.Content,
			"tags":       doc.Tags,
			"metadata":   doc.Metadata,
			"updated_at": doc.UpdatedAt,
			"version":    doc.Version,
		},
	}

	result, err := coll.UpdateOne(ctx, bson.M{"_id": doc.ID}, update)
	if err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("document not found")
	}

	return nil
}

// DeleteDocument deletes a document by ID
func (m *MongoDB) DeleteDocument(ctx context.Context, collection, id string) error {
	ctx, cancel := context.WithTimeout(ctx, m.config.QueryTimeout)
	defer cancel()

	coll := m.database.Collection(collection)
	result, err := coll.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("document not found")
	}

	return nil
}

// QueryDocuments performs a query on the specified collection
func (m *MongoDB) QueryDocuments(ctx context.Context, query mcp.DatabaseQuery) ([]*mcp.Document, error) {
	ctx, cancel := context.WithTimeout(ctx, m.config.QueryTimeout)
	defer cancel()

	coll := m.database.Collection(query.Collection)

	// Build find options
	findOptions := options.Find()
	if query.Limit > 0 {
		findOptions.SetLimit(int64(query.Limit))
	}
	if query.Skip > 0 {
		findOptions.SetSkip(int64(query.Skip))
	}
	if query.Sort != nil && len(query.Sort) > 0 {
		findOptions.SetSort(query.Sort)
	}

	// Build filter
	filter := bson.M{}
	if query.Filter != nil {
		filter = query.Filter
	}

	cursor, err := coll.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer cursor.Close(ctx)

	var documents []*mcp.Document
	for cursor.Next(ctx) {
		// Decode to bson.M first to handle ObjectID properly
		var rawDoc bson.M
		if err := cursor.Decode(&rawDoc); err != nil {
			return nil, fmt.Errorf("failed to decode raw document: %w", err)
		}

		// Convert to Document struct with ObjectID handling
		doc, err := m.convertToDocument(rawDoc)
		if err != nil {
			return nil, fmt.Errorf("failed to convert document: %w", err)
		}
		
		documents = append(documents, doc)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return documents, nil
}

// SearchDocuments performs a text search on documents
func (m *MongoDB) SearchDocuments(ctx context.Context, collection, searchText string, limit int) ([]*mcp.Document, error) {
	ctx, cancel := context.WithTimeout(ctx, m.config.QueryTimeout)
	defer cancel()

	coll := m.database.Collection(collection)

	// Create text search filter
	filter := bson.M{
		"$text": bson.M{
			"$search": searchText,
		},
	}

	findOptions := options.Find()
	if limit > 0 {
		findOptions.SetLimit(int64(limit))
	}
	// Sort by text score
	findOptions.SetSort(bson.M{"score": bson.M{"$meta": "textScore"}})

	cursor, err := coll.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}
	defer cursor.Close(ctx)

	var documents []*mcp.Document
	for cursor.Next(ctx) {
		// Decode to bson.M first to handle ObjectID properly
		var rawDoc bson.M
		if err := cursor.Decode(&rawDoc); err != nil {
			return nil, fmt.Errorf("failed to decode raw document: %w", err)
		}

		// Convert to Document struct with ObjectID handling
		doc, err := m.convertToDocument(rawDoc)
		if err != nil {
			return nil, fmt.Errorf("failed to convert document: %w", err)
		}
		
		documents = append(documents, doc)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return documents, nil
}

// CountDocuments counts documents matching the filter
func (m *MongoDB) CountDocuments(ctx context.Context, collection string, filter map[string]interface{}) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, m.config.QueryTimeout)
	defer cancel()

	coll := m.database.Collection(collection)
	
	if filter == nil {
		filter = bson.M{}
	}

	count, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count documents: %w", err)
	}

	return count, nil
}

// CreateIndexes creates indexes for better performance
func (m *MongoDB) CreateIndexes(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, m.config.QueryTimeout)
	defer cancel()

	// Create text indexes for search
	collections := []string{"documents", "search_cache"}
	
	for _, collName := range collections {
		coll := m.database.Collection(collName)
		
		// Text index for search
		textIndex := mongo.IndexModel{
			Keys: bson.M{
				"title":   "text",
				"content": "text",
			},
		}
		
		// Other useful indexes
		indexes := []mongo.IndexModel{
			textIndex,
			{Keys: bson.M{"created_at": -1}},
			{Keys: bson.M{"updated_at": -1}},
			{Keys: bson.M{"tags": 1}},
		}
		
		if collName == "search_cache" {
			// TTL index for search cache expiration
			ttlIndex := mongo.IndexModel{
				Keys:    bson.M{"timestamp": 1},
				Options: options.Index().SetExpireAfterSeconds(3600), // 1 hour
			}
			indexes = append(indexes, ttlIndex)
		}
		
		_, err := coll.Indexes().CreateMany(ctx, indexes)
		if err != nil {
			return fmt.Errorf("failed to create indexes for %s: %w", collName, err)
		}
	}

	return nil
}

// HealthCheck performs a health check on the database
func (m *MongoDB) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return m.client.Ping(ctx, nil)
}

// convertToDocument converts a bson.M to a Document struct with proper ObjectID handling
func (m *MongoDB) convertToDocument(rawDoc bson.M) (*mcp.Document, error) {
	doc := &mcp.Document{}
	
	// Handle _id field (ObjectID)
	if id, ok := rawDoc["_id"]; ok {
		switch v := id.(type) {
		case bson.ObjectID:
			doc.ID = v.Hex()
		case string:
			doc.ID = v
		default:
			doc.ID = fmt.Sprintf("%v", v)
		}
	}
	
	// Handle other fields
	if title, ok := rawDoc["title"].(string); ok {
		doc.Title = title
	}
	if content, ok := rawDoc["content"].(string); ok {
		doc.Content = content
	}
	if category, ok := rawDoc["category"].(string); ok {
		doc.Category = category
	}
	if tags, ok := rawDoc["tags"].(bson.A); ok {
		for _, tag := range tags {
			if tagStr, ok := tag.(string); ok {
				doc.Tags = append(doc.Tags, tagStr)
			}
		}
	}
	if metadata, ok := rawDoc["metadata"].(bson.M); ok {
		doc.Metadata = make(map[string]interface{})
		for k, v := range metadata {
			doc.Metadata[k] = v
		}
	}
	if createdAt, ok := rawDoc["created_at"].(time.Time); ok {
		doc.CreatedAt = createdAt
	}
	if updatedAt, ok := rawDoc["updated_at"].(time.Time); ok {
		doc.UpdatedAt = updatedAt
	}
	if version, ok := rawDoc["version"].(int32); ok {
		doc.Version = int(version)
	}
	
	return doc, nil
}

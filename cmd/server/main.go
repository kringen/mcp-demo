package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kringen/go-mcp-server/internal/database"
	"github.com/kringen/go-mcp-server/internal/search"
	"github.com/kringen/go-mcp-server/internal/server"
	"github.com/kringen/go-mcp-server/internal/tools"
)

func main() {
	// Get default values from environment variables
	defaultAddr := os.Getenv("ADDR")
	if defaultAddr == "" {
		defaultAddr = "localhost:8080"
	}
	
	defaultMongoURI := os.Getenv("MONGO_URI")
	if defaultMongoURI == "" {
		defaultMongoURI = "mongodb://admin:password@localhost:27017"
	}
	
	defaultDBName := os.Getenv("DB_NAME")
	if defaultDBName == "" {
		defaultDBName = "mcp_server"
	}
	
	defaultDebug := os.Getenv("DEBUG") == "true"

	// Command line flags
	var (
		addr       = flag.String("addr", defaultAddr, "Server address")
		mongoURI   = flag.String("mongo-uri", defaultMongoURI, "MongoDB connection URI")
		dbName     = flag.String("db-name", defaultDBName, "MongoDB database name")
		debug      = flag.Bool("debug", defaultDebug, "Enable debug mode")
	)
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize MongoDB
	log.Println("Connecting to MongoDB...")
	dbConfig := database.Config{
		URI:            *mongoURI,
		Database:       *dbName,
		ConnectTimeout: 10 * time.Second,
		QueryTimeout:   30 * time.Second,
	}

	db, err := database.NewMongoDB(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err := db.Close(context.Background()); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	// Create indexes for better performance
	log.Println("Creating database indexes...")
	if err := db.CreateIndexes(ctx); err != nil {
		log.Printf("Warning: Failed to create indexes: %v", err)
	}

	// Initialize web searcher
	log.Println("Initializing web search service...")
	searchConfig := search.DefaultConfig()
	searchConfig.EnableDebug = *debug
	searcher := search.NewCollySearcher(searchConfig)

	// Create and configure the MCP server
	log.Println("Creating MCP server...")
	mcpServer := server.NewMCPServer()
	
	// Add tool providers
	mcpServer.RegisterToolProvider(tools.NewMathToolProvider())
	mcpServer.RegisterToolProvider(tools.NewSearchTool(searcher))
	mcpServer.RegisterToolProvider(tools.NewDatabaseTool(db))
	
	// Start the server
	log.Printf("Starting MCP server on %s...", *addr)
	go func() {
		if err := mcpServer.Start(ctx, *addr); err != nil {
			log.Printf("Server error: %v", err)
			cancel()
		}
	}()

	// Test connectivity
	log.Println("Testing service connectivity...")
	
	// Test database
	if err := db.HealthCheck(ctx); err != nil {
		log.Printf("Warning: Database health check failed: %v", err)
	} else {
		log.Println("✓ Database connection is healthy")
	}

	// Test search service
	if err := searcher.HealthCheck(ctx); err != nil {
		log.Printf("Warning: Search service health check failed: %v", err)
	} else {
		log.Println("✓ Search service is healthy")
	}

	log.Printf("✓ MCP Server is ready!")
	log.Printf("  - WebSocket endpoint: ws://%s/mcp", *addr)
	log.Printf("  - Health check: http://%s/health", *addr)
	log.Printf("  - Web interface: http://%s/", *addr)
	log.Println()
	log.Println("Available tools:")
	log.Println("  Math: add, multiply, divide, power")
	log.Println("  Search: web_search, search_health_check")
	log.Println("  Database: db_create_document, db_get_document, db_update_document,")
	log.Println("           db_delete_document, db_query_documents, db_search_documents,")
	log.Println("           db_count_documents, db_health_check")
	log.Println()
	log.Println("To start MongoDB: make mongo-up")
	log.Println("To stop the server: Ctrl+C")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		log.Printf("Received signal %v, shutting down...", sig)
	case <-ctx.Done():
		log.Println("Context cancelled, shutting down...")
	}

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := mcpServer.Stop(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}

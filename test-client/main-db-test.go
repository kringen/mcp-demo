package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"

	"github.com/gorilla/websocket"
)

func main() {
	// Parse command line flags
	var (
		host   = flag.String("host", "localhost", "Server host/IP address")
		port   = flag.String("port", "8080", "Server port")
		help   = flag.Bool("help", false, "Show usage information")
		dbTest = flag.Bool("db-test", false, "Run comprehensive database tests with knowledge base")
	)
	flag.Parse()

	if *help {
		fmt.Println("🧪 MCP WebSocket Test Client")
		fmt.Println("==============================")
		fmt.Println("Usage:")
		fmt.Println("  go run main.go [options]")
		fmt.Println("")
		fmt.Println("Options:")
		fmt.Println("  -host string    Server host/IP address (default: localhost)")
		fmt.Println("  -port string    Server port (default: 8080)")
		fmt.Println("  -db-test       Run comprehensive database tests (default: false)")
		fmt.Println("  -help          Show this help message")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  go run main.go                                    # Basic test")
		fmt.Println("  go run main.go -db-test                          # Full database test")
		fmt.Println("  go run main.go -host 192.168.1.49 -port 80      # LoadBalancer test")
		fmt.Println("  go run main.go -host example.com -port 443       # Remote server")
		return
	}

	fmt.Println("🧪 MCP WebSocket Test Client")
	fmt.Println("==============================")

	// Connect to WebSocket
	u := url.URL{Scheme: "ws", Host: *host + ":" + *port, Path: "/mcp"}
	fmt.Printf("🔌 Connecting to %s\n", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("❌ Connection failed:", err)
	}
	defer c.Close()

	fmt.Println("✅ Connected to WebSocket!")

	// Test 1: Initialize
	fmt.Println("\n📤 Sending initialize message...")
	initMsg := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{"tools": map[string]interface{}{}},
			"clientInfo":      map[string]interface{}{"name": "test-client", "version": "1.0.0"},
		},
	}

	if err := c.WriteJSON(initMsg); err != nil {
		log.Fatal("Write error:", err)
	}

	var response map[string]interface{}
	if err := c.ReadJSON(&response); err != nil {
		log.Fatal("Read error:", err)
	}

	respBytes, _ := json.MarshalIndent(response, "", "  ")
	fmt.Printf("📥 Response: %s\n", respBytes)

	// Send initialized notification
	fmt.Println("\n📤 Sending initialized notification...")
	initializedMsg := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "initialized",
	}

	if err := c.WriteJSON(initializedMsg); err != nil {
		log.Fatal("Write error:", err)
	}

	// Test 2: List tools
	fmt.Println("\n📤 Requesting tools list...")
	toolsMsg := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/list",
	}

	if err := c.WriteJSON(toolsMsg); err != nil {
		log.Fatal("Write error:", err)
	}

	if err := c.ReadJSON(&response); err != nil {
		log.Fatal("Read error:", err)
	}

	fmt.Println("📥 Available tools:")
	if result, ok := response["result"].(map[string]interface{}); ok {
		if tools, ok := result["tools"].([]interface{}); ok {
			for i, tool := range tools {
				if toolMap, ok := tool.(map[string]interface{}); ok {
					if name, ok := toolMap["name"].(string); ok {
						fmt.Printf("   %d. %s\n", i+1, name)
					}
				}
			}
		}
	}

	// Test 3: Math tool
	fmt.Println("\n📤 Testing math tool (add 5 + 3)...")
	mathMsg := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      3,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name":      "add",
			"arguments": map[string]interface{}{"a": 5, "b": 3},
		},
	}

	if err := c.WriteJSON(mathMsg); err != nil {
		log.Fatal("Write error:", err)
	}

	if err := c.ReadJSON(&response); err != nil {
		log.Fatal("Read error:", err)
	}

	if result, ok := response["result"].(map[string]interface{}); ok {
		if content, ok := result["content"].([]interface{}); ok && len(content) > 0 {
			if textContent, ok := content[0].(map[string]interface{}); ok {
				if text, ok := textContent["text"].(string); ok {
					fmt.Printf("📥 Math result: %s\n", text)
				}
			}
		}
	}

	// Test 4: Database health check
	fmt.Println("\n📤 Testing database health check...")
	dbMsg := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      4,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name": "db_health_check",
		},
	}

	if err := c.WriteJSON(dbMsg); err != nil {
		log.Fatal("Write error:", err)
	}

	if err := c.ReadJSON(&response); err != nil {
		log.Fatal("Read error:", err)
	}

	if result, ok := response["result"].(map[string]interface{}); ok {
		if content, ok := result["content"].([]interface{}); ok && len(content) > 0 {
			if textContent, ok := content[0].(map[string]interface{}); ok {
				if text, ok := textContent["text"].(string); ok {
					fmt.Printf("📥 Database health: %s\n", text)
				}
			}
		}
	}

	// Comprehensive database tests if requested
	if *dbTest {
		fmt.Println("\n🗃️  Running comprehensive database tests...")
		runDatabaseTests(c)
	}

	fmt.Println("\n🎉 All tests completed successfully!")
}

func runDatabaseTests(c *websocket.Conn) {
	var response map[string]interface{}
	msgID := 5

	// Test 1: Count all documents
	fmt.Println("\n📊 Counting all knowledge base articles...")
	countMsg := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      msgID,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name": "db_count_documents",
			"arguments": map[string]interface{}{
				"collection": "knowledgebase",
				"filter":     map[string]interface{}{},
			},
		},
	}
	msgID++

	if err := c.WriteJSON(countMsg); err != nil {
		log.Printf("Write error: %v", err)
		return
	}

	if err := c.ReadJSON(&response); err != nil {
		log.Printf("Read error: %v", err)
		return
	}

	if result, ok := response["result"].(map[string]interface{}); ok {
		if content, ok := result["content"].([]interface{}); ok && len(content) > 0 {
			if textContent, ok := content[0].(map[string]interface{}); ok {
				if text, ok := textContent["text"].(string); ok {
					fmt.Printf("📥 Total articles: %s\n", text)
				}
			}
		}
	}

	// Test 2: Search for Kubernetes articles
	fmt.Println("\n🔍 Searching for 'kubernetes' articles...")
	searchMsg := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      msgID,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name": "db_search_documents",
			"arguments": map[string]interface{}{
				"collection":  "knowledgebase",
				"search_text": "kubernetes",
				"limit":       3,
			},
		},
	}
	msgID++

	if err := c.WriteJSON(searchMsg); err != nil {
		log.Printf("Write error: %v", err)
		return
	}

	if err := c.ReadJSON(&response); err != nil {
		log.Printf("Read error: %v", err)
		return
	}

	fmt.Println("📥 Kubernetes search results:")
	printSearchResults(response)

	// Test 3: Query by category
	fmt.Println("\n📂 Querying 'Security' category articles...")
	queryMsg := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      msgID,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name": "db_query_documents",
			"arguments": map[string]interface{}{
				"collection": "knowledgebase",
				"filter":     map[string]interface{}{"category": "Security"},
				"limit":      3,
			},
		},
	}
	msgID++

	if err := c.WriteJSON(queryMsg); err != nil {
		log.Printf("Write error: %v", err)
		return
	}

	if err := c.ReadJSON(&response); err != nil {
		log.Printf("Read error: %v", err)
		return
	}

	fmt.Println("📥 Security category results:")
	printQueryResults(response)

	// Test 4: Search for database-related articles
	fmt.Println("\n🗄️  Searching for 'database' articles...")
	dbSearchMsg := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      msgID,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name": "db_search_documents",
			"arguments": map[string]interface{}{
				"collection":  "knowledgebase",
				"search_text": "database connection",
				"limit":       2,
			},
		},
	}
	msgID++

	if err := c.WriteJSON(dbSearchMsg); err != nil {
		log.Printf("Write error: %v", err)
		return
	}

	if err := c.ReadJSON(&response); err != nil {
		log.Printf("Read error: %v", err)
		return
	}

	fmt.Println("📥 Database search results:")
	printSearchResults(response)

	// Test 5: Count high priority articles
	fmt.Println("\n⚠️  Counting high priority articles...")
	highPriorityMsg := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      msgID,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name": "db_count_documents",
			"arguments": map[string]interface{}{
				"collection": "knowledgebase",
				"filter":     map[string]interface{}{"priority": "high"},
			},
		},
	}

	if err := c.WriteJSON(highPriorityMsg); err != nil {
		log.Printf("Write error: %v", err)
		return
	}

	if err := c.ReadJSON(&response); err != nil {
		log.Printf("Read error: %v", err)
		return
	}

	if result, ok := response["result"].(map[string]interface{}); ok {
		if content, ok := result["content"].([]interface{}); ok && len(content) > 0 {
			if textContent, ok := content[0].(map[string]interface{}); ok {
				if text, ok := textContent["text"].(string); ok {
					fmt.Printf("📥 High priority articles: %s\n", text)
				}
			}
		}
	}
}

func printSearchResults(response map[string]interface{}) {
	if result, ok := response["result"].(map[string]interface{}); ok {
		if content, ok := result["content"].([]interface{}); ok && len(content) > 0 {
			if textContent, ok := content[0].(map[string]interface{}); ok {
				if text, ok := textContent["text"].(string); ok {
					var docs []map[string]interface{}
					if err := json.Unmarshal([]byte(text), &docs); err == nil {
						for i, doc := range docs {
							if title, ok := doc["title"].(string); ok {
								if category, ok := doc["category"].(string); ok {
									fmt.Printf("   %d. [%s] %s\n", i+1, category, title)
								}
							}
						}
					} else {
						fmt.Printf("   %s\n", text)
					}
				}
			}
		}
	}
}

func printQueryResults(response map[string]interface{}) {
	printSearchResults(response) // Same format for query results
}

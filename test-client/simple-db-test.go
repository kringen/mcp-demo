package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"

	"github.com/gorilla/websocket"
)

func main() {
	fmt.Println("🧪 Simple Database Test")
	fmt.Println("=====================")

	// Connect to WebSocket
	u := url.URL{Scheme: "ws", Host: "192.168.1.49:80", Path: "/mcp"}
	fmt.Printf("🔌 Connecting to %s\n", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("❌ Connection failed:", err)
	}
	defer c.Close()

	fmt.Println("✅ Connected to WebSocket!")

	// Initialize
	initMsg := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{"tools": map[string]interface{}{}},
			"clientInfo":      map[string]interface{}{"name": "simple-test", "version": "1.0.0"},
		},
	}

	if err := c.WriteJSON(initMsg); err != nil {
		log.Fatal("Write error:", err)
	}

	var response map[string]interface{}
	if err := c.ReadJSON(&response); err != nil {
		log.Fatal("Read error:", err)
	}

	// Send initialized notification
	initializedMsg := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "initialized",
	}

	if err := c.WriteJSON(initializedMsg); err != nil {
		log.Fatal("Write error:", err)
	}

	// Test 1: Count documents
	fmt.Println("\n📊 Counting knowledge base documents...")
	countMsg := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name": "db_count_documents",
			"arguments": map[string]interface{}{
				"collection": "knowledgebase",
				"filter":     map[string]interface{}{},
			},
		},
	}

	if err := c.WriteJSON(countMsg); err != nil {
		log.Fatal("Write error:", err)
	}

	if err := c.ReadJSON(&response); err != nil {
		log.Fatal("Read error:", err)
	}

	respBytes, _ := json.MarshalIndent(response, "", "  ")
	fmt.Printf("📥 Count Response: %s\n", respBytes)

	// Test 2: Query by category
	fmt.Println("\n📂 Querying Security category...")
	queryMsg := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      3,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name": "db_query_documents",
			"arguments": map[string]interface{}{
				"collection": "knowledgebase",
				"filter":     map[string]interface{}{"category": "Security"},
				"limit":      2,
			},
		},
	}

	if err := c.WriteJSON(queryMsg); err != nil {
		log.Fatal("Write error:", err)
	}

	if err := c.ReadJSON(&response); err != nil {
		log.Fatal("Read error:", err)
	}

	respBytes, _ = json.MarshalIndent(response, "", "  ")
	fmt.Printf("📥 Query Response: %s\n", respBytes)

	fmt.Println("\n🎉 Tests completed!")
}

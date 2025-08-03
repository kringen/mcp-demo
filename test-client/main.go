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
		host = flag.String("host", "localhost", "Server host/IP address")
		port = flag.String("port", "8080", "Server port")
		help = flag.Bool("help", false, "Show usage information")
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
		fmt.Println("  -help          Show this help message")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  go run main.go                                    # Test localhost:8080")
		fmt.Println("  go run main.go -host 192.168.1.92 -port 30800   # Test NodePort")
		fmt.Println("  go run main.go -host example.com -port 443       # Test remote server")
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

	respBytes, _ = json.MarshalIndent(response, "", "  ")
	fmt.Printf("📥 Response: %s\n", respBytes)

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

	respBytes, _ = json.MarshalIndent(response, "", "  ")
	fmt.Printf("📥 Response: %s\n", respBytes)

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

	respBytes, _ = json.MarshalIndent(response, "", "  ")
	fmt.Printf("📥 Response: %s\n", respBytes)

	fmt.Println("\n🎉 All tests completed successfully!")
}

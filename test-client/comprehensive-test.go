package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type MCPRequest struct {
	ID      int64       `json:"id"`
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type MCPResponse struct {
	ID      int64           `json:"id"`
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   interface{}     `json:"error,omitempty"`
}

type ToolCallParams struct {
	Name      string            `json:"name"`
	Arguments map[string]string `json:"arguments"`
}

func main() {
	fmt.Println("🧪 Comprehensive MCP Database Test")
	fmt.Println("===================================")

	// Connect to LoadBalancer endpoint
	url := "ws://192.168.1.49:80/mcp"
	fmt.Printf("🔌 Connecting to %s\n", url)

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("❌ Connection failed:", err)
	}
	defer conn.Close()
	fmt.Println("✅ Connected to LoadBalancer WebSocket!")

	var messageID int64 = 1

	// Initialize connection
	initMsg := MCPRequest{
		ID:      messageID,
		JSONRPC: "2.0",
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "Test Client",
				"version": "1.0.0",
			},
		},
	}

	fmt.Println("📤 Sending initialize message...")
	if err := conn.WriteJSON(initMsg); err != nil {
		log.Fatal("❌ Failed to send initialize:", err)
	}

	var response MCPResponse
	if err := conn.ReadJSON(&response); err != nil {
		log.Fatal("❌ Failed to read response:", err)
	}
	fmt.Printf("📥 Initialize response received\n")

	messageID++

	// Send initialized notification
	initNotification := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "initialized",
	}
	if err := conn.WriteJSON(initNotification); err != nil {
		log.Fatal("❌ Failed to send initialized notification:", err)
	}
	
	// Small delay to ensure server processes the notification
	time.Sleep(500 * time.Millisecond)

	// Database Tests
	fmt.Println("\n📊 Starting Database Tests...")

	tests := []struct {
		name   string
		tool   string
		args   map[string]string
		expect string
	}{
		{
			name: "Database Health Check",
			tool: "db_health_check",
			args: map[string]string{},
			expect: "healthy",
		},
		{
			name: "Count All Documents",
			tool: "db_count_documents",
			args: map[string]string{
				"collection": "knowledgebase",
			},
			expect: "count",
		},
		{
			name: "Query Security Documents",
			tool: "db_query_documents",
			args: map[string]string{
				"collection": "knowledgebase",
				"filter":     `{"category": "Security"}`,
			},
			expect: "documents",
		},
		{
			name: "Search for Kubernetes",
			tool: "db_search_documents",
			args: map[string]string{
				"collection":  "knowledgebase",
				"search_text": "kubernetes",
			},
			expect: "documents",
		},
		{
			name: "Search for SSL",
			tool: "db_search_documents",
			args: map[string]string{
				"collection":  "knowledgebase",
				"search_text": "ssl certificate",
			},
			expect: "documents",
		},
		{
			name: "Search for Docker",
			tool: "db_search_documents",
			args: map[string]string{
				"collection":  "knowledgebase",
				"search_text": "docker container",
			},
			expect: "documents",
		},
	}

	for i, test := range tests {
		fmt.Printf("\n🧪 Test %d: %s\n", i+1, test.name)
		
		toolCall := MCPRequest{
			ID:      messageID,
			JSONRPC: "2.0",
			Method:  "tools/call",
			Params: ToolCallParams{
				Name:      test.tool,
				Arguments: test.args,
			},
		}

		fmt.Printf("📤 Calling tool: %s\n", test.tool)
		if err := conn.WriteJSON(toolCall); err != nil {
			fmt.Printf("❌ Failed to send tool call: %v\n", err)
			continue
		}

		var toolResponse MCPResponse
		if err := conn.ReadJSON(&toolResponse); err != nil {
			fmt.Printf("❌ Failed to read tool response: %v\n", err)
			continue
		}

		if toolResponse.Error != nil {
			fmt.Printf("❌ Tool error: %v\n", toolResponse.Error)
		} else {
			var result map[string]interface{}
			if err := json.Unmarshal(toolResponse.Result, &result); err != nil {
				fmt.Printf("❌ Failed to parse result: %v\n", err)
			} else {
				if content, ok := result["content"]; ok {
					if contentArray, ok := content.([]interface{}); ok && len(contentArray) > 0 {
						if textContent, ok := contentArray[0].(map[string]interface{}); ok {
							if text, ok := textContent["text"].(string); ok {
								// Parse the JSON text to check structure
								var jsonData interface{}
								if err := json.Unmarshal([]byte(text), &jsonData); err == nil {
									switch test.expect {
									case "healthy":
										fmt.Println("✅ Database is healthy")
									case "count":
										if data, ok := jsonData.(map[string]interface{}); ok {
											if count, ok := data["count"]; ok {
												fmt.Printf("✅ Document count: %.0f\n", count)
											}
										}
									case "documents":
										if data, ok := jsonData.(map[string]interface{}); ok {
											if docs, ok := data["documents"].([]interface{}); ok {
												fmt.Printf("✅ Found %d documents\n", len(docs))
												// Show first document summary
												if len(docs) > 0 {
													if doc, ok := docs[0].(map[string]interface{}); ok {
														if title, ok := doc["title"].(string); ok {
															fmt.Printf("   First result: %s\n", title)
														}
													}
												}
											}
										}
									}
								} else {
									fmt.Printf("📋 Raw response: %s\n", text[:min(200, len(text))])
								}
							}
						}
					}
				}
			}
		}

		messageID++
		time.Sleep(100 * time.Millisecond) // Small delay between tests
	}

	fmt.Println("\n🎉 All database tests completed!")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

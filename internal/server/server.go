package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/kringen/go-mcp-server/pkg/mcp"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins in this example
		// In production, implement proper origin checking
		return true
	},
}

// MCPServer implements the MCP server
type MCPServer struct {
	mu                sync.RWMutex
	toolProviders     []mcp.ToolProvider
	resourceProviders []mcp.ResourceProvider
	connections       map[*websocket.Conn]*Connection
	server            *http.Server
	initialized       bool
}

// Connection represents a client connection
type Connection struct {
	conn        *websocket.Conn
	server      *MCPServer
	initialized bool
	mu          sync.Mutex
}

// NewMCPServer creates a new MCP server instance
func NewMCPServer() *MCPServer {
	return &MCPServer{
		connections: make(map[*websocket.Conn]*Connection),
	}
}

// Start starts the MCP server
func (s *MCPServer) Start(ctx context.Context, addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", s.handleWebSocket)
	mux.HandleFunc("/health", s.handleHealth)
	
	s.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	log.Printf("Starting MCP server on %s", addr)
	
	// Start server in a goroutine
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()
	return s.Stop(context.Background())
}

// Stop stops the MCP server
func (s *MCPServer) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Close all connections
	for conn := range s.connections {
		conn.Close()
	}
	s.connections = make(map[*websocket.Conn]*Connection)

	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

// RegisterToolProvider registers a tool provider
func (s *MCPServer) RegisterToolProvider(provider mcp.ToolProvider) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.toolProviders = append(s.toolProviders, provider)
}

// RegisterResourceProvider registers a resource provider
func (s *MCPServer) RegisterResourceProvider(provider mcp.ResourceProvider) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resourceProviders = append(s.resourceProviders, provider)
}

// handleWebSocket handles WebSocket connections
func (s *MCPServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	connection := &Connection{
		conn:   conn,
		server: s,
	}

	s.mu.Lock()
	s.connections[conn] = connection
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.connections, conn)
		s.mu.Unlock()
		conn.Close()
	}()

	// Handle messages
	for {
		var message mcp.Message
		err := conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		response := connection.handleMessage(&message)
		if response != nil {
			if err := conn.WriteJSON(response); err != nil {
				log.Printf("Failed to write response: %v", err)
				break
			}
		}
	}
}

// handleHealth handles health check requests
func (s *MCPServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	response := map[string]interface{}{
		"status": "healthy",
		"service": "mcp-server",
		"timestamp": fmt.Sprintf("%d", time.Now().Unix()),
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleMessage processes incoming messages
func (c *Connection) handleMessage(message *mcp.Message) interface{} {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Handle requests
	if message.Method != "" && message.ID != nil {
		return c.handleRequest(message)
	}

	// Handle notifications
	if message.Method != "" && message.ID == nil {
		c.handleNotification(message)
		return nil
	}

	// Invalid message
	return mcp.NewErrorResponse(message.ID, mcp.ErrorCodeInvalidRequest, "Invalid message format", nil)
}

// handleRequest processes MCP requests
func (c *Connection) handleRequest(message *mcp.Message) *mcp.Response {
	switch message.Method {
	case mcp.MethodInitialize:
		return c.handleInitialize(message)
	case mcp.MethodListTools:
		return c.handleListTools(message)
	case mcp.MethodCallTool:
		return c.handleCallTool(message)
	case mcp.MethodListResources:
		return c.handleListResources(message)
	case mcp.MethodReadResource:
		return c.handleReadResource(message)
	default:
		return mcp.NewErrorResponse(message.ID, mcp.ErrorCodeMethodNotFound, 
			fmt.Sprintf("Method not found: %s", message.Method), nil)
	}
}

// handleNotification processes MCP notifications
func (c *Connection) handleNotification(message *mcp.Message) {
	switch message.Method {
	case mcp.MethodInitialized:
		c.initialized = true
		log.Println("Client initialized")
	default:
		log.Printf("Unknown notification: %s", message.Method)
	}
}

// handleInitialize processes initialize requests
func (c *Connection) handleInitialize(message *mcp.Message) *mcp.Response {
	var req mcp.InitializeRequest
	if message.Params != nil {
		paramsBytes, _ := json.Marshal(message.Params)
		if err := json.Unmarshal(paramsBytes, &req); err != nil {
			return mcp.NewErrorResponse(message.ID, mcp.ErrorCodeInvalidParams, 
				"Invalid initialize parameters", err.Error())
		}
	}

	response := mcp.InitializeResponse{
		ProtocolVersion: mcp.ProtocolVersion,
		Capabilities: mcp.ServerCapabilities{
			Tools: &mcp.ToolsCapability{
				ListChanged: false,
			},
			Resources: &mcp.ResourcesCapability{
				Subscribe:   false,
				ListChanged: false,
			},
		},
		ServerInfo: mcp.ServerInfo{
			Name:    "MCP Server Go",
			Version: "1.0.0",
		},
	}

	return mcp.NewResponse(message.ID, response)
}

// handleListTools processes list tools requests
func (c *Connection) handleListTools(message *mcp.Message) *mcp.Response {
	if !c.initialized {
		return mcp.NewErrorResponse(message.ID, mcp.ErrorCodeInvalidRequest, 
			"Client not initialized", nil)
	}

	var allTools []mcp.Tool
	
	c.server.mu.RLock()
	for _, provider := range c.server.toolProviders {
		tools, err := provider.ListTools(context.Background())
		if err != nil {
			c.server.mu.RUnlock()
			return mcp.NewErrorResponse(message.ID, mcp.ErrorCodeInternalError, 
				"Failed to list tools", err.Error())
		}
		allTools = append(allTools, tools...)
	}
	c.server.mu.RUnlock()

	result := map[string]interface{}{
		"tools": allTools,
	}

	return mcp.NewResponse(message.ID, result)
}

// handleCallTool processes tool call requests
func (c *Connection) handleCallTool(message *mcp.Message) *mcp.Response {
	if !c.initialized {
		return mcp.NewErrorResponse(message.ID, mcp.ErrorCodeInvalidRequest, 
			"Client not initialized", nil)
	}

	var req mcp.ToolCallRequest
	if message.Params != nil {
		paramsBytes, _ := json.Marshal(message.Params)
		if err := json.Unmarshal(paramsBytes, &req); err != nil {
			return mcp.NewErrorResponse(message.ID, mcp.ErrorCodeInvalidParams, 
				"Invalid tool call parameters", err.Error())
		}
	}

	c.server.mu.RLock()
	defer c.server.mu.RUnlock()

	for _, provider := range c.server.toolProviders {
		// Check if this provider has the requested tool
		tools, err := provider.ListTools(context.Background())
		if err != nil {
			continue
		}

		for _, tool := range tools {
			if tool.Name == req.Name {
				response, err := provider.CallTool(context.Background(), req)
				if err != nil {
					return mcp.NewErrorResponse(message.ID, mcp.ErrorCodeInternalError, 
						"Tool execution failed", err.Error())
				}
				return mcp.NewResponse(message.ID, response)
			}
		}
	}

	return mcp.NewErrorResponse(message.ID, mcp.ErrorCodeMethodNotFound, 
		fmt.Sprintf("Tool not found: %s", req.Name), nil)
}

// handleListResources processes list resources requests
func (c *Connection) handleListResources(message *mcp.Message) *mcp.Response {
	if !c.initialized {
		return mcp.NewErrorResponse(message.ID, mcp.ErrorCodeInvalidRequest, 
			"Client not initialized", nil)
	}

	var allResources []mcp.Resource
	
	c.server.mu.RLock()
	for _, provider := range c.server.resourceProviders {
		resources, err := provider.ListResources(context.Background())
		if err != nil {
			c.server.mu.RUnlock()
			return mcp.NewErrorResponse(message.ID, mcp.ErrorCodeInternalError, 
				"Failed to list resources", err.Error())
		}
		allResources = append(allResources, resources...)
	}
	c.server.mu.RUnlock()

	result := map[string]interface{}{
		"resources": allResources,
	}

	return mcp.NewResponse(message.ID, result)
}

// handleReadResource processes read resource requests
func (c *Connection) handleReadResource(message *mcp.Message) *mcp.Response {
	if !c.initialized {
		return mcp.NewErrorResponse(message.ID, mcp.ErrorCodeInvalidRequest, 
			"Client not initialized", nil)
	}

	var req mcp.ResourceReadRequest
	if message.Params != nil {
		paramsBytes, _ := json.Marshal(message.Params)
		if err := json.Unmarshal(paramsBytes, &req); err != nil {
			return mcp.NewErrorResponse(message.ID, mcp.ErrorCodeInvalidParams, 
				"Invalid resource read parameters", err.Error())
		}
	}

	c.server.mu.RLock()
	defer c.server.mu.RUnlock()

	for _, provider := range c.server.resourceProviders {
		// Check if this provider has the requested resource
		resources, err := provider.ListResources(context.Background())
		if err != nil {
			continue
		}

		for _, resource := range resources {
			if resource.URI == req.URI {
				response, err := provider.ReadResource(context.Background(), req.URI)
				if err != nil {
					return mcp.NewErrorResponse(message.ID, mcp.ErrorCodeInternalError, 
						"Resource read failed", err.Error())
				}
				return mcp.NewResponse(message.ID, response)
			}
		}
	}

	return mcp.NewErrorResponse(message.ID, mcp.ErrorCodeMethodNotFound, 
		fmt.Sprintf("Resource not found: %s", req.URI), nil)
}

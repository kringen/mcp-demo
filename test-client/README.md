# MCP WebSocket Test Client

A simple Go client to test the MCP (Model Context Protocol) WebSocket server.

## Usage

```bash
go run main.go [options]
```

## Options

- `-host string` - Server host/IP address (default: localhost)
- `-port string` - Server port (default: 8080)  
- `-help` - Show usage information

## Examples

### Local Development (Port Forwarding)
```bash
# Test local server or port-forwarded service
go run main.go
go run main.go -host localhost -port 8080
```

### LoadBalancer Testing (Recommended)
```bash
# Test Kubernetes LoadBalancer service (production approach)
go run main.go -host your-loadbalancer-ip -port 80
go run main.go -host mcp.example.com -port 80

# With custom domain/hostname
go run main.go -host mcp-server.your-domain.com -port 443
```

### Remote Server Testing
```bash
# Test remote MCP server
go run main.go -host example.com -port 443
go run main.go -host mcp.mycompany.com -port 8080
```

## Test Sequence

The client performs these tests in order:

1. **WebSocket Connection** - Establishes WS connection
2. **Initialize** - Sends MCP initialize message
3. **Initialized** - Sends initialized notification
4. **List Tools** - Requests available tools
5. **Math Tool** - Tests add function (5 + 3)
6. **Database Health** - Tests database connectivity

## Expected Output

```
🧪 MCP WebSocket Test Client
==============================
🔌 Connecting to ws://your-loadbalancer-ip/mcp
✅ Connected to WebSocket!

📤 Sending initialize message...
📥 Response: { ... }

📤 Sending initialized notification...

📤 Requesting tools list...
📥 Response: { ... }

📤 Testing math tool (add 5 + 3)...
📥 Response: { ... }

📤 Testing database health check...
📥 Response: { ... }

🎉 All tests completed successfully!
```

## Available Tools

The MCP server provides these tools:

### Math Tools
- `add` - Add two numbers
- `multiply` - Multiply two numbers  
- `power` - Calculate power (base^exponent)

### Search Tools
- `web_search` - Search the web
- `search_health_check` - Check search service health

### Database Tools
- `db_create_document` - Create new document
- `db_get_document` - Get document by ID
- `db_update_document` - Update existing document
- `db_delete_document` - Delete document
- `db_query_documents` - Query documents
- `db_search_documents` - Text search documents
- `db_count_documents` - Count documents
- `db_health_check` - Check database health

## Troubleshooting

### Connection Failed
```bash
❌ Connection failed: dial tcp your-loadbalancer-ip:80: connect: connection refused
```

**Solutions:**
- Verify the server is running: `kubectl get pods -n mcp-server`
- Check the LoadBalancer: `kubectl get svc -n mcp-server`
- Verify external IP assignment: `kubectl describe svc mcp-server-loadbalancer -n mcp-server`
- Test health endpoint: `curl http://your-loadbalancer-ip/health`

### Timeout
```bash
# Use timeout for long-running tests
timeout 15s go run main.go -host your-loadbalancer-ip -port 80
```

### Network Issues
- Verify firewall rules allow traffic on port 80/443
- Check if LoadBalancer is accessible from your location
- Verify DNS resolution if using hostname
- Check cloud provider LoadBalancer configuration

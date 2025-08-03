# MCP Server Testing Guide

This comprehensive guide shows you how to test the MCP server in all deployment scenarios.

## 🚀 Quick Testing

### Automated Test Script
```bash
# Run the comprehensive test script
./test-services.sh
```

This script automatically tests:
- ✅ Health endpoint accessibility  
- ✅ Service container status
- ✅ MongoDB connectivity
- ✅ MongoDB Express admin interface
- ✅ WebSocket endpoint availability

### Manual Quick Tests
```bash
# Test health endpoint
curl http://localhost:8080/health

# Check Docker services
docker-compose ps

# Test WebSocket MCP protocol
cd test-client && go run main.go
```

## 1. Service Status & Health Checks

### Check Running Containers
```bash
# View all service status
docker-compose ps

# Expected output:
#   mcp-server      Up (healthy)   0.0.0.0:8080->8080/tcp
#   mcp-mongodb     Up             0.0.0.0:27017->27017/tcp  
#   mcp-mongo-express Up           0.0.0.0:8081->8081/tcp
```

### Service Logs
```bash
# View all service logs
make docker-logs

# View specific service logs
docker-compose logs -f mcp-server
docker-compose logs -f mongodb
docker-compose logs --tail=20 mcp-server
```

### Container Health Status
```bash
# Check Docker health status
docker inspect mcp-server | jq '.[].State.Health'

# Monitor resource usage
docker stats mcp-server mcp-mongodb
```

## 2. Health Checks & Endpoints

### HTTP Health Endpoint
```bash
# Basic health check
curl http://localhost:8080/health

# Pretty-formatted health check  
curl -s http://localhost:8080/health | jq .

# Expected response:
{
  "service": "mcp-server",
  "status": "healthy",
  "timestamp": "1753757075"
}
```

### Service Endpoints Summary
- **MCP WebSocket**: `ws://localhost:8080/mcp`
- **Health Check**: `http://localhost:8080/health` 
- **MongoDB**: `mongodb://admin:password@localhost:27017/mcp_server`
- **MongoDB Express**: `http://localhost:8081` (admin interface)

### Endpoint Accessibility Tests
```bash
# Test WebSocket upgrade (should return HTTP 101)
curl -i -H "Connection: Upgrade" \
     -H "Upgrade: websocket" \
     -H "Sec-WebSocket-Version: 13" \
     -H "Sec-WebSocket-Key: test" \
     http://localhost:8080/mcp

# Test MongoDB Express (should return HTTP 200)
curl -s -o /dev/null -w "%{http_code}" http://localhost:8081
```

## 3. WebSocket & MCP Protocol Testing

### Using the Built-in Test Client (Recommended)

The easiest way to test the MCP protocol:

```bash
cd test-client && go run main.go
```

**Expected Output:**
```
🧪 MCP WebSocket Test Client
==============================
🔌 Connecting to ws://localhost:8080/mcp
✅ Connected to WebSocket!

📤 Sending initialize message...
📥 Response: { "result": { "capabilities": {...}, "protocolVersion": "2024-11-05" } }

📤 Sending initialized notification...

📤 Requesting tools list...
📥 Response: { "result": { "tools": [13 tools listed] } }

📤 Testing math tool (add 5 + 3)...
📥 Response: { "result": { "content": [{ "text": "5.00 + 3.00 = 8.00" }] } }

📤 Testing database health check...
📥 Response: { "result": { "content": [{ "text": "Database is healthy" }] } }

🎉 All tests completed successfully!
```

### Using wscat (Alternative)

If you have Node.js available:

```bash
# Install wscat
npm install -g wscat

# Connect to MCP server
wscat -c ws://localhost:8080/mcp
```

### Manual WebSocket Testing

#### Basic Connection Test
```bash
# Test WebSocket endpoint availability (should return HTTP 101)
curl -i -N \
  -H "Connection: Upgrade" \
  -H "Upgrade: websocket" \
  -H "Sec-WebSocket-Version: 13" \
  -H "Sec-WebSocket-Key: test" \
  http://localhost:8080/mcp
```

### Test MCP Protocol Messages

Once connected via wscat, try these MCP messages:

#### 1. Initialize Connection
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "initialize",
  "params": {
    "protocolVersion": "2024-11-05",
    "capabilities": {
      "tools": {}
    },
    "clientInfo": {
      "name": "test-client",
      "version": "1.0.0"
    }
  }
}
```

#### 2. List Available Tools
```json
{
  "jsonrpc": "2.0", 
  "id": 2,
  "method": "tools/list"
}
```

#### 3. Test Math Tool
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "add",
    "arguments": {
      "a": 5,
      "b": 3
    }
  }
}
```

#### 4. Test Database Health Check
```json
{
  "jsonrpc": "2.0",
  "id": 4,
  "method": "tools/call",
  "params": {
    "name": "db_health_check"
  }
}
```

#### 5. Test Web Search
```json
{
  "jsonrpc": "2.0",
  "id": 5,
  "method": "tools/call",
  "params": {
    "name": "web_search",
    "arguments": {
      "query": "golang tutorial",
      "max_results": 3
    }
  }
}
```

## 4. Database Testing

### Direct MongoDB Access
```bash
# Connect to MongoDB container
docker exec -it mcp-mongodb mongosh

# Or using connection string
mongosh "mongodb://admin:password@localhost:27017/mcp_server?authSource=admin"
```

### MongoDB Express (if admin profile is running)
```bash
# Start services with admin interface
make docker-run-all

# Access MongoDB Express
open http://localhost:8081
```

## 5. Integration Testing

### Run the integrated test suite
```bash
# Run integration tests (requires services to be running)
make test

# Run specific integration tests
go test -v ./tests/...
```

### Test Database Operations
```json
# Create a document
{
  "jsonrpc": "2.0",
  "id": 6,
  "method": "tools/call",
  "params": {
    "name": "db_create_document",
    "arguments": {
      "collection": "test_collection",
      "document": {
        "name": "Test Document",
        "created_at": "2025-01-01T00:00:00Z",
        "tags": ["test", "demo"]
      }
    }
  }
}
```

```json
# Query documents
{
  "jsonrpc": "2.0",
  "id": 7,
  "method": "tools/call",
  "params": {
    "name": "db_query_documents",
    "arguments": {
      "collection": "test_collection",
      "filter": {
        "tags": "test"
      },
      "limit": 10
    }
  }
}
```

## 6. Performance Testing

### Load Testing with WebSocket
You can use tools like `artillery` or custom scripts to test WebSocket performance:

```bash
# Install artillery
npm install -g artillery

# Create a load test configuration
cat > load-test.yml << EOF
config:
  target: 'ws://localhost:8080'
  phases:
    - duration: 60
      arrivalRate: 10
  engines:
    ws: {}

scenarios:
  - name: "MCP WebSocket Load Test"
    engine: ws
    flow:
      - connect:
          url: "/mcp"
      - send:
          payload: |
            {
              "jsonrpc": "2.0",
              "id": 1,
              "method": "tools/list"
            }
      - think: 1
EOF

# Run load test
artillery run load-test.yml
```

## 7. Troubleshooting

### Common Issues and Solutions

#### Service won't start
```bash
# Check logs
docker-compose logs mcp-server

# Check if ports are in use
lsof -i :8080
lsof -i :27017
```

#### Connection refused
```bash
# Verify services are running
docker-compose ps

# Check network connectivity
docker exec mcp-server ping mongodb
```

#### Health check fails
```bash
# Test health endpoint manually
curl -v http://localhost:8080/health

# Check server binding address
docker-compose logs mcp-server | grep "Starting MCP server"
```

#### Database connection issues
```bash
# Test MongoDB connection
docker exec mcp-mongodb mongosh --eval "db.adminCommand('ping')"

# Check MongoDB logs
docker-compose logs mongodb
```

## 8. Monitoring and Metrics

### Container Resource Usage
```bash
# Monitor container stats
docker stats mcp-server mcp-mongodb

# Check container health
docker inspect mcp-server | jq '.[].State.Health'
```

### Service Endpoints Summary
- **MCP WebSocket**: `ws://localhost:8080/mcp`
- **Health Check**: `http://localhost:8080/health` 
- **MongoDB**: `mongodb://admin:password@localhost:27017/mcp_server`
- **MongoDB Express**: `http://localhost:8081` (when admin profile is active)

## 9. Useful Make Commands

```bash
# Start core services
make docker-run

# Start all services with admin interface  
make docker-run-all

# Stop all services
make docker-stop

# View logs
make docker-logs

# Clean up (stop and remove volumes)
make docker-clean

# Rebuild and restart
make docker-stop && make docker-build && make docker-run
```

This testing guide covers all the major ways to verify that your MCP server is working correctly in the Docker environment.

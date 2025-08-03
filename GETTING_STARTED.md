# Getting Started with Go MCP Server

This guide will get you up and running with the MCP server in under 5 minutes.

## 🎯 Prerequisites

- **Docker & Docker Compose** (recommended approach)
- **OR** Go 1.21+ and MongoDB (for local development)

## 🚀 Quick Start (Docker - Recommended)

### 1. Clone and Start

```bash
# Clone the repository
git clone https://github.com/kringen/go-mcp-server.git
cd go-mcp-server

# Start all services (this may take a few minutes on first run)
make docker-run-all
```

### 2. Verify Installation

```bash
# Run the automated test script
./test-services.sh
```

**Expected output:**
```
🔍 MCP Server Testing Script
============================
1. Testing Health Endpoint...
✅ Health endpoint: {"service":"mcp-server","status":"healthy","timestamp":"..."}

2. Checking Service Status...
   mcp-server      Up (healthy)   0.0.0.0:8080->8080/tcp
   mcp-mongodb     Up             0.0.0.0:27017->27017/tcp  
   mcp-mongo-express Up           0.0.0.0:8081->8081/tcp

3. Testing MongoDB Connection...
✅ MongoDB is responding

4. Testing MongoDB Express...
✅ MongoDB Express is accessible at http://localhost:8081

5. Testing WebSocket Endpoint...
✅ WebSocket endpoint accepting connections (ws://localhost:8080/mcp)

6. Testing MCP Protocol...
✅ MCP protocol test passed
   📊 Tools available: 13 tools detected

🎉 Testing Complete!
```

### 3. Test the MCP Protocol

```bash
# Test WebSocket MCP communication
cd test-client && go run main.go
```

This will connect to your MCP server and test:
- ✅ WebSocket connection
- ✅ MCP protocol handshake
- ✅ Tools discovery (13 tools)
- ✅ Math tool execution (5 + 3 = 8)
- ✅ Database health check

## 🎉 You're Ready!

Your MCP server is now running with these endpoints:

### **Service Endpoints**
- **MCP WebSocket**: `ws://localhost:8080/mcp`
- **Health Check**: `http://localhost:8080/health`
- **MongoDB**: `mongodb://admin:password@localhost:27017/mcp_server`
- **Admin Interface**: `http://localhost:8081`

### **Available Tools (13 total)**
- **Math**: `add`, `multiply`, `divide`, `power`
- **Search**: `web_search`, `search_health_check`
- **Database**: `db_create_document`, `db_get_document`, `db_update_document`, `db_delete_document`, `db_query_documents`, `db_search_documents`, `db_count_documents`, `db_health_check`

## 🔧 Management Commands

```bash
# View logs
make docker-logs

# Stop services
make docker-stop

# Restart services
make docker-run-all

# Clean restart (removes data)
make docker-clean && make docker-run-all
```

## 🧪 Advanced Testing

### WebSocket Client Example

You can test the MCP protocol with any WebSocket client. Here's the basic flow:

1. **Connect** to `ws://localhost:8080/mcp`
2. **Initialize**:
   ```json
   {
     "jsonrpc": "2.0",
     "id": 1,
     "method": "initialize",
     "params": {
       "protocolVersion": "2024-11-05",
       "capabilities": {"tools": {}},
       "clientInfo": {"name": "test-client", "version": "1.0.0"}
     }
   }
   ```
3. **Send initialized notification**:
   ```json
   {"jsonrpc": "2.0", "method": "initialized"}
   ```
4. **List tools**:
   ```json
   {"jsonrpc": "2.0", "id": 2, "method": "tools/list"}
   ```
5. **Call a tool**:
   ```json
   {
     "jsonrpc": "2.0",
     "id": 3,
     "method": "tools/call",
     "params": {
       "name": "add",
       "arguments": {"a": 5, "b": 3}
     }
   }
   ```

### Database Testing

Access the MongoDB admin interface at http://localhost:8081 to:
- View collections and documents
- Run database queries
- Monitor database performance

### Integration with MCP Clients

This server is compatible with any MCP client. Popular clients include:
- **Claude Desktop** (with MCP support)
- **Custom WebSocket clients**
- **Any application supporting MCP protocol**

## 🐛 Troubleshooting

### Services won't start
```bash
# Check if ports are in use
lsof -i :8080
lsof -i :27017
lsof -i :8081

# Check Docker logs
make docker-logs
```

### Health check fails
```bash
# Test manually
curl -v http://localhost:8080/health

# Check server logs
docker-compose logs mcp-server
```

### WebSocket connection issues
```bash
# Test WebSocket endpoint
curl -i -H "Connection: Upgrade" -H "Upgrade: websocket" \
     -H "Sec-WebSocket-Version: 13" -H "Sec-WebSocket-Key: test" \
     http://localhost:8080/mcp
```

### Reset everything
```bash
# Complete reset
make docker-stop
docker system prune -f
make docker-run-all
```

## 📚 Next Steps

- **Read**: [TESTING.md](TESTING.md) for comprehensive testing guide
- **Explore**: [README.md](README.md) for detailed architecture
- **Develop**: Check the `/internal` directory for extending functionality
- **Deploy**: See Docker configuration for production deployment

## 💡 Tips

1. **First time?** The Docker images need to download, so first start takes longer
2. **Development?** Use `make docker-logs` to see what's happening
3. **Testing?** Always run `./test-services.sh` after changes
4. **Stuck?** Check the troubleshooting section above

**Happy coding with MCP! 🚀**

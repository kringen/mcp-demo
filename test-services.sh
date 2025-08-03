#!/bin/bash

# Simple test script for MCP Server
echo "🔍 MCP Server Testing Script"
echo "============================"

# Test 1: Health Check
echo "1. Testing Health Endpoint..."
health_response=$(curl -s http://localhost:8080/health)
if [ $? -eq 0 ]; then
    echo "✅ Health endpoint: $health_response"
else
    echo "❌ Health endpoint failed"
fi

# Test 2: Check all services are running
echo ""
echo "2. Checking Service Status..."
echo "Docker services:"
docker-compose ps

# Test 3: Test MongoDB connectivity
echo ""
echo "3. Testing MongoDB Connection..."
mongodb_test=$(docker exec mcp-mongodb mongosh --quiet --eval "db.adminCommand('ping').ok" 2>/dev/null)
if [ "$mongodb_test" = "1" ]; then
    echo "✅ MongoDB is responding"
else
    echo "❌ MongoDB connection failed"
fi

# Test 4: Check if MongoDB Express is accessible
echo ""
echo "4. Testing MongoDB Express..."
express_test=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8081)
if [ "$express_test" = "200" ]; then
    echo "✅ MongoDB Express is accessible at http://localhost:8081"
else
    echo "❌ MongoDB Express not accessible (HTTP $express_test)"
fi

# Test 5: WebSocket endpoint basic connectivity
echo ""
echo "5. Testing WebSocket Endpoint..."
# Using curl to test WebSocket upgrade
ws_test=$(curl -s -o /dev/null -w "%{http_code}" \
    -H "Connection: Upgrade" \
    -H "Upgrade: websocket" \
    -H "Sec-WebSocket-Version: 13" \
    -H "Sec-WebSocket-Key: test" \
    http://localhost:8080/mcp)

if [ "$ws_test" = "101" ]; then
    echo "✅ WebSocket endpoint accepting connections (ws://localhost:8080/mcp)"
else
    echo "❌ WebSocket endpoint not responding properly (HTTP $ws_test)"
fi

# Test 6: MCP Protocol Test
echo ""
echo "6. Testing MCP Protocol..."
if [ -f "test-client/main.go" ]; then
    echo "Running MCP protocol test..."
    cd test-client && timeout 10s go run main.go > /tmp/mcp_test.out 2>&1
    test_result=$?
    cd ..
    
    if [ $test_result -eq 0 ]; then
        echo "✅ MCP protocol test passed"
        echo "   📊 Tools available: $(grep -o '"name":[^,]*' /tmp/mcp_test.out | wc -l) tools detected"
    else
        echo "⚠️  MCP protocol test had issues (timeout or error)"
        echo "   💡 Run manually: cd test-client && go run main.go"
    fi
else
    echo "⚠️  MCP test client not found. Run manually with:"
    echo "   cd test-client && go run main.go"
fi

echo ""
echo "🎉 Testing Complete!"
echo ""
echo "Available Endpoints:"
echo "  - Health Check: http://localhost:8080/health"
echo "  - WebSocket MCP: ws://localhost:8080/mcp"
echo "  - MongoDB: mongodb://admin:password@localhost:27017/mcp_server"
echo "  - MongoDB Express: http://localhost:8081"
echo ""
echo "Next Steps:"
echo "  1. Test MCP Protocol:     cd test-client && go run main.go"
echo "  2. View server logs:      make docker-logs"
echo "  3. Stop services:         make docker-stop"
echo "  4. Restart services:      make docker-run-all"
echo ""
echo "📖 For detailed testing guide, see: TESTING.md"
echo "📊 For admin interface, visit: http://localhost:8081"

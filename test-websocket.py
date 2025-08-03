#!/usr/bin/env python3
"""
Simple WebSocket test client for MCP server
"""
import asyncio
import websockets
import json

async def test_mcp_server():
    uri = "ws://localhost:8080/mcp"
    
    try:
        print("🔌 Connecting to MCP server...")
        async with websockets.connect(uri) as websocket:
            print("✅ Connected to WebSocket!")
            
            # Test 1: Initialize connection
            print("\n📤 Sending initialize message...")
            init_message = {
                "jsonrpc": "2.0",
                "id": 1,
                "method": "initialize",
                "params": {
                    "protocolVersion": "2024-11-05",
                    "capabilities": {"tools": {}},
                    "clientInfo": {"name": "test-client", "version": "1.0.0"}
                }
            }
            
            await websocket.send(json.dumps(init_message))
            response = await websocket.recv()
            print(f"📥 Response: {response}")
            
            # Test 2: List tools
            print("\n📤 Requesting tools list...")
            tools_message = {
                "jsonrpc": "2.0",
                "id": 2,
                "method": "tools/list"
            }
            
            await websocket.send(json.dumps(tools_message))
            response = await websocket.recv()
            print(f"📥 Response: {response}")
            
            # Test 3: Call math tool
            print("\n📤 Testing math tool (add)...")
            math_message = {
                "jsonrpc": "2.0",
                "id": 3,
                "method": "tools/call",
                "params": {
                    "name": "add",
                    "arguments": {"a": 5, "b": 3}
                }
            }
            
            await websocket.send(json.dumps(math_message))
            response = await websocket.recv()
            print(f"📥 Response: {response}")
            
            # Test 4: Database health check
            print("\n📤 Testing database health check...")
            db_health_message = {
                "jsonrpc": "2.0",
                "id": 4,
                "method": "tools/call",
                "params": {
                    "name": "db_health_check"
                }
            }
            
            await websocket.send(json.dumps(db_health_message))
            response = await websocket.recv()
            print(f"📥 Response: {response}")
            
            print("\n🎉 All tests completed successfully!")
            
    except websockets.exceptions.ConnectionRefused:
        print("❌ Connection refused. Is the MCP server running?")
    except Exception as e:
        print(f"❌ Error: {e}")

if __name__ == "__main__":
    print("🧪 MCP WebSocket Test Client")
    print("=" * 30)
    asyncio.run(test_mcp_server())

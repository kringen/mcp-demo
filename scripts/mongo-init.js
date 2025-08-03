// MongoDB initialization script
db = db.getSiblingDB('mcp_server');

// Create collections
db.createCollection('search_cache');
db.createCollection('documents');
db.createCollection('users');

// Create indexes for better performance
db.search_cache.createIndex({ "query": 1, "timestamp": 1 });
db.search_cache.createIndex({ "timestamp": 1 }, { expireAfterSeconds: 3600 }); // TTL index for 1 hour cache

db.documents.createIndex({ "title": "text", "content": "text" });
db.documents.createIndex({ "created_at": -1 });
db.documents.createIndex({ "tags": 1 });

db.users.createIndex({ "username": 1 }, { unique: true });
db.users.createIndex({ "email": 1 }, { unique: true });

print('Database initialized successfully');

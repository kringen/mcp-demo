#!/bin/bash

# Load test data into MongoDB for MCP Server testing
echo "📚 Loading Knowledge Base Test Data"
echo "=================================="

# Configuration
NAMESPACE="mcp-server"
MONGODB_SERVICE="mongodb-service"
DATABASE="mcp_server"
COLLECTION="knowledgebase"
JSON_FILE="test-data/kb-articles.json"

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl is not installed or not in PATH"
    exit 1
fi

# Check if jq is available
if ! command -v jq &> /dev/null; then
    echo "❌ jq is not installed. Installing..."
    sudo apt-get update && sudo apt-get install -y jq
fi

# Check if MongoDB is running
echo "🔍 Checking MongoDB service..."
if ! kubectl get service $MONGODB_SERVICE -n $NAMESPACE &> /dev/null; then
    echo "❌ MongoDB service not found. Please deploy the MCP server first:"
    echo "   ./k8s/deploy.sh"
    exit 1
fi

# Check if JSON file exists
if [ ! -f "$JSON_FILE" ]; then
    echo "❌ Test data file not found: $JSON_FILE"
    exit 1
fi

echo "✅ MongoDB service found"
echo "✅ Test data file found: $JSON_FILE"

# Get MongoDB pod name
MONGO_POD=$(kubectl get pods -n $NAMESPACE -l app=mongodb -o jsonpath='{.items[0].metadata.name}')
if [ -z "$MONGO_POD" ]; then
    echo "❌ MongoDB pod not found"
    exit 1
fi

echo "✅ MongoDB pod: $MONGO_POD"

# Copy JSON file to MongoDB pod
echo ""
echo "📤 Copying test data to MongoDB pod..."
kubectl cp "$JSON_FILE" "$NAMESPACE/$MONGO_POD:/tmp/kb-articles.json"

# Import data using mongoimport
echo "📥 Importing test data into MongoDB..."
kubectl exec -n $NAMESPACE $MONGO_POD -- mongoimport \
    --host localhost:27017 \
    --username admin \
    --password password \
    --authenticationDatabase admin \
    --db $DATABASE \
    --collection $COLLECTION \
    --file /tmp/kb-articles.json \
    --jsonArray \
    --drop

if [ $? -eq 0 ]; then
    echo "✅ Test data imported successfully!"
else
    echo "❌ Failed to import test data"
    exit 1
fi

# Create text index for search functionality
echo ""
echo "🔍 Creating text search index..."
kubectl exec -n $NAMESPACE $MONGO_POD -- mongosh \
    --host localhost:27017 \
    --username admin \
    --password password \
    --authenticationDatabase admin \
    --quiet \
    --eval "
        use $DATABASE;
        try {
            db.$COLLECTION.createIndex(
                { 
                    title: 'text', 
                    content: 'text', 
                    tags: 'text',
                    category: 'text'
                },
                { 
                    name: 'kb_text_index',
                    weights: { 
                        title: 10, 
                        tags: 5, 
                        category: 3, 
                        content: 1 
                    }
                }
            );
            print('✅ Text search index created successfully');
        } catch (e) {
            if (e.code === 85) {
                print('✅ Text search index already exists');
            } else {
                print('❌ Error creating index: ' + e.message);
            }
        }
    "

# Verify data was loaded
echo ""
echo "📊 Verifying data import..."
DOCUMENT_COUNT=$(kubectl exec -n $NAMESPACE $MONGO_POD -- mongosh \
    --host localhost:27017 \
    --username admin \
    --password password \
    --authenticationDatabase admin \
    --quiet \
    --eval "use $DATABASE; print(db.$COLLECTION.countDocuments({})); quit()")

echo "✅ Documents imported: $DOCUMENT_COUNT"

# Show sample data
echo ""
echo "📖 Sample document:"
kubectl exec -n $NAMESPACE $MONGO_POD -- mongosh \
    --host localhost:27017 \
    --username admin \
    --password password \
    --authenticationDatabase admin \
    --quiet \
    --eval "
        use $DATABASE; 
        const doc = db.$COLLECTION.findOne({}, {title: 1, category: 1, tags: 1, _id: 0});
        if (doc) {
            print('Title: ' + doc.title);
            print('Category: ' + doc.category);
            print('Tags: ' + (doc.tags ? doc.tags.join(', ') : 'none'));
        } else {
            print('No documents found');
        }
        quit();
    "

# Clean up temporary file
kubectl exec -n $NAMESPACE $MONGO_POD -- rm -f /tmp/kb-articles.json

echo ""
echo "🎉 Test data loading complete!"
echo ""
echo "📋 You can now test the following database operations:"
echo "   • Search: db_search_documents with queries like 'kubernetes', 'ssl', 'database'"
echo "   • Query: db_query_documents with filters like {category: 'Networking'}"
echo "   • Count: db_count_documents with filters"
echo "   • Get: db_get_document with a document ID"
echo ""
echo "🧪 Test with the WebSocket client:"
echo "   cd test-client && go run main.go -host your-loadbalancer-ip -port 80"

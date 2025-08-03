# Knowledge Base Test Data

This directory contains comprehensive test data for the MCP Server database functionality.

## 📋 Contents

### `kb-articles.json`
- **37 technical support articles** covering common IT scenarios
- **Categories**: Authentication, Networking, Database, Kubernetes, API, Security, Docker, Version Control, Load Balancing, Performance, Backup, Microservices, Caching, CI/CD, Messaging, Search, Container Registry, Web Development, Monitoring, Infrastructure, Streaming, Mobile, Windows, Serverless, API Gateway
- **Rich metadata**: Priority levels, difficulty ratings, view counts, helpful votes, timestamps
- **Searchable content**: Titles, content, tags, and categories optimized for text search

### Article Categories and Examples

#### 🔐 Security (3 articles)
- SSL Certificate Validation Failures
- OAuth 2.0 Authentication Failures  
- Container Image Security Vulnerability Scanning

#### 🌐 Networking (4 articles)
- Troubleshooting WebSocket Connection Issues
- DNS Resolution Problems
- VPN Connection Drops and Instability
- Load Balancer Health Check Failures

#### 🗄️ Database (4 articles)
- Database Connection Timeout Errors
- MySQL Deadlock Detection and Resolution
- PostgreSQL Connection Pool Exhaustion
- Redis Cache Connection Issues

#### ☸️ Kubernetes (2 articles)
- Kubernetes Pod CrashLoopBackOff Error
- Load Balancer Health Check Failures

#### 🐳 Containers (3 articles)
- Docker Container Out of Memory Issues
- Container Registry Push/Pull Failures
- Container Image Security Vulnerability Scanning

#### 🔧 DevOps & CI/CD (3 articles)
- CI/CD Pipeline Build Failures
- Jenkins Build Agent Connectivity Issues
- Terraform State Lock Conflicts

#### 📱 Applications (6 articles)
- API Rate Limiting and 429 Errors
- GraphQL Query Performance Issues
- Cross-Origin Resource Sharing (CORS) Errors
- Mobile App Push Notification Failures
- Serverless Function Cold Start Optimization
- Application Performance Optimization

#### 🏗️ Infrastructure & Operations (6 articles)
- Backup and Disaster Recovery Best Practices
- Prometheus Monitoring Alert Fatigue
- Nginx Reverse Proxy Timeout Errors
- Windows Service Startup Failures
- S3 Bucket Access Denied Errors
- API Gateway Rate Limiting Configuration

#### 🔄 Integration & Messaging (6 articles)
- Microservices Communication Failures
- Message Queue Processing Delays
- Apache Kafka Consumer Lag Issues
- ElasticSearch Index Performance Issues
- LDAP Authentication Integration Issues
- Git Merge Conflicts Resolution

## 🚀 Loading Test Data

### Prerequisites
- Kubernetes cluster with MCP server deployed
- MongoDB instance running
- kubectl configured and accessible

### Load Data
```bash
# Run the automated loading script
./scripts/load-test-data.sh
```

### What the Script Does
1. **Copies JSON data** to MongoDB pod
2. **Imports articles** using `mongoimport` with `--jsonArray` flag
3. **Creates text search index** for full-text search capabilities
4. **Verifies import** by counting documents
5. **Shows sample data** to confirm structure

## 🧪 Testing Database Functionality

### Basic Tests
```bash
# Basic functionality test
cd test-client && go run main.go -host your-loadbalancer-ip -port 80

# Count documents
# Health check
```

### Comprehensive Database Tests
```bash
# Enhanced database testing
cd test-client && go run main-db-test.go -host your-loadbalancer-ip -port 80 -db-test
```

### Available Database Operations

#### Document Counting
```bash
# Count all articles
db_count_documents: {"collection": "knowledgebase", "filter": {}}

# Count by priority
db_count_documents: {"collection": "knowledgebase", "filter": {"priority": "high"}}

# Count by category  
db_count_documents: {"collection": "knowledgebase", "filter": {"category": "Security"}}
```

#### Text Search
```bash
# Search for specific topics
db_search_documents: {"collection": "knowledgebase", "search_text": "kubernetes", "limit": 5}
db_search_documents: {"collection": "knowledgebase", "search_text": "database connection", "limit": 3}
db_search_documents: {"collection": "knowledgebase", "search_text": "ssl certificate", "limit": 5}
```

#### Document Queries
```bash
# Query by category
db_query_documents: {"collection": "knowledgebase", "filter": {"category": "Networking"}, "limit": 5}

# Query by priority and difficulty
db_query_documents: {"collection": "knowledgebase", "filter": {"priority": "critical", "difficulty": "hard"}, "limit": 3}

# Query with sorting
db_query_documents: {"collection": "knowledgebase", "filter": {}, "sort": {"views": -1}, "limit": 5}
```

## 📊 Data Statistics

- **Total Articles**: 37
- **Categories**: 15 unique categories
- **Priority Levels**: 
  - Critical: 4 articles
  - High: 14 articles  
  - Medium: 19 articles
- **Difficulty Levels**:
  - Easy: 4 articles
  - Medium: 24 articles
  - Hard: 9 articles
- **Rich Metadata**: Views, helpful votes, creation/update timestamps, resolution times

## 🔍 Search Index Configuration

The text search index is configured with weighted fields:

```javascript
{
  title: 'text',     // Weight: 10 (highest priority)
  tags: 'text',      // Weight: 5
  category: 'text',  // Weight: 3
  content: 'text'    // Weight: 1
}
```

This ensures that matches in titles and tags are prioritized over content matches.

## 🛠️ Known Issues

### ObjectID Decoding
- **Issue**: MongoDB ObjectIDs require special handling in Go
- **Symptom**: `error decoding key _id: decoding an object ID into a string is not supported`
- **Status**: Known limitation, count operations work correctly
- **Workaround**: Use count operations and text search which work properly

### Text Search Requirements
- **Requirement**: Text index must be created before search operations
- **Solution**: The loading script automatically creates the required index
- **Verification**: Use `db_search_documents` to test search functionality

## 🎯 Use Cases

This test data enables testing of:

1. **Technical Support Scenarios**: Real-world IT troubleshooting articles
2. **Search Functionality**: Full-text search across technical content
3. **Categorization**: Organized content by technical domain
4. **Metadata Queries**: Filter by priority, difficulty, views, etc.
5. **Performance Testing**: 37 documents with rich content for load testing
6. **Integration Testing**: End-to-end database operations via MCP protocol

## 📝 Data Schema

Each article contains:

```json
{
  "title": "string",           // Article title
  "content": "string",         // Detailed troubleshooting content
  "category": "string",        // Technical category
  "tags": ["string"],          // Searchable tags
  "priority": "string",        // critical|high|medium|low
  "status": "string",          // published|draft|archived
  "author": "string",          // Team responsible
  "views": number,             // View count
  "helpful_votes": number,     // Usefulness rating
  "created_at": "ISO8601",     // Creation timestamp
  "updated_at": "ISO8601",     // Last update timestamp
  "resolution_time": "string", // Expected resolution time
  "difficulty": "string"       // easy|medium|hard
}
```

This comprehensive test dataset provides a realistic foundation for testing and demonstrating the MCP server's database capabilities!

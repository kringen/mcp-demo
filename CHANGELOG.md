# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-07-26

### Added
- Initial release of Go MCP Server
- Complete Model Context Protocol (MCP) 2024-11-05 implementation
- WebSocket-based MCP server with JSON-RPC 2.0 support
- Mathematical operations tools (add, multiply, divide, power)
- Web search functionality using Colly v2 with DuckDuckGo and Startpage
- MongoDB v2 integration with full CRUD operations
- Comprehensive test suite with unit and integration tests
- Docker support with multi-stage builds
- Health check endpoints for monitoring
- Configuration via command-line flags
- Makefile with development and deployment commands
- golangci-lint configuration for code quality
- Documentation and usage examples

### Technical Details
- **Module**: `github.com/kringen/go-mcp-server`
- **Go Version**: 1.21+
- **MongoDB Driver**: v2.2.2 (latest stable)
- **WebSocket Library**: Gorilla WebSocket v1.5.0
- **Web Scraping**: Colly v2.1.0
- **Testing Framework**: Testify v1.8.4

### Architecture
- Modular design with clear separation of concerns
- Clean interfaces for easy testing and extensibility
- Context-aware operations with proper timeout handling
- Graceful shutdown with signal handling
- Connection pooling and health monitoring
- Rate-limited web scraping with ethical practices

### Available Tools
- **Math Tools**: `add`, `multiply`, `divide`, `power`
- **Search Tools**: `web_search`, `search_health_check`
- **Database Tools**: `db_create_document`, `db_get_document`, `db_update_document`, `db_delete_document`, `db_query_documents`, `db_search_documents`, `db_count_documents`, `db_health_check`

### Configuration
- Server address and port configuration
- MongoDB connection string customization
- Debug mode for detailed logging
- Timeout and connection settings

### Dependencies
- No external runtime dependencies beyond Go standard library
- MongoDB database (can run via Docker)
- Optional: golangci-lint for development
- Optional: Docker for containerized deployment

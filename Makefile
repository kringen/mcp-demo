.PHONY: build run test test-coverage lint clean deps dev docker-build docker-build-publish docker-publish publish build-info k8s-deploy docker-run docker-run-all docker-stop docker-logs docker-admin

# Variables
BINARY_NAME=mcp-server
BUILD_DIR=./bin
CMD_DIR=./cmd/server
DOCKER_IMAGE=mcp-server
DOCKER_TAG=latest
DOCKER_REGISTRY=kringen
BUILD_ID=$(shell git rev-parse --short HEAD 2>/dev/null || echo "dev-$(shell date +%Y%m%d-%H%M%S)")

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOLINT=golangci-lint

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)

# Run the application
run: build
	@echo "Running $(BINARY_NAME)..."
	$(BUILD_DIR)/$(BINARY_NAME)

# Run in development mode with live reload
dev:
	@echo "Running in development mode..."
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Installing air for live reload..."; \
		$(GOGET) -u github.com/cosmtrek/air; \
		air; \
	fi

# Install dependencies
deps:
	@echo "Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) verify

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests for specific module
test-server:
	$(GOTEST) -v ./internal/server/...

test-tools:
	$(GOTEST) -v ./internal/tools/...

test-resources:
	$(GOTEST) -v ./internal/resources/...

test-handlers:
	$(GOTEST) -v ./internal/handlers/...

# Lint the code
lint:
	@echo "Running linter..."
	@if command -v $(GOLINT) > /dev/null; then \
		$(GOLINT) run; \
	else \
		echo "golangci-lint not installed. Installing..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.54.2; \
		$(GOLINT) run; \
	fi

# Format the code
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

# Vet the code
vet:
	@echo "Vetting code..."
	$(GOCMD) vet ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Update dependencies
update-deps:
	@echo "Updating dependencies..."
	$(GOGET) -u ./...
	$(GOMOD) tidy

# Security check
security:
	@echo "Running security checks..."
	@if command -v gosec > /dev/null; then \
		gosec ./...; \
	else \
		echo "Installing gosec..."; \
		$(GOGET) github.com/securecodewarrior/gosec/v2/cmd/gosec; \
		gosec ./...; \
	fi

# Docker build
docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

# Docker build and tag for publishing
docker-build-publish:
	@echo "Building Docker image for publishing..."
	@echo "Build ID: $(BUILD_ID)"
	@echo "Building $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(BUILD_ID) and $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):latest"
	docker build -t $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(BUILD_ID) -t $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):latest .
	@echo "✅ Images built successfully"
	@echo "Verifying images exist..."
	docker images | grep "$(DOCKER_REGISTRY)/$(DOCKER_IMAGE)"

# Docker publish - build and push to registry
docker-publish: 
	@echo "Publishing Docker image to $(DOCKER_REGISTRY)..."
	@echo "Checking Docker Hub login..."
	@if ! docker info | grep -q "Username"; then \
		echo "⚠️  Not logged into Docker Hub. Please run: docker login"; \
		echo "   Or if using a different registry, make sure you're authenticated"; \
	fi
	@echo "Rebuilding images to ensure consistency..."
	$(MAKE) docker-build-publish
	@echo "Pushing latest image first..."
	docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):latest
	@echo "Getting all build ID tags to push..."
	@for tag in $$(docker images $(DOCKER_REGISTRY)/$(DOCKER_IMAGE) --format "{{.Tag}}" | grep -v latest); do \
		echo "Pushing $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$$tag..."; \
		docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$$tag; \
	done
	@echo "✅ Successfully published all images:"
	@docker images $(DOCKER_REGISTRY)/$(DOCKER_IMAGE) --format "  - $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):{{.Tag}}"

# Publish command alias for convenience
publish: docker-publish

# Show current build information
build-info:
	@echo "Build Information:"
	@echo "=================="
	@echo "Binary Name: $(BINARY_NAME)"
	@echo "Docker Image: $(DOCKER_REGISTRY)/$(DOCKER_IMAGE)"
	@echo "Build ID: $(BUILD_ID)"
	@echo "Tags that will be created:"
	@echo "  - $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(BUILD_ID)"
	@echo "  - $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):latest"

# Kubernetes deploy (requires kubectl and cluster connection)
k8s-deploy:
	@echo "Deploying to Kubernetes..."
	@if command -v kubectl > /dev/null; then \
		./k8s/deploy.sh; \
	else \
		echo "kubectl not found. Please install kubectl first."; \
		exit 1; \
	fi

# Docker run - start all services (MongoDB + optional admin interface)
docker-run:
	@echo "Starting core services (MCP server + MongoDB)..."
	docker-compose up -d mongodb mcp-server

# Docker run all - start all services including admin interface
docker-run-all:
	@echo "Starting all services with admin interface..."
	docker-compose --profile admin up -d

# Docker stop - stop all services
docker-stop:
	@echo "Stopping all services..."
	docker-compose down

# Docker logs - show service logs
docker-logs:
	@echo "Showing service logs..."
	docker-compose logs -f

# Docker admin - start services with admin interface
docker-admin:
	@echo "Starting services with admin interface..."
	docker-compose --profile admin up -d

# Docker clean - stop and remove volumes
docker-clean:
	@echo "Stopping services and cleaning volumes..."
	docker-compose down -v

# Legacy aliases for backward compatibility
mongo-up: docker-run
mongo-down: docker-stop
mongo-logs: docker-logs
mongo-admin: docker-admin

# Help
help:
	@echo "Available commands:"
	@echo "  build         - Build the application"
	@echo "  run           - Build and run the application"
	@echo "  dev           - Run in development mode with live reload"
	@echo "  test          - Run all tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  test-server   - Run server module tests"
	@echo "  test-tools    - Run tools module tests"
	@echo "  test-resources- Run resources module tests"
	@echo "  test-handlers - Run handlers module tests"
	@echo "  lint          - Run linter"
	@echo "  fmt           - Format code"
	@echo "  vet           - Vet code"
	@echo "  clean         - Clean build artifacts"
	@echo "  deps          - Install dependencies"
	@echo "  update-deps   - Update dependencies"
	@echo "  security      - Run security checks"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-build-publish - Build Docker image with registry tags"
	@echo "  docker-publish- Build and push Docker image to registry"
	@echo "  publish       - Alias for docker-publish"
	@echo "  build-info    - Show build information and tags"
	@echo "  k8s-deploy    - Deploy to Kubernetes cluster"
	@echo "  docker-run    - Start core services (MCP server + MongoDB)"
	@echo "  docker-run-all- Start all services with admin interface"
	@echo "  docker-stop   - Stop all services"
	@echo "  docker-logs   - Show service logs"
	@echo "  docker-admin  - Start services with admin interface"
	@echo "  docker-clean  - Stop services and remove volumes"
	@echo "  mongo-up      - Alias for docker-run (backward compatibility)"
	@echo "  mongo-down    - Alias for docker-stop (backward compatibility)"
	@echo "  mongo-logs    - Alias for docker-logs (backward compatibility)"
	@echo "  mongo-admin   - Alias for docker-admin (backward compatibility)"
	@echo "  help          - Show this help message"

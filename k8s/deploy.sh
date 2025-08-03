#!/bin/bash

# Kubernetes Deployment Script for MCP Server
echo "🚀 Deploying MCP Server to Kubernetes"
echo "===================================="

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl is not installed or not in PATH"
    exit 1
fi

# Check if we can connect to cluster
if ! kubectl cluster-info &> /dev/null; then
    echo "❌ Cannot connect to Kubernetes cluster"
    echo "   Make sure you're connected to a cluster: kubectl config current-context"
    exit 1
fi

echo "✅ Connected to cluster: $(kubectl config current-context)"

# Check if Docker image exists
echo ""
echo "📦 Checking Docker image..."
REGISTRY_IMAGE="kringen/mcp-server:latest"
LOCAL_IMAGE="mcp-server:latest"

# Try to pull the published image first
if docker pull $REGISTRY_IMAGE 2>/dev/null; then
    echo "✅ Using published image: $REGISTRY_IMAGE"
    USE_PUBLISHED=true
elif docker images | grep -q "mcp-server"; then
    echo "✅ Using local image: $LOCAL_IMAGE"
    USE_PUBLISHED=false
else
    echo "⚠️  Building local Docker image..."
    docker build -t $LOCAL_IMAGE .
    if [ $? -eq 0 ]; then
        echo "✅ Local Docker image built successfully"
        USE_PUBLISHED=false
    else
        echo "❌ Failed to build Docker image"
        exit 1
    fi
fi

# For kind/minikube, load the image
if kubectl config current-context | grep -E "(kind|minikube)" > /dev/null; then
    echo "📤 Loading image to cluster..."
    if [ "$USE_PUBLISHED" = "true" ]; then
        IMAGE_TO_LOAD=$REGISTRY_IMAGE
    else
        IMAGE_TO_LOAD=$LOCAL_IMAGE
    fi
    
    if kubectl config current-context | grep "kind" > /dev/null; then
        kind load docker-image $IMAGE_TO_LOAD
    elif kubectl config current-context | grep "minikube" > /dev/null; then
        minikube image load $IMAGE_TO_LOAD
    fi
fi

# Deploy resources
echo ""
echo "📋 Deploying Kubernetes resources..."

echo "1. Creating namespace and configuration..."
kubectl apply -f k8s/00-namespace-and-config.yaml

echo "2. Deploying MongoDB..."
kubectl apply -f k8s/01-mongodb.yaml

echo "3. Deploying MCP Server..."
kubectl apply -f k8s/02-mcp-server.yaml

echo "4. Deploying MongoDB Express (optional)..."
kubectl apply -f k8s/04-mongo-express.yaml

echo "5. Setting up scaling and policies..."
kubectl apply -f k8s/05-scaling-and-policies.yaml

# Wait for deployments
echo ""
echo "⏳ Waiting for deployments to be ready..."
kubectl wait --for=condition=available --timeout=300s deployment/mongodb -n mcp-server
kubectl wait --for=condition=available --timeout=300s deployment/mcp-server -n mcp-server

# Show status
echo ""
echo "📊 Deployment Status:"
kubectl get pods -n mcp-server
echo ""
kubectl get services -n mcp-server
echo ""

# Get access information
echo "🌐 Access Information:"
echo "===================="

# LoadBalancer access
echo "⏳ Waiting for LoadBalancer to get external IP..."
kubectl wait --for=jsonpath='{.status.loadBalancer.ingress}' --timeout=300s service/mcp-server-loadbalancer -n mcp-server

EXTERNAL_IP=$(kubectl get service mcp-server-loadbalancer -n mcp-server -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
EXTERNAL_HOSTNAME=$(kubectl get service mcp-server-loadbalancer -n mcp-server -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')

if [ -n "$EXTERNAL_IP" ]; then
    echo "🔗 MCP Server (LoadBalancer): ws://$EXTERNAL_IP/mcp"
    echo "🔗 Health Check: http://$EXTERNAL_IP/health"
    echo "🔗 Web Interface: http://$EXTERNAL_IP/"
elif [ -n "$EXTERNAL_HOSTNAME" ]; then
    echo "🔗 MCP Server (LoadBalancer): ws://$EXTERNAL_HOSTNAME/mcp"
    echo "🔗 Health Check: http://$EXTERNAL_HOSTNAME/health"
    echo "🔗 Web Interface: http://$EXTERNAL_HOSTNAME/"
else
    echo "⚠️  LoadBalancer external IP/hostname not yet assigned"
    echo "   Check status with: kubectl get service mcp-server-loadbalancer -n mcp-server"
    echo "   This may take a few minutes depending on your cloud provider"
fi

# Port forwarding option for local development
echo ""
echo "🔧 For local testing, you can also use port forwarding:"
echo "   kubectl port-forward -n mcp-server service/mcp-server-loadbalancer 8080:80"
echo "   Then access: ws://localhost:8080/mcp"

echo ""
echo "🎉 Deployment complete!"
echo ""
echo "📋 Useful commands:"
echo "   kubectl logs -f deployment/mcp-server -n mcp-server"
echo "   kubectl get pods -n mcp-server -w"
echo "   kubectl describe pod <pod-name> -n mcp-server"
echo "   kubectl delete namespace mcp-server  # To remove everything"

# Kubernetes Deployment Guide for MCP Server

This guide explains how to deploy the MCP WebSocket server on Kubernetes with proper WebSocket support.

## 🌐 WebSocket on Kubernetes: Key Considerations

### **Session Affinity (Sticky Sessions)**
WebSocket connections are stateful and need to stay connected to the same pod:
- ✅ **ClientIP session affinity** configured
- ✅ **24-hour timeout** for long-lived connections
- ✅ **Cookie-based affinity** in ingress controller

### **Load Balancer Configuration**
Special annotations for WebSocket support:
- ✅ **Extended timeouts** (3600s) for long connections
- ✅ **WebSocket-specific routing** 
- ✅ **Graceful termination** handling

### **Health Checks**
HTTP health checks work alongside WebSocket:
- ✅ **HTTP /health endpoint** for liveness/readiness
- ✅ **Separate from WebSocket traffic**
- ✅ **Non-disruptive** to active connections

## 🚀 Quick Deployment

### Prerequisites
- Kubernetes cluster (local or cloud)
- `kubectl` configured  

### Option 1: Using Published Image (Recommended)
```bash
# Publish to Docker Hub under kringen account
make publish

# Deploy to Kubernetes (uses published image)
make k8s-deploy
```

### Option 2: One-Command Deployment (Local Build)
```bash
# One-command deployment with local build
./k8s/deploy.sh
```

### Option 3: Manual Deployment
```bash
# Step by step
kubectl apply -f k8s/00-namespace-and-config.yaml
kubectl apply -f k8s/01-mongodb.yaml
kubectl apply -f k8s/02-mcp-server.yaml
kubectl apply -f k8s/04-mongo-express.yaml
kubectl apply -f k8s/05-scaling-and-policies.yaml
```

## 📋 Architecture Overview

```
Client → NodePort Service (Session Affinity) → Pods (WebSocket)
                   ↓
             MongoDB Service → MongoDB Pod
```

### **Components Deployed:**

1. **MCP Server Pods** (2 replicas, auto-scaling 2-10)
   - WebSocket server on port 8080
   - Health endpoint `/health`
   - Session affinity ensures connection persistence

2. **MongoDB Pod** (1 replica with persistent storage)
   - Database for MCP tools
   - 10GB persistent volume
   - Initialization scripts included

3. **MongoDB Express** (optional admin interface)
   - Web UI for database management
   - Restricted network access

4. **NodePort Service**
   - Direct access to pods on port 30800
   - Session affinity with ClientIP
   - 24-hour timeout for long connections

## 🔧 Access Methods

### NodePort (Direct Access)
```bash
# Get node IP and port
kubectl get service mcp-server-nodeport -n mcp-server

# Access via: ws://<node-ip>:30800/mcp
```

### Port Forwarding (Local Testing)
```bash
kubectl port-forward -n mcp-server service/mcp-server-nodeport 8080:8080
# Access via: ws://localhost:8080/mcp
```

## ⚙️ WebSocket-Specific Configuration

### **Service Configuration**
```yaml
spec:
  sessionAffinity: ClientIP
  sessionAffinityConfig:
    clientIP:
      timeoutSeconds: 86400  # 24 hours
```

### **Ingress Annotations**
```yaml
annotations:
  # WebSocket timeouts
  nginx.ingress.kubernetes.io/proxy-read-timeout: "3600"
  nginx.ingress.kubernetes.io/proxy-send-timeout: "3600"
  
  # Session affinity
  nginx.ingress.kubernetes.io/affinity: "cookie"
  nginx.ingress.kubernetes.io/session-cookie-expires: "86400"
```

### **Pod Termination**
```yaml
spec:
  terminationGracePeriodSeconds: 30
  containers:
  - lifecycle:
      preStop:
        exec:
          command: ["/bin/sh", "-c", "sleep 15"]
```

## 📊 Scaling & High Availability

### **Horizontal Pod Autoscaler**
- **Min replicas**: 2 (for HA)
- **Max replicas**: 10 (for load)
- **CPU threshold**: 70%
- **Memory threshold**: 80%

### **Pod Disruption Budget**
- **Min available**: 1 pod always running
- **Graceful rolling updates**

### **Resource Limits**
```yaml
resources:
  requests:
    memory: "256Mi"
    cpu: "100m"
  limits:
    memory: "512Mi"
    cpu: "500m"
```

## 🧪 Testing the Deployment

### Health Check
```bash
# Get service endpoint
kubectl get service -n mcp-server

# Test health
curl http://<service-ip>:8080/health
```

### WebSocket Test
```bash
# Port forward for testing
kubectl port-forward -n mcp-server service/mcp-server-service 8080:8080

# Run test client
cd test-client && go run main.go
```

### Load Testing
```bash
# Scale up for testing
kubectl scale deployment mcp-server --replicas=5 -n mcp-server

# Monitor scaling
kubectl get hpa -n mcp-server -w
```

## 🛠️ Troubleshooting

### **WebSocket Connection Issues**

```bash
# Check pod logs
kubectl logs -f deployment/mcp-server -n mcp-server

# Check ingress logs
kubectl logs -f -n ingress-nginx deployment/ingress-nginx-controller

# Test direct pod connection
kubectl port-forward -n mcp-server <pod-name> 8080:8080
```

### **Session Affinity Problems**

```bash
# Check service configuration
kubectl describe service mcp-server-service -n mcp-server

# Verify ingress annotations
kubectl describe ingress mcp-server-ingress -n mcp-server
```

### **Scaling Issues**

```bash
# Check HPA status
kubectl describe hpa mcp-server-hpa -n mcp-server

# Check resource usage
kubectl top pods -n mcp-server
```

## 🔒 Security Considerations

### **Network Policies**
- Pods can only communicate with MongoDB
- External internet access for web search
- Ingress restricted to necessary ports

### **MongoDB Security**
- Username/password stored in Kubernetes secrets
- Network isolation within cluster
- Optional: Enable MongoDB authentication

### **Ingress Security**
- Rate limiting configured
- Optional TLS termination
- Admin interface IP restrictions

## 🌍 Production Deployment

### **Docker Image Publishing**

The project supports automated image publishing to Docker Hub under the `kringen` account:

```bash
# Build and publish with build ID and latest tags
make publish

# Images created:
# - kringen/mcp-server:<build-id>  (git commit hash or timestamp)
# - kringen/mcp-server:latest
```

**Build ID Generation:**
- **Git repository**: Uses short commit hash (`git rev-parse --short HEAD`)
- **Non-git environment**: Uses timestamp (`YYYYMMDDHHMMSS`)

**Make Commands:**
```bash
make docker-build-publish  # Build with registry tags
make docker-publish        # Build and push to registry  
make publish               # Alias for docker-publish
make k8s-deploy           # Deploy to Kubernetes
```

### **Image Strategy**
The deployment script intelligently selects images:

1. **Published image**: Tries to pull `kringen/mcp-server:latest` first
2. **Local fallback**: Uses locally built `mcp-server:latest` if published unavailable
3. **Auto-build**: Builds image if neither exists

### **Cloud Provider Specifics**

#### **AWS EKS**
```bash
# Use ALB ingress controller
annotations:
  kubernetes.io/ingress.class: alb
  alb.ingress.kubernetes.io/load-balancer-attributes: idle_timeout.timeout_seconds=4000
```

#### **Google GKE**
```bash
# Use GCE ingress controller
annotations:
  kubernetes.io/ingress.class: gce
  cloud.google.com/backend-config: '{"default": "mcp-backendconfig"}'
```

#### **Azure AKS**
```bash
# Use Application Gateway
annotations:
  kubernetes.io/ingress.class: azure/application-gateway
  appgw.ingress.kubernetes.io/connection-draining-timeout: "30"
```

### **Monitoring & Observability**

```yaml
# Add to deployment
spec:
  template:
    metadata:
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8080"
        prometheus.io/path: "/metrics"
```

## 📚 Further Reading

- [Kubernetes Ingress Controllers](https://kubernetes.io/docs/concepts/services-networking/ingress-controllers/)
- [NGINX WebSocket Proxying](https://nginx.org/en/docs/http/websocket.html)
- [Kubernetes Session Affinity](https://kubernetes.io/docs/concepts/services-networking/service/#session-affinity)

## 💡 Tips

1. **Always use session affinity** for WebSocket services
2. **Configure proper timeouts** in ingress controllers  
3. **Test with multiple pods** to verify load balancing
4. **Monitor connection distribution** across pods
5. **Use graceful termination** for clean shutdowns

Your MCP WebSocket server is now production-ready on Kubernetes! 🚀

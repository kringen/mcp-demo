# LoadBalancer vs NodePort for WebSocket Services

## Why LoadBalancer is Better for Production

### 🔒 Security Benefits
- **No direct node access**: Clients don't need to know individual node IPs
- **Cloud provider integration**: Leverages cloud provider's security features
- **Standard ports**: Uses standard ports (80/443) instead of high-numbered NodePorts
- **Network isolation**: Better isolation between cluster internals and external access

### 🌐 Networking Advantages
- **External IP management**: Cloud provider handles IP allocation and DNS
- **Load balancing**: Proper load balancing across pods (not just session affinity)
- **Health checking**: Cloud load balancers perform health checks
- **SSL termination**: Can handle TLS termination at the load balancer level

### 📈 Scalability & Reliability
- **High availability**: Cloud load balancers are highly available
- **Global load balancing**: Some providers support global load balancing
- **Traffic management**: Better traffic shaping and rate limiting
- **Monitoring**: Cloud provider monitoring and logging integration

### 🛠 Operational Benefits
- **Easier maintenance**: No need to update client configurations when nodes change
- **Standard practices**: Follows Kubernetes best practices
- **Cloud integration**: Works seamlessly with cloud provider tools
- **Automatic failover**: Handles node failures gracefully

## LoadBalancer Configuration

### Service Definition
```yaml
apiVersion: v1
kind: Service
metadata:
  name: mcp-server-loadbalancer
  annotations:
    # Cloud-specific annotations for customization
spec:
  type: LoadBalancer
  ports:
  - port: 80          # External port
    targetPort: 8080  # Pod port
  sessionAffinity: ClientIP  # Important for WebSocket persistence
```

### Cloud Provider Annotations

#### AWS (EKS)
```yaml
annotations:
  service.beta.kubernetes.io/aws-load-balancer-type: "nlb"
  service.beta.kubernetes.io/aws-load-balancer-scheme: "internet-facing"
  service.beta.kubernetes.io/aws-load-balancer-backend-protocol: "tcp"
```

#### Azure (AKS)
```yaml
annotations:
  service.beta.kubernetes.io/azure-load-balancer-internal: "false"
  service.beta.kubernetes.io/azure-dns-label-name: "mcp-server"
```

#### Google Cloud (GKE)
```yaml
annotations:
  cloud.google.com/load-balancer-type: "External"
  cloud.google.com/backend-config: '{"default": "websocket-config"}'
```

## WebSocket Considerations

### Session Affinity
- **Required**: WebSocket connections need to stick to the same pod
- **ClientIP**: Use `sessionAffinity: ClientIP` for session persistence
- **Timeout**: Set appropriate timeout for session affinity

### Health Checks
- **HTTP health checks**: Load balancers can use HTTP health endpoints
- **Graceful shutdown**: Ensure pods handle termination gracefully
- **Connection draining**: Allow existing connections to complete

### SSL/TLS (Recommended)
```yaml
ports:
- port: 443
  targetPort: 8080
  name: websocket-tls
```

## Migration from NodePort

### 1. Update Service
```bash
# Apply the new LoadBalancer service
kubectl apply -f k8s/02-mcp-server.yaml

# Remove old NodePort service if needed
kubectl delete service mcp-server-nodeport -n mcp-server
```

### 2. Wait for External IP
```bash
# Watch for external IP assignment
kubectl get service mcp-server-loadbalancer -n mcp-server -w

# Check LoadBalancer status
kubectl describe service mcp-server-loadbalancer -n mcp-server
```

### 3. Update Client Applications
```bash
# Old NodePort access
ws://node-ip:30800/mcp

# New LoadBalancer access  
ws://external-ip/mcp
# or with hostname
ws://mcp-server.example.com/mcp
```

### 4. DNS Configuration (Optional)
```bash
# Create DNS A record pointing to LoadBalancer IP
mcp-server.example.com -> LoadBalancer-External-IP

# Or use cloud provider DNS features
# AWS: Route53 alias records
# Azure: DNS zones
# GCP: Cloud DNS
```

## Troubleshooting

### LoadBalancer Stuck in Pending
```bash
# Check events
kubectl describe service mcp-server-loadbalancer -n mcp-server

# Common issues:
# - Cloud provider limits/quotas
# - Invalid annotations
# - Networking policy restrictions
# - Service account permissions
```

### Connection Issues
```bash
# Test external connectivity
curl -I http://external-ip/health

# Check pod logs
kubectl logs -f deployment/mcp-server -n mcp-server

# Verify session affinity
kubectl get service mcp-server-loadbalancer -n mcp-server -o yaml
```

## Cost Considerations

### Cloud Provider Costs
- **AWS**: Network Load Balancer charges
- **Azure**: Load Balancer resource costs
- **GCP**: Load Balancer forwarding rules
- **Traffic**: Data transfer costs

### Optimization
- Use internal LoadBalancer for cluster-internal services
- Consider regional vs global load balancing
- Monitor and optimize based on usage patterns

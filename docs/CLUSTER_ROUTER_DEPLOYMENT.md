# 🚀 Cluster Router Deployment Guide

## 📋 Table of Contents

1. [Quick Start](#quick-start)
2. [Docker Deployment](#docker-deployment)
3. [Kubernetes Deployment](#kubernetes-deployment)
4. [Configuration](#configuration)
5. [Monitoring Setup](#monitoring-setup)
6. [Production Checklist](#production-checklist)
7. [Troubleshooting](#troubleshooting)

---

## 🚀 Quick Start

### Prerequisites
- Docker and Docker Compose
- Go 1.21+ (for development)
- 4GB RAM minimum
- 2 CPU cores minimum

### Basic Docker Deployment

```bash
# Clone the repository
git clone https://github.com/palaseus/adrenochain.git
cd adrenochain

# Start the cluster router with monitoring
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f cluster-router
```

### Verify Installation

```bash
# Check health
curl http://localhost:8080/health

# Check metrics
curl http://localhost:9090/metrics

# List clusters
curl http://localhost:8080/api/v1/clusters
```

---

## 🐳 Docker Deployment

### Dockerfile

```dockerfile
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o cluster-router ./cmd/cluster-router

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates curl

WORKDIR /root/

# Copy binary and configs
COPY --from=builder /app/cluster-router .
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/scripts ./scripts

# Create non-root user
RUN adduser -D -s /bin/sh cluster-router
USER cluster-router

# Expose ports
EXPOSE 8080 9090

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

# Start the application
CMD ["./cluster-router", "--config", "/root/configs/production.yaml"]
```

### Docker Compose

```yaml
version: '3.8'

services:
  cluster-router:
    build: .
    container_name: adrenochain-cluster-router
    ports:
      - "8080:8080"  # API Gateway
      - "9090:9090"  # Metrics
    environment:
      - CONFIG_PATH=/root/configs/production.yaml
      - LOG_LEVEL=info
      - NODE_ENV=production
    volumes:
      - ./configs:/root/configs:ro
      - ./logs:/root/logs
      - cluster-router-data:/root/data
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    networks:
      - adrenochain-network
    depends_on:
      - prometheus
      - grafana

  prometheus:
    image: prom/prometheus:v2.45.0
    container_name: adrenochain-prometheus
    ports:
      - "9091:9090"
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml:ro
      - prometheus-data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'
      - '--storage.tsdb.retention.time=30d'
      - '--web.enable-lifecycle'
    restart: unless-stopped
    networks:
      - adrenochain-network

  grafana:
    image: grafana/grafana:10.0.0
    container_name: adrenochain-grafana
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin123
      - GF_USERS_ALLOW_SIGN_UP=false
      - GF_INSTALL_PLUGINS=grafana-piechart-panel
    volumes:
      - grafana-data:/var/lib/grafana
      - ./monitoring/grafana/dashboards:/etc/grafana/provisioning/dashboards:ro
      - ./monitoring/grafana/datasources:/etc/grafana/provisioning/datasources:ro
    restart: unless-stopped
    networks:
      - adrenochain-network
    depends_on:
      - prometheus

  redis:
    image: redis:7-alpine
    container_name: adrenochain-redis
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data
    restart: unless-stopped
    networks:
      - adrenochain-network
    command: redis-server --appendonly yes

volumes:
  cluster-router-data:
  prometheus-data:
  grafana-data:
  redis-data:

networks:
  adrenochain-network:
    driver: bridge
```

### Environment Configuration

```bash
# .env file
NODE_ENV=production
LOG_LEVEL=info
CONFIG_PATH=/root/configs/production.yaml

# Database
REDIS_URL=redis://redis:6379

# Monitoring
PROMETHEUS_URL=http://prometheus:9090
GRAFANA_URL=http://grafana:3000

# Security
JWT_SECRET=your-jwt-secret-here
API_KEY=your-api-key-here
```

---

## ☸️ Kubernetes Deployment

### Namespace

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: adrenochain
  labels:
    name: adrenochain
```

### ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cluster-router-config
  namespace: adrenochain
data:
  production.yaml: |
    cluster_router:
      max_clusters: 1000
      max_nodes_per_cluster: 100
      health_check_interval: 30s
      load_update_interval: 10s
      routing_strategy: "adaptive"
      enable_failover: true
      enable_load_balancing: true
      enable_metrics: true
      max_retries: 3
      timeout: 30s

    health_monitor:
      check_interval: 30s
      timeout: 5s
      recovery_threshold: 3
      failure_threshold: 5
      enable_history: true
      max_history_size: 100

    discovery:
      enable_mdns: false
      enable_dns: true
      enable_bootstrap: true
      enable_broadcast: false
      discovery_interval: 60s
      bootstrap_peers:
        - "cluster-router-service:8080"
      dns_seeds:
        - "seed1.adrenochain.com"
      service_name: "adrenochain-cluster"
      timeout: 10s

    api_gateway:
      listen_addr: "0.0.0.0"
      port: 8080
      enable_cors: true
      enable_auth: true
      rate_limit: 1000
      timeout: 30s
      enable_metrics: true
      enable_health: true

    logging:
      level: "info"
      format: "json"
      output: "stdout"

    metrics:
      enable_prometheus: true
      prometheus_port: 9090
      collect_interval: 10s
```

### Secret

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: cluster-router-secrets
  namespace: adrenochain
type: Opaque
data:
  jwt-secret: <base64-encoded-jwt-secret>
  api-key: <base64-encoded-api-key>
```

### Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cluster-router
  namespace: adrenochain
  labels:
    app: cluster-router
spec:
  replicas: 3
  selector:
    matchLabels:
      app: cluster-router
  template:
    metadata:
      labels:
        app: cluster-router
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "9090"
        prometheus.io/path: "/metrics"
    spec:
      containers:
      - name: cluster-router
        image: adrenochain/cluster-router:latest
        imagePullPolicy: Always
        ports:
        - containerPort: 8080
          name: api
          protocol: TCP
        - containerPort: 9090
          name: metrics
          protocol: TCP
        env:
        - name: CONFIG_PATH
          value: "/etc/config/production.yaml"
        - name: LOG_LEVEL
          value: "info"
        - name: NODE_ENV
          value: "production"
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: cluster-router-secrets
              key: jwt-secret
        - name: API_KEY
          valueFrom:
            secretKeyRef:
              name: cluster-router-secrets
              key: api-key
        volumeMounts:
        - name: config
          mountPath: /etc/config
          readOnly: true
        - name: logs
          mountPath: /var/log/cluster-router
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
            scheme: HTTP
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
            scheme: HTTP
          initialDelaySeconds: 5
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 3
        resources:
          requests:
            memory: "512Mi"
            cpu: "500m"
          limits:
            memory: "1Gi"
            cpu: "1000m"
        securityContext:
          runAsNonRoot: true
          runAsUser: 1000
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: true
          capabilities:
            drop:
            - ALL
      volumes:
      - name: config
        configMap:
          name: cluster-router-config
      - name: logs
        emptyDir: {}
      securityContext:
        fsGroup: 1000
      restartPolicy: Always
```

### Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: cluster-router-service
  namespace: adrenochain
  labels:
    app: cluster-router
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/port: "9090"
    prometheus.io/path: "/metrics"
spec:
  selector:
    app: cluster-router
  ports:
  - name: api
    port: 8080
    targetPort: 8080
    protocol: TCP
  - name: metrics
    port: 9090
    targetPort: 9090
    protocol: TCP
  type: ClusterIP
```

### Ingress

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: cluster-router-ingress
  namespace: adrenochain
  annotations:
    kubernetes.io/ingress.class: "nginx"
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/force-ssl-redirect: "true"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  tls:
  - hosts:
    - cluster-router.adrenochain.com
    secretName: cluster-router-tls
  rules:
  - host: cluster-router.adrenochain.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: cluster-router-service
            port:
              number: 8080
```

### HorizontalPodAutoscaler

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: cluster-router-hpa
  namespace: adrenochain
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: cluster-router
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 10
        periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 60
      policies:
      - type: Percent
        value: 50
        periodSeconds: 60
```

---

## 📊 Monitoring Setup

### Prometheus Configuration

```yaml
# monitoring/prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "rules/*.yml"

alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - alertmanager:9093

scrape_configs:
  - job_name: 'cluster-router'
    static_configs:
      - targets: ['cluster-router:9090']
    scrape_interval: 10s
    metrics_path: /metrics
    scheme: http

  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']

  - job_name: 'grafana'
    static_configs:
      - targets: ['grafana:3000']
```

### Grafana Datasource

```yaml
# monitoring/grafana/datasources/prometheus.yml
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
    editable: true
```

### Grafana Dashboard

```json
{
  "dashboard": {
    "id": null,
    "title": "Cluster Router Monitoring",
    "tags": ["adrenochain", "cluster-router"],
    "timezone": "browser",
    "panels": [
      {
        "id": 1,
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(cluster_router_requests_total[5m])",
            "legendFormat": "{{cluster_id}} - {{status}}",
            "refId": "A"
          }
        ],
        "yAxes": [
          {
            "label": "Requests/sec",
            "min": 0
          }
        ],
        "xAxis": {
          "mode": "time"
        }
      },
      {
        "id": 2,
        "title": "Response Time",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(cluster_router_request_duration_seconds_bucket[5m]))",
            "legendFormat": "95th percentile",
            "refId": "A"
          },
          {
            "expr": "histogram_quantile(0.50, rate(cluster_router_request_duration_seconds_bucket[5m]))",
            "legendFormat": "50th percentile",
            "refId": "B"
          }
        ],
        "yAxes": [
          {
            "label": "Seconds",
            "min": 0
          }
        ]
      },
      {
        "id": 3,
        "title": "Node Health Status",
        "type": "table",
        "targets": [
          {
            "expr": "cluster_router_node_health_score",
            "format": "table",
            "refId": "A"
          }
        ]
      },
      {
        "id": 4,
        "title": "Active Connections",
        "type": "singlestat",
        "targets": [
          {
            "expr": "sum(cluster_router_active_connections)",
            "refId": "A"
          }
        ],
        "valueName": "current"
      }
    ],
    "time": {
      "from": "now-1h",
      "to": "now"
    },
    "refresh": "30s"
  }
}
```

### Alert Rules

```yaml
# monitoring/prometheus/rules/cluster-router.yml
groups:
- name: cluster-router
  rules:
  - alert: HighErrorRate
    expr: rate(cluster_router_requests_total{status="error"}[5m]) > 0.1
    for: 2m
    labels:
      severity: warning
    annotations:
      summary: "High error rate detected"
      description: "Error rate is {{ $value }} errors per second"

  - alert: HighResponseTime
    expr: histogram_quantile(0.95, rate(cluster_router_request_duration_seconds_bucket[5m])) > 1
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "High response time detected"
      description: "95th percentile response time is {{ $value }} seconds"

  - alert: NodeDown
    expr: cluster_router_node_health_score < 0.5
    for: 1m
    labels:
      severity: critical
    annotations:
      summary: "Node is down"
      description: "Node {{ $labels.node_id }} health score is {{ $value }}"

  - alert: ClusterDegraded
    expr: cluster_router_cluster_health_score < 0.7
    for: 2m
    labels:
      severity: warning
    annotations:
      summary: "Cluster is degraded"
      description: "Cluster {{ $labels.cluster_id }} health score is {{ $value }}"
```

---

## ✅ Production Checklist

### Pre-Deployment

- [ ] **Configuration Review**
  - [ ] Validate all configuration files
  - [ ] Set appropriate resource limits
  - [ ] Configure security settings
  - [ ] Set up monitoring and alerting

- [ ] **Security**
  - [ ] Generate secure JWT secrets
  - [ ] Configure API keys
  - [ ] Set up TLS certificates
  - [ ] Enable authentication
  - [ ] Configure rate limiting

- [ ] **Monitoring**
  - [ ] Set up Prometheus
  - [ ] Configure Grafana dashboards
  - [ ] Set up alert rules
  - [ ] Test alert notifications

### Deployment

- [ ] **Infrastructure**
  - [ ] Provision sufficient resources
  - [ ] Set up load balancers
  - [ ] Configure DNS
  - [ ] Set up backup systems

- [ ] **Application**
  - [ ] Deploy cluster router
  - [ ] Register initial clusters
  - [ ] Configure health checks
  - [ ] Test failover mechanisms

### Post-Deployment

- [ ] **Verification**
  - [ ] Health checks passing
  - [ ] Metrics collection working
  - [ ] Alerts configured
  - [ ] Load testing completed

- [ ] **Documentation**
  - [ ] Update runbooks
  - [ ] Document procedures
  - [ ] Train operations team
  - [ ] Set up on-call rotation

---

## 🔧 Troubleshooting

### Common Issues

#### 1. Container Won't Start

```bash
# Check container logs
docker logs adrenochain-cluster-router

# Check configuration
docker exec adrenochain-cluster-router cat /root/configs/production.yaml

# Verify health endpoint
curl -v http://localhost:8080/health
```

#### 2. High Memory Usage

```bash
# Check memory metrics
curl http://localhost:9090/metrics | grep go_memstats

# Check cluster/node counts
curl http://localhost:8080/api/v1/clusters | jq '. | length'

# Monitor resource usage
docker stats adrenochain-cluster-router
```

#### 3. Connection Issues

```bash
# Check network connectivity
docker exec adrenochain-cluster-router ping -c 3 8.8.8.8

# Check port binding
netstat -tlnp | grep :8080

# Test internal connectivity
docker exec adrenochain-cluster-router curl -f http://localhost:8080/health
```

### Debug Commands

```bash
# Get cluster status
curl -s http://localhost:8080/api/v1/clusters | jq '.[] | {id, status, load, health_score}'

# Get node status
curl -s http://localhost:8080/api/v1/nodes | jq '.[] | {id, status, load, health_score}'

# Get health status
curl -s http://localhost:8080/api/v1/health/clusters | jq '.'

# Get metrics
curl -s http://localhost:9090/metrics | grep cluster_router

# Check logs
docker-compose logs -f cluster-router
```

### Performance Tuning

```yaml
# Optimize for high throughput
cluster_router:
  max_clusters: 2000
  max_nodes_per_cluster: 200
  health_check_interval: 60s  # Reduce frequency
  load_update_interval: 5s    # Increase frequency
  routing_strategy: "least_load"
  max_retries: 2              # Reduce retries
  timeout: 15s                # Reduce timeout

# Optimize for low latency
cluster_router:
  health_check_interval: 15s  # Increase frequency
  load_update_interval: 1s    # Real-time updates
  routing_strategy: "least_latency"
  max_retries: 1
  timeout: 5s
```

---

## 📚 Additional Resources

- [Architecture Guide](./CLUSTER_ROUTER_ARCHITECTURE.md)
- [API Reference](./CLUSTER_ROUTER_API.md)
- [Configuration Guide](./CONFIGURATION.md)
- [Troubleshooting Guide](./TROUBLESHOOTING.md)

---

*This deployment guide provides comprehensive instructions for deploying the Adrenochain Cluster Router in various environments. For additional support, refer to the troubleshooting section or contact the development team.*

# GoChain Deployment Guide

## Overview

This guide covers deploying GoChain in various environments, from local development to production systems.

## Prerequisites

- Go 1.21 or later
- At least 1GB RAM
- 10GB+ disk space for blockchain data
- Network access for peer communication

## Quick Start

### 1. Build the Binary

```bash
# Clone the repository
git clone https://github.com/yourusername/gochain.git
cd gochain

# Build the binary
make build

# Or manually
go build -o gochain ./cmd/gochain
```

### 2. Run GoChain

```bash
# Basic run
./gochain

# With custom data directory
./gochain -data-dir ./blockchain-data

# With custom port
./gochain -port 8334

# With custom peers
./gochain -peers "192.168.1.100:8333,192.168.1.101:8333"
```

## Configuration

### Configuration File

Create a `config.yaml` file in your data directory:

```yaml
# GoChain Configuration
network:
  port: 8333
  peers:
    - "192.168.1.100:8333"
    - "192.168.1.101:8333"
  max_connections: 50
  timeout: 30s

blockchain:
  data_dir: "./data"
  genesis_block: true
  difficulty: 1000000
  block_time: 10m

mining:
  enabled: true
  threads: 4
  reward_address: "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"

storage:
  engine: "leveldb"
  cache_size: "256MB"
  write_buffer_size: "64MB"

wallet:
  encryption_enabled: true
  key_derivation_iterations: 100000
  backup_enabled: true
  backup_interval: "24h"

logging:
  level: "info"
  format: "json"
  output: "stdout"
  file: "./logs/gochain.log"
  max_size: "100MB"
  max_age: "30d"
  max_backups: 10
```

### Environment Variables

You can also use environment variables for configuration:

```bash
export GOCHAIN_PORT=8333
export GOCHAIN_DATA_DIR="./data"
export GOCHAIN_PEERS="192.168.1.100:8333,192.168.1.101:8333"
export GOCHAIN_DIFFICULTY=1000000
export GOCHAIN_MINING_ENABLED=true
export GOCHAIN_LOG_LEVEL=info
```

## Deployment Scenarios

### 1. Local Development

```bash
# Simple local run
make run

# With custom configuration
./gochain -config ./config-dev.yaml

# Enable debug logging
./gochain -log-level debug
```

### 2. Single Node Production

```bash
# Create service user
sudo useradd -r -s /bin/false gochain

# Create directories
sudo mkdir -p /var/lib/gochain
sudo mkdir -p /var/log/gochain
sudo chown gochain:gochain /var/lib/gochain /var/log/gochain

# Copy binary
sudo cp gochain /usr/local/bin/
sudo chown gochain:gochain /usr/local/bin/gochain

# Create systemd service
sudo tee /etc/systemd/system/gochain.service > /dev/null <<EOF
[Unit]
Description=GoChain Node
After=network.target

[Service]
Type=simple
User=gochain
Group=gochain
ExecStart=/usr/local/bin/gochain -config /etc/gochain/config.yaml
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable gochain
sudo systemctl start gochain

# Check status
sudo systemctl status gochain
```

### 3. Multi-Node Cluster

#### Node 1 (Bootstrap Node)

```bash
# Start bootstrap node
./gochain -port 8333 -data-dir ./node1-data -mining-enabled true
```

#### Node 2

```bash
# Start second node
./gochain -port 8334 -data-dir ./node2-data -peers "localhost:8333"
```

#### Node 3

```bash
# Start third node
./gochain -port 8335 -data-dir ./node3-data -peers "localhost:8333,localhost:8334"
```

### 4. Docker Deployment

#### Dockerfile

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gochain ./cmd/gochain

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/gochain .
COPY --from=builder /app/config.yaml .

EXPOSE 8333
CMD ["./gochain"]
```

#### Docker Compose

```yaml
version: '3.8'

services:
  gochain:
    build: .
    ports:
      - "8333:8333"
    volumes:
      - gochain-data:/data
      - ./config.yaml:/root/config.yaml
    environment:
      - GOCHAIN_DATA_DIR=/data
    restart: unless-stopped

volumes:
  gochain-data:
```

#### Run with Docker

```bash
# Build and run
docker-compose up -d

# Check logs
docker-compose logs -f gochain

# Stop
docker-compose down
```

### 5. Kubernetes Deployment

#### ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: gochain-config
data:
  config.yaml: |
    network:
      port: 8333
      max_connections: 50
    blockchain:
      data_dir: "/data"
      difficulty: 1000000
    mining:
      enabled: true
      threads: 4
```

#### Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gochain
spec:
  replicas: 3
  selector:
    matchLabels:
      app: gochain
  template:
    metadata:
      labels:
        app: gochain
    spec:
      containers:
      - name: gochain
        image: gochain:latest
        ports:
        - containerPort: 8333
        volumeMounts:
        - name: config
          mountPath: /root/config.yaml
          subPath: config.yaml
        - name: data
          mountPath: /data
        resources:
          requests:
            memory: "512Mi"
            cpu: "250m"
          limits:
            memory: "1Gi"
            cpu: "500m"
      volumes:
      - name: config
        configMap:
          name: gochain-config
      - name: data
        persistentVolumeClaim:
          claimName: gochain-pvc
```

#### Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: gochain-service
spec:
  selector:
    app: gochain
  ports:
  - port: 8333
    targetPort: 8333
  type: ClusterIP
```

## Monitoring and Logging

### Health Checks

```bash
# Check node status
curl http://localhost:8333/health

# Check blockchain info
curl http://localhost:8333/api/v1/blockchain/info

# Check peer connections
curl http://localhost:8333/api/v1/network/peers
```

### Log Management

```bash
# View logs
tail -f /var/log/gochain/gochain.log

# Search for errors
grep "ERROR" /var/log/gochain/gochain.log

# Monitor in real-time
journalctl -u gochain -f
```

### Metrics Collection

GoChain exposes Prometheus metrics at `/metrics`:

```bash
# Scrape metrics
curl http://localhost:8333/metrics

# Prometheus configuration
scrape_configs:
  - job_name: 'gochain'
    static_configs:
      - targets: ['localhost:8333']
    metrics_path: /metrics
```

## Security Considerations

### Firewall Configuration

```bash
# Allow GoChain port
sudo ufw allow 8333/tcp

# Allow specific peer IPs
sudo ufw allow from 192.168.1.100 to any port 8333

# Check firewall status
sudo ufw status
```

### SSL/TLS Configuration

```bash
# Generate certificates
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes

# Run with TLS
./gochain -tls-cert cert.pem -tls-key key.pem
```

### Access Control

```bash
# Restrict file permissions
chmod 600 config.yaml
chmod 700 data/
chown gochain:gochain data/

# Use non-root user
sudo -u gochain ./gochain
```

## Backup and Recovery

### Automated Backups

```bash
#!/bin/bash
# backup.sh
BACKUP_DIR="/backup/gochain"
DATA_DIR="/var/lib/gochain"
DATE=$(date +%Y%m%d_%H%M%S)

# Stop service
sudo systemctl stop gochain

# Create backup
tar -czf "$BACKUP_DIR/gochain_$DATE.tar.gz" -C "$DATA_DIR" .

# Start service
sudo systemctl start gochain

# Clean old backups (keep last 7 days)
find "$BACKUP_DIR" -name "gochain_*.tar.gz" -mtime +7 -delete
```

### Recovery

```bash
# Stop service
sudo systemctl stop gochain

# Restore from backup
tar -xzf gochain_20240101_120000.tar.gz -C /var/lib/gochain/

# Fix permissions
sudo chown -R gochain:gochain /var/lib/gochain

# Start service
sudo systemctl start gochain
```

## Troubleshooting

### Common Issues

#### Node Won't Start

```bash
# Check logs
journalctl -u gochain -n 50

# Check configuration
./gochain -config-check

# Verify permissions
ls -la /var/lib/gochain/
```

#### Connection Issues

```bash
# Check network connectivity
telnet localhost 8333

# Check firewall
sudo ufw status

# Verify peer configuration
grep "peers" config.yaml
```

#### Performance Issues

```bash
# Check resource usage
htop
iotop

# Monitor disk I/O
iostat -x 1

# Check memory usage
free -h
```

### Debug Mode

```bash
# Enable debug logging
./gochain -log-level debug -log-format json

# Enable profiling
./gochain -profile-cpu cpu.prof -profile-mem mem.prof

# Analyze profiles
go tool pprof cpu.prof
go tool pprof mem.prof
```

## Performance Tuning

### System Tuning

```bash
# Increase file descriptor limits
echo "* soft nofile 65536" >> /etc/security/limits.conf
echo "* hard nofile 65536" >> /etc/security/limits.conf

# Optimize disk I/O
echo 'vm.dirty_ratio = 15' >> /etc/sysctl.conf
echo 'vm.dirty_background_ratio = 5' >> /etc/sysctl.conf

# Apply changes
sysctl -p
```

### GoChain Tuning

```yaml
# config.yaml optimizations
storage:
  cache_size: "1GB"
  write_buffer_size: "128MB"
  max_open_files: 1000

network:
  max_connections: 100
  timeout: 60s
  keepalive: 30s

mining:
  threads: 8
  batch_size: 1000
```

## Scaling Considerations

### Horizontal Scaling

- Deploy multiple nodes behind a load balancer
- Use consistent hashing for request distribution
- Implement proper peer discovery and synchronization

### Vertical Scaling

- Increase CPU cores for mining operations
- Add more RAM for caching and UTXO management
- Use SSD storage for better I/O performance

### Database Scaling

- Consider using distributed databases (CockroachDB, TiDB)
- Implement database sharding for large UTXO sets
- Use read replicas for balance queries

## Maintenance

### Regular Tasks

```bash
# Daily
- Check logs for errors
- Monitor disk space
- Verify peer connections

# Weekly
- Review performance metrics
- Update dependencies
- Backup configuration

# Monthly
- Security updates
- Performance analysis
- Capacity planning
```

### Update Procedures

```bash
# Backup current version
cp gochain gochain.backup

# Stop service
sudo systemctl stop gochain

# Update binary
sudo cp gochain.new /usr/local/bin/gochain

# Start service
sudo systemctl start gochain

# Verify operation
sudo systemctl status gochain
```

---

For additional support, refer to the [API Documentation](API.md) and [Contributing Guide](../CONTRIBUTING.md). 
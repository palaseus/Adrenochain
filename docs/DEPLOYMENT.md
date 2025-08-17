# adrenochain Production Deployment Guide

## 🚀 **Overview**

This guide provides comprehensive instructions for deploying adrenochain in production environments. It covers security hardening, monitoring, scaling, and maintenance procedures.

## 📋 **Prerequisites**

### **System Requirements**
- **OS**: Linux (Ubuntu 20.04+ recommended)
- **CPU**: 4+ cores (8+ recommended for high-traffic)
- **RAM**: 16GB+ (32GB+ recommended)
- **Storage**: 500GB+ SSD (1TB+ recommended)
- **Network**: 100Mbps+ bandwidth

### **Dependencies**
- Go 1.21+
- Docker (optional, for containerized deployment)
- PostgreSQL 13+ (for production database)
- Redis 6+ (for caching and session management)

## 🔒 **Security Hardening**

### **1. Network Security**
```bash
# Configure firewall
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow ssh
sudo ufw allow 8545/tcp  # JSON-RPC
sudo ufw allow 30303/tcp # P2P
sudo ufw enable

# Configure fail2ban
sudo apt install fail2ban
sudo systemctl enable fail2ban
sudo systemctl start fail2ban
```

### **2. Application Security**
```yaml
# config/security.yaml
security:
  enable_tls: true
  enable_auth: true
  api_key_required: true
  rate_limiting: true
  max_connections: 1000
  
  # Admin whitelist
  admin_whitelist:
    - "0x1234567890123456789012345678901234567890"
    - "0x0987654321098765432109876543210987654321"
  
  # Alert thresholds
  alert_thresholds:
    failed_logins: 5
    suspicious_ips: 10
    api_abuse: 100
```

### **3. Cryptographic Configuration**
```yaml
# config/crypto.yaml
crypto:
  key_size: 256
  hash_algorithm: "SHA-256"
  encryption_method: "AES-256-GCM"
  key_rotation_days: 90
```

## 🏗️ **Deployment Methods**

### **Method 1: Binary Deployment**
```bash
# Build production binary
make build-prod

# Create systemd service
sudo tee /etc/systemd/system/adrenochain.service << EOF
[Unit]
Description=adrenochain Node
After=network.target

[Service]
Type=simple
User=adrenochain
WorkingDirectory=/opt/adrenochain
ExecStart=/opt/adrenochain/adrenochain
Restart=always
RestartSec=10
Environment=GOMAXPROCS=8
Environment=GOGC=100

[Install]
WantedBy=multi-user.target
EOF

# Start service
sudo systemctl daemon-reload
sudo systemctl enable adrenochain
sudo systemctl start adrenochain
```

### **Method 2: Docker Deployment**
```yaml
# docker-compose.yml
version: '3.8'
services:
  adrenochain:
    image: adrenochain/adrenochain:latest
    container_name: adrenochain-node
    restart: unless-stopped
    ports:
      - "8545:8545"
      - "30303:30303"
    volumes:
      - ./data:/opt/adrenochain/data
      - ./config:/opt/adrenochain/config
      - ./logs:/opt/adrenochain/logs
    environment:
      - GOMAXPROCS=8
      - GOGC=100
    networks:
      - adrenochain-network

  postgres:
    image: postgres:13
    container_name: adrenochain-db
    restart: unless-stopped
    environment:
      POSTGRES_DB: adrenochain
      POSTGRES_USER: adrenochain
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - adrenochain-network

  redis:
    image: redis:6-alpine
    container_name: adrenochain-cache
    restart: unless-stopped
    command: redis-server --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis_data:/data
    networks:
      - adrenochain-network

networks:
  adrenochain-network:
    driver: bridge

volumes:
  postgres_data:
  redis_data:
```

## 📊 **Monitoring & Observability**

### **1. Metrics Collection**
```yaml
# config/monitoring.yaml
monitoring:
  enable_metrics: true
  metrics_port: 9090
  enable_prometheus: true
  enable_grafana: true
  
  # Health checks
  health_check_interval: 30s
  health_check_timeout: 10s
  
  # Performance metrics
  collect_performance_metrics: true
  collect_network_metrics: true
  collect_storage_metrics: true
```

### **2. Logging Configuration**
```yaml
# config/logging.yaml
logging:
  level: info
  format: json
  output: file
  file_path: /var/log/adrenochain/app.log
  max_size: 100MB
  max_age: 30
  max_backups: 10
  
  # Structured logging
  enable_structured_logging: true
  include_timestamp: true
  include_level: true
  include_caller: true
```

### **3. Alerting Rules**
```yaml
# config/alerts.yaml
alerts:
  # Performance alerts
  high_cpu_usage:
    threshold: 80
    duration: 5m
    severity: warning
    
  high_memory_usage:
    threshold: 85
    duration: 5m
    severity: warning
    
  # Network alerts
  high_latency:
    threshold: 1000ms
    duration: 2m
    severity: critical
    
  # Security alerts
  failed_login_attempts:
    threshold: 5
    duration: 1m
    severity: critical
```

## 🔄 **Scaling & Performance**

### **1. Horizontal Scaling**
```bash
# Load balancer configuration (nginx)
upstream adrenochain_backend {
    least_conn;
    server 192.168.1.10:8545;
    server 192.168.1.11:8545;
    server 192.168.1.12:8545;
}

server {
    listen 80;
    server_name adrenochain.example.com;
    
    location / {
        proxy_pass http://adrenochain_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

### **2. Performance Tuning**
```bash
# System tuning
echo 'net.core.somaxconn = 65535' | sudo tee -a /etc/sysctl.conf
echo 'net.ipv4.tcp_max_syn_backlog = 65535' | sudo tee -a /etc/sysctl.conf
echo 'net.core.netdev_max_backlog = 5000' | sudo tee -a /etc/sysctl.conf
sudo sysctl -p

# Go runtime tuning
export GOMAXPROCS=8
export GOGC=100
export GOMEMLIMIT=8GiB
```

## 🛠️ **Maintenance & Updates**

### **1. Backup Procedures**
```bash
#!/bin/bash
# backup.sh
BACKUP_DIR="/backup/adrenochain"
DATE=$(date +%Y%m%d_%H%M%S)

# Create backup directory
mkdir -p $BACKUP_DIR

# Backup data directory
tar -czf $BACKUP_DIR/data_$DATE.tar.gz /opt/adrenochain/data

# Backup configuration
tar -czf $BACKUP_DIR/config_$DATE.tar.gz /opt/adrenochain/config

# Backup logs
tar -czf $BACKUP_DIR/logs_$DATE.tar.gz /opt/adrenochain/logs

# Clean old backups (keep last 30 days)
find $BACKUP_DIR -name "*.tar.gz" -mtime +30 -delete
```

### **2. Update Procedures**
```bash
#!/bin/bash
# update.sh
set -e

echo "Starting adrenochain update..."

# Stop service
sudo systemctl stop adrenochain

# Backup current version
cp /opt/adrenochain/adrenochain /opt/adrenochain/adrenochain.backup

# Download new version
wget -O /tmp/adrenochain.tar.gz https://github.com/adrenochain/adrenochain/releases/latest/download/adrenochain_linux_amd64.tar.gz

# Extract and install
tar -xzf /tmp/adrenochain.tar.gz -C /tmp
sudo cp /tmp/adrenochain /opt/adrenochain/

# Verify binary
/opt/adrenochain/adrenochain version

# Start service
sudo systemctl start adrenochain

# Check status
sudo systemctl status adrenochain

echo "Update completed successfully!"
```

### **3. Health Check Scripts**
```bash
#!/bin/bash
# health_check.sh
HEALTH_ENDPOINT="http://localhost:8545/health"
MAX_RETRIES=3
RETRY_DELAY=5

for i in $(seq 1 $MAX_RETRIES); do
    if curl -f -s $HEALTH_ENDPOINT > /dev/null; then
        echo "Health check passed"
        exit 0
    else
        echo "Health check failed (attempt $i/$MAX_RETRIES)"
        if [ $i -lt $MAX_RETRIES ]; then
            sleep $RETRY_DELAY
        fi
    fi
done

echo "Health check failed after $MAX_RETRIES attempts"
exit 1
```

## 🚨 **Disaster Recovery**

### **1. Recovery Procedures**
```bash
# Database recovery
pg_restore -d adrenochain /backup/adrenochain/db_20231201_120000.dump

# Data recovery
tar -xzf /backup/adrenochain/data_20231201_120000.tar.gz -C /

# Configuration recovery
tar -xzf /backup/adrenochain/config_20231201_120000.tar.gz -C /
```

### **2. Failover Configuration**
```yaml
# config/failover.yaml
failover:
  enable: true
  primary_node: "192.168.1.10:8545"
  backup_nodes:
    - "192.168.1.11:8545"
    - "192.168.1.12:8545"
  
  health_check_interval: 10s
  failover_threshold: 3
  auto_failback: true
```

## 📈 **Performance Benchmarks**

### **Expected Performance Metrics**
- **Transaction Throughput**: 10,000+ TPS
- **Block Time**: < 15 seconds
- **Network Latency**: < 100ms (local), < 500ms (global)
- **Memory Usage**: < 8GB under normal load
- **CPU Usage**: < 70% under normal load

### **Load Testing**
```bash
# Run performance tests
make test-performance

# Run load tests
make test-load

# Run stress tests
make test-stress
```

## 🔍 **Troubleshooting**

### **Common Issues & Solutions**

#### **1. High Memory Usage**
```bash
# Check memory usage
free -h
ps aux --sort=-%mem | head -10

# Restart service if needed
sudo systemctl restart adrenochain
```

#### **2. Network Connectivity Issues**
```bash
# Check network status
netstat -tulpn | grep 8545
netstat -tulpn | grep 30303

# Check firewall
sudo ufw status
```

#### **3. Database Connection Issues**
```bash
# Check database status
sudo systemctl status postgresql
sudo -u postgres psql -c "SELECT version();"
```

## 📚 **Additional Resources**

- [adrenochain Architecture Documentation](ARCHITECTURE.md)
- [API Reference](API.md)
- [Security Best Practices](SECURITY.md)
- [Performance Tuning Guide](PERFORMANCE.md)

## 🆘 **Support**

For production deployment support:
- **Email**: support@adrenochain.io
- **Documentation**: https://docs.adrenochain.io
- **Community**: https://community.adrenochain.io
- **Emergency**: +1-800-adrenochain

---

**⚠️ Important**: This guide covers production deployment. Always test in staging environments first and ensure proper security measures are in place before deploying to production. 
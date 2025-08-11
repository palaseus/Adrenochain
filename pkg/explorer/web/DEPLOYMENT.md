# GoChain Explorer Deployment Guide

This guide covers deploying the GoChain Explorer web interface to production environments.

## 🚀 **Quick Start Deployment**

### 1. **Docker Deployment (Recommended)**

```bash
# Build the Docker image
docker build -t gochain-explorer .

# Run the container
docker run -d \
  --name gochain-explorer \
  -p 8080:8080 \
  -e EXPLORER_PORT=8080 \
  -e EXPLORER_ENV=production \
  gochain-explorer
```

### 2. **Binary Deployment**

```bash
# Build the binary
go build -o gochain-explorer ./cmd/explorer

# Run the explorer
./gochain-explorer --config config.yaml
```

## 🏗️ **Production Architecture**

### **Recommended Setup**

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Load Balancer │    │   Reverse Proxy │    │  GoChain Nodes │
│   (nginx/HAProxy)│    │   (nginx)       │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │  GoChain        │
                    │  Explorer       │
                    │  (Multiple      │
                    │   Instances)    │
                    └─────────────────┘
                                 │
                    ┌─────────────────┐
                    │  Database       │
                    │  (PostgreSQL/   │
                    │   Redis)        │
                    └─────────────────┘
```

## 📋 **Prerequisites**

### **System Requirements**

- **CPU**: 2+ cores (4+ recommended for production)
- **RAM**: 4GB+ (8GB+ recommended)
- **Storage**: 50GB+ SSD (100GB+ for high-traffic)
- **OS**: Linux (Ubuntu 20.04+ recommended)

### **Software Dependencies**

```bash
# Ubuntu/Debian
sudo apt update
sudo apt install -y \
  curl \
  wget \
  git \
  build-essential \
  ca-certificates \
  nginx \
  redis-server \
  postgresql

# CentOS/RHEL
sudo yum install -y \
  curl \
  wget \
  git \
  gcc \
  make \
  nginx \
  redis \
  postgresql-server
```

## 🔧 **Configuration**

### **Environment Variables**

```bash
# Core Configuration
export EXPLORER_ENV=production
export EXPLORER_PORT=8080
export EXPLORER_HOST=0.0.0.0

# Database Configuration
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=gochain_explorer
export DB_USER=explorer_user
export DB_PASSWORD=secure_password

# Redis Configuration
export REDIS_HOST=localhost
export REDIS_PORT=6379
export REDIS_PASSWORD=secure_redis_password

# Security Configuration
export JWT_SECRET=your_jwt_secret_here
export API_KEY=your_api_key_here
export CORS_ORIGIN=https://yourdomain.com

# Monitoring Configuration
export PROMETHEUS_ENABLED=true
export PROMETHEUS_PORT=9090
export LOG_LEVEL=info
```

### **Configuration File (config.yaml)**

```yaml
# Explorer Configuration
explorer:
  environment: production
  port: 8080
  host: "0.0.0.0"
  
  # Rate Limiting
  rate_limit:
    requests_per_minute: 100
    burst_size: 20
  
  # Caching
  cache:
    enabled: true
    ttl: 300s
    max_size: 1000

# Database Configuration
database:
  host: localhost
  port: 5432
  name: gochain_explorer
  user: explorer_user
  password: secure_password
  ssl_mode: require
  max_connections: 100
  connection_timeout: 30s

# Redis Configuration
redis:
  host: localhost
  port: 6379
  password: secure_redis_password
  db: 0
  pool_size: 10

# Security Configuration
security:
  jwt_secret: your_jwt_secret_here
  api_key: your_api_key_here
  cors_origin: https://yourdomain.com
  rate_limiting: true
  input_validation: true

# Monitoring Configuration
monitoring:
  prometheus_enabled: true
  prometheus_port: 9090
  log_level: info
  metrics_enabled: true
  health_check_interval: 30s

# Web Interface Configuration
web:
  static_files_path: ./static
  templates_path: ./templates
  session_timeout: 3600s
  max_upload_size: 10MB
```

## 🌐 **Reverse Proxy Setup**

### **Nginx Configuration**

```nginx
# /etc/nginx/sites-available/gochain-explorer
upstream explorer_backend {
    server 127.0.0.1:8080;
    server 127.0.0.1:8081;
    server 127.0.0.1:8082;
}

server {
    listen 80;
    server_name explorer.yourdomain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name explorer.yourdomain.com;

    # SSL Configuration
    ssl_certificate /etc/letsencrypt/live/explorer.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/explorer.yourdomain.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;

    # Security Headers
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self';";

    # Rate Limiting
    limit_req_zone $binary_remote_addr zone=explorer:10m rate=10r/s;
    limit_req zone=explorer burst=20 nodelay;

    # Gzip Compression
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_proxied any;
    gzip_comp_level 6;
    gzip_types
        text/plain
        text/css
        text/xml
        text/javascript
        application/json
        application/javascript
        application/xml+rss
        application/atom+xml
        image/svg+xml;

    # Static Files
    location /static/ {
        alias /path/to/explorer/static/;
        expires 1y;
        add_header Cache-Control "public, immutable";
        add_header X-Content-Type-Options nosniff;
    }

    # API Endpoints
    location /api/ {
        limit_req zone=explorer burst=20 nodelay;
        proxy_pass http://explorer_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_connect_timeout 30s;
        proxy_send_timeout 30s;
        proxy_read_timeout 30s;
    }

    # Web Interface
    location / {
        proxy_pass http://explorer_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_connect_timeout 30s;
        proxy_send_timeout 30s;
        proxy_read_timeout 30s;
    }

    # Health Check
    location /health {
        access_log off;
        return 200 "healthy\n";
        add_header Content-Type text/plain;
    }
}
```

### **Enable the Site**

```bash
# Create symlink
sudo ln -s /etc/nginx/sites-available/gochain-explorer /etc/nginx/sites-enabled/

# Test configuration
sudo nginx -t

# Reload nginx
sudo systemctl reload nginx
```

## 🔒 **SSL Certificate Setup**

### **Let's Encrypt (Free SSL)**

```bash
# Install Certbot
sudo apt install certbot python3-certbot-nginx

# Obtain certificate
sudo certbot --nginx -d explorer.yourdomain.com

# Auto-renewal
sudo crontab -e
# Add this line:
0 12 * * * /usr/bin/certbot renew --quiet
```

## 📊 **Monitoring & Logging**

### **Prometheus Metrics**

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'gochain-explorer'
    static_configs:
      - targets: ['localhost:9090']
    metrics_path: /metrics
    scrape_interval: 30s
```

### **Grafana Dashboard**

```json
{
  "dashboard": {
    "title": "GoChain Explorer Metrics",
    "panels": [
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total[5m])",
            "legendFormat": "{{method}} {{endpoint}}"
          }
        ]
      },
      {
        "title": "Response Time",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "95th percentile"
          }
        ]
      }
    ]
  }
}
```

### **Logging Configuration**

```yaml
# logback.xml
<configuration>
  <appender name="FILE" class="ch.qos.logback.core.rolling.RollingFileAppender">
    <file>logs/explorer.log</file>
    <rollingPolicy class="ch.qos.logback.core.rolling.TimeBasedRollingPolicy">
      <fileNamePattern>logs/explorer.%d{yyyy-MM-dd}.log</fileNamePattern>
      <maxHistory>30</maxHistory>
    </rollingPolicy>
    <encoder>
      <pattern>%d{yyyy-MM-dd HH:mm:ss} [%thread] %-5level %logger{36} - %msg%n</pattern>
    </encoder>
  </appender>

  <appender name="STDOUT" class="ch.qos.logback.core.ConsoleAppender">
    <encoder>
      <pattern>%d{HH:mm:ss.SSS} [%thread] %-5level %logger{36} - %msg%n</pattern>
    </encoder>
  </appender>

  <root level="INFO">
    <appender-ref ref="FILE" />
    <appender-ref ref="STDOUT" />
  </root>
</configuration>
```

## 🚀 **Deployment Scripts**

### **Deploy Script**

```bash
#!/bin/bash
# deploy.sh

set -e

echo "🚀 Deploying GoChain Explorer..."

# Pull latest changes
git pull origin main

# Build the application
echo "📦 Building application..."
go build -o gochain-explorer ./cmd/explorer

# Stop existing service
echo "⏹️  Stopping existing service..."
sudo systemctl stop gochain-explorer || true

# Backup current binary
echo "💾 Creating backup..."
sudo cp /usr/local/bin/gochain-explorer /usr/local/bin/gochain-explorer.backup || true

# Install new binary
echo "📥 Installing new binary..."
sudo cp gochain-explorer /usr/local/bin/
sudo chmod +x /usr/local/bin/gochain-explorer

# Start service
echo "▶️  Starting service..."
sudo systemctl start gochain-explorer

# Check status
echo "🔍 Checking service status..."
sudo systemctl status gochain-explorer

echo "✅ Deployment completed successfully!"
```

### **Systemd Service**

```ini
# /etc/systemd/system/gochain-explorer.service
[Unit]
Description=GoChain Explorer
After=network.target postgresql.service redis.service

[Service]
Type=simple
User=explorer
Group=explorer
WorkingDirectory=/opt/gochain-explorer
ExecStart=/usr/local/bin/gochain-explorer --config /opt/gochain-explorer/config.yaml
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=gochain-explorer

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/gochain-explorer/logs /opt/gochain-explorer/data

# Resource limits
LimitNOFILE=65536
LimitNPROC=4096

[Install]
WantedBy=multi-user.target
```

## 🔧 **Maintenance & Updates**

### **Health Check Endpoints**

```bash
# Basic health check
curl http://localhost:8080/health

# Detailed health check
curl http://localhost:8080/health/detailed

# Metrics endpoint
curl http://localhost:9090/metrics
```

### **Backup Script**

```bash
#!/bin/bash
# backup.sh

BACKUP_DIR="/backups/explorer"
DATE=$(date +%Y%m%d_%H%M%S)

echo "💾 Creating backup for $DATE..."

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Backup database
pg_dump gochain_explorer > "$BACKUP_DIR/db_backup_$DATE.sql"

# Backup configuration
cp config.yaml "$BACKUP_DIR/config_$DATE.yaml"

# Backup logs
tar -czf "$BACKUP_DIR/logs_$DATE.tar.gz" logs/

# Clean old backups (keep last 7 days)
find "$BACKUP_DIR" -name "*.sql" -mtime +7 -delete
find "$BACKUP_DIR" -name "*.yaml" -mtime +7 -delete
find "$BACKUP_DIR" -name "*.tar.gz" -mtime +7 -delete

echo "✅ Backup completed: $BACKUP_DIR"
```

## 🚨 **Troubleshooting**

### **Common Issues**

1. **Service won't start**
   ```bash
   # Check logs
   sudo journalctl -u gochain-explorer -f
   
   # Check configuration
   sudo -u explorer gochain-explorer --config-check
   ```

2. **High memory usage**
   ```bash
   # Check memory usage
   ps aux | grep gochain-explorer
   
   # Check for memory leaks
   curl http://localhost:9090/metrics | grep go_memstats
   ```

3. **Database connection issues**
   ```bash
   # Test database connection
   psql -h localhost -U explorer_user -d gochain_explorer -c "SELECT 1"
   
   # Check database status
   sudo systemctl status postgresql
   ```

### **Performance Tuning**

```bash
# Increase file descriptor limits
echo "* soft nofile 65536" >> /etc/security/limits.conf
echo "* hard nofile 65536" >> /etc/security/limits.conf

# Optimize kernel parameters
echo "net.core.somaxconn = 65535" >> /etc/sysctl.conf
echo "net.ipv4.tcp_max_syn_backlog = 65535" >> /etc/sysctl.conf
sysctl -p
```

## 📚 **Additional Resources**

- [GoChain Documentation](https://docs.gochain.io)
- [Nginx Configuration Guide](https://nginx.org/en/docs/)
- [Prometheus Best Practices](https://prometheus.io/docs/practices/)
- [Docker Best Practices](https://docs.docker.com/develop/dev-best-practices/)

## 🆘 **Support**

For deployment issues:
- Check the logs: `sudo journalctl -u gochain-explorer -f`
- Review configuration: `sudo -u explorer gochain-explorer --config-check`
- Open an issue: [GitHub Issues](https://github.com/gochain/gochain/issues)

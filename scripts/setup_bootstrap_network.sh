#!/bin/bash

# 🚀 Adrenochain Bootstrap Network Setup Script
# This script sets up the initial bootstrap network for adrenochain

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BOOTSTRAP_DIR="$PROJECT_ROOT/bootstrap_network"
GENESIS_FILE="$BOOTSTRAP_DIR/genesis.json"
CONFIG_DIR="$BOOTSTRAP_DIR/configs"

echo -e "${BLUE}🚀 Setting up Adrenochain Bootstrap Network...${NC}"

# Create bootstrap directory structure
echo -e "${BLUE}📁 Creating bootstrap network directory structure...${NC}"
mkdir -p "$BOOTSTRAP_DIR"/{configs,data,logs,validators}

# Create bootstrap configuration
echo -e "${BLUE}⚙️  Creating bootstrap network configuration...${NC}"
cat > "$CONFIG_DIR/bootstrap_config.yaml" << 'EOF'
# Bootstrap Network Configuration
network:
  name: "adrenochain-bootstrap"
  chain_id: "adrenochain-1"
  listen_port: 30303
  bootstrap_peers: []
  enable_mdns: true
  max_peers: 10

blockchain:
  genesis_block_reward: 1000000000
  block_time: 10s
  difficulty_adjustment_interval: 2016
  target_block_time: 10s
  max_block_size: 1000000

mining:
  enabled: true
  mining_threads: 2
  coinbase_address: "bootstrap_miner"

storage:
  data_dir: "./bootstrap_data"
  db_type: "leveldb"

logging:
  level: "info"
  format: "text"
  log_file: "./bootstrap_logs/adrenochain.log"

api:
  enabled: true
  listen_addr: "127.0.0.1:8080"
  cors_enabled: true

monitoring:
  enabled: true
  metrics:
    enabled: true
    listen_addr: "127.0.0.1:9090"
  health:
    enabled: true
    listen_addr: "127.0.0.1:8081"
EOF

# Create genesis block generation script
echo -e "${BLUE}🏗️  Creating genesis block generation script...${NC}"
cat > "$BOOTSTRAP_DIR/generate_genesis.sh" << 'EOF'
#!/bin/bash

# Generate genesis block for bootstrap network
echo "🏗️  Generating genesis block..."

# Create initial validator keys
mkdir -p validators
for i in {1..3}; do
    echo "Creating validator $i..."
    go run cmd/gochain/main.go wallet --wallet-file "validators/validator_$i.dat" --passphrase "validator$i"
done

# Generate genesis block
go run cmd/gochain/main.go --generate-genesis --config configs/bootstrap_config.yaml

echo "✅ Genesis block generated successfully!"
EOF

chmod +x "$BOOTSTRAP_DIR/generate_genesis.sh"

# Create validator setup script
echo -e "${BLUE}👥 Creating validator setup script...${NC}"
cat > "$BOOTSTRAP_DIR/setup_validators.sh" << 'EOF'
#!/bin/bash

# Setup validators for bootstrap network
echo "👥 Setting up validators..."

# Create validator configurations
for i in {1..3}; do
    echo "Setting up validator $i..."
    
    # Create validator directory
    mkdir -p "validators/validator_$i"
    
    # Create validator config
    cat > "validators/validator_$i/config.yaml" << EOL
network:
  listen_port: $((30303 + i))
  bootstrap_peers:
    - "/ip4/127.0.0.1/tcp/30303/p2p/bootstrap_node"

mining:
  enabled: true
  mining_threads: 1
  coinbase_address: "validator_$i"

storage:
  data_dir: "./validator_$i_data"
  db_type: "leveldb"

api:
  enabled: true
  listen_addr: "127.0.0.1:$((8080 + i))"

monitoring:
  enabled: true
  metrics:
    listen_addr: "127.0.0.1:$((9090 + i))"
  health:
    listen_addr: "127.0.0.1:$((8081 + i))"
EOL
done

echo "✅ Validators setup completed!"
EOF

chmod +x "$BOOTSTRAP_DIR/setup_validators.sh"

# Create network startup script
echo -e "${BLUE}🚀 Creating network startup script...${NC}"
cat > "$BOOTSTRAP_DIR/start_network.sh" << 'EOF'
#!/bin/bash

# Start bootstrap network
echo "🚀 Starting Adrenochain Bootstrap Network..."

# Start bootstrap node
echo "Starting bootstrap node..."
go run cmd/gochain/main.go --config configs/bootstrap_config.yaml --mining &
BOOTSTRAP_PID=$!

# Wait for bootstrap node to start
sleep 10

# Start validators
for i in {1..3}; do
    echo "Starting validator $i..."
    go run cmd/gochain/main.go --config validators/validator_$i/config.yaml --mining &
    VALIDATOR_PID=$!
    echo $VALIDATOR_PID > "validators/validator_$i/pid"
    sleep 5
done

echo "✅ Bootstrap network started!"
echo "Bootstrap node PID: $BOOTSTRAP_PID"
echo "Network status: http://localhost:8081/health"
echo "Metrics: http://localhost:9090/metrics"

# Create stop script
cat > stop_network.sh << 'EOL'
#!/bin/bash
echo "🛑 Stopping bootstrap network..."
kill $BOOTSTRAP_PID 2>/dev/null || true
for i in {1..3}; do
    if [ -f "validators/validator_$i/pid" ]; then
        kill $(cat "validators/validator_$i/pid") 2>/dev/null || true
    fi
done
echo "✅ Network stopped!"
EOL
chmod +x stop_network.sh
EOF

chmod +x "$BOOTSTRAP_DIR/start_network.sh"

# Create network monitoring script
echo -e "${BLUE}📊 Creating network monitoring script...${NC}"
cat > "$BOOTSTRAP_DIR/monitor_network.sh" << 'EOF'
#!/bin/bash

# Monitor bootstrap network
echo "📊 Adrenochain Bootstrap Network Monitor"
echo "========================================"

while true; do
    clear
    echo "📊 Adrenochain Bootstrap Network Status"
    echo "======================================"
    echo "Timestamp: $(date)"
    echo
    
    # Check bootstrap node
    if curl -s http://localhost:8081/health > /dev/null; then
        echo "✅ Bootstrap Node: Running"
    else
        echo "❌ Bootstrap Node: Down"
    fi
    
    # Check validators
    for i in {1..3}; do
        port=$((8081 + i))
        if curl -s http://localhost:$port/health > /dev/null; then
            echo "✅ Validator $i: Running (port $port)"
        else
            echo "❌ Validator $i: Down (port $port)"
        fi
    done
    
    echo
    echo "Press Ctrl+C to exit"
    sleep 5
done
EOF

chmod +x "$BOOTSTRAP_DIR/monitor_network.sh"

# Create README for bootstrap network
echo -e "${BLUE}📝 Creating bootstrap network documentation...${NC}"
cat > "$BOOTSTRAP_DIR/README.md" << 'EOF'
# Adrenochain Bootstrap Network

This directory contains the bootstrap network setup for adrenochain.

## Quick Start

1. **Generate Genesis Block**
   ```bash
   ./generate_genesis.sh
   ```

2. **Setup Validators**
   ```bash
   ./setup_validators.sh
   ```

3. **Start Network**
   ```bash
   ./start_network.sh
   ```

4. **Monitor Network**
   ```bash
   ./monitor_network.sh
   ```

5. **Stop Network**
   ```bash
   ./stop_network.sh
   ```

## Network Architecture

- **Bootstrap Node**: Port 30303 (P2P), 8080 (API), 8081 (Health), 9090 (Metrics)
- **Validator 1**: Port 30304 (P2P), 8081 (API), 8082 (Health), 9091 (Metrics)
- **Validator 2**: Port 30305 (P2P), 8082 (API), 8083 (Health), 9092 (Metrics)
- **Validator 3**: Port 30306 (P2P), 8083 (API), 8084 (Health), 9093 (Metrics)

## Health Checks

- Bootstrap Node: http://localhost:8081/health
- Validator 1: http://localhost:8082/health
- Validator 2: http://localhost:8083/health
- Validator 3: http://localhost:8084/health

## Metrics

- Bootstrap Node: http://localhost:9090/metrics
- Validator 1: http://localhost:9091/metrics
- Validator 2: http://localhost:9092/metrics
- Validator 3: http://localhost:9093/metrics

## Configuration

All configuration files are in the `configs/` directory. Each validator has its own configuration in `validators/validator_X/config.yaml`.
EOF

echo -e "${GREEN}✅ Bootstrap network setup completed!${NC}"
echo -e "${BLUE}📁 Bootstrap network directory: $BOOTSTRAP_DIR${NC}"
echo -e "${BLUE}📝 Next steps:${NC}"
echo -e "   1. cd $BOOTSTRAP_DIR"
echo -e "   2. ./generate_genesis.sh"
echo -e "   3. ./setup_validators.sh"
echo -e "   4. ./start_network.sh"
echo -e "   5. ./monitor_network.sh"

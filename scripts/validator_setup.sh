#!/bin/bash

# 👥 Adrenochain Validator Setup Script
# This script sets up validators for the adrenochain network

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
VALIDATORS_DIR="$PROJECT_ROOT/validators"
DEFAULT_VALIDATOR_COUNT=3

echo -e "${BLUE}👥 Setting up Adrenochain Validators...${NC}"

# Get validator count from user
read -p "Number of validators to create (default: $DEFAULT_VALIDATOR_COUNT): " VALIDATOR_COUNT
VALIDATOR_COUNT=${VALIDATOR_COUNT:-$DEFAULT_VALIDATOR_COUNT}

# Create validators directory
echo -e "${BLUE}📁 Creating validators directory...${NC}"
mkdir -p "$VALIDATORS_DIR"

# Create validator setup function
setup_validator() {
    local validator_id=$1
    local validator_dir="$VALIDATORS_DIR/validator_$validator_id"
    
    echo -e "${BLUE}🔧 Setting up validator $validator_id...${NC}"
    
    # Create validator directory
    mkdir -p "$validator_dir"/{config,data,logs,keys}
    
    # Create validator configuration
    cat > "$validator_dir/config/validator_config.yaml" << EOF
# Validator $validator_id Configuration
network:
  name: "adrenochain-validator-$validator_id"
  listen_port: $((30303 + validator_id))
  bootstrap_peers:
    - "/ip4/127.0.0.1/tcp/30303/p2p/bootstrap_node"
  max_peers: 50
  enable_mdns: true

blockchain:
  genesis_block_reward: 1000000000
  block_time: 10s
  difficulty_adjustment_interval: 2016
  target_block_time: 10s
  max_block_size: 1000000

mining:
  enabled: true
  mining_threads: 2
  coinbase_address: "validator_$validator_id"
  coinbase_reward: 1000000000

storage:
  data_dir: "./data"
  db_type: "leveldb"

logging:
  level: "info"
  format: "text"
  log_file: "./logs/validator_$validator_id.log"
  max_size: 104857600  # 100MB
  max_backups: 5

api:
  enabled: true
  listen_addr: "127.0.0.1:$((8080 + validator_id))"
  cors_enabled: true
  rate_limit: 1000

monitoring:
  enabled: true
  metrics:
    enabled: true
    listen_addr: "127.0.0.1:$((9090 + validator_id))"
    collect_interval: 15s
    prometheus_enabled: true
  health:
    enabled: true
    listen_addr: "127.0.0.1:$((8081 + validator_id))"
    check_interval: 15s

# Validator specific settings
validator:
  name: "validator_$validator_id"
  commission_rate: 0.1
  min_self_delegation: 1000000
  max_commission_rate: 0.2
  max_commission_change_rate: 0.01
EOF

    # Create validator key generation script
    cat > "$validator_dir/generate_keys.sh" << 'EOF'
#!/bin/bash

# Generate validator keys
echo "🔑 Generating validator keys..."

# Create wallet for validator
go run ../../cmd/gochain/main.go wallet \
    --wallet-file "keys/validator_wallet.dat" \
    --passphrase "validator_password"

echo "✅ Validator keys generated!"
echo "📁 Keys saved to: keys/validator_wallet.dat"
EOF

    chmod +x "$validator_dir/generate_keys.sh"

    # Create validator startup script
    cat > "$validator_dir/start_validator.sh" << 'EOF'
#!/bin/bash

# Start validator
echo "🚀 Starting validator..."

# Load configuration
CONFIG_FILE="config/validator_config.yaml"

# Start the validator node
go run ../../cmd/gochain/main.go \
    --config "$CONFIG_FILE" \
    --mining \
    --wallet-file "keys/validator_wallet.dat" \
    --passphrase "validator_password" &

VALIDATOR_PID=$!
echo $VALIDATOR_PID > validator.pid

echo "✅ Validator started!"
echo "📊 PID: $VALIDATOR_PID"
echo "🌐 API: http://localhost:8080"
echo "❤️  Health: http://localhost:8081/health"
echo "📈 Metrics: http://localhost:9090/metrics"

# Create stop script
cat > stop_validator.sh << 'EOL'
#!/bin/bash
echo "🛑 Stopping validator..."
if [ -f validator.pid ]; then
    kill $(cat validator.pid) 2>/dev/null || true
    rm validator.pid
fi
echo "✅ Validator stopped!"
EOL
chmod +x stop_validator.sh
EOF

    chmod +x "$validator_dir/start_validator.sh"

    # Create validator monitoring script
    cat > "$validator_dir/monitor_validator.sh" << 'EOF'
#!/bin/bash

# Monitor validator
echo "📊 Validator Monitor"
echo "==================="

while true; do
    clear
    echo "📊 Validator Status"
    echo "=================="
    echo "Timestamp: $(date)"
    echo
    
    # Check if validator is running
    if [ -f validator.pid ]; then
        PID=$(cat validator.pid)
        if ps -p $PID > /dev/null; then
            echo "✅ Validator: Running (PID: $PID)"
        else
            echo "❌ Validator: Not running"
        fi
    else
        echo "❌ Validator: Not started"
    fi
    
    # Check health endpoint
    if curl -s http://localhost:8081/health > /dev/null; then
        echo "✅ Health Check: OK"
    else
        echo "❌ Health Check: Failed"
    fi
    
    # Check metrics endpoint
    if curl -s http://localhost:9090/metrics > /dev/null; then
        echo "✅ Metrics: Available"
    else
        echo "❌ Metrics: Unavailable"
    fi
    
    echo
    echo "Press Ctrl+C to exit"
    sleep 5
done
EOF

    chmod +x "$validator_dir/monitor_validator.sh"

    # Create validator README
    cat > "$validator_dir/README.md" << EOF
# Validator $validator_id

This directory contains the configuration and scripts for validator $validator_id.

## Quick Start

1. **Generate Keys**
   \`\`\`bash
   ./generate_keys.sh
   \`\`\`

2. **Start Validator**
   \`\`\`bash
   ./start_validator.sh
   \`\`\`

3. **Monitor Validator**
   \`\`\`bash
   ./monitor_validator.sh
   \`\`\`

4. **Stop Validator**
   \`\`\`bash
   ./stop_validator.sh
   \`\`\`

## Configuration

- **Config File**: \`config/validator_config.yaml\`
- **Data Directory**: \`data/\`
- **Log Directory**: \`logs/\`
- **Keys Directory**: \`keys/\`

## Endpoints

- **API**: http://localhost:$((8080 + validator_id))
- **Health**: http://localhost:$((8081 + validator_id))/health
- **Metrics**: http://localhost:$((9090 + validator_id))/metrics
- **P2P**: Port $((30303 + validator_id))

## Files

- \`config/validator_config.yaml\` - Validator configuration
- \`keys/validator_wallet.dat\` - Validator wallet (generated)
- \`logs/validator_$validator_id.log\` - Validator logs
- \`validator.pid\` - Validator process ID (created when started)
EOF

    echo -e "${GREEN}✅ Validator $validator_id setup completed!${NC}"
}

# Setup all validators
for i in $(seq 1 $VALIDATOR_COUNT); do
    setup_validator $i
done

# Create master validator management script
echo -e "${BLUE}🎛️  Creating master validator management script...${NC}"
cat > "$VALIDATORS_DIR/manage_validators.sh" << 'EOF'
#!/bin/bash

# Master validator management script
echo "🎛️  Adrenochain Validator Manager"
echo "================================="

case "$1" in
    start)
        echo "🚀 Starting all validators..."
        for dir in validator_*; do
            if [ -d "$dir" ]; then
                echo "Starting $dir..."
                cd "$dir"
                ./start_validator.sh
                cd ..
            fi
        done
        echo "✅ All validators started!"
        ;;
    stop)
        echo "🛑 Stopping all validators..."
        for dir in validator_*; do
            if [ -d "$dir" ]; then
                echo "Stopping $dir..."
                cd "$dir"
                ./stop_validator.sh
                cd ..
            fi
        done
        echo "✅ All validators stopped!"
        ;;
    status)
        echo "📊 Validator Status"
        echo "=================="
        for dir in validator_*; do
            if [ -d "$dir" ]; then
                echo "📋 $dir:"
                cd "$dir"
                if [ -f validator.pid ]; then
                    PID=$(cat validator.pid)
                    if ps -p $PID > /dev/null; then
                        echo "  ✅ Running (PID: $PID)"
                    else
                        echo "  ❌ Not running"
                    fi
                else
                    echo "  ❌ Not started"
                fi
                cd ..
            fi
        done
        ;;
    generate-keys)
        echo "🔑 Generating keys for all validators..."
        for dir in validator_*; do
            if [ -d "$dir" ]; then
                echo "Generating keys for $dir..."
                cd "$dir"
                ./generate_keys.sh
                cd ..
            fi
        done
        echo "✅ All validator keys generated!"
        ;;
    *)
        echo "Usage: $0 {start|stop|status|generate-keys}"
        echo
        echo "Commands:"
        echo "  start         - Start all validators"
        echo "  stop          - Stop all validators"
        echo "  status        - Show status of all validators"
        echo "  generate-keys - Generate keys for all validators"
        exit 1
        ;;
esac
EOF

chmod +x "$VALIDATORS_DIR/manage_validators.sh"

# Create validator network monitoring script
echo -e "${BLUE}📊 Creating validator network monitoring script...${NC}"
cat > "$VALIDATORS_DIR/monitor_network.sh" << 'EOF'
#!/bin/bash

# Monitor entire validator network
echo "📊 Adrenochain Validator Network Monitor"
echo "========================================"

while true; do
    clear
    echo "📊 Adrenochain Validator Network Status"
    echo "======================================"
    echo "Timestamp: $(date)"
    echo
    
    # Check each validator
    for dir in validator_*; do
        if [ -d "$dir" ]; then
            validator_id=$(echo "$dir" | sed 's/validator_//')
            api_port=$((8080 + validator_id))
            health_port=$((8081 + validator_id))
            metrics_port=$((9090 + validator_id))
            
            echo "📋 $dir:"
            
            # Check if validator is running
            if [ -f "$dir/validator.pid" ]; then
                PID=$(cat "$dir/validator.pid")
                if ps -p $PID > /dev/null; then
                    echo "  ✅ Process: Running (PID: $PID)"
                else
                    echo "  ❌ Process: Not running"
                fi
            else
                echo "  ❌ Process: Not started"
            fi
            
            # Check health endpoint
            if curl -s http://localhost:$health_port/health > /dev/null; then
                echo "  ✅ Health: OK"
            else
                echo "  ❌ Health: Failed"
            fi
            
            # Check metrics endpoint
            if curl -s http://localhost:$metrics_port/metrics > /dev/null; then
                echo "  ✅ Metrics: Available"
            else
                echo "  ❌ Metrics: Unavailable"
            fi
            
            echo
        fi
    done
    
    echo "Press Ctrl+C to exit"
    sleep 5
done
EOF

chmod +x "$VALIDATORS_DIR/monitor_network.sh"

# Create validator README
echo -e "${BLUE}📝 Creating validator documentation...${NC}"
cat > "$VALIDATORS_DIR/README.md" << EOF
# Adrenochain Validators

This directory contains the validator setup for the adrenochain network.

## Quick Start

1. **Generate Keys for All Validators**
   \`\`\`bash
   ./manage_validators.sh generate-keys
   \`\`\`

2. **Start All Validators**
   \`\`\`bash
   ./manage_validators.sh start
   \`\`\`

3. **Monitor Network**
   \`\`\`bash
   ./monitor_network.sh
   \`\`\`

4. **Stop All Validators**
   \`\`\`bash
   ./manage_validators.sh stop
   \`\`\`

## Individual Validator Management

Each validator has its own directory with the following structure:

\`\`\`
validator_X/
├── config/
│   └── validator_config.yaml
├── data/                    # Blockchain data
├── logs/                    # Validator logs
├── keys/                    # Validator keys
├── generate_keys.sh         # Generate validator keys
├── start_validator.sh       # Start validator
├── stop_validator.sh        # Stop validator
├── monitor_validator.sh     # Monitor validator
└── README.md               # Validator documentation
\`\`\`

## Network Architecture

- **Validator 1**: API: 8081, Health: 8082, Metrics: 9091, P2P: 30304
- **Validator 2**: API: 8082, Health: 8083, Metrics: 9092, P2P: 30305
- **Validator 3**: API: 8083, Health: 8084, Metrics: 9093, P2P: 30306
- **...and so on**

## Management Scripts

- \`manage_validators.sh\` - Master script to manage all validators
- \`monitor_network.sh\` - Monitor the entire validator network
- \`validator_X/start_validator.sh\` - Start individual validator
- \`validator_X/stop_validator.sh\` - Stop individual validator
- \`validator_X/monitor_validator.sh\` - Monitor individual validator

## Configuration

Each validator has its own configuration file with:
- Network settings (ports, peers)
- Mining configuration
- Storage settings
- Logging configuration
- API settings
- Monitoring settings
- Validator-specific settings

## Security

- Each validator has its own wallet and keys
- Keys are stored in encrypted format
- Validators use different ports to avoid conflicts
- Each validator has its own data directory
EOF

echo -e "${GREEN}✅ Validator setup completed!${NC}"
echo -e "${BLUE}📁 Validators directory: $VALIDATORS_DIR${NC}"
echo -e "${BLUE}👥 Created $VALIDATOR_COUNT validators${NC}"
echo -e "${BLUE}📝 Next steps:${NC}"
echo -e "   1. cd $VALIDATORS_DIR"
echo -e "   2. ./manage_validators.sh generate-keys"
echo -e "   3. ./manage_validators.sh start"
echo -e "   4. ./monitor_network.sh"

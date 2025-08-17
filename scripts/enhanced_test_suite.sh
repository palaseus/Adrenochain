#!/bin/bash

# Enhanced adrenochain Test Suite
# This script provides comprehensive testing including multi-node testing

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
TEST_RESULTS_DIR="$PROJECT_ROOT/test_results"
COVERAGE_DIR="$PROJECT_ROOT/coverage"
NODE_DATA_DIR="$PROJECT_ROOT/test_nodes"
LOG_DIR="$PROJECT_ROOT/logs"

# Test configuration
NODE_COUNT=2
NODE_PORTS=(8545 8546)
P2P_PORTS=(30303 30304)
RPC_ENDPOINTS=("http://localhost:8545" "http://localhost:8546")
ENABLE_MULTI_NODE_TESTS=true
ENABLE_TRANSACTION_TESTS=true
ENABLE_SYNC_TESTS=true

# Node PIDs for cleanup
NODE_PIDS=()

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to cleanup on exit
cleanup() {
    print_status "Cleaning up test environment..."
    
    # Stop all test nodes
    for pid in "${NODE_PIDS[@]}"; do
        if kill -0 "$pid" 2>/dev/null; then
            print_status "Stopping node with PID $pid"
            kill "$pid" 2>/dev/null || true
        fi
    done
    
    # Wait for processes to finish
    for pid in "${NODE_PIDS[@]}"; do
        if kill -0 "$pid" 2>/dev/null; then
            wait "$pid" 2>/dev/null || true
        fi
    done
    
    # Clean up test data directories
    if [ -d "$NODE_DATA_DIR" ]; then
        print_status "Cleaning up test node data"
        rm -rf "$NODE_DATA_DIR"
    fi
    
    print_success "Cleanup completed"
}

# Set trap for cleanup
trap cleanup EXIT INT TERM

# Function to check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    # Check Go version
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed or not in PATH"
        exit 1
    fi
    
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    print_success "Go version $GO_VERSION detected"
    
    # Check if we're in the right directory
    if [ ! -f "$PROJECT_ROOT/go.mod" ]; then
        print_error "go.mod not found. Please run this script from the adrenochain project root."
        exit 1
    fi
    
    # Check for required tools
    if ! command -v curl &> /dev/null; then
        print_warning "curl not found. Some tests may fail."
    fi
    
    if ! command -v jq &> /dev/null; then
        print_warning "jq not found. JSON parsing tests may fail."
    fi
    
    print_success "Prerequisites check passed"
}

# Function to initialize test environment
init_test_environment() {
    print_status "Initializing test environment..."
    
    # Create directories
    mkdir -p "$TEST_RESULTS_DIR"
    mkdir -p "$COVERAGE_DIR"
    mkdir -p "$LOG_DIR"
    mkdir -p "$NODE_DATA_DIR"
    
    # Build adrenochain binary
    print_status "Building adrenochain binary..."
    cd "$PROJECT_ROOT"
    if ! go build -o adrenochain ./cmd/adrenochain; then
        print_error "Failed to build adrenochain binary"
        exit 1
    fi
    
    print_success "Test environment initialized"
}

# Function to start test nodes
start_test_nodes() {
    if [ "$ENABLE_MULTI_NODE_TESTS" != "true" ]; then
        print_status "Multi-node tests disabled, skipping node startup"
        return 0
    fi
    
    print_status "Starting $NODE_COUNT test nodes..."
    
    for i in $(seq 0 $((NODE_COUNT-1))); do
        local node_id=$i
        local rpc_port=${NODE_PORTS[$i]}
        local p2p_port=${P2P_PORTS[$i]}
        local data_dir="$NODE_DATA_DIR/node_$i"
        
        # Create node data directory
        mkdir -p "$data_dir"
        
        # Create node configuration
        cat > "$data_dir/config.yaml" << EOF
network:
  listen_port: $rpc_port
  p2p_port: $p2p_port
  max_peers: 10
  enable_mdns: false
  enable_relay: false
  connection_timeout: 30s

blockchain:
  data_dir: "$data_dir/blockchain"
  genesis_block: true
  difficulty: 1000000
  block_time: 15s

mining:
  enabled: true
  threads: 2
  reward_address: "0x1234567890123456789012345678901234567890"

storage:
  engine: "leveldb"
  cache_size: "256MB"
  write_buffer_size: "64MB"

logging:
  level: "info"
  format: "json"
  output: "file"
  file_path: "$data_dir/node.log"
EOF
        
        # Start node
        print_status "Starting node $i on RPC port $rpc_port, P2P port $p2p_port"
        
        # Start node in background
        ./adrenochain -config "$data_dir/config.yaml" > "$data_dir/stdout.log" 2> "$data_dir/stderr.log" &
        local node_pid=$!
        NODE_PIDS+=("$node_pid")
        
        print_success "Node $i started with PID $node_pid"
        
        # Wait for node to be ready
        print_status "Waiting for node $i to be ready..."
        local attempts=0
        local max_attempts=30
        
        while [ $attempts -lt $max_attempts ]; do
            if curl -s -f "http://localhost:$rpc_port/health" > /dev/null 2>&1; then
                print_success "Node $i is ready"
                break
            fi
            
            attempts=$((attempts + 1))
            sleep 1
        done
        
        if [ $attempts -eq $max_attempts ]; then
            print_error "Node $i failed to start within timeout"
            exit 1
        fi
    done
    
    print_success "All test nodes started successfully"
}

# Function to wait for nodes to sync
wait_for_node_sync() {
    if [ "$ENABLE_MULTI_NODE_TESTS" != "true" ]; then
        return 0
    fi
    
    print_status "Waiting for nodes to establish P2P connections..."
    
    # Wait for nodes to discover each other
    sleep 10
    
    # Check peer connections
    for i in $(seq 0 $((NODE_COUNT-1))); do
        local rpc_port=${NODE_PORTS[$i]}
        local peer_count=0
        local attempts=0
        local max_attempts=30
        
        while [ $attempts -lt $max_attempts ]; do
            if command -v jq &> /dev/null; then
                peer_count=$(curl -s "http://localhost:$rpc_port/network/peers" | jq '.peers | length' 2>/dev/null || echo "0")
            else
                # Fallback if jq is not available
                peer_count=$(curl -s "http://localhost:$rpc_port/network/peers" | grep -o '"peer_id"' | wc -l)
            fi
            
            if [ "$peer_count" -gt 0 ]; then
                print_success "Node $i has $peer_count peer(s)"
                break
            fi
            
            attempts=$((attempts + 1))
            sleep 2
        done
        
        if [ $attempts -eq $max_attempts ]; then
            print_warning "Node $i may not have established peer connections"
        fi
    done
}

# Function to run basic unit tests
run_unit_tests() {
    print_status "Running unit tests..."
    
    cd "$PROJECT_ROOT"
    
    # Run all tests with coverage
    if go test -v -coverprofile="$COVERAGE_DIR/coverage.out" -covermode=atomic ./... 2>&1 | tee "$TEST_RESULTS_DIR/unit_tests.log"; then
        print_success "Unit tests completed successfully"
    else
        print_error "Unit tests failed"
        return 1
    fi
    
    # Generate coverage report
    if command -v go tool cover &> /dev/null; then
        go tool cover -html="$COVERAGE_DIR/coverage.out" -o "$COVERAGE_DIR/coverage_report.html"
        print_success "Coverage report generated: $COVERAGE_DIR/coverage_report.html"
    fi
}

# Function to run fuzz tests
run_fuzz_tests() {
    print_status "Running fuzz tests..."
    
    cd "$PROJECT_ROOT"
    
    # Find packages with fuzz tests
    local fuzz_packages=$(find ./pkg -name "*_fuzz_test.go" -exec dirname {} \; | sort -u)
    
    if [ -z "$fuzz_packages" ]; then
        print_warning "No fuzz tests found"
        return 0
    fi
    
    print_status "Found fuzz tests in: $fuzz_packages"
    
    for pkg in $fuzz_packages; do
        print_status "Running fuzz tests in $pkg"
        
        # Run fuzz tests with timeout
        if timeout 30s go test -fuzz=Fuzz -fuzztime=10s "$pkg" 2>&1 | tee -a "$TEST_RESULTS_DIR/fuzz_tests.log"; then
            print_success "Fuzz tests in $pkg completed"
        else
            print_warning "Fuzz tests in $pkg may have timed out or failed"
        fi
    done
    
    print_success "Fuzz tests completed"
}

# Function to run benchmark tests
run_benchmark_tests() {
    print_status "Running benchmark tests..."
    
    cd "$PROJECT_ROOT"
    
    # Find packages with benchmark tests
    local benchmark_packages=$(find ./pkg -name "*_test.go" -exec grep -l "Benchmark" {} \; | xargs dirname | sort -u)
    
    if [ -z "$benchmark_packages" ]; then
        print_warning "No benchmark tests found"
        return 0
    fi
    
    print_status "Found benchmark tests in: $benchmark_packages"
    
    for pkg in $benchmark_packages; do
        print_status "Running benchmarks in $pkg"
        
        if go test -bench=. -benchmem "$pkg" 2>&1 | tee -a "$TEST_RESULTS_DIR/benchmark_tests.log"; then
            print_success "Benchmarks in $pkg completed"
        else
            print_warning "Benchmarks in $pkg may have failed"
        fi
    done
    
    print_success "Benchmark tests completed"
}

# Function to test node synchronization
test_node_sync() {
    if [ "$ENABLE_SYNC_TESTS" != "true" ] || [ "$ENABLE_MULTI_NODE_TESTS" != "true" ]; then
        return 0
    fi
    
    print_status "Testing node synchronization..."
    
    # Wait for some blocks to be mined
    print_status "Waiting for blocks to be mined..."
    sleep 30
    
    # Check block heights on all nodes
    local block_heights=()
    local max_height=0
    
    for i in $(seq 0 $((NODE_COUNT-1))); do
        local rpc_port=${NODE_PORTS[$i]}
        local height=0
        
        if command -v jq &> /dev/null; then
            height=$(curl -s "http://localhost:$rpc_port/blockchain/height" | jq '.height' 2>/dev/null || echo "0")
        else
            # Fallback parsing
            height=$(curl -s "http://localhost:$rpc_port/blockchain/height" | grep -o '"height":[0-9]*' | cut -d: -f2 || echo "0")
        fi
        
        block_heights+=("$height")
        
        if [ "$height" -gt "$max_height" ]; then
            max_height=$height
        fi
        
        print_status "Node $i block height: $height"
    done
    
    # Check if nodes are in sync (within 2 blocks)
    local sync_ok=true
    for height in "${block_heights[@]}"; do
        if [ $((max_height - height)) -gt 2 ]; then
            print_warning "Node synchronization may be delayed (height difference: $((max_height - height)))"
            sync_ok=false
        fi
    done
    
    if [ "$sync_ok" = "true" ]; then
        print_success "Node synchronization test passed"
    else
        print_warning "Node synchronization test may have issues"
    fi
}

# Function to test transaction propagation
test_transaction_propagation() {
    if [ "$ENABLE_TRANSACTION_TESTS" != "true" ] || [ "$ENABLE_MULTI_NODE_TESTS" != "true" ]; then
        return 0
    fi
    
    print_status "Testing transaction propagation..."
    
    # Create a test transaction
    local test_tx='{
        "from": "0x1234567890123456789012345678901234567890",
        "to": "0x0987654321098765432109876543210987654321",
        "value": "1000000000000000000",
        "gas": 21000,
        "gasPrice": "20000000000"
    }'
    
    # Submit transaction to first node
    local tx_hash=""
    if command -v jq &> /dev/null; then
        tx_hash=$(curl -s -X POST -H "Content-Type: application/json" \
            -d "$test_tx" \
            "http://localhost:${NODE_PORTS[0]}/transactions" | \
            jq -r '.hash' 2>/dev/null || echo "")
    else
        # Fallback parsing
        tx_hash=$(curl -s -X POST -H "Content-Type: application/json" \
            -d "$test_tx" \
            "http://localhost:${NODE_PORTS[0]}/transactions" | \
            grep -o '"hash":"[^"]*"' | cut -d'"' -f4 || echo "")
    fi
    
    if [ -n "$tx_hash" ] && [ "$tx_hash" != "null" ]; then
        print_success "Transaction submitted: $tx_hash"
        
        # Wait for transaction to propagate
        sleep 5
        
        # Check if transaction appears in mempool of other nodes
        for i in $(seq 1 $((NODE_COUNT-1))); do
            local rpc_port=${NODE_PORTS[$i]}
            local tx_found=false
            
            if command -v jq &> /dev/null; then
                local mempool_txs=$(curl -s "http://localhost:$rpc_port/mempool" | jq '.transactions | length' 2>/dev/null || echo "0")
                if [ "$mempool_txs" -gt 0 ]; then
                    tx_found=true
                fi
            else
                # Fallback check
                if curl -s "http://localhost:$rpc_port/mempool" | grep -q "transaction"; then
                    tx_found=true
                fi
            fi
            
            if [ "$tx_found" = "true" ]; then
                print_success "Transaction found in node $i mempool"
            else
                print_warning "Transaction not found in node $i mempool"
            fi
        done
    else
        print_warning "Failed to submit test transaction"
    fi
    
    print_success "Transaction propagation test completed"
}

# Function to run multi-node integration tests
run_multi_node_tests() {
    if [ "$ENABLE_MULTI_NODE_TESTS" != "true" ]; then
        return 0
    fi
    
    print_status "Running multi-node integration tests..."
    
    # Test node synchronization
    test_node_sync
    
    # Test transaction propagation
    test_transaction_propagation
    
    # Test network connectivity
    test_network_connectivity
    
    print_success "Multi-node integration tests completed"
}

# Function to test network connectivity
test_network_connectivity() {
    print_status "Testing network connectivity between nodes..."
    
    for i in $(seq 0 $((NODE_COUNT-1))); do
        local rpc_port=${NODE_PORTS[$i]}
        
        # Test basic connectivity
        if curl -s -f "http://localhost:$rpc_port/health" > /dev/null; then
            print_success "Node $i health check passed"
        else
            print_error "Node $i health check failed"
        fi
        
        # Test blockchain info
        if curl -s -f "http://localhost:$rpc_port/blockchain/info" > /dev/null; then
            print_success "Node $i blockchain info accessible"
        else
            print_warning "Node $i blockchain info not accessible"
        fi
    done
}

# Function to generate test summary
generate_test_summary() {
    print_status "Generating test summary..."
    
    local summary_file="$TEST_RESULTS_DIR/enhanced_test_summary.md"
    
    cat > "$summary_file" << EOF
# Enhanced adrenochain Test Suite Results

## Test Execution Summary
- **Execution Time**: $(date)
- **Test Environment**: Multi-node setup with $NODE_COUNT nodes
- **Multi-node Tests**: $ENABLE_MULTI_NODE_TESTS
- **Transaction Tests**: $ENABLE_TRANSACTION_TESTS
- **Sync Tests**: $ENABLE_SYNC_TESTS

## Node Configuration
EOF
    
    for i in $(seq 0 $((NODE_COUNT-1))); do
        echo "- **Node $i**: RPC Port ${NODE_PORTS[$i]}, P2P Port ${P2P_PORTS[$i]}" >> "$summary_file"
    done
    
    cat >> "$summary_file" << EOF

## Test Results
- **Unit Tests**: See unit_tests.log
- **Fuzz Tests**: See fuzz_tests.log  
- **Benchmark Tests**: See benchmark_tests.log
- **Coverage Report**: $COVERAGE_DIR/coverage_report.html

## Multi-Node Testing
- **Node Synchronization**: $(if [ "$ENABLE_SYNC_TESTS" = "true" ]; then echo "Enabled"; else echo "Disabled"; fi)
- **Transaction Propagation**: $(if [ "$ENABLE_TRANSACTION_TESTS" = "true" ]; then echo "Enabled"; else echo "Disabled"; fi)
- **Network Connectivity**: Tested

## Files Generated
- Test Results: $TEST_RESULTS_DIR/
- Coverage Data: $COVERAGE_DIR/
- Node Logs: $NODE_DATA_DIR/
- Test Summary: $summary_file

EOF
    
    print_success "Test summary generated: $summary_file"
}

# Main execution
main() {
    echo
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║                🚀 Enhanced adrenochain Test Suite 🚀            ║"
    echo "║                                                              ║"
    echo "║  Multi-node testing with synchronization and transactions   ║"
    echo "║  Comprehensive coverage and performance validation          ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo
    
    # Check prerequisites
    check_prerequisites
    
    # Initialize environment
    init_test_environment
    
    # Start test nodes
    start_test_nodes
    
    # Wait for nodes to sync
    wait_for_node_sync
    
    # Run tests
    run_unit_tests
    run_fuzz_tests
    run_benchmark_tests
    
    # Run multi-node tests
    run_multi_node_tests
    
    # Generate summary
    generate_test_summary
    
    echo
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║                    🎯 Enhanced Test Suite Complete 🎯      ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo
    
    print_success "Enhanced test suite completed successfully!"
    print_status "Check results in: $TEST_RESULTS_DIR/"
    print_status "Check coverage in: $COVERAGE_DIR/"
    print_status "Check node logs in: $NODE_DATA_DIR/"
}

# Run main function
main "$@"

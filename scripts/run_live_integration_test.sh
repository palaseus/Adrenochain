#!/bin/bash

# Live Node Integration Test Runner
# This script runs comprehensive live blockchain node integration tests

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Default configuration
VERBOSE=false
COVERAGE=false
TIMEOUT="30m"
TEST_PACKAGE="./pkg/testing"
TEST_PATTERN="TestLiveNodeIntegration"
PARALLEL_JOBS=1

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -c|--coverage)
            COVERAGE=true
            shift
            ;;
        -t|--timeout)
            TIMEOUT="$2"
            shift 2
            ;;
        -j|--jobs)
            PARALLEL_JOBS="$2"
            shift 2
            ;;
        -h|--help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "OPTIONS:"
            echo "  -v, --verbose        Enable verbose output"
            echo "  -c, --coverage       Generate coverage report"
            echo "  -t, --timeout DURATION    Set test timeout (default: 30m)"
            echo "  -j, --jobs N         Number of parallel test jobs (default: 1)"
            echo "  -h, --help           Show this help message"
            echo ""
            echo "EXAMPLES:"
            echo "  $0                   Run basic live node integration test"
            echo "  $0 -v -c             Run with verbose output and coverage"
            echo "  $0 -t 45m            Run with 45 minute timeout"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

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

print_header() {
    echo -e "${PURPLE}=== $1 ===${NC}"
}

# Function to check system resources
check_system_resources() {
    print_header "SYSTEM RESOURCE CHECK"
    
    # Check available memory
    local available_memory=$(free -m | awk 'NR==2{printf "%.1f", $7/1024}')
    print_status "Available memory: ${available_memory}GB"
    
    if (( $(echo "$available_memory < 2.0" | bc -l) )); then
        print_warning "Low available memory detected. Live node tests may fail."
    fi
    
    # Check available disk space
    local available_disk=$(df -h . | awk 'NR==2{print $4}')
    print_status "Available disk space: $available_disk"
    
    # Check if ports are available
    local test_ports=(8000 8001 8002 9000 9001 9002)
    for port in "${test_ports[@]}"; do
        if lsof -i :$port >/dev/null 2>&1; then
            print_warning "Port $port is in use. This may interfere with node testing."
        fi
    done
    
    print_success "System resource check completed"
}

# Function to check Go environment and dependencies
check_go_environment() {
    print_header "GO ENVIRONMENT CHECK"
    
    # Check Go version
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed or not in PATH"
        exit 1
    fi
    
    local go_version=$(go version | awk '{print $3}')
    print_status "Go version: $go_version"
    
    # Check if we're in a Go module
    if [[ ! -f "go.mod" ]]; then
        print_error "go.mod not found. Run this script from the project root."
        exit 1
    fi
    
    # Verify dependencies
    print_status "Checking Go dependencies..."
    if ! go mod verify; then
        print_warning "Go module verification failed, attempting to download dependencies..."
        go mod download
        go mod tidy
    fi
    
    print_success "Go environment check completed"
}

# Function to run pre-test checks
run_pre_test_checks() {
    print_header "PRE-TEST CHECKS"
    
    # Check if test file exists
    local test_file="pkg/testing/live_node_integration_test.go"
    if [[ ! -f "$test_file" ]]; then
        print_error "Live node integration test file not found: $test_file"
        exit 1
    fi
    print_status "Test file found: $test_file"
    
    # Compile test to check for syntax errors
    print_status "Compiling test package..."
    if ! go test -c "$TEST_PACKAGE" -o /tmp/live_integration_test >/dev/null 2>&1; then
        print_error "Test compilation failed. Check for syntax errors."
        go test -c "$TEST_PACKAGE" -o /tmp/live_integration_test
        exit 1
    fi
    rm -f /tmp/live_integration_test
    
    print_success "Pre-test checks completed"
}

# Function to build the project
build_project() {
    print_header "PROJECT BUILD"
    
    print_status "Building adrenochain project..."
    if ! go build ./...; then
        print_error "Project build failed"
        exit 1
    fi
    
    # Build the main binary if it exists
    if [[ -f "cmd/gochain/main.go" ]]; then
        print_status "Building gochain binary..."
        if ! go build -o adrenochain cmd/gochain/main.go; then
            print_warning "Failed to build gochain binary, but continuing with tests..."
        else
            print_success "Built adrenochain binary"
        fi
    fi
    
    print_success "Project build completed"
}

# Function to run the live node integration test
run_live_integration_test() {
    print_header "LIVE NODE INTEGRATION TEST"
    
    local test_args=()
    
    # Add verbose flag if requested
    if [[ "$VERBOSE" == "true" ]]; then
        test_args+=("-v")
    fi
    
    # Add timeout
    test_args+=("-timeout" "$TIMEOUT")
    
    # Add parallel jobs
    if [[ "$PARALLEL_JOBS" -gt 1 ]]; then
        test_args+=("-parallel" "$PARALLEL_JOBS")
    fi
    
    # Add coverage if requested
    if [[ "$COVERAGE" == "true" ]]; then
        test_args+=("-coverprofile=live_integration_coverage.out")
        test_args+=("-covermode=atomic")
    fi
    
    # Add test pattern
    test_args+=("-run" "$TEST_PATTERN")
    
    print_status "Running live node integration test with timeout: $TIMEOUT"
    print_status "Test command: go test ${test_args[*]} $TEST_PACKAGE"
    
    local start_time=$(date +%s)
    
    # Run the test
    if go test "${test_args[@]}" "$TEST_PACKAGE"; then
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        print_success "Live node integration test PASSED (${duration}s)"
        
        # Generate coverage report if requested
        if [[ "$COVERAGE" == "true" && -f "live_integration_coverage.out" ]]; then
            print_status "Generating coverage report..."
            go tool cover -html=live_integration_coverage.out -o live_integration_coverage.html
            local coverage_percentage=$(go tool cover -func=live_integration_coverage.out | tail -1 | awk '{print $3}')
            print_status "Coverage report generated: live_integration_coverage.html"
            print_status "Total coverage: $coverage_percentage"
        fi
        
        return 0
    else
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        print_error "Live node integration test FAILED (${duration}s)"
        return 1
    fi
}

# Function to clean up any lingering processes
cleanup_processes() {
    print_status "Cleaning up any lingering test processes..."
    
    # Kill any adrenochain processes
    pkill -f "adrenochain" 2>/dev/null || true
    pkill -f "live_integration_test" 2>/dev/null || true
    
    # Clean up temporary files
    rm -f live_integration_coverage.out
    rm -f /tmp/live_integration_test
    
    # Clean up any test data directories
    find /tmp -name "adrenochain_live_test_*" -type d -exec rm -rf {} + 2>/dev/null || true
    
    print_status "Cleanup completed"
}

# Function to generate test report
generate_test_report() {
    local test_result="$1"
    local report_file="live_integration_test_report.md"
    
    print_header "GENERATING TEST REPORT"
    
    cat > "$report_file" << EOF
# Live Node Integration Test Report

**Date:** $(date)
**Test Result:** $test_result
**Timeout:** $TIMEOUT
**Verbose Mode:** $VERBOSE
**Coverage:** $COVERAGE

## Test Configuration

- Test Package: $TEST_PACKAGE
- Test Pattern: $TEST_PATTERN
- Parallel Jobs: $PARALLEL_JOBS

## System Information

- OS: $(uname -a)
- Go Version: $(go version)
- Available Memory: $(free -h | awk 'NR==2{print $7}')
- Available Disk: $(df -h . | awk 'NR==2{print $4}')

## Test Details

The live node integration test performs the following operations:

1. **Network Setup**: Creates multiple live blockchain nodes with real P2P networking
2. **Node Communication**: Tests P2P discovery and connection establishment
3. **Mining and Consensus**: Validates real block production and consensus mechanisms
4. **Transaction Processing**: Creates and processes real transactions through the network
5. **Network Synchronization**: Tests blockchain synchronization across multiple nodes
6. **Stress Testing**: Evaluates network performance under load

## Coverage Report

EOF

    if [[ "$COVERAGE" == "true" && -f "live_integration_coverage.html" ]]; then
        echo "Coverage report available: live_integration_coverage.html" >> "$report_file"
        if command -v go &> /dev/null; then
            local coverage_percentage=$(go tool cover -func=live_integration_coverage.out 2>/dev/null | tail -1 | awk '{print $3}' || echo "N/A")
            echo "Total coverage: $coverage_percentage" >> "$report_file"
        fi
    else
        echo "No coverage report generated" >> "$report_file"
    fi
    
    cat >> "$report_file" << EOF

## Recommendations

- Run this test regularly to ensure live node functionality
- Monitor system resources during test execution
- Use verbose mode (-v) for detailed debugging information
- Generate coverage reports (-c) to track test completeness

EOF
    
    print_success "Test report generated: $report_file"
}

# Main execution function
main() {
    local start_time=$(date)
    local test_result="UNKNOWN"
    
    print_header "LIVE NODE INTEGRATION TEST RUNNER"
    print_status "Starting live node integration test at: $start_time"
    
    # Set up signal handlers for cleanup
    trap cleanup_processes EXIT
    trap 'echo -e "\n${YELLOW}Interrupted by user${NC}"; cleanup_processes; exit 1' INT TERM
    
    # Run all checks and tests
    check_system_resources
    check_go_environment
    run_pre_test_checks
    build_project
    
    # Run the main test
    if run_live_integration_test; then
        test_result="PASSED"
    else
        test_result="FAILED"
    fi
    
    # Generate report
    generate_test_report "$test_result"
    
    # Print final summary
    local end_time=$(date)
    print_header "TEST EXECUTION SUMMARY"
    echo "Start Time: $start_time"
    echo "End Time: $end_time"
    echo "Test Result: $test_result"
    echo "Timeout: $TIMEOUT"
    echo "Coverage: $COVERAGE"
    echo "Verbose: $VERBOSE"
    
    if [[ "$test_result" == "PASSED" ]]; then
        print_success "✓ Live node integration test completed successfully!"
        exit 0
    else
        print_error "✗ Live node integration test failed"
        exit 1
    fi
}

# Run main function
main "$@"

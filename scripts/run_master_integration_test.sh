#!/bin/bash

# Master Integration Test Runner
# This script runs the comprehensive end-to-end blockchain test suite

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
TEST_TIMEOUT=1800  # 30 minutes
PARALLEL_JOBS=4
VERBOSE=false
COVERAGE=false
STRESS_TEST=false
PERFORMANCE_TEST=false

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
        -s|--stress)
            STRESS_TEST=true
            shift
            ;;
        -p|--performance)
            PERFORMANCE_TEST=true
            shift
            ;;
        -t|--timeout)
            TEST_TIMEOUT="$2"
            shift 2
            ;;
        -j|--jobs)
            PARALLEL_JOBS="$2"
            shift 2
            ;;
        -h|--help)
            echo "Usage: $0 [OPTIONS]"
            echo "Options:"
            echo "  -v, --verbose      Enable verbose output"
            echo "  -c, --coverage     Enable coverage reporting"
            echo "  -s, --stress       Enable stress testing"
            echo "  -p, --performance  Enable performance benchmarking"
            echo "  -t, --timeout      Set test timeout in seconds (default: 1800)"
            echo "  -j, --jobs         Set parallel jobs (default: 4)"
            echo "  -h, --help         Show this help message"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

echo -e "${BLUE}=== MASTER INTEGRATION TEST RUNNER ===${NC}"
echo "Starting comprehensive blockchain integration test suite..."
echo ""

# Check if we're in the right directory
if [[ ! -f "go.mod" ]]; then
    echo -e "${RED}Error: Must run from project root directory${NC}"
    exit 1
fi

# Create test results directory
TEST_RESULTS_DIR="test_results/master_integration_$(date +%Y%m%d_%H%M%S)"
mkdir -p "$TEST_RESULTS_DIR"
echo "Test results will be saved to: $TEST_RESULTS_DIR"

# Set environment variables for testing
export TEST_ENV=integration
export TEST_TIMEOUT=$TEST_TIMEOUT
export GO_TEST_TIMEOUT=$TEST_TIMEOUT

# Function to log with timestamp
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

# Function to run tests with timeout
run_test_with_timeout() {
    local test_name="$1"
    local test_cmd="$2"
    local timeout="$3"
    
    log "Starting $test_name..."
    log "Command: $test_cmd"
    log "Timeout: ${timeout}s"
    
    # Run test with timeout
    if timeout "$timeout" bash -c "$test_cmd"; then
        log "✓ $test_name completed successfully"
        return 0
    else
        log "✗ $test_name failed or timed out after ${timeout}s"
        return 1
    fi
}

# Function to check system resources
check_system_resources() {
    log "Checking system resources..."
    
    # Check available memory
    local mem_available=$(free -m | awk 'NR==2{printf "%.0f", $7}')
    if [[ $mem_available -lt 2048 ]]; then
        echo -e "${YELLOW}Warning: Low memory available (${mem_available}MB). Recommended: 2GB+${NC}"
    else
        log "Memory: ${mem_available}MB available ✓"
    fi
    
    # Check available disk space
    local disk_available=$(df . | awk 'NR==2{printf "%.0f", $4}')
    if [[ $disk_available -lt 1024 ]]; then
        echo -e "${YELLOW}Warning: Low disk space available (${disk_available}MB). Recommended: 1GB+${NC}"
    else
        log "Disk space: ${disk_available}MB available ✓"
    fi
    
    # Check CPU cores
    local cpu_cores=$(nproc)
    log "CPU cores: $cpu_cores ✓"
    
    # Check if timeout command is available
    if ! command -v timeout &> /dev/null; then
        echo -e "${RED}Error: 'timeout' command not found. Please install coreutils.${NC}"
        exit 1
    fi
}

# Function to run pre-test checks
run_pre_test_checks() {
    log "Running pre-test checks..."
    
    # Check Go installation
    if ! command -v go &> /dev/null; then
        echo -e "${RED}Error: Go is not installed${NC}"
        exit 1
    fi
    
    # Check Go version
    local go_version=$(go version | awk '{print $3}')
    log "Go version: $go_version"
    
    # Check if dependencies are available
    log "Checking dependencies..."
    if ! go mod download; then
        echo -e "${RED}Error: Failed to download Go dependencies${NC}"
        exit 1
    fi
    
    # Check if test files exist
    if [[ ! -f "pkg/testing/master_integration_test.go" ]]; then
        echo -e "${RED}Error: Master integration test file not found${NC}"
        exit 1
    fi
    
    log "Pre-test checks completed ✓"
}

# Function to run the master integration test
run_master_integration_test() {
    log "Running master integration test..."
    
    local test_cmd="go test -v -timeout ${TEST_TIMEOUT}s ./pkg/testing -run TestMasterIntegration"
    
    if [[ "$COVERAGE" == "true" ]]; then
        test_cmd="$test_cmd -coverprofile=$TEST_RESULTS_DIR/master_integration_coverage.out"
    fi
    
    if [[ "$VERBOSE" == "true" ]]; then
        test_cmd="$test_cmd -v"
    fi
    
    # Run the test
    if run_test_with_timeout "Master Integration Test" "$test_cmd" "$TEST_TIMEOUT"; then
        echo -e "${GREEN}✓ Master integration test completed successfully${NC}"
        return 0
    else
        echo -e "${RED}✗ Master integration test failed${NC}"
        return 1
    fi
}

# Function to run additional test suites
run_additional_tests() {
    log "Running additional test suites..."
    
    local additional_tests=(
        "pkg/consensus"
        "pkg/ai"
        "pkg/governance"
        "pkg/exchange"
        "pkg/defi"
        "pkg/sync"
    )
    
    local failed_tests=()
    
    for test_dir in "${additional_tests[@]}"; do
        if [[ -d "$test_dir" ]]; then
            log "Testing $test_dir..."
            local test_cmd="go test -v -timeout 300s ./$test_dir"
            
            if [[ "$COVERAGE" == "true" ]]; then
                test_cmd="$test_cmd -coverprofile=$TEST_RESULTS_DIR/${test_dir//\//_}_coverage.out"
            fi
            
            if timeout 300 bash -c "$test_cmd"; then
                log "✓ $test_dir tests passed"
            else
                log "✗ $test_dir tests failed"
                failed_tests+=("$test_dir")
            fi
        fi
    done
    
    if [[ ${#failed_tests[@]} -gt 0 ]]; then
        echo -e "${YELLOW}Warning: Some additional tests failed:${NC}"
        printf '%s\n' "${failed_tests[@]}"
        return 1
    fi
    
    return 0
}

# Function to run stress tests
run_stress_tests() {
    if [[ "$STRESS_TEST" != "true" ]]; then
        return 0
    fi
    
    log "Running stress tests..."
    
    # Run tests with high concurrency
    local stress_cmd="go test -v -timeout 600s -parallel 8 ./pkg/testing -run TestMasterIntegration"
    
    if run_test_with_timeout "Stress Test" "$stress_cmd" 600; then
        log "✓ Stress tests completed successfully"
        return 0
    else
        log "✗ Stress tests failed"
        return 1
    fi
}

# Function to run performance tests
run_performance_tests() {
    if [[ "$PERFORMANCE_TEST" != "true" ]]; then
        return 0
    fi
    
    log "Running performance tests..."
    
    # Run tests multiple times to get performance metrics
    local perf_cmd="for i in {1..5}; do echo '=== Run $i ==='; go test -v -timeout 300s ./pkg/testing -run TestMasterIntegration; done"
    
    if run_test_with_timeout "Performance Test" "$perf_cmd" 1800; then
        log "✓ Performance tests completed successfully"
        return 0
    else
        log "✗ Performance tests failed"
        return 1
    fi
}

# Function to generate test report
generate_test_report() {
    log "Generating test report..."
    
    local report_file="$TEST_RESULTS_DIR/test_report.md"
    
    cat > "$report_file" << EOF
# Master Integration Test Report

**Generated:** $(date)
**Test Duration:** $(($(date +%s) - $(date -d "$START_TIME" +%s))) seconds

## Test Configuration
- **Timeout:** ${TEST_TIMEOUT}s
- **Parallel Jobs:** ${PARALLEL_JOBS}
- **Coverage:** ${COVERAGE}
- **Stress Test:** ${STRESS_TEST}
- **Performance Test:** ${PERFORMANCE_TEST}

## Test Results
- **Master Integration Test:** ${MASTER_TEST_RESULT}
- **Additional Tests:** ${ADDITIONAL_TESTS_RESULT}
- **Stress Tests:** ${STRESS_TEST_RESULT}
- **Performance Tests:** ${PERFORMANCE_TEST_RESULT}

## System Information
- **Go Version:** $(go version)
- **OS:** $(uname -a)
- **CPU Cores:** $(nproc)
- **Memory:** $(free -h | awk 'NR==2{print $2}')
- **Disk Space:** $(df -h . | awk 'NR==2{print $4}')

## Coverage Reports
EOF

    if [[ "$COVERAGE" == "true" ]]; then
        for coverage_file in "$TEST_RESULTS_DIR"/*_coverage.out; do
            if [[ -f "$coverage_file" ]]; then
                local package_name=$(basename "$coverage_file" _coverage.out)
                echo "- [$package_name]($coverage_file)" >> "$report_file"
            fi
        done
    fi

    echo "" >> "$report_file"
    echo "## Test Logs" >> "$report_file"
    echo "Check the test results directory for detailed logs and coverage reports." >> "$report_file"
    
    log "Test report generated: $report_file"
}

# Function to cleanup
cleanup() {
    log "Cleaning up..."
    
    # Kill any remaining test processes
    pkill -f "go test" 2>/dev/null || true
    
    # Remove temporary files
    rm -f /tmp/go-test-*
    
    log "Cleanup completed"
}

# Main execution
main() {
    local START_TIME=$(date)
    local MASTER_TEST_RESULT="UNKNOWN"
    local ADDITIONAL_TESTS_RESULT="UNKNOWN"
    local STRESS_TEST_RESULT="UNKNOWN"
    local PERFORMANCE_TEST_RESULT="UNKNOWN"
    
    # Set up signal handlers
    trap cleanup EXIT
    trap 'echo -e "\n${YELLOW}Interrupted by user${NC}"; cleanup; exit 1' INT TERM
    
    # Check system resources
    check_system_resources
    
    # Run pre-test checks
    run_pre_test_checks
    
    # Run master integration test
    if run_master_integration_test; then
        MASTER_TEST_RESULT="PASSED"
    else
        MASTER_TEST_RESULT="FAILED"
    fi
    
    # Run additional tests
    if run_additional_tests; then
        ADDITIONAL_TESTS_RESULT="PASSED"
    else
        ADDITIONAL_TESTS_RESULT="FAILED"
    fi
    
    # Run stress tests if enabled
    if [[ "$STRESS_TEST" == "true" ]]; then
        if run_stress_tests; then
            STRESS_TEST_RESULT="PASSED"
        else
            STRESS_TEST_RESULT="FAILED"
        fi
    else
        STRESS_TEST_RESULT="SKIPPED"
    fi
    
    # Run performance tests if enabled
    if [[ "$PERFORMANCE_TEST" == "true" ]]; then
        if run_performance_tests; then
            PERFORMANCE_TEST_RESULT="PASSED"
        else
            PERFORMANCE_TEST_RESULT="FAILED"
        fi
    else
        PERFORMANCE_TEST_RESULT="SKIPPED"
    fi
    
    # Generate test report
    generate_test_report
    
    # Print final summary
    echo ""
    echo -e "${BLUE}=== TEST EXECUTION SUMMARY ===${NC}"
    echo "Master Integration Test: $MASTER_TEST_RESULT"
    echo "Additional Tests: $ADDITIONAL_TESTS_RESULT"
    echo "Stress Tests: $STRESS_TEST_RESULT"
    echo "Performance Tests: $PERFORMANCE_TEST_RESULT"
    echo ""
    echo "Test results saved to: $TEST_RESULTS_DIR"
    
    # Determine overall success
    if [[ "$MASTER_TEST_RESULT" == "PASSED" ]]; then
        echo -e "${GREEN}✓ All critical tests passed!${NC}"
        exit 0
    else
        echo -e "${RED}✗ Some critical tests failed${NC}"
        exit 1
    fi
}

# Run main function
main "$@"

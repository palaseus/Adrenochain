#!/bin/bash

# 🚀 GoChain Comprehensive Test Suite
# This script provides a unified testing experience for the entire GoChain project

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TEST_RESULTS_DIR="$PROJECT_ROOT/test_results"
COVERAGE_DIR="$PROJECT_ROOT/coverage"
LOG_FILE="$TEST_RESULTS_DIR/test_suite.log"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
GO_VERSION=$(go version | awk '{print $3}')

# Test configuration
TIMEOUT=300s  # 5 minutes per test package
RACE_DETECTION=true
COVERAGE_ENABLED=true
VERBOSE_TESTS=true
PARALLEL_TESTS=true
FUZZ_TESTS=true
BENCHMARK_TESTS=true

# Statistics
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
SKIPPED_TESTS=0
TOTAL_PACKAGES=0
PASSED_PACKAGES=0
FAILED_PACKAGES=0

# Initialize test environment
init_test_environment() {
    echo -e "${BLUE}🔧 Initializing GoChain Test Suite...${NC}"
    
    # Create directories
    mkdir -p "$TEST_RESULTS_DIR"
    mkdir -p "$COVERAGE_DIR"
    
    # Clean previous results
    rm -rf "$TEST_RESULTS_DIR"/*
    rm -rf "$COVERAGE_DIR"/*
    
    # Start log file
    {
        echo "=== GoChain Test Suite Execution ==="
        echo "Timestamp: $(date)"
        echo "Go Version: $GO_VERSION"
        echo "Project Root: $PROJECT_ROOT"
        echo "Test Results Dir: $TEST_RESULTS_DIR"
        echo "Coverage Dir: $COVERAGE_DIR"
        echo "====================================="
        echo
    } > "$LOG_FILE"
    
    echo -e "${GREEN}✅ Test environment initialized${NC}"
}

# Print banner
print_banner() {
    echo -e "${CYAN}"
    cat << "EOF"
    ╔══════════════════════════════════════════════════════════════╗
    ║                    🚀 GoChain Test Suite 🚀                ║
    ║                                                              ║
    ║  Comprehensive testing for the GoChain blockchain project   ║
    ║                                                              ║
    ║  Features:                                                   ║
    ║  • Unit tests with coverage                                 ║
    ║  • Fuzz testing                                             ║
    ║  • Benchmark testing                                        ║
    ║  • Race condition detection                                 ║
    ║  • Detailed reporting                                       ║
    ║  • Performance metrics                                      ║
    ╚══════════════════════════════════════════════════════════════╝
EOF
    echo -e "${NC}"
}

# Check prerequisites
check_prerequisites() {
    echo -e "${BLUE}🔍 Checking prerequisites...${NC}"
    
    # Check Go installation
    if ! command -v go &> /dev/null; then
        echo -e "${RED}❌ Go is not installed or not in PATH${NC}"
        exit 1
    fi
    
    # Check Go version
    if [[ ! "$GO_VERSION" =~ ^go1\.(1[8-9]|2[0-9]) ]]; then
        echo -e "${YELLOW}⚠️  Warning: Go version $GO_VERSION detected. Go 1.18+ recommended.${NC}"
    else
        echo -e "${GREEN}✅ Go version $GO_VERSION detected${NC}"
    fi
    
    # Check if we're in the right directory
    if [[ ! -f "$PROJECT_ROOT/go.mod" ]]; then
        echo -e "${RED}❌ Not in GoChain project root (go.mod not found)${NC}"
        exit 1
    fi
    
    # Check for test dependencies
    if ! go list -f '{{.TestGoFiles}}' ./pkg/... | grep -q .; then
        echo -e "${YELLOW}⚠️  Warning: No test files found in packages${NC}"
    fi
    
    echo -e "${GREEN}✅ Prerequisites check passed${NC}"
}

# Get all test packages
get_test_packages() {
    echo -e "${BLUE}🔍 Discovering test packages...${NC}"
    
    # Get all packages that have tests
    local packages=($(go list -f '{{if .TestGoFiles}}{{.ImportPath}}{{end}}' ./...))
    
    if [[ ${#packages[@]} -eq 0 ]]; then
        echo -e "${YELLOW}⚠️  No test packages found${NC}"
        return
    fi
    
    echo -e "${GREEN}✅ Found ${#packages[@]} test packages:${NC}"
    for pkg in "${packages[@]}"; do
        echo -e "   📦 $pkg"
    done
    
    TOTAL_PACKAGES=${#packages[@]}
    echo
}

# Run tests for a specific package
run_package_tests() {
    local package_path="$1"
    local package_name=$(basename "$package_path")
    local test_file="$TEST_RESULTS_DIR/${package_name}_tests.log"
    local coverage_file="$COVERAGE_DIR/${package_name}_coverage.out"
    
    echo -e "${BLUE}🧪 Testing package: $package_path${NC}"
    
    # Build test binary first
    if ! go build -o /dev/null "$package_path" 2>/dev/null; then
        echo -e "${YELLOW}⚠️  Package $package_path has build issues, skipping${NC}"
        SKIPPED_TESTS=$((SKIPPED_TESTS + 1))
        return 1
    fi
    
    # Run tests with various options
    local test_cmd="go test"
    local test_args=()
    
    if [[ "$VERBOSE_TESTS" == true ]]; then
        test_args+=("-v")
    fi
    
    if [[ "$RACE_DETECTION" == true ]]; then
        test_args+=("-race")
    fi
    
    if [[ "$COVERAGE_ENABLED" == true ]]; then
        test_args+=("-coverprofile=$coverage_file" "-covermode=atomic")
    fi
    
    test_args+=("-timeout=$TIMEOUT" "$package_path")
    
    # Run the tests
    local start_time=$(date +%s)
    local exit_code=0
    
    echo "Running: $test_cmd ${test_args[*]}" | tee -a "$LOG_FILE"
    
    if $test_cmd "${test_args[@]}" 2>&1 | tee "$test_file"; then
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        
        echo -e "${GREEN}✅ Package $package_name tests passed (${duration}s)${NC}"
        PASSED_PACKAGES=$((PASSED_PACKAGES + 1))
        
        # Count tests from output (only actual test results, not test function names)
        local test_count=$(grep -c "^--- PASS\|^--- FAIL\|^--- SKIP" "$test_file" 2>/dev/null | tr -d '\n' || echo "0")
        test_count=${test_count:-0}
        TOTAL_TESTS=$((TOTAL_TESTS + test_count))
        
        # Count passed tests
        local passed_count=$(grep -c "^--- PASS" "$test_file" 2>/dev/null | tr -d '\n' || echo "0")
        passed_count=${passed_count:-0}
        PASSED_TESTS=$((PASSED_TESTS + passed_count))
        
        # Count failed tests
        local failed_count=$(grep -c "^--- FAIL" "$test_file" 2>/dev/null | tr -d '\n' || echo "0")
        failed_count=${failed_count:-0}
        FAILED_TESTS=$((FAILED_TESTS + failed_count))
        
        # Count skipped tests
        local skipped_count=$(grep -c "^--- SKIP" "$test_file" 2>/dev/null | tr -d '\n' || echo "0")
        skipped_count=${skipped_count:-0}
        SKIPPED_TESTS=$((SKIPPED_TESTS + skipped_count))
        
    else
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        
        echo -e "${RED}❌ Package $package_name tests failed (${duration}s)${NC}"
        FAILED_PACKAGES=$((FAILED_PACKAGES + 1))
        exit_code=1
        
        # Count tests from output (only actual test results, not test function names)
        local test_count=$(grep -c "^--- PASS\|^--- FAIL\|^--- SKIP" "$test_file" 2>/dev/null | tr -d '\n' || echo "0")
        test_count=${test_count:-0}
        TOTAL_TESTS=$((TOTAL_TESTS + test_count))
        
        # Count passed tests
        local passed_count=$(grep -c "^--- PASS" "$test_file" 2>/dev/null | tr -d '\n' || echo "0")
        passed_count=${passed_count:-0}
        PASSED_TESTS=$((PASSED_TESTS + passed_count))
        
        # Count failed tests
        local failed_count=$(grep -c "^--- FAIL" "$test_file" 2>/dev/null | tr -d '\n' || echo "0")
        failed_count=${failed_count:-0}
        FAILED_TESTS=$((FAILED_TESTS + failed_count))
        
        # Count skipped tests
        local skipped_count=$(grep -c "^--- SKIP" "$test_file" 2>/dev/null | tr -d '\n' || echo "0")
        skipped_count=${skipped_count:-0}
        SKIPPED_TESTS=$((SKIPPED_TESTS + skipped_count))
    fi
    
    echo "Test results saved to: $test_file"
    if [[ "$COVERAGE_ENABLED" == true ]] && [[ -f "$coverage_file" ]]; then
        echo "Coverage data saved to: $coverage_file"
    fi
    
    echo
    return $exit_code
}

# Run fuzz tests
run_fuzz_tests() {
    if [[ "$FUZZ_TESTS" != true ]]; then
        return 0
    fi
    
    echo -e "${BLUE}🧪 Running fuzz tests...${NC}"
    
    # Find packages with fuzz tests
    local fuzz_packages=($(find ./pkg -name "*_fuzz_test.go" -exec dirname {} \; | sort -u))
    
    if [[ ${#fuzz_packages[@]} -eq 0 ]]; then
        echo -e "${YELLOW}⚠️  No fuzz tests found${NC}"
        return 0
    fi
    
    echo -e "${GREEN}✅ Found ${#fuzz_packages[@]} packages with fuzz tests:${NC}"
    
    for pkg in "${fuzz_packages[@]}"; do
        local package_name=$(basename "$pkg")
        echo -e "   🧪 $pkg"
        
        # Find all fuzz test functions in the package
        local fuzz_functions=($(grep -h "func Fuzz[A-Za-z]*(" "$pkg"/*_fuzz_test.go | sed 's/func //' | sed 's/(.*//'))
        
        if [[ ${#fuzz_functions[@]} -eq 0 ]]; then
            echo -e "   ⚠️  No fuzz functions found in $pkg"
            continue
        fi
        
        # Run each fuzz test function separately
        for fuzz_func in "${fuzz_functions[@]}"; do
            echo -e "      🧪 Running $fuzz_func..."
            if go test -fuzz="$fuzz_func" -fuzztime=5s "$pkg" 2>&1 | tee -a "$TEST_RESULTS_DIR/${package_name}_fuzz.log"; then
                echo -e "      ✅ $fuzz_func completed successfully"
            else
                echo -e "      ⚠️  $fuzz_func had issues"
            fi
        done
        
        echo -e "   ✅ Fuzz tests for $package_name completed"
    done
    
    echo
}

# Run benchmark tests
run_benchmark_tests() {
    if [[ "$BENCHMARK_TESTS" != true ]]; then
        return 0
    fi
    
    echo -e "${BLUE}📊 Running benchmark tests...${NC}"
    
    # Find packages with benchmark tests
    local bench_packages=()
    while IFS= read -r -d '' file; do
        if grep -q "Benchmark" "$file"; then
            local dir=$(dirname "$file")
            if [[ ! " ${bench_packages[@]} " =~ " ${dir} " ]]; then
                bench_packages+=("$dir")
            fi
        fi
    done < <(find ./pkg -name "*_test.go" -print0)
    
    if [[ ${#bench_packages[@]} -eq 0 ]]; then
        echo -e "${YELLOW}⚠️  No benchmark tests found${NC}"
        return 0
    fi
    
    echo -e "${GREEN}✅ Found ${#bench_packages[@]} packages with benchmark tests:${NC}"
    
    for pkg in "${bench_packages[@]}"; do
        local package_name=$(basename "$pkg")
        echo -e "   📊 $pkg"
        
        # Run benchmarks
        if go test -bench=. -benchmem "$pkg" 2>&1 | tee "$TEST_RESULTS_DIR/${package_name}_bench.log"; then
            echo -e "${GREEN}✅ Benchmarks for $package_name completed${NC}"
        else
            echo -e "${YELLOW}⚠️  Benchmarks for $package_name had issues${NC}"
        fi
    done
    
    echo
}

# Generate coverage report
generate_coverage_report() {
    if [[ "$COVERAGE_ENABLED" != true ]]; then
        return 0
    fi
    
    echo -e "${BLUE}📊 Generating coverage report...${NC}"
    
    # Combine all coverage files
    local combined_coverage="$COVERAGE_DIR/combined_coverage.out"
    local coverage_report="$COVERAGE_DIR/coverage_report.html"
    
    # Find all coverage files
    local coverage_files=($(find "$COVERAGE_DIR" -name "*_coverage.out"))
    
    if [[ ${#coverage_files[@]} -eq 0 ]]; then
        echo -e "${YELLOW}⚠️  No coverage files found${NC}"
        return 0
    fi
    
    echo -e "${GREEN}✅ Found ${#coverage_files[@]} coverage files${NC}"
    
    # Combine coverage files
    if [[ ${#coverage_files[@]} -gt 1 ]]; then
        echo "mode: atomic" > "$combined_coverage"
        for file in "${coverage_files[@]}"; do
            tail -n +2 "$file" >> "$combined_coverage" 2>/dev/null || true
        done
    else
        cp "${coverage_files[0]}" "$combined_coverage"
    fi
    
    # Generate HTML report
    if go tool cover -html="$combined_coverage" -o "$coverage_report" 2>/dev/null; then
        echo -e "${GREEN}✅ Coverage report generated: $coverage_report${NC}"
    else
        echo -e "${YELLOW}⚠️  Failed to generate HTML coverage report${NC}"
    fi
    
    # Show coverage summary
    if command -v go tool cover &> /dev/null; then
        echo -e "${BLUE}📊 Coverage Summary:${NC}"
        go tool cover -func="$combined_coverage" | tail -1
    fi
    
    echo
}

# Generate test summary report
generate_test_summary() {
    echo -e "${BLUE}📋 Generating test summary report...${NC}"
    
    local summary_file="$TEST_RESULTS_DIR/test_summary.md"
    
    {
        echo "# GoChain Test Suite Results"
        echo
        echo "**Execution Date:** $(date)"
        echo "**Go Version:** $GO_VERSION"
        echo "**Total Duration:** $(($(date +%s) - $(date -d @$(cat /proc/uptime | awk '{print $1}' | cut -d. -f1) +%s)))s"
        echo
        echo "## 📊 Test Statistics"
        echo
        echo "| Metric | Count |"
        echo "|--------|-------|"
        echo "| **Total Packages** | $TOTAL_PACKAGES |"
        echo "| **Passed Packages** | $PASSED_PACKAGES |"
        echo "| **Failed Packages** | $FAILED_PACKAGES |"
        echo "| **Total Tests** | $TOTAL_TESTS |"
        echo "| **Passed Tests** | $PASSED_TESTS |"
        echo "| **Failed Tests** | $FAILED_TESTS |"
        echo "| **Skipped Tests** | $SKIPPED_TESTS |"
        echo
        echo "## 🎯 Success Rate"
        echo
        if [[ $TOTAL_PACKAGES -gt 0 ]]; then
            local package_success_rate=$((PASSED_PACKAGES * 100 / TOTAL_PACKAGES))
            echo "- **Package Success Rate:** ${package_success_rate}%"
        fi
        
        if [[ $TOTAL_TESTS -gt 0 ]]; then
            local test_success_rate=$((PASSED_TESTS * 100 / TOTAL_TESTS))
            echo "- **Test Success Rate:** ${test_success_rate}%"
        fi
        echo
        echo "## 📁 Test Results"
        echo
        echo "Detailed test results are available in the following files:"
        echo
        for file in "$TEST_RESULTS_DIR"/*.log; do
            if [[ -f "$file" ]]; then
                local filename=$(basename "$file")
                echo "- [$filename]($file)"
            fi
        done
        echo
        echo "## 📊 Coverage Report"
        echo
        if [[ -f "$COVERAGE_DIR/coverage_report.html" ]]; then
            echo "- [Coverage Report]($COVERAGE_DIR/coverage_report.html)"
        fi
        if [[ -f "$COVERAGE_DIR/combined_coverage.out" ]]; then
            echo "- [Combined Coverage Data]($COVERAGE_DIR/combined_coverage.out)"
        fi
        echo
        echo "## 🚀 Next Steps"
        echo
        if [[ $FAILED_PACKAGES -gt 0 ]]; then
            echo "❌ **Action Required:** $FAILED_PACKAGES package(s) failed tests"
            echo "   - Review failed test logs in $TEST_RESULTS_DIR"
            echo "   - Fix failing tests before proceeding"
        else
            echo "✅ **All tests passed successfully!**"
            echo "   - Ready for deployment or further development"
        fi
        echo
        echo "---"
        echo "*Generated by GoChain Test Suite*"
    } > "$summary_file"
    
    echo -e "${GREEN}✅ Test summary generated: $summary_file${NC}"
}

# Print final results
print_final_results() {
    echo -e "${CYAN}"
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║                    🎯 Test Suite Complete 🎯                ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
    
    echo -e "${BLUE}📊 Final Results Summary:${NC}"
    echo -e "   📦 Packages: ${GREEN}$PASSED_PACKAGES passed${NC}, ${RED}$FAILED_PACKAGES failed${NC} (Total: $TOTAL_PACKAGES)"
    echo -e "   🧪 Tests: ${GREEN}$PASSED_TESTS passed${NC}, ${RED}$FAILED_TESTS failed${NC}, ${YELLOW}$SKIPPED_TESTS skipped${NC} (Total: $TOTAL_TESTS)"
    
    if [[ $TOTAL_PACKAGES -gt 0 ]]; then
        local package_success_rate=$((PASSED_PACKAGES * 100 / TOTAL_PACKAGES))
        echo -e "   📈 Package Success Rate: ${GREEN}${package_success_rate}%${NC}"
    fi
    
    if [[ $TOTAL_TESTS -gt 0 ]]; then
        local test_success_rate=$((PASSED_TESTS * 100 / TOTAL_TESTS))
        echo -e "   📈 Test Success Rate: ${GREEN}${test_success_rate}%${NC}"
    fi
    
    echo
    echo -e "${BLUE}📁 Results Location:${NC}"
    echo -e "   📋 Test Results: ${CYAN}$TEST_RESULTS_DIR${NC}"
    echo -e "   📊 Coverage: ${CYAN}$COVERAGE_DIR${NC}"
    echo -e "   📝 Summary: ${CYAN}$TEST_RESULTS_DIR/test_summary.md${NC}"
    echo -e "   📄 Log: ${CYAN}$LOG_FILE${NC}"
    
    echo
    if [[ $FAILED_PACKAGES -gt 0 ]]; then
        echo -e "${RED}❌ Some tests failed. Please review the logs and fix the issues.${NC}"
        exit 1
    else
        echo -e "${GREEN}🎉 All tests passed successfully! GoChain is ready for action! 🚀${NC}"
    fi
}

# Main execution function
main() {
    print_banner
    check_prerequisites
    init_test_environment
    get_test_packages
    
    echo -e "${BLUE}🚀 Starting comprehensive test suite...${NC}"
    echo
    
    # Run tests for each package
    local packages=($(go list -f '{{if .TestGoFiles}}{{.ImportPath}}{{end}}' ./...))
    local overall_success=true
    
    for pkg in "${packages[@]}"; do
        if ! run_package_tests "$pkg"; then
            overall_success=false
        fi
    done
    
    # Run additional test types
    run_fuzz_tests
    run_benchmark_tests
    
    # Generate reports
    generate_coverage_report
    generate_test_summary
    
    # Print results
    print_final_results
    
    # Exit with appropriate code
    if [[ "$overall_success" == true ]]; then
        exit 0
    else
        exit 1
    fi
}

# Handle script arguments
case "${1:-}" in
    --help|-h)
        echo "Usage: $0 [OPTIONS]"
        echo
        echo "Options:"
        echo "  --help, -h     Show this help message"
        echo "  --no-race      Disable race detection"
        echo "  --no-coverage  Disable coverage reporting"
        echo "  --no-fuzz      Disable fuzz testing"
        echo "  --no-bench     Disable benchmark testing"
        echo "  --verbose      Enable verbose output"
        echo "  --timeout N    Set test timeout (default: 300s)"
        echo
        echo "Examples:"
        echo "  $0                    # Run all tests with default settings"
        echo "  $0 --no-race         # Run tests without race detection"
        echo "  $0 --timeout 600s    # Run tests with 10 minute timeout"
        echo "  $0 --no-coverage     # Run tests without coverage"
        exit 0
        ;;
    --no-race)
        RACE_DETECTION=false
        shift
        ;;
    --no-coverage)
        COVERAGE_ENABLED=false
        shift
        ;;
    --no-fuzz)
        FUZZ_TESTS=false
        shift
        ;;
    --no-bench)
        BENCHMARK_TESTS=false
        shift
        ;;
    --verbose)
        VERBOSE_TESTS=true
        shift
        ;;
    --timeout)
        if [[ -n "${2:-}" ]]; then
            TIMEOUT="$2"
            shift 2
        else
            echo -e "${RED}Error: --timeout requires a value${NC}"
            exit 1
        fi
        ;;
esac

# Run the main function
main "$@"

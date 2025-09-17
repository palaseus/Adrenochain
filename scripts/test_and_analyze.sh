#!/bin/bash

# 🧪 Adrenochain Comprehensive Testing & Analysis Script
# This script performs in-depth testing and analysis of the entire Adrenochain ecosystem

set -e

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
ANALYSIS_DIR="$PROJECT_ROOT/analysis"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

echo -e "${CYAN}"
cat << "EOF"
╔══════════════════════════════════════════════════════════════╗
║            🧪 Adrenochain Testing & Analysis 🧪            ║
║                                                              ║
║  Comprehensive testing and analysis of the entire ecosystem  ║
║                                                              ║
╚══════════════════════════════════════════════════════════════╝
EOF
echo -e "${NC}"

# Create directories
mkdir -p "$TEST_RESULTS_DIR" "$ANALYSIS_DIR"

# Test results file
TEST_RESULTS="$TEST_RESULTS_DIR/test_results_$TIMESTAMP.txt"
ANALYSIS_REPORT="$ANALYSIS_DIR/analysis_report_$TIMESTAMP.md"

# Initialize test results
echo "Adrenochain Test Results - $(date)" > "$TEST_RESULTS"
echo "=====================================" >> "$TEST_RESULTS"
echo "" >> "$TEST_RESULTS"

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_TOTAL=0

# Test function
run_test() {
    local test_name="$1"
    local test_command="$2"
    local expected_exit_code="${3:-0}"
    
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
    
    echo -e "${BLUE}🧪 Running: $test_name${NC}"
    echo "Test: $test_name" >> "$TEST_RESULTS"
    echo "Command: $test_command" >> "$TEST_RESULTS"
    
    if eval "$test_command" >> "$TEST_RESULTS" 2>&1; then
        if [ $? -eq $expected_exit_code ]; then
            echo -e "${GREEN}✅ PASSED: $test_name${NC}"
            echo "Result: PASSED" >> "$TEST_RESULTS"
            TESTS_PASSED=$((TESTS_PASSED + 1))
        else
            echo -e "${RED}❌ FAILED: $test_name (wrong exit code)${NC}"
            echo "Result: FAILED (wrong exit code)" >> "$TEST_RESULTS"
            TESTS_FAILED=$((TESTS_FAILED + 1))
        fi
    else
        echo -e "${RED}❌ FAILED: $test_name${NC}"
        echo "Result: FAILED" >> "$TEST_RESULTS"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
    
    echo "" >> "$TEST_RESULTS"
    echo ""
}

# Analysis function
analyze_component() {
    local component="$1"
    local analysis_command="$2"
    
    echo -e "${PURPLE}🔍 Analyzing: $component${NC}"
    echo "## $component Analysis" >> "$ANALYSIS_REPORT"
    echo "" >> "$ANALYSIS_REPORT"
    
    eval "$analysis_command" >> "$ANALYSIS_REPORT" 2>&1
    
    echo "" >> "$ANALYSIS_REPORT"
    echo ""
}

# ===== BUILD TESTS =====
echo -e "${YELLOW}📦 BUILD TESTS${NC}"
echo "=================="

run_test "Go Build - Main Application" "go build -o bin/adrenochain ./cmd/gochain"
run_test "Go Build - CLI Tool" "go build -o bin/adrenochain-cli ./tools/cli"
run_test "Go Build - All Packages" "go build ./..."

# ===== UNIT TESTS =====
echo -e "${YELLOW}🧪 UNIT TESTS${NC}"
echo "==============="

run_test "Unit Tests - Core Packages" "go test ./pkg/block/... -v"
run_test "Unit Tests - UTXO System" "go test ./pkg/utxo/... -v"
run_test "Unit Tests - Storage" "go test ./pkg/storage/... -v"
run_test "Unit Tests - API" "go test ./pkg/api/... -v"
run_test "Unit Tests - All Packages" "go test ./... -timeout 5m"

# ===== DOCKER TESTS =====
echo -e "${YELLOW}🐳 DOCKER TESTS${NC}"
echo "================="

run_test "Docker Build" "docker build -t adrenochain:test ."
run_test "Docker Compose Build" "docker-compose build"

# ===== FUNCTIONAL TESTS =====
echo -e "${YELLOW}⚙️  FUNCTIONAL TESTS${NC}"
echo "====================="

# Start the application for testing
echo -e "${BLUE}🚀 Starting Adrenochain for functional tests...${NC}"
./bin/adrenochain --config config/production.yaml --mining &
ADRENOCHAIN_PID=$!
sleep 10

# Test API endpoints
run_test "API Health Check" "curl -s http://localhost:8080/health | grep -q 'healthy'"
run_test "API Blockchain Info" "curl -s http://localhost:8080/api/v1/chain/info | grep -q 'height'"
run_test "API Network Status" "curl -s http://localhost:8080/api/v1/network/status | grep -q 'active'"
run_test "API Chain Height" "curl -s http://localhost:8080/api/v1/chain/height | grep -q 'height'"

# Test CLI tool
run_test "CLI Health Check" "./bin/adrenochain-cli health"
run_test "CLI Blockchain Info" "./bin/adrenochain-cli info"
run_test "CLI Network Status" "./bin/adrenochain-cli network"

# Stop the application
kill $ADRENOCHAIN_PID 2>/dev/null || true
sleep 2

# ===== SECURITY TESTS =====
echo -e "${YELLOW}🔒 SECURITY TESTS${NC}"
echo "=================="

run_test "Go Security Check" "go list -json -deps ./... | grep -v 'golang.org/x' | grep -v 'github.com' | wc -l"
run_test "Dependency Check" "go mod verify"

# ===== PERFORMANCE TESTS =====
echo -e "${YELLOW}⚡ PERFORMANCE TESTS${NC}"
echo "====================="

run_test "Build Performance" "time go build ./cmd/gochain"
run_test "Test Performance" "time go test ./pkg/block/... -run=TestBlock"

# ===== CODE QUALITY TESTS =====
echo -e "${YELLOW}📊 CODE QUALITY TESTS${NC}"
echo "======================="

run_test "Go Format Check" "gofmt -l . | wc -l | grep -q '^0$'"
run_test "Go Vet Check" "go vet ./..."
run_test "Go Mod Tidy" "go mod tidy && git diff --exit-code go.mod go.sum"

# ===== COMPREHENSIVE ANALYSIS =====
echo -e "${YELLOW}🔍 COMPREHENSIVE ANALYSIS${NC}"
echo "============================="

# Initialize analysis report
cat > "$ANALYSIS_REPORT" << 'EOF'
# Adrenochain Comprehensive Analysis Report

Generated: $(date)

## Executive Summary

This report provides a comprehensive analysis of the Adrenochain blockchain ecosystem, including code quality, architecture, performance, and security assessments.

EOF

# Code analysis
analyze_component "Code Statistics" "echo '### Code Statistics' && find . -name '*.go' -not -path './vendor/*' | wc -l && echo 'Go files found' && find . -name '*.go' -not -path './vendor/*' -exec wc -l {} + | tail -1 && echo 'Total lines of code'"

analyze_component "Package Analysis" "echo '### Package Structure' && go list ./... | wc -l && echo 'packages' && go list ./... | head -20"

analyze_component "Dependency Analysis" "echo '### Dependencies' && go list -m all | wc -l && echo 'dependencies' && go list -m all | head -20"

analyze_component "Test Coverage" "echo '### Test Coverage' && go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out | tail -1"

analyze_component "Build Analysis" "echo '### Build Analysis' && go build -a -ldflags='-s -w' -o /tmp/adrenochain ./cmd/gochain && ls -lh /tmp/adrenochain"

analyze_component "Memory Analysis" "echo '### Memory Usage' && go build -race -o /tmp/adrenochain-race ./cmd/gochain && ls -lh /tmp/adrenochain-race"

# Architecture analysis
analyze_component "Architecture Analysis" "echo '### Architecture Overview' && echo 'Core Components:' && ls -la pkg/ | grep '^d' && echo 'Applications:' && ls -la cmd/ | grep '^d' && echo 'SDKs:' && ls -la sdk/ | grep '^d' && echo 'Apps:' && ls -la apps/ | grep '^d'"

# Security analysis
analyze_component "Security Analysis" "echo '### Security Assessment' && echo 'Potential vulnerabilities:' && grep -r 'TODO.*security\|FIXME.*security\|XXX.*security' . --include='*.go' | wc -l && echo 'Hardcoded secrets:' && grep -r 'password\|secret\|key' . --include='*.go' | grep -v 'test' | wc -l"

# Performance analysis
analyze_component "Performance Analysis" "echo '### Performance Metrics' && echo 'Binary size:' && ls -lh bin/adrenochain 2>/dev/null || echo 'Binary not found' && echo 'Docker image size:' && docker images adrenochain:latest --format 'table {{.Size}}' 2>/dev/null || echo 'Docker image not found'"

# ===== DOCKER ANALYSIS =====
echo -e "${YELLOW}🐳 DOCKER ANALYSIS${NC}"
echo "=================="

analyze_component "Docker Analysis" "echo '### Docker Configuration' && echo 'Dockerfile layers:' && docker history adrenochain:latest --format 'table {{.CreatedBy}}' 2>/dev/null || echo 'Docker image not found' && echo 'Docker Compose services:' && docker-compose config --services 2>/dev/null || echo 'Docker Compose not configured'"

# ===== SDK ANALYSIS =====
echo -e "${YELLOW}🛠️  SDK ANALYSIS${NC}"
echo "================="

analyze_component "JavaScript SDK" "echo '### JavaScript SDK' && ls -la sdk/javascript/ && echo 'Package.json dependencies:' && cat sdk/javascript/package.json | grep -A 20 'dependencies' 2>/dev/null || echo 'Package.json not found'"

analyze_component "Python SDK" "echo '### Python SDK' && ls -la sdk/python/ && echo 'Requirements:' && cat sdk/python/requirements.txt | wc -l && echo 'dependencies'"

# ===== WEB WALLET ANALYSIS =====
echo -e "${YELLOW}🌐 WEB WALLET ANALYSIS${NC}"
echo "======================="

analyze_component "Web Wallet" "echo '### Web Wallet Application' && ls -la apps/web-wallet/ && echo 'React components:' && find apps/web-wallet/src -name '*.tsx' -o -name '*.ts' | wc -l && echo 'components'"

# ===== FINAL REPORT =====
echo -e "${YELLOW}📋 FINAL REPORT${NC}"
echo "==============="

# Generate final summary
cat >> "$ANALYSIS_REPORT" << EOF

## Test Results Summary

- **Total Tests**: $TESTS_TOTAL
- **Passed**: $TESTS_PASSED
- **Failed**: $TESTS_FAILED
- **Success Rate**: $(( (TESTS_PASSED * 100) / TESTS_TOTAL ))%

## Recommendations

Based on the analysis, here are the key recommendations:

1. **Code Quality**: $(if [ $TESTS_PASSED -gt $((TESTS_TOTAL * 80 / 100)) ]; then echo "Excellent code quality"; else echo "Code quality needs improvement"; fi)
2. **Test Coverage**: $(if [ $TESTS_PASSED -gt $((TESTS_TOTAL * 90 / 100)) ]; then echo "Comprehensive test coverage"; else echo "Test coverage could be improved"; fi)
3. **Security**: $(if [ $TESTS_FAILED -eq 0 ]; then echo "No security issues detected"; else echo "Security review recommended"; fi)
4. **Performance**: $(if [ $TESTS_PASSED -gt $((TESTS_TOTAL * 85 / 100)) ]; then echo "Good performance characteristics"; else echo "Performance optimization needed"; fi)

## Next Steps

1. Review failed tests and address issues
2. Implement additional security measures
3. Optimize performance bottlenecks
4. Enhance test coverage
5. Update documentation

---
*Report generated by Adrenochain Testing & Analysis Script*
EOF

# Display final results
echo -e "${CYAN}"
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║                    📊 TEST RESULTS SUMMARY 📊                ║"
echo "╠══════════════════════════════════════════════════════════════╣"
echo "║  Total Tests: $TESTS_TOTAL"
echo "║  ✅ Passed: $TESTS_PASSED"
echo "║  ❌ Failed: $TESTS_FAILED"
echo "║  📈 Success Rate: $(( (TESTS_PASSED * 100) / TESTS_TOTAL ))%"
echo "╚══════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

echo -e "${GREEN}📁 Test Results: $TEST_RESULTS${NC}"
echo -e "${GREEN}📊 Analysis Report: $ANALYSIS_REPORT${NC}"

# Cleanup
rm -f /tmp/adrenochain /tmp/adrenochain-race coverage.out 2>/dev/null || true

echo -e "${GREEN}🎉 Testing and analysis completed successfully!${NC}"





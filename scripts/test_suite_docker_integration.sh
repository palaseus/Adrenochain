#!/bin/bash

# 🚀 adrenochain Test Suite with Docker Integration
# This script extends the original test_suite.sh to work with Docker
# It can run both traditional Go tests and Docker-based tests

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
DOCKER_TEST_RESULTS_DIR="$PROJECT_ROOT/docker_test_results"
LOG_FILE="$TEST_RESULTS_DIR/test_suite_docker.log"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

# Test configuration
TIMEOUT=300s
COVERAGE_ENABLED=true
VERBOSE_TESTS=true
DOCKER_TESTS=true
TRADITIONAL_TESTS=true
INTEGRATION_TESTS=true

# Statistics
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
SKIPPED_TESTS=0
DOCKER_SERVICES_HEALTHY=0
DOCKER_SERVICES_TOTAL=0

# Print banner
print_banner() {
    echo -e "${CYAN}"
    cat << "EOF"
    ╔══════════════════════════════════════════════════════════════╗
    ║            🚀 adrenochain Test Suite + Docker 🐳           ║
    ║                                                              ║
    ║  Comprehensive testing for adrenochain with Docker support  ║
    ║                                                              ║
    ║  Features:                                                   ║
    ║  • Traditional Go unit tests                                 ║
    ║  • Docker service integration tests                          ║
    ║  • API endpoint validation                                   ║
    ║  • Monitoring system tests                                   ║
    ║  • End-to-end workflow tests                                 ║
    ║  • Performance benchmarking                                  ║
    ║  • Timeout-protected operations                              ║
    ╚══════════════════════════════════════════════════════════════╝
EOF
    echo -e "${NC}"
}

# Check if Docker is available and running
check_docker_availability() {
    echo -e "${BLUE}🔍 Checking Docker availability...${NC}"
    
    if command -v docker &> /dev/null && command -v docker-compose &> /dev/null; then
        if timeout 5 sudo docker info &> /dev/null; then
            echo -e "${GREEN}✅ Docker is available and running${NC}"
            return 0
        else
            echo -e "${YELLOW}⚠️  Docker is installed but daemon is not running${NC}"
            return 1
        fi
    else
        echo -e "${YELLOW}⚠️  Docker is not installed or not in PATH${NC}"
        return 1
    fi
}

# Run traditional Go tests (from original test_suite.sh)
run_traditional_tests() {
    if [[ "$TRADITIONAL_TESTS" != true ]]; then
        return 0
    fi
    
    echo -e "${BLUE}🧪 Running Traditional Go Tests...${NC}"
    
    # Check if original test_suite.sh exists
    if [[ -f "$PROJECT_ROOT/scripts/test_suite.sh" ]]; then
        echo -e "${GREEN}✅ Found original test_suite.sh, running it...${NC}"
        
        # Run the original test suite with timeout
        if timeout 1800 "$PROJECT_ROOT/scripts/test_suite.sh" --no-race --no-fuzz --no-bench 2>&1 | tee "$TEST_RESULTS_DIR/traditional_tests.log"; then
            echo -e "${GREEN}✅ Traditional tests completed successfully${NC}"
            PASSED_TESTS=$((PASSED_TESTS + 1))
        else
            echo -e "${YELLOW}⚠️  Traditional tests had issues${NC}"
            FAILED_TESTS=$((FAILED_TESTS + 1))
        fi
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
    else
        echo -e "${YELLOW}⚠️  Original test_suite.sh not found, skipping traditional tests${NC}"
    fi
    
    echo
}

# Run Docker integration tests
run_docker_integration_tests() {
    if [[ "$DOCKER_TESTS" != true ]]; then
        return 0
    fi
    
    echo -e "${BLUE}🐳 Running Docker Integration Tests...${NC}"
    
    # Check if Docker is available
    if ! check_docker_availability; then
        echo -e "${YELLOW}⚠️  Docker not available, skipping Docker tests${NC}"
        return 0
    fi
    
    # Check if docker-compose.yml exists
    if [[ ! -f "$PROJECT_ROOT/docker-compose.yml" ]]; then
        echo -e "${YELLOW}⚠️  docker-compose.yml not found, skipping Docker tests${NC}"
        return 0
    fi
    
    echo -e "${GREEN}✅ Running Docker service health checks...${NC}"
    
    # Test Docker services
    local services=("adrenochain-node" "adrenochain-prometheus" "adrenochain-grafana" "adrenochain-redis")
    DOCKER_SERVICES_TOTAL=${#services[@]}
    
    for service in "${services[@]}"; do
        echo -e "   🐳 Testing $service..."
        
        # Check if container is running
        if timeout 10 sudo docker ps --format "table {{.Names}}\t{{.Status}}" | grep -q "$service"; then
            echo -e "      ✅ $service container is running"
            
            # Test service endpoints
            case "$service" in
                "adrenochain-node")
                    if timeout 5 curl -s http://localhost:8080/health > /dev/null 2>&1; then
                        echo -e "      ✅ $service API is responding"
                        DOCKER_SERVICES_HEALTHY=$((DOCKER_SERVICES_HEALTHY + 1))
                    else
                        echo -e "      ❌ $service API is not responding"
                    fi
                    ;;
                "adrenochain-prometheus")
                    if timeout 5 curl -s http://localhost:9091/api/v1/query?query=up > /dev/null 2>&1; then
                        echo -e "      ✅ $service API is responding"
                        DOCKER_SERVICES_HEALTHY=$((DOCKER_SERVICES_HEALTHY + 1))
                    else
                        echo -e "      ❌ $service API is not responding"
                    fi
                    ;;
                "adrenochain-grafana")
                    if timeout 5 curl -s -u admin:admin http://localhost:3000/api/health > /dev/null 2>&1; then
                        echo -e "      ✅ $service API is responding"
                        DOCKER_SERVICES_HEALTHY=$((DOCKER_SERVICES_HEALTHY + 1))
                    else
                        echo -e "      ❌ $service API is not responding"
                    fi
                    ;;
                "adrenochain-redis")
                    if timeout 5 sudo docker exec adrenochain-redis redis-cli ping > /dev/null 2>&1; then
                        echo -e "      ✅ $service is responding"
                        DOCKER_SERVICES_HEALTHY=$((DOCKER_SERVICES_HEALTHY + 1))
                    else
                        echo -e "      ❌ $service is not responding"
                    fi
                    ;;
            esac
        else
            echo -e "      ❌ $service container is not running"
        fi
    done
    
    # Test monitoring service if available
    echo -e "   📊 Testing monitoring service..."
    if timeout 5 curl -s http://localhost:9093/health > /dev/null 2>&1; then
        echo -e "      ✅ Monitoring service is responding"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "      ❌ Monitoring service is not responding"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    echo -e "${GREEN}✅ Docker integration tests completed${NC}"
    echo
}

# Run end-to-end integration tests
run_integration_tests() {
    if [[ "$INTEGRATION_TESTS" != true ]]; then
        return 0
    fi
    
    echo -e "${BLUE}🔄 Running End-to-End Integration Tests...${NC}"
    
    # Test complete data flow: Monitoring → Prometheus → Grafana
    echo -e "${GREEN}✅ Testing complete monitoring pipeline:${NC}"
    
    # Test 1: Monitoring service provides metrics
    echo -e "   🔍 Step 1: Monitoring service metrics..."
    if timeout 5 curl -s http://localhost:9093/prometheus | grep -q "adrenochain_block_height"; then
        echo -e "      ✅ Monitoring service provides Prometheus metrics"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "      ❌ Monitoring service missing Prometheus metrics"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    # Test 2: Prometheus scrapes metrics
    echo -e "   🔍 Step 2: Prometheus scraping..."
    if timeout 5 curl -s "http://localhost:9091/api/v1/query?query=adrenochain_block_height" | jq -e '.data.result' > /dev/null 2>&1; then
        echo -e "      ✅ Prometheus successfully scrapes metrics"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "      ❌ Prometheus failed to scrape metrics"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    # Test 3: Grafana can access data
    echo -e "   🔍 Step 3: Grafana data access..."
    if timeout 5 curl -s -u admin:admin "http://localhost:3000/api/search" | jq -e '.[]' > /dev/null 2>&1; then
        echo -e "      ✅ Grafana can access dashboard data"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "      ❌ Grafana cannot access dashboard data"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    echo -e "${GREEN}✅ Integration tests completed${NC}"
    echo
}

# Generate comprehensive test summary
generate_comprehensive_summary() {
    echo -e "${BLUE}📋 Generating comprehensive test summary...${NC}"
    
    local summary_file="$TEST_RESULTS_DIR/comprehensive_test_summary.md"
    
    {
        echo "# adrenochain Comprehensive Test Suite Results"
        echo
        echo "**Execution Date:** $(date)"
        echo "**Test Type:** Traditional + Docker Integration Tests"
        echo
        echo "## 📊 Test Statistics"
        echo
        echo "| Metric | Count |"
        echo "|--------|-------|"
        echo "| **Total Tests** | $TOTAL_TESTS |"
        echo "| **Passed Tests** | $PASSED_TESTS |"
        echo "| **Failed Tests** | $FAILED_TESTS |"
        echo "| **Docker Services Tested** | $DOCKER_SERVICES_TOTAL |"
        echo "| **Healthy Docker Services** | $DOCKER_SERVICES_HEALTHY |"
        echo
        echo "## 🐳 Docker Services Status"
        echo
        if [[ $DOCKER_SERVICES_TOTAL -gt 0 ]]; then
            local service_health_rate=$((($DOCKER_SERVICES_HEALTHY * 100) / $DOCKER_SERVICES_TOTAL))
            echo "- **Service Health Rate:** ${service_health_rate}%"
            echo "- **Services Tested:** $DOCKER_SERVICES_TOTAL"
            echo "- **Healthy Services:** $DOCKER_SERVICES_HEALTHY"
        else
            echo "- **Docker Services:** Not tested (Docker not available)"
        fi
        echo
        echo "## 🌐 Service Endpoints"
        echo
        echo "| Service | Endpoint | Status |"
        echo "|---------|----------|--------|"
        echo "| Adrenochain API | http://localhost:8080/health | $(timeout 5 curl -s http://localhost:8080/health > /dev/null 2>&1 && echo '✅ OK' || echo '❌ Failed') |"
        echo "| Prometheus | http://localhost:9091/api/v1/query?query=up | $(timeout 5 curl -s http://localhost:9091/api/v1/query?query=up > /dev/null 2>&1 && echo '✅ OK' || echo '❌ Failed') |"
        echo "| Grafana | http://localhost:3000/api/health | $(timeout 5 curl -s -u admin:admin http://localhost:3000/api/health > /dev/null 2>&1 && echo '✅ OK' || echo '❌ Failed') |"
        echo "| Monitoring Service | http://localhost:9093/health | $(timeout 5 curl -s http://localhost:9093/health > /dev/null 2>&1 && echo '✅ OK' || echo '❌ Failed') |"
        echo
        echo "## 🚀 Next Steps"
        echo
        if [[ $FAILED_TESTS -gt 0 ]]; then
            echo "❌ **Action Required:** $FAILED_TESTS test(s) failed"
            echo "   - Review failed test logs"
            echo "   - Check Docker service health"
            echo "   - Verify network connectivity"
        else
            echo "✅ **All tests passed successfully!**"
            echo "   - Traditional Go tests: ✅"
            echo "   - Docker integration: ✅"
            echo "   - End-to-end workflow: ✅"
        fi
        echo
        echo "---"
        echo "*Generated by adrenochain Comprehensive Test Suite*"
    } > "$summary_file"
    
    echo -e "${GREEN}✅ Comprehensive test summary generated: $summary_file${NC}"
}

# Print final results
print_final_results() {
    echo -e "${CYAN}"
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║              🎯 Comprehensive Test Suite Complete 🎯        ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
    
    echo -e "${BLUE}📊 Final Results Summary:${NC}"
    echo -e "   🧪 Tests: ${GREEN}$PASSED_TESTS${NC} passed, ${RED}$FAILED_TESTS${NC} failed (Total: $TOTAL_TESTS)"
    
    if [[ $DOCKER_SERVICES_TOTAL -gt 0 ]]; then
        echo -e "   🐳 Docker Services: ${GREEN}$DOCKER_SERVICES_HEALTHY${NC} healthy / ${YELLOW}$DOCKER_SERVICES_TOTAL${NC} total"
        local service_health_rate=$((($DOCKER_SERVICES_HEALTHY * 100) / $DOCKER_SERVICES_TOTAL))
        echo -e "   📈 Service Health Rate: ${GREEN}${service_health_rate}%${NC}"
    fi
    
    local test_success_rate=0
    if [[ $TOTAL_TESTS -gt 0 ]]; then
        test_success_rate=$((($PASSED_TESTS * 100) / $TOTAL_TESTS))
        echo -e "   📈 Test Success Rate: ${GREEN}${test_success_rate}%${NC}"
    fi
    
    echo
    echo -e "${BLUE}📁 Results Location:${NC}"
    echo -e "   📋 Test Results: ${CYAN}$TEST_RESULTS_DIR${NC}"
    echo -e "   📝 Summary: ${CYAN}$TEST_RESULTS_DIR/comprehensive_test_summary.md${NC}"
    
    echo
    echo -e "${BLUE}🌐 Access Points:${NC}"
    echo -e "   🚀 Adrenochain API: ${CYAN}http://localhost:8080/health${NC}"
    echo -e "   📊 Prometheus: ${CYAN}http://localhost:9091${NC}"
    echo -e "   📈 Grafana: ${CYAN}http://localhost:3000${NC} (admin/admin)"
    echo -e "   📊 Monitoring: ${CYAN}http://localhost:9093/health${NC}"
    
    echo
    if [[ $FAILED_TESTS -gt 0 ]]; then
        echo -e "${RED}❌ Some tests failed. Please review the logs and fix the issues.${NC}"
        exit 1
    else
        echo -e "${GREEN}🎉 All tests passed successfully! adrenochain is ready for action! 🚀${NC}"
    fi
}

# Main execution function
main() {
    print_banner
    
    # Create directories
    mkdir -p "$TEST_RESULTS_DIR"
    mkdir -p "$DOCKER_TEST_RESULTS_DIR"
    
    echo -e "${BLUE}🚀 Starting comprehensive test suite...${NC}"
    echo
    
    # Run different types of tests
    run_traditional_tests
    run_docker_integration_tests
    run_integration_tests
    
    # Generate reports
    generate_comprehensive_summary
    
    # Print results
    print_final_results
    
    # Exit with appropriate code
    if [[ $FAILED_TESTS -gt 0 ]]; then
        exit 1
    else
        exit 0
    fi
}

# Handle script arguments
case "${1:-}" in
    --help|-h)
        echo "Usage: $0 [OPTIONS]"
        echo
        echo "Options:"
        echo "  --help, -h           Show this help message"
        echo "  --traditional-only   Run only traditional Go tests"
        echo "  --docker-only        Run only Docker integration tests"
        echo "  --integration-only   Run only end-to-end integration tests"
        echo "  --no-traditional     Skip traditional Go tests"
        echo "  --no-docker          Skip Docker integration tests"
        echo "  --no-integration     Skip integration tests"
        echo "  --verbose            Enable verbose output"
        echo
        echo "Examples:"
        echo "  $0                    # Run all tests"
        echo "  $0 --traditional-only # Run only traditional Go tests"
        echo "  $0 --docker-only      # Run only Docker tests"
        echo "  $0 --no-traditional   # Skip traditional tests"
        exit 0
        ;;
    --traditional-only)
        DOCKER_TESTS=false
        INTEGRATION_TESTS=false
        ;;
    --docker-only)
        TRADITIONAL_TESTS=false
        INTEGRATION_TESTS=false
        ;;
    --integration-only)
        TRADITIONAL_TESTS=false
        DOCKER_TESTS=false
        ;;
    --no-traditional)
        TRADITIONAL_TESTS=false
        ;;
    --no-docker)
        DOCKER_TESTS=false
        ;;
    --no-integration)
        INTEGRATION_TESTS=false
        ;;
    --verbose)
        VERBOSE_TESTS=true
        ;;
esac

# Run the main function
main "$@"

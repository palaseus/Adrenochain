#!/bin/bash

# 🐳 adrenochain Docker Test Suite
# This script provides comprehensive testing for the Docker-based adrenochain implementation
# It runs tests both inside Docker containers and against the running Docker services

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
LOG_FILE="$DOCKER_TEST_RESULTS_DIR/docker_test_suite.log"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

# Test configuration
TIMEOUT=300s  # 5 minutes per test
COVERAGE_ENABLED=true
VERBOSE_TESTS=true
DOCKER_TESTS=true
INTEGRATION_TESTS=true
MONITORING_TESTS=true
API_TESTS=true
PERFORMANCE_TESTS=true

# Statistics
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
SKIPPED_TESTS=0
DOCKER_SERVICES_TESTED=0
DOCKER_SERVICES_HEALTHY=0

# Initialize test environment
init_docker_test_environment() {
    echo -e "${BLUE}🐳 Initializing Docker Test Suite...${NC}"
    
    # Create directories
    mkdir -p "$TEST_RESULTS_DIR"
    mkdir -p "$DOCKER_TEST_RESULTS_DIR"
    
    # Clean previous results
    rm -rf "$DOCKER_TEST_RESULTS_DIR"/*
    
    # Reset counters
    TOTAL_TESTS=0
    PASSED_TESTS=0
    FAILED_TESTS=0
    SKIPPED_TESTS=0
    DOCKER_SERVICES_TESTED=0
    DOCKER_SERVICES_HEALTHY=0
    
    # Start log file
    {
        echo "=== adrenochain Docker Test Suite Execution ==="
        echo "Timestamp: $(date)"
        echo "Project Root: $PROJECT_ROOT"
        echo "Docker Test Results Dir: $DOCKER_TEST_RESULTS_DIR"
        echo "============================================="
        echo
    } > "$LOG_FILE"
    
    echo -e "${GREEN}✅ Docker test environment initialized${NC}"
}

# Print banner
print_banner() {
    echo -e "${CYAN}"
    cat << "EOF"
    ╔══════════════════════════════════════════════════════════════╗
    ║                🐳 adrenochain Docker Test Suite 🐳           ║
    ║                                                              ║
    ║  Comprehensive testing for Docker-based adrenochain          ║
    ║                                                              ║
    ║  Features:                                                   ║
    ║  • Docker service health checks                              ║
    ║  • Container integration tests                               ║
    ║  • API endpoint validation                                   ║
    ║  • Monitoring system tests                                   ║
    ║  • Performance benchmarking                                  ║
    ║  • End-to-end workflow tests                                 ║
    ║  • Timeout-protected operations                              ║
    ╚══════════════════════════════════════════════════════════════╝
EOF
    echo -e "${NC}"
}

# Check Docker prerequisites
check_docker_prerequisites() {
    echo -e "${BLUE}🔍 Checking Docker prerequisites...${NC}"
    
    # Check Docker installation
    if ! command -v docker &> /dev/null; then
        echo -e "${RED}❌ Docker is not installed or not in PATH${NC}"
        exit 1
    fi
    
    # Check Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        echo -e "${RED}❌ Docker Compose is not installed or not in PATH${NC}"
        exit 1
    fi
    
    # Check if Docker daemon is running
    if ! timeout 5 sudo docker info &> /dev/null; then
        echo -e "${RED}❌ Docker daemon is not running${NC}"
        exit 1
    fi
    
    # Check if we're in the right directory
    if [[ ! -f "$PROJECT_ROOT/docker-compose.yml" ]]; then
        echo -e "${RED}❌ Not in adrenochain project root (docker-compose.yml not found)${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✅ Docker prerequisites check passed${NC}"
}

# Test Docker services health
test_docker_services_health() {
    echo -e "${BLUE}🐳 Testing Docker Services Health...${NC}"
    
    local services=("adrenochain-node" "adrenochain-prometheus" "adrenochain-grafana" "adrenochain-redis")
    local service_count=${#services[@]}
    DOCKER_SERVICES_TESTED=$service_count
    
    echo -e "${GREEN}✅ Testing $service_count Docker services:${NC}"
    
    for service in "${services[@]}"; do
        echo -e "   🐳 Testing $service..."
        
        # Check if container is running
        if timeout 10 sudo docker ps --format "table {{.Names}}\t{{.Status}}" | grep -q "$service"; then
            echo -e "      ✅ $service container is running"
            
            # Check container health
            local health_status=$(timeout 10 sudo docker inspect --format='{{.State.Health.Status}}' "$service" 2>/dev/null || echo "unknown")
            if [[ "$health_status" == "healthy" ]]; then
                echo -e "      ✅ $service health check: healthy"
                DOCKER_SERVICES_HEALTHY=$((DOCKER_SERVICES_HEALTHY + 1))
            elif [[ "$health_status" == "unhealthy" ]]; then
                echo -e "      ❌ $service health check: unhealthy"
            else
                echo -e "      ⚠️  $service health check: $health_status"
            fi
            
            # Test service-specific endpoints
            case "$service" in
                "adrenochain-node")
                    test_adrenochain_node_endpoints
                    ;;
                "adrenochain-prometheus")
                    test_prometheus_endpoints
                    ;;
                "adrenochain-grafana")
                    test_grafana_endpoints
                    ;;
                "adrenochain-redis")
                    test_redis_endpoints
                    ;;
            esac
            
        else
            echo -e "      ❌ $service container is not running"
        fi
        echo
    done
    
    echo -e "${GREEN}✅ Docker services health test completed${NC}"
    echo -e "   📊 Services tested: $DOCKER_SERVICES_TESTED"
    echo -e "   ✅ Healthy services: $DOCKER_SERVICES_HEALTHY"
    echo
}

# Test Adrenochain node endpoints
test_adrenochain_node_endpoints() {
    echo -e "      🌐 Testing Adrenochain API endpoints..."
    
    local endpoints=(
        "http://localhost:8080/health"
        "http://localhost:8080/api/v1/chain/info"
        "http://localhost:8080/api/v1/network/status"
    )
    
    for endpoint in "${endpoints[@]}"; do
        if timeout 5 curl -s "$endpoint" > /dev/null 2>&1; then
            echo -e "         ✅ $endpoint - responding"
            PASSED_TESTS=$((PASSED_TESTS + 1))
        else
            echo -e "         ❌ $endpoint - not responding"
            FAILED_TESTS=$((FAILED_TESTS + 1))
        fi
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
    done
}

# Test Prometheus endpoints
test_prometheus_endpoints() {
    echo -e "      📊 Testing Prometheus endpoints..."
    
    local endpoints=(
        "http://localhost:9091/api/v1/query?query=up"
        "http://localhost:9091/api/v1/targets"
        "http://localhost:9091/api/v1/query?query=adrenochain_block_height"
    )
    
    for endpoint in "${endpoints[@]}"; do
        if timeout 5 curl -s "$endpoint" > /dev/null 2>&1; then
            echo -e "         ✅ $endpoint - responding"
            PASSED_TESTS=$((PASSED_TESTS + 1))
        else
            echo -e "         ❌ $endpoint - not responding"
            FAILED_TESTS=$((FAILED_TESTS + 1))
        fi
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
    done
}

# Test Grafana endpoints
test_grafana_endpoints() {
    echo -e "      📈 Testing Grafana endpoints..."
    
    local endpoints=(
        "http://localhost:3000/api/health"
        "http://localhost:3000/api/search"
    )
    
    for endpoint in "${endpoints[@]}"; do
        if timeout 5 curl -s -u admin:admin "$endpoint" > /dev/null 2>&1; then
            echo -e "         ✅ $endpoint - responding"
            PASSED_TESTS=$((PASSED_TESTS + 1))
        else
            echo -e "         ❌ $endpoint - not responding"
            FAILED_TESTS=$((FAILED_TESTS + 1))
        fi
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
    done
}

# Test Redis endpoints
test_redis_endpoints() {
    echo -e "      🗄️ Testing Redis endpoints..."
    
    if timeout 5 sudo docker exec adrenochain-redis redis-cli ping > /dev/null 2>&1; then
        echo -e "         ✅ Redis PING - responding"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "         ❌ Redis PING - not responding"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
}

# Test monitoring service
test_monitoring_service() {
    echo -e "${BLUE}📊 Testing Monitoring Service...${NC}"
    
    echo -e "${GREEN}✅ Testing monitoring endpoints:${NC}"
    
    local monitoring_endpoints=(
        "http://localhost:9093/health"
        "http://localhost:9093/metrics"
        "http://localhost:9093/prometheus"
    )
    
    for endpoint in "${monitoring_endpoints[@]}"; do
        echo -e "   📊 Testing $endpoint..."
        
        if timeout 5 curl -s "$endpoint" > /dev/null 2>&1; then
            echo -e "      ✅ $endpoint - responding"
            PASSED_TESTS=$((PASSED_TESTS + 1))
        else
            echo -e "      ❌ $endpoint - not responding"
            FAILED_TESTS=$((FAILED_TESTS + 1))
        fi
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
    done
    
    # Test metrics data quality
    echo -e "   📊 Testing metrics data quality..."
    local metrics_response=$(timeout 5 curl -s http://localhost:9093/metrics 2>/dev/null || echo "")
    if [[ -n "$metrics_response" ]] && echo "$metrics_response" | jq -e '.blockchain.height' > /dev/null 2>&1; then
        echo -e "      ✅ Metrics data structure is valid"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "      ❌ Metrics data structure is invalid"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    echo -e "${GREEN}✅ Monitoring service test completed${NC}"
    echo
}

# Test end-to-end workflow
test_end_to_end_workflow() {
    echo -e "${BLUE}🔄 Testing End-to-End Workflow...${NC}"
    
    echo -e "${GREEN}✅ Testing complete data flow:${NC}"
    echo -e "   📊 Monitoring Service → Prometheus → Grafana"
    
    # Test 1: Monitoring service provides metrics
    echo -e "   🔍 Step 1: Monitoring service metrics..."
    local monitoring_metrics=$(timeout 5 curl -s http://localhost:9093/prometheus 2>/dev/null || echo "")
    if [[ -n "$monitoring_metrics" ]] && echo "$monitoring_metrics" | grep -q "adrenochain_block_height"; then
        echo -e "      ✅ Monitoring service provides Prometheus metrics"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "      ❌ Monitoring service missing Prometheus metrics"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    # Test 2: Prometheus scrapes metrics
    echo -e "   🔍 Step 2: Prometheus scraping..."
    local prometheus_metrics=$(timeout 5 curl -s "http://localhost:9091/api/v1/query?query=adrenochain_block_height" 2>/dev/null || echo "")
    if [[ -n "$prometheus_metrics" ]] && echo "$prometheus_metrics" | jq -e '.data.result' > /dev/null 2>&1; then
        echo -e "      ✅ Prometheus successfully scrapes metrics"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "      ❌ Prometheus failed to scrape metrics"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    # Test 3: Grafana can access data
    echo -e "   🔍 Step 3: Grafana data access..."
    local grafana_dashboards=$(timeout 5 curl -s -u admin:admin "http://localhost:3000/api/search" 2>/dev/null || echo "")
    if [[ -n "$grafana_dashboards" ]] && echo "$grafana_dashboards" | jq -e '.[]' > /dev/null 2>&1; then
        echo -e "      ✅ Grafana can access dashboard data"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "      ❌ Grafana cannot access dashboard data"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    echo -e "${GREEN}✅ End-to-end workflow test completed${NC}"
    echo
}

# Test performance benchmarks
test_performance_benchmarks() {
    echo -e "${BLUE}📊 Testing Performance Benchmarks...${NC}"
    
    echo -e "${GREEN}✅ Running performance tests:${NC}"
    
    # Test API response times
    echo -e "   ⏱️ Testing API response times..."
    local start_time=$(date +%s%N)
    if timeout 5 curl -s http://localhost:8080/health > /dev/null 2>&1; then
        local end_time=$(date +%s%N)
        local response_time=$(( (end_time - start_time) / 1000000 ))  # Convert to milliseconds
        echo -e "      ✅ API health endpoint: ${response_time}ms"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "      ❌ API health endpoint: timeout"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    # Test monitoring service response times
    echo -e "   ⏱️ Testing monitoring service response times..."
    local start_time=$(date +%s%N)
    if timeout 5 curl -s http://localhost:9093/health > /dev/null 2>&1; then
        local end_time=$(date +%s%N)
        local response_time=$(( (end_time - start_time) / 1000000 ))  # Convert to milliseconds
        echo -e "      ✅ Monitoring health endpoint: ${response_time}ms"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "      ❌ Monitoring health endpoint: timeout"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    # Test Prometheus query performance
    echo -e "   ⏱️ Testing Prometheus query performance..."
    local start_time=$(date +%s%N)
    if timeout 5 curl -s "http://localhost:9091/api/v1/query?query=up" > /dev/null 2>&1; then
        local end_time=$(date +%s%N)
        local response_time=$(( (end_time - start_time) / 1000000 ))  # Convert to milliseconds
        echo -e "      ✅ Prometheus query: ${response_time}ms"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "      ❌ Prometheus query: timeout"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    echo -e "${GREEN}✅ Performance benchmarks completed${NC}"
    echo
}

# Test Go unit tests inside Docker
test_go_unit_tests_in_docker() {
    echo -e "${BLUE}🧪 Testing Go Unit Tests in Docker...${NC}"
    
    echo -e "${GREEN}✅ Running Go tests inside Docker container:${NC}"
    
    # Run tests inside the adrenochain container
    echo -e "   🐳 Running tests in adrenochain-node container..."
    if timeout 60 sudo docker exec adrenochain-node go test ./pkg/... -v 2>&1 | tee "$DOCKER_TEST_RESULTS_DIR/go_unit_tests.log"; then
        echo -e "      ✅ Go unit tests passed"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "      ❌ Go unit tests failed"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    echo -e "${GREEN}✅ Go unit tests in Docker completed${NC}"
    echo
}

# Generate Docker test summary
generate_docker_test_summary() {
    echo -e "${BLUE}📋 Generating Docker test summary...${NC}"
    
    local summary_file="$DOCKER_TEST_RESULTS_DIR/docker_test_summary.md"
    
    {
        echo "# adrenochain Docker Test Suite Results"
        echo
        echo "**Execution Date:** $(date)"
        echo "**Test Type:** Docker Integration Tests"
        echo
        echo "## 🐳 Docker Services Status"
        echo
        echo "| Service | Status | Health |"
        echo "|---------|--------|--------|"
        echo "| adrenochain-node | $(timeout 5 sudo docker ps --format '{{.Status}}' --filter name=adrenochain-node 2>/dev/null || echo 'Not running') | $(timeout 5 sudo docker inspect --format='{{.State.Health.Status}}' adrenochain-node 2>/dev/null || echo 'unknown') |"
        echo "| adrenochain-prometheus | $(timeout 5 sudo docker ps --format '{{.Status}}' --filter name=adrenochain-prometheus 2>/dev/null || echo 'Not running') | $(timeout 5 sudo docker inspect --format='{{.State.Health.Status}}' adrenochain-prometheus 2>/dev/null || echo 'unknown') |"
        echo "| adrenochain-grafana | $(timeout 5 sudo docker ps --format '{{.Status}}' --filter name=adrenochain-grafana 2>/dev/null || echo 'Not running') | $(timeout 5 sudo docker inspect --format='{{.State.Health.Status}}' adrenochain-grafana 2>/dev/null || echo 'unknown') |"
        echo "| adrenochain-redis | $(timeout 5 sudo docker ps --format '{{.Status}}' --filter name=adrenochain-redis 2>/dev/null || echo 'Not running') | $(timeout 5 sudo docker inspect --format='{{.State.Health.Status}}' adrenochain-redis 2>/dev/null || echo 'unknown') |"
        echo
        echo "## 📊 Test Statistics"
        echo
        echo "| Metric | Count |"
        echo "|--------|-------|"
        echo "| **Total Tests** | $TOTAL_TESTS |"
        echo "| **Passed Tests** | $PASSED_TESTS |"
        echo "| **Failed Tests** | $FAILED_TESTS |"
        echo "| **Docker Services Tested** | $DOCKER_SERVICES_TESTED |"
        echo "| **Healthy Services** | $DOCKER_SERVICES_HEALTHY |"
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
            echo "✅ **All Docker tests passed successfully!**"
            echo "   - Docker services are healthy"
            echo "   - All endpoints are responding"
            echo "   - Monitoring integration is working"
        fi
        echo
        echo "---"
        echo "*Generated by adrenochain Docker Test Suite*"
    } > "$summary_file"
    
    echo -e "${GREEN}✅ Docker test summary generated: $summary_file${NC}"
}

# Print final results
print_final_results() {
    echo -e "${CYAN}"
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║                🎯 Docker Test Suite Complete 🎯              ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
    
    echo -e "${BLUE}📊 Final Results Summary:${NC}"
    echo -e "   🐳 Docker Services: ${GREEN}$DOCKER_SERVICES_HEALTHY${NC} healthy / ${YELLOW}$DOCKER_SERVICES_TESTED${NC} total"
    echo -e "   🧪 Tests: ${GREEN}$PASSED_TESTS${NC} passed, ${RED}$FAILED_TESTS${NC} failed (Total: $TOTAL_TESTS)"
    
    local service_health_rate=0
    if [[ $DOCKER_SERVICES_TESTED -gt 0 ]]; then
        service_health_rate=$((($DOCKER_SERVICES_HEALTHY * 100) / $DOCKER_SERVICES_TESTED))
        echo -e "   📈 Service Health Rate: ${GREEN}${service_health_rate}%${NC}"
    fi
    
    local test_success_rate=0
    if [[ $TOTAL_TESTS -gt 0 ]]; then
        test_success_rate=$((($PASSED_TESTS * 100) / $TOTAL_TESTS))
        echo -e "   📈 Test Success Rate: ${GREEN}${test_success_rate}%${NC}"
    fi
    
    echo
    echo -e "${BLUE}📁 Results Location:${NC}"
    echo -e "   📋 Docker Test Results: ${CYAN}$DOCKER_TEST_RESULTS_DIR${NC}"
    echo -e "   📝 Summary: ${CYAN}$DOCKER_TEST_RESULTS_DIR/docker_test_summary.md${NC}"
    echo -e "   📄 Log: ${CYAN}$LOG_FILE${NC}"
    
    echo
    echo -e "${BLUE}🌐 Access Points:${NC}"
    echo -e "   🚀 Adrenochain API: ${CYAN}http://localhost:8080/health${NC}"
    echo -e "   📊 Prometheus: ${CYAN}http://localhost:9091${NC}"
    echo -e "   📈 Grafana: ${CYAN}http://localhost:3000${NC} (admin/admin)"
    echo -e "   📊 Monitoring: ${CYAN}http://localhost:9093/health${NC}"
    
    echo
    if [[ $FAILED_TESTS -gt 0 ]]; then
        echo -e "${RED}❌ Some Docker tests failed. Please review the logs and fix the issues.${NC}"
        exit 1
    else
        echo -e "${GREEN}🎉 All Docker tests passed successfully! adrenochain Docker setup is ready! 🐳${NC}"
    fi
}

# Main execution function
main() {
    print_banner
    check_docker_prerequisites
    init_docker_test_environment
    
    echo -e "${BLUE}🐳 Starting Docker test suite...${NC}"
    echo
    
    # Run Docker-specific tests
    test_docker_services_health
    
    if [[ "$MONITORING_TESTS" == true ]]; then
        test_monitoring_service
    fi
    
    if [[ "$INTEGRATION_TESTS" == true ]]; then
        test_end_to_end_workflow
    fi
    
    if [[ "$PERFORMANCE_TESTS" == true ]]; then
        test_performance_benchmarks
    fi
    
    if [[ "$DOCKER_TESTS" == true ]]; then
        test_go_unit_tests_in_docker
    fi
    
    # Generate reports
    generate_docker_test_summary
    
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
        echo "  --no-monitoring      Disable monitoring tests"
        echo "  --no-integration     Disable integration tests"
        echo "  --no-performance     Disable performance tests"
        echo "  --no-docker-tests    Disable Docker Go unit tests"
        echo "  --services-only      Test only Docker services health"
        echo "  --monitoring-only    Test only monitoring system"
        echo "  --verbose            Enable verbose output"
        echo
        echo "Examples:"
        echo "  $0                    # Run all Docker tests"
        echo "  $0 --services-only   # Test only Docker services"
        echo "  $0 --monitoring-only # Test only monitoring system"
        echo "  $0 --no-performance  # Skip performance tests"
        exit 0
        ;;
    --no-monitoring)
        MONITORING_TESTS=false
        shift
        ;;
    --no-integration)
        INTEGRATION_TESTS=false
        shift
        ;;
    --no-performance)
        PERFORMANCE_TESTS=false
        shift
        ;;
    --no-docker-tests)
        DOCKER_TESTS=false
        shift
        ;;
    --services-only)
        echo -e "${BLUE}🐳 Running Docker Services Health Tests Only...${NC}"
        test_docker_services_health
        exit 0
        ;;
    --monitoring-only)
        echo -e "${BLUE}📊 Running Monitoring System Tests Only...${NC}"
        test_monitoring_service
        exit 0
        ;;
    --verbose)
        VERBOSE_TESTS=true
        shift
        ;;
esac

# Run the main function
main "$@"

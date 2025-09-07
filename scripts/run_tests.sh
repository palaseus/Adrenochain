#!/bin/bash

# 🚀 adrenochain Test Runner
# This script provides a unified interface to run different types of tests

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# Print banner
print_banner() {
    echo -e "${CYAN}"
    cat << "EOF"
    ╔══════════════════════════════════════════════════════════════╗
    ║                    🚀 adrenochain Test Runner 🚀           ║
    ║                                                              ║
    ║  Unified interface for running adrenochain tests            ║
    ║                                                              ║
    ║  Available test suites:                                      ║
    ║  • Traditional Go tests (test_suite.sh)                     ║
    ║  • Docker integration tests (docker_test_suite.sh)          ║
    ║  • Comprehensive tests (test_suite_docker_integration.sh)   ║
    ║  • Health checks (health_check.sh)                          ║
    ║  • Monitoring tests (test_complete_monitoring_with_timeouts.sh) ║
    ╚══════════════════════════════════════════════════════════════╝
EOF
    echo -e "${NC}"
}

# Show help
show_help() {
    echo "Usage: $0 [TEST_TYPE] [OPTIONS]"
    echo
    echo "Test Types:"
    echo "  traditional     Run traditional Go tests (test_suite.sh)"
    echo "  docker          Run Docker integration tests (docker_test_suite.sh)"
    echo "  comprehensive   Run comprehensive tests with Docker integration"
    echo "  health          Run health checks (health_check.sh)"
    echo "  monitoring      Run monitoring system tests"
    echo "  all             Run all available tests"
    echo
    echo "Options:"
    echo "  --help, -h      Show this help message"
    echo "  --quick         Run quick tests only (skip long-running tests)"
    echo "  --verbose       Enable verbose output"
    echo
    echo "Examples:"
    echo "  $0 traditional           # Run traditional Go tests"
    echo "  $0 docker               # Run Docker integration tests"
    echo "  $0 comprehensive        # Run comprehensive tests"
    echo "  $0 health               # Run health checks"
    echo "  $0 monitoring           # Run monitoring tests"
    echo "  $0 all                  # Run all tests"
    echo "  $0 docker --quick       # Run quick Docker tests"
    echo
    echo "Available test scripts:"
    echo "  📄 scripts/test_suite.sh                    - Traditional Go tests"
    echo "  📄 scripts/docker_test_suite.sh             - Docker integration tests"
    echo "  📄 scripts/test_suite_docker_integration.sh - Comprehensive tests"
    echo "  📄 scripts/health_check.sh                  - Health checks"
    echo "  📄 scripts/test_complete_monitoring_with_timeouts.sh - Monitoring tests"
}

# Check if a script exists and is executable
check_script() {
    local script_path="$1"
    if [[ -f "$script_path" && -x "$script_path" ]]; then
        return 0
    else
        return 1
    fi
}

# Run traditional tests
run_traditional_tests() {
    echo -e "${BLUE}🧪 Running Traditional Go Tests...${NC}"
    
    local script_path="$PROJECT_ROOT/scripts/test_suite.sh"
    if check_script "$script_path"; then
        echo -e "${GREEN}✅ Found test_suite.sh, running it...${NC}"
        if [[ "$QUICK_MODE" == true ]]; then
            "$script_path" --no-race --no-fuzz --no-bench --timeout 60s
        else
            "$script_path"
        fi
    else
        echo -e "${RED}❌ test_suite.sh not found or not executable${NC}"
        return 1
    fi
}

# Run Docker tests
run_docker_tests() {
    echo -e "${BLUE}🐳 Running Docker Integration Tests...${NC}"
    
    local script_path="$PROJECT_ROOT/scripts/docker_test_suite.sh"
    if check_script "$script_path"; then
        echo -e "${GREEN}✅ Found docker_test_suite.sh, running it...${NC}"
        if [[ "$QUICK_MODE" == true ]]; then
            "$script_path" --services-only
        else
            "$script_path"
        fi
    else
        echo -e "${RED}❌ docker_test_suite.sh not found or not executable${NC}"
        return 1
    fi
}

# Run comprehensive tests
run_comprehensive_tests() {
    echo -e "${BLUE}🔄 Running Comprehensive Tests...${NC}"
    
    local script_path="$PROJECT_ROOT/scripts/test_suite_docker_integration.sh"
    if check_script "$script_path"; then
        echo -e "${GREEN}✅ Found test_suite_docker_integration.sh, running it...${NC}"
        if [[ "$QUICK_MODE" == true ]]; then
            "$script_path" --docker-only
        else
            "$script_path"
        fi
    else
        echo -e "${RED}❌ test_suite_docker_integration.sh not found or not executable${NC}"
        return 1
    fi
}

# Run health checks
run_health_checks() {
    echo -e "${BLUE}🏥 Running Health Checks...${NC}"
    
    local script_path="$PROJECT_ROOT/scripts/health_check.sh"
    if check_script "$script_path"; then
        echo -e "${GREEN}✅ Found health_check.sh, running it...${NC}"
        sudo "$script_path"
    else
        echo -e "${RED}❌ health_check.sh not found or not executable${NC}"
        return 1
    fi
}

# Run monitoring tests
run_monitoring_tests() {
    echo -e "${BLUE}📊 Running Monitoring Tests...${NC}"
    
    local script_path="$PROJECT_ROOT/scripts/test_complete_monitoring_with_timeouts.sh"
    if check_script "$script_path"; then
        echo -e "${GREEN}✅ Found test_complete_monitoring_with_timeouts.sh, running it...${NC}"
        "$script_path"
    else
        echo -e "${RED}❌ test_complete_monitoring_with_timeouts.sh not found or not executable${NC}"
        return 1
    fi
}

# Run all tests
run_all_tests() {
    echo -e "${BLUE}🚀 Running All Available Tests...${NC}"
    
    local overall_success=true
    
    # Run health checks first
    echo -e "${YELLOW}📋 Step 1/5: Health Checks${NC}"
    if ! run_health_checks; then
        echo -e "${YELLOW}⚠️  Health checks failed, continuing with other tests...${NC}"
        overall_success=false
    fi
    echo
    
    # Run monitoring tests
    echo -e "${YELLOW}📋 Step 2/5: Monitoring Tests${NC}"
    if ! run_monitoring_tests; then
        echo -e "${YELLOW}⚠️  Monitoring tests failed, continuing with other tests...${NC}"
        overall_success=false
    fi
    echo
    
    # Run Docker tests
    echo -e "${YELLOW}📋 Step 3/5: Docker Integration Tests${NC}"
    if ! run_docker_tests; then
        echo -e "${YELLOW}⚠️  Docker tests failed, continuing with other tests...${NC}"
        overall_success=false
    fi
    echo
    
    # Run comprehensive tests
    echo -e "${YELLOW}📋 Step 4/5: Comprehensive Tests${NC}"
    if ! run_comprehensive_tests; then
        echo -e "${YELLOW}⚠️  Comprehensive tests failed, continuing with other tests...${NC}"
        overall_success=false
    fi
    echo
    
    # Run traditional tests last (they take the longest)
    echo -e "${YELLOW}📋 Step 5/5: Traditional Go Tests${NC}"
    if ! run_traditional_tests; then
        echo -e "${YELLOW}⚠️  Traditional tests failed${NC}"
        overall_success=false
    fi
    
    echo
    if [[ "$overall_success" == true ]]; then
        echo -e "${GREEN}🎉 All tests completed successfully!${NC}"
    else
        echo -e "${YELLOW}⚠️  Some tests failed, but all test suites were executed${NC}"
    fi
}

# Main execution
main() {
    print_banner
    
    # Parse arguments
    local test_type=""
    QUICK_MODE=false
    VERBOSE=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --help|-h)
                show_help
                exit 0
                ;;
            --quick)
                QUICK_MODE=true
                shift
                ;;
            --verbose)
                VERBOSE=true
                shift
                ;;
            traditional|docker|comprehensive|health|monitoring|all)
                test_type="$1"
                shift
                ;;
            *)
                echo -e "${RED}❌ Unknown option: $1${NC}"
                show_help
                exit 1
                ;;
        esac
    done
    
    # If no test type specified, show help
    if [[ -z "$test_type" ]]; then
        show_help
        exit 0
    fi
    
    # Run the specified test type
    case "$test_type" in
        traditional)
            run_traditional_tests
            ;;
        docker)
            run_docker_tests
            ;;
        comprehensive)
            run_comprehensive_tests
            ;;
        health)
            run_health_checks
            ;;
        monitoring)
            run_monitoring_tests
            ;;
        all)
            run_all_tests
            ;;
        *)
            echo -e "${RED}❌ Unknown test type: $test_type${NC}"
            show_help
            exit 1
            ;;
    esac
}

# Run main function
main "$@"

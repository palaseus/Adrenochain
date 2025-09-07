#!/bin/bash

# 🏥 Adrenochain Health Check & Monitoring Script
# This script provides comprehensive health monitoring for the Adrenochain ecosystem

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
HEALTH_LOG="$PROJECT_ROOT/logs/health_check.log"
ALERT_LOG="$PROJECT_ROOT/logs/alerts.log"

# Create logs directory
mkdir -p "$PROJECT_ROOT/logs"

# Health check function
check_health() {
    local service="$1"
    local check_command="$2"
    local expected_output="$3"
    
    echo -e "${BLUE}🔍 Checking: $service${NC}"
    
    if eval "$check_command" > /dev/null 2>&1; then
        if [ -n "$expected_output" ]; then
            if eval "$check_command" | grep -q "$expected_output"; then
                echo -e "${GREEN}✅ $service: HEALTHY${NC}"
                return 0
            else
                echo -e "${RED}❌ $service: UNHEALTHY (unexpected output)${NC}"
                return 1
            fi
        else
            echo -e "${GREEN}✅ $service: HEALTHY${NC}"
            return 0
        fi
    else
        echo -e "${RED}❌ $service: UNHEALTHY${NC}"
        return 1
    fi
}

# Alert function
send_alert() {
    local message="$1"
    local severity="$2"
    
    echo "$(date): [$severity] $message" >> "$ALERT_LOG"
    
    case "$severity" in
        "CRITICAL")
            echo -e "${RED}🚨 CRITICAL ALERT: $message${NC}"
            ;;
        "WARNING")
            echo -e "${YELLOW}⚠️  WARNING: $message${NC}"
            ;;
        "INFO")
            echo -e "${BLUE}ℹ️  INFO: $message${NC}"
            ;;
    esac
}

# Log function
log_health() {
    local message="$1"
    echo "$(date): $message" >> "$HEALTH_LOG"
}

# Main health check function
main_health_check() {
    echo -e "${CYAN}"
    cat << "EOF"
╔══════════════════════════════════════════════════════════════╗
║                🏥 Adrenochain Health Check 🏥               ║
║                                                              ║
║  Comprehensive health monitoring for the entire ecosystem    ║
║                                                              ║
╚══════════════════════════════════════════════════════════════╝
EOF
    echo -e "${NC}"
    
    log_health "Starting comprehensive health check"
    
    local healthy_services=0
    local total_services=0
    
    # ===== CORE BLOCKCHAIN HEALTH =====
    echo -e "${YELLOW}🔗 CORE BLOCKCHAIN HEALTH${NC}"
    echo "=========================="
    
    # Check if main process is running
    total_services=$((total_services + 1))
    if pgrep -f "adrenochain.*--config" > /dev/null; then
        echo -e "${GREEN}✅ Main Process: RUNNING${NC}"
        healthy_services=$((healthy_services + 1))
    else
        echo -e "${RED}❌ Main Process: NOT RUNNING${NC}"
        send_alert "Main Adrenochain process is not running" "CRITICAL"
    fi
    
    # Check API health
    total_services=$((total_services + 1))
    if check_health "API Server" "curl -s http://localhost:8080/health" "healthy"; then
        healthy_services=$((healthy_services + 1))
    else
        send_alert "API server is not responding" "CRITICAL"
    fi
    
    # Check blockchain info
    total_services=$((total_services + 1))
    if check_health "Blockchain Info" "curl -s http://localhost:8080/api/v1/chain/info" "height"; then
        healthy_services=$((healthy_services + 1))
    else
        send_alert "Blockchain info endpoint not responding" "WARNING"
    fi
    
    # Check network status
    total_services=$((total_services + 1))
    if check_health "Network Status" "curl -s http://localhost:8080/api/v1/network/status" "active"; then
        healthy_services=$((healthy_services + 1))
    else
        send_alert "Network status endpoint not responding" "WARNING"
    fi
    
    # ===== DOCKER SERVICES HEALTH =====
    echo -e "${YELLOW}🐳 DOCKER SERVICES HEALTH${NC}"
    echo "=========================="
    
    # Check Docker daemon
    total_services=$((total_services + 1))
    if check_health "Docker Daemon" "docker info" ""; then
        healthy_services=$((healthy_services + 1))
    else
        send_alert "Docker daemon is not running" "CRITICAL"
    fi
    
    # Check Docker Compose services
    if command -v docker-compose > /dev/null; then
        total_services=$((total_services + 1))
        if check_health "Docker Compose" "docker-compose ps" ""; then
            healthy_services=$((healthy_services + 1))
        else
            send_alert "Docker Compose services are not running" "WARNING"
        fi
        
        # Check individual services
        for service in adrenochain prometheus grafana redis; do
            total_services=$((total_services + 1))
            if docker-compose ps | grep -q "$service.*Up"; then
                echo -e "${GREEN}✅ $service: RUNNING${NC}"
                healthy_services=$((healthy_services + 1))
            else
                echo -e "${RED}❌ $service: NOT RUNNING${NC}"
                send_alert "Docker service $service is not running" "WARNING"
            fi
        done
    fi
    
    # ===== MONITORING HEALTH =====
    echo -e "${YELLOW}📊 MONITORING HEALTH${NC}"
    echo "====================="
    
    # Check Prometheus
    total_services=$((total_services + 1))
    if check_health "Prometheus" "curl -s http://localhost:9091/-/healthy" ""; then
        healthy_services=$((healthy_services + 1))
    else
        send_alert "Prometheus is not responding" "WARNING"
    fi
    
    # Check Grafana
    total_services=$((total_services + 1))
    if check_health "Grafana" "curl -s http://localhost:3000/api/health" ""; then
        healthy_services=$((healthy_services + 1))
    else
        send_alert "Grafana is not responding" "WARNING"
    fi
    
    # Check Redis
    total_services=$((total_services + 1))
    if check_health "Redis" "redis-cli ping" "PONG"; then
        healthy_services=$((healthy_services + 1))
    else
        send_alert "Redis is not responding" "WARNING"
    fi
    
    # ===== SYSTEM HEALTH =====
    echo -e "${YELLOW}💻 SYSTEM HEALTH${NC}"
    echo "=================="
    
    # Check disk space
    total_services=$((total_services + 1))
    disk_usage=$(df / | awk 'NR==2 {print $5}' | sed 's/%//')
    if [ "$disk_usage" -lt 90 ]; then
        echo -e "${GREEN}✅ Disk Space: ${disk_usage}% used${NC}"
        healthy_services=$((healthy_services + 1))
    else
        echo -e "${RED}❌ Disk Space: ${disk_usage}% used (CRITICAL)${NC}"
        send_alert "Disk space is critically low: ${disk_usage}%" "CRITICAL"
    fi
    
    # Check memory usage
    total_services=$((total_services + 1))
    memory_usage=$(free | awk 'NR==2{printf "%.0f", $3*100/$2}')
    if [ "$memory_usage" -lt 90 ]; then
        echo -e "${GREEN}✅ Memory Usage: ${memory_usage}%${NC}"
        healthy_services=$((healthy_services + 1))
    else
        echo -e "${RED}❌ Memory Usage: ${memory_usage}% (HIGH)${NC}"
        send_alert "Memory usage is high: ${memory_usage}%" "WARNING"
    fi
    
    # Check CPU load
    total_services=$((total_services + 1))
    cpu_load=$(uptime | awk -F'load average:' '{print $2}' | awk '{print $1}' | sed 's/,//')
    if (( $(echo "$cpu_load < 5.0" | bc -l) )); then
        echo -e "${GREEN}✅ CPU Load: $cpu_load${NC}"
        healthy_services=$((healthy_services + 1))
    else
        echo -e "${RED}❌ CPU Load: $cpu_load (HIGH)${NC}"
        send_alert "CPU load is high: $cpu_load" "WARNING"
    fi
    
    # ===== NETWORK HEALTH =====
    echo -e "${YELLOW}🌐 NETWORK HEALTH${NC}"
    echo "=================="
    
    # Check network connectivity
    total_services=$((total_services + 1))
    if check_health "Internet Connectivity" "ping -c 1 8.8.8.8" ""; then
        healthy_services=$((healthy_services + 1))
    else
        send_alert "No internet connectivity" "WARNING"
    fi
    
    # Check DNS resolution
    total_services=$((total_services + 1))
    if check_health "DNS Resolution" "nslookup google.com" ""; then
        healthy_services=$((healthy_services + 1))
    else
        send_alert "DNS resolution is not working" "WARNING"
    fi
    
    # ===== APPLICATION HEALTH =====
    echo -e "${YELLOW}📱 APPLICATION HEALTH${NC}"
    echo "====================="
    
    # Check CLI tool
    total_services=$((total_services + 1))
    if check_health "CLI Tool" "./bin/adrenochain-cli --help" ""; then
        healthy_services=$((healthy_services + 1))
    else
        send_alert "CLI tool is not working" "WARNING"
    fi
    
    # Check JavaScript SDK
    total_services=$((total_services + 1))
    if [ -f "sdk/javascript/package.json" ]; then
        echo -e "${GREEN}✅ JavaScript SDK: AVAILABLE${NC}"
        healthy_services=$((healthy_services + 1))
    else
        echo -e "${RED}❌ JavaScript SDK: NOT FOUND${NC}"
        send_alert "JavaScript SDK is missing" "WARNING"
    fi
    
    # Check Python SDK
    total_services=$((total_services + 1))
    if [ -f "sdk/python/setup.py" ]; then
        echo -e "${GREEN}✅ Python SDK: AVAILABLE${NC}"
        healthy_services=$((healthy_services + 1))
    else
        echo -e "${RED}❌ Python SDK: NOT FOUND${NC}"
        send_alert "Python SDK is missing" "WARNING"
    fi
    
    # Check Web Wallet
    total_services=$((total_services + 1))
    if [ -f "apps/web-wallet/package.json" ]; then
        echo -e "${GREEN}✅ Web Wallet: AVAILABLE${NC}"
        healthy_services=$((healthy_services + 1))
    else
        echo -e "${RED}❌ Web Wallet: NOT FOUND${NC}"
        send_alert "Web Wallet is missing" "WARNING"
    fi
    
    # ===== FINAL SUMMARY =====
    echo -e "${CYAN}"
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║                    📊 HEALTH SUMMARY 📊                     ║"
    echo "╠══════════════════════════════════════════════════════════════╣"
    echo "║  Total Services: $total_services"
    echo "║  ✅ Healthy: $healthy_services"
    echo "║  ❌ Unhealthy: $((total_services - healthy_services))"
    echo "║  📈 Health Score: $(( (healthy_services * 100) / total_services ))%"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
    
    # Log final results
    log_health "Health check completed: $healthy_services/$total_services services healthy"
    
    # Determine overall health status
    if [ $healthy_services -eq $total_services ]; then
        echo -e "${GREEN}🎉 ALL SYSTEMS HEALTHY! 🎉${NC}"
        return 0
    elif [ $healthy_services -gt $((total_services * 80 / 100)) ]; then
        echo -e "${YELLOW}⚠️  MOSTLY HEALTHY - Some issues detected${NC}"
        return 1
    else
        echo -e "${RED}🚨 CRITICAL ISSUES DETECTED! 🚨${NC}"
        return 2
    fi
}

# Continuous monitoring function
continuous_monitoring() {
    local interval="${1:-60}"
    
    echo -e "${BLUE}🔄 Starting continuous monitoring (interval: ${interval}s)${NC}"
    echo "Press Ctrl+C to stop"
    
    while true; do
        clear
        main_health_check
        echo -e "${BLUE}⏰ Next check in ${interval} seconds...${NC}"
        sleep "$interval"
    done
}

# Alert monitoring function
monitor_alerts() {
    echo -e "${BLUE}🚨 Monitoring alerts...${NC}"
    
    if [ -f "$ALERT_LOG" ]; then
        echo -e "${YELLOW}Recent alerts:${NC}"
        tail -10 "$ALERT_LOG"
    else
        echo -e "${GREEN}No alerts found${NC}"
    fi
}

# Main function
case "${1:-check}" in
    "check")
        main_health_check
        ;;
    "monitor")
        continuous_monitoring "${2:-60}"
        ;;
    "alerts")
        monitor_alerts
        ;;
    "help"|"-h"|"--help")
        echo "Usage: $0 [check|monitor|alerts]"
        echo ""
        echo "Commands:"
        echo "  check    - Run a single health check (default)"
        echo "  monitor  - Run continuous monitoring with specified interval"
        echo "  alerts   - Show recent alerts"
        echo "  help     - Show this help message"
        echo ""
        echo "Examples:"
        echo "  $0 check              # Run single health check"
        echo "  $0 monitor 30         # Monitor every 30 seconds"
        echo "  $0 alerts             # Show recent alerts"
        ;;
    *)
        echo "Unknown command: $1"
        echo "Use '$0 help' for usage information"
        exit 1
        ;;
esac




#!/bin/bash

# 🎛️ Adrenochain Management Script
# This script provides comprehensive management capabilities for the Adrenochain ecosystem

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
LOG_DIR="$PROJECT_ROOT/logs"
DATA_DIR="$PROJECT_ROOT/data"
CONFIG_DIR="$PROJECT_ROOT/config"

# Create directories
mkdir -p "$LOG_DIR" "$DATA_DIR"

# Log function
log() {
    local message="$1"
    local level="${2:-INFO}"
    echo "$(date '+%Y-%m-%d %H:%M:%S') [$level] $message" | tee -a "$LOG_DIR/management.log"
}

# Error handling
handle_error() {
    local exit_code=$?
    log "Command failed with exit code $exit_code" "ERROR"
    exit $exit_code
}

trap handle_error ERR

# Display banner
show_banner() {
    echo -e "${CYAN}"
    cat << "EOF"
╔══════════════════════════════════════════════════════════════╗
║                🎛️  Adrenochain Management 🎛️                ║
║                                                              ║
║  Comprehensive management for the entire ecosystem           ║
║                                                              ║
╚══════════════════════════════════════════════════════════════╝
EOF
    echo -e "${NC}"
}

# Start services
start_services() {
    local mode="${1:-development}"
    
    log "Starting Adrenochain services in $mode mode"
    
    case "$mode" in
        "development")
            echo -e "${BLUE}🚀 Starting development environment...${NC}"
            
            # Start main application
            if ! pgrep -f "adrenochain.*--config" > /dev/null; then
                echo -e "${BLUE}Starting main application...${NC}"
                nohup ./bin/adrenochain --config config/config.yaml --mining > "$LOG_DIR/adrenochain.log" 2>&1 &
                sleep 5
                
                if pgrep -f "adrenochain.*--config" > /dev/null; then
                    echo -e "${GREEN}✅ Main application started${NC}"
                    log "Main application started successfully"
                else
                    echo -e "${RED}❌ Failed to start main application${NC}"
                    log "Failed to start main application" "ERROR"
                    return 1
                fi
            else
                echo -e "${YELLOW}⚠️  Main application already running${NC}"
            fi
            ;;
            
        "production")
            echo -e "${BLUE}🏭 Starting production environment...${NC}"
            
            # Start with Docker Compose
            if command -v docker-compose > /dev/null; then
                echo -e "${BLUE}Starting Docker services...${NC}"
                docker-compose up -d
                sleep 10
                
                if docker-compose ps | grep -q "Up"; then
                    echo -e "${GREEN}✅ Docker services started${NC}"
                    log "Docker services started successfully"
                else
                    echo -e "${RED}❌ Failed to start Docker services${NC}"
                    log "Failed to start Docker services" "ERROR"
                    return 1
                fi
            else
                echo -e "${RED}❌ Docker Compose not available${NC}"
                log "Docker Compose not available" "ERROR"
                return 1
            fi
            ;;
            
        "docker")
            echo -e "${BLUE}🐳 Starting Docker environment...${NC}"
            
            if command -v docker > /dev/null; then
                echo -e "${BLUE}Building Docker image...${NC}"
                docker build -t adrenochain:latest .
                
                echo -e "${BLUE}Starting Docker container...${NC}"
                docker run -d \
                    --name adrenochain-node \
                    -p 8080:8080 \
                    -p 8081:8081 \
                    -p 9090:9090 \
                    -p 30303:30303 \
                    -v "$DATA_DIR:/app/data" \
                    -v "$LOG_DIR:/app/logs" \
                    adrenochain:latest
                
                sleep 5
                
                if docker ps | grep -q "adrenochain-node"; then
                    echo -e "${GREEN}✅ Docker container started${NC}"
                    log "Docker container started successfully"
                else
                    echo -e "${RED}❌ Failed to start Docker container${NC}"
                    log "Failed to start Docker container" "ERROR"
                    return 1
                fi
            else
                echo -e "${RED}❌ Docker not available${NC}"
                log "Docker not available" "ERROR"
                return 1
            fi
            ;;
            
        *)
            echo -e "${RED}❌ Unknown mode: $mode${NC}"
            echo "Available modes: development, production, docker"
            return 1
            ;;
    esac
    
    echo -e "${GREEN}🎉 Services started successfully!${NC}"
    log "Services started successfully in $mode mode"
}

# Stop services
stop_services() {
    local mode="${1:-all}"
    
    log "Stopping Adrenochain services"
    
    case "$mode" in
        "all"|"development")
            echo -e "${BLUE}🛑 Stopping all services...${NC}"
            
            # Stop main application
            if pgrep -f "adrenochain.*--config" > /dev/null; then
                echo -e "${BLUE}Stopping main application...${NC}"
                pkill -f "adrenochain.*--config"
                sleep 2
                echo -e "${GREEN}✅ Main application stopped${NC}"
                log "Main application stopped"
            fi
            ;;
            
        "production"|"docker")
            echo -e "${BLUE}🐳 Stopping Docker services...${NC}"
            
            if command -v docker-compose > /dev/null; then
                docker-compose down
                echo -e "${GREEN}✅ Docker services stopped${NC}"
                log "Docker services stopped"
            fi
            
            if docker ps | grep -q "adrenochain-node"; then
                docker stop adrenochain-node
                docker rm adrenochain-node
                echo -e "${GREEN}✅ Docker container stopped${NC}"
                log "Docker container stopped"
            fi
            ;;
    esac
    
    echo -e "${GREEN}🎉 Services stopped successfully!${NC}"
    log "Services stopped successfully"
}

# Restart services
restart_services() {
    local mode="${1:-development}"
    
    log "Restarting Adrenochain services in $mode mode"
    
    echo -e "${BLUE}🔄 Restarting services...${NC}"
    stop_services "$mode"
    sleep 3
    start_services "$mode"
    
    echo -e "${GREEN}🎉 Services restarted successfully!${NC}"
    log "Services restarted successfully in $mode mode"
}

# Show status
show_status() {
    echo -e "${BLUE}📊 Adrenochain Status${NC}"
    echo "====================="
    
    # Check main process
    if pgrep -f "adrenochain.*--config" > /dev/null; then
        local pid=$(pgrep -f "adrenochain.*--config")
        echo -e "${GREEN}✅ Main Process: Running (PID: $pid)${NC}"
    else
        echo -e "${RED}❌ Main Process: Not running${NC}"
    fi
    
    # Check API
    if curl -s http://localhost:8080/health > /dev/null 2>&1; then
        echo -e "${GREEN}✅ API Server: Running${NC}"
    else
        echo -e "${RED}❌ API Server: Not responding${NC}"
    fi
    
    # Check Docker services
    if command -v docker-compose > /dev/null; then
        echo -e "${BLUE}🐳 Docker Services:${NC}"
        docker-compose ps
    fi
    
    # Check system resources
    echo -e "${BLUE}💻 System Resources:${NC}"
    echo "CPU Load: $(uptime | awk -F'load average:' '{print $2}')"
    echo "Memory: $(free -h | awk 'NR==2{printf "%.1f%%", $3*100/$2}')"
    echo "Disk: $(df -h / | awk 'NR==2{print $5}')"
}

# Show logs
show_logs() {
    local service="${1:-all}"
    local lines="${2:-50}"
    
    echo -e "${BLUE}📋 Showing logs for: $service${NC}"
    echo "================================"
    
    case "$service" in
        "all"|"main")
            if [ -f "$LOG_DIR/adrenochain.log" ]; then
                echo -e "${BLUE}Main Application Logs:${NC}"
                tail -n "$lines" "$LOG_DIR/adrenochain.log"
            fi
            ;;
        "docker")
            if command -v docker-compose > /dev/null; then
                echo -e "${BLUE}Docker Logs:${NC}"
                docker-compose logs --tail="$lines"
            fi
            ;;
        "management")
            if [ -f "$LOG_DIR/management.log" ]; then
                echo -e "${BLUE}Management Logs:${NC}"
                tail -n "$lines" "$LOG_DIR/management.log"
            fi
            ;;
        *)
            echo -e "${RED}❌ Unknown service: $service${NC}"
            echo "Available services: all, main, docker, management"
            ;;
    esac
}

# Backup data
backup_data() {
    local backup_dir="${1:-$PROJECT_ROOT/backups}"
    local timestamp=$(date +"%Y%m%d_%H%M%S")
    local backup_file="$backup_dir/adrenochain_backup_$timestamp.tar.gz"
    
    log "Creating backup: $backup_file"
    
    mkdir -p "$backup_dir"
    
    echo -e "${BLUE}💾 Creating backup...${NC}"
    
    tar -czf "$backup_file" \
        --exclude='*.log' \
        --exclude='node_modules' \
        --exclude='.git' \
        --exclude='backups' \
        -C "$PROJECT_ROOT" \
        .
    
    if [ -f "$backup_file" ]; then
        local size=$(du -h "$backup_file" | cut -f1)
        echo -e "${GREEN}✅ Backup created: $backup_file ($size)${NC}"
        log "Backup created successfully: $backup_file ($size)"
    else
        echo -e "${RED}❌ Failed to create backup${NC}"
        log "Failed to create backup" "ERROR"
        return 1
    fi
}

# Restore data
restore_data() {
    local backup_file="$1"
    
    if [ -z "$backup_file" ]; then
        echo -e "${RED}❌ Please specify backup file${NC}"
        echo "Usage: $0 restore <backup_file>"
        return 1
    fi
    
    if [ ! -f "$backup_file" ]; then
        echo -e "${RED}❌ Backup file not found: $backup_file${NC}"
        return 1
    fi
    
    log "Restoring from backup: $backup_file"
    
    echo -e "${YELLOW}⚠️  This will overwrite current data. Continue? (y/N)${NC}"
    read -r response
    if [[ ! "$response" =~ ^[Yy]$ ]]; then
        echo "Restore cancelled"
        return 0
    fi
    
    echo -e "${BLUE}🔄 Restoring from backup...${NC}"
    
    # Stop services first
    stop_services
    
    # Extract backup
    tar -xzf "$backup_file" -C "$PROJECT_ROOT"
    
    echo -e "${GREEN}✅ Data restored successfully${NC}"
    log "Data restored successfully from $backup_file"
}

# Update system
update_system() {
    log "Updating Adrenochain system"
    
    echo -e "${BLUE}🔄 Updating system...${NC}"
    
    # Stop services
    stop_services
    
    # Update dependencies
    echo -e "${BLUE}Updating Go dependencies...${NC}"
    go mod tidy
    go mod download
    
    # Rebuild
    echo -e "${BLUE}Rebuilding application...${NC}"
    go build -o bin/adrenochain ./cmd/gochain
    go build -o bin/adrenochain-cli ./tools/cli
    
    # Rebuild Docker image
    if command -v docker > /dev/null; then
        echo -e "${BLUE}Rebuilding Docker image...${NC}"
        docker build -t adrenochain:latest .
    fi
    
    echo -e "${GREEN}✅ System updated successfully${NC}"
    log "System updated successfully"
}

# Clean system
clean_system() {
    log "Cleaning Adrenochain system"
    
    echo -e "${BLUE}🧹 Cleaning system...${NC}"
    
    # Stop services
    stop_services
    
    # Clean Go cache
    echo -e "${BLUE}Cleaning Go cache...${NC}"
    go clean -cache -testcache -modcache
    
    # Clean Docker
    if command -v docker > /dev/null; then
        echo -e "${BLUE}Cleaning Docker...${NC}"
        docker system prune -f
        docker volume prune -f
    fi
    
    # Clean logs
    echo -e "${BLUE}Cleaning logs...${NC}"
    find "$LOG_DIR" -name "*.log" -mtime +7 -delete
    
    # Clean temporary files
    echo -e "${BLUE}Cleaning temporary files...${NC}"
    find "$PROJECT_ROOT" -name "*.tmp" -delete
    find "$PROJECT_ROOT" -name ".DS_Store" -delete
    
    echo -e "${GREEN}✅ System cleaned successfully${NC}"
    log "System cleaned successfully"
}

# Show help
show_help() {
    echo "Usage: $0 <command> [options]"
    echo ""
    echo "Commands:"
    echo "  start [mode]     - Start services (development|production|docker)"
    echo "  stop [mode]      - Stop services (all|development|production|docker)"
    echo "  restart [mode]   - Restart services"
    echo "  status           - Show system status"
    echo "  logs [service]   - Show logs (all|main|docker|management)"
    echo "  backup [dir]     - Create backup"
    echo "  restore <file>   - Restore from backup"
    echo "  update           - Update system"
    echo "  clean            - Clean system"
    echo "  help             - Show this help"
    echo ""
    echo "Examples:"
    echo "  $0 start development    # Start in development mode"
    echo "  $0 start production     # Start in production mode"
    echo "  $0 stop all            # Stop all services"
    echo "  $0 status              # Show status"
    echo "  $0 logs main           # Show main application logs"
    echo "  $0 backup              # Create backup"
    echo "  $0 restore backup.tar.gz # Restore from backup"
}

# Main function
main() {
    show_banner
    
    case "${1:-help}" in
        "start")
            start_services "${2:-development}"
            ;;
        "stop")
            stop_services "${2:-all}"
            ;;
        "restart")
            restart_services "${2:-development}"
            ;;
        "status")
            show_status
            ;;
        "logs")
            show_logs "${2:-all}" "${3:-50}"
            ;;
        "backup")
            backup_data "${2:-$PROJECT_ROOT/backups}"
            ;;
        "restore")
            restore_data "$2"
            ;;
        "update")
            update_system
            ;;
        "clean")
            clean_system
            ;;
        "help"|"-h"|"--help")
            show_help
            ;;
        *)
            echo -e "${RED}❌ Unknown command: $1${NC}"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

# Run main function
main "$@"




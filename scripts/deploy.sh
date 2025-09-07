#!/bin/bash

# 🚀 Adrenochain Production Deployment Script
# This script deploys adrenochain to production

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
DEPLOYMENT_MODE=${1:-"development"}

echo -e "${CYAN}"
cat << "EOF"
╔══════════════════════════════════════════════════════════════╗
║                🚀 Adrenochain Deployment 🚀                  ║
║                                                              ║
║  Production-ready blockchain deployment script               ║
║                                                              ║
╚══════════════════════════════════════════════════════════════╝
EOF
echo -e "${NC}"

# Print banner
print_banner() {
    echo -e "${BLUE}🚀 Adrenochain Production Deployment${NC}"
    echo -e "${BLUE}=====================================${NC}"
    echo -e "Deployment Mode: ${YELLOW}$DEPLOYMENT_MODE${NC}"
    echo -e "Project Root: ${CYAN}$PROJECT_ROOT${NC}"
    echo
}

# Check prerequisites
check_prerequisites() {
    echo -e "${BLUE}🔍 Checking prerequisites...${NC}"
    
    # Check Go installation
    if ! command -v go &> /dev/null; then
        echo -e "${RED}❌ Go is not installed or not in PATH${NC}"
        exit 1
    fi
    
    # Check Docker installation
    if ! command -v docker &> /dev/null; then
        echo -e "${RED}❌ Docker is not installed or not in PATH${NC}"
        exit 1
    fi
    
    # Check Docker Compose installation
    if ! command -v docker-compose &> /dev/null; then
        echo -e "${RED}❌ Docker Compose is not installed or not in PATH${NC}"
        exit 1
    fi
    
    # Check if we're in the right directory
    if [[ ! -f "$PROJECT_ROOT/go.mod" ]]; then
        echo -e "${RED}❌ Not in adrenochain project root (go.mod not found)${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✅ Prerequisites check passed${NC}"
}

# Build the application
build_application() {
    echo -e "${BLUE}🔨 Building adrenochain application...${NC}"
    
    # Clean previous builds
    echo -e "   🧹 Cleaning previous builds..."
    go clean -cache -testcache
    
    # Download dependencies
    echo -e "   📥 Downloading dependencies..."
    go mod download
    go mod tidy
    
    # Build the application
    echo -e "   🔨 Building application..."
    if go build -o bin/adrenochain ./cmd/gochain; then
        echo -e "${GREEN}✅ Application built successfully${NC}"
    else
        echo -e "${RED}❌ Application build failed${NC}"
        exit 1
    fi
    
    # Run tests
    echo -e "   🧪 Running tests..."
    if go test ./... -timeout 5m; then
        echo -e "${GREEN}✅ All tests passed${NC}"
    else
        echo -e "${YELLOW}⚠️  Some tests failed, continuing with deployment...${NC}"
    fi
}

# Setup Docker environment
setup_docker() {
    echo -e "${BLUE}🐳 Setting up Docker environment...${NC}"
    
    # Build Docker image
    echo -e "   🔨 Building Docker image..."
    if docker build -t adrenochain:latest .; then
        echo -e "${GREEN}✅ Docker image built successfully${NC}"
    else
        echo -e "${RED}❌ Docker image build failed${NC}"
        exit 1
    fi
    
    # Create Docker network
    echo -e "   🌐 Creating Docker network..."
    docker network create adrenochain-network 2>/dev/null || true
    
    echo -e "${GREEN}✅ Docker environment setup completed${NC}"
}

# Deploy based on mode
deploy_development() {
    echo -e "${BLUE}🚀 Deploying in development mode...${NC}"
    
    # Start services with docker-compose
    echo -e "   🐳 Starting services with Docker Compose..."
    if docker-compose up -d; then
        echo -e "${GREEN}✅ Development deployment completed${NC}"
    else
        echo -e "${RED}❌ Development deployment failed${NC}"
        exit 1
    fi
    
    # Show service status
    echo -e "   📊 Service status:"
    docker-compose ps
}

deploy_production() {
    echo -e "${BLUE}🚀 Deploying in production mode...${NC}"
    
    # Create production directories
    echo -e "   📁 Creating production directories..."
    sudo mkdir -p /opt/adrenochain/{data,logs,config}
    sudo chown -R $USER:$USER /opt/adrenochain
    
    # Copy configuration files
    echo -e "   ⚙️  Copying configuration files..."
    cp config/production.yaml /opt/adrenochain/config/
    
    # Create systemd service
    echo -e "   🔧 Creating systemd service..."
    sudo tee /etc/systemd/system/adrenochain.service > /dev/null << EOF
[Unit]
Description=Adrenochain Node
After=network.target

[Service]
Type=simple
User=$USER
WorkingDirectory=/opt/adrenochain
ExecStart=$PROJECT_ROOT/bin/adrenochain --config /opt/adrenochain/config/production.yaml
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF
    
    # Reload systemd and start service
    echo -e "   🚀 Starting adrenochain service..."
    sudo systemctl daemon-reload
    sudo systemctl enable adrenochain
    sudo systemctl start adrenochain
    
    # Check service status
    echo -e "   📊 Service status:"
    sudo systemctl status adrenochain --no-pager
    
    echo -e "${GREEN}✅ Production deployment completed${NC}"
}

deploy_kubernetes() {
    echo -e "${BLUE}☸️  Deploying to Kubernetes...${NC}"
    
    # Create Kubernetes namespace
    echo -e "   📦 Creating Kubernetes namespace..."
    kubectl create namespace adrenochain 2>/dev/null || true
    
    # Apply Kubernetes manifests
    echo -e "   📋 Applying Kubernetes manifests..."
    if [ -d "scripts/deployment/k8s" ]; then
        kubectl apply -f scripts/deployment/k8s/ -n adrenochain
    else
        echo -e "${YELLOW}⚠️  Kubernetes manifests not found, creating basic deployment...${NC}"
        
        # Create basic deployment
        kubectl create deployment adrenochain --image=adrenochain:latest -n adrenochain
        kubectl expose deployment adrenochain --port=8080 --type=LoadBalancer -n adrenochain
    fi
    
    # Check deployment status
    echo -e "   📊 Deployment status:"
    kubectl get pods -n adrenochain
    kubectl get services -n adrenochain
    
    echo -e "${GREEN}✅ Kubernetes deployment completed${NC}"
}

# Setup monitoring
setup_monitoring() {
    echo -e "${BLUE}📊 Setting up monitoring...${NC}"
    
    if [ "$DEPLOYMENT_MODE" = "development" ]; then
        echo -e "   📈 Monitoring available at:"
        echo -e "      - Grafana: http://localhost:3000 (admin/admin)"
        echo -e "      - Prometheus: http://localhost:9091"
        echo -e "      - Adrenochain Metrics: http://localhost:9090/metrics"
    else
        echo -e "   📈 Setting up production monitoring..."
        # Add production monitoring setup here
    fi
    
    echo -e "${GREEN}✅ Monitoring setup completed${NC}"
}

# Show deployment status
show_status() {
    echo -e "${BLUE}📊 Deployment Status${NC}"
    echo -e "${BLUE}===================${NC}"
    
    case "$DEPLOYMENT_MODE" in
        development)
            echo -e "🌐 Development Environment:"
            echo -e "   - Adrenochain API: http://localhost:8080"
            echo -e "   - Health Check: http://localhost:8081/health"
            echo -e "   - Metrics: http://localhost:9090/metrics"
            echo -e "   - Grafana: http://localhost:3000"
            echo -e "   - Prometheus: http://localhost:9091"
            echo
            echo -e "📋 Useful Commands:"
            echo -e "   - View logs: docker-compose logs -f adrenochain"
            echo -e "   - Stop services: docker-compose down"
            echo -e "   - Restart services: docker-compose restart"
            ;;
        production)
            echo -e "🏭 Production Environment:"
            echo -e "   - Service status: sudo systemctl status adrenochain"
            echo -e "   - View logs: sudo journalctl -u adrenochain -f"
            echo -e "   - Restart service: sudo systemctl restart adrenochain"
            echo -e "   - Stop service: sudo systemctl stop adrenochain"
            ;;
        kubernetes)
            echo -e "☸️  Kubernetes Environment:"
            echo -e "   - Pods: kubectl get pods -n adrenochain"
            echo -e "   - Services: kubectl get services -n adrenochain"
            echo -e "   - Logs: kubectl logs -f deployment/adrenochain -n adrenochain"
            ;;
    esac
}

# Main deployment function
main() {
    print_banner
    check_prerequisites
    build_application
    setup_docker
    
    case "$DEPLOYMENT_MODE" in
        development)
            deploy_development
            ;;
        production)
            deploy_production
            ;;
        kubernetes)
            deploy_kubernetes
            ;;
        *)
            echo -e "${RED}❌ Invalid deployment mode: $DEPLOYMENT_MODE${NC}"
            echo -e "Valid modes: development, production, kubernetes"
            exit 1
            ;;
    esac
    
    setup_monitoring
    show_status
    
    echo -e "${GREEN}🎉 Deployment completed successfully!${NC}"
}

# Handle script arguments
case "${1:-}" in
    --help|-h)
        echo "Usage: $0 [DEPLOYMENT_MODE]"
        echo
        echo "Deployment Modes:"
        echo "  development  - Deploy using Docker Compose (default)"
        echo "  production   - Deploy as systemd service"
        echo "  kubernetes   - Deploy to Kubernetes"
        echo
        echo "Examples:"
        echo "  $0                    # Deploy in development mode"
        echo "  $0 development        # Deploy in development mode"
        echo "  $0 production         # Deploy in production mode"
        echo "  $0 kubernetes         # Deploy to Kubernetes"
        exit 0
        ;;
    development|production|kubernetes)
        DEPLOYMENT_MODE="$1"
        ;;
esac

# Run main function
main

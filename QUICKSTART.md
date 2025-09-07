# 🚀 Adrenochain Quick Start Guide

Welcome to Adrenochain! This guide will get you up and running with the blockchain in minutes.

## 📋 Prerequisites

- **Go 1.23+** - [Download here](https://golang.org/dl/)
- **Docker & Docker Compose** - [Install Docker](https://docs.docker.com/get-docker/)
- **Git** - [Install Git](https://git-scm.com/downloads)

## 🏃‍♂️ Quick Start (5 minutes)

### 1. Clone and Build
```bash
git clone https://github.com/palaseus/adrenochain.git
cd adrenochain
go mod tidy
go build -o bin/adrenochain ./cmd/gochain
```

### 2. Start Development Environment
```bash
# Start with Docker Compose (includes monitoring)
./scripts/deploy.sh development

# Or start manually
./bin/adrenochain --config config/config.yaml
```

### 3. Verify Installation
```bash
# Check health
curl http://localhost:8081/health

# Check blockchain info
curl http://localhost:8080/api/v1/blockchain/info

# View metrics
curl http://localhost:9090/metrics
```

## 🎯 What's Available

### 🌐 Web Interfaces
- **API**: http://localhost:8080
- **Health Check**: http://localhost:8081/health
- **Metrics**: http://localhost:9090/metrics
- **Grafana Dashboard**: http://localhost:3000 (admin/admin)
- **Prometheus**: http://localhost:9091

### 🛠️ Command Line Tools
```bash
# Create a wallet
./bin/adrenochain wallet --wallet-file my_wallet.dat --passphrase my_password

# Check balance
./bin/adrenochain balance 0x1234567890123456789012345678901234567890

# Send transaction
./bin/adrenochain tx send --from 0x123... --to 0x456... --amount 1000

# Get blockchain info
./bin/adrenochain info

# Explore blocks
./bin/adrenochain explore block 0xabc123...
```

### 📊 Monitoring
- **Real-time metrics** via Prometheus
- **Beautiful dashboards** via Grafana
- **Health monitoring** with automatic alerts
- **Performance tracking** for blocks and transactions

## 🏗️ Development Setup

### 1. Bootstrap Network
```bash
# Set up bootstrap network
./scripts/setup_bootstrap_network.sh

# Generate genesis block
cd bootstrap_network
./generate_genesis.sh

# Start network
./start_network.sh
```

### 2. Validator Setup
```bash
# Set up validators
./scripts/validator_setup.sh

# Generate keys for all validators
cd validators
./manage_validators.sh generate-keys

# Start all validators
./manage_validators.sh start

# Monitor network
./monitor_network.sh
```

### 3. Production Deployment
```bash
# Deploy to production
./scripts/deploy.sh production

# Or deploy to Kubernetes
./scripts/deploy.sh kubernetes
```

## 🔧 Configuration

### Main Configuration
Edit `config/config.yaml`:
```yaml
network:
  listen_port: 30303
  bootstrap_peers: []

blockchain:
  genesis_block_reward: 1000000000
  block_time: 10s

mining:
  enabled: true
  mining_threads: 2

api:
  enabled: true
  listen_addr: "127.0.0.1:8080"
```

### Production Configuration
Edit `config/production.yaml` for production settings.

## 📚 Next Steps

### For Developers
1. **Explore the codebase** - Start with `pkg/` directory
2. **Read the documentation** - Check `docs/` folder
3. **Run tests** - `go test ./...`
4. **Build your first dApp** - Use the JavaScript SDK

### For Validators
1. **Set up validator node** - Use validator setup scripts
2. **Join the network** - Connect to bootstrap nodes
3. **Monitor performance** - Use Grafana dashboards
4. **Participate in governance** - Vote on proposals

### For Users
1. **Create a wallet** - Use CLI or web interface
2. **Get test tokens** - Use faucet (when available)
3. **Explore DeFi** - Interact with protocols
4. **Trade tokens** - Use built-in exchange

## 🆘 Troubleshooting

### Common Issues

**Build fails:**
```bash
go clean -cache
go mod tidy
go build ./...
```

**Docker issues:**
```bash
docker system prune -a
docker-compose down -v
docker-compose up --build
```

**Port conflicts:**
- Change ports in `config/config.yaml`
- Update `docker-compose.yml` if using Docker

**Permission issues:**
```bash
sudo chown -R $USER:$USER .
chmod +x scripts/*.sh
```

### Getting Help

- **Documentation**: Check `docs/` folder
- **Issues**: [GitHub Issues](https://github.com/palaseus/adrenochain/issues)
- **Discord**: [Join our community](https://discord.gg/adrenochain)
- **Email**: support@adrenochain.com

## 🎉 You're Ready!

You now have a fully functional Adrenochain node running! 

- ✅ **Blockchain**: Creating and validating blocks
- ✅ **API**: RESTful interface for interactions
- ✅ **Monitoring**: Real-time metrics and health checks
- ✅ **Tools**: CLI and SDK for development

**Happy building! 🚀**

---

*Need more help? Check out the [full documentation](docs/) or [join our community](https://discord.gg/adrenochain).*

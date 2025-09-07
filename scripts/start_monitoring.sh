#!/bin/bash

# Script to start monitoring service manually
echo "🔧 Starting Adrenochain Monitoring Service"

# Check if container is running
if ! sudo docker ps | grep -q adrenochain-node; then
    echo "❌ Adrenochain container is not running"
    exit 1
fi

# Check current configuration
echo "📋 Current monitoring configuration:"
sudo docker exec adrenochain-node cat /app/config/production.yaml | grep -A 20 monitoring

# Check if monitoring service is enabled
echo ""
echo "🔍 Checking monitoring service status:"
sudo docker exec adrenochain-node netstat -tlnp | grep -E ':(9090|8081)'

# Check logs for monitoring service
echo ""
echo "📝 Recent logs:"
sudo docker logs adrenochain-node --tail 10

# Test API endpoints
echo ""
echo "🧪 Testing API endpoints:"
echo "  - Health: $(curl -s http://localhost:8080/health | jq -r '.status' 2>/dev/null || echo 'Failed')"
echo "  - Chain Info: $(curl -s http://localhost:8080/api/v1/chain/info | jq -r '.height' 2>/dev/null || echo 'Failed')"

# Test Prometheus targets
echo ""
echo "📊 Prometheus targets:"
curl -s http://localhost:9091/api/v1/targets | jq '.data.activeTargets[] | select(.job=="adrenochain") | {job: .job, health: .health, lastError: .lastError}' 2>/dev/null || echo "Failed to get targets"

echo ""
echo "✅ Monitoring diagnostic complete"

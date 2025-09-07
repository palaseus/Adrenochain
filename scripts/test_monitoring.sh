#!/bin/bash

# Test script for monitoring integration
echo "🔍 Testing Adrenochain Monitoring Integration"
echo "=============================================="

# Test API endpoints
echo "📡 Testing API Endpoints:"
echo "  - Health: $(curl -s http://localhost:8080/health | jq -r '.status')"
echo "  - Chain Info: $(curl -s http://localhost:8080/api/v1/chain/info | jq -r '.height') blocks"
echo "  - Network Status: $(curl -s http://localhost:8080/api/v1/network/status | jq -r '.status')"

# Test Prometheus
echo ""
echo "📊 Testing Prometheus:"
echo "  - Prometheus Status: $(curl -s http://localhost:9091/api/v1/query?query=up | jq -r '.data.result[] | select(.metric.job=="prometheus") | .value[1]')"
echo "  - Adrenochain API: $(curl -s http://localhost:9091/api/v1/query?query=up | jq -r '.data.result[] | select(.metric.job=="adrenochain-api") | .value[1]')"

# Test Grafana
echo ""
echo "📈 Testing Grafana:"
echo "  - Grafana Status: $(curl -s http://localhost:3000/api/health | jq -r '.database')"

# Test Redis
echo ""
echo "🗄️ Testing Redis:"
echo "  - Redis Status: $(sudo docker exec adrenochain-redis redis-cli ping 2>/dev/null || echo "Not responding")"

echo ""
echo "✅ Monitoring Integration Test Complete!"

#!/bin/bash

echo "🔍 Testing Complete Monitoring Integration (with timeouts)"
echo "========================================================="

# Test monitoring service
echo "📊 Monitoring Service:"
echo "  - Health: $(timeout 5 curl -s http://localhost:9093/health | jq -r '.status' 2>/dev/null || echo 'Failed')"
echo "  - Metrics: $(timeout 5 curl -s http://localhost:9093/metrics | jq -r '.blockchain.height' 2>/dev/null || echo 'Failed')"
echo "  - Prometheus: $(timeout 5 curl -s http://localhost:9093/prometheus | head -1 | cut -d' ' -f1 2>/dev/null || echo 'Failed')"

# Test API endpoints
echo ""
echo "📡 API Endpoints:"
echo "  - Health: $(timeout 5 curl -s http://localhost:8080/health | jq -r '.status' 2>/dev/null || echo 'Failed')"
echo "  - Chain Info: $(timeout 5 curl -s http://localhost:8080/api/v1/chain/info | jq -r '.height' 2>/dev/null || echo 'Failed') blocks"

# Test Prometheus
echo ""
echo "📊 Prometheus:"
echo "  - Status: $(timeout 5 curl -s http://localhost:9091/api/v1/query?query=up | jq -r '.data.result[] | select(.metric.job=="prometheus") | .value[1]' 2>/dev/null || echo 'Failed')"
echo "  - Targets: $(timeout 5 curl -s http://localhost:9091/api/v1/targets | jq '.data.activeTargets | length' 2>/dev/null || echo 'Failed')"
echo "  - Adrenochain metrics: $(timeout 5 curl -s http://localhost:9091/api/v1/query?query=adrenochain_block_height | jq -r '.data.result[0].value[1]' 2>/dev/null || echo 'No data')"

# Test Grafana
echo ""
echo "📈 Grafana:"
echo "  - Status: $(timeout 5 curl -s http://localhost:3000/api/health | jq -r '.database' 2>/dev/null || echo 'Failed')"
echo "  - Dashboards: $(timeout 5 curl -s http://localhost:3000/api/search -u admin:admin | jq 'length' 2>/dev/null || echo 'Failed')"

# Test Redis
echo ""
echo "🗄️ Redis:"
echo "  - Status: $(timeout 5 sudo docker exec adrenochain-redis redis-cli ping 2>/dev/null || echo 'Not responding')"

# Test Docker services
echo ""
echo "🐳 Docker Services:"
timeout 5 sudo docker-compose ps --format "table {{.Name}}\t{{.Status}}" 2>/dev/null || echo "Failed to get Docker status"

echo ""
echo "✅ Complete Monitoring Test Finished!"
echo ""
echo "🌐 Access Points:"
echo "  - Grafana: http://localhost:3000 (admin/admin)"
echo "  - Prometheus: http://localhost:9091"
echo "  - Monitoring Service: http://localhost:9093/health"
echo "  - API: http://localhost:8080/health"

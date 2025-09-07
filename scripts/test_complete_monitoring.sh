#!/bin/bash

echo "🔍 Testing Complete Monitoring Integration"
echo "=========================================="

# Test all components
echo "📡 API Endpoints:"
echo "  - Health: $(curl -s http://localhost:8080/health | jq -r '.status' 2>/dev/null || echo 'Failed')"
echo "  - Chain Info: $(curl -s http://localhost:8080/api/v1/chain/info | jq -r '.height' 2>/dev/null || echo 'Failed') blocks"

echo ""
echo "📊 Prometheus:"
echo "  - Status: $(curl -s http://localhost:9091/api/v1/query?query=up | jq -r '.data.result[] | select(.metric.job=="prometheus") | .value[1]' 2>/dev/null || echo 'Failed')"
echo "  - Targets: $(curl -s http://localhost:9091/api/v1/targets | jq '.data.activeTargets | length' 2>/dev/null || echo 'Failed')"

echo ""
echo "📈 Grafana:"
echo "  - Status: $(curl -s http://localhost:3000/api/health | jq -r '.database' 2>/dev/null || echo 'Failed')"
echo "  - Dashboards: $(curl -s http://localhost:3000/api/search -u admin:admin | jq 'length' 2>/dev/null || echo 'Failed')"

echo ""
echo "🗄️ Redis:"
echo "  - Status: $(sudo docker exec adrenochain-redis redis-cli ping 2>/dev/null || echo 'Not responding')"

echo ""
echo "🐳 Docker Services:"
sudo docker-compose ps --format "table {{.Name}}\t{{.Status}}\t{{.Ports}}"

echo ""
echo "✅ Monitoring Integration Test Complete!"
echo ""
echo "🌐 Access Points:"
echo "  - Grafana: http://localhost:3000 (admin/admin)"
echo "  - Prometheus: http://localhost:9091"
echo "  - API: http://localhost:8080/health"

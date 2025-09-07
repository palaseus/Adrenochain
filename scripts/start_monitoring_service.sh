#!/bin/bash

echo "🔧 Starting Adrenochain Monitoring Service"

# Kill any existing monitoring processes
pkill -f monitoring_fix 2>/dev/null || true

# Start the monitoring service in background
echo "Starting monitoring service on port 9093..."
nohup go run scripts/monitoring_fix.go > /tmp/monitoring.log 2>&1 &
MONITORING_PID=$!

# Wait a moment for service to start
sleep 3

# Test the endpoints with timeouts
echo "Testing monitoring endpoints..."

echo "  - Health endpoint:"
timeout 5 curl -s http://localhost:9093/health | jq . 2>/dev/null || echo "    ❌ Health endpoint not responding"

echo "  - Metrics endpoint:"
timeout 5 curl -s http://localhost:9093/metrics | jq . 2>/dev/null || echo "    ❌ Metrics endpoint not responding"

echo "  - Prometheus endpoint:"
timeout 5 curl -s http://localhost:9093/prometheus | head -5 || echo "    ❌ Prometheus endpoint not responding"

# Check if service is running
if ps -p $MONITORING_PID > /dev/null; then
    echo "✅ Monitoring service started successfully (PID: $MONITORING_PID)"
    echo "📝 Logs: /tmp/monitoring.log"
else
    echo "❌ Monitoring service failed to start"
    echo "📝 Check logs: /tmp/monitoring.log"
fi

echo "🌐 Endpoints:"
echo "  - Health: http://localhost:9093/health"
echo "  - Metrics: http://localhost:9093/metrics"
echo "  - Prometheus: http://localhost:9093/prometheus"

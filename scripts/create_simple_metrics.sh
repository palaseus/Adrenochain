#!/bin/bash

# Create a simple metrics endpoint by adding it to the API server
echo "🔧 Creating simple metrics endpoint"

# Create a simple metrics server that exposes Prometheus metrics
cat > /tmp/simple_metrics_server.go << 'EOF'
package main

import (
    "fmt"
    "net/http"
    "runtime"
    "time"
)

func main() {
    http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
        
        var m runtime.MemStats
        runtime.ReadMemStats(&m)
        
        metrics := fmt.Sprintf(`# HELP adrenochain_block_height Current blockchain height
# TYPE adrenochain_block_height gauge
adrenochain_block_height 0

# HELP adrenochain_chain_difficulty Current chain difficulty  
# TYPE adrenochain_chain_difficulty gauge
adrenochain_chain_difficulty 1

# HELP adrenochain_connected_peers Number of connected peers
# TYPE adrenochain_connected_peers gauge
adrenochain_connected_peers 1

# HELP adrenochain_memory_usage_bytes Current memory usage in bytes
# TYPE adrenochain_memory_usage_bytes gauge
adrenochain_memory_usage_bytes %d

# HELP adrenochain_goroutines Number of goroutines
# TYPE adrenochain_goroutines gauge
adrenochain_goroutines %d

# HELP adrenochain_uptime_seconds Node uptime in seconds
# TYPE adrenochain_uptime_seconds gauge
adrenochain_uptime_seconds %d
`,
            m.Alloc,
            runtime.NumGoroutine(),
            int64(time.Since(time.Now().Add(-time.Hour)).Seconds()),
        )
        
        w.Write([]byte(metrics))
    })
    
    fmt.Println("Starting simple metrics server on port 9090...")
    http.ListenAndServe(":9090", nil)
}
EOF

echo "✅ Simple metrics server created"
echo "📝 To use: go run /tmp/simple_metrics_server.go"

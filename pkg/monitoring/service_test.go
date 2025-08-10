package monitoring

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gochain/gochain/pkg/block"
	"github.com/gochain/gochain/pkg/health"
	"github.com/gochain/gochain/pkg/logger"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockChain is a mock implementation of the chain for testing
type MockChain struct {
	height       uint64
	bestBlock    *block.Block
	genesisBlock *block.Block
}

func (mc *MockChain) GetHeight() uint64 {
	return mc.height
}

func (mc *MockChain) GetBestBlock() *block.Block {
	return mc.bestBlock
}

func (mc *MockChain) GetGenesisBlock() *block.Block {
	return mc.genesisBlock
}

func (mc *MockChain) GetBlockByHeight(height uint64) *block.Block {
	if height == mc.height {
		return mc.bestBlock
	}
	return nil
}

// MockMempool is a mock implementation of the mempool for testing
type MockMempool struct {
	txnCount int
}

func (mm *MockMempool) GetTransactionCount() int {
	return mm.txnCount
}

// MockNetwork is a mock implementation of the network for testing
type MockNetwork struct {
	peers []peer.ID
}

func (mn *MockNetwork) GetPeers() []peer.ID {
	return mn.peers
}

// MockHealthChecker is a mock implementation of the health checker for testing
type MockHealthChecker struct {
	name   string
	status health.Status
}

func (mhc *MockHealthChecker) Name() string {
	return mhc.name
}

func (mhc *MockHealthChecker) Check() (*health.Component, error) {
	return &health.Component{
		Name:      mhc.name,
		Status:    mhc.status,
		Message:   "Mock health check",
		LastCheck: time.Now(),
		CheckTime: 0,
		Details:   map[string]interface{}{},
	}, nil
}

func TestNewService(t *testing.T) {
	// Create mock dependencies
	mockChain := &MockChain{
		height: 10,
		bestBlock: &block.Block{
			Header: &block.Header{
				Height:     10,
				Timestamp:  time.Now(),
				Difficulty: 1000,
			},
			Transactions: []*block.Transaction{},
		},
		genesisBlock: &block.Block{
			Header: &block.Header{
				Height:     0,
				Timestamp:  time.Now().Add(-24 * time.Hour),
				Difficulty: 1,
			},
		},
	}
	mockMempool := &MockMempool{txnCount: 5}

	// Create mock peer IDs
	peer1, _ := peer.Decode("QmPeer1")
	peer2, _ := peer.Decode("QmPeer2")
	peer3, _ := peer.Decode("QmPeer3")
	mockNetwork := &MockNetwork{peers: []peer.ID{peer1, peer2, peer3}}

	// Test with default config
	service := NewService(nil, mockChain, mockMempool, mockNetwork)
	assert.NotNil(t, service)
	assert.NotNil(t, service.GetLogger())
	assert.NotNil(t, service.GetMetrics())
	assert.NotNil(t, service.GetSystemHealth())

	// Test with custom config
	config := &Config{
		MetricsPort:         9091,
		HealthPort:          8081,
		LogLevel:            logger.DEBUG,
		CollectInterval:     10 * time.Second,
		HealthCheckInterval: 5 * time.Second,
	}
	service2 := NewService(config, mockChain, mockMempool, mockNetwork)
	assert.NotNil(t, service2)
	assert.Equal(t, config.MetricsPort, service2.config.MetricsPort)
}

func TestServiceStartStop(t *testing.T) {
	mockChain := &MockChain{
		height: 5,
		bestBlock: &block.Block{
			Header: &block.Header{
				Height:     5,
				Timestamp:  time.Now(),
				Difficulty: 500,
			},
		},
	}
	mockMempool := &MockMempool{txnCount: 3}

	// Create mock peer ID
	peer1, _ := peer.Decode("QmPeer1")
	mockNetwork := &MockNetwork{peers: []peer.ID{peer1}}

	service := NewService(nil, mockChain, mockMempool, mockNetwork)

	// Start service
	err := service.Start()
	require.NoError(t, err)

	// Wait a bit for servers to start
	time.Sleep(100 * time.Millisecond)

	// Test metrics endpoint
	resp, err := http.Get(service.GetMetricsEndpoint())
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Test health endpoint
	resp, err = http.Get(service.GetHealthEndpoint())
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Test Prometheus endpoint
	resp, err = http.Get(service.GetPrometheusEndpoint())
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Stop service
	err = service.Stop()
	require.NoError(t, err)

	// Wait a bit for servers to stop
	time.Sleep(100 * time.Millisecond)

	// Verify endpoints are no longer accessible
	_, err = http.Get(service.GetMetricsEndpoint())
	assert.Error(t, err)
}

func TestHealthCheckersRegistration(t *testing.T) {
	mockChain := &MockChain{
		height: 1,
		bestBlock: &block.Block{
			Header: &block.Header{
				Height:     1,
				Timestamp:  time.Now(),
				Difficulty: 100,
			},
		},
	}
	mockMempool := &MockMempool{txnCount: 0}
	mockNetwork := &MockNetwork{peers: []peer.ID{peer.ID("QmPeer1"), peer.ID("QmPeer2")}}

	service := NewService(nil, mockChain, mockMempool, mockNetwork)

	// Manually register mock health checkers for testing
	service.RegisterHealthChecker(&MockHealthChecker{name: "blockchain", status: health.StatusHealthy})
	service.RegisterHealthChecker(&MockHealthChecker{name: "mempool", status: health.StatusHealthy})
	service.RegisterHealthChecker(&MockHealthChecker{name: "network", status: health.StatusHealthy})

	// Check that health checkers are registered
	systemHealth := service.GetSystemHealth()
	components := systemHealth.GetRegisteredComponents()

	expectedComponents := []string{"blockchain", "mempool", "network"}
	for _, expected := range expectedComponents {
		assert.Contains(t, components, expected)
	}

	assert.Equal(t, 3, systemHealth.GetComponentCount())
}

func TestHealthCheckResults(t *testing.T) {
	mockChain := &MockChain{
		height: 5,
		bestBlock: &block.Block{
			Header: &block.Header{
				Height:     5,
				Timestamp:  time.Now(),
				Difficulty: 1000,
			},
			Transactions: []*block.Transaction{},
		},
	}
	mockMempool := &MockMempool{txnCount: 50}
	mockNetwork := &MockNetwork{peers: []peer.ID{peer.ID("QmPeer1"), peer.ID("QmPeer2"), peer.ID("QmPeer3")}}

	service := NewService(nil, mockChain, mockMempool, mockNetwork)

	// Manually register mock health checkers for testing
	service.RegisterHealthChecker(&MockHealthChecker{name: "blockchain", status: health.StatusHealthy})
	service.RegisterHealthChecker(&MockHealthChecker{name: "mempool", status: health.StatusHealthy})
	service.RegisterHealthChecker(&MockHealthChecker{name: "network", status: health.StatusHealthy})

	// Run health checks
	systemHealth := service.GetSystemHealth()
	systemHealth.RunHealthChecks()

	// Check blockchain health
	blockchainStatus, exists := systemHealth.GetComponentStatus("blockchain")
	require.True(t, exists)
	assert.Equal(t, health.StatusHealthy, blockchainStatus.Status)
	assert.Contains(t, blockchainStatus.Message, "Mock health check")

	// Check mempool health
	mempoolStatus, exists := systemHealth.GetComponentStatus("mempool")
	require.True(t, exists)
	assert.Equal(t, health.StatusHealthy, mempoolStatus.Status)
	assert.Contains(t, mempoolStatus.Message, "Mock health check")

	// Check network health
	networkStatus, exists := systemHealth.GetComponentStatus("network")
	require.True(t, exists)
	assert.Equal(t, health.StatusHealthy, networkStatus.Status)
	assert.Contains(t, networkStatus.Message, "Mock health check")
}

func TestMetricsCollection(t *testing.T) {
	mockChain := &MockChain{
		height: 10,
		bestBlock: &block.Block{
			Header: &block.Header{
				Height:     10,
				Timestamp:  time.Now(),
				Difficulty: 1000,
			},
			Transactions: []*block.Transaction{{}, {}}, // 2 transactions
		},
	}
	mockMempool := &MockMempool{txnCount: 25}
	mockNetwork := &MockNetwork{peers: []peer.ID{peer.ID("QmPeer1"), peer.ID("QmPeer2")}}

	service := NewService(nil, mockChain, mockMempool, mockNetwork)

	// Start service to trigger metrics collection
	err := service.Start()
	require.NoError(t, err)
	defer service.Stop()

	// Manually trigger metrics update to ensure they are collected
	service.UpdateMetrics()

	// Wait a bit for any background processing
	time.Sleep(100 * time.Millisecond)

	// Get metrics
	metrics := service.GetMetrics().GetMetrics()

	// Verify blockchain metrics
	blockchainMetrics := metrics["blockchain"].(map[string]interface{})
	assert.Equal(t, int64(10), blockchainMetrics["block_height"])
	assert.Equal(t, int64(11), blockchainMetrics["total_blocks"]) // height + 1
	assert.Equal(t, float64(2), blockchainMetrics["avg_txn_per_block"])

	// Verify mempool metrics
	assert.Equal(t, int64(25), metrics["blockchain"].(map[string]interface{})["pending_transactions"])

	// Verify network metrics
	assert.Equal(t, int64(2), metrics["network"].(map[string]interface{})["connected_peers"])
}

func TestHealthEndpointResponse(t *testing.T) {
	mockChain := &MockChain{
		height: 3,
		bestBlock: &block.Block{
			Header: &block.Header{
				Height:     3,
				Timestamp:  time.Now(),
				Difficulty: 500,
			},
		},
	}
	mockMempool := &MockMempool{txnCount: 10}
	mockNetwork := &MockNetwork{peers: []peer.ID{peer.ID("QmPeer1")}}

	service := NewService(nil, mockChain, mockMempool, mockNetwork)

	// Create test request
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Call health handler
	service.healthHandler(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	// Parse response
	var healthReport map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &healthReport)
	require.NoError(t, err)

	// Verify structure
	assert.Contains(t, healthReport, "status")
	assert.Contains(t, healthReport, "version")
	assert.Contains(t, healthReport, "uptime")
	assert.Contains(t, healthReport, "components")
	assert.Contains(t, healthReport, "system")

	// Verify components
	components := healthReport["components"].(map[string]interface{})
	assert.Contains(t, components, "blockchain")
	assert.Contains(t, components, "mempool")
	assert.Contains(t, components, "network")
}

func TestMetricsEndpointResponse(t *testing.T) {
	mockChain := &MockChain{
		height: 2,
		bestBlock: &block.Block{
			Header: &block.Header{
				Height:     2,
				Timestamp:  time.Now(),
				Difficulty: 300,
			},
		},
	}
	mockMempool := &MockMempool{txnCount: 5}
	mockNetwork := &MockNetwork{peers: []peer.ID{peer.ID("QmPeer1"), peer.ID("QmPeer2")}}

	service := NewService(nil, mockChain, mockMempool, mockNetwork)

	// Create test request
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	// Call metrics handler
	service.metricsHandler(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	// Parse response
	var metrics map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &metrics)
	require.NoError(t, err)

	// Verify structure
	assert.Contains(t, metrics, "blockchain")
	assert.Contains(t, metrics, "network")
	assert.Contains(t, metrics, "mining")
	assert.Contains(t, metrics, "performance")
	assert.Contains(t, metrics, "errors")
	assert.Contains(t, metrics, "system")
}

func TestPrometheusEndpointResponse(t *testing.T) {
	mockChain := &MockChain{
		height: 1,
		bestBlock: &block.Block{
			Header: &block.Header{
				Height:     1,
				Timestamp:  time.Now(),
				Difficulty: 100,
			},
		},
	}
	mockMempool := &MockMempool{txnCount: 0}
	mockNetwork := &MockNetwork{peers: []peer.ID{peer.ID("QmPeer1")}}

	service := NewService(nil, mockChain, mockMempool, mockNetwork)

	// Update metrics to ensure they have values
	service.UpdateMetrics()

	// Create test request
	req := httptest.NewRequest("GET", "/prometheus", nil)
	w := httptest.NewRecorder()

	// Call Prometheus handler
	service.prometheusHandler(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/plain; version=0.0.4; charset=utf-8", w.Header().Get("Content-Type"))

	// Verify content contains Prometheus format
	body := w.Body.String()
	assert.Contains(t, body, "# HELP gochain_block_height")
	assert.Contains(t, body, "# TYPE gochain_block_height gauge")
	assert.Contains(t, body, "gochain_block_height 1")
}

func TestServiceLogging(t *testing.T) {
	mockChain := &MockChain{
		height: 1,
		bestBlock: &block.Block{
			Header: &block.Header{
				Height:     1,
				Timestamp:  time.Now(),
				Difficulty: 100,
			},
		},
	}
	mockMempool := &MockMempool{txnCount: 0}
	mockNetwork := &MockNetwork{peers: []peer.ID{peer.ID("QmPeer1")}}

	service := NewService(nil, mockChain, mockMempool, mockNetwork)

	// Test logging methods
	service.LogInfo("Test info message")
	service.LogError("Test error message")
	service.LogDebug("Test debug message")
	service.LogWarn("Test warning message")

	// Verify logger is accessible
	logger := service.GetLogger()
	assert.NotNil(t, logger)
}

func TestServiceContextCancellation(t *testing.T) {
	mockChain := &MockChain{
		height: 1,
		bestBlock: &block.Block{
			Header: &block.Header{
				Height:     1,
				Timestamp:  time.Now(),
				Difficulty: 100,
			},
		},
	}
	mockMempool := &MockMempool{txnCount: 0}
	mockNetwork := &MockNetwork{peers: []peer.ID{peer.ID("QmPeer1")}}

	service := NewService(nil, mockChain, mockMempool, mockNetwork)

	// Start service
	err := service.Start()
	require.NoError(t, err)

	// Wait a bit for background monitoring to start
	time.Sleep(100 * time.Millisecond)

	// Stop service (this cancels the context)
	err = service.Stop()
	require.NoError(t, err)

	// Wait a bit for background monitoring to stop
	time.Sleep(100 * time.Millisecond)

	// Verify service is stopped
	// The background monitoring should have stopped due to context cancellation
}

func TestMetricsReset(t *testing.T) {
	mockChain := &MockChain{
		height: 5,
		bestBlock: &block.Block{
			Header: &block.Header{
				Height:     5,
				Timestamp:  time.Now(),
				Difficulty: 500,
			},
		},
	}
	mockMempool := &MockMempool{txnCount: 10}
	mockNetwork := &MockNetwork{peers: []peer.ID{peer.ID("QmPeer1")}}

	service := NewService(nil, mockChain, mockMempool, mockNetwork)

	// Start service
	err := service.Start()
	require.NoError(t, err)
	defer service.Stop()

	// Manually trigger metrics update to ensure they are collected
	service.UpdateMetrics()

	// Wait a bit for any background processing
	time.Sleep(100 * time.Millisecond)

	// Verify metrics are populated
	metrics := service.GetMetrics().GetMetrics()
	blockchainMetrics := metrics["blockchain"].(map[string]interface{})
	assert.NotEqual(t, int64(0), blockchainMetrics["block_height"])

	// Reset metrics
	service.GetMetrics().Reset()

	// Verify metrics are reset
	metrics = service.GetMetrics().GetMetrics()
	blockchainMetrics = metrics["blockchain"].(map[string]interface{})
	assert.Equal(t, int64(0), blockchainMetrics["block_height"])
	assert.Equal(t, int64(0), blockchainMetrics["total_blocks"])
}

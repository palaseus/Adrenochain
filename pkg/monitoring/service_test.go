package monitoring

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/palaseus/adrenochain/pkg/block"
	"github.com/palaseus/adrenochain/pkg/health"
	"github.com/palaseus/adrenochain/pkg/logger"
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

// getAvailablePort finds an available port for testing
func getAvailablePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()

	return l.Addr().(*net.TCPAddr).Port, nil
}

// createTestConfig creates a test configuration with dynamic ports
func createTestConfig() (*Config, error) {
	metricsPort, err := getAvailablePort()
	if err != nil {
		return nil, err
	}

	healthPort, err := getAvailablePort()
	if err != nil {
		return nil, err
	}

	return &Config{
		MetricsPort:         metricsPort,
		HealthPort:          healthPort,
		LogLevel:            logger.INFO,
		LogJSON:             false,
		LogFile:             "",
		MetricsPath:         "/metrics",
		HealthPath:          "/health",
		PrometheusPath:      "/prometheus",
		CollectInterval:     30 * time.Second,
		HealthCheckInterval: 10 * time.Second,
		EnablePrometheus:    true,
	}, nil
}

func TestNewService(t *testing.T) {
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
	assert.NotNil(t, service)
	assert.NotNil(t, service.GetLogger())
	assert.NotNil(t, service.GetMetrics())
	assert.NotNil(t, service.GetSystemHealth())
}

func TestServiceStartStop(t *testing.T) {
	// Create test configuration with dynamic ports
	config, err := createTestConfig()
	require.NoError(t, err)

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

	service := NewService(config, mockChain, mockMempool, mockNetwork)

	// Start service
	err = service.Start()
	require.NoError(t, err)

	// Wait a bit for servers to start
	time.Sleep(200 * time.Millisecond)

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

	// Wait longer for servers to fully stop
	time.Sleep(500 * time.Millisecond)

	// Verify endpoints are no longer accessible
	// Use a timeout context to prevent hanging
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Test metrics endpoint after stop
	req, err := http.NewRequestWithContext(ctx, "GET", service.GetMetricsEndpoint(), nil)
	require.NoError(t, err)

	client := &http.Client{Timeout: 1 * time.Second}
	_, err = client.Do(req)
	assert.Error(t, err, "Expected error when accessing stopped service")

	// Test health endpoint after stop
	req, err = http.NewRequestWithContext(ctx, "GET", service.GetHealthEndpoint(), nil)
	require.NoError(t, err)

	_, err = client.Do(req)
	assert.Error(t, err, "Expected error when accessing stopped service")
}

func TestHealthCheckersRegistration(t *testing.T) {
	// Create test configuration with dynamic ports
	config, err := createTestConfig()
	require.NoError(t, err)

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

	service := NewService(config, mockChain, mockMempool, mockNetwork)

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
	// Create test configuration with dynamic ports
	config, err := createTestConfig()
	require.NoError(t, err)

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

	service := NewService(config, mockChain, mockMempool, mockNetwork)

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
	// Create test configuration with dynamic ports
	config, err := createTestConfig()
	require.NoError(t, err)

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

	service := NewService(config, mockChain, mockMempool, mockNetwork)

	// Start service to trigger metrics collection
	err = service.Start()
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
	// Create test configuration with dynamic ports
	config, err := createTestConfig()
	require.NoError(t, err)

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

	service := NewService(config, mockChain, mockMempool, mockNetwork)

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
	err2 := json.Unmarshal(w.Body.Bytes(), &healthReport)
	require.NoError(t, err2)

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
	// Create test configuration with dynamic ports
	config, err := createTestConfig()
	require.NoError(t, err)

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

	service := NewService(config, mockChain, mockMempool, mockNetwork)

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
	err3 := json.Unmarshal(w.Body.Bytes(), &metrics)
	require.NoError(t, err3)

	// Verify structure
	assert.Contains(t, metrics, "blockchain")
	assert.Contains(t, metrics, "network")
	assert.Contains(t, metrics, "mining")
	assert.Contains(t, metrics, "performance")
	assert.Contains(t, metrics, "errors")
	assert.Contains(t, metrics, "system")
}

func TestPrometheusEndpointResponse(t *testing.T) {
	// Create test configuration with dynamic ports
	config, err := createTestConfig()
	require.NoError(t, err)

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

	service := NewService(config, mockChain, mockMempool, mockNetwork)

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
	assert.Contains(t, body, "# HELP adrenochain_block_height")
	assert.Contains(t, body, "# TYPE adrenochain_block_height gauge")
	assert.Contains(t, body, "adrenochain_block_height 1")
}

func TestServiceLogging(t *testing.T) {
	// Create test configuration with dynamic ports
	config, err := createTestConfig()
	require.NoError(t, err)

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

	service := NewService(config, mockChain, mockMempool, mockNetwork)

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
	// Create test configuration with dynamic ports
	config, err := createTestConfig()
	require.NoError(t, err)

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

	service := NewService(config, mockChain, mockMempool, mockNetwork)

	// Start service
	err = service.Start()
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
	// Create test configuration with dynamic ports
	config, err := createTestConfig()
	require.NoError(t, err)

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

	service := NewService(config, mockChain, mockMempool, mockNetwork)

	// Start service
	err = service.Start()
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

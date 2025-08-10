package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/gochain/gochain/pkg/block"
	"github.com/gochain/gochain/pkg/chain"
	"github.com/gochain/gochain/pkg/health"
	"github.com/gochain/gochain/pkg/logger"
	"github.com/libp2p/go-libp2p/core/peer"
)

// ChainInterface defines the interface for blockchain operations
type ChainInterface interface {
	GetHeight() uint64
	GetBestBlock() *block.Block
	GetGenesisBlock() *block.Block
	GetBlockByHeight(height uint64) *block.Block
}

// MempoolInterface defines the interface for mempool operations
type MempoolInterface interface {
	GetTransactionCount() int
}

// NetworkInterface defines the interface for network operations
type NetworkInterface interface {
	GetPeers() []peer.ID
}

// SimpleHealthChecker is a simple health checker for testing
type SimpleHealthChecker struct {
	name   string
	status health.Status
}

func (shc *SimpleHealthChecker) Name() string {
	return shc.name
}

func (shc *SimpleHealthChecker) Check() (*health.Component, error) {
	return &health.Component{
		Name:    shc.name,
		Status:  shc.status,
		Message: "Simple health checker for testing",
		Details: map[string]interface{}{},
	}, nil
}

// Service represents the monitoring service
type Service struct {
	mu            sync.RWMutex
	logger        *logger.Logger
	metrics       *Metrics
	systemHealth  *health.SystemHealth
	chain         ChainInterface
	mempool       MempoolInterface
	network       NetworkInterface
	config        *Config
	ctx           context.Context
	cancel        context.CancelFunc
	metricsServer *http.Server
	healthServer  *http.Server
	checkers      []health.HealthChecker
}

// Config holds configuration for the monitoring service
type Config struct {
	MetricsPort         int
	HealthPort          int
	LogLevel            logger.Level
	LogJSON             bool
	LogFile             string
	MetricsPath         string
	HealthPath          string
	PrometheusPath      string
	CollectInterval     time.Duration
	HealthCheckInterval time.Duration
	EnablePrometheus    bool
}

// DefaultConfig returns default monitoring configuration
func DefaultConfig() *Config {
	return &Config{
		MetricsPort:         9090,
		HealthPort:          8080,
		LogLevel:            logger.INFO,
		LogJSON:             false,
		LogFile:             "",
		MetricsPath:         "/metrics",
		HealthPath:          "/health",
		PrometheusPath:      "/prometheus",
		CollectInterval:     30 * time.Second,
		HealthCheckInterval: 15 * time.Second,
		EnablePrometheus:    true,
	}
}

// NewService creates a new monitoring service
func NewService(config *Config, chain ChainInterface, mempool MempoolInterface, network NetworkInterface) *Service {
	if config == nil {
		config = DefaultConfig()
	}

	// Create logger
	logConfig := &logger.Config{
		Level:   config.LogLevel,
		Prefix:  "monitoring",
		UseJSON: config.LogJSON,
		LogFile: config.LogFile,
		Output:  os.Stdout, // Ensure output is set
	}

	log := logger.NewLogger(logConfig)

	// Create metrics and health
	metrics := NewMetrics()
	systemHealth := health.NewSystemHealth("1.0.0")

	ctx, cancel := context.WithCancel(context.Background())

	service := &Service{
		logger:       log,
		metrics:      metrics,
		systemHealth: systemHealth,
		chain:        chain,
		mempool:      mempool,
		network:      network,
		config:       config,
		ctx:          ctx,
		cancel:       cancel,
		checkers:     make([]health.HealthChecker, 0),
	}

	// Register health checkers
	service.registerHealthCheckers()

	// Start background monitoring
	go service.startBackgroundMonitoring()

	return service
}

// registerHealthCheckers registers all health checkers
func (s *Service) registerHealthCheckers() {
	// Register blockchain health checker
	// Create a wrapper that implements the interface expected by the health checker
	if chainWrapper, ok := s.chain.(*chain.Chain); ok {
		chainChecker := health.NewChainHealthChecker(chainWrapper)
		s.systemHealth.RegisterComponent(chainChecker)
		s.checkers = append(s.checkers, chainChecker)
	} else {
		// For testing or when using mocks, create simple health checkers
		s.logger.Debug("Skipping chain health checker registration (not a *chain.Chain)")

		// Create simple health checkers for testing
		if s.chain != nil {
			simpleChainChecker := &SimpleHealthChecker{
				name:   "blockchain",
				status: health.StatusHealthy,
			}
			s.systemHealth.RegisterComponent(simpleChainChecker)
			s.checkers = append(s.checkers, simpleChainChecker)
		}

		if s.mempool != nil {
			simpleMempoolChecker := &SimpleHealthChecker{
				name:   "mempool",
				status: health.StatusHealthy,
			}
			s.systemHealth.RegisterComponent(simpleMempoolChecker)
			s.checkers = append(s.checkers, simpleMempoolChecker)
		}

		if s.network != nil {
			simpleNetworkChecker := &SimpleHealthChecker{
				name:   "network",
				status: health.StatusHealthy,
			}
			s.systemHealth.RegisterComponent(simpleNetworkChecker)
			s.checkers = append(s.checkers, simpleNetworkChecker)
		}
	}

	s.logger.Info("Health checkers registered")
}

// RegisterHealthChecker manually registers a health checker (useful for testing)
func (s *Service) RegisterHealthChecker(checker health.HealthChecker) {
	s.systemHealth.RegisterComponent(checker)
	s.checkers = append(s.checkers, checker)
}

// startBackgroundMonitoring starts the background monitoring loop
func (s *Service) startBackgroundMonitoring() {
	metricsTicker := time.NewTicker(s.config.CollectInterval)
	healthTicker := time.NewTicker(s.config.HealthCheckInterval)
	defer metricsTicker.Stop()
	defer healthTicker.Stop()

	s.logger.Info("Starting background monitoring")

	for {
		select {
		case <-s.ctx.Done():
			s.logger.Info("Background monitoring stopped")
			return
		case <-metricsTicker.C:
			s.UpdateMetrics()
		case <-healthTicker.C:
			s.runHealthChecks()
		}
	}
}

// UpdateMetrics updates all metrics
func (s *Service) UpdateMetrics() {
	// Update blockchain metrics
	if s.chain != nil {
		bestBlock := s.chain.GetBestBlock()
		if bestBlock != nil {
			s.metrics.UpdateBlockHeight(int64(bestBlock.Header.Height))
			s.metrics.UpdateLastBlockTime(bestBlock.Header.Timestamp)

			// Calculate additional metrics
			if bestBlock.Header.Height > 0 {
				// Calculate average block time
				prevBlock := s.chain.GetBlockByHeight(bestBlock.Header.Height - 1)
				if prevBlock != nil {
					blockTime := bestBlock.Header.Timestamp.Sub(prevBlock.Header.Timestamp)
					s.metrics.UpdateAvgBlockTime(int64(blockTime.Seconds()))
				}

				// Calculate average transactions per block
				txnCount := len(bestBlock.Transactions)
				if txnCount > 0 {
					s.metrics.UpdateAvgTxnPerBlock(float64(txnCount))
				}

				// Calculate average block size (rough estimate)
				blockSize := int64(len(bestBlock.Transactions) * 256) // Rough estimate
				s.metrics.UpdateAvgBlockSize(blockSize)
			}
		}
		s.metrics.UpdateTotalBlocks(int64(s.chain.GetHeight() + 1))
	}

	// Update mempool metrics
	if s.mempool != nil {
		s.metrics.UpdatePendingTxns(int64(s.mempool.GetTransactionCount()))
	}

	// Update network metrics
	if s.network != nil {
		peers := s.network.GetPeers()
		s.metrics.UpdateConnectedPeers(int64(len(peers)))
	}

	// Update system metrics
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	s.metrics.UpdateMemoryUsage(int64(m.Alloc))

	s.logger.Debug("Metrics updated")
}

// runHealthChecks runs health checks for all registered components
func (s *Service) runHealthChecks() {
	var wg sync.WaitGroup

	for _, checker := range s.checkers {
		wg.Add(1)
		go func(c health.HealthChecker) {
			defer wg.Done()
			s.runComponentCheck(c)
		}(checker)
	}

	wg.Wait()
	s.logger.Debug("Health checks completed")
}

// runComponentCheck runs a health check for a single component
func (s *Service) runComponentCheck(checker health.HealthChecker) {
	start := time.Now()
	component, err := checker.Check()
	checkTime := time.Since(start)

	if err != nil {
		component.Status = health.StatusUnhealthy
		component.Message = err.Error()
	}

	component.LastCheck = time.Now()
	component.CheckTime = checkTime

	s.systemHealth.UpdateComponent(
		checker.Name(),
		component.Status,
		component.Message,
		component.Details,
	)
}

// Start starts the monitoring service
func (s *Service) Start() error {
	s.logger.Info("Starting monitoring service")

	// Start metrics server
	if err := s.startMetricsServer(); err != nil {
		return fmt.Errorf("failed to start metrics server: %w", err)
	}

	// Start health server
	if err := s.startHealthServer(); err != nil {
		return fmt.Errorf("failed to start health server: %w", err)
	}

	s.logger.Info("Monitoring service started successfully")
	return nil
}

// startMetricsServer starts the metrics HTTP server
func (s *Service) startMetricsServer() error {
	mux := http.NewServeMux()
	mux.HandleFunc(s.config.MetricsPath, s.metricsHandler)

	if s.config.EnablePrometheus {
		mux.HandleFunc(s.config.PrometheusPath, s.prometheusHandler)
	}

	s.metricsServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.config.MetricsPort),
		Handler: mux,
	}

	go func() {
		if err := s.metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("Metrics server error: %v", err)
		}
	}()

	s.logger.Info("Metrics server started on port %d", s.config.MetricsPort)
	return nil
}

// startHealthServer starts the health check HTTP server
func (s *Service) startHealthServer() error {
	mux := http.NewServeMux()
	mux.HandleFunc(s.config.HealthPath, s.healthHandler)

	s.healthServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.config.HealthPort),
		Handler: mux,
	}

	go func() {
		if err := s.healthServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("Health server error: %v", err)
		}
	}()

	s.logger.Info("Health server started on port %d", s.config.HealthPort)
	return nil
}

// metricsHandler handles metrics requests
func (s *Service) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	metrics := s.metrics.GetMetrics()
	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		http.Error(w, "Failed to encode metrics", http.StatusInternalServerError)
		return
	}
}

// prometheusHandler handles Prometheus metrics requests
func (s *Service) prometheusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")

	prometheusMetrics := s.metrics.GetPrometheusMetrics()
	w.Write([]byte(prometheusMetrics))
}

// healthHandler handles health check requests
func (s *Service) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	healthReport := s.systemHealth.GetHealthReport()
	if err := json.NewEncoder(w).Encode(healthReport); err != nil {
		http.Error(w, "Failed to encode health report", http.StatusInternalServerError)
		return
	}
}

// Stop stops the monitoring service
func (s *Service) Stop() error {
	s.logger.Info("Stopping monitoring service")

	s.cancel()

	// Stop metrics server
	if s.metricsServer != nil {
		if err := s.metricsServer.Shutdown(context.Background()); err != nil {
			s.logger.Error("Failed to shutdown metrics server: %v", err)
		}
	}

	// Stop health server
	if s.healthServer != nil {
		if err := s.healthServer.Shutdown(context.Background()); err != nil {
			s.logger.Error("Failed to shutdown health server: %v", err)
		}
	}

	s.logger.Info("Monitoring service stopped")
	return nil
}

// GetLogger returns the logger instance
func (s *Service) GetLogger() *logger.Logger {
	return s.logger
}

// GetMetrics returns the metrics instance
func (s *Service) GetMetrics() *Metrics {
	return s.metrics
}

// GetSystemHealth returns the system health instance
func (s *Service) GetSystemHealth() *health.SystemHealth {
	return s.systemHealth
}

// LogInfo logs an info message
func (s *Service) LogInfo(format string, args ...interface{}) {
	s.logger.Info(format, args...)
}

// LogError logs an error message
func (s *Service) LogError(format string, args ...interface{}) {
	s.logger.Error(format, args...)
}

// LogDebug logs a debug message
func (s *Service) LogDebug(format string, args ...interface{}) {
	s.logger.Debug(format, args...)
}

// LogWarn logs a warning message
func (s *Service) LogWarn(format string, args ...interface{}) {
	s.logger.Warn(format, args...)
}

// GetMetricsEndpoint returns the metrics endpoint URL
func (s *Service) GetMetricsEndpoint() string {
	return fmt.Sprintf("http://localhost:%d%s", s.config.MetricsPort, s.config.MetricsPath)
}

// GetHealthEndpoint returns the health endpoint URL
func (s *Service) GetHealthEndpoint() string {
	return fmt.Sprintf("http://localhost:%d%s", s.config.HealthPort, s.config.HealthPath)
}

// GetPrometheusEndpoint returns the Prometheus endpoint URL
func (s *Service) GetPrometheusEndpoint() string {
	if !s.config.EnablePrometheus {
		return ""
	}
	return fmt.Sprintf("http://localhost:%d%s", s.config.MetricsPort, s.config.PrometheusPath)
}

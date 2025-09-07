package router

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

// HealthMonitor monitors the health of clusters and nodes
type HealthMonitor struct {
	mu            sync.RWMutex
	checkInterval time.Duration
	timeout       time.Duration
	healthChecks  map[NodeID]*HealthCheck
	healthHistory map[NodeID][]*HealthRecord
	callbacks     []HealthCallback
	ctx           context.Context
	cancel        context.CancelFunc
	logger        *Logger
}

// HealthCheck represents a health check configuration for a node
type HealthCheck struct {
	NodeID      NodeID
	Address     string
	Port        int
	CheckType   HealthCheckType
	Interval    time.Duration
	Timeout     time.Duration
	Retries     int
	LastCheck   time.Time
	LastResult  *HealthResult
	IsEnabled   bool
	CustomCheck func(*Node) *HealthResult
}

// HealthResult represents the result of a health check
type HealthResult struct {
	NodeID    NodeID
	IsHealthy bool
	Latency   time.Duration
	Error     error
	Details   map[string]interface{}
	Timestamp time.Time
	CheckType HealthCheckType
}

// HealthRecord represents a historical health record
type HealthRecord struct {
	Result    *HealthResult
	Timestamp time.Time
}

// HealthCheckType represents the type of health check
type HealthCheckType string

const (
	HealthCheckTypeTCP    HealthCheckType = "tcp"
	HealthCheckTypeHTTP   HealthCheckType = "http"
	HealthCheckTypeCustom HealthCheckType = "custom"
	HealthCheckTypePing   HealthCheckType = "ping"
)

// HealthCallback is a function called when health status changes
type HealthCallback func(nodeID NodeID, result *HealthResult)

// NewHealthMonitor creates a new health monitor
func NewHealthMonitor(checkInterval time.Duration) *HealthMonitor {
	ctx, cancel := context.WithCancel(context.Background())

	monitor := &HealthMonitor{
		checkInterval: checkInterval,
		timeout:       5 * time.Second,
		healthChecks:  make(map[NodeID]*HealthCheck),
		healthHistory: make(map[NodeID][]*HealthRecord),
		callbacks:     make([]HealthCallback, 0),
		ctx:           ctx,
		cancel:        cancel,
		logger:        NewLogger("health-monitor"),
	}

	// Start monitoring loop
	go monitor.startMonitoring()

	return monitor
}

// RegisterNode registers a node for health monitoring
func (hm *HealthMonitor) RegisterNode(node *Node, checkType HealthCheckType) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	healthCheck := &HealthCheck{
		NodeID:    node.ID,
		Address:   node.Address,
		Port:      node.Port,
		CheckType: checkType,
		Interval:  hm.checkInterval,
		Timeout:   hm.timeout,
		Retries:   3,
		IsEnabled: true,
	}

	hm.healthChecks[node.ID] = healthCheck
	hm.healthHistory[node.ID] = make([]*HealthRecord, 0)

	hm.logger.Info("Registered health check for node %s", node.ID)
	return nil
}

// UnregisterNode removes a node from health monitoring
func (hm *HealthMonitor) UnregisterNode(nodeID NodeID) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	delete(hm.healthChecks, nodeID)
	delete(hm.healthHistory, nodeID)

	hm.logger.Info("Unregistered health check for node %s", nodeID)
}

// AddCallback adds a health status change callback
func (hm *HealthMonitor) AddCallback(callback HealthCallback) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	hm.callbacks = append(hm.callbacks, callback)
}

// CheckNodeHealth performs an immediate health check on a node
func (hm *HealthMonitor) CheckNodeHealth(nodeID NodeID) (*HealthResult, error) {
	hm.mu.RLock()
	healthCheck, exists := hm.healthChecks[nodeID]
	hm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("node %s not registered for health monitoring", nodeID)
	}

	return hm.performHealthCheck(healthCheck)
}

// GetNodeHealth returns the latest health status for a node
func (hm *HealthMonitor) GetNodeHealth(nodeID NodeID) (*HealthResult, bool) {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	healthCheck, exists := hm.healthChecks[nodeID]
	if !exists || healthCheck.LastResult == nil {
		return nil, false
	}

	return healthCheck.LastResult, true
}

// GetHealthHistory returns the health history for a node
func (hm *HealthMonitor) GetHealthHistory(nodeID NodeID, limit int) []*HealthRecord {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	history, exists := hm.healthHistory[nodeID]
	if !exists {
		return nil
	}

	if limit <= 0 || limit >= len(history) {
		return history
	}

	// Return the most recent records
	start := len(history) - limit
	return history[start:]
}

// GetUnhealthyNodes returns all currently unhealthy nodes
func (hm *HealthMonitor) GetUnhealthyNodes() []NodeID {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	var unhealthy []NodeID
	for nodeID, healthCheck := range hm.healthChecks {
		if healthCheck.LastResult != nil && !healthCheck.LastResult.IsHealthy {
			unhealthy = append(unhealthy, nodeID)
		}
	}

	return unhealthy
}

// GetHealthyNodes returns all currently healthy nodes
func (hm *HealthMonitor) GetHealthyNodes() []NodeID {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	var healthy []NodeID
	for nodeID, healthCheck := range hm.healthChecks {
		if healthCheck.LastResult != nil && healthCheck.LastResult.IsHealthy {
			healthy = append(healthy, nodeID)
		}
	}

	return healthy
}

// startMonitoring starts the health monitoring loop
func (hm *HealthMonitor) startMonitoring() {
	ticker := time.NewTicker(hm.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-hm.ctx.Done():
			return
		case <-ticker.C:
			hm.performAllHealthChecks()
		}
	}
}

// performAllHealthChecks performs health checks on all registered nodes
func (hm *HealthMonitor) performAllHealthChecks() {
	hm.mu.RLock()
	healthChecks := make([]*HealthCheck, 0, len(hm.healthChecks))
	for _, check := range hm.healthChecks {
		if check.IsEnabled {
			healthChecks = append(healthChecks, check)
		}
	}
	hm.mu.RUnlock()

	// Perform checks concurrently
	var wg sync.WaitGroup
	for _, healthCheck := range healthChecks {
		wg.Add(1)
		go func(check *HealthCheck) {
			defer wg.Done()
			_, _ = hm.performHealthCheck(check)
		}(healthCheck)
	}

	wg.Wait()
}

// performHealthCheck performs a health check on a specific node
func (hm *HealthMonitor) performHealthCheck(healthCheck *HealthCheck) (*HealthResult, error) {
	var result *HealthResult
	var err error

	// Perform check with retries
	for attempt := 0; attempt <= healthCheck.Retries; attempt++ {
		result, err = hm.executeHealthCheck(healthCheck)
		if err != nil {
			result.IsHealthy = false
			result.Error = err
		}

		if result.IsHealthy || attempt == healthCheck.Retries {
			break
		}

		// Wait before retry
		time.Sleep(time.Duration(attempt+1) * time.Second)
	}

	// Update health check
	hm.mu.Lock()
	healthCheck.LastCheck = time.Now()
	healthCheck.LastResult = result

	// Store in history
	record := &HealthRecord{
		Result:    result,
		Timestamp: time.Now(),
	}
	hm.healthHistory[healthCheck.NodeID] = append(hm.healthHistory[healthCheck.NodeID], record)

	// Keep only last 100 records
	if len(hm.healthHistory[healthCheck.NodeID]) > 100 {
		hm.healthHistory[healthCheck.NodeID] = hm.healthHistory[healthCheck.NodeID][1:]
	}

	// Notify callbacks
	callbacks := make([]HealthCallback, len(hm.callbacks))
	copy(callbacks, hm.callbacks)
	hm.mu.Unlock()

	// Call callbacks
	for _, callback := range callbacks {
		go callback(healthCheck.NodeID, result)
	}

	return result, nil
}

// executeHealthCheck executes the actual health check
func (hm *HealthMonitor) executeHealthCheck(healthCheck *HealthCheck) (*HealthResult, error) {
	start := time.Now()

	result := &HealthResult{
		NodeID:    healthCheck.NodeID,
		CheckType: healthCheck.CheckType,
		Timestamp: time.Now(),
		Details:   make(map[string]interface{}),
	}

	var err error
	switch healthCheck.CheckType {
	case HealthCheckTypeTCP:
		result, err = hm.performTCPCheck(healthCheck, result)
	case HealthCheckTypeHTTP:
		result, err = hm.performHTTPCheck(healthCheck, result)
	case HealthCheckTypeCustom:
		if healthCheck.CustomCheck != nil {
			// Custom check would need the node object
			result.IsHealthy = false
			result.Error = fmt.Errorf("custom check not implemented")
		} else {
			result.IsHealthy = false
			result.Error = fmt.Errorf("no custom check function provided")
		}
	case HealthCheckTypePing:
		result, err = hm.performPingCheck(healthCheck, result)
	default:
		result.IsHealthy = false
		result.Error = fmt.Errorf("unknown health check type: %s", healthCheck.CheckType)
	}

	result.Latency = time.Since(start)
	return result, err
}

// performTCPCheck performs a TCP connection check
func (hm *HealthMonitor) performTCPCheck(healthCheck *HealthCheck, result *HealthResult) (*HealthResult, error) {
	address := fmt.Sprintf("%s:%d", healthCheck.Address, healthCheck.Port)

	conn, err := net.DialTimeout("tcp", address, healthCheck.Timeout)
	if err != nil {
		result.IsHealthy = false
		result.Error = err
		result.Details["error"] = err.Error()
		return result, nil
	}

	conn.Close()
	result.IsHealthy = true
	result.Details["connection"] = "successful"

	return result, nil
}

// performHTTPCheck performs an HTTP health check
func (hm *HealthMonitor) performHTTPCheck(healthCheck *HealthCheck, result *HealthResult) (*HealthResult, error) {
	// This would implement HTTP health checks
	// For now, fall back to TCP check
	return hm.performTCPCheck(healthCheck, result)
}

// performPingCheck performs a ping check
func (hm *HealthMonitor) performPingCheck(healthCheck *HealthCheck, result *HealthResult) (*HealthResult, error) {
	// This would implement ICMP ping checks
	// For now, fall back to TCP check
	return hm.performTCPCheck(healthCheck, result)
}

// SetCheckInterval updates the health check interval
func (hm *HealthMonitor) SetCheckInterval(interval time.Duration) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	hm.checkInterval = interval

	// Update all registered health checks
	for _, healthCheck := range hm.healthChecks {
		healthCheck.Interval = interval
	}
}

// SetTimeout updates the health check timeout
func (hm *HealthMonitor) SetTimeout(timeout time.Duration) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	hm.timeout = timeout

	// Update all registered health checks
	for _, healthCheck := range hm.healthChecks {
		healthCheck.Timeout = timeout
	}
}

// EnableNode enables health checking for a node
func (hm *HealthMonitor) EnableNode(nodeID NodeID) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if healthCheck, exists := hm.healthChecks[nodeID]; exists {
		healthCheck.IsEnabled = true
		hm.logger.Info("Enabled health check for node %s", nodeID)
	}
}

// DisableNode disables health checking for a node
func (hm *HealthMonitor) DisableNode(nodeID NodeID) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if healthCheck, exists := hm.healthChecks[nodeID]; exists {
		healthCheck.IsEnabled = false
		hm.logger.Info("Disabled health check for node %s", nodeID)
	}
}

// GetStats returns health monitoring statistics
func (hm *HealthMonitor) GetStats() *HealthMonitorStats {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	stats := &HealthMonitorStats{
		TotalNodes:     len(hm.healthChecks),
		EnabledNodes:   0,
		HealthyNodes:   0,
		UnhealthyNodes: 0,
		CheckInterval:  hm.checkInterval,
		Timeout:        hm.timeout,
	}

	for _, healthCheck := range hm.healthChecks {
		if healthCheck.IsEnabled {
			stats.EnabledNodes++
		}

		if healthCheck.LastResult != nil {
			if healthCheck.LastResult.IsHealthy {
				stats.HealthyNodes++
			} else {
				stats.UnhealthyNodes++
			}
		}
	}

	return stats
}

// HealthMonitorStats contains statistics about the health monitor
type HealthMonitorStats struct {
	TotalNodes     int
	EnabledNodes   int
	HealthyNodes   int
	UnhealthyNodes int
	CheckInterval  time.Duration
	Timeout        time.Duration
}

// Close shuts down the health monitor
func (hm *HealthMonitor) Close() error {
	hm.cancel()
	hm.logger.Info("Health monitor shut down")
	return nil
}

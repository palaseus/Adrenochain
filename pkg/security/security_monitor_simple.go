package security

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/palaseus/adrenochain/pkg/logger"
)

// SimpleSecurityMonitor provides basic security monitoring
type SimpleSecurityMonitor struct {
	mu                  sync.RWMutex
	logger              *logger.Logger
	events              map[string]*SimpleSecurityEvent
	metrics             map[string]interface{}
	running             bool
	ctx                 context.Context
	cancel              context.CancelFunc
	transactionCount    int
	totalGasUsed        int
	lastTransactionTime time.Time
	startTime           time.Time
	alertManager        *AlertManager
}

// SimpleSecurityEvent represents a security event
type SimpleSecurityEvent struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Severity    string                 `json:"severity"`
	Timestamp   time.Time              `json:"timestamp"`
	Source      string                 `json:"source"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// SimpleSecurityMetrics tracks security metrics
type SimpleSecurityMetrics struct {
	TotalEvents       int64            `json:"total_events"`
	EventsByType      map[string]int64 `json:"events_by_type"`
	EventsBySeverity  map[string]int64 `json:"events_by_severity"`
	AnomaliesDetected int64            `json:"anomalies_detected"`
	AlertsTriggered   int64            `json:"alerts_triggered"`
	LastUpdated       time.Time        `json:"last_updated"`
}

// NewSimpleSecurityMonitor creates a new simple security monitor
func NewSimpleSecurityMonitor(logger *logger.Logger) *SimpleSecurityMonitor {
	return &SimpleSecurityMonitor{
		logger:  logger,
		events:  make(map[string]*SimpleSecurityEvent),
		metrics: make(map[string]interface{}),
	}
}

// LogEvent logs a security event
func (sm *SimpleSecurityMonitor) LogEvent(event *SimpleSecurityEvent) {
	if event == nil {
		return
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Set timestamp if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Generate ID if not provided
	if event.ID == "" {
		event.ID = fmt.Sprintf("evt_%d", time.Now().UnixNano())
	}

	// Store event
	sm.events[event.ID] = event

	// Update metrics
	sm.metrics["total_events"] = sm.metrics["total_events"].(int64) + 1
	sm.metrics["events_by_type"] = sm.metrics["events_by_type"].(map[string]int64)
	sm.metrics["events_by_severity"] = sm.metrics["events_by_severity"].(map[string]int64)
	sm.metrics["last_updated"] = time.Now()

	// Log the event
	sm.logger.Info("Security event: %s - %s (%s)", event.Type, event.Description, event.Severity)

	// Check for critical events
	if event.Severity == "critical" {
		sm.logger.Error("CRITICAL SECURITY EVENT: %s", event.Description)
		sm.metrics["alerts_triggered"] = sm.metrics["alerts_triggered"].(int64) + 1
	}
}

// GetMetrics returns current security metrics
func (sm *SimpleSecurityMonitor) GetMetrics() map[string]interface{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Return a copy
	metrics := make(map[string]interface{})
	for k, v := range sm.metrics {
		metrics[k] = v
	}

	return metrics
}

// GetRecentEvents returns recent security events
func (sm *SimpleSecurityMonitor) GetRecentEvents(limit int) []*SimpleSecurityEvent {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	events := make([]*SimpleSecurityEvent, 0, len(sm.events))
	for _, event := range sm.events {
		events = append(events, event)
	}

	// Sort by timestamp (most recent first)
	for i := 0; i < len(events)-1; i++ {
		for j := i + 1; j < len(events); j++ {
			if events[i].Timestamp.Before(events[j].Timestamp) {
				events[i], events[j] = events[j], events[i]
			}
		}
	}

	if limit > 0 && limit < len(events) {
		events = events[:limit]
	}

	return events
}

// Start begins monitoring with active threat detection
func (sm *SimpleSecurityMonitor) Start(ctx context.Context) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.running {
		return fmt.Errorf("security monitor is already running")
	}

	sm.running = true
	sm.ctx, sm.cancel = context.WithCancel(ctx)

	// Start background monitoring goroutine
	go sm.monitoringLoop()

	sm.logger.Info("Simple security monitor started with active threat detection")
	return nil
}

// Stop stops monitoring with graceful shutdown
func (sm *SimpleSecurityMonitor) Stop() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if !sm.running {
		return
	}

	sm.running = false
	if sm.cancel != nil {
		sm.cancel()
	}

	// Final metrics collection
	sm.collectFinalMetrics()

	sm.logger.Info("Simple security monitor stopped gracefully")
}

// monitoringLoop runs the background monitoring process
func (sm *SimpleSecurityMonitor) monitoringLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sm.ctx.Done():
			return
		case <-ticker.C:
			sm.performSecurityScan()
		}
	}
}

// performSecurityScan performs a comprehensive security scan
func (sm *SimpleSecurityMonitor) performSecurityScan() {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Check for suspicious patterns
	sm.checkSuspiciousPatterns()

	// Monitor resource usage
	sm.monitorResourceUsage()

	// Check for anomalies
	sm.checkAnomalies()

	// Update metrics
	sm.updateMetrics()
}

// checkSuspiciousPatterns looks for suspicious activity patterns
func (sm *SimpleSecurityMonitor) checkSuspiciousPatterns() {
	// Check for rapid successive transactions
	if sm.transactionCount > 100 {
		sm.logger.Warn("High transaction volume detected: %d transactions", sm.transactionCount)
		sm.alertManager.SendAlert("High transaction volume", "WARNING")
	}

	// Check for unusual gas usage patterns
	if sm.totalGasUsed > 1000000 {
		sm.logger.Warn("High gas usage detected: %d gas", sm.totalGasUsed)
		sm.alertManager.SendAlert("High gas usage", "WARNING")
	}
}

// monitorResourceUsage monitors system resource usage
func (sm *SimpleSecurityMonitor) monitorResourceUsage() {
	// In a real implementation, this would check actual system resources
	// For now, we'll simulate resource monitoring

	// Check memory usage (simulated)
	memoryUsage := float64(sm.transactionCount) * 0.1
	if memoryUsage > 80.0 {
		sm.logger.Warn("High memory usage detected: %.2f%%", memoryUsage)
		sm.alertManager.SendAlert("High memory usage", "WARNING")
	}

	// Check CPU usage (simulated)
	cpuUsage := float64(sm.totalGasUsed) / 10000.0
	if cpuUsage > 90.0 {
		sm.logger.Warn("High CPU usage detected: %.2f%%", cpuUsage)
		sm.alertManager.SendAlert("High CPU usage", "WARNING")
	}
}

// checkAnomalies checks for anomalous behavior
func (sm *SimpleSecurityMonitor) checkAnomalies() {
	// Check for unusual transaction patterns
	if sm.transactionCount > 0 {
		avgGasPerTx := float64(sm.totalGasUsed) / float64(sm.transactionCount)
		if avgGasPerTx > 50000 {
			sm.logger.Warn("Unusual gas usage per transaction: %.2f", avgGasPerTx)
			sm.alertManager.SendAlert("Unusual gas usage pattern", "WARNING")
		}
	}

	// Check for time-based anomalies
	now := time.Now()
	if sm.lastTransactionTime.IsZero() || now.Sub(sm.lastTransactionTime) > 10*time.Minute {
		sm.logger.Info("No recent transaction activity")
	}
}

// updateMetrics updates security metrics
func (sm *SimpleSecurityMonitor) updateMetrics() {
	sm.metrics["transactions_processed"] = sm.transactionCount
	sm.metrics["total_gas_used"] = sm.totalGasUsed
	sm.metrics["last_scan_time"] = time.Now().Unix()
	sm.metrics["monitoring_duration"] = time.Since(sm.startTime).Seconds()
}

// collectFinalMetrics collects final metrics before shutdown
func (sm *SimpleSecurityMonitor) collectFinalMetrics() {
	sm.metrics["final_transaction_count"] = sm.transactionCount
	sm.metrics["final_gas_used"] = sm.totalGasUsed
	sm.metrics["shutdown_time"] = time.Now().Unix()
	sm.metrics["total_monitoring_time"] = time.Since(sm.startTime).Seconds()

	sm.logger.Info("Final metrics collected: %+v", sm.metrics)
}

// AlertManager handles security alerts
type AlertManager struct {
	logger *logger.Logger
}

// NewAlertManager creates a new alert manager
func NewAlertManager() *AlertManager {
	return &AlertManager{
		logger: logger.NewLogger(&logger.Config{Level: logger.INFO, Prefix: "alert_manager"}),
	}
}

// SendAlert sends a security alert
func (am *AlertManager) SendAlert(message, severity string) {
	am.logger.Warn("Security Alert [%s]: %s", severity, message)
}

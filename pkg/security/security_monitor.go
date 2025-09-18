package security

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/palaseus/adrenochain/pkg/logger"
)

// SecurityMonitor provides comprehensive security monitoring and anomaly detection
type SecurityMonitor struct {
	mu              sync.RWMutex
	logger          *logger.Logger
	events          map[string]*SecurityMonitorEvent
	anomalies       []*Anomaly
	metrics         *SecurityMetrics
	alertThresholds *AlertThresholds
	eventChannel    chan *SecurityMonitorEvent
	stopChannel     chan struct{}
	isRunning       bool
}

// SecurityMonitorEvent represents a security-related event for monitoring
type SecurityMonitorEvent struct {
	ID          string                 `json:"id"`
	Type        MonitorEventType       `json:"type"`
	Severity    MonitorSeverityLevel   `json:"severity"`
	Timestamp   time.Time              `json:"timestamp"`
	Source      string                 `json:"source"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
	IPAddress   string                 `json:"ip_address,omitempty"`
	UserAgent   string                 `json:"user_agent,omitempty"`
}

// MonitorEventType represents the type of security event
type MonitorEventType string

const (
	MonitorEventTypeAuthentication     MonitorEventType = "authentication"
	MonitorEventTypeAuthorization      MonitorEventType = "authorization"
	MonitorEventTypeInputValidation    MonitorEventType = "input_validation"
	MonitorEventTypeRateLimit          MonitorEventType = "rate_limit"
	MonitorEventTypeSuspiciousActivity MonitorEventType = "suspicious_activity"
	MonitorEventTypeSystemError        MonitorEventType = "system_error"
	MonitorEventTypeDataAccess         MonitorEventType = "data_access"
	MonitorEventTypeNetwork            MonitorEventType = "network"
	MonitorEventTypeContract           MonitorEventType = "contract"
	MonitorEventTypeDeFi               MonitorEventType = "defi"
)

// MonitorSeverityLevel represents the severity of a security event
type MonitorSeverityLevel string

const (
	MonitorSeverityLow      MonitorSeverityLevel = "low"
	MonitorSeverityMedium   MonitorSeverityLevel = "medium"
	MonitorSeverityHigh     MonitorSeverityLevel = "high"
	MonitorSeverityCritical MonitorSeverityLevel = "critical"
)

// Anomaly represents a detected security anomaly
type Anomaly struct {
	ID          string                 `json:"id"`
	Type        AnomalyType            `json:"type"`
	Severity    string                 `json:"severity"`
	Timestamp   time.Time              `json:"timestamp"`
	Description string                 `json:"description"`
	Confidence  float64                `json:"confidence"`
	Metadata    map[string]interface{} `json:"metadata"`
	Resolved    bool                   `json:"resolved"`
}

// AnomalyType represents the type of anomaly detected
type AnomalyType string

const (
	AnomalyTypeUnusualTraffic    AnomalyType = "unusual_traffic"
	AnomalyTypeSuspiciousPattern AnomalyType = "suspicious_pattern"
	AnomalyTypeFailedAttempts    AnomalyType = "failed_attempts"
	AnomalyTypeDataExfiltration  AnomalyType = "data_exfiltration"
	AnomalyTypeSystemIntrusion   AnomalyType = "system_intrusion"
	AnomalyTypeContractExploit   AnomalyType = "contract_exploit"
	AnomalyTypeDeFiManipulation  AnomalyType = "defi_manipulation"
)

// SecurityMetrics tracks security-related metrics
type SecurityMetrics struct {
	TotalEvents       int64                          `json:"total_events"`
	EventsByType      map[MonitorEventType]int64     `json:"events_by_type"`
	EventsBySeverity  map[MonitorSeverityLevel]int64 `json:"events_by_severity"`
	AnomaliesDetected int64                          `json:"anomalies_detected"`
	AlertsTriggered   int64                          `json:"alerts_triggered"`
	LastUpdated       time.Time                      `json:"last_updated"`
}

// AlertThresholds defines thresholds for triggering alerts
type AlertThresholds struct {
	MaxEventsPerMinute    int           `json:"max_events_per_minute"`
	MaxFailedAttempts     int           `json:"max_failed_attempts"`
	MaxSuspiciousActivity int           `json:"max_suspicious_activity"`
	AlertCooldown         time.Duration `json:"alert_cooldown"`
}

// NewSecurityMonitor creates a new security monitor
func NewSecurityMonitor(logger *logger.Logger) *SecurityMonitor {
	return &SecurityMonitor{
		logger:    logger,
		events:    make(map[string]*SecurityMonitorEvent),
		anomalies: make([]*Anomaly, 0),
		metrics: &SecurityMetrics{
			EventsByType:     make(map[MonitorEventType]int64),
			EventsBySeverity: make(map[MonitorSeverityLevel]int64),
		},
		alertThresholds: &AlertThresholds{
			MaxEventsPerMinute:    100,
			MaxFailedAttempts:     10,
			MaxSuspiciousActivity: 5,
			AlertCooldown:         5 * time.Minute,
		},
		eventChannel: make(chan *SecurityMonitorEvent, 1000),
		stopChannel:  make(chan struct{}),
	}
}

// Start begins the security monitoring
func (sm *SecurityMonitor) Start(ctx context.Context) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.isRunning {
		return fmt.Errorf("security monitor is already running")
	}

	sm.isRunning = true
	go sm.eventProcessor(ctx)
	go sm.anomalyDetector(ctx)
	go sm.metricsCollector(ctx)

	sm.logger.Info("Security monitor started")
	return nil
}

// Stop stops the security monitoring
func (sm *SecurityMonitor) Stop() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if !sm.isRunning {
		return
	}

	close(sm.stopChannel)
	sm.isRunning = false
	sm.logger.Info("Security monitor stopped")
}

// LogEvent logs a security event
func (sm *SecurityMonitor) LogEvent(event *SecurityMonitorEvent) {
	if event == nil {
		return
	}

	// Set timestamp if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Generate ID if not provided
	if event.ID == "" {
		event.ID = generateEventID()
	}

	// Send to event channel for processing
	select {
	case sm.eventChannel <- event:
	default:
		sm.logger.Warn("Event channel full, dropping event: %s", event.ID)
	}
}

// eventProcessor processes security events
func (sm *SecurityMonitor) eventProcessor(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-sm.stopChannel:
			return
		case event := <-sm.eventChannel:
			sm.processEvent(event)
		}
	}
}

// processEvent processes a single security event
func (sm *SecurityMonitor) processEvent(event *SecurityMonitorEvent) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Store event
	sm.events[event.ID] = event

	// Update metrics
	sm.metrics.TotalEvents++
	sm.metrics.EventsByType[event.Type]++
	sm.metrics.EventsBySeverity[event.Severity]++
	sm.metrics.LastUpdated = time.Now()

	// Log the event
	sm.logger.Info("Security event: %s - %s (%s)", event.Type, event.Description, event.Severity)

	// Check for immediate alerts
	sm.checkImmediateAlerts(event)
}

// anomalyDetector detects security anomalies
func (sm *SecurityMonitor) anomalyDetector(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-sm.stopChannel:
			return
		case <-ticker.C:
			sm.detectAnomalies()
		}
	}
}

// detectAnomalies detects various types of security anomalies
func (sm *SecurityMonitor) detectAnomalies() {
	sm.mu.RLock()
	events := make([]*SecurityMonitorEvent, 0, len(sm.events))
	for _, event := range sm.events {
		events = append(events, event)
	}
	sm.mu.RUnlock()

	// Detect unusual traffic patterns
	sm.detectUnusualTraffic(events)

	// Detect suspicious patterns
	sm.detectSuspiciousPatterns(events)

	// Detect failed authentication attempts
	sm.detectFailedAttempts(events)

	// Detect potential contract exploits
	sm.detectContractExploits(events)
}

// detectUnusualTraffic detects unusual traffic patterns
func (sm *SecurityMonitor) detectUnusualTraffic(events []*SecurityMonitorEvent) {
	// Count events in the last minute
	recentEvents := 0
	cutoff := time.Now().Add(-1 * time.Minute)

	for _, event := range events {
		if event.Timestamp.After(cutoff) {
			recentEvents++
		}
	}

	// Check if traffic exceeds threshold
	if recentEvents > sm.alertThresholds.MaxEventsPerMinute {
		anomaly := &Anomaly{
			ID:          generateAnomalyID(),
			Type:        AnomalyTypeUnusualTraffic,
			Severity:    "high",
			Timestamp:   time.Now(),
			Description: fmt.Sprintf("Unusual traffic detected: %d events in last minute", recentEvents),
			Confidence:  0.85,
			Metadata: map[string]interface{}{
				"event_count": recentEvents,
				"threshold":   sm.alertThresholds.MaxEventsPerMinute,
			},
		}
		sm.addAnomaly(anomaly)
	}
}

// detectSuspiciousPatterns detects suspicious activity patterns
func (sm *SecurityMonitor) detectSuspiciousPatterns(events []*SecurityMonitorEvent) {
	// Count suspicious events in the last 5 minutes
	suspiciousCount := 0
	cutoff := time.Now().Add(-5 * time.Minute)

	for _, event := range events {
		if event.Timestamp.After(cutoff) && event.Type == MonitorEventTypeSuspiciousActivity {
			suspiciousCount++
		}
	}

	if suspiciousCount > sm.alertThresholds.MaxSuspiciousActivity {
		anomaly := &Anomaly{
			ID:          generateAnomalyID(),
			Type:        AnomalyTypeSuspiciousPattern,
			Severity:    "medium",
			Timestamp:   time.Now(),
			Description: fmt.Sprintf("Suspicious activity pattern detected: %d events in last 5 minutes", suspiciousCount),
			Confidence:  0.75,
			Metadata: map[string]interface{}{
				"suspicious_count": suspiciousCount,
				"threshold":        sm.alertThresholds.MaxSuspiciousActivity,
			},
		}
		sm.addAnomaly(anomaly)
	}
}

// detectFailedAttempts detects failed authentication attempts
func (sm *SecurityMonitor) detectFailedAttempts(events []*SecurityMonitorEvent) {
	// Count failed authentication attempts in the last 10 minutes
	failedCount := 0
	cutoff := time.Now().Add(-10 * time.Minute)

	for _, event := range events {
		if event.Timestamp.After(cutoff) &&
			event.Type == MonitorEventTypeAuthentication &&
			event.Severity == MonitorSeverityHigh {
			failedCount++
		}
	}

	if failedCount > sm.alertThresholds.MaxFailedAttempts {
		anomaly := &Anomaly{
			ID:          generateAnomalyID(),
			Type:        AnomalyTypeFailedAttempts,
			Severity:    "high",
			Timestamp:   time.Now(),
			Description: fmt.Sprintf("Multiple failed authentication attempts: %d in last 10 minutes", failedCount),
			Confidence:  0.90,
			Metadata: map[string]interface{}{
				"failed_count": failedCount,
				"threshold":    sm.alertThresholds.MaxFailedAttempts,
			},
		}
		sm.addAnomaly(anomaly)
	}
}

// detectContractExploits detects potential smart contract exploits
func (sm *SecurityMonitor) detectContractExploits(events []*SecurityMonitorEvent) {
	// Count contract-related high severity events
	exploitCount := 0
	cutoff := time.Now().Add(-15 * time.Minute)

	for _, event := range events {
		if event.Timestamp.After(cutoff) &&
			event.Type == MonitorEventTypeContract &&
			event.Severity == MonitorSeverityCritical {
			exploitCount++
		}
	}

	if exploitCount > 0 {
		anomaly := &Anomaly{
			ID:          generateAnomalyID(),
			Type:        AnomalyTypeContractExploit,
			Severity:    "critical",
			Timestamp:   time.Now(),
			Description: fmt.Sprintf("Potential contract exploit detected: %d critical events", exploitCount),
			Confidence:  0.95,
			Metadata: map[string]interface{}{
				"exploit_count": exploitCount,
			},
		}
		sm.addAnomaly(anomaly)
	}
}

// addAnomaly adds a new anomaly
func (sm *SecurityMonitor) addAnomaly(anomaly *Anomaly) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.anomalies = append(sm.anomalies, anomaly)
	sm.metrics.AnomaliesDetected++

	sm.logger.Warn("Security anomaly detected: %s - %s (confidence: %.2f)",
		anomaly.Type, anomaly.Description, anomaly.Confidence)

	// Trigger alert if severity is high enough
	if anomaly.Severity == "high" || anomaly.Severity == "critical" {
		sm.triggerAlert(anomaly)
	}
}

// triggerAlert triggers a security alert
func (sm *SecurityMonitor) triggerAlert(anomaly *Anomaly) {
	sm.metrics.AlertsTriggered++

	alert := map[string]interface{}{
		"type":        "security_alert",
		"anomaly_id":  anomaly.ID,
		"severity":    anomaly.Severity,
		"description": anomaly.Description,
		"timestamp":   time.Now(),
		"confidence":  anomaly.Confidence,
	}

	// Log critical alert
	alertJSON, _ := json.Marshal(alert)
	sm.logger.Error("SECURITY ALERT: %s", string(alertJSON))

	// In a real implementation, this would send alerts to:
	// - Security team
	// - Monitoring systems
	// - Incident response systems
}

// checkImmediateAlerts checks for immediate alert conditions
func (sm *SecurityMonitor) checkImmediateAlerts(event *SecurityMonitorEvent) {
	// Immediate alert for critical events
	if event.Severity == MonitorSeverityCritical {
		sm.triggerImmediateAlert(event)
	}
}

// triggerImmediateAlert triggers an immediate alert
func (sm *SecurityMonitor) triggerImmediateAlert(event *SecurityMonitorEvent) {
	alert := map[string]interface{}{
		"type":        "immediate_security_alert",
		"event_id":    event.ID,
		"severity":    event.Severity,
		"description": event.Description,
		"timestamp":   time.Now(),
		"source":      event.Source,
	}

	alertJSON, _ := json.Marshal(alert)
	sm.logger.Error("IMMEDIATE SECURITY ALERT: %s", string(alertJSON))
}

// metricsCollector collects and updates security metrics
func (sm *SecurityMonitor) metricsCollector(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-sm.stopChannel:
			return
		case <-ticker.C:
			sm.updateMetrics()
		}
	}
}

// updateMetrics updates security metrics
func (sm *SecurityMonitor) updateMetrics() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Clean up old events (keep last 24 hours)
	cutoff := time.Now().Add(-24 * time.Hour)
	for id, event := range sm.events {
		if event.Timestamp.Before(cutoff) {
			delete(sm.events, id)
		}
	}

	// Clean up old anomalies (keep last 7 days)
	anomalyCutoff := time.Now().Add(-7 * 24 * time.Hour)
	var recentAnomalies []*Anomaly
	for _, anomaly := range sm.anomalies {
		if anomaly.Timestamp.After(anomalyCutoff) {
			recentAnomalies = append(recentAnomalies, anomaly)
		}
	}
	sm.anomalies = recentAnomalies
}

// GetMetrics returns current security metrics
func (sm *SecurityMonitor) GetMetrics() *SecurityMetrics {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Return a copy to avoid race conditions
	metrics := &SecurityMetrics{
		TotalEvents:       sm.metrics.TotalEvents,
		EventsByType:      make(map[MonitorEventType]int64),
		EventsBySeverity:  make(map[MonitorSeverityLevel]int64),
		AnomaliesDetected: sm.metrics.AnomaliesDetected,
		AlertsTriggered:   sm.metrics.AlertsTriggered,
		LastUpdated:       sm.metrics.LastUpdated,
	}

	for k, v := range sm.metrics.EventsByType {
		metrics.EventsByType[k] = v
	}
	for k, v := range sm.metrics.EventsBySeverity {
		metrics.EventsBySeverity[k] = v
	}

	return metrics
}

// GetAnomalies returns recent anomalies
func (sm *SecurityMonitor) GetAnomalies(limit int) []*Anomaly {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if limit <= 0 || limit > len(sm.anomalies) {
		limit = len(sm.anomalies)
	}

	// Return most recent anomalies
	start := len(sm.anomalies) - limit
	if start < 0 {
		start = 0
	}

	anomalies := make([]*Anomaly, limit)
	copy(anomalies, sm.anomalies[start:])
	return anomalies
}

// Helper functions
func generateEventID() string {
	return fmt.Sprintf("evt_%d", time.Now().UnixNano())
}

func generateAnomalyID() string {
	return fmt.Sprintf("anom_%d", time.Now().UnixNano())
}

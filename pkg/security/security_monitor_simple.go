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
	mu      sync.RWMutex
	logger  *logger.Logger
	events  map[string]*SimpleSecurityEvent
	metrics *SimpleSecurityMetrics
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
		logger: logger,
		events: make(map[string]*SimpleSecurityEvent),
		metrics: &SimpleSecurityMetrics{
			EventsByType:     make(map[string]int64),
			EventsBySeverity: make(map[string]int64),
		},
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
	sm.metrics.TotalEvents++
	sm.metrics.EventsByType[event.Type]++
	sm.metrics.EventsBySeverity[event.Severity]++
	sm.metrics.LastUpdated = time.Now()

	// Log the event
	sm.logger.Info("Security event: %s - %s (%s)", event.Type, event.Description, event.Severity)

	// Check for critical events
	if event.Severity == "critical" {
		sm.logger.Error("CRITICAL SECURITY EVENT: %s", event.Description)
		sm.metrics.AlertsTriggered++
	}
}

// GetMetrics returns current security metrics
func (sm *SimpleSecurityMonitor) GetMetrics() *SimpleSecurityMetrics {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Return a copy
	metrics := &SimpleSecurityMetrics{
		TotalEvents:       sm.metrics.TotalEvents,
		EventsByType:      make(map[string]int64),
		EventsBySeverity:  make(map[string]int64),
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

// Start begins monitoring (placeholder for interface compatibility)
func (sm *SimpleSecurityMonitor) Start(ctx context.Context) error {
	sm.logger.Info("Simple security monitor started")
	return nil
}

// Stop stops monitoring (placeholder for interface compatibility)
func (sm *SimpleSecurityMonitor) Stop() {
	sm.logger.Info("Simple security monitor stopped")
}

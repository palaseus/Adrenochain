package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/palaseus/adrenochain/pkg/logger"
	"github.com/palaseus/adrenochain/pkg/security"
)

// SecurityMonitoringIntegration demonstrates how to integrate security monitoring
func main() {
	// Initialize logger with sanitization enabled
	loggerConfig := &logger.Config{
		Level:        logger.INFO,
		Prefix:       "security",
		UseJSON:      true,
		SanitizeLogs: true, // Enable log sanitization
	}

	securityLogger := logger.NewLogger(loggerConfig)

	// Create security monitor
	monitor := security.NewSecurityMonitor(securityLogger)

	// Start monitoring
	ctx := context.Background()
	if err := monitor.Start(ctx); err != nil {
		log.Fatalf("Failed to start security monitor: %v", err)
	}
	defer monitor.Stop()

	// Create security dashboard
	dashboard := security.NewSecurityDashboard(monitor, securityLogger)

	// Start dashboard server in a goroutine
	go func() {
		if err := dashboard.StartDashboard(8080); err != nil {
			securityLogger.Error("Failed to start dashboard: %v", err)
		}
	}()

	securityLogger.Info("Security monitoring system started")
	securityLogger.Info("Dashboard available at: http://localhost:8080")

	// Simulate various security events
	simulateSecurityEvents(monitor)

	// Keep running
	select {}
}

// simulateSecurityEvents simulates various types of security events
func simulateSecurityEvents(monitor *security.SecurityMonitor) {
	// Simulate normal authentication events
	for i := 0; i < 5; i++ {
		event := &security.SecurityEvent{
			Type:        security.EventTypeAuthentication,
			Severity:    security.SeverityLow,
			Source:      "api",
			Description: fmt.Sprintf("User login successful: user_%d", i),
			IPAddress:   "192.168.1.100",
			Metadata: map[string]interface{}{
				"user_id": fmt.Sprintf("user_%d", i),
				"method":  "password",
			},
		}
		monitor.LogEvent(event)
		time.Sleep(100 * time.Millisecond)
	}

	// Simulate failed authentication attempts
	for i := 0; i < 3; i++ {
		event := &security.SecurityEvent{
			Type:        security.EventTypeAuthentication,
			Severity:    security.SeverityHigh,
			Source:      "api",
			Description: fmt.Sprintf("Failed login attempt: user_%d", i),
			IPAddress:   "192.168.1.200",
			Metadata: map[string]interface{}{
				"user_id": fmt.Sprintf("user_%d", i),
				"reason":  "invalid_password",
			},
		}
		monitor.LogEvent(event)
		time.Sleep(200 * time.Millisecond)
	}

	// Simulate suspicious activity
	event := &security.SecurityEvent{
		Type:        security.EventTypeSuspiciousActivity,
		Severity:    security.SeverityMedium,
		Source:      "network",
		Description: "Unusual traffic pattern detected from IP 10.0.0.1",
		IPAddress:   "10.0.0.1",
		Metadata: map[string]interface{}{
			"traffic_volume": 1000,
			"duration":       "5 minutes",
		},
	}
	monitor.LogEvent(event)

	// Simulate rate limiting
	event = &security.SecurityEvent{
		Type:        security.EventTypeRateLimit,
		Severity:    security.SeverityMedium,
		Source:      "api",
		Description: "Rate limit exceeded for IP 192.168.1.50",
		IPAddress:   "192.168.1.50",
		Metadata: map[string]interface{}{
			"requests_per_minute": 150,
			"limit":               100,
		},
	}
	monitor.LogEvent(event)

	// Simulate contract-related event
	event = &security.SecurityEvent{
		Type:        security.EventTypeContract,
		Severity:    security.SeverityHigh,
		Source:      "smart_contract",
		Description: "Suspicious contract interaction detected",
		Metadata: map[string]interface{}{
			"contract_address": "0x1234567890abcdef",
			"function":         "transfer",
			"amount":           "1000000",
		},
	}
	monitor.LogEvent(event)

	// Simulate DeFi event
	event = &security.SecurityEvent{
		Type:        security.EventTypeDeFi,
		Severity:    security.SeverityMedium,
		Source:      "defi_protocol",
		Description: "Large liquidity withdrawal detected",
		Metadata: map[string]interface{}{
			"protocol": "lending",
			"amount":   "5000000",
			"asset":    "USDC",
		},
	}
	monitor.LogEvent(event)

	// Simulate input validation failure
	event = &security.SecurityEvent{
		Type:        security.EventTypeInputValidation,
		Severity:    security.SeverityHigh,
		Source:      "api",
		Description: "Malicious input detected in API request",
		IPAddress:   "192.168.1.75",
		Metadata: map[string]interface{}{
			"endpoint": "/api/block",
			"input":    "<script>alert('xss')</script>",
			"pattern":  "xss_attempt",
		},
	}
	monitor.LogEvent(event)

	// Simulate system error
	event = &security.SecurityEvent{
		Type:        security.EventTypeSystemError,
		Severity:    security.SeverityCritical,
		Source:      "storage",
		Description: "Critical storage system error",
		Metadata: map[string]interface{}{
			"error_code": "STORAGE_001",
			"component":  "block_storage",
		},
	}
	monitor.LogEvent(event)

	// Simulate network event
	event = &security.SecurityEvent{
		Type:        security.EventTypeNetwork,
		Severity:    security.SeverityMedium,
		Source:      "p2p",
		Description: "Peer reputation score dropped below threshold",
		Metadata: map[string]interface{}{
			"peer_id":    "12D3KooW...",
			"reputation": 5.0,
			"threshold":  10.0,
		},
	}
	monitor.LogEvent(event)

	// Simulate data access event
	event = &security.SecurityEvent{
		Type:        security.EventTypeDataAccess,
		Severity:    security.SeverityLow,
		Source:      "api",
		Description: "Sensitive data access logged",
		Metadata: map[string]interface{}{
			"user_id":   "admin_001",
			"data_type": "user_private_keys",
			"operation": "read",
		},
	}
	monitor.LogEvent(event)

	// Wait for anomaly detection to process
	time.Sleep(2 * time.Second)

	// Display current metrics
	metrics := monitor.GetMetrics()
	fmt.Printf("\n=== Security Metrics ===\n")
	fmt.Printf("Total Events: %d\n", metrics.TotalEvents)
	fmt.Printf("Anomalies Detected: %d\n", metrics.AnomaliesDetected)
	fmt.Printf("Alerts Triggered: %d\n", metrics.AlertsTriggered)

	// Display recent anomalies
	anomalies := monitor.GetAnomalies(10)
	fmt.Printf("\n=== Recent Anomalies ===\n")
	for _, anomaly := range anomalies {
		fmt.Printf("- %s: %s (Confidence: %.2f)\n",
			anomaly.Type, anomaly.Description, anomaly.Confidence)
	}

	fmt.Printf("\n=== Security Monitoring Active ===\n")
	fmt.Printf("Dashboard: http://localhost:8080\n")
	fmt.Printf("API Endpoints:\n")
	fmt.Printf("  - GET /api/metrics\n")
	fmt.Printf("  - GET /api/anomalies\n")
	fmt.Printf("  - GET /api/events\n")
	fmt.Printf("  - GET /api/status\n")
}

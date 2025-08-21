package security

import (
	"testing"
	"time"
)

func TestDefaultSecurityConfig(t *testing.T) {
	config := DefaultSecurityConfig()

	if config == nil {
		t.Fatal("expected non-nil config")
	}

	if config.KeySize != 256 {
		t.Errorf("expected KeySize 256, got %d", config.KeySize)
	}

	if config.HashAlgorithm != "SHA-256" {
		t.Errorf("expected HashAlgorithm SHA-256, got %s", config.HashAlgorithm)
	}

	if config.EncryptionMethod != "AES-256-GCM" {
		t.Errorf("expected EncryptionMethod AES-256-GCM, got %s", config.EncryptionMethod)
	}

	if !config.EnableTLS {
		t.Error("expected EnableTLS to be true")
	}

	if !config.RateLimiting {
		t.Error("expected RateLimiting to be true")
	}

	if config.MaxConnections != 1000 {
		t.Errorf("expected MaxConnections 1000, got %d", config.MaxConnections)
	}

	if !config.EnableAuth {
		t.Error("expected EnableAuth to be true")
	}

	if !config.APIKeyRequired {
		t.Error("expected APIKeyRequired to be true")
	}

	if !config.EnableAuditLog {
		t.Error("expected EnableAuditLog to be true")
	}

	if config.LogRetentionDays != 90 {
		t.Errorf("expected LogRetentionDays 90, got %d", config.LogRetentionDays)
	}

	if !config.EnableIntrusionDetection {
		t.Error("expected EnableIntrusionDetection to be true")
	}

	if config.SuspiciousActivityThreshold != 3 {
		t.Errorf("expected SuspiciousActivityThreshold 3, got %d", config.SuspiciousActivityThreshold)
	}

	// Check alert thresholds
	expectedThresholds := map[string]int{
		"failed_logins":  5,
		"suspicious_ips": 10,
		"api_abuse":      100,
	}

	for key, expected := range expectedThresholds {
		if actual, exists := config.AlertThresholds[key]; !exists {
			t.Errorf("expected alert threshold for %s", key)
		} else if actual != expected {
			t.Errorf("expected alert threshold %s to be %d, got %d", key, expected, actual)
		}
	}
}

func TestNewSecurityManager(t *testing.T) {
	config := DefaultSecurityConfig()
	manager := NewSecurityManager(config)

	if manager == nil {
		t.Fatal("expected non-nil manager")
	}

	if manager.config != config {
		t.Error("expected config to be set")
	}

	if manager.auditLog == nil {
		t.Error("expected auditLog to be initialized")
	}

	if manager.blockedIPs == nil {
		t.Error("expected blockedIPs to be initialized")
	}

	if manager.rateLimit == nil {
		t.Error("expected rateLimit to be initialized")
	}
}

func TestSecurityManager_GenerateSecureKey(t *testing.T) {
	config := DefaultSecurityConfig()
	manager := NewSecurityManager(config)

	key, err := manager.GenerateSecureKey()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if key == nil {
		t.Fatal("expected non-nil key")
	}

	expectedKeySize := config.KeySize / 8
	if len(key) != expectedKeySize {
		t.Errorf("expected key size %d, got %d", expectedKeySize, len(key))
	}

	// Test with custom key size
	customConfig := &SecurityConfig{KeySize: 128}
	customManager := NewSecurityManager(customConfig)

	customKey, err := customManager.GenerateSecureKey()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(customKey) != 16 { // 128/8
		t.Errorf("expected custom key size 16, got %d", len(customKey))
	}
}

func TestSecurityManager_HashData(t *testing.T) {
	config := DefaultSecurityConfig()
	manager := NewSecurityManager(config)

	testData := []byte("test data")

	hash, err := manager.HashData(testData)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if hash == nil {
		t.Fatal("expected non-nil hash")
	}

	if len(hash) != 32 { // SHA-256 produces 32 bytes
		t.Errorf("expected hash size 32, got %d", len(hash))
	}

	// Test with unsupported algorithm
	unsupportedConfig := &SecurityConfig{HashAlgorithm: "MD5"}
	unsupportedManager := NewSecurityManager(unsupportedConfig)

	_, err = unsupportedManager.HashData(testData)
	if err == nil {
		t.Error("expected error for unsupported hash algorithm")
	}
}

func TestSecurityManager_ValidateAPIKey(t *testing.T) {
	config := DefaultSecurityConfig()
	manager := NewSecurityManager(config)

	// Test valid API key
	validKey := "valid_api_key_that_is_long_enough_to_pass_validation"
	if !manager.ValidateAPIKey(validKey) {
		t.Error("expected valid API key to pass validation")
	}

	// Test short API key
	shortKey := "short"
	if manager.ValidateAPIKey(shortKey) {
		t.Error("expected short API key to fail validation")
	}

	// Test very long API key
	longKey := string(make([]byte, 129)) // 129 bytes
	if manager.ValidateAPIKey(longKey) {
		t.Error("expected very long API key to fail validation")
	}

	// Test with API key not required
	noAPIKeyConfig := &SecurityConfig{APIKeyRequired: false}
	noAPIKeyManager := NewSecurityManager(noAPIKeyConfig)

	if !noAPIKeyManager.ValidateAPIKey("") {
		t.Error("expected empty API key to pass when not required")
	}
}

func TestSecurityManager_CheckRateLimit(t *testing.T) {
	config := DefaultSecurityConfig()
	manager := NewSecurityManager(config)

	identifier := "test_user"
	maxRequests := 3
	window := 100 * time.Millisecond

	// First 3 requests should be allowed
	for i := 0; i < 3; i++ {
		if !manager.CheckRateLimit(identifier, maxRequests, window) {
			t.Errorf("expected request %d to be allowed", i+1)
		}
	}

	// 4th request should be rate limited
	if manager.CheckRateLimit(identifier, maxRequests, window) {
		t.Error("expected 4th request to be rate limited")
	}

	// Wait for window to expire
	time.Sleep(150 * time.Millisecond)

	// Should be allowed again
	if !manager.CheckRateLimit(identifier, maxRequests, window) {
		t.Error("expected request to be allowed after window expired")
	}

	// Test different identifier
	differentIdentifier := "different_user"
	if !manager.CheckRateLimit(differentIdentifier, maxRequests, window) {
		t.Error("expected different identifier to be allowed")
	}
}

func TestSecurityManager_LogSecurityEvent(t *testing.T) {
	config := DefaultSecurityConfig()
	manager := NewSecurityManager(config)

	eventType := "LOGIN_ATTEMPT"
	description := "Failed login attempt"
	ipAddress := "192.168.1.1"
	userID := "test_user"
	severity := SeverityMedium
	metadata := map[string]interface{}{"source": "web"}

	initialLogCount := len(manager.GetAuditLog())

	manager.LogSecurityEvent(eventType, description, ipAddress, userID, severity, metadata)

	logs := manager.GetAuditLog()
	if len(logs) != initialLogCount+1 {
		t.Errorf("expected log count to increase by 1, got %d", len(logs))
	}

	lastEvent := logs[len(logs)-1]
	if lastEvent.EventType != eventType {
		t.Errorf("expected EventType %s, got %s", eventType, lastEvent.EventType)
	}

	if lastEvent.Description != description {
		t.Errorf("expected Description %s, got %s", description, lastEvent.Description)
	}

	if lastEvent.IPAddress != ipAddress {
		t.Errorf("expected IPAddress %s, got %s", ipAddress, lastEvent.IPAddress)
	}

	if lastEvent.UserID != userID {
		t.Errorf("expected UserID %s, got %s", userID, lastEvent.UserID)
	}

	if lastEvent.Severity != severity {
		t.Errorf("expected Severity %v, got %v", severity, lastEvent.Severity)
	}

	if lastEvent.Metadata["source"] != "web" {
		t.Errorf("expected metadata source 'web', got %v", lastEvent.Metadata["source"])
	}
}

func TestSecurityManager_BlockIP(t *testing.T) {
	config := DefaultSecurityConfig()
	manager := NewSecurityManager(config)

	ipAddress := "192.168.1.100"
	duration := 100 * time.Millisecond

	// IP should not be blocked initially
	if manager.IsIPBlocked(ipAddress) {
		t.Error("expected IP to not be blocked initially")
	}

	// Block the IP
	manager.BlockIP(ipAddress, duration)

	// IP should be blocked
	if !manager.IsIPBlocked(ipAddress) {
		t.Error("expected IP to be blocked")
	}

	// Wait for block to expire
	time.Sleep(150 * time.Millisecond)

	// IP should no longer be blocked
	if manager.IsIPBlocked(ipAddress) {
		t.Error("expected IP to no longer be blocked after expiration")
	}

	// Check that security event was logged
	logs := manager.GetAuditLog()
	blockEventFound := false
	for _, event := range logs {
		if event.EventType == "IP_BLOCKED" && event.IPAddress == ipAddress {
			blockEventFound = true
			break
		}
	}

	if !blockEventFound {
		t.Error("expected IP_BLOCKED event to be logged")
	}
}

func TestSecurityManager_GetAuditLog(t *testing.T) {
	config := DefaultSecurityConfig()
	manager := NewSecurityManager(config)

	// Log some events
	manager.LogSecurityEvent("TEST_EVENT", "Test description", "127.0.0.1", "test", SeverityLow, nil)
	manager.LogSecurityEvent("TEST_EVENT_2", "Test description 2", "127.0.0.2", "test2", SeverityHigh, nil)

	logs := manager.GetAuditLog()
	if len(logs) != 2 {
		t.Errorf("expected 2 log entries, got %d", len(logs))
	}

	// Verify logs are copies (modifying returned logs shouldn't affect internal state)
	logs[0].Description = "modified"

	// Get logs again
	logs2 := manager.GetAuditLog()
	if logs2[0].Description == "modified" {
		t.Error("expected returned logs to be copies, not references")
	}
}

func TestSecurityManager_GetSecurityStats(t *testing.T) {
	config := DefaultSecurityConfig()
	manager := NewSecurityManager(config)

	// Log some events and block some IPs
	manager.LogSecurityEvent("TEST_EVENT", "Test", "127.0.0.1", "test", SeverityLow, nil)
	manager.LogSecurityEvent("TEST_EVENT", "Test", "127.0.0.2", "test", SeverityHigh, nil)
	manager.BlockIP("192.168.1.1", time.Hour)

	stats := manager.GetSecurityStats()

	if stats["total_events"] != 3 {
		t.Errorf("expected total_events 3, got %v", stats["total_events"])
	}

	if stats["blocked_ips"] != 1 {
		t.Errorf("expected blocked_ips 1, got %v", stats["blocked_ips"])
	}

	if stats["rate_limit_trackers"] != 0 {
		t.Errorf("expected rate_limit_trackers 0, got %v", stats["rate_limit_trackers"])
	}

	severityCounts := stats["severity_counts"].(map[SecuritySeverity]int)
	if severityCounts[SeverityLow] != 1 {
		t.Errorf("expected SeverityLow count 1, got %d", severityCounts[SeverityLow])
	}

	if severityCounts[SeverityHigh] != 1 {
		t.Errorf("expected SeverityHigh count 1, got %d", severityCounts[SeverityHigh])
	}
}

func TestSecurityManager_CleanupExpiredBlocks(t *testing.T) {
	config := DefaultSecurityConfig()
	manager := NewSecurityManager(config)

	// Block an IP with short duration
	ipAddress := "192.168.1.100"
	manager.BlockIP(ipAddress, 50*time.Millisecond)

	// Verify it's blocked
	if !manager.IsIPBlocked(ipAddress) {
		t.Error("expected IP to be blocked")
	}

	// Wait for block to expire
	time.Sleep(100 * time.Millisecond)

	// Clean up expired blocks
	manager.CleanupExpiredBlocks()

	// IP should no longer be blocked
	if manager.IsIPBlocked(ipAddress) {
		t.Error("expected IP to no longer be blocked after cleanup")
	}
}

func TestSecurityManager_CheckAlertThresholds(t *testing.T) {
	config := DefaultSecurityConfig()
	manager := NewSecurityManager(config)

	// Log events up to the threshold
	eventType := "failed_logins"
	threshold := config.AlertThresholds[eventType]

	for i := 0; i < threshold; i++ {
		manager.LogSecurityEvent(eventType, "Failed login", "127.0.0.1", "user", SeverityMedium, nil)
	}

	// Get logs to check if alert was generated
	logs := manager.GetAuditLog()
	alertFound := false
	for _, event := range logs {
		if event.EventType == "ALERT_THRESHOLD_EXCEEDED" {
			alertFound = true
			break
		}
	}

	if !alertFound {
		t.Error("expected alert threshold exceeded event to be generated")
	}
}

func TestSecurityManager_Concurrency(t *testing.T) {
	config := DefaultSecurityConfig()
	manager := NewSecurityManager(config)

	// Test concurrent access to security manager
	done := make(chan bool)

	// Start multiple goroutines
	for i := 0; i < 10; i++ {
		go func(id int) {
			// Log events
			manager.LogSecurityEvent("CONCURRENT_TEST", "Test", "127.0.0.1", "user", SeverityLow, nil)

			// Check rate limits
			manager.CheckRateLimit("user", 5, time.Second)

			// Block IPs
			manager.BlockIP("192.168.1.1", time.Millisecond)

			// Get stats
			_ = manager.GetSecurityStats()

			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify no panics occurred and data is consistent
	stats := manager.GetSecurityStats()
	if stats["total_events"] == nil {
		t.Error("expected stats to be accessible after concurrent operations")
	}
}

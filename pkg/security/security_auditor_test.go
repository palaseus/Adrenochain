package security

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestNewSecurityAuditor(t *testing.T) {
	config := SecurityAuditConfig{
		EnableConsensusAuditing: true,
		EnableContractAuditing:  true,
		EnableNetworkAuditing:   true,
		EnableEconomicAuditing:  true,
	}

	auditor := NewSecurityAuditor(config)
	if auditor == nil {
		t.Fatal("NewSecurityAuditor returned nil")
	}

	if auditor.consensusAuditor == nil {
		t.Error("consensusAuditor should be initialized")
	}
	if auditor.contractAuditor == nil {
		t.Error("contractAuditor should be initialized")
	}
	if auditor.networkAuditor == nil {
		t.Error("networkAuditor should be initialized")
	}
	if auditor.economicAuditor == nil {
		t.Error("economicAuditor should be initialized")
	}

	if auditor.config != config {
		t.Error("config should be set correctly")
	}

	if auditor.auditResults == nil {
		t.Error("auditResults map should be initialized")
	}
	if auditor.auditHistory == nil {
		t.Error("auditHistory slice should be initialized")
	}

	if auditor.TotalAudits != 0 {
		t.Error("TotalAudits should start at 0")
	}
}

func TestSecurityAuditor_RunComprehensiveAudit(t *testing.T) {
	config := SecurityAuditConfig{
		EnableConsensusAuditing: true,
		EnableContractAuditing:  true,
		EnableNetworkAuditing:   true,
		EnableEconomicAuditing:  true,
	}

	auditor := NewSecurityAuditor(config)
	ctx := context.Background()

	result, err := auditor.RunComprehensiveAudit(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result but got nil")
	}

	if result.ID == "" {
		t.Error("expected non-empty audit ID")
	}

	if result.Type != AuditTypeComprehensive {
		t.Errorf("expected Type AuditTypeComprehensive, got %v", result.Type)
	}

	if result.Component != "adrenochain Platform" {
		t.Errorf("expected Component 'adrenochain Platform', got %s", result.Component)
	}

	if result.Status != AuditStatusCompleted {
		t.Errorf("expected Status AuditStatusCompleted, got %v", result.Status)
	}

	if result.Auditor != "adrenochain Security Auditor" {
		t.Errorf("expected Auditor 'adrenochain Security Auditor', got %s", result.Auditor)
	}

	if result.Version != "1.0.0" {
		t.Errorf("expected Version '1.0.0', got %s", result.Version)
	}

	if result.StartTime.IsZero() {
		t.Error("expected non-zero start time")
	}

	if result.EndTime.IsZero() {
		t.Error("expected non-zero end time")
	}

	if result.Duration <= 0 {
		t.Error("expected positive duration")
	}

	if result.Score < 0 || result.Score > 100 {
		t.Errorf("expected score between 0 and 100, got %f", result.Score)
	}

	if result.RiskLevel < RiskLevelLow || result.RiskLevel > RiskLevelCritical {
		t.Errorf("expected valid risk level, got %v", result.RiskLevel)
	}

	if len(result.Recommendations) == 0 {
		t.Error("expected recommendations to be generated")
	}

	// Check that statistics were updated
	if auditor.TotalAudits != 1 {
		t.Errorf("expected TotalAudits to be 1, got %d", auditor.TotalAudits)
	}

	if auditor.LastAudit.IsZero() {
		t.Error("expected LastAudit to be updated")
	}
}

func TestSecurityAuditor_RunComprehensiveAudit_DisabledComponents(t *testing.T) {
	config := SecurityAuditConfig{
		EnableConsensusAuditing: false,
		EnableContractAuditing:  false,
		EnableNetworkAuditing:   false,
		EnableEconomicAuditing:  false,
	}

	auditor := NewSecurityAuditor(config)
	ctx := context.Background()

	result, err := auditor.RunComprehensiveAudit(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result but got nil")
	}

	// Should still complete successfully even with all components disabled
	if result.Status != AuditStatusCompleted {
		t.Errorf("expected Status AuditStatusCompleted, got %v", result.Status)
	}
}

func TestSecurityAuditor_RunConsensusAudit(t *testing.T) {
	config := SecurityAuditConfig{
		EnableConsensusAuditing: true,
	}

	auditor := NewSecurityAuditor(config)
	ctx := context.Background()

	result, err := auditor.RunConsensusAudit(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result but got nil")
	}

	if result.Type != AuditTypeConsensus {
		t.Errorf("expected Type AuditTypeConsensus, got %v", result.Type)
	}

	if result.Status != AuditStatusCompleted {
		t.Errorf("expected Status AuditStatusCompleted, got %v", result.Status)
	}
}

func TestSecurityAuditor_RunConsensusAudit_Disabled(t *testing.T) {
	config := SecurityAuditConfig{
		EnableConsensusAuditing: false,
	}

	auditor := NewSecurityAuditor(config)
	ctx := context.Background()

	result, err := auditor.RunConsensusAudit(ctx)
	if err == nil {
		t.Error("expected error when consensus auditing is disabled")
	}

	if result != nil {
		t.Error("expected nil result when consensus auditing is disabled")
	}

	if err != ErrConsensusAuditingNotEnabled {
		t.Errorf("expected ErrConsensusAuditingNotEnabled, got %v", err)
	}
}

func TestSecurityAuditor_RunContractAudit(t *testing.T) {
	config := SecurityAuditConfig{
		EnableContractAuditing: true,
	}

	auditor := NewSecurityAuditor(config)
	ctx := context.Background()

	result, err := auditor.RunContractAudit(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result but got nil")
	}

	if result.Type != AuditTypeContract {
		t.Errorf("expected Type AuditTypeContract, got %v", result.Type)
	}
}

func TestSecurityAuditor_RunNetworkAudit(t *testing.T) {
	config := SecurityAuditConfig{
		EnableNetworkAuditing: true,
	}

	auditor := NewSecurityAuditor(config)
	ctx := context.Background()

	result, err := auditor.RunNetworkAudit(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result but got nil")
	}

	if result.Type != AuditTypeNetwork {
		t.Errorf("expected Type AuditTypeNetwork, got %v", result.Type)
	}
}

func TestSecurityAuditor_RunEconomicAudit(t *testing.T) {
	config := SecurityAuditConfig{
		EnableEconomicAuditing: true,
	}

	auditor := NewSecurityAuditor(config)
	ctx := context.Background()

	result, err := auditor.RunEconomicAudit(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result but got nil")
	}

	if result.Type != AuditTypeEconomic {
		t.Errorf("expected Type AuditTypeEconomic, got %v", result.Type)
	}
}

func TestSecurityAuditor_GetAuditResult(t *testing.T) {
	config := SecurityAuditConfig{
		EnableConsensusAuditing: true,
	}

	auditor := NewSecurityAuditor(config)
	ctx := context.Background()

	// Run an audit first
	result, err := auditor.RunConsensusAudit(ctx)
	if err != nil {
		t.Fatalf("failed to run audit: %v", err)
	}

	// Get the audit result
	retrievedResult := auditor.GetAuditResult(result.ID)
	if retrievedResult == nil {
		t.Fatal("expected to retrieve audit result")
	}

	if retrievedResult.ID != result.ID {
		t.Errorf("expected ID %s, got %s", result.ID, retrievedResult.ID)
	}
}

func TestSecurityAuditor_GetAuditHistory(t *testing.T) {
	config := SecurityAuditConfig{
		EnableConsensusAuditing: true,
		EnableContractAuditing:  true,
	}

	auditor := NewSecurityAuditor(config)
	ctx := context.Background()

	// Run multiple audits
	_, err := auditor.RunConsensusAudit(ctx)
	if err != nil {
		t.Fatalf("failed to run consensus audit: %v", err)
	}

	_, err = auditor.RunContractAudit(ctx)
	if err != nil {
		t.Fatalf("failed to run contract audit: %v", err)
	}

	history := auditor.GetAuditHistory(10)
	if len(history) != 2 {
		t.Errorf("expected 2 audit results in history, got %d", len(history))
	}
}

func TestSecurityAuditor_GetSecurityStatistics(t *testing.T) {
	config := SecurityAuditConfig{
		EnableConsensusAuditing: true,
	}

	auditor := NewSecurityAuditor(config)
	ctx := context.Background()

	// Run an audit
	_, err := auditor.RunConsensusAudit(ctx)
	if err != nil {
		t.Fatalf("failed to run audit: %v", err)
	}

	stats := auditor.GetSecurityStatistics()
	if stats == nil {
		t.Fatal("expected security statistics")
	}

	if stats.TotalAudits != 1 {
		t.Errorf("expected TotalAudits to be 1, got %d", stats.TotalAudits)
	}
}

func TestSecurityAuditor_calculateAuditScore(t *testing.T) {
	config := SecurityAuditConfig{}
	auditor := NewSecurityAuditor(config)

	// Test with no issues
	result := &AuditResult{
		CriticalIssues: []SecurityIssue{},
		HighIssues:     []SecurityIssue{},
		MediumIssues:   []SecurityIssue{},
		LowIssues:      []SecurityIssue{},
	}

	score := auditor.calculateAuditScore(result)
	if score != 100.0 {
		t.Errorf("expected score 100.0 for no issues, got %f", score)
	}

	// Test with some issues
	result.CriticalIssues = []SecurityIssue{{ID: "test"}}
	result.HighIssues = []SecurityIssue{{ID: "test"}}

	score = auditor.calculateAuditScore(result)
	if score >= 100.0 {
		t.Errorf("expected score < 100.0 for issues, got %f", score)
	}

	if score < 0.0 {
		t.Errorf("expected score >= 0.0, got %f", score)
	}
}

func TestSecurityAuditor_calculateRiskLevel(t *testing.T) {
	config := SecurityAuditConfig{}
	auditor := NewSecurityAuditor(config)

	// Test with no issues
	result := &AuditResult{
		CriticalIssues: []SecurityIssue{},
		HighIssues:     []SecurityIssue{},
		MediumIssues:   []SecurityIssue{},
		LowIssues:      []SecurityIssue{},
	}

	riskLevel := auditor.calculateRiskLevel(result)
	if riskLevel != RiskLevelLow {
		t.Errorf("expected RiskLevelLow for no issues, got %v", riskLevel)
	}

	// Test with critical issues
	result.CriticalIssues = []SecurityIssue{{ID: "test"}}
	riskLevel = auditor.calculateRiskLevel(result)
	if riskLevel != RiskLevelCritical {
		t.Errorf("expected RiskLevelCritical for critical issues, got %v", riskLevel)
	}

	// Test with high issues only
	result.CriticalIssues = []SecurityIssue{}
	result.HighIssues = []SecurityIssue{{ID: "test"}}
	riskLevel = auditor.calculateRiskLevel(result)
	if riskLevel != RiskLevelHigh {
		t.Errorf("expected RiskLevelHigh for high issues, got %v", riskLevel)
	}
}

func TestSecurityAuditor_generateRecommendations(t *testing.T) {
	config := SecurityAuditConfig{}
	auditor := NewSecurityAuditor(config)

	// Test with no issues
	result := &AuditResult{
		CriticalIssues: []SecurityIssue{},
		HighIssues:     []SecurityIssue{},
		MediumIssues:   []SecurityIssue{},
		LowIssues:      []SecurityIssue{},
	}

	recommendations := auditor.generateRecommendations(result)
	if len(recommendations) == 0 {
		t.Error("expected recommendations even for clean audit")
	}

	// Test with issues
	result.CriticalIssues = []SecurityIssue{{ID: "test"}}
	recommendations = auditor.generateRecommendations(result)
	if len(recommendations) == 0 {
		t.Error("expected recommendations for audit with issues")
	}
}

func TestSecurityAuditor_updateAuditStatistics(t *testing.T) {
	config := SecurityAuditConfig{}
	auditor := NewSecurityAuditor(config)

	initialTotal := auditor.TotalAudits
	initialPassed := auditor.PassedAudits

	result := &AuditResult{
		Status:         AuditStatusCompleted,
		CriticalIssues: []SecurityIssue{},
		HighIssues:     []SecurityIssue{},
		MediumIssues:   []SecurityIssue{},
		LowIssues:      []SecurityIssue{},
	}

	auditor.updateAuditStatistics(result)

	if auditor.TotalAudits != initialTotal+1 {
		t.Errorf("expected TotalAudits to increment, got %d", auditor.TotalAudits)
	}

	if auditor.PassedAudits != initialPassed+1 {
		t.Errorf("expected PassedAudits to increment, got %d", auditor.PassedAudits)
	}
}

func TestSecurityAuditor_copyAuditResult(t *testing.T) {
	config := SecurityAuditConfig{}
	auditor := NewSecurityAuditor(config)

	original := &AuditResult{
		ID:              "test_id",
		Type:            AuditTypeConsensus,
		Component:       "test_component",
		StartTime:       time.Now(),
		EndTime:         time.Now(),
		Duration:        time.Second,
		Status:          AuditStatusCompleted,
		Score:           95.0,
		CriticalIssues:  []SecurityIssue{{ID: "critical"}},
		HighIssues:      []SecurityIssue{{ID: "high"}},
		MediumIssues:    []SecurityIssue{{ID: "medium"}},
		LowIssues:       []SecurityIssue{{ID: "low"}},
		Recommendations: []string{"test recommendation"},
		RiskLevel:       RiskLevelLow,
		Auditor:         "test_auditor",
		Version:         "1.0.0",
		Timestamp:       time.Now(),
	}

	copied := auditor.copyAuditResult(original)
	if copied == nil {
		t.Fatal("expected copied result but got nil")
	}

	if copied.ID != original.ID {
		t.Errorf("expected ID %s, got %s", original.ID, copied.ID)
	}

	if copied.Type != original.Type {
		t.Errorf("expected Type %v, got %v", original.Type, copied.Type)
	}

	if copied.Component != original.Component {
		t.Errorf("expected Component %s, got %s", original.Component, copied.Component)
	}

	if copied.Score != original.Score {
		t.Errorf("expected Score %f, got %f", original.Score, copied.Score)
	}

	if copied.RiskLevel != original.RiskLevel {
		t.Errorf("expected RiskLevel %v, got %v", original.RiskLevel, copied.RiskLevel)
	}

	if len(copied.CriticalIssues) != len(original.CriticalIssues) {
		t.Errorf("expected %d critical issues, got %d", len(original.CriticalIssues), len(copied.CriticalIssues))
	}

	if len(copied.Recommendations) != len(original.Recommendations) {
		t.Errorf("expected %d recommendations, got %d", len(original.Recommendations), len(copied.Recommendations))
	}
}

func TestSecurityAuditor_generateAuditID(t *testing.T) {
	id1 := generateAuditID()
	id2 := generateAuditID()

	if id1 == "" {
		t.Error("expected non-empty audit ID")
	}

	if id2 == "" {
		t.Error("expected non-empty audit ID")
	}

	if id1 == id2 {
		t.Error("expected unique audit IDs")
	}
}

func TestSecurityAuditor_Concurrency(t *testing.T) {
	config := SecurityAuditConfig{
		EnableConsensusAuditing: true,
		EnableContractAuditing:  true,
	}

	auditor := NewSecurityAuditor(config)
	ctx := context.Background()

	// Test concurrent audit execution
	const numGoroutines = 5
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			_, err := auditor.RunConsensusAudit(ctx)
			if err != nil {
				t.Errorf("goroutine %d failed: %v", id, err)
			}
		}(i)
	}

	wg.Wait()

	// Verify all audits completed
	if auditor.TotalAudits != uint64(numGoroutines) {
		t.Errorf("expected %d total audits, got %d", numGoroutines, auditor.TotalAudits)
	}
}

func TestSecurityAuditor_EdgeCases(t *testing.T) {
	config := SecurityAuditConfig{}
	auditor := NewSecurityAuditor(config)

	// Test with nil context
	result, err := auditor.RunComprehensiveAudit(nil)
	if err != nil {
		t.Errorf("unexpected error with nil context: %v", err)
	}
	if result == nil {
		t.Fatal("expected result with nil context")
	}

	// Test with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err = auditor.RunComprehensiveAudit(ctx)
	if err != nil {
		t.Errorf("unexpected error with cancelled context: %v", err)
	}
	if result == nil {
		t.Fatal("expected result with cancelled context")
	}

	// Test with timeout context
	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	result, err = auditor.RunComprehensiveAudit(ctx)
	if err != nil {
		t.Errorf("unexpected error with timeout context: %v", err)
	}
	if result == nil {
		t.Fatal("expected result with timeout context")
	}
}

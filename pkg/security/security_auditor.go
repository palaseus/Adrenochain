package security

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// SecurityAuditor provides comprehensive security auditing capabilities
type SecurityAuditor struct {
	mu sync.RWMutex

	// Core components
	consensusAuditor *ConsensusAuditor
	contractAuditor  *ContractAuditor
	networkAuditor   *NetworkAuditor
	economicAuditor  *EconomicAuditor
	
	// Security configuration
	config SecurityAuditConfig
	
	// Audit results
	auditResults map[string]*AuditResult
	auditHistory []*AuditResult
	
	// Statistics
	TotalAudits     uint64
	PassedAudits    uint64
	FailedAudits    uint64
	CriticalIssues  uint64
	HighIssues      uint64
	MediumIssues    uint64
	LowIssues       uint64
	LastAudit       time.Time
}

// SecurityAuditConfig holds configuration for security auditing
type SecurityAuditConfig struct {
	// Audit settings
	EnableConsensusAuditing bool
	EnableContractAuditing  bool
	EnableNetworkAuditing   bool
	EnableEconomicAuditing  bool
	
	// Security thresholds
	MaxCriticalIssues uint64
	MaxHighIssues     uint64
	MaxMediumIssues   uint64
	MaxLowIssues      uint64
	
	// Audit frequency
	AuditInterval     time.Duration
	EnableContinuousAuditing bool
	
	// Reporting
	EnableDetailedReports bool
	EnableVulnerabilityDatabase bool
	ReportOutputPath     string
}

// AuditResult contains the result of a security audit
type AuditResult struct {
	ID              string
	Type            AuditType
	Component       string
	StartTime       time.Time
	EndTime         time.Time
	Duration        time.Duration
	Status          AuditStatus
	Score           float64
	
	// Issues found
	CriticalIssues  []SecurityIssue
	HighIssues      []SecurityIssue
	MediumIssues    []SecurityIssue
	LowIssues       []SecurityIssue
	
	// Recommendations
	Recommendations []string
	RiskLevel       RiskLevel
	
	// Metadata
	Auditor         string
	Version         string
	Timestamp       time.Time
}

// AuditType indicates the type of security audit
type AuditType int

const (
	AuditTypeConsensus AuditType = iota
	AuditTypeContract
	AuditTypeNetwork
	AuditTypeEconomic
	AuditTypeComprehensive
)

// AuditStatus indicates the status of an audit
type AuditStatus int

const (
	AuditStatusPending AuditStatus = iota
	AuditStatusRunning
	AuditStatusCompleted
	AuditStatusFailed
	AuditStatusBlocked
)

// RiskLevel indicates the overall risk level
type RiskLevel int

const (
	RiskLevelLow RiskLevel = iota
	RiskLevelMedium
	RiskLevelHigh
	RiskLevelCritical
)

// SecurityIssue represents a security vulnerability or issue
type SecurityIssue struct {
	ID              string
	Title           string
	Description     string
	Severity        IssueSeverity
	Category        IssueCategory
	CVSS            float64
	Status          IssueStatus
	DiscoveredAt    time.Time
	FixedAt         *time.Time
	FixVersion      string
	References      []string
	Exploitability  ExploitabilityLevel
	Impact          ImpactLevel
}

// IssueSeverity indicates the severity of a security issue
type IssueSeverity int

const (
	IssueSeverityLow IssueSeverity = iota
	IssueSeverityMedium
	IssueSeverityHigh
	IssueSeverityCritical
)

// IssueCategory indicates the category of a security issue
type IssueCategory int

const (
	IssueCategoryReentrancy IssueCategory = iota
	IssueCategoryIntegerOverflow
	IssueCategoryAccessControl
	IssueCategoryLogicError
	IssueCategoryDoS
	IssueCategorySybil
	IssueCategoryConsensus
	IssueCategoryNetwork
	IssueCategoryEconomic
	IssueCategoryOther
)

// IssueStatus indicates the status of a security issue
type IssueStatus int

const (
	IssueStatusOpen IssueStatus = iota
	IssueStatusInProgress
	IssueStatusFixed
	IssueStatusVerified
	IssueStatusClosed
)

// ExploitabilityLevel indicates how easily an issue can be exploited
type ExploitabilityLevel int

const (
	ExploitabilityLevelNone ExploitabilityLevel = iota
	ExploitabilityLevelLow
	ExploitabilityLevelMedium
	ExploitabilityLevelHigh
	ExploitabilityLevelCritical
)

// ImpactLevel indicates the potential impact of an issue
type ImpactLevel int

const (
	ImpactLevelNone ImpactLevel = iota
	ImpactLevelLow
	ImpactLevelMedium
	ImpactLevelHigh
	ImpactLevelCritical
)

// NewSecurityAuditor creates a new security auditor
func NewSecurityAuditor(config SecurityAuditConfig) *SecurityAuditor {
	return &SecurityAuditor{
		consensusAuditor: NewConsensusAuditor(),
		contractAuditor:  NewContractAuditor(),
		networkAuditor:   NewNetworkAuditor(),
		economicAuditor:  NewEconomicAuditor(),
		config:           config,
		auditResults:     make(map[string]*AuditResult),
		auditHistory:     make([]*AuditResult, 0),
		TotalAudits:      0,
		PassedAudits:     0,
		FailedAudits:     0,
		CriticalIssues:   0,
		HighIssues:       0,
		MediumIssues:     0,
		LowIssues:        0,
		LastAudit:        time.Time{},
	}
}

// RunComprehensiveAudit runs a comprehensive security audit
func (sa *SecurityAuditor) RunComprehensiveAudit(ctx context.Context) (*AuditResult, error) {
	sa.mu.Lock()
	defer sa.mu.Unlock()
	
	auditID := generateAuditID()
	startTime := time.Now()
	
	auditResult := &AuditResult{
		ID:        auditID,
		Type:      AuditTypeComprehensive,
		Component: "adrenochain Platform",
		StartTime: startTime,
		Status:    AuditStatusRunning,
		Auditor:   "adrenochain Security Auditor",
		Version:   "1.0.0",
		Timestamp: startTime,
	}
	
	// Run consensus audit
	if sa.config.EnableConsensusAuditing {
		if consensusResult, err := sa.consensusAuditor.Audit(ctx); err == nil {
			auditResult.CriticalIssues = append(auditResult.CriticalIssues, consensusResult.CriticalIssues...)
			auditResult.HighIssues = append(auditResult.HighIssues, consensusResult.HighIssues...)
			auditResult.MediumIssues = append(auditResult.MediumIssues, consensusResult.MediumIssues...)
			auditResult.LowIssues = append(auditResult.LowIssues, consensusResult.LowIssues...)
		}
	}
	
	// Run contract audit
	if sa.config.EnableContractAuditing {
		if contractResult, err := sa.contractAuditor.Audit(ctx); err == nil {
			auditResult.CriticalIssues = append(auditResult.CriticalIssues, contractResult.CriticalIssues...)
			auditResult.HighIssues = append(auditResult.HighIssues, contractResult.HighIssues...)
			auditResult.MediumIssues = append(auditResult.MediumIssues, contractResult.MediumIssues...)
			auditResult.LowIssues = append(auditResult.LowIssues, contractResult.LowIssues...)
		}
	}
	
	// Run network audit
	if sa.config.EnableNetworkAuditing {
		if networkResult, err := sa.networkAuditor.Audit(ctx); err == nil {
			auditResult.CriticalIssues = append(auditResult.CriticalIssues, networkResult.CriticalIssues...)
			auditResult.HighIssues = append(auditResult.HighIssues, networkResult.HighIssues...)
			auditResult.MediumIssues = append(auditResult.MediumIssues, networkResult.MediumIssues...)
			auditResult.LowIssues = append(auditResult.LowIssues, networkResult.LowIssues...)
		}
	}
	
	// Run economic audit
	if sa.config.EnableEconomicAuditing {
		if economicResult, err := sa.economicAuditor.Audit(ctx); err == nil {
			auditResult.CriticalIssues = append(auditResult.CriticalIssues, economicResult.CriticalIssues...)
			auditResult.HighIssues = append(auditResult.HighIssues, economicResult.HighIssues...)
			auditResult.MediumIssues = append(auditResult.MediumIssues, economicResult.MediumIssues...)
			auditResult.LowIssues = append(auditResult.LowIssues, economicResult.LowIssues...)
		}
	}
	
	// Calculate audit score and risk level
	auditResult.Score = sa.calculateAuditScore(auditResult)
	auditResult.RiskLevel = sa.calculateRiskLevel(auditResult)
	
	// Generate recommendations
	auditResult.Recommendations = sa.generateRecommendations(auditResult)
	
	// Finalize audit
	auditResult.EndTime = time.Now()
	auditResult.Duration = auditResult.EndTime.Sub(auditResult.StartTime)
	auditResult.Status = AuditStatusCompleted
	
	// Update statistics
	sa.updateAuditStatistics(auditResult)
	
	// Store result
	sa.auditResults[auditID] = auditResult
	sa.auditHistory = append(sa.auditHistory, auditResult)
	
	return auditResult, nil
}

// RunConsensusAudit runs a consensus security audit
func (sa *SecurityAuditor) RunConsensusAudit(ctx context.Context) (*AuditResult, error) {
	if !sa.config.EnableConsensusAuditing {
		return nil, ErrConsensusAuditingNotEnabled
	}
	
	return sa.consensusAuditor.Audit(ctx)
}

// RunContractAudit runs a contract security audit
func (sa *SecurityAuditor) RunContractAudit(ctx context.Context) (*AuditResult, error) {
	if !sa.config.EnableContractAuditing {
		return nil, ErrContractAuditingNotEnabled
	}
	
	return sa.contractAuditor.Audit(ctx)
}

// RunNetworkAudit runs a network security audit
func (sa *SecurityAuditor) RunNetworkAudit(ctx context.Context) (*AuditResult, error) {
	if !sa.config.EnableNetworkAuditing {
		return nil, ErrNetworkAuditingNotEnabled
	}
	
	return sa.networkAuditor.Audit(ctx)
}

// RunEconomicAudit runs an economic security audit
func (sa *SecurityAuditor) RunEconomicAudit(ctx context.Context) (*AuditResult, error) {
	if !sa.config.EnableEconomicAuditing {
		return nil, ErrEconomicAuditingNotEnabled
	}
	
	return sa.economicAuditor.Audit(ctx)
}

// GetAuditResult returns a specific audit result
func (sa *SecurityAuditor) GetAuditResult(auditID string) *AuditResult {
	sa.mu.RLock()
	defer sa.mu.RUnlock()
	
	if result, exists := sa.auditResults[auditID]; exists {
		// Return a copy to avoid race conditions
		resultCopy := &AuditResult{
			ID:              result.ID,
			Type:            result.Type,
			Component:       result.Component,
			StartTime:       result.StartTime,
			EndTime:         result.EndTime,
			Duration:        result.Duration,
			Status:          result.Status,
			Score:           result.Score,
			RiskLevel:       result.RiskLevel,
			Auditor:         result.Auditor,
			Version:         result.Version,
			Timestamp:       result.Timestamp,
			CriticalIssues:  make([]SecurityIssue, len(result.CriticalIssues)),
			HighIssues:      make([]SecurityIssue, len(result.HighIssues)),
			MediumIssues:    make([]SecurityIssue, len(result.MediumIssues)),
			LowIssues:       make([]SecurityIssue, len(result.LowIssues)),
			Recommendations: make([]string, len(result.Recommendations)),
		}
		
		// Copy issues
		copy(resultCopy.CriticalIssues, result.CriticalIssues)
		copy(resultCopy.HighIssues, result.HighIssues)
		copy(resultCopy.MediumIssues, result.MediumIssues)
		copy(resultCopy.LowIssues, result.LowIssues)
		copy(resultCopy.Recommendations, result.Recommendations)
		
		return resultCopy
	}
	
	return nil
}

// GetAuditHistory returns audit history
func (sa *SecurityAuditor) GetAuditHistory(limit int) []*AuditResult {
	sa.mu.RLock()
	defer sa.mu.RUnlock()
	
	if limit <= 0 || limit > len(sa.auditHistory) {
		limit = len(sa.auditHistory)
	}
	
	// Return recent audits
	recentAudits := sa.auditHistory[len(sa.auditHistory)-limit:]
	
	// Return copies to avoid race conditions
	audits := make([]*AuditResult, len(recentAudits))
	for i, audit := range recentAudits {
		audits[i] = sa.copyAuditResult(audit)
	}
	
	return audits
}

// GetSecurityStatistics returns security audit statistics
func (sa *SecurityAuditor) GetSecurityStatistics() *SecurityStatistics {
	sa.mu.RLock()
	defer sa.mu.RUnlock()
	
	return &SecurityStatistics{
		TotalAudits:    sa.TotalAudits,
		PassedAudits:   sa.PassedAudits,
		FailedAudits:   sa.FailedAudits,
		CriticalIssues: sa.CriticalIssues,
		HighIssues:     sa.HighIssues,
		MediumIssues:   sa.MediumIssues,
		LowIssues:      sa.LowIssues,
		LastAudit:      sa.LastAudit,
		RiskLevel:      sa.calculateOverallRiskLevel(),
	}
}

// Helper functions
func (sa *SecurityAuditor) calculateAuditScore(result *AuditResult) float64 {
	totalIssues := len(result.CriticalIssues) + len(result.HighIssues) + len(result.MediumIssues) + len(result.LowIssues)
	
	if totalIssues == 0 {
		return 100.0
	}
	
	// Weight issues by severity
	criticalWeight := 10.0
	highWeight := 5.0
	mediumWeight := 2.0
	lowWeight := 0.5
	
	weightedScore := float64(len(result.CriticalIssues)) * criticalWeight +
		float64(len(result.HighIssues)) * highWeight +
		float64(len(result.MediumIssues)) * mediumWeight +
		float64(len(result.LowIssues)) * lowWeight
	
	// Convert to 0-100 scale
	score := 100.0 - (weightedScore * 100.0 / 100.0) // Normalize to 100
	
	if score < 0 {
		score = 0
	}
	
	return score
}

func (sa *SecurityAuditor) calculateRiskLevel(result *AuditResult) RiskLevel {
	if len(result.CriticalIssues) > 0 {
		return RiskLevelCritical
	}
	
	if len(result.HighIssues) > 0 {
		return RiskLevelHigh
	}
	
	if len(result.MediumIssues) > 0 {
		return RiskLevelMedium
	}
	
	if len(result.LowIssues) > 0 {
		return RiskLevelLow
	}
	
	return RiskLevelLow
}

func (sa *SecurityAuditor) calculateOverallRiskLevel() RiskLevel {
	if sa.CriticalIssues > 0 {
		return RiskLevelCritical
	}
	
	if sa.HighIssues > 0 {
		return RiskLevelHigh
	}
	
	if sa.MediumIssues > 0 {
		return RiskLevelMedium
	}
	
	return RiskLevelLow
}

func (sa *SecurityAuditor) generateRecommendations(result *AuditResult) []string {
	var recommendations []string
	
	// Critical issues
	if len(result.CriticalIssues) > 0 {
		recommendations = append(recommendations, 
			fmt.Sprintf("Immediately address %d critical security issues", len(result.CriticalIssues)))
	}
	
	// High issues
	if len(result.HighIssues) > 0 {
		recommendations = append(recommendations, 
			fmt.Sprintf("Address %d high-priority security issues within 30 days", len(result.HighIssues)))
	}
	
	// Medium issues
	if len(result.MediumIssues) > 0 {
		recommendations = append(recommendations, 
			fmt.Sprintf("Plan to address %d medium-priority security issues", len(result.MediumIssues)))
	}
	
	// Low issues
	if len(result.LowIssues) > 0 {
		recommendations = append(recommendations, 
			fmt.Sprintf("Consider addressing %d low-priority security issues", len(result.LowIssues)))
	}
	
	// General recommendations
	if result.Score < 80.0 {
		recommendations = append(recommendations, 
			"Implement comprehensive security review process")
	}
	
	if result.Score < 60.0 {
		recommendations = append(recommendations, 
			"Consider engaging external security audit firm")
	}
	
	return recommendations
}

func (sa *SecurityAuditor) updateAuditStatistics(result *AuditResult) {
	sa.TotalAudits++
	sa.LastAudit = time.Now()
	
	// Update issue counts
	sa.CriticalIssues += uint64(len(result.CriticalIssues))
	sa.HighIssues += uint64(len(result.HighIssues))
	sa.MediumIssues += uint64(len(result.MediumIssues))
	sa.LowIssues += uint64(len(result.LowIssues))
	
	// Update audit status counts
	if result.RiskLevel == RiskLevelCritical || result.RiskLevel == RiskLevelHigh {
		sa.FailedAudits++
	} else {
		sa.PassedAudits++
	}
}

func (sa *SecurityAuditor) copyAuditResult(result *AuditResult) *AuditResult {
	// Deep copy implementation
	return &AuditResult{
		ID:              result.ID,
		Type:            result.Type,
		Component:       result.Component,
		StartTime:       result.StartTime,
		EndTime:         result.EndTime,
		Duration:        result.Duration,
		Status:          result.Status,
		Score:           result.Score,
		RiskLevel:       result.RiskLevel,
		Auditor:         result.Auditor,
		Version:         result.Version,
		Timestamp:       result.Timestamp,
		CriticalIssues:  result.CriticalIssues,
		HighIssues:      result.HighIssues,
		MediumIssues:    result.MediumIssues,
		LowIssues:       result.LowIssues,
		Recommendations: result.Recommendations,
	}
}

func generateAuditID() string {
	// Generate unique audit ID
	return fmt.Sprintf("audit_%d", time.Now().UnixNano())
}

// SecurityStatistics contains security audit statistics
type SecurityStatistics struct {
	TotalAudits    uint64
	PassedAudits   uint64
	FailedAudits   uint64
	CriticalIssues uint64
	HighIssues     uint64
	MediumIssues   uint64
	LowIssues      uint64
	LastAudit      time.Time
	RiskLevel      RiskLevel
}

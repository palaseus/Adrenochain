package delegated

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/palaseus/adrenochain/pkg/security"
)

// DelegationType represents the type of delegation
type DelegationType int

const (
	DelegationTypeFull DelegationType = iota
	DelegationTypePartial
	DelegationTypeConditional
	DelegationTypeTemporary
	DelegationTypeCustom
)

// DelegationType.String() returns the string representation
func (dt DelegationType) String() string {
	switch dt {
	case DelegationTypeFull:
		return "full"
	case DelegationTypePartial:
		return "partial"
	case DelegationTypeConditional:
		return "conditional"
	case DelegationTypeTemporary:
		return "temporary"
	case DelegationTypeCustom:
		return "custom"
	default:
		return "unknown"
	}
}

// DelegationStatus represents the status of a delegation
type DelegationStatus int

const (
	DelegationStatusActive DelegationStatus = iota
	DelegationStatusPaused
	DelegationStatusRevoked
	DelegationStatusExpired
	DelegationStatusSuspended
)

// DelegationStatus.String() returns the string representation
func (ds DelegationStatus) String() string {
	switch ds {
	case DelegationStatusActive:
		return "active"
	case DelegationStatusPaused:
		return "paused"
	case DelegationStatusRevoked:
		return "revoked"
	case DelegationStatusExpired:
		return "expired"
	case DelegationStatusSuspended:
		return "suspended"
	default:
		return "unknown"
	}
}

// Delegator represents a delegator in the system
type Delegator struct {
	ID              string                 `json:"id"`
	Address         string                 `json:"address"`
	TotalPower      *big.Int               `json:"total_power"`
	DelegatedPower  *big.Int               `json:"delegated_power"`
	RetainedPower   *big.Int               `json:"retained_power"`
	Reputation      *big.Int               `json:"reputation"`
	CreatedAt       time.Time              `json:"created_at"`
	LastDelegation  time.Time              `json:"last_delegation"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// Delegate represents a delegate in the system
type Delegate struct {
	ID              string                 `json:"id"`
	Address         string                 `json:"address"`
	TotalDelegated  *big.Int               `json:"total_delegated"`
	ActiveDelegations uint64               `json:"active_delegations"`
	Reputation      *big.Int               `json:"reputation"`
	Performance     *big.Int               `json:"performance"`
	CreatedAt       time.Time              `json:"created_at"`
	LastVote        time.Time              `json:"last_vote"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// Delegation represents a delegation relationship
type Delegation struct {
	ID              string                 `json:"id"`
	DelegatorID     string                 `json:"delegator_id"`
	DelegateID      string                 `json:"delegate_id"`
	Type            DelegationType         `json:"type"`
	Power           *big.Int               `json:"power"`
	Conditions      []DelegationCondition  `json:"conditions"`
	Status          DelegationStatus       `json:"status"`
	CreatedAt       time.Time              `json:"created_at"`
	ExpiresAt       time.Time              `json:"expires_at"`
	LastUsed        time.Time              `json:"last_used"`
	ZKProof         *security.ZKProof      `json:"zk_proof"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// DelegationCondition represents a condition for delegation
type DelegationCondition struct {
	ID              string                 `json:"id"`
	Type            ConditionType          `json:"type"`
	Parameter       string                 `json:"parameter"`
	Value           interface{}            `json:"value"`
	Operator        ConditionOperator      `json:"operator"`
	Active          bool                   `json:"active"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// ConditionType represents the type of delegation condition
type ConditionType int

const (
	ConditionTypeTimeBased ConditionType = iota
	ConditionTypePerformanceBased
	ConditionTypeReputationBased
	ConditionTypeVoteBased
	ConditionTypeCustom
)

// ConditionType.String() returns the string representation
func (ct ConditionType) String() string {
	switch ct {
	case ConditionTypeTimeBased:
		return "time_based"
	case ConditionTypePerformanceBased:
		return "performance_based"
	case ConditionTypeReputationBased:
		return "reputation_based"
	case ConditionTypeVoteBased:
		return "vote_based"
	case ConditionTypeCustom:
		return "custom"
	default:
		return "unknown"
	}
}

// ConditionOperator represents the operator for condition evaluation
type ConditionOperator int

const (
	ConditionOperatorGreaterThan ConditionOperator = iota
	ConditionOperatorLessThan
	ConditionOperatorEqualTo
	ConditionOperatorNotEqualTo
	ConditionOperatorContains
	ConditionOperatorCustom
)

// ConditionOperator.String() returns the string representation
func (co ConditionOperator) String() string {
	switch co {
	case ConditionOperatorGreaterThan:
		return "greater_than"
	case ConditionOperatorLessThan:
		return "less_than"
	case ConditionOperatorEqualTo:
		return "equal_to"
	case ConditionOperatorNotEqualTo:
		return "not_equal_to"
	case ConditionOperatorContains:
		return "contains"
	case ConditionOperatorCustom:
		return "custom"
	default:
		return "unknown"
	}
}

// ProxyVote represents a vote cast by a delegate on behalf of delegators
type ProxyVote struct {
	ID              string                 `json:"id"`
	DelegationID    string                 `json:"delegation_id"`
	ProposalID      string                 `json:"proposal_id"`
	VoteType        string                 `json:"vote_type"`
	VotePower       *big.Int               `json:"vote_power"`
	DelegatorCount  uint64                 `json:"delegator_count"`
	Timestamp       time.Time              `json:"timestamp"`
	ZKProof         *security.ZKProof      `json:"zk_proof"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// DelegationMetrics represents metrics for delegation performance
type DelegationMetrics struct {
	DelegationID    string                 `json:"delegation_id"`
	TotalVotes      uint64                 `json:"total_votes"`
	SuccessfulVotes uint64                 `json:"successful_votes"`
	VoteAccuracy    float64                `json:"vote_accuracy"`
	Performance     *big.Int               `json:"performance"`
	LastUpdated     time.Time              `json:"last_updated"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// DelegatedGovernanceConfig represents configuration for the Delegated Governance system
type DelegatedGovernanceConfig struct {
	MaxDelegators       uint64        `json:"max_delegators"`
	MaxDelegates        uint64        `json:"max_delegates"`
	MaxDelegations      uint64        `json:"max_delegations"`
	MaxProxyVotes       uint64        `json:"max_proxy_votes"`
	EncryptionKeySize   int           `json:"encryption_key_size"`
	ZKProofType         security.ProofType `json:"zk_proof_type"`
	DelegationTimeout   time.Duration `json:"delegation_timeout"`
	CleanupInterval     time.Duration `json:"cleanup_interval"`
	MinDelegationPower  *big.Int      `json:"min_delegation_power"`
	MaxDelegationPower  *big.Int      `json:"max_delegation_power"`
}

// DelegatedGovernance represents the main Delegated Governance system
type DelegatedGovernance struct {
	mu              sync.RWMutex
	Delegators      map[string]*Delegator       `json:"delegators"`
	Delegates       map[string]*Delegate        `json:"delegates"`
	Delegations     map[string]*Delegation      `json:"delegations"`
	ProxyVotes      map[string]*ProxyVote       `json:"proxy_votes"`
	Metrics         map[string]*DelegationMetrics `json:"metrics"`
	Config          DelegatedGovernanceConfig   `json:"config"`
	encryptionKey   []byte
	running         bool
	stopChan        chan struct{}
}

// NewDelegatedGovernance creates a new Delegated Governance system
func NewDelegatedGovernance(config DelegatedGovernanceConfig) *DelegatedGovernance {
	// Set default values if not provided
	if config.MaxDelegators == 0 {
		config.MaxDelegators = 10000
	}
	if config.MaxDelegates == 0 {
		config.MaxDelegates = 1000
	}
	if config.MaxDelegations == 0 {
		config.MaxDelegations = 50000
	}
	if config.MaxProxyVotes == 0 {
		config.MaxProxyVotes = 100000
	}
	if config.EncryptionKeySize == 0 {
		config.EncryptionKeySize = 32 // 256 bits
	}
	if config.ZKProofType == 0 {
		config.ZKProofType = security.ProofTypeBulletproofs
	}
	if config.DelegationTimeout == 0 {
		config.DelegationTimeout = time.Hour * 24 * 30 // 30 days
	}
	if config.CleanupInterval == 0 {
		config.CleanupInterval = time.Hour
	}
	if config.MinDelegationPower == nil {
		config.MinDelegationPower = big.NewInt(100)
	}
	if config.MaxDelegationPower == nil {
		config.MaxDelegationPower = big.NewInt(1000000)
	}

	// Generate encryption key
	encryptionKey := make([]byte, config.EncryptionKeySize)
	if _, err := rand.Read(encryptionKey); err != nil {
		panic(fmt.Sprintf("Failed to generate encryption key: %v", err))
	}

	return &DelegatedGovernance{
		Delegators:     make(map[string]*Delegator),
		Delegates:      make(map[string]*Delegate),
		Delegations:    make(map[string]*Delegation),
		ProxyVotes:     make(map[string]*ProxyVote),
		Metrics:        make(map[string]*DelegationMetrics),
		Config:         config,
		encryptionKey:  encryptionKey,
		stopChan:       make(chan struct{}),
	}
}

// Start begins the Delegated Governance system operations
func (dg *DelegatedGovernance) Start() error {
	dg.mu.Lock()
	defer dg.mu.Unlock()

	if dg.running {
		return fmt.Errorf("Delegated Governance system is already running")
	}

	dg.running = true

	// Start background goroutines
	go dg.delegationManagementLoop()
	go dg.metricsUpdateLoop()
	go dg.cleanupLoop()

	return nil
}

// Stop halts all Delegated Governance system operations
func (dg *DelegatedGovernance) Stop() error {
	dg.mu.Lock()
	defer dg.mu.Unlock()

	if !dg.running {
		return fmt.Errorf("Delegated Governance system is not running")
	}

	close(dg.stopChan)
	dg.running = false

	return nil
}

// RegisterDelegator registers a new delegator in the system
func (dg *DelegatedGovernance) RegisterDelegator(
	address string,
	totalPower *big.Int,
	metadata map[string]interface{},
) (*Delegator, error) {
	dg.mu.Lock()
	defer dg.mu.Unlock()

	// Check limits
	if uint64(len(dg.Delegators)) >= dg.Config.MaxDelegators {
		return nil, fmt.Errorf("delegator limit reached")
	}

	// Check if delegator already exists
	for _, delegator := range dg.Delegators {
		if delegator.Address == address {
			return nil, fmt.Errorf("delegator already exists: %s", address)
		}
	}

	// Validate parameters
	if address == "" {
		return nil, fmt.Errorf("address cannot be empty")
	}
	if totalPower.Cmp(big.NewInt(0)) <= 0 {
		return nil, fmt.Errorf("total power must be positive")
	}

	delegator := &Delegator{
		ID:             generateDelegatorID(),
		Address:        address,
		TotalPower:     totalPower,
		DelegatedPower: big.NewInt(0),
		RetainedPower:  totalPower, // Initially all power is retained
		Reputation:     big.NewInt(100), // Start with base reputation
		CreatedAt:      time.Now(),
		Metadata:       metadata,
	}

	dg.Delegators[delegator.ID] = delegator
	return delegator, nil
}

// RegisterDelegate registers a new delegate in the system
func (dg *DelegatedGovernance) RegisterDelegate(
	address string,
	metadata map[string]interface{},
) (*Delegate, error) {
	dg.mu.Lock()
	defer dg.mu.Unlock()

	// Check limits
	if uint64(len(dg.Delegates)) >= dg.Config.MaxDelegates {
		return nil, fmt.Errorf("delegate limit reached")
	}

	// Check if delegate already exists
	for _, delegate := range dg.Delegates {
		if delegate.Address == address {
			return nil, fmt.Errorf("delegate already exists: %s", address)
		}
	}

	// Validate parameters
	if address == "" {
		return nil, fmt.Errorf("address cannot be empty")
	}

	delegate := &Delegate{
		ID:               generateDelegateID(),
		Address:          address,
		TotalDelegated:   big.NewInt(0),
		ActiveDelegations: 0,
		Reputation:       big.NewInt(100), // Start with base reputation
		Performance:      big.NewInt(0),
		CreatedAt:        time.Now(),
		Metadata:         metadata,
	}

	dg.Delegates[delegate.ID] = delegate
	return delegate, nil
}

// CreateDelegation creates a new delegation relationship
func (dg *DelegatedGovernance) CreateDelegation(
	delegatorAddress string,
	delegateAddress string,
	delegationType DelegationType,
	power *big.Int,
	conditions []DelegationCondition,
	expiresAt time.Time,
	metadata map[string]interface{},
) (*Delegation, error) {
	dg.mu.Lock()
	defer dg.mu.Unlock()

	// Check limits
	if uint64(len(dg.Delegations)) >= dg.Config.MaxDelegations {
		return nil, fmt.Errorf("delegation limit reached")
	}

	// Find delegator and delegate
	var delegator *Delegator
	var delegate *Delegate

	for _, d := range dg.Delegators {
		if d.Address == delegatorAddress {
			delegator = d
			break
		}
	}

	for _, d := range dg.Delegates {
		if d.Address == delegateAddress {
			delegate = d
			break
		}
	}

	if delegator == nil {
		return nil, fmt.Errorf("delegator not found: %s", delegatorAddress)
	}
	if delegate == nil {
		return nil, fmt.Errorf("delegate not found: %s", delegateAddress)
	}

	// Validate delegation power
	if power.Cmp(dg.Config.MinDelegationPower) < 0 {
		return nil, fmt.Errorf("delegation power below minimum: %s < %s", power.String(), dg.Config.MinDelegationPower.String())
	}
	if power.Cmp(dg.Config.MaxDelegationPower) > 0 {
		return nil, fmt.Errorf("delegation power above maximum: %s > %s", power.String(), dg.Config.MaxDelegationPower.String())
	}

	// Check if delegator has enough power
	availablePower := new(big.Int).Sub(delegator.RetainedPower, delegator.DelegatedPower)
	if power.Cmp(availablePower) > 0 {
		return nil, fmt.Errorf("insufficient power: %s < %s", availablePower.String(), power.String())
	}

	// Generate ZK proof for the delegation
	zkProver := security.NewZKProver(dg.Config.ZKProofType)
	statement := []byte(fmt.Sprintf("%s:%s:%s:%s", delegatorAddress, delegateAddress, delegationType.String(), power.String()))
	witness := []byte(fmt.Sprintf("%s", delegatorAddress))
	zkProof, err := zkProver.GenerateProof(statement, witness)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ZK proof: %v", err)
	}

	delegation := &Delegation{
		ID:          generateDelegationID(),
		DelegatorID: delegator.ID,
		DelegateID:  delegate.ID,
		Type:        delegationType,
		Power:       power,
		Conditions:  conditions,
		Status:      DelegationStatusActive,
		CreatedAt:   time.Now(),
		ExpiresAt:   expiresAt,
		LastUsed:    time.Now(),
		ZKProof:     zkProof,
		Metadata:    metadata,
	}

	dg.Delegations[delegation.ID] = delegation

	// Update delegator and delegate
	delegator.DelegatedPower.Add(delegator.DelegatedPower, power)
	delegate.TotalDelegated.Add(delegate.TotalDelegated, power)
	delegate.ActiveDelegations++

	// Initialize metrics
	dg.Metrics[delegation.ID] = &DelegationMetrics{
		DelegationID:    delegation.ID,
		TotalVotes:      0,
		SuccessfulVotes: 0,
		VoteAccuracy:    0.0,
		Performance:     big.NewInt(0),
		LastUpdated:     time.Now(),
		Metadata:        make(map[string]interface{}),
	}

	return delegation, nil
}

// CastProxyVote casts a vote on behalf of delegators
func (dg *DelegatedGovernance) CastProxyVote(
	delegationID string,
	proposalID string,
	voteType string,
	votePower *big.Int,
	metadata map[string]interface{},
) (*ProxyVote, error) {
	dg.mu.Lock()
	defer dg.mu.Unlock()

	// Check limits
	if uint64(len(dg.ProxyVotes)) >= dg.Config.MaxProxyVotes {
		return nil, fmt.Errorf("proxy vote limit reached")
	}

	// Find delegation
	delegation, exists := dg.Delegations[delegationID]
	if !exists {
		return nil, fmt.Errorf("delegation not found: %s", delegationID)
	}

	if delegation.Status != DelegationStatusActive {
		return nil, fmt.Errorf("delegation is not active: %s", delegation.Status.String())
	}

	// Check if delegation has expired
	if time.Now().After(delegation.ExpiresAt) {
		delegation.Status = DelegationStatusExpired
		return nil, fmt.Errorf("delegation has expired")
	}

	// Validate vote power
	if votePower.Cmp(big.NewInt(0)) <= 0 {
		return nil, fmt.Errorf("vote power must be positive")
	}
	if votePower.Cmp(delegation.Power) > 0 {
		return nil, fmt.Errorf("vote power exceeds delegation power: %s > %s", votePower.String(), delegation.Power.String())
	}

	// Check conditions
	if !dg.evaluateConditions(delegation.Conditions) {
		return nil, fmt.Errorf("delegation conditions not met")
	}

	// Generate ZK proof for the proxy vote
	zkProver := security.NewZKProver(dg.Config.ZKProofType)
	statement := []byte(fmt.Sprintf("%s:%s:%s:%s", delegationID, proposalID, voteType, votePower.String()))
	witness := []byte(fmt.Sprintf("%s", delegationID))
	zkProof, err := zkProver.GenerateProof(statement, witness)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ZK proof: %v", err)
	}

	proxyVote := &ProxyVote{
		ID:             generateProxyVoteID(),
		DelegationID:   delegationID,
		ProposalID:     proposalID,
		VoteType:       voteType,
		VotePower:      votePower,
		DelegatorCount: 1, // For now, assume single delegator
		Timestamp:      time.Now(),
		ZKProof:        zkProof,
		Metadata:       metadata,
	}

	dg.ProxyVotes[proxyVote.ID] = proxyVote

	// Update delegation last used time
	delegation.LastUsed = time.Now()

	// Update metrics
	if metrics, exists := dg.Metrics[delegationID]; exists {
		metrics.TotalVotes++
		metrics.LastUpdated = time.Now()
	}

	return proxyVote, nil
}

// RevokeDelegation revokes an active delegation
func (dg *DelegatedGovernance) RevokeDelegation(delegationID string) error {
	dg.mu.Lock()
	defer dg.mu.Unlock()

	delegation, exists := dg.Delegations[delegationID]
	if !exists {
		return fmt.Errorf("delegation not found: %s", delegationID)
	}

	if delegation.Status != DelegationStatusActive {
		return fmt.Errorf("delegation is not active: %s", delegation.Status.String())
	}

	// Update status
	delegation.Status = DelegationStatusRevoked

	// Find delegator and delegate
	delegator, exists := dg.Delegators[delegation.DelegatorID]
	if exists {
		delegator.DelegatedPower.Sub(delegator.DelegatedPower, delegation.Power)
	}

	delegate, exists := dg.Delegates[delegation.DelegateID]
	if exists {
		delegate.TotalDelegated.Sub(delegate.TotalDelegated, delegation.Power)
		if delegate.ActiveDelegations > 0 {
			delegate.ActiveDelegations--
		}
	}

	return nil
}

// PauseDelegation pauses an active delegation
func (dg *DelegatedGovernance) PauseDelegation(delegationID string) error {
	dg.mu.Lock()
	defer dg.mu.Unlock()

	delegation, exists := dg.Delegations[delegationID]
	if !exists {
		return fmt.Errorf("delegation not found: %s", delegationID)
	}

	if delegation.Status != DelegationStatusActive {
		return fmt.Errorf("delegation is not active: %s", delegation.Status.String())
	}

	delegation.Status = DelegationStatusPaused
	return nil
}

// ResumeDelegation resumes a paused delegation
func (dg *DelegatedGovernance) ResumeDelegation(delegationID string) error {
	dg.mu.Lock()
	defer dg.mu.Unlock()

	delegation, exists := dg.Delegations[delegationID]
	if !exists {
		return fmt.Errorf("delegation not found: %s", delegationID)
	}

	if delegation.Status != DelegationStatusPaused {
		return fmt.Errorf("delegation is not paused: %s", delegation.Status.String())
	}

	delegation.Status = DelegationStatusActive
	return nil
}

// GetDelegator retrieves a delegator by address
func (dg *DelegatedGovernance) GetDelegator(address string) (*Delegator, error) {
	dg.mu.RLock()
	defer dg.mu.RUnlock()

	for _, delegator := range dg.Delegators {
		if delegator.Address == address {
			return dg.copyDelegator(delegator), nil
		}
	}

	return nil, fmt.Errorf("delegator not found: %s", address)
}

// GetDelegate retrieves a delegate by address
func (dg *DelegatedGovernance) GetDelegate(address string) (*Delegate, error) {
	dg.mu.RLock()
	defer dg.mu.RUnlock()

	for _, delegate := range dg.Delegates {
		if delegate.Address == address {
			return dg.copyDelegate(delegate), nil
		}
	}

	return nil, fmt.Errorf("delegate not found: %s", address)
}

// GetDelegation retrieves a delegation by ID
func (dg *DelegatedGovernance) GetDelegation(delegationID string) (*Delegation, error) {
	dg.mu.RLock()
	defer dg.mu.RUnlock()

	delegation, exists := dg.Delegations[delegationID]
	if !exists {
		return nil, fmt.Errorf("delegation not found: %s", delegationID)
	}

	return dg.copyDelegation(delegation), nil
}

// GetDelegationsByDelegator retrieves all delegations for a delegator
func (dg *DelegatedGovernance) GetDelegationsByDelegator(delegatorAddress string) ([]*Delegation, error) {
	dg.mu.RLock()
	defer dg.mu.RUnlock()

	var delegations []*Delegation
	for _, delegation := range dg.Delegations {
		if delegator, exists := dg.Delegators[delegation.DelegatorID]; exists && delegator.Address == delegatorAddress {
			delegations = append(delegations, dg.copyDelegation(delegation))
		}
	}

	return delegations, nil
}

// GetDelegationsByDelegate retrieves all delegations for a delegate
func (dg *DelegatedGovernance) GetDelegationsByDelegate(delegateAddress string) ([]*Delegation, error) {
	dg.mu.RLock()
	defer dg.mu.RUnlock()

	var delegations []*Delegation
	for _, delegation := range dg.Delegations {
		if delegate, exists := dg.Delegates[delegation.DelegateID]; exists && delegate.Address == delegateAddress {
			delegations = append(delegations, dg.copyDelegation(delegation))
		}
	}

	return delegations, nil
}

// Background loops
func (dg *DelegatedGovernance) delegationManagementLoop() {
	ticker := time.NewTicker(time.Minute * 5)
	defer ticker.Stop()

	for {
		select {
		case <-dg.stopChan:
			return
		case <-ticker.C:
			dg.manageDelegations()
		}
	}
}

func (dg *DelegatedGovernance) metricsUpdateLoop() {
	ticker := time.NewTicker(time.Minute * 10)
	defer ticker.Stop()

	for {
		select {
		case <-dg.stopChan:
			return
		case <-ticker.C:
			dg.updateMetrics()
		}
	}
}

func (dg *DelegatedGovernance) cleanupLoop() {
	ticker := time.NewTicker(dg.Config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-dg.stopChan:
			return
		case <-ticker.C:
			dg.cleanupOldData()
		}
	}
}

// Helper functions for background loops
func (dg *DelegatedGovernance) manageDelegations() {
	dg.mu.Lock()
	defer dg.mu.Unlock()

	now := time.Now()

	for _, delegation := range dg.Delegations {
		if delegation.Status == DelegationStatusActive && now.After(delegation.ExpiresAt) {
			delegation.Status = DelegationStatusExpired
		}
	}
}

func (dg *DelegatedGovernance) updateMetrics() {
	dg.mu.Lock()
	defer dg.mu.Unlock()

	for delegationID, metrics := range dg.Metrics {
		if delegation, exists := dg.Delegations[delegationID]; exists {
			// Update performance based on delegation status and activity
			if delegation.Status == DelegationStatusActive {
				metrics.Performance.Add(metrics.Performance, big.NewInt(1))
			}
			metrics.LastUpdated = time.Now()
		}
	}
}

func (dg *DelegatedGovernance) cleanupOldData() {
	dg.mu.Lock()
	defer dg.mu.Unlock()

	cutoffTime := time.Now().Add(-dg.Config.DelegationTimeout)

	// Clean up expired delegations
	for id, delegation := range dg.Delegations {
		if delegation.Status == DelegationStatusExpired && delegation.ExpiresAt.Before(cutoffTime) {
			delete(dg.Delegations, id)
			delete(dg.Metrics, id)
		}
	}

	// Clean up old proxy votes
	for id, vote := range dg.ProxyVotes {
		if vote.Timestamp.Before(cutoffTime) {
			delete(dg.ProxyVotes, id)
		}
	}
}

// Helper functions
func (dg *DelegatedGovernance) evaluateConditions(conditions []DelegationCondition) bool {
	if len(conditions) == 0 {
		return true // No conditions means always allow
	}

	for _, condition := range conditions {
		if !condition.Active {
			continue
		}

		if !dg.evaluateCondition(condition) {
			return false
		}
	}

	return true
}

func (dg *DelegatedGovernance) evaluateCondition(condition DelegationCondition) bool {
	// Simple condition evaluation - in real implementation, this would be more sophisticated
	switch condition.Type {
	case ConditionTypeTimeBased:
		// Time-based conditions would check current time against condition value
		return true
	case ConditionTypePerformanceBased:
		// Performance-based conditions would check delegate performance
		return true
	case ConditionTypeReputationBased:
		// Reputation-based conditions would check delegate reputation
		return true
	case ConditionTypeVoteBased:
		// Vote-based conditions would check voting history
		return true
	default:
		return true
	}
}

// Deep copy functions
func (dg *DelegatedGovernance) copyDelegator(delegator *Delegator) *Delegator {
	if delegator == nil {
		return nil
	}

	copied := *delegator
	copied.Metadata = dg.copyMap(delegator.Metadata)
	return &copied
}

func (dg *DelegatedGovernance) copyDelegate(delegate *Delegate) *Delegate {
	if delegate == nil {
		return nil
	}

	copied := *delegate
	copied.Metadata = dg.copyMap(delegate.Metadata)
	return &copied
}

func (dg *DelegatedGovernance) copyDelegation(delegation *Delegation) *Delegation {
	if delegation == nil {
		return nil
	}

	copied := *delegation
	copied.Conditions = make([]DelegationCondition, len(delegation.Conditions))
	copy(copied.Conditions, delegation.Conditions)
	
	if delegation.ZKProof != nil {
		copied.ZKProof = &security.ZKProof{
			Type:            delegation.ZKProof.Type,
			Proof:           make([]byte, len(delegation.ZKProof.Proof)),
			PublicInputs:    make([]byte, len(delegation.ZKProof.PublicInputs)),
			VerificationKey: make([]byte, len(delegation.ZKProof.VerificationKey)),
			Timestamp:       delegation.ZKProof.Timestamp,
		}
		copy(copied.ZKProof.Proof, delegation.ZKProof.Proof)
		copy(copied.ZKProof.PublicInputs, delegation.ZKProof.PublicInputs)
		copy(copied.ZKProof.VerificationKey, delegation.ZKProof.VerificationKey)
	}

	copied.Metadata = dg.copyMap(delegation.Metadata)
	return &copied
}

func (dg *DelegatedGovernance) copyMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		return nil
	}

	copied := make(map[string]interface{})
	for k, v := range m {
		copied[k] = v
	}
	return copied
}

// ID generation functions
func generateDelegatorID() string {
	return fmt.Sprintf("delegator_%d", time.Now().UnixNano())
}

func generateDelegateID() string {
	return fmt.Sprintf("delegate_%d", time.Now().UnixNano())
}

func generateDelegationID() string {
	return fmt.Sprintf("delegation_%d", time.Now().UnixNano())
}

func generateProxyVoteID() string {
	return fmt.Sprintf("proxy_vote_%d", time.Now().UnixNano())
}

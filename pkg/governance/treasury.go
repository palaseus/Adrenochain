package governance

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"
	"time"
)

// TreasuryManager handles DAO treasury operations
type TreasuryManager struct {
	balances        map[string]*big.Int
	transactions    []*TreasuryTransaction
	proposals       map[string]*TreasuryProposal
	multisig        *MultisigWallet
	mutex           sync.RWMutex
	maxTransactionAmount *big.Int
	dailyLimit      *big.Int
	dailyUsed       *big.Int
	lastReset       time.Time
}

// TreasuryTransaction represents a treasury transaction
type TreasuryTransaction struct {
	ID              string                `json:"id"`
	Type            TreasuryTransactionType `json:"type"`
	Amount          *big.Int              `json:"amount"`
	Asset           string                `json:"asset"`
	From            string                `json:"from"`
	To              string                `json:"to"`
	Description     string                `json:"description"`
	ProposalID      string                `json:"proposal_id,omitempty"`
	Status          TransactionStatus     `json:"status"`
	ExecutedBy      string                `json:"executed_by"`
	CreatedAt       time.Time             `json:"created_at"`
	ExecutedAt      *time.Time            `json:"executed_at,omitempty"`
	GasUsed         *big.Int              `json:"gas_used,omitempty"`
	TxHash          string                `json:"tx_hash,omitempty"`
}

// TreasuryTransactionType represents the type of treasury transaction
type TreasuryTransactionType string

const (
	TreasuryTransactionTypeTransfer    TreasuryTransactionType = "transfer"
	TreasuryTransactionTypeWithdrawal  TreasuryTransactionType = "withdrawal"
	TreasuryTransactionTypeDeposit     TreasuryTransactionType = "deposit"
	TreasuryTransactionTypeInvestment  TreasuryTransactionType = "investment"
	TreasuryTransactionTypeReward      TreasuryTransactionType = "reward"
)

// TransactionStatus represents the status of a treasury transaction
type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusApproved  TransactionStatus = "approved"
	TransactionStatusExecuted  TransactionStatus = "executed"
	TransactionStatusRejected  TransactionStatus = "rejected"
	TransactionStatusFailed    TransactionStatus = "failed"
)

// TreasuryProposal represents a treasury proposal
type TreasuryProposal struct {
	ID              string                `json:"id"`
	Title           string                `json:"title"`
	Description     string                `json:"description"`
	Proposer        string                `json:"proposer"`
	Amount          *big.Int              `json:"amount"`
	Asset           string                `json:"asset"`
	Recipient       string                `json:"recipient"`
	Purpose         string                `json:"purpose"`
	Status          TreasuryProposalStatus `json:"status"`
	VotesFor        *big.Int              `json:"votes_for"`
	VotesAgainst    *big.Int              `json:"votes_against"`
	CreatedAt       time.Time             `json:"created_at"`
	UpdatedAt       time.Time             `json:"updated_at"`
	ExecutedAt      *time.Time            `json:"executed_at,omitempty"`
}

// TreasuryProposalStatus represents the status of a treasury proposal
type TreasuryProposalStatus string

const (
	TreasuryProposalStatusDraft      TreasuryProposalStatus = "draft"
	TreasuryProposalStatusActive     TreasuryProposalStatus = "active"
	TreasuryProposalStatusPassed     TreasuryProposalStatus = "passed"
	TreasuryProposalStatusRejected   TreasuryProposalStatus = "rejected"
	TreasuryProposalStatusExecuted   TreasuryProposalStatus = "executed"
	TreasuryProposalStatusCancelled  TreasuryProposalStatus = "cancelled"
)

// MultisigWallet represents a multisignature wallet for treasury operations
type MultisigWallet struct {
	addresses       []string
	requiredSignatures int
	signatures      map[string]map[string]*Signature
	mutex           sync.RWMutex
}

// Signature represents a signature on a transaction
type Signature struct {
	Signer         string    `json:"signer"`
	TransactionID  string    `json:"transaction_id"`
	Signature      []byte    `json:"signature"`
	Timestamp      time.Time `json:"timestamp"`
}

// NewTreasuryManager creates a new treasury manager
func NewTreasuryManager(
	maxTransactionAmount *big.Int,
	dailyLimit *big.Int,
	multisigAddresses []string,
	requiredSignatures int,
) *TreasuryManager {
	return &TreasuryManager{
		balances:             make(map[string]*big.Int),
		transactions:         make([]*TreasuryTransaction, 0),
		proposals:            make(map[string]*TreasuryProposal, 0),
		multisig:             NewMultisigWallet(multisigAddresses, requiredSignatures),
		maxTransactionAmount: maxTransactionAmount,
		dailyLimit:           dailyLimit,
		dailyUsed:            big.NewInt(0),
		lastReset:            time.Now(),
	}
}

// NewMultisigWallet creates a new multisignature wallet
func NewMultisigWallet(addresses []string, requiredSignatures int) *MultisigWallet {
	return &MultisigWallet{
		addresses:          addresses,
		requiredSignatures: requiredSignatures,
		signatures:         make(map[string]map[string]*Signature),
	}
}

// GetBalance returns the balance of an asset
func (tm *TreasuryManager) GetBalance(asset string) *big.Int {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	balance, exists := tm.balances[asset]
	if !exists {
		return big.NewInt(0)
	}

	return balance
}

// SetBalance sets the balance of an asset
func (tm *TreasuryManager) SetBalance(asset string, amount *big.Int) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	tm.balances[asset] = amount
}

// CreateTreasuryProposal creates a new treasury proposal
func (tm *TreasuryManager) CreateTreasuryProposal(
	title string,
	description string,
	proposer string,
	amount *big.Int,
	asset string,
	recipient string,
	purpose string,
) (*TreasuryProposal, error) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	// Validate amount
	if amount.Cmp(big.NewInt(0)) <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}

	if amount.Cmp(tm.maxTransactionAmount) > 0 {
		return nil, fmt.Errorf("amount exceeds maximum transaction amount")
	}

	// Check if treasury has sufficient balance
	currentBalance := tm.balances[asset]
	if currentBalance == nil || currentBalance.Cmp(amount) < 0 {
		return nil, fmt.Errorf("insufficient treasury balance for asset %s", asset)
	}

	proposal := &TreasuryProposal{
		ID:           tm.generateProposalID(),
		Title:        title,
		Description:  description,
		Proposer:     proposer,
		Amount:       amount,
		Asset:        asset,
		Recipient:    recipient,
		Purpose:      purpose,
		Status:       TreasuryProposalStatusDraft,
		VotesFor:     big.NewInt(0),
		VotesAgainst: big.NewInt(0),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	tm.proposals[proposal.ID] = proposal

	return proposal, nil
}

// ActivateTreasuryProposal activates a treasury proposal for voting
func (tm *TreasuryManager) ActivateTreasuryProposal(proposalID string) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	proposal, exists := tm.proposals[proposalID]
	if !exists {
		return fmt.Errorf("proposal not found: %s", proposalID)
	}

	if proposal.Status != TreasuryProposalStatusDraft {
		return fmt.Errorf("proposal status is %s, expected draft", proposal.Status)
	}

	proposal.Status = TreasuryProposalStatusActive
	proposal.UpdatedAt = time.Now()

	return nil
}

// VoteOnTreasuryProposal votes on a treasury proposal
func (tm *TreasuryManager) VoteOnTreasuryProposal(
	proposalID string,
	voter string,
	voteChoice VoteChoice,
	votingPower *big.Int,
) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	proposal, exists := tm.proposals[proposalID]
	if !exists {
		return fmt.Errorf("proposal not found: %s", proposalID)
	}

	if proposal.Status != TreasuryProposalStatusActive {
		return fmt.Errorf("proposal is not active for voting")
	}

	// Update vote counts
	switch voteChoice {
	case VoteChoiceFor:
		proposal.VotesFor.Add(proposal.VotesFor, votingPower)
	case VoteChoiceAgainst:
		proposal.VotesAgainst.Add(proposal.VotesAgainst, votingPower)
	}

	proposal.UpdatedAt = time.Now()

	return nil
}

// FinalizeTreasuryProposal finalizes a treasury proposal
func (tm *TreasuryManager) FinalizeTreasuryProposal(proposalID string) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	proposal, exists := tm.proposals[proposalID]
	if !exists {
		return fmt.Errorf("proposal not found: %s", proposalID)
	}

	if proposal.Status != TreasuryProposalStatusActive {
		return fmt.Errorf("proposal is not active")
	}

	// Determine if proposal passed (simple majority)
	if proposal.VotesFor.Cmp(proposal.VotesAgainst) > 0 {
		proposal.Status = TreasuryProposalStatusPassed
	} else {
		proposal.Status = TreasuryProposalStatusRejected
	}

	proposal.UpdatedAt = time.Now()

	return nil
}

// ExecuteTreasuryProposal executes a passed treasury proposal
func (tm *TreasuryManager) ExecuteTreasuryProposal(
	proposalID string,
	executor string,
) (*TreasuryTransaction, error) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	proposal, exists := tm.proposals[proposalID]
	if !exists {
		return nil, fmt.Errorf("proposal not found: %s", proposalID)
	}

	if proposal.Status != TreasuryProposalStatusPassed {
		return nil, fmt.Errorf("proposal status is %s, expected passed", proposal.Status)
	}

	// Check daily limit
	if err := tm.checkDailyLimit(proposal.Amount); err != nil {
		return nil, err
	}

	// Create and execute transaction
	transaction := &TreasuryTransaction{
		ID:          tm.generateTransactionID(),
		Type:        TreasuryTransactionTypeTransfer,
		Amount:      proposal.Amount,
		Asset:       proposal.Asset,
		From:        "treasury",
		To:          proposal.Recipient,
		Description: proposal.Purpose,
		ProposalID:  proposalID,
		Status:      TransactionStatusPending,
		ExecutedBy:  executor,
		CreatedAt:   time.Now(),
	}

	// Execute the transaction
	if err := tm.executeTransaction(transaction); err != nil {
		return nil, err
	}

	// Update proposal status
	proposal.Status = TreasuryProposalStatusExecuted
	proposal.ExecutedAt = &time.Time{}
	*proposal.ExecutedAt = time.Now()
	proposal.UpdatedAt = time.Now()

	// Update daily usage
	tm.dailyUsed.Add(tm.dailyUsed, proposal.Amount)

	return transaction, nil
}

// CreateDirectTransaction creates a direct treasury transaction
func (tm *TreasuryManager) CreateDirectTransaction(
	transactionType TreasuryTransactionType,
	amount *big.Int,
	asset string,
	to string,
	description string,
	executor string,
) (*TreasuryTransaction, error) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	// Validate amount
	if amount.Cmp(big.NewInt(0)) <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}

	if amount.Cmp(tm.maxTransactionAmount) > 0 {
		return nil, fmt.Errorf("amount exceeds maximum transaction amount")
	}

	// Check daily limit
	if err := tm.checkDailyLimit(amount); err != nil {
		return nil, err
	}

	// Create transaction
	transaction := &TreasuryTransaction{
		ID:          tm.generateTransactionID(),
		Type:        transactionType,
		Amount:      amount,
		Asset:       asset,
		From:        "treasury",
		To:          to,
		Description: description,
		Status:      TransactionStatusPending,
		ExecutedBy:  executor,
		CreatedAt:   time.Now(),
	}

	// Execute the transaction
	if err := tm.executeTransaction(transaction); err != nil {
		return nil, err
	}

	// Update daily usage
	tm.dailyUsed.Add(tm.dailyUsed, amount)

	return transaction, nil
}

// executeTransaction executes a treasury transaction
func (tm *TreasuryManager) executeTransaction(transaction *TreasuryTransaction) error {
	// Check if treasury has sufficient balance
	currentBalance := tm.balances[transaction.Asset]
	if currentBalance == nil || currentBalance.Cmp(transaction.Amount) < 0 {
		return fmt.Errorf("insufficient treasury balance for asset %s", transaction.Asset)
	}

	// Update balances
	tm.balances[transaction.Asset] = new(big.Int).Sub(currentBalance, transaction.Amount)

	// Update transaction status
	transaction.Status = TransactionStatusExecuted
	transaction.ExecutedAt = &time.Time{}
	*transaction.ExecutedAt = time.Now()

	// Add to transaction history
	tm.transactions = append(tm.transactions, transaction)

	return nil
}

// checkDailyLimit checks if a transaction would exceed daily limits
func (tm *TreasuryManager) checkDailyLimit(amount *big.Int) error {
	// Reset daily limit if it's a new day
	now := time.Now()
	if now.Sub(tm.lastReset) >= 24*time.Hour {
		tm.dailyUsed = big.NewInt(0)
		tm.lastReset = now
	}

	// Check if transaction would exceed daily limit
	if tm.dailyUsed.Add(tm.dailyUsed, amount).Cmp(tm.dailyLimit) > 0 {
		return fmt.Errorf("transaction would exceed daily limit")
	}

	return nil
}

// GetTreasuryProposal returns a treasury proposal by ID
func (tm *TreasuryManager) GetTreasuryProposal(proposalID string) (*TreasuryProposal, error) {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	proposal, exists := tm.proposals[proposalID]
	if !exists {
		return nil, fmt.Errorf("proposal not found: %s", proposalID)
	}

	return proposal, nil
}

// GetTreasuryProposalsByStatus returns treasury proposals by status
func (tm *TreasuryManager) GetTreasuryProposalsByStatus(status TreasuryProposalStatus) []*TreasuryProposal {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	var result []*TreasuryProposal
	for _, proposal := range tm.proposals {
		if proposal.Status == status {
			result = append(result, proposal)
		}
	}

	return result
}

// GetTreasuryTransactions returns treasury transactions
func (tm *TreasuryManager) GetTreasuryTransactions(limit int) []*TreasuryTransaction {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	if limit <= 0 || limit > len(tm.transactions) {
		limit = len(tm.transactions)
	}

	result := make([]*TreasuryTransaction, limit)
	copy(result, tm.transactions[len(tm.transactions)-limit:])

	return result
}

// GetTreasuryStats returns treasury statistics
func (tm *TreasuryManager) GetTreasuryStats() map[string]interface{} {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	stats := make(map[string]interface{})
	stats["total_proposals"] = len(tm.proposals)
	stats["active_proposals"] = len(tm.GetTreasuryProposalsByStatus(TreasuryProposalStatusActive))
	stats["passed_proposals"] = len(tm.GetTreasuryProposalsByStatus(TreasuryProposalStatusPassed))
	stats["total_transactions"] = len(tm.transactions)
	stats["daily_limit"] = tm.dailyLimit.String()
	stats["daily_used"] = tm.dailyUsed.String()
	stats["multisig_signers"] = len(tm.multisig.addresses)
	stats["required_signatures"] = tm.multisig.requiredSignatures

	// Add asset balances
	for asset, balance := range tm.balances {
		stats["balance_"+asset] = balance.String()
	}

	return stats
}

// Multisig Wallet Methods

// SignTransaction signs a transaction with a multisig wallet
func (mw *MultisigWallet) SignTransaction(transactionID string, signer string, signature []byte) error {
	mw.mutex.Lock()
	defer mw.mutex.Unlock()

	// Validate signer
	validSigner := false
	for _, addr := range mw.addresses {
		if addr == signer {
			validSigner = true
			break
		}
	}

	if !validSigner {
		return fmt.Errorf("invalid signer: %s", signer)
	}

	// Initialize signatures map for transaction if needed
	if mw.signatures[transactionID] == nil {
		mw.signatures[transactionID] = make(map[string]*Signature)
	}

	// Add signature
	mw.signatures[transactionID][signer] = &Signature{
		Signer:        signer,
		TransactionID: transactionID,
		Signature:     signature,
		Timestamp:     time.Now(),
	}

	return nil
}

// HasEnoughSignatures checks if a transaction has enough signatures
func (mw *MultisigWallet) HasEnoughSignatures(transactionID string) bool {
	mw.mutex.RLock()
	defer mw.mutex.RUnlock()

	signatures, exists := mw.signatures[transactionID]
	if !exists {
		return false
	}

	return len(signatures) >= mw.requiredSignatures
}

// GetSignatures returns all signatures for a transaction
func (mw *MultisigWallet) GetSignatures(transactionID string) []*Signature {
	mw.mutex.RLock()
	defer mw.mutex.RUnlock()

	signatures, exists := mw.signatures[transactionID]
	if !exists {
		return nil
	}

	var result []*Signature
	for _, signature := range signatures {
		result = append(result, signature)
	}

	return result
}

// Helper methods

// generateProposalID generates a unique proposal ID
func (tm *TreasuryManager) generateProposalID() string {
	data := fmt.Sprintf("treasury_proposal_%d", time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:8])
}

// generateTransactionID generates a unique transaction ID
func (tm *TreasuryManager) generateTransactionID() string {
	data := fmt.Sprintf("treasury_tx_%d", time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:8])
}

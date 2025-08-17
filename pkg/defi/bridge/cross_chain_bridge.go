package bridge

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math/big"
	"sync"
	"time"
)

// BridgeStatus represents the status of a bridge
type BridgeStatus int

const (
	BridgeInactive BridgeStatus = iota
	BridgeActive
	BridgePaused
	BridgeEmergency
)

// Chain represents a blockchain network
type Chain struct {
	ID          string
	Name        string
	Symbol      string
	ChainID     uint64
	RPCURL      string
	ExplorerURL string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Asset represents a cross-chain asset
type Asset struct {
	ID           string
	Symbol       string
	Name         string
	Decimals     uint8
	TotalSupply  *big.Int
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// BridgeAsset represents an asset on a specific chain
type BridgeAsset struct {
	ID           string
	AssetID      string
	ChainID      string
	ContractAddr string
	Balance      *big.Int
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Transfer represents a cross-chain transfer
type Transfer struct {
	ID              string
	SourceChainID   string
	TargetChainID   string
	AssetID         string
	Amount          *big.Int
	Sender          string
	Recipient       string
	Status          TransferStatus
	SourceTxHash    string
	TargetTxHash    string
	BridgeFee       *big.Int
	GasFee          *big.Int
	CreatedAt       time.Time
	UpdatedAt       time.Time
	CompletedAt     *time.Time
}

// TransferStatus represents the status of a transfer
type TransferStatus int

const (
	TransferPending TransferStatus = iota
	TransferConfirmed
	TransferProcessing
	TransferCompleted
	TransferFailed
	TransferCancelled
)

// Validator represents a bridge validator
type Validator struct {
	ID          string
	Address     string
	ChainID     string
	StakeAmount *big.Int
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Bridge represents a cross-chain bridge
type Bridge struct {
	ID              string
	Name            string
	Description     string
	SourceChainID   string
	TargetChainID   string
	Status          BridgeStatus
	MinTransfer     *big.Int
	MaxTransfer     *big.Int
	BridgeFee       *big.Int
	Validators      []string
	IsActive        bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// BridgeManager manages cross-chain bridges
type BridgeManager struct {
	Chains      map[string]*Chain
	Assets      map[string]*Asset
	BridgeAssets map[string]*BridgeAsset
	Bridges     map[string]*Bridge
	Transfers   map[string]*Transfer
	Validators  map[string]*Validator
	mu          sync.RWMutex
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewBridgeManager creates a new bridge manager
func NewBridgeManager() *BridgeManager {
	now := time.Now()
	
	return &BridgeManager{
		Chains:       make(map[string]*Chain),
		Assets:       make(map[string]*Asset),
		BridgeAssets: make(map[string]*BridgeAsset),
		Bridges:      make(map[string]*Bridge),
		Transfers:    make(map[string]*Transfer),
		Validators:   make(map[string]*Validator),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// NewChain creates a new blockchain network
func NewChain(id, name, symbol string, chainID uint64, rpcURL, explorerURL string) (*Chain, error) {
	if id == "" {
		return nil, errors.New("chain ID cannot be empty")
	}
	if name == "" {
		return nil, errors.New("chain name cannot be empty")
	}
	if symbol == "" {
		return nil, errors.New("chain symbol cannot be empty")
	}
	if rpcURL == "" {
		return nil, errors.New("RPC URL cannot be empty")
	}
	
	now := time.Now()
	
	return &Chain{
		ID:          id,
		Name:        name,
		Symbol:      symbol,
		ChainID:     chainID,
		RPCURL:      rpcURL,
		ExplorerURL: explorerURL,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// NewAsset creates a new cross-chain asset
func NewAsset(id, symbol, name string, decimals uint8, totalSupply *big.Int) (*Asset, error) {
	if id == "" {
		return nil, errors.New("asset ID cannot be empty")
	}
	if symbol == "" {
		return nil, errors.New("asset symbol cannot be empty")
	}
	if name == "" {
		return nil, errors.New("asset name cannot be empty")
	}
	if totalSupply == nil || totalSupply.Sign() < 0 {
		return nil, errors.New("total supply must be non-negative")
	}
	
	now := time.Now()
	
	return &Asset{
		ID:          id,
		Symbol:      symbol,
		Name:        name,
		Decimals:    decimals,
		TotalSupply: new(big.Int).Set(totalSupply),
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// NewBridgeAsset creates a new bridge asset
func NewBridgeAsset(id, assetID, chainID, contractAddr string, balance *big.Int) (*BridgeAsset, error) {
	if id == "" {
		return nil, errors.New("bridge asset ID cannot be empty")
	}
	if assetID == "" {
		return nil, errors.New("asset ID cannot be empty")
	}
	if chainID == "" {
		return nil, errors.New("chain ID cannot be empty")
	}
	if contractAddr == "" {
		return nil, errors.New("contract address cannot be empty")
	}
	if balance == nil || balance.Sign() < 0 {
		return nil, errors.New("balance must be non-negative")
	}
	
	now := time.Now()
	
	return &BridgeAsset{
		ID:           id,
		AssetID:      assetID,
		ChainID:      chainID,
		ContractAddr: contractAddr,
		Balance:      new(big.Int).Set(balance),
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// NewBridge creates a new cross-chain bridge
func NewBridge(id, name, description, sourceChainID, targetChainID string, minTransfer, maxTransfer, bridgeFee *big.Int) (*Bridge, error) {
	if id == "" {
		return nil, errors.New("bridge ID cannot be empty")
	}
	if name == "" {
		return nil, errors.New("bridge name cannot be empty")
	}
	if sourceChainID == "" {
		return nil, errors.New("source chain ID cannot be empty")
	}
	if targetChainID == "" {
		return nil, errors.New("target chain ID cannot be empty")
	}
	if sourceChainID == targetChainID {
		return nil, errors.New("source and target chains must be different")
	}
	if minTransfer == nil || minTransfer.Sign() < 0 {
		return nil, errors.New("min transfer must be non-negative")
	}
	if maxTransfer == nil || maxTransfer.Sign() < 0 {
		return nil, errors.New("max transfer must be non-negative")
	}
	if minTransfer.Cmp(maxTransfer) > 0 {
		return nil, errors.New("min transfer cannot be greater than max transfer")
	}
	if bridgeFee == nil || bridgeFee.Sign() < 0 {
		return nil, errors.New("bridge fee must be non-negative")
	}
	
	now := time.Now()
	
	return &Bridge{
		ID:            id,
		Name:          name,
		Description:   description,
		SourceChainID: sourceChainID,
		TargetChainID: targetChainID,
		Status:        BridgeActive,
		MinTransfer:   new(big.Int).Set(minTransfer),
		MaxTransfer:   new(big.Int).Set(maxTransfer),
		BridgeFee:     new(big.Int).Set(bridgeFee),
		Validators:    []string{},
		IsActive:      true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

// NewTransfer creates a new cross-chain transfer
func NewTransfer(id, sourceChainID, targetChainID, assetID string, amount *big.Int, sender, recipient string) (*Transfer, error) {
	if id == "" {
		return nil, errors.New("transfer ID cannot be empty")
	}
	if sourceChainID == "" {
		return nil, errors.New("source chain ID cannot be empty")
	}
	if targetChainID == "" {
		return nil, errors.New("target chain ID cannot be empty")
	}
	if sourceChainID == targetChainID {
		return nil, errors.New("source and target chains must be different")
	}
	if assetID == "" {
		return nil, errors.New("asset ID cannot be empty")
	}
	if amount == nil || amount.Sign() <= 0 {
		return nil, errors.New("amount must be positive")
	}
	if sender == "" {
		return nil, errors.New("sender cannot be empty")
	}
	if recipient == "" {
		return nil, errors.New("recipient cannot be empty")
	}
	
	now := time.Now()
	
	return &Transfer{
		ID:            id,
		SourceChainID: sourceChainID,
		TargetChainID: targetChainID,
		AssetID:       assetID,
		Amount:        new(big.Int).Set(amount),
		Sender:        sender,
		Recipient:     recipient,
		Status:        TransferPending,
		SourceTxHash:  "",
		TargetTxHash:  "",
		BridgeFee:     big.NewInt(0),
		GasFee:        big.NewInt(0),
		CreatedAt:     now,
		UpdatedAt:     now,
		CompletedAt:   nil,
	}, nil
}

// NewValidator creates a new bridge validator
func NewValidator(id, address, chainID string, stakeAmount *big.Int) (*Validator, error) {
	if id == "" {
		return nil, errors.New("validator ID cannot be empty")
	}
	if address == "" {
		return nil, errors.New("validator address cannot be empty")
	}
	if chainID == "" {
		return nil, errors.New("chain ID cannot be empty")
	}
	if stakeAmount == nil || stakeAmount.Sign() < 0 {
		return nil, errors.New("stake amount must be non-negative")
	}
	
	now := time.Now()
	
	return &Validator{
		ID:          id,
		Address:     address,
		ChainID:     chainID,
		StakeAmount: new(big.Int).Set(stakeAmount),
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// AddChain adds a chain to the bridge manager
func (bm *BridgeManager) AddChain(chain *Chain) error {
	if chain == nil {
		return errors.New("chain cannot be nil")
	}
	
	bm.mu.Lock()
	defer bm.mu.Unlock()
	
	if _, exists := bm.Chains[chain.ID]; exists {
		return errors.New("chain with this ID already exists")
	}
	
	bm.Chains[chain.ID] = chain
	bm.UpdatedAt = time.Now()
	
	return nil
}

// AddAsset adds an asset to the bridge manager
func (bm *BridgeManager) AddAsset(asset *Asset) error {
	if asset == nil {
		return errors.New("asset cannot be nil")
	}
	
	bm.mu.Lock()
	defer bm.mu.Unlock()
	
	if _, exists := bm.Assets[asset.ID]; exists {
		return errors.New("asset with this ID already exists")
	}
	
	bm.Assets[asset.ID] = asset
	bm.UpdatedAt = time.Now()
	
	return nil
}

// AddBridgeAsset adds a bridge asset to the bridge manager
func (bm *BridgeManager) AddBridgeAsset(bridgeAsset *BridgeAsset) error {
	if bridgeAsset == nil {
		return errors.New("bridge asset cannot be nil")
	}
	
	bm.mu.Lock()
	defer bm.mu.Unlock()
	
	// Validate asset exists
	if _, exists := bm.Assets[bridgeAsset.AssetID]; !exists {
		return errors.New("asset not found")
	}
	
	// Validate chain exists
	if _, exists := bm.Chains[bridgeAsset.ChainID]; !exists {
		return errors.New("chain not found")
	}
	
	if _, exists := bm.BridgeAssets[bridgeAsset.ID]; exists {
		return errors.New("bridge asset with this ID already exists")
	}
	
	bm.BridgeAssets[bridgeAsset.ID] = bridgeAsset
	bm.UpdatedAt = time.Now()
	
	return nil
}

// AddBridge adds a bridge to the bridge manager
func (bm *BridgeManager) AddBridge(bridge *Bridge) error {
	if bridge == nil {
		return errors.New("bridge cannot be nil")
	}
	
	bm.mu.Lock()
	defer bm.mu.Unlock()
	
	// Validate source chain exists
	if _, exists := bm.Chains[bridge.SourceChainID]; !exists {
		return errors.New("source chain not found")
	}
	
	// Validate target chain exists
	if _, exists := bm.Chains[bridge.TargetChainID]; !exists {
		return errors.New("target chain not found")
	}
	
	if _, exists := bm.Bridges[bridge.ID]; exists {
		return errors.New("bridge with this ID already exists")
	}
	
	bm.Bridges[bridge.ID] = bridge
	bm.UpdatedAt = time.Now()
	
	return nil
}

// AddValidator adds a validator to the bridge manager
func (bm *BridgeManager) AddValidator(validator *Validator) error {
	if validator == nil {
		return errors.New("validator cannot be nil")
	}
	
	bm.mu.Lock()
	defer bm.mu.Unlock()
	
	// Validate chain exists
	if _, exists := bm.Chains[validator.ChainID]; !exists {
		return errors.New("chain not found")
	}
	
	if _, exists := bm.Validators[validator.ID]; exists {
		return errors.New("validator with this ID already exists")
	}
	
	bm.Validators[validator.ID] = validator
	bm.UpdatedAt = time.Now()
	
	return nil
}

// InitiateTransfer initiates a cross-chain transfer
func (bm *BridgeManager) InitiateTransfer(transfer *Transfer) error {
	if transfer == nil {
		return errors.New("transfer cannot be nil")
	}
	
	bm.mu.Lock()
	defer bm.mu.Unlock()
	
	// Validate transfer parameters
	if _, exists := bm.Chains[transfer.SourceChainID]; !exists {
		return errors.New("source chain not found")
	}
	if _, exists := bm.Chains[transfer.TargetChainID]; !exists {
		return errors.New("target chain not found")
	}
	if _, exists := bm.Assets[transfer.AssetID]; !exists {
		return errors.New("asset not found")
	}
	
	// Find bridge for this transfer
	var bridge *Bridge
	for _, b := range bm.Bridges {
		if b.SourceChainID == transfer.SourceChainID && 
		   b.TargetChainID == transfer.TargetChainID && 
		   b.IsActive && 
		   b.Status == BridgeActive {
			bridge = b
			break
		}
	}
	
	if bridge == nil {
		return errors.New("no active bridge found for this transfer")
	}
	
	// Validate transfer amount
	if transfer.Amount.Cmp(bridge.MinTransfer) < 0 {
		return errors.New("transfer amount below minimum")
	}
	if transfer.Amount.Cmp(bridge.MaxTransfer) > 0 {
		return errors.New("transfer amount above maximum")
	}
	
	// Calculate bridge fee
	transfer.BridgeFee = new(big.Int).Set(bridge.BridgeFee)
	
	// Check if transfer already exists
	if _, exists := bm.Transfers[transfer.ID]; exists {
		return errors.New("transfer with this ID already exists")
	}
	
	bm.Transfers[transfer.ID] = transfer
	bm.UpdatedAt = time.Now()
	
	return nil
}

// ConfirmTransfer confirms a transfer on the source chain
func (bm *BridgeManager) ConfirmTransfer(transferID, sourceTxHash string) error {
	if transferID == "" {
		return errors.New("transfer ID cannot be empty")
	}
	if sourceTxHash == "" {
		return errors.New("source transaction hash cannot be empty")
	}
	
	bm.mu.Lock()
	defer bm.mu.Unlock()
	
	transfer, exists := bm.Transfers[transferID]
	if !exists {
		return errors.New("transfer not found")
	}
	
	if transfer.Status != TransferPending {
		return errors.New("transfer is not in pending status")
	}
	
	transfer.SourceTxHash = sourceTxHash
	transfer.Status = TransferConfirmed
	transfer.UpdatedAt = time.Now()
	bm.UpdatedAt = time.Now()
	
	return nil
}

// ProcessTransfer processes a confirmed transfer on the target chain
func (bm *BridgeManager) ProcessTransfer(transferID, targetTxHash string) error {
	if transferID == "" {
		return errors.New("transfer ID cannot be empty")
	}
	if targetTxHash == "" {
		return errors.New("target transaction hash cannot be empty")
	}
	
	bm.mu.Lock()
	defer bm.mu.Unlock()
	
	transfer, exists := bm.Transfers[transferID]
	if !exists {
		return errors.New("transfer not found")
	}
	
	if transfer.Status != TransferConfirmed {
		return errors.New("transfer is not in confirmed status")
	}
	
	transfer.TargetTxHash = targetTxHash
	transfer.Status = TransferProcessing
	transfer.UpdatedAt = time.Now()
	bm.UpdatedAt = time.Now()
	
	return nil
}

// CompleteTransfer completes a processed transfer
func (bm *BridgeManager) CompleteTransfer(transferID string) error {
	if transferID == "" {
		return errors.New("transfer ID cannot be empty")
	}
	
	bm.mu.Lock()
	defer bm.mu.Unlock()
	
	transfer, exists := bm.Transfers[transferID]
	if !exists {
		return errors.New("transfer not found")
	}
	
	if transfer.Status != TransferProcessing {
		return errors.New("transfer is not in processing status")
	}
	
	transfer.Status = TransferCompleted
	now := time.Now()
	transfer.CompletedAt = &now
	transfer.UpdatedAt = now
	bm.UpdatedAt = now
	
	return nil
}

// FailTransfer marks a transfer as failed
func (bm *BridgeManager) FailTransfer(transferID string, reason string) error {
	if transferID == "" {
		return errors.New("transfer ID cannot be empty")
	}
	
	bm.mu.Lock()
	defer bm.mu.Unlock()
	
	transfer, exists := bm.Transfers[transferID]
	if !exists {
		return errors.New("transfer not found")
	}
	
	if transfer.Status == TransferCompleted || transfer.Status == TransferFailed {
		return errors.New("transfer cannot be marked as failed")
	}
	
	transfer.Status = TransferFailed
	transfer.UpdatedAt = time.Now()
	bm.UpdatedAt = time.Now()
	
	return nil
}

// GetTransfer returns a transfer by ID
func (bm *BridgeManager) GetTransfer(transferID string) (*Transfer, error) {
	if transferID == "" {
		return nil, errors.New("transfer ID cannot be empty")
	}
	
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	
	transfer, exists := bm.Transfers[transferID]
	if !exists {
		return nil, errors.New("transfer not found")
	}
	
	return transfer, nil
}

// GetTransfersByStatus returns all transfers with a specific status
func (bm *BridgeManager) GetTransfersByStatus(status TransferStatus) []*Transfer {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	
	var transfers []*Transfer
	for _, transfer := range bm.Transfers {
		if transfer.Status == status {
			transfers = append(transfers, transfer)
		}
	}
	
	return transfers
}

// GetTransfersByChain returns all transfers for a specific chain
func (bm *BridgeManager) GetTransfersByChain(chainID string) []*Transfer {
	if chainID == "" {
		return nil
	}
	
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	
	var transfers []*Transfer
	for _, transfer := range bm.Transfers {
		if transfer.SourceChainID == chainID || transfer.TargetChainID == chainID {
			transfers = append(transfers, transfer)
		}
	}
	
	return transfers
}

// GetBridge returns a bridge by ID
func (bm *BridgeManager) GetBridge(bridgeID string) (*Bridge, error) {
	if bridgeID == "" {
		return nil, errors.New("bridge ID cannot be empty")
	}
	
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	
	bridge, exists := bm.Bridges[bridgeID]
	if !exists {
		return nil, errors.New("bridge not found")
	}
	
	return bridge, nil
}

// GetBridgesByChain returns all bridges for a specific chain
func (bm *BridgeManager) GetBridgesByChain(chainID string) []*Bridge {
	if chainID == "" {
		return nil
	}
	
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	
	var bridges []*Bridge
	for _, bridge := range bm.Bridges {
		if bridge.SourceChainID == chainID || bridge.TargetChainID == chainID {
			bridges = append(bridges, bridge)
		}
	}
	
	return bridges
}

// UpdateBridgeStatus updates the status of a bridge
func (bm *BridgeManager) UpdateBridgeStatus(bridgeID string, status BridgeStatus) error {
	if bridgeID == "" {
		return errors.New("bridge ID cannot be empty")
	}
	
	bm.mu.Lock()
	defer bm.mu.Unlock()
	
	bridge, exists := bm.Bridges[bridgeID]
	if !exists {
		return errors.New("bridge not found")
	}
	
	bridge.Status = status
	bridge.UpdatedAt = time.Now()
	bm.UpdatedAt = time.Now()
	
	return nil
}

// AddValidatorToBridge adds a validator to a bridge
func (bm *BridgeManager) AddValidatorToBridge(bridgeID, validatorID string) error {
	if bridgeID == "" {
		return errors.New("bridge ID cannot be empty")
	}
	if validatorID == "" {
		return errors.New("validator ID cannot be empty")
	}
	
	bm.mu.Lock()
	defer bm.mu.Unlock()
	
	bridge, exists := bm.Bridges[bridgeID]
	if !exists {
		return errors.New("bridge not found")
	}
	
	if _, exists := bm.Validators[validatorID]; !exists {
		return errors.New("validator not found")
	}
	
	// Check if validator is already in the bridge
	for _, vID := range bridge.Validators {
		if vID == validatorID {
			return errors.New("validator already in bridge")
		}
	}
	
	bridge.Validators = append(bridge.Validators, validatorID)
	bridge.UpdatedAt = time.Now()
	bm.UpdatedAt = time.Now()
	
	return nil
}

// RemoveValidatorFromBridge removes a validator from a bridge
func (bm *BridgeManager) RemoveValidatorFromBridge(bridgeID, validatorID string) error {
	if bridgeID == "" {
		return errors.New("bridge ID cannot be empty")
	}
	if validatorID == "" {
		return errors.New("validator ID cannot be empty")
	}
	
	bm.mu.Lock()
	defer bm.mu.Unlock()
	
	bridge, exists := bm.Bridges[bridgeID]
	if !exists {
		return errors.New("bridge not found")
	}
	
	// Find and remove validator
	for i, vID := range bridge.Validators {
		if vID == validatorID {
			bridge.Validators = append(bridge.Validators[:i], bridge.Validators[i+1:]...)
			bridge.UpdatedAt = time.Now()
			bm.UpdatedAt = time.Now()
			return nil
		}
	}
	
	return errors.New("validator not found in bridge")
}

// GetTotalTransfers returns the total number of transfers
func (bm *BridgeManager) GetTotalTransfers() int {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	
	return len(bm.Transfers)
}

// GetTotalVolume returns the total volume of completed transfers
func (bm *BridgeManager) GetTotalVolume() *big.Int {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	
	totalVolume := big.NewInt(0)
	for _, transfer := range bm.Transfers {
		if transfer.Status == TransferCompleted {
			totalVolume.Add(totalVolume, transfer.Amount)
		}
	}
	
	return totalVolume
}

// GetTotalFees returns the total fees collected
func (bm *BridgeManager) GetTotalFees() *big.Int {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	
	totalFees := big.NewInt(0)
	for _, transfer := range bm.Transfers {
		if transfer.Status == TransferCompleted {
			totalFees.Add(totalFees, transfer.BridgeFee)
		}
	}
	
	return totalFees
}

// GenerateTransferID generates a unique transfer ID
func GenerateTransferID(sourceChainID, targetChainID, assetID string, timestamp time.Time) string {
	data := sourceChainID + targetChainID + assetID + timestamp.Format(time.RFC3339)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:8]) // Use first 8 bytes for shorter ID
}

// ValidateTransfer validates a transfer request
func (bm *BridgeManager) ValidateTransfer(transfer *Transfer) error {
	if transfer == nil {
		return errors.New("transfer cannot be nil")
	}
	
	// Validate chains
	if _, exists := bm.Chains[transfer.SourceChainID]; !exists {
		return errors.New("source chain not found")
	}
	if _, exists := bm.Chains[transfer.TargetChainID]; !exists {
		return errors.New("target chain not found")
	}
	
	// Validate asset
	if _, exists := bm.Assets[transfer.AssetID]; !exists {
		return errors.New("asset not found")
	}
	
	// Validate amount
	if transfer.Amount.Sign() <= 0 {
		return errors.New("amount must be positive")
	}
	
	// Validate addresses
	if transfer.Sender == "" {
		return errors.New("sender address cannot be empty")
	}
	if transfer.Recipient == "" {
		return errors.New("recipient address cannot be empty")
	}
	
	return nil
}

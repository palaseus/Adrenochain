package tokens

import (
	"math/big"
	"sync"
	"time"

	"github.com/palaseus/adrenochain/pkg/contracts/engine"
)

// ERC20Token represents an ERC-20 compliant fungible token
type ERC20Token struct {
	mu sync.RWMutex

	// Token metadata
	Name        string
	Symbol      string
	Decimals    uint8
	TotalSupply *big.Int

	// State
	Balances    map[engine.Address]*big.Int
	Allowances  map[engine.Address]map[engine.Address]*big.Int
	Owner       engine.Address
	Paused      bool
	Blacklisted map[engine.Address]bool

	// Events
	TransferEvents    []TransferEvent
	ApprovalEvents   []ApprovalEvent
	MintEvents       []MintEvent
	BurnEvents       []BurnEvent
	PauseEvents      []PauseEvent
	BlacklistEvents  []BlacklistEvent

	// Configuration
	MaxSupply        *big.Int
	TransferFee      *big.Int
	TransferFeeRecipient engine.Address
	Mintable        bool
	Burnable        bool
	Pausable        bool
	Blacklistable   bool
}

// NewERC20Token creates a new ERC-20 token
func NewERC20Token(
	name, symbol string,
	decimals uint8,
	totalSupply *big.Int,
	owner engine.Address,
	config TokenConfig,
) *ERC20Token {
	token := &ERC20Token{
		Name:                 name,
		Symbol:               symbol,
		Decimals:             decimals,
		TotalSupply:          new(big.Int).Set(totalSupply),
		Balances:             make(map[engine.Address]*big.Int),
		Allowances:           make(map[engine.Address]map[engine.Address]*big.Int),
		Owner:                owner,
		Paused:               false,
		Blacklisted:          make(map[engine.Address]bool),
		TransferEvents:       make([]TransferEvent, 0),
		ApprovalEvents:       make([]ApprovalEvent, 0),
		MintEvents:           make([]MintEvent, 0),
		BurnEvents:           make([]BurnEvent, 0),
		PauseEvents:          make([]PauseEvent, 0),
		BlacklistEvents:      make([]BlacklistEvent, 0),
		MaxSupply:            config.MaxSupply,
		TransferFee:          config.TransferFee,
		TransferFeeRecipient: config.TransferFeeRecipient,
		Mintable:             config.Mintable,
		Burnable:             config.Burnable,
		Pausable:             config.Pausable,
		Blacklistable:        config.Blacklistable,
	}
	
	// Set initial balance for owner
	token.Balances[owner] = new(big.Int).Set(totalSupply)
	
	return token
}

// TokenConfig holds configuration options for ERC-20 tokens
type TokenConfig struct {
	MaxSupply            *big.Int
	TransferFee          *big.Int
	TransferFeeRecipient engine.Address
	Mintable             bool
	Burnable             bool
	Pausable             bool
	Blacklistable        bool
}

// DefaultTokenConfig returns a default token configuration
func DefaultTokenConfig() TokenConfig {
	return TokenConfig{
		MaxSupply:            nil, // No max supply
		TransferFee:          big.NewInt(0), // No transfer fee
		TransferFeeRecipient: engine.Address{},
		Mintable:             false,
		Burnable:             false,
		Pausable:             false,
		Blacklistable:        false,
	}
}

// TransferEvent represents a token transfer event
type TransferEvent struct {
	From   engine.Address
	To     engine.Address
	Value  *big.Int
	TxHash engine.Hash
	Block  uint64
	Time   time.Time
}

// ApprovalEvent represents a token approval event
type ApprovalEvent struct {
	Owner   engine.Address
	Spender engine.Address
	Value   *big.Int
	TxHash  engine.Hash
	Block   uint64
	Time    time.Time
}

// MintEvent represents a token minting event
type MintEvent struct {
	To     engine.Address
	Value  *big.Int
	TxHash engine.Hash
	Block  uint64
	Time   time.Time
}

// BurnEvent represents a token burning event
type BurnEvent struct {
	From   engine.Address
	Value  *big.Int
	TxHash engine.Hash
	Block  uint64
	Time   time.Time
}

// PauseEvent represents a token pause/unpause event
type PauseEvent struct {
	Paused bool
	TxHash engine.Hash
	Block  uint64
	Time   time.Time
}

// BlacklistEvent represents a blacklist add/remove event
type BlacklistEvent struct {
	Address engine.Address
	Blacklisted bool
	TxHash     engine.Hash
	Block      uint64
	Time       time.Time
}

// GetName returns the token name
func (t *ERC20Token) GetName() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Name
}

// GetSymbol returns the token symbol
func (t *ERC20Token) GetSymbol() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Symbol
}

// GetDecimals returns the token decimals
func (t *ERC20Token) GetDecimals() uint8 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Decimals
}

// GetTotalSupply returns the total token supply
func (t *ERC20Token) GetTotalSupply() *big.Int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return new(big.Int).Set(t.TotalSupply)
}

// GetBalance returns the token balance of an address
func (t *ERC20Token) GetBalance(address engine.Address) *big.Int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	if balance, exists := t.Balances[address]; exists {
		return new(big.Int).Set(balance)
	}
	return big.NewInt(0)
}

// GetAllowance returns the allowance given by owner to spender
func (t *ERC20Token) GetAllowance(owner, spender engine.Address) *big.Int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	if ownerAllowances, exists := t.Allowances[owner]; exists {
		if allowance, exists := ownerAllowances[spender]; exists {
			return new(big.Int).Set(allowance)
		}
	}
	return big.NewInt(0)
}

// Transfer transfers tokens from the caller to the specified address
func (t *ERC20Token) Transfer(from, to engine.Address, value *big.Int, txHash engine.Hash, block uint64) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Check if token is paused
	if t.Paused {
		return ErrTokenPaused
	}

	// Check if addresses are blacklisted
	if t.Blacklisted[from] || t.Blacklisted[to] {
		return ErrAddressBlacklisted
	}

	// Validate input
	if value == nil || value.Sign() <= 0 {
		return ErrInvalidAmount
	}

	if from == to {
		return ErrSelfTransfer
	}

	// Check balance
	fromBalance := t.getBalanceInternal(from)
	if fromBalance.Cmp(value) < 0 {
		return ErrInsufficientBalance
	}

	// Calculate transfer fee
	transferFee := big.NewInt(0)
	if t.TransferFee != nil && t.TransferFee.Sign() > 0 {
		transferFee = new(big.Int).Mul(value, t.TransferFee)
		transferFee.Div(transferFee, big.NewInt(10000)) // Basis points (10000 = 100%)
	}

	// Calculate actual amount to transfer
	actualAmount := new(big.Int).Sub(value, transferFee)

	// Update balances
	t.Balances[from] = new(big.Int).Sub(fromBalance, value)
	t.Balances[to] = new(big.Int).Add(t.getBalanceInternal(to), actualAmount)

	// Add transfer fee to recipient if specified
	if transferFee.Sign() > 0 && t.TransferFeeRecipient != (engine.Address{}) {
		t.Balances[t.TransferFeeRecipient] = new(big.Int).Add(
			t.getBalanceInternal(t.TransferFeeRecipient),
			transferFee,
		)
	}

	// Record event
	event := TransferEvent{
		From:   from,
		To:     to,
		Value:  new(big.Int).Set(value),
		TxHash: txHash,
		Block:  block,
		Time:   time.Now(),
	}
	t.TransferEvents = append(t.TransferEvents, event)

	return nil
}

// Approve approves the specified address to spend tokens on behalf of the caller
func (t *ERC20Token) Approve(owner, spender engine.Address, value *big.Int, txHash engine.Hash, block uint64) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Check if token is paused
	if t.Paused {
		return ErrTokenPaused
	}

	// Check if addresses are blacklisted
	if t.Blacklisted[owner] || t.Blacklisted[spender] {
		return ErrAddressBlacklisted
	}

	// Validate input
	if value == nil || value.Sign() < 0 {
		return ErrInvalidAmount
	}

	// Initialize allowances map if needed
	if t.Allowances[owner] == nil {
		t.Allowances[owner] = make(map[engine.Address]*big.Int)
	}

	// Update allowance
	t.Allowances[owner][spender] = new(big.Int).Set(value)

	// Record event
	event := ApprovalEvent{
		Owner:   owner,
		Spender: spender,
		Value:   new(big.Int).Set(value),
		TxHash:  txHash,
		Block:   block,
		Time:    time.Now(),
	}
	t.ApprovalEvents = append(t.ApprovalEvents, event)

	return nil
}

// TransferFrom transfers tokens from one address to another using allowance
func (t *ERC20Token) TransferFrom(spender, from, to engine.Address, value *big.Int, txHash engine.Hash, block uint64) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Check if token is paused
	if t.Paused {
		return ErrTokenPaused
	}

	// Check if addresses are blacklisted
	if t.Blacklisted[from] || t.Blacklisted[to] || t.Blacklisted[spender] {
		return ErrAddressBlacklisted
	}

	// Validate input
	if value == nil || value.Sign() <= 0 {
		return ErrInvalidAmount
	}

	if from == to {
		return ErrSelfTransfer
	}

	// Check allowance
	allowance := t.getAllowanceInternal(from, spender)
	if allowance.Cmp(value) < 0 {
		return ErrInsufficientAllowance
	}

	// Check balance
	fromBalance := t.getBalanceInternal(from)
	if fromBalance.Cmp(value) < 0 {
		return ErrInsufficientBalance
	}

	// Calculate transfer fee
	transferFee := big.NewInt(0)
	if t.TransferFee != nil && t.TransferFee.Sign() > 0 {
		transferFee = new(big.Int).Mul(value, t.TransferFee)
		transferFee.Div(transferFee, big.NewInt(10000)) // Basis points
	}

	// Calculate actual amount to transfer
	actualAmount := new(big.Int).Sub(value, transferFee)

	// Update balances
	t.Balances[from] = new(big.Int).Sub(fromBalance, value)
	t.Balances[to] = new(big.Int).Add(t.getBalanceInternal(to), actualAmount)

	// Add transfer fee to recipient if specified
	if transferFee.Sign() > 0 && t.TransferFeeRecipient != (engine.Address{}) {
		t.Balances[t.TransferFeeRecipient] = new(big.Int).Add(
			t.getBalanceInternal(t.TransferFeeRecipient),
			transferFee,
		)
	}

	// Update allowance
	t.Allowances[from][spender] = new(big.Int).Sub(allowance, value)

	// Record event
	event := TransferEvent{
		From:   from,
		To:     to,
		Value:  new(big.Int).Set(value),
		TxHash: txHash,
		Block:  block,
		Time:   time.Now(),
	}
	t.TransferEvents = append(t.TransferEvents, event)

	return nil
}

// Mint creates new tokens and assigns them to the specified address
func (t *ERC20Token) Mint(to engine.Address, value *big.Int, txHash engine.Hash, block uint64) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Check if minting is allowed
	if !t.Mintable {
		return ErrMintingNotAllowed
	}

	// Check if token is paused
	if t.Paused {
		return ErrTokenPaused
	}

	// Check if address is blacklisted
	if t.Blacklisted[to] {
		return ErrAddressBlacklisted
	}

	// Validate input
	if value == nil || value.Sign() <= 0 {
		return ErrInvalidAmount
	}

	// Check max supply if specified
	if t.MaxSupply != nil {
		newTotalSupply := new(big.Int).Add(t.TotalSupply, value)
		if newTotalSupply.Cmp(t.MaxSupply) > 0 {
			return ErrExceedsMaxSupply
		}
	}

	// Update balances and total supply
	t.Balances[to] = new(big.Int).Add(t.getBalanceInternal(to), value)
	t.TotalSupply = new(big.Int).Add(t.TotalSupply, value)

	// Record event
	event := MintEvent{
		To:     to,
		Value:  new(big.Int).Set(value),
		TxHash: txHash,
		Block:  block,
		Time:   time.Now(),
	}
	t.MintEvents = append(t.MintEvents, event)

	return nil
}

// Burn destroys tokens from the specified address
func (t *ERC20Token) Burn(from engine.Address, value *big.Int, txHash engine.Hash, block uint64) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Check if burning is allowed
	if !t.Burnable {
		return ErrBurningNotAllowed
	}

	// Check if token is paused
	if t.Paused {
		return ErrTokenPaused
	}

	// Check if address is blacklisted
	if t.Blacklisted[from] {
		return ErrAddressBlacklisted
	}

	// Validate input
	if value == nil || value.Sign() <= 0 {
		return ErrInvalidAmount
	}

	// Check balance
	fromBalance := t.getBalanceInternal(from)
	if fromBalance.Cmp(value) < 0 {
		return ErrInsufficientBalance
	}

	// Update balances and total supply
	t.Balances[from] = new(big.Int).Sub(fromBalance, value)
	t.TotalSupply = new(big.Int).Sub(t.TotalSupply, value)

	// Record event
	event := BurnEvent{
		From:   from,
		Value:  new(big.Int).Set(value),
		TxHash: txHash,
		Block:  block,
		Time:   time.Now(),
	}
	t.BurnEvents = append(t.BurnEvents, event)

	return nil
}

// Pause pauses all token operations
func (t *ERC20Token) Pause(txHash engine.Hash, block uint64) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Check if pausing is allowed
	if !t.Pausable {
		return ErrPausingNotAllowed
	}

	if t.Paused {
		return ErrTokenAlreadyPaused
	}

	t.Paused = true

	// Record event
	event := PauseEvent{
		Paused: true,
		TxHash: txHash,
		Block:  block,
		Time:   time.Now(),
	}
	t.PauseEvents = append(t.PauseEvents, event)

	return nil
}

// Unpause resumes all token operations
func (t *ERC20Token) Unpause(txHash engine.Hash, block uint64) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Check if pausing is allowed
	if !t.Pausable {
		return ErrPausingNotAllowed
	}

	if !t.Paused {
		return ErrTokenNotPaused
	}

	t.Paused = false

	// Record event
	event := PauseEvent{
		Paused: false,
		TxHash: txHash,
		Block:  block,
		Time:   time.Now(),
	}
	t.PauseEvents = append(t.PauseEvents, event)

	return nil
}

// AddToBlacklist adds an address to the blacklist
func (t *ERC20Token) AddToBlacklist(address engine.Address, txHash engine.Hash, block uint64) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Check if blacklisting is allowed
	if !t.Blacklistable {
		return ErrBlacklistingNotAllowed
	}

	if t.Blacklisted[address] {
		return ErrAddressAlreadyBlacklisted
	}

	t.Blacklisted[address] = true

	// Record event
	event := BlacklistEvent{
		Address:    address,
		Blacklisted: true,
		TxHash:     txHash,
		Block:      block,
		Time:       time.Now(),
	}
	t.BlacklistEvents = append(t.BlacklistEvents, event)

	return nil
}

// RemoveFromBlacklist removes an address from the blacklist
func (t *ERC20Token) RemoveFromBlacklist(address engine.Address, txHash engine.Hash, block uint64) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Check if blacklisting is allowed
	if !t.Blacklistable {
		return ErrBlacklistingNotAllowed
	}

	if !t.Blacklisted[address] {
		return ErrAddressNotBlacklisted
	}

	delete(t.Blacklisted, address)

	// Record event
	event := BlacklistEvent{
		Address:    address,
		Blacklisted: false,
		TxHash:     txHash,
		Block:      block,
		Time:       time.Now(),
	}
	t.BlacklistEvents = append(t.BlacklistEvents, event)

	return nil
}

// GetTransferEvents returns all transfer events
func (t *ERC20Token) GetTransferEvents() []TransferEvent {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	events := make([]TransferEvent, len(t.TransferEvents))
	copy(events, t.TransferEvents)
	return events
}

// GetApprovalEvents returns all approval events
func (t *ERC20Token) GetApprovalEvents() []ApprovalEvent {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	events := make([]ApprovalEvent, len(t.ApprovalEvents))
	copy(events, t.ApprovalEvents)
	return events
}

// GetMintEvents returns all mint events
func (t *ERC20Token) GetMintEvents() []MintEvent {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	events := make([]MintEvent, len(t.MintEvents))
	copy(events, t.MintEvents)
	return events
}

// GetBurnEvents returns all burn events
func (t *ERC20Token) GetBurnEvents() []BurnEvent {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	events := make([]BurnEvent, len(t.BurnEvents))
	copy(events, t.BurnEvents)
	return events
}

// GetPauseEvents returns all pause events
func (t *ERC20Token) GetPauseEvents() []PauseEvent {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	events := make([]PauseEvent, len(t.PauseEvents))
	copy(events, t.PauseEvents)
	return events
}

// GetBlacklistEvents returns all blacklist events
func (t *ERC20Token) GetBlacklistEvents() []BlacklistEvent {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	events := make([]BlacklistEvent, len(t.BlacklistEvents))
	copy(events, t.BlacklistEvents)
	return events
}

// SetBalanceForTesting sets a balance for testing purposes only
func (t *ERC20Token) SetBalanceForTesting(address engine.Address, balance *big.Int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Balances[address] = new(big.Int).Set(balance)
}

// SetAllowanceForTesting sets an allowance for testing purposes only
func (t *ERC20Token) SetAllowanceForTesting(owner, spender engine.Address, allowance *big.Int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.Allowances[owner] == nil {
		t.Allowances[owner] = make(map[engine.Address]*big.Int)
	}
	t.Allowances[owner][spender] = new(big.Int).Set(allowance)
}

// IsPaused returns whether the token is paused
func (t *ERC20Token) IsPaused() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Paused
}

// IsBlacklisted returns whether an address is blacklisted
func (t *ERC20Token) IsBlacklisted(address engine.Address) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Blacklisted[address]
}

// GetOwner returns the token owner
func (t *ERC20Token) GetOwner() engine.Address {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Owner
}

// GetTransferFee returns the transfer fee (in basis points)
func (t *ERC20Token) GetTransferFee() *big.Int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.TransferFee == nil {
		return big.NewInt(0)
	}
	return new(big.Int).Set(t.TransferFee)
}

// GetTransferFeeRecipient returns the transfer fee recipient
func (t *ERC20Token) GetTransferFeeRecipient() engine.Address {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.TransferFeeRecipient
}

// GetMaxSupply returns the maximum token supply
func (t *ERC20Token) GetMaxSupply() *big.Int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.MaxSupply == nil {
		return nil
	}
	return new(big.Int).Set(t.MaxSupply)
}

// IsMintable returns whether the token is mintable
func (t *ERC20Token) IsMintable() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Mintable
}

// IsBurnable returns whether the token is burnable
func (t *ERC20Token) IsBurnable() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Burnable
}

// IsPausable returns whether the token is pausable
func (t *ERC20Token) IsPausable() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Pausable
}

// IsBlacklistable returns whether the token supports blacklisting
func (t *ERC20Token) IsBlacklistable() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Blacklistable
}

// Clone creates a deep copy of the token
func (t *ERC20Token) Clone() *ERC20Token {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	clone := &ERC20Token{
		Name:                 t.Name,
		Symbol:               t.Symbol,
		Decimals:             t.Decimals,
		TotalSupply:          new(big.Int).Set(t.TotalSupply),
		Balances:             make(map[engine.Address]*big.Int),
		Allowances:           make(map[engine.Address]map[engine.Address]*big.Int),
		Owner:                t.Owner,
		Paused:               t.Paused,
		Blacklisted:          make(map[engine.Address]bool),
		TransferEvents:       make([]TransferEvent, len(t.TransferEvents)),
		ApprovalEvents:       make([]ApprovalEvent, len(t.ApprovalEvents)),
		MintEvents:           make([]MintEvent, len(t.MintEvents)),
		BurnEvents:           make([]BurnEvent, len(t.BurnEvents)),
		PauseEvents:          make([]PauseEvent, len(t.PauseEvents)),
		BlacklistEvents:      make([]BlacklistEvent, len(t.BlacklistEvents)),
		MaxSupply:            nil,
		TransferFee:          nil,
		TransferFeeRecipient: t.TransferFeeRecipient,
		Mintable:             t.Mintable,
		Burnable:             t.Burnable,
		Pausable:             t.Pausable,
		Blacklistable:        t.Blacklistable,
	}
	
	// Copy balances
	for addr, balance := range t.Balances {
		clone.Balances[addr] = new(big.Int).Set(balance)
	}
	
	// Copy allowances
	for owner, allowances := range t.Allowances {
		clone.Allowances[owner] = make(map[engine.Address]*big.Int)
		for spender, allowance := range allowances {
			clone.Allowances[owner][spender] = new(big.Int).Set(allowance)
		}
	}
	
	// Copy blacklisted addresses
	for addr, blacklisted := range t.Blacklisted {
		clone.Blacklisted[addr] = blacklisted
	}
	
	// Copy max supply
	if t.MaxSupply != nil {
		clone.MaxSupply = new(big.Int).Set(t.MaxSupply)
	}
	
	// Copy transfer fee
	if t.TransferFee != nil {
		clone.TransferFee = new(big.Int).Set(t.TransferFee)
	}
	
	// Copy events
	copy(clone.TransferEvents, t.TransferEvents)
	copy(clone.ApprovalEvents, t.ApprovalEvents)
	copy(clone.MintEvents, t.MintEvents)
	copy(clone.BurnEvents, t.BurnEvents)
	copy(clone.PauseEvents, t.PauseEvents)
	copy(clone.BlacklistEvents, t.BlacklistEvents)
	
	return clone
}

// Internal helper methods (not thread-safe, caller must hold lock)
func (t *ERC20Token) getBalanceInternal(address engine.Address) *big.Int {
	if balance, exists := t.Balances[address]; exists {
		return balance
	}
	return big.NewInt(0)
}

func (t *ERC20Token) getAllowanceInternal(owner, spender engine.Address) *big.Int {
	if ownerAllowances, exists := t.Allowances[owner]; exists {
		if allowance, exists := ownerAllowances[spender]; exists {
			return allowance
		}
	}
	return big.NewInt(0)
}

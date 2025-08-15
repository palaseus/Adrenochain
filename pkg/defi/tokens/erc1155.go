package tokens

import (
	"math/big"
	"sync"
	"time"

	"github.com/gochain/gochain/pkg/contracts/engine"
)

// ERC1155Token represents an ERC-1155 compliant multi-token standard
type ERC1155Token struct {
	mu sync.RWMutex

	// Token metadata
	URI         string
	Owner       engine.Address
	Paused      bool
	Blacklisted map[engine.Address]bool

	// State
	Balances       map[engine.Address]map[uint64]*big.Int     // address -> tokenID -> balance
	TokenApprovals map[engine.Address]map[engine.Address]bool // owner -> operator -> approved
	TokenURIs      map[uint64]string                          // tokenID -> URI

	// Events
	TransferSingleEvents []ERC1155TransferSingleEvent
	TransferBatchEvents  []ERC1155TransferBatchEvent
	ApprovalForAllEvents []ERC1155ApprovalForAllEvent
	URIEvents            []ERC1155URIEvent
	MintEvents           []ERC1155MintEvent
	BurnEvents           []ERC1155BurnEvent
	PauseEvents          []PauseEvent
	BlacklistEvents      []BlacklistEvent

	// Configuration
	MaxSupply     map[uint64]*big.Int // tokenID -> max supply
	Mintable      bool
	Burnable      bool
	Pausable      bool
	Blacklistable bool
	MetadataURI   bool
}

// NewERC1155Token creates a new ERC-1155 token
func NewERC1155Token(
	uri string,
	owner engine.Address,
	config ERC1155TokenConfig,
) *ERC1155Token {
	token := &ERC1155Token{
		URI:                  uri,
		Owner:                owner,
		Paused:               false,
		Blacklisted:          make(map[engine.Address]bool),
		Balances:             make(map[engine.Address]map[uint64]*big.Int),
		TokenApprovals:       make(map[engine.Address]map[engine.Address]bool),
		TokenURIs:            make(map[uint64]string),
		TransferSingleEvents: make([]ERC1155TransferSingleEvent, 0),
		TransferBatchEvents:  make([]ERC1155TransferBatchEvent, 0),
		ApprovalForAllEvents: make([]ERC1155ApprovalForAllEvent, 0),
		URIEvents:            make([]ERC1155URIEvent, 0),
		MintEvents:           make([]ERC1155MintEvent, 0),
		BurnEvents:           make([]ERC1155BurnEvent, 0),
		PauseEvents:          make([]PauseEvent, 0),
		BlacklistEvents:      make([]BlacklistEvent, 0),
		MaxSupply:            config.MaxSupply,
		Mintable:             config.Mintable,
		Burnable:             config.Burnable,
		Pausable:             config.Pausable,
		Blacklistable:        config.Blacklistable,
		MetadataURI:          config.MetadataURI,
	}

	return token
}

// ERC1155TokenConfig holds configuration options for ERC-1155 tokens
type ERC1155TokenConfig struct {
	MaxSupply     map[uint64]*big.Int
	Mintable      bool
	Burnable      bool
	Pausable      bool
	Blacklistable bool
	MetadataURI   bool
}

// DefaultERC1155TokenConfig returns a default ERC-1155 token configuration
func DefaultERC1155TokenConfig() ERC1155TokenConfig {
	return ERC1155TokenConfig{
		MaxSupply:     make(map[uint64]*big.Int),
		Mintable:      false,
		Burnable:      false,
		Pausable:      false,
		Blacklistable: false,
		MetadataURI:   true,
	}
}

// ERC1155TransferSingleEvent represents a single token transfer event
type ERC1155TransferSingleEvent struct {
	Operator engine.Address
	From     engine.Address
	To       engine.Address
	ID       uint64
	Value    *big.Int
	TxHash   engine.Hash
	Block    uint64
	Time     time.Time
}

// ERC1155TransferBatchEvent represents a batch token transfer event
type ERC1155TransferBatchEvent struct {
	Operator engine.Address
	From     engine.Address
	To       engine.Address
	IDs      []uint64
	Values   []*big.Int
	TxHash   engine.Hash
	Block    uint64
	Time     time.Time
}

// ERC1155ApprovalForAllEvent represents an operator approval event
type ERC1155ApprovalForAllEvent struct {
	Owner    engine.Address
	Operator engine.Address
	Approved bool
	TxHash   engine.Hash
	Block    uint64
	Time     time.Time
}

// ERC1155URIEvent represents a URI update event
type ERC1155URIEvent struct {
	URI    string
	ID     uint64
	TxHash engine.Hash
	Block  uint64
	Time   time.Time
}

// ERC1155MintEvent represents a token minting event
type ERC1155MintEvent struct {
	To     engine.Address
	ID     uint64
	Value  *big.Int
	TxHash engine.Hash
	Block  uint64
	Time   time.Time
}

// ERC1155BurnEvent represents a token burning event
type ERC1155BurnEvent struct {
	From   engine.Address
	ID     uint64
	Value  *big.Int
	TxHash engine.Hash
	Block  uint64
	Time   time.Time
}

// GetURI returns the base URI for token metadata
func (t *ERC1155Token) GetURI() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.URI
}

// GetOwner returns the token owner
func (t *ERC1155Token) GetOwner() engine.Address {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Owner
}

// GetTokenURI returns the URI for a specific token
func (t *ERC1155Token) GetTokenURI(tokenID uint64) string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if uri, exists := t.TokenURIs[tokenID]; exists {
		return uri
	}
	return t.URI
}

// GetBalance returns the balance of a specific token for an address
func (t *ERC1155Token) GetBalance(account engine.Address, tokenID uint64) *big.Int {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if accountBalances, exists := t.Balances[account]; exists {
		if balance, exists := accountBalances[tokenID]; exists {
			return new(big.Int).Set(balance)
		}
	}
	return big.NewInt(0)
}

// GetBalances returns all token balances for an address
func (t *ERC1155Token) GetBalances(account engine.Address) map[uint64]*big.Int {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if accountBalances, exists := t.Balances[account]; exists {
		result := make(map[uint64]*big.Int)
		for tokenID, balance := range accountBalances {
			result[tokenID] = new(big.Int).Set(balance)
		}
		return result
	}
	return make(map[uint64]*big.Int)
}

// IsApprovedForAll checks if an operator is approved for all tokens of an owner
func (t *ERC1155Token) IsApprovedForAll(owner, operator engine.Address) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if operatorApprovals, exists := t.TokenApprovals[owner]; exists {
		return operatorApprovals[operator]
	}
	return false
}

// SetApprovalForAll approves or revokes approval for an operator to manage all tokens
func (t *ERC1155Token) SetApprovalForAll(owner, operator engine.Address, approved bool, txHash engine.Hash, block uint64) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Check if token is paused
	if t.Paused {
		return ErrTokenPaused
	}

	// Check if addresses are blacklisted
	if t.Blacklisted[owner] || t.Blacklisted[operator] {
		return ErrAddressBlacklisted
	}

	// Initialize approvals map if needed
	if t.TokenApprovals[owner] == nil {
		t.TokenApprovals[owner] = make(map[engine.Address]bool)
	}

	// Update approval
	t.TokenApprovals[owner][operator] = approved

	// Record event
	event := ERC1155ApprovalForAllEvent{
		Owner:    owner,
		Operator: operator,
		Approved: approved,
		TxHash:   txHash,
		Block:    block,
		Time:     time.Now(),
	}
	t.ApprovalForAllEvents = append(t.ApprovalForAllEvents, event)

	return nil
}

// SafeTransferFrom transfers a single token using approval
func (t *ERC1155Token) SafeTransferFrom(operator, from, to engine.Address, tokenID uint64, amount *big.Int, data []byte, txHash engine.Hash, block uint64) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Check if token is paused
	if t.Paused {
		return ErrTokenPaused
	}

	// Check if addresses are blacklisted
	if t.Blacklisted[from] || t.Blacklisted[to] || t.Blacklisted[operator] {
		return ErrAddressBlacklisted
	}

	// Validate input
	if amount == nil || amount.Sign() <= 0 {
		return ErrInvalidAmount
	}

	if from == to {
		return ErrSelfTransfer
	}

	// Check if from owns the token
	fromBalance := t.getBalanceInternal(from, tokenID)
	if fromBalance.Cmp(amount) < 0 {
		return ErrInsufficientBalance
	}

	// Check if caller is approved
	if operator != from && !t.isApprovedForAllInternal(from, operator) {
		return ErrUnauthorizedOperation
	}

	// Update balances
	t.updateBalanceInternal(from, tokenID, new(big.Int).Sub(fromBalance, amount))
	t.updateBalanceInternal(to, tokenID, new(big.Int).Add(t.getBalanceInternal(to, tokenID), amount))

	// Record event
	event := ERC1155TransferSingleEvent{
		Operator: operator,
		From:     from,
		To:       to,
		ID:       tokenID,
		Value:    new(big.Int).Set(amount),
		TxHash:   txHash,
		Block:    block,
		Time:     time.Now(),
	}
	t.TransferSingleEvents = append(t.TransferSingleEvents, event)

	return nil
}

// SafeBatchTransferFrom transfers multiple tokens using approval
func (t *ERC1155Token) SafeBatchTransferFrom(operator, from, to engine.Address, tokenIDs []uint64, amounts []*big.Int, data []byte, txHash engine.Hash, block uint64) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Check if token is paused
	if t.Paused {
		return ErrTokenPaused
	}

	// Check if addresses are blacklisted
	if t.Blacklisted[from] || t.Blacklisted[to] || t.Blacklisted[operator] {
		return ErrAddressBlacklisted
	}

	// Validate input
	if len(tokenIDs) != len(amounts) {
		return ErrInvalidBatchTransfer
	}

	if from == to {
		return ErrSelfTransfer
	}

	// Check if caller is approved
	if operator != from && !t.isApprovedForAllInternal(from, operator) {
		return ErrUnauthorizedOperation
	}

	// Validate all transfers first
	for i, tokenID := range tokenIDs {
		amount := amounts[i]
		if amount == nil || amount.Sign() <= 0 {
			return ErrInvalidAmount
		}

		fromBalance := t.getBalanceInternal(from, tokenID)
		if fromBalance.Cmp(amount) < 0 {
			return ErrInsufficientBalance
		}
	}

	// Execute all transfers
	for i, tokenID := range tokenIDs {
		amount := amounts[i]
		fromBalance := t.getBalanceInternal(from, tokenID)

		t.updateBalanceInternal(from, tokenID, new(big.Int).Sub(fromBalance, amount))
		t.updateBalanceInternal(to, tokenID, new(big.Int).Add(t.getBalanceInternal(to, tokenID), amount))
	}

	// Record event
	event := ERC1155TransferBatchEvent{
		Operator: operator,
		From:     from,
		To:       to,
		IDs:      make([]uint64, len(tokenIDs)),
		Values:   make([]*big.Int, len(amounts)),
		TxHash:   txHash,
		Block:    block,
		Time:     time.Now(),
	}
	copy(event.IDs, tokenIDs)
	for i, amount := range amounts {
		event.Values[i] = new(big.Int).Set(amount)
	}
	t.TransferBatchEvents = append(t.TransferBatchEvents, event)

	return nil
}

// Mint creates new tokens and assigns them to the specified address
func (t *ERC1155Token) Mint(to engine.Address, tokenID uint64, amount *big.Int, data []byte, txHash engine.Hash, block uint64) error {
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
	if amount == nil || amount.Sign() <= 0 {
		return ErrInvalidAmount
	}

	// Check max supply if specified
	if maxSupply, exists := t.MaxSupply[tokenID]; exists {
		currentSupply := t.getTotalSupplyInternal(tokenID)
		newSupply := new(big.Int).Add(currentSupply, amount)
		if newSupply.Cmp(maxSupply) > 0 {
			return ErrExceedsMaxSupply
		}
	}

	// Update balance
	currentBalance := t.getBalanceInternal(to, tokenID)
	t.updateBalanceInternal(to, tokenID, new(big.Int).Add(currentBalance, amount))

	// Record event
	event := ERC1155MintEvent{
		To:     to,
		ID:     tokenID,
		Value:  new(big.Int).Set(amount),
		TxHash: txHash,
		Block:  block,
		Time:   time.Now(),
	}
	t.MintEvents = append(t.MintEvents, event)

	return nil
}

// MintBatch creates multiple tokens and assigns them to the specified address
func (t *ERC1155Token) MintBatch(to engine.Address, tokenIDs []uint64, amounts []*big.Int, data []byte, txHash engine.Hash, block uint64) error {
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
	if len(tokenIDs) != len(amounts) {
		return ErrInvalidBatchTransfer
	}

	// Validate all mints first
	for i, tokenID := range tokenIDs {
		amount := amounts[i]
		if amount == nil || amount.Sign() <= 0 {
			return ErrInvalidAmount
		}

		// Check max supply if specified
		if maxSupply, exists := t.MaxSupply[tokenID]; exists {
			currentSupply := t.getTotalSupplyInternal(tokenID)
			newSupply := new(big.Int).Add(currentSupply, amount)
			if newSupply.Cmp(maxSupply) > 0 {
				return ErrExceedsMaxSupply
			}
		}
	}

	// Execute all mints
	for i, tokenID := range tokenIDs {
		amount := amounts[i]
		currentBalance := t.getBalanceInternal(to, tokenID)
		t.updateBalanceInternal(to, tokenID, new(big.Int).Add(currentBalance, amount))
	}

	// Record events
	for i, tokenID := range tokenIDs {
		amount := amounts[i]
		event := ERC1155MintEvent{
			To:     to,
			ID:     tokenID,
			Value:  new(big.Int).Set(amount),
			TxHash: txHash,
			Block:  block,
			Time:   time.Now(),
		}
		t.MintEvents = append(t.MintEvents, event)
	}

	return nil
}

// Burn destroys tokens from the specified address
func (t *ERC1155Token) Burn(from engine.Address, tokenID uint64, amount *big.Int, txHash engine.Hash, block uint64) error {
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
	if amount == nil || amount.Sign() <= 0 {
		return ErrInvalidAmount
	}

	// Check balance
	fromBalance := t.getBalanceInternal(from, tokenID)
	if fromBalance.Cmp(amount) < 0 {
		return ErrInsufficientBalance
	}

	// Update balance
	t.updateBalanceInternal(from, tokenID, new(big.Int).Sub(fromBalance, amount))

	// Record event
	event := ERC1155BurnEvent{
		From:   from,
		ID:     tokenID,
		Value:  new(big.Int).Set(amount),
		TxHash: txHash,
		Block:  block,
		Time:   time.Now(),
	}
	t.BurnEvents = append(t.BurnEvents, event)

	return nil
}

// BurnBatch destroys multiple tokens from the specified address
func (t *ERC1155Token) BurnBatch(from engine.Address, tokenIDs []uint64, amounts []*big.Int, txHash engine.Hash, block uint64) error {
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
	if len(tokenIDs) != len(amounts) {
		return ErrInvalidBatchTransfer
	}

	// Validate all burns first
	for i, tokenID := range tokenIDs {
		amount := amounts[i]
		if amount == nil || amount.Sign() <= 0 {
			return ErrInvalidAmount
		}

		fromBalance := t.getBalanceInternal(from, tokenID)
		if fromBalance.Cmp(amount) < 0 {
			return ErrInsufficientBalance
		}
	}

	// Execute all burns
	for i, tokenID := range tokenIDs {
		amount := amounts[i]
		fromBalance := t.getBalanceInternal(from, tokenID)
		t.updateBalanceInternal(from, tokenID, new(big.Int).Sub(fromBalance, amount))
	}

	// Record events
	for i, tokenID := range tokenIDs {
		amount := amounts[i]
		event := ERC1155BurnEvent{
			From:   from,
			ID:     tokenID,
			Value:  new(big.Int).Set(amount),
			TxHash: txHash,
			Block:  block,
			Time:   time.Now(),
		}
		t.BurnEvents = append(t.BurnEvents, event)
	}

	return nil
}

// SetTokenURI sets the URI for a specific token
func (t *ERC1155Token) SetTokenURI(tokenID uint64, uri string, txHash engine.Hash, block uint64) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Check if token is paused
	if t.Paused {
		return ErrTokenPaused
	}

	// Update token URI
	t.TokenURIs[tokenID] = uri

	// Record event
	event := ERC1155URIEvent{
		URI:    uri,
		ID:     tokenID,
		TxHash: txHash,
		Block:  block,
		Time:   time.Now(),
	}
	t.URIEvents = append(t.URIEvents, event)

	return nil
}

// Pause pauses all token operations
func (t *ERC1155Token) Pause(txHash engine.Hash, block uint64) error {
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
func (t *ERC1155Token) Unpause(txHash engine.Hash, block uint64) error {
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
func (t *ERC1155Token) AddToBlacklist(address engine.Address, txHash engine.Hash, block uint64) error {
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
		Address:     address,
		Blacklisted: true,
		TxHash:      txHash,
		Block:       block,
		Time:        time.Now(),
	}
	t.BlacklistEvents = append(t.BlacklistEvents, event)

	return nil
}

// RemoveFromBlacklist removes an address from the blacklist
func (t *ERC1155Token) RemoveFromBlacklist(address engine.Address, txHash engine.Hash, block uint64) error {
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
		Address:     address,
		Blacklisted: false,
		TxHash:      txHash,
		Block:       block,
		Time:        time.Now(),
	}
	t.BlacklistEvents = append(t.BlacklistEvents, event)

	return nil
}

// GetTransferSingleEvents returns all single transfer events
func (t *ERC1155Token) GetTransferSingleEvents() []ERC1155TransferSingleEvent {
	t.mu.RLock()
	defer t.mu.RUnlock()

	events := make([]ERC1155TransferSingleEvent, len(t.TransferSingleEvents))
	copy(events, t.TransferSingleEvents)
	return events
}

// GetTransferBatchEvents returns all batch transfer events
func (t *ERC1155Token) GetTransferBatchEvents() []ERC1155TransferBatchEvent {
	t.mu.RLock()
	defer t.mu.RUnlock()

	events := make([]ERC1155TransferBatchEvent, len(t.TransferBatchEvents))
	copy(events, t.TransferBatchEvents)
	return events
}

// GetApprovalForAllEvents returns all approval for all events
func (t *ERC1155Token) GetApprovalForAllEvents() []ERC1155ApprovalForAllEvent {
	t.mu.RLock()
	defer t.mu.RUnlock()

	events := make([]ERC1155ApprovalForAllEvent, len(t.ApprovalForAllEvents))
	copy(events, t.ApprovalForAllEvents)
	return events
}

// GetURIEvents returns all URI events
func (t *ERC1155Token) GetURIEvents() []ERC1155URIEvent {
	t.mu.RLock()
	defer t.mu.RUnlock()

	events := make([]ERC1155URIEvent, len(t.URIEvents))
	copy(events, t.URIEvents)
	return events
}

// GetMintEvents returns all mint events
func (t *ERC1155Token) GetMintEvents() []ERC1155MintEvent {
	t.mu.RLock()
	defer t.mu.RUnlock()

	events := make([]ERC1155MintEvent, len(t.MintEvents))
	copy(events, t.MintEvents)
	return events
}

// GetBurnEvents returns all burn events
func (t *ERC1155Token) GetBurnEvents() []ERC1155BurnEvent {
	t.mu.RLock()
	defer t.mu.RUnlock()

	events := make([]ERC1155BurnEvent, len(t.BurnEvents))
	copy(events, t.BurnEvents)
	return events
}

// GetPauseEvents returns all pause events
func (t *ERC1155Token) GetPauseEvents() []PauseEvent {
	t.mu.RLock()
	defer t.mu.RUnlock()

	events := make([]PauseEvent, len(t.PauseEvents))
	copy(events, t.PauseEvents)
	return events
}

// GetBlacklistEvents returns all blacklist events
func (t *ERC1155Token) GetBlacklistEvents() []BlacklistEvent {
	t.mu.RLock()
	defer t.mu.RUnlock()

	events := make([]BlacklistEvent, len(t.BlacklistEvents))
	copy(events, t.BlacklistEvents)
	return events
}

// IsPaused returns whether the token is paused
func (t *ERC1155Token) IsPaused() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Paused
}

// IsBlacklisted returns whether an address is blacklisted
func (t *ERC1155Token) IsBlacklisted(address engine.Address) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Blacklisted[address]
}

// GetTotalSupply returns the total supply for a specific token
func (t *ERC1155Token) GetTotalSupply(tokenID uint64) *big.Int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.getTotalSupplyInternal(tokenID)
}

// GetMaxSupply returns the maximum supply for a specific token
func (t *ERC1155Token) GetMaxSupply(tokenID uint64) *big.Int {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if maxSupply, exists := t.MaxSupply[tokenID]; exists {
		return new(big.Int).Set(maxSupply)
	}
	return nil
}

// IsMintable returns whether the token is mintable
func (t *ERC1155Token) IsMintable() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Mintable
}

// IsBurnable returns whether the token is burnable
func (t *ERC1155Token) IsBurnable() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Burnable
}

// IsPausable returns whether the token is pausable
func (t *ERC1155Token) IsPausable() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Pausable
}

// IsBlacklistable returns whether the token supports blacklisting
func (t *ERC1155Token) IsBlacklistable() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Blacklistable
}

// SetBalanceForTesting sets a balance for testing purposes only
func (t *ERC1155Token) SetBalanceForTesting(address engine.Address, balance *big.Int) {
	// ERC-1155 doesn't have single balances, this method is kept for interface compatibility
}

// SetAllowanceForTesting sets an allowance for testing purposes only
func (t *ERC1155Token) SetAllowanceForTesting(owner, spender engine.Address, allowance *big.Int) {
	// ERC-1155 doesn't have allowances in the traditional sense
	// This method is kept for interface compatibility
}

// Clone creates a deep copy of the token
func (t *ERC1155Token) Clone() *ERC1155Token {
	t.mu.RLock()
	defer t.mu.RUnlock()

	clone := &ERC1155Token{
		URI:                  t.URI,
		Owner:                t.Owner,
		Paused:               t.Paused,
		Blacklisted:          make(map[engine.Address]bool),
		Balances:             make(map[engine.Address]map[uint64]*big.Int),
		TokenApprovals:       make(map[engine.Address]map[engine.Address]bool),
		TokenURIs:            make(map[uint64]string),
		TransferSingleEvents: make([]ERC1155TransferSingleEvent, len(t.TransferSingleEvents)),
		TransferBatchEvents:  make([]ERC1155TransferBatchEvent, len(t.TransferBatchEvents)),
		ApprovalForAllEvents: make([]ERC1155ApprovalForAllEvent, len(t.ApprovalForAllEvents)),
		URIEvents:            make([]ERC1155URIEvent, len(t.URIEvents)),
		MintEvents:           make([]ERC1155MintEvent, len(t.MintEvents)),
		BurnEvents:           make([]ERC1155BurnEvent, len(t.BurnEvents)),
		PauseEvents:          make([]PauseEvent, len(t.PauseEvents)),
		BlacklistEvents:      make([]BlacklistEvent, len(t.BlacklistEvents)),
		MaxSupply:            make(map[uint64]*big.Int),
		Mintable:             t.Mintable,
		Burnable:             t.Burnable,
		Pausable:             t.Pausable,
		Blacklistable:        t.Blacklistable,
		MetadataURI:          t.MetadataURI,
	}

	// Copy blacklisted addresses
	for addr, blacklisted := range t.Blacklisted {
		clone.Blacklisted[addr] = blacklisted
	}

	// Copy balances
	for account, accountBalances := range t.Balances {
		clone.Balances[account] = make(map[uint64]*big.Int)
		for tokenID, balance := range accountBalances {
			clone.Balances[account][tokenID] = new(big.Int).Set(balance)
		}
	}

	// Copy token approvals
	for owner, approvals := range t.TokenApprovals {
		clone.TokenApprovals[owner] = make(map[engine.Address]bool)
		for operator, approved := range approvals {
			clone.TokenApprovals[owner][operator] = approved
		}
	}

	// Copy token URIs
	for tokenID, uri := range t.TokenURIs {
		clone.TokenURIs[tokenID] = uri
	}

	// Copy max supply
	for tokenID, maxSupply := range t.MaxSupply {
		clone.MaxSupply[tokenID] = new(big.Int).Set(maxSupply)
	}

	// Copy events
	copy(clone.TransferSingleEvents, t.TransferSingleEvents)
	copy(clone.TransferBatchEvents, t.TransferBatchEvents)
	copy(clone.ApprovalForAllEvents, t.ApprovalForAllEvents)
	copy(clone.URIEvents, t.URIEvents)
	copy(clone.MintEvents, t.MintEvents)
	copy(clone.BurnEvents, t.BurnEvents)
	copy(clone.PauseEvents, t.PauseEvents)
	copy(clone.BlacklistEvents, t.BlacklistEvents)

	return clone
}

// Internal helper methods (not thread-safe, caller must hold lock)
func (t *ERC1155Token) getBalanceInternal(account engine.Address, tokenID uint64) *big.Int {
	if accountBalances, exists := t.Balances[account]; exists {
		if balance, exists := accountBalances[tokenID]; exists {
			return balance
		}
	}
	return big.NewInt(0)
}

func (t *ERC1155Token) updateBalanceInternal(account engine.Address, tokenID uint64, newBalance *big.Int) {
	if t.Balances[account] == nil {
		t.Balances[account] = make(map[uint64]*big.Int)
	}
	t.Balances[account][tokenID] = newBalance
}

func (t *ERC1155Token) isApprovedForAllInternal(owner, operator engine.Address) bool {
	if operatorApprovals, exists := t.TokenApprovals[owner]; exists {
		return operatorApprovals[operator]
	}
	return false
}

func (t *ERC1155Token) getTotalSupplyInternal(tokenID uint64) *big.Int {
	total := big.NewInt(0)
	for _, accountBalances := range t.Balances {
		if balance, exists := accountBalances[tokenID]; exists {
			total.Add(total, balance)
		}
	}
	return total
}

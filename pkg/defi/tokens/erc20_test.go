package tokens

import (
	"math/big"
	"testing"

	"github.com/gochain/gochain/pkg/contracts/engine"
)

// Helper function to generate random addresses
func generateRandomAddress() engine.Address {
	// Use a simple counter to generate unique addresses for testing
	staticCounter++
	return engine.Address{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, byte(staticCounter)}
}

var staticCounter byte = 0

// Helper function to generate random hashes
func generateRandomHash() engine.Hash {
	return engine.Hash{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20}
}

// TestNewERC20Token tests ERC-20 token creation
func TestNewERC20Token(t *testing.T) {
	name := "Test Token"
	symbol := "TEST"
	decimals := uint8(18)
	totalSupply := big.NewInt(1000) // 1000 tokens
	owner := generateRandomAddress()
	config := DefaultTokenConfig()

	token := NewERC20Token(name, symbol, decimals, totalSupply, owner, config)

	if token == nil {
		t.Fatal("expected non-nil token")
	}

	if token.GetName() != name {
		t.Errorf("expected name %s, got %s", name, token.GetName())
	}

	if token.GetSymbol() != symbol {
		t.Errorf("expected symbol %s, got %s", symbol, token.GetSymbol())
	}

	if token.GetDecimals() != decimals {
		t.Errorf("expected decimals %d, got %d", decimals, token.GetDecimals())
	}

	if token.GetTotalSupply().Cmp(totalSupply) != 0 {
		t.Errorf("expected total supply %s, got %s", totalSupply.String(), token.GetTotalSupply().String())
	}

	if token.GetOwner() != owner {
		t.Errorf("expected owner %v, got %v", owner, token.GetOwner())
	}

	// Check initial balance of owner
	ownerBalance := token.GetBalance(owner)
	if ownerBalance.Cmp(totalSupply) != 0 {
		t.Errorf("expected owner balance %s, got %s", totalSupply.String(), ownerBalance.String())
	}
}

// TestERC20TokenTransfer tests basic token transfer functionality
func TestERC20TokenTransfer(t *testing.T) {
	token := NewERC20Token("Test", "TEST", 18, big.NewInt(1000), generateRandomAddress(), DefaultTokenConfig())
	from := generateRandomAddress()
	to := generateRandomAddress()
	amount := big.NewInt(100)
	txHash := generateRandomHash()
	block := uint64(12345)

	// Set initial balance for from address
	token.SetBalanceForTesting(from, big.NewInt(500))

	// Test successful transfer
	err := token.Transfer(from, to, amount, txHash, block)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check balances
	fromBalance := token.GetBalance(from)
	expectedFromBalance := big.NewInt(400) // 500 - 100
	if fromBalance.Cmp(expectedFromBalance) != 0 {
		t.Errorf("expected from balance %s, got %s", expectedFromBalance.String(), fromBalance.String())
	}

	toBalance := token.GetBalance(to)
	expectedToBalance := big.NewInt(100)
	if toBalance.Cmp(expectedToBalance) != 0 {
		t.Errorf("expected to balance %s, got %s", expectedToBalance.String(), toBalance.String())
	}

	// Check transfer events
	events := token.GetTransferEvents()
	if len(events) != 1 {
		t.Errorf("expected 1 transfer event, got %d", len(events))
	}

	event := events[0]
	if event.From != from {
		t.Errorf("expected event from %v, got %v", from, event.From)
	}
	if event.To != to {
		t.Errorf("expected event to %v, got %v", to, event.To)
	}
	if event.Value.Cmp(amount) != 0 {
		t.Errorf("expected event value %s, got %s", amount.String(), event.Value.String())
	}
}

// TestERC20TokenTransferInsufficientBalance tests transfer with insufficient balance
func TestERC20TokenTransferInsufficientBalance(t *testing.T) {
	token := NewERC20Token("Test", "TEST", 18, big.NewInt(1000), generateRandomAddress(), DefaultTokenConfig())
	from := generateRandomAddress()
	to := generateRandomAddress()
	amount := big.NewInt(100)
	txHash := generateRandomHash()
	block := uint64(12345)

	// Set insufficient balance for from address
	token.SetBalanceForTesting(from, big.NewInt(50))

	// Test transfer with insufficient balance
	err := token.Transfer(from, to, amount, txHash, block)
	if err != ErrInsufficientBalance {
		t.Errorf("expected ErrInsufficientBalance, got %v", err)
	}

	// Check that balances remain unchanged
	fromBalance := token.GetBalance(from)
	expectedFromBalance := big.NewInt(50)
	if fromBalance.Cmp(expectedFromBalance) != 0 {
		t.Errorf("expected from balance %s, got %s", expectedFromBalance.String(), fromBalance.String())
	}

	toBalance := token.GetBalance(to)
	expectedToBalance := big.NewInt(0)
	if toBalance.Cmp(expectedToBalance) != 0 {
		t.Errorf("expected to balance %s, got %s", expectedToBalance.String(), toBalance.String())
	}
}

// TestERC20TokenTransferSelf tests transfer to self (should fail)
func TestERC20TokenTransferSelf(t *testing.T) {
	token := NewERC20Token("Test", "TEST", 18, big.NewInt(1000), generateRandomAddress(), DefaultTokenConfig())
	from := generateRandomAddress()
	amount := big.NewInt(100)
	txHash := generateRandomHash()
	block := uint64(12345)

	// Set initial balance
	token.SetBalanceForTesting(from, big.NewInt(500))

	// Test transfer to self
	err := token.Transfer(from, from, amount, txHash, block)
	if err != ErrSelfTransfer {
		t.Errorf("expected ErrSelfTransfer, got %v", err)
	}
}

// TestERC20TokenTransferInvalidAmount tests transfer with invalid amount
func TestERC20TokenTransferInvalidAmount(t *testing.T) {
	token := NewERC20Token("Test", "TEST", 18, big.NewInt(1000), generateRandomAddress(), DefaultTokenConfig())
	from := generateRandomAddress()
	to := generateRandomAddress()
	txHash := generateRandomHash()
	block := uint64(12345)

	// Test transfer with nil amount
	err := token.Transfer(from, to, nil, txHash, block)
	if err != ErrInvalidAmount {
		t.Errorf("expected ErrInvalidAmount, got %v", err)
	}

	// Test transfer with zero amount
	err = token.Transfer(from, to, big.NewInt(0), txHash, block)
	if err != ErrInvalidAmount {
		t.Errorf("expected ErrInvalidAmount, got %v", err)
	}

	// Test transfer with negative amount
	err = token.Transfer(from, to, big.NewInt(-100), txHash, block)
	if err != ErrInvalidAmount {
		t.Errorf("expected ErrInvalidAmount, got %v", err)
	}
}

// TestERC20TokenApprove tests token approval functionality
func TestERC20TokenApprove(t *testing.T) {
	token := NewERC20Token("Test", "TEST", 18, big.NewInt(1000), generateRandomAddress(), DefaultTokenConfig())
	owner := generateRandomAddress()
	spender := generateRandomAddress()
	amount := big.NewInt(100)
	txHash := generateRandomHash()
	block := uint64(12345)

	// Test approval
	err := token.Approve(owner, spender, amount, txHash, block)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check allowance
	allowance := token.GetAllowance(owner, spender)
	if allowance.Cmp(amount) != 0 {
		t.Errorf("expected allowance %s, got %s", amount.String(), allowance.String())
	}

	// Check approval events
	events := token.GetApprovalEvents()
	if len(events) != 1 {
		t.Errorf("expected 1 approval event, got %d", len(events))
	}

	event := events[0]
	if event.Owner != owner {
		t.Errorf("expected event owner %v, got %v", owner, event.Owner)
	}
	if event.Spender != spender {
		t.Errorf("expected event spender %v, got %v", spender, event.Spender)
	}
	if event.Value.Cmp(amount) != 0 {
		t.Errorf("expected event value %s, got %s", amount.String(), event.Value.String())
	}
}

// TestERC20TokenTransferFrom tests transfer using allowance
func TestERC20TokenTransferFrom(t *testing.T) {
	token := NewERC20Token("Test", "TEST", 18, big.NewInt(1000), generateRandomAddress(), DefaultTokenConfig())
	owner := generateRandomAddress()
	spender := generateRandomAddress()
	to := generateRandomAddress()
	amount := big.NewInt(100)
	txHash := generateRandomHash()
	block := uint64(12345)

	// Set initial balance and allowance
	token.SetBalanceForTesting(owner, big.NewInt(500))
	token.SetAllowanceForTesting(owner, spender, big.NewInt(150))

	// Test transfer from
	err := token.TransferFrom(spender, owner, to, amount, txHash, block)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check balances
	ownerBalance := token.GetBalance(owner)
	expectedOwnerBalance := big.NewInt(400) // 500 - 100
	if ownerBalance.Cmp(expectedOwnerBalance) != 0 {
		t.Errorf("expected owner balance %s, got %s", expectedOwnerBalance.String(), ownerBalance.String())
	}

	toBalance := token.GetBalance(to)
	expectedToBalance := big.NewInt(100)
	if toBalance.Cmp(expectedToBalance) != 0 {
		t.Errorf("expected to balance %s, got %s", expectedToBalance.String(), toBalance.String())
	}

	// Check allowance is reduced
	allowance := token.GetAllowance(owner, spender)
	expectedAllowance := big.NewInt(50) // 150 - 100
	if allowance.Cmp(expectedAllowance) != 0 {
		t.Errorf("expected allowance %s, got %s", expectedAllowance.String(), allowance.String())
	}
}

// TestERC20TokenTransferFromInsufficientAllowance tests transfer with insufficient allowance
func TestERC20TokenTransferFromInsufficientAllowance(t *testing.T) {
	token := NewERC20Token("Test", "TEST", 18, big.NewInt(1000), generateRandomAddress(), DefaultTokenConfig())
	owner := generateRandomAddress()
	spender := generateRandomAddress()
	to := generateRandomAddress()
	amount := big.NewInt(100)
	txHash := generateRandomHash()
	block := uint64(12345)

	// Set initial balance and insufficient allowance
	token.Balances[owner] = big.NewInt(500)
	token.Allowances[owner] = make(map[engine.Address]*big.Int)
	token.Allowances[owner][spender] = big.NewInt(50)

	// Test transfer from with insufficient allowance
	err := token.TransferFrom(spender, owner, to, amount, txHash, block)
	if err != ErrInsufficientAllowance {
		t.Errorf("expected ErrInsufficientAllowance, got %v", err)
	}
}

// TestERC20TokenMint tests token minting functionality
func TestERC20TokenMint(t *testing.T) {
	config := DefaultTokenConfig()
	config.Mintable = true
	token := NewERC20Token("Test", "TEST", 18, big.NewInt(1000), generateRandomAddress(), config)
	to := generateRandomAddress()
	amount := big.NewInt(100)
	txHash := generateRandomHash()
	block := uint64(12345)

	// Test minting
	err := token.Mint(to, amount, txHash, block)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check balance
	toBalance := token.GetBalance(to)
	expectedToBalance := big.NewInt(100)
	if toBalance.Cmp(expectedToBalance) != 0 {
		t.Errorf("expected to balance %s, got %s", expectedToBalance.String(), toBalance.String())
	}

	// Check total supply
	totalSupply := token.GetTotalSupply()
	expectedTotalSupply := big.NewInt(1100) // 1000 + 100
	if totalSupply.Cmp(expectedTotalSupply) != 0 {
		t.Errorf("expected total supply %s, got %s", expectedTotalSupply.String(), totalSupply.String())
	}

	// Check mint events
	events := token.GetMintEvents()
	if len(events) != 1 {
		t.Errorf("expected 1 mint event, got %d", len(events))
	}
}

// TestERC20TokenMintNotAllowed tests minting when not allowed
func TestERC20TokenMintNotAllowed(t *testing.T) {
	token := NewERC20Token("Test", "TEST", 18, big.NewInt(1000), generateRandomAddress(), DefaultTokenConfig())
	to := generateRandomAddress()
	amount := big.NewInt(100)
	txHash := generateRandomHash()
	block := uint64(12345)

	// Test minting when not allowed
	err := token.Mint(to, amount, txHash, block)
	if err != ErrMintingNotAllowed {
		t.Errorf("expected ErrMintingNotAllowed, got %v", err)
	}
}

// TestERC20TokenBurn tests token burning functionality
func TestERC20TokenBurn(t *testing.T) {
	config := DefaultTokenConfig()
	config.Burnable = true
	token := NewERC20Token("Test", "TEST", 18, big.NewInt(1000), generateRandomAddress(), config)
	from := generateRandomAddress()
	amount := big.NewInt(100)
	txHash := generateRandomHash()
	block := uint64(12345)

	// Set initial balance
	token.Balances[from] = big.NewInt(500)

	// Test burning
	err := token.Burn(from, amount, txHash, block)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check balance
	fromBalance := token.GetBalance(from)
	expectedFromBalance := big.NewInt(400) // 500 - 100
	if fromBalance.Cmp(expectedFromBalance) != 0 {
		t.Errorf("expected from balance %s, got %s", expectedFromBalance.String(), fromBalance.String())
	}

	// Check total supply
	totalSupply := token.GetTotalSupply()
	expectedTotalSupply := big.NewInt(900) // 1000 - 100
	if totalSupply.Cmp(expectedTotalSupply) != 0 {
		t.Errorf("expected total supply %s, got %s", expectedTotalSupply.String(), totalSupply.String())
	}

	// Check burn events
	events := token.GetBurnEvents()
	if len(events) != 1 {
		t.Errorf("expected 1 burn event, got %d", len(events))
	}
}

// TestERC20TokenBurnNotAllowed tests burning when not allowed
func TestERC20TokenBurnNotAllowed(t *testing.T) {
	token := NewERC20Token("Test", "TEST", 18, big.NewInt(1000), generateRandomAddress(), DefaultTokenConfig())
	from := generateRandomAddress()
	amount := big.NewInt(100)
	txHash := generateRandomHash()
	block := uint64(12345)

	// Set initial balance
	token.SetBalanceForTesting(from, big.NewInt(500))

	// Test burning when not allowed
	err := token.Burn(from, amount, txHash, block)
	if err != ErrBurningNotAllowed {
		t.Errorf("expected ErrBurningNotAllowed, got %v", err)
	}
}

// TestERC20TokenPause tests token pausing functionality
func TestERC20TokenPause(t *testing.T) {
	config := DefaultTokenConfig()
	config.Pausable = true
	token := NewERC20Token("Test", "TEST", 18, big.NewInt(1000), generateRandomAddress(), config)
	txHash := generateRandomHash()
	block := uint64(12345)

	// Test pausing
	err := token.Pause(txHash, block)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check if paused
	if !token.IsPaused() {
		t.Error("expected token to be paused")
	}

	// Check pause events
	events := token.GetPauseEvents()
	if len(events) != 1 {
		t.Errorf("expected 1 pause event, got %d", len(events))
	}

	event := events[0]
	if !event.Paused {
		t.Error("expected event to show paused")
	}
}

// TestERC20TokenUnpause tests token unpausing functionality
func TestERC20TokenUnpause(t *testing.T) {
	config := DefaultTokenConfig()
	config.Pausable = true
	token := NewERC20Token("Test", "TEST", 18, big.NewInt(1000), generateRandomAddress(), config)
	txHash := generateRandomHash()
	block := uint64(12345)

	// First pause the token
	err := token.Pause(txHash, block)
	if err != nil {
		t.Fatalf("failed to pause token: %v", err)
	}

	// Test unpausing
	err = token.Unpause(txHash, block)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check if unpaused
	if token.IsPaused() {
		t.Error("expected token to be unpaused")
	}

	// Check pause events
	events := token.GetPauseEvents()
	if len(events) != 2 {
		t.Errorf("expected 2 pause events, got %d", len(events))
	}

	lastEvent := events[len(events)-1]
	if lastEvent.Paused {
		t.Error("expected last event to show unpaused")
	}
}

// TestERC20TokenBlacklist tests token blacklisting functionality
func TestERC20TokenBlacklist(t *testing.T) {
	config := DefaultTokenConfig()
	config.Blacklistable = true
	token := NewERC20Token("Test", "TEST", 18, big.NewInt(1000), generateRandomAddress(), config)
	address := generateRandomAddress()
	txHash := generateRandomHash()
	block := uint64(12345)

	// Test adding to blacklist
	err := token.AddToBlacklist(address, txHash, block)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check if blacklisted
	if !token.IsBlacklisted(address) {
		t.Error("expected address to be blacklisted")
	}

	// Check blacklist events
	events := token.GetBlacklistEvents()
	if len(events) != 1 {
		t.Errorf("expected 1 blacklist event, got %d", len(events))
	}

	event := events[0]
	if !event.Blacklisted {
		t.Error("expected event to show blacklisted")
	}

	// Test removing from blacklist
	err = token.RemoveFromBlacklist(address, txHash, block)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check if removed from blacklist
	if token.IsBlacklisted(address) {
		t.Error("expected address to not be blacklisted")
	}
}

// TestERC20TokenTransferFee tests transfer fee functionality
func TestERC20TokenTransferFee(t *testing.T) {
	config := DefaultTokenConfig()
	config.TransferFee = big.NewInt(100) // 1% fee (100 basis points)
	config.TransferFeeRecipient = generateRandomAddress()
	token := NewERC20Token("Test", "TEST", 18, big.NewInt(1000), generateRandomAddress(), config)
	from := generateRandomAddress()
	to := generateRandomAddress()
	amount := big.NewInt(1000) // 1000 tokens
	txHash := generateRandomHash()
	block := uint64(12345)

	// Set initial balances
	token.Balances[from] = big.NewInt(2000)
	token.Balances[config.TransferFeeRecipient] = big.NewInt(0)

	// Test transfer with fee
	err := token.Transfer(from, to, amount, txHash, block)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check balances
	fromBalance := token.GetBalance(from)
	expectedFromBalance := big.NewInt(1000) // 2000 - 1000
	if fromBalance.Cmp(expectedFromBalance) != 0 {
		t.Errorf("expected from balance %s, got %s", expectedFromBalance.String(), fromBalance.String())
	}

	toBalance := token.GetBalance(to)
	expectedToBalance := big.NewInt(990) // 1000 - 10 (1% fee)
	if toBalance.Cmp(expectedToBalance) != 0 {
		t.Errorf("expected to balance %s, got %s", expectedToBalance.String(), toBalance.String())
	}

	feeRecipientBalance := token.GetBalance(config.TransferFeeRecipient)
	expectedFeeRecipientBalance := big.NewInt(10) // 1% of 1000
	if feeRecipientBalance.Cmp(expectedFeeRecipientBalance) != 0 {
		t.Errorf("expected fee recipient balance %s, got %s", expectedFeeRecipientBalance.String(), feeRecipientBalance.String())
	}
}

// TestERC20TokenClone tests token cloning functionality
func TestERC20TokenClone(t *testing.T) {
	config := DefaultTokenConfig()
	config.Mintable = true
	config.Burnable = true
	config.Pausable = true
	config.Blacklistable = true
	token := NewERC20Token("Test", "TEST", 18, big.NewInt(1000), generateRandomAddress(), config)

	// Add some state
	address1 := generateRandomAddress()
	address2 := generateRandomAddress()
	token.Balances[address1] = big.NewInt(100)
	token.Allowances[address1] = make(map[engine.Address]*big.Int)
	token.Allowances[address1][address2] = big.NewInt(50)

	// Clone the token
	clone := token.Clone()

	if clone == token {
		t.Error("expected clone to be different from original")
	}

	// Check that values are copied
	if clone.GetName() != token.GetName() {
		t.Error("expected clone name to match original")
	}
	if clone.GetSymbol() != token.GetSymbol() {
		t.Error("expected clone symbol to match original")
	}
	if clone.GetDecimals() != token.GetDecimals() {
		t.Error("expected clone decimals to match original")
	}
	if clone.GetTotalSupply().Cmp(token.GetTotalSupply()) != 0 {
		t.Error("expected clone total supply to match original")
	}

	// Check that balances are copied
	if clone.GetBalance(address1).Cmp(token.GetBalance(address1)) != 0 {
		t.Error("expected clone balance to match original")
	}

	// Check that allowances are copied
	if clone.GetAllowance(address1, address2).Cmp(token.GetAllowance(address1, address2)) != 0 {
		t.Error("expected clone allowance to match original")
	}

	// Modify clone and verify original is unchanged
	clone.Balances[address1] = big.NewInt(200)
	if token.GetBalance(address1).Cmp(big.NewInt(100)) != 0 {
		t.Error("expected original balance to remain unchanged")
	}
}

// TestERC20TokenConcurrency tests concurrent access to token
func TestERC20TokenConcurrency(t *testing.T) {
	token := NewERC20Token("Test", "TEST", 18, big.NewInt(1000), generateRandomAddress(), DefaultTokenConfig())
	
	// Test concurrent reads
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = token.GetName()
				_ = token.GetSymbol()
				_ = token.GetDecimals()
				_ = token.GetTotalSupply()
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

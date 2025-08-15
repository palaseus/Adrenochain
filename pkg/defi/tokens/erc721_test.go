package tokens

import (
	"math/big"
	"testing"

	"github.com/gochain/gochain/pkg/contracts/engine"
)

// TestNewERC721Token tests ERC-721 token creation
func TestNewERC721Token(t *testing.T) {
	name := "Test NFT"
	symbol := "TNFT"
	baseURI := "https://api.example.com/metadata/"
	owner := generateRandomAddress()
	config := DefaultERC721TokenConfig()

	token := NewERC721Token(name, symbol, baseURI, owner, config)

	if token == nil {
		t.Fatal("expected non-nil token")
	}

	if token.GetName() != name {
		t.Errorf("expected name %s, got %s", name, token.GetName())
	}

	if token.GetSymbol() != symbol {
		t.Errorf("expected symbol %s, got %s", symbol, token.GetSymbol())
	}

	if token.GetBaseURI() != baseURI {
		t.Errorf("expected base URI %s, got %s", baseURI, token.GetBaseURI())
	}

	if token.GetTotalSupply().Cmp(big.NewInt(0)) != 0 {
		t.Errorf("expected total supply 0, got %s", token.GetTotalSupply().String())
	}

	if token.GetOwner() != owner {
		t.Errorf("expected owner %v, got %v", owner, token.GetOwner())
	}
}

// TestERC721TokenMint tests token minting functionality
func TestERC721TokenMint(t *testing.T) {
	config := DefaultERC721TokenConfig()
	config.Mintable = true
	token := NewERC721Token("Test NFT", "TNFT", "https://api.example.com/metadata/", generateRandomAddress(), config)
	to := generateRandomAddress()
	tokenID := uint64(1)
	txHash := generateRandomHash()
	block := uint64(12345)

	// Test minting
	err := token.Mint(to, tokenID, txHash, block)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check token ownership
	owner := token.GetOwnerOf(tokenID)
	if owner != to {
		t.Errorf("expected token owner %v, got %v", to, owner)
	}

	// Check total supply
	totalSupply := token.GetTotalSupply()
	if totalSupply.Cmp(big.NewInt(1)) != 0 {
		t.Errorf("expected total supply 1, got %s", totalSupply.String())
	}

	// Check balance of owner
	balance := token.GetBalanceOf(to)
	if balance != 1 {
		t.Errorf("expected balance 1, got %d", balance)
	}

	// Check tokens of owner
	tokens := token.GetTokensOfOwner(to)
	if len(tokens) != 1 || tokens[0] != tokenID {
		t.Errorf("expected tokens [%d], got %v", tokenID, tokens)
	}

	// Check mint events
	events := token.GetMintEvents()
	if len(events) != 1 {
		t.Errorf("expected 1 mint event, got %d", len(events))
	}

	event := events[0]
	if event.To != to {
		t.Errorf("expected event to %v, got %v", to, event.To)
	}
	if event.TokenID != tokenID {
		t.Errorf("expected event token ID %d, got %d", tokenID, event.TokenID)
	}
}

// TestERC721TokenMintNotAllowed tests minting when not allowed
func TestERC721TokenMintNotAllowed(t *testing.T) {
	token := NewERC721Token("Test NFT", "TNFT", "https://api.example.com/metadata/", generateRandomAddress(), DefaultERC721TokenConfig())
	to := generateRandomAddress()
	tokenID := uint64(1)
	txHash := generateRandomHash()
	block := uint64(12345)

	// Test minting when not allowed
	err := token.Mint(to, tokenID, txHash, block)
	if err != ErrMintingNotAllowed {
		t.Errorf("expected ErrMintingNotAllowed, got %v", err)
	}
}

// TestERC721TokenMintDuplicateID tests minting with duplicate token ID
func TestERC721TokenMintDuplicateID(t *testing.T) {
	config := DefaultERC721TokenConfig()
	config.Mintable = true
	token := NewERC721Token("Test NFT", "TNFT", "https://api.example.com/metadata/", generateRandomAddress(), config)
	to := generateRandomAddress()
	tokenID := uint64(1)
	txHash := generateRandomHash()
	block := uint64(12345)

	// First mint
	err := token.Mint(to, tokenID, txHash, block)
	if err != nil {
		t.Fatalf("failed to mint first token: %v", err)
	}

	// Try to mint with same ID
	err = token.Mint(to, tokenID, txHash, block)
	if err != ErrTokenAlreadyExists {
		t.Errorf("expected ErrTokenAlreadyExists, got %v", err)
	}
}

// TestERC721TokenTransfer tests token transfer functionality
func TestERC721TokenTransfer(t *testing.T) {
	config := DefaultERC721TokenConfig()
	config.Mintable = true
	token := NewERC721Token("Test NFT", "TNFT", "https://api.example.com/metadata/", generateRandomAddress(), config)
	from := generateRandomAddress()
	to := generateRandomAddress()
	tokenID := uint64(1)
	txHash := generateRandomHash()
	block := uint64(12345)

	// Mint token to from address
	err := token.Mint(from, tokenID, txHash, block)
	if err != nil {
		t.Fatalf("failed to mint token: %v", err)
	}

	// Test transfer
	err = token.Transfer(from, to, tokenID, txHash, block)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check token ownership
	owner := token.GetOwnerOf(tokenID)
	if owner != to {
		t.Errorf("expected token owner %v, got %v", to, owner)
	}

	// Check balances
	fromBalance := token.GetBalanceOf(from)
	if fromBalance != 0 {
		t.Errorf("expected from balance 0, got %d", fromBalance)
	}

	toBalance := token.GetBalanceOf(to)
	if toBalance != 1 {
		t.Errorf("expected to balance 1, got %d", toBalance)
	}

	// Check transfer events
	events := token.GetTransferEvents()
	if len(events) != 2 { // 1 mint + 1 transfer
		t.Errorf("expected 2 transfer events, got %d", len(events))
	}

	// Check that approval was cleared
	approved := token.GetTokenApproval(tokenID)
	if approved != (engine.Address{}) {
		t.Error("expected approval to be cleared after transfer")
	}
}

// TestERC721TokenTransferSelf tests transfer to self (should fail)
func TestERC721TokenTransferSelf(t *testing.T) {
	config := DefaultERC721TokenConfig()
	config.Mintable = true
	token := NewERC721Token("Test NFT", "TNFT", "https://api.example.com/metadata/", generateRandomAddress(), config)
	from := generateRandomAddress()
	tokenID := uint64(1)
	txHash := generateRandomHash()
	block := uint64(12345)

	// Mint token to from address
	err := token.Mint(from, tokenID, txHash, block)
	if err != nil {
		t.Fatalf("failed to mint token: %v", err)
	}

	// Test transfer to self
	err = token.Transfer(from, from, tokenID, txHash, block)
	if err != ErrSelfTransfer {
		t.Errorf("expected ErrSelfTransfer, got %v", err)
	}
}

// TestERC721TokenTransferUnauthorized tests transfer without ownership
func TestERC721TokenTransferUnauthorized(t *testing.T) {
	config := DefaultERC721TokenConfig()
	config.Mintable = true
	token := NewERC721Token("Test NFT", "TNFT", "https://api.example.com/metadata/", generateRandomAddress(), config)
	owner := generateRandomAddress()
	unauthorized := generateRandomAddress()
	to := generateRandomAddress()
	tokenID := uint64(1)
	txHash := generateRandomHash()
	block := uint64(12345)

	// Mint token to owner
	err := token.Mint(owner, tokenID, txHash, block)
	if err != nil {
		t.Fatalf("failed to mint token: %v", err)
	}

	// Try to transfer from unauthorized address
	err = token.Transfer(unauthorized, to, tokenID, txHash, block)
	if err != ErrUnauthorizedOperation {
		t.Errorf("expected ErrUnauthorizedOperation, got %v", err)
	}
}

// TestERC721TokenApprove tests token approval functionality
func TestERC721TokenApprove(t *testing.T) {
	config := DefaultERC721TokenConfig()
	config.Mintable = true
	token := NewERC721Token("Test NFT", "TNFT", "https://api.example.com/metadata/", generateRandomAddress(), config)
	owner := generateRandomAddress()
	approved := generateRandomAddress()
	tokenID := uint64(1)
	txHash := generateRandomHash()
	block := uint64(12345)

	// Mint token to owner
	err := token.Mint(owner, tokenID, txHash, block)
	if err != nil {
		t.Fatalf("failed to mint token: %v", err)
	}

	// Test approval
	err = token.Approve(owner, approved, tokenID, txHash, block)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check approval
	tokenApproved := token.GetTokenApproval(tokenID)
	if tokenApproved != approved {
		t.Errorf("expected approved address %v, got %v", approved, tokenApproved)
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
	if event.Approved != approved {
		t.Errorf("expected event approved %v, got %v", approved, event.Approved)
	}
	if event.TokenID != tokenID {
		t.Errorf("expected event token ID %d, got %d", tokenID, event.TokenID)
	}
}

// TestERC721TokenApproveUnauthorized tests approval without ownership
func TestERC721TokenApproveUnauthorized(t *testing.T) {
	config := DefaultERC721TokenConfig()
	config.Mintable = true
	token := NewERC721Token("Test NFT", "TNFT", "https://api.example.com/metadata/", generateRandomAddress(), config)
	owner := generateRandomAddress()
	unauthorized := generateRandomAddress()
	approved := generateRandomAddress()
	tokenID := uint64(1)
	txHash := generateRandomHash()
	block := uint64(12345)

	// Mint token to owner
	err := token.Mint(owner, tokenID, txHash, block)
	if err != nil {
		t.Fatalf("failed to mint token: %v", err)
	}

	// Try to approve from unauthorized address
	err = token.Approve(unauthorized, approved, tokenID, txHash, block)
	if err != ErrUnauthorizedOperation {
		t.Errorf("expected ErrUnauthorizedOperation, got %v", err)
	}
}

// TestERC721TokenSetApprovalForAll tests operator approval functionality
func TestERC721TokenSetApprovalForAll(t *testing.T) {
	token := NewERC721Token("Test NFT", "TNFT", "https://api.example.com/metadata/", generateRandomAddress(), DefaultERC721TokenConfig())
	owner := generateRandomAddress()
	operator := generateRandomAddress()
	txHash := generateRandomHash()
	block := uint64(12345)

	// Test approval for all
	err := token.SetApprovalForAll(owner, operator, true, txHash, block)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check approval
	approved := token.IsApprovedForAll(owner, operator)
	if !approved {
		t.Error("expected operator to be approved for all")
	}

	// Test revocation
	err = token.SetApprovalForAll(owner, operator, false, txHash, block)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check approval is revoked
	approved = token.IsApprovedForAll(owner, operator)
	if approved {
		t.Error("expected operator to not be approved for all")
	}

	// Check approval for all events
	events := token.GetApprovalForAllEvents()
	if len(events) != 2 {
		t.Errorf("expected 2 approval for all events, got %d", len(events))
	}
}

// TestERC721TokenTransferFrom tests transfer using approval
func TestERC721TokenTransferFrom(t *testing.T) {
	config := DefaultERC721TokenConfig()
	config.Mintable = true
	token := NewERC721Token("Test NFT", "TNFT", "https://api.example.com/metadata/", generateRandomAddress(), config)
	owner := generateRandomAddress()
	operator := generateRandomAddress()
	to := generateRandomAddress()
	tokenID := uint64(1)
	txHash := generateRandomHash()
	block := uint64(12345)

	// Mint token to owner
	err := token.Mint(owner, tokenID, txHash, block)
	if err != nil {
		t.Fatalf("failed to mint token: %v", err)
	}

	// Approve operator for all tokens
	err = token.SetApprovalForAll(owner, operator, true, txHash, block)
	if err != nil {
		t.Fatalf("failed to approve operator: %v", err)
	}

	// Test transfer from using operator
	err = token.TransferFrom(owner, to, tokenID, txHash, block)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check token ownership
	newOwner := token.GetOwnerOf(tokenID)
	if newOwner != to {
		t.Errorf("expected token owner %v, got %v", to, newOwner)
	}
}

// TestERC721TokenBurn tests token burning functionality
func TestERC721TokenBurn(t *testing.T) {
	config := DefaultERC721TokenConfig()
	config.Mintable = true
	config.Burnable = true
	token := NewERC721Token("Test NFT", "TNFT", "https://api.example.com/metadata/", generateRandomAddress(), config)
	owner := generateRandomAddress()
	tokenID := uint64(1)
	txHash := generateRandomHash()
	block := uint64(12345)

	// Mint token to owner
	err := token.Mint(owner, tokenID, txHash, block)
	if err != nil {
		t.Fatalf("failed to mint token: %v", err)
	}

	// Test burning
	err = token.Burn(tokenID, txHash, block)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check token no longer exists
	burnedOwner := token.GetOwnerOf(tokenID)
	if burnedOwner != (engine.Address{}) {
		t.Error("expected token to not exist after burning")
	}

	// Check total supply
	totalSupply := token.GetTotalSupply()
	if totalSupply.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("expected total supply 0, got %s", totalSupply.String())
	}

	// Check burn events
	events := token.GetBurnEvents()
	if len(events) != 1 {
		t.Errorf("expected 1 burn event, got %d", len(events))
	}

	event := events[0]
	if event.From != owner {
		t.Errorf("expected event from %v, got %v", owner, event.From)
	}
	if event.TokenID != tokenID {
		t.Errorf("expected event token ID %d, got %d", tokenID, event.TokenID)
	}
}

// TestERC721TokenBurnNotAllowed tests burning when not allowed
func TestERC721TokenBurnNotAllowed(t *testing.T) {
	config := DefaultERC721TokenConfig()
	config.Mintable = true
	token := NewERC721Token("Test NFT", "TNFT", "https://api.example.com/metadata/", generateRandomAddress(), config)
	owner := generateRandomAddress()
	tokenID := uint64(1)
	txHash := generateRandomHash()
	block := uint64(12345)

	// Mint token to owner
	err := token.Mint(owner, tokenID, txHash, block)
	if err != nil {
		t.Fatalf("failed to mint token: %v", err)
	}

	// Try to burn when not allowed
	err = token.Burn(tokenID, txHash, block)
	if err != ErrBurningNotAllowed {
		t.Errorf("expected ErrBurningNotAllowed, got %v", err)
	}
}

// TestERC721TokenPause tests token pausing functionality
func TestERC721TokenPause(t *testing.T) {
	config := DefaultERC721TokenConfig()
	config.Pausable = true
	token := NewERC721Token("Test NFT", "TNFT", "https://api.example.com/metadata/", generateRandomAddress(), config)
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

// TestERC721TokenBlacklist tests token blacklisting functionality
func TestERC721TokenBlacklist(t *testing.T) {
	config := DefaultERC721TokenConfig()
	config.Blacklistable = true
	token := NewERC721Token("Test NFT", "TNFT", "https://api.example.com/metadata/", generateRandomAddress(), config)
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

// TestERC721TokenClone tests token cloning functionality
func TestERC721TokenClone(t *testing.T) {
	config := DefaultERC721TokenConfig()
	config.Mintable = true
	config.Burnable = true
	config.Pausable = true
	config.Blacklistable = true
	token := NewERC721Token("Test NFT", "TNFT", "https://api.example.com/metadata/", generateRandomAddress(), config)

	// Add some state
	address1 := generateRandomAddress()
	tokenID := uint64(1)
	
	// Mint a token
	err := token.Mint(address1, tokenID, generateRandomHash(), 12345)
	if err != nil {
		t.Fatalf("failed to mint token: %v", err)
	}

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
	if clone.GetBaseURI() != token.GetBaseURI() {
		t.Error("expected clone base URI to match original")
	}
	if clone.GetTotalSupply().Cmp(token.GetTotalSupply()) != 0 {
		t.Error("expected clone total supply to match original")
	}

	// Check that token ownership is copied
	if clone.GetOwnerOf(tokenID) != token.GetOwnerOf(tokenID) {
		t.Error("expected clone token ownership to match original")
	}
}

// TestERC721TokenConcurrency tests concurrent access to token
func TestERC721TokenConcurrency(t *testing.T) {
	token := NewERC721Token("Test NFT", "TNFT", "https://api.example.com/metadata/", generateRandomAddress(), DefaultERC721TokenConfig())
	
	// Test concurrent reads
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = token.GetName()
				_ = token.GetSymbol()
				_ = token.GetBaseURI()
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

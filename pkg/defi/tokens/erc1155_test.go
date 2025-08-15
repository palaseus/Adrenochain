package tokens

import (
	"math/big"
	"testing"
)

// TestNewERC1155Token tests ERC-1155 token creation
func TestNewERC1155Token(t *testing.T) {
	uri := "https://api.example.com/metadata/"
	owner := generateRandomAddress()
	config := DefaultERC1155TokenConfig()

	token := NewERC1155Token(uri, owner, config)

	if token == nil {
		t.Fatal("expected non-nil token")
	}

	if token.GetURI() != uri {
		t.Errorf("expected URI %s, got %s", uri, token.GetURI())
	}

	if token.GetOwner() != owner {
		t.Errorf("expected owner %v, got %v", owner, token.GetOwner())
	}

	if token.GetTotalSupply(1) == nil {
		t.Error("expected non-nil total supply")
	}
}

// TestERC1155TokenMint tests single token minting functionality
func TestERC1155TokenMint(t *testing.T) {
	config := DefaultERC1155TokenConfig()
	config.Mintable = true
	token := NewERC1155Token("https://api.example.com/metadata/", generateRandomAddress(), config)
	to := generateRandomAddress()
	tokenID := uint64(1)
	amount := big.NewInt(100)
	txHash := generateRandomHash()
	block := uint64(12345)

	// Test minting
	err := token.Mint(to, tokenID, amount, nil, txHash, block)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check token balance
	balance := token.GetBalance(to, tokenID)
	if balance.Cmp(amount) != 0 {
		t.Errorf("expected balance %s, got %s", amount.String(), balance.String())
	}

	// Check all balances for the address
	allBalances := token.GetBalances(to)
	if len(allBalances) != 1 {
		t.Errorf("expected 1 token balance, got %d", len(allBalances))
	}
	if allBalances[tokenID].Cmp(amount) != 0 {
		t.Errorf("expected all balances to show %s, got %s", amount.String(), allBalances[tokenID].String())
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
	if event.ID != tokenID {
		t.Errorf("expected event ID %d, got %d", tokenID, event.ID)
	}
	if event.Value.Cmp(amount) != 0 {
		t.Errorf("expected event value %s, got %s", amount.String(), event.Value.String())
	}
}

// TestERC1155TokenMintBatch tests batch token minting functionality
func TestERC1155TokenMintBatch(t *testing.T) {
	config := DefaultERC1155TokenConfig()
	config.Mintable = true
	token := NewERC1155Token("https://api.example.com/metadata/", generateRandomAddress(), config)
	to := generateRandomAddress()
	tokenIDs := []uint64{1, 2, 3}
	amounts := []*big.Int{big.NewInt(100), big.NewInt(200), big.NewInt(300)}
	txHash := generateRandomHash()
	block := uint64(12345)

	// Test batch minting
	err := token.MintBatch(to, tokenIDs, amounts, nil, txHash, block)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check all token balances
	for i, tokenID := range tokenIDs {
		expectedAmount := amounts[i]
		balance := token.GetBalance(to, tokenID)
		if balance.Cmp(expectedAmount) != 0 {
			t.Errorf("expected balance for token %d to be %s, got %s", tokenID, expectedAmount.String(), balance.String())
		}
	}

	// Check all balances for the address
	allBalances := token.GetBalances(to)
	if len(allBalances) != 3 {
		t.Errorf("expected 3 token balances, got %d", len(allBalances))
	}

	// Check mint events
	events := token.GetMintEvents()
	if len(events) != 3 {
		t.Errorf("expected 3 mint events, got %d", len(events))
	}
}

// TestERC1155TokenMintNotAllowed tests minting when not allowed
func TestERC1155TokenMintNotAllowed(t *testing.T) {
	token := NewERC1155Token("https://api.example.com/metadata/", generateRandomAddress(), DefaultERC1155TokenConfig())
	to := generateRandomAddress()
	tokenID := uint64(1)
	amount := big.NewInt(100)
	txHash := generateRandomHash()
	block := uint64(12345)

	// Test minting when not allowed
	err := token.Mint(to, tokenID, amount, nil, txHash, block)
	if err != ErrMintingNotAllowed {
		t.Errorf("expected ErrMintingNotAllowed, got %v", err)
	}
}

// TestERC1155TokenSafeTransferFrom tests single token transfer functionality
func TestERC1155TokenSafeTransferFrom(t *testing.T) {
	config := DefaultERC1155TokenConfig()
	config.Mintable = true
	token := NewERC1155Token("https://api.example.com/metadata/", generateRandomAddress(), config)
	from := generateRandomAddress()
	to := generateRandomAddress()
	operator := generateRandomAddress()
	tokenID := uint64(1)
	amount := big.NewInt(50)
	txHash := generateRandomHash()
	block := uint64(12345)

	// Mint token to from address
	err := token.Mint(from, tokenID, big.NewInt(100), nil, txHash, block)
	if err != nil {
		t.Fatalf("failed to mint token: %v", err)
	}

	// Approve operator for all tokens
	err = token.SetApprovalForAll(from, operator, true, txHash, block)
	if err != nil {
		t.Fatalf("failed to approve operator: %v", err)
	}

	// Test transfer
	err = token.SafeTransferFrom(operator, from, to, tokenID, amount, nil, txHash, block)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check balances
	fromBalance := token.GetBalance(from, tokenID)
	expectedFromBalance := big.NewInt(50) // 100 - 50
	if fromBalance.Cmp(expectedFromBalance) != 0 {
		t.Errorf("expected from balance %s, got %s", expectedFromBalance.String(), fromBalance.String())
	}

	toBalance := token.GetBalance(to, tokenID)
	if toBalance.Cmp(amount) != 0 {
		t.Errorf("expected to balance %s, got %s", amount.String(), toBalance.String())
	}

	// Check transfer events
	events := token.GetTransferSingleEvents()
	if len(events) != 1 {
		t.Errorf("expected 1 transfer event, got %d", len(events))
	}

	event := events[0]
	if event.Operator != operator {
		t.Errorf("expected event operator %v, got %v", operator, event.Operator)
	}
	if event.From != from {
		t.Errorf("expected event from %v, got %v", from, event.From)
	}
	if event.To != to {
		t.Errorf("expected event to %v, got %v", to, event.To)
	}
	if event.ID != tokenID {
		t.Errorf("expected event ID %d, got %d", tokenID, event.ID)
	}
	if event.Value.Cmp(amount) != 0 {
		t.Errorf("expected event value %s, got %s", amount.String(), event.Value.String())
	}
}

// TestERC1155TokenSafeBatchTransferFrom tests batch token transfer functionality
func TestERC1155TokenSafeBatchTransferFrom(t *testing.T) {
	config := DefaultERC1155TokenConfig()
	config.Mintable = true
	token := NewERC1155Token("https://api.example.com/metadata/", generateRandomAddress(), config)
	from := generateRandomAddress()
	to := generateRandomAddress()
	operator := generateRandomAddress()
	tokenIDs := []uint64{1, 2, 3}
	amounts := []*big.Int{big.NewInt(50), big.NewInt(100), big.NewInt(150)}
	txHash := generateRandomHash()
	block := uint64(12345)

	// Mint tokens to from address
	for _, tokenID := range tokenIDs {
		err := token.Mint(from, tokenID, big.NewInt(200), nil, txHash, block)
		if err != nil {
			t.Fatalf("failed to mint token %d: %v", tokenID, err)
		}
	}

	// Approve operator for all tokens
	err := token.SetApprovalForAll(from, operator, true, txHash, block)
	if err != nil {
		t.Fatalf("failed to approve operator: %v", err)
	}

	// Test batch transfer
	err = token.SafeBatchTransferFrom(operator, from, to, tokenIDs, amounts, nil, txHash, block)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check balances
	for i, tokenID := range tokenIDs {
		expectedAmount := amounts[i]
		toBalance := token.GetBalance(to, tokenID)
		if toBalance.Cmp(expectedAmount) != 0 {
			t.Errorf("expected to balance for token %d to be %s, got %s", tokenID, expectedAmount.String(), toBalance.String())
		}
	}

	// Check batch transfer events
	events := token.GetTransferBatchEvents()
	if len(events) != 1 {
		t.Errorf("expected 1 batch transfer event, got %d", len(events))
	}

	event := events[0]
	if event.Operator != operator {
		t.Errorf("expected event operator %v, got %v", operator, event.Operator)
	}
	if event.From != from {
		t.Errorf("expected event from %v, got %v", from, event.From)
	}
	if event.To != to {
		t.Errorf("expected event to %v, got %v", to, event.To)
	}
	if len(event.IDs) != len(tokenIDs) {
		t.Errorf("expected %d token IDs in event, got %d", len(tokenIDs), len(event.IDs))
	}
	if len(event.Values) != len(amounts) {
		t.Errorf("expected %d amounts in event, got %d", len(amounts), len(event.Values))
	}
}

// TestERC1155TokenSafeTransferFromUnauthorized tests transfer without approval
func TestERC1155TokenSafeTransferFromUnauthorized(t *testing.T) {
	config := DefaultERC1155TokenConfig()
	config.Mintable = true
	token := NewERC1155Token("https://api.example.com/metadata/", generateRandomAddress(), config)
	from := generateRandomAddress()
	to := generateRandomAddress()
	unauthorized := generateRandomAddress()
	tokenID := uint64(1)
	amount := big.NewInt(50)
	txHash := generateRandomHash()
	block := uint64(12345)

	// Mint token to from address
	err := token.Mint(from, tokenID, big.NewInt(100), nil, txHash, block)
	if err != nil {
		t.Fatalf("failed to mint token: %v", err)
	}

	// Try to transfer without approval
	err = token.SafeTransferFrom(unauthorized, from, to, tokenID, amount, nil, txHash, block)
	if err != ErrUnauthorizedOperation {
		t.Errorf("expected ErrUnauthorizedOperation, got %v", err)
	}
}

// TestERC1155TokenSafeTransferFromInsufficientBalance tests transfer with insufficient balance
func TestERC1155TokenSafeTransferFromInsufficientBalance(t *testing.T) {
	config := DefaultERC1155TokenConfig()
	config.Mintable = true
	token := NewERC1155Token("https://api.example.com/metadata/", generateRandomAddress(), config)
	from := generateRandomAddress()
	to := generateRandomAddress()
	tokenID := uint64(1)
	amount := big.NewInt(150)
	txHash := generateRandomHash()
	block := uint64(12345)

	// Mint token to from address with insufficient balance
	err := token.Mint(from, tokenID, big.NewInt(100), nil, txHash, block)
	if err != nil {
		t.Fatalf("failed to mint token: %v", err)
	}

	// Try to transfer more than available
	err = token.SafeTransferFrom(from, from, to, tokenID, amount, nil, txHash, block)
	if err != ErrInsufficientBalance {
		t.Errorf("expected ErrInsufficientBalance, got %v", err)
	}
}

// TestERC1155TokenSetApprovalForAll tests operator approval functionality
func TestERC1155TokenSetApprovalForAll(t *testing.T) {
	token := NewERC1155Token("https://api.example.com/metadata/", generateRandomAddress(), DefaultERC1155TokenConfig())
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

// TestERC1155TokenBurn tests single token burning functionality
func TestERC1155TokenBurn(t *testing.T) {
	config := DefaultERC1155TokenConfig()
	config.Mintable = true
	config.Burnable = true
	token := NewERC1155Token("https://api.example.com/metadata/", generateRandomAddress(), config)
	owner := generateRandomAddress()
	tokenID := uint64(1)
	amount := big.NewInt(50)
	txHash := generateRandomHash()
	block := uint64(12345)

	// Mint token to owner
	err := token.Mint(owner, tokenID, big.NewInt(100), nil, txHash, block)
	if err != nil {
		t.Fatalf("failed to mint token: %v", err)
	}

	// Test burning
	err = token.Burn(owner, tokenID, amount, txHash, block)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check balance after burning
	balance := token.GetBalance(owner, tokenID)
	expectedBalance := big.NewInt(50) // 100 - 50
	if balance.Cmp(expectedBalance) != 0 {
		t.Errorf("expected balance after burning %s, got %s", expectedBalance.String(), balance.String())
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
	if event.ID != tokenID {
		t.Errorf("expected event ID %d, got %d", tokenID, event.ID)
	}
	if event.Value.Cmp(amount) != 0 {
		t.Errorf("expected event value %s, got %s", amount.String(), event.Value.String())
	}
}

// TestERC1155TokenBurnBatch tests batch token burning functionality
func TestERC1155TokenBurnBatch(t *testing.T) {
	config := DefaultERC1155TokenConfig()
	config.Mintable = true
	config.Burnable = true
	token := NewERC1155Token("https://api.example.com/metadata/", generateRandomAddress(), config)
	owner := generateRandomAddress()
	tokenIDs := []uint64{1, 2, 3}
	amounts := []*big.Int{big.NewInt(50), big.NewInt(100), big.NewInt(150)}
	txHash := generateRandomHash()
	block := uint64(12345)

	// Mint tokens to owner
	for _, tokenID := range tokenIDs {
		err := token.Mint(owner, tokenID, big.NewInt(200), nil, txHash, block)
		if err != nil {
			t.Fatalf("failed to mint token %d: %v", tokenID, err)
		}
	}

	// Test batch burning
	err := token.BurnBatch(owner, tokenIDs, amounts, txHash, block)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check balances after burning
	for i, tokenID := range tokenIDs {
		burnedAmount := amounts[i]
		expectedBalance := big.NewInt(200 - burnedAmount.Int64())
		balance := token.GetBalance(owner, tokenID)
		if balance.Cmp(expectedBalance) != 0 {
			t.Errorf("expected balance for token %d after burning %s, got %s", tokenID, expectedBalance.String(), balance.String())
		}
	}

	// Check burn events
	burnEvents := token.GetBurnEvents()
	if len(burnEvents) != 3 { // 3 burn events
		t.Errorf("expected 3 burn events, got %d", len(burnEvents))
	}
	
	// Check mint events
	mintEvents := token.GetMintEvents()
	if len(mintEvents) != 3 { // 3 mint events
		t.Errorf("expected 3 mint events, got %d", len(mintEvents))
	}
}

// TestERC1155TokenBurnNotAllowed tests burning when not allowed
func TestERC1155TokenBurnNotAllowed(t *testing.T) {
	config := DefaultERC1155TokenConfig()
	config.Mintable = true
	token := NewERC1155Token("https://api.example.com/metadata/", generateRandomAddress(), config)
	owner := generateRandomAddress()
	tokenID := uint64(1)
	amount := big.NewInt(50)
	txHash := generateRandomHash()
	block := uint64(12345)

	// Mint token to owner
	err := token.Mint(owner, tokenID, big.NewInt(100), nil, txHash, block)
	if err != nil {
		t.Fatalf("failed to mint token: %v", err)
	}

	// Try to burn when not allowed
	err = token.Burn(owner, tokenID, amount, txHash, block)
	if err != ErrBurningNotAllowed {
		t.Errorf("expected ErrBurningNotAllowed, got %v", err)
	}
}

// TestERC1155TokenSetTokenURI tests token URI setting functionality
func TestERC1155TokenSetTokenURI(t *testing.T) {
	token := NewERC1155Token("https://api.example.com/metadata/", generateRandomAddress(), DefaultERC1155TokenConfig())
	tokenID := uint64(1)
	customURI := "https://api.example.com/metadata/token1.json"
	txHash := generateRandomHash()
	block := uint64(12345)

	// Test setting token URI
	err := token.SetTokenURI(tokenID, customURI, txHash, block)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check token URI
	uri := token.GetTokenURI(tokenID)
	if uri != customURI {
		t.Errorf("expected token URI %s, got %s", customURI, uri)
	}

	// Check URI events
	events := token.GetURIEvents()
	if len(events) != 1 {
		t.Errorf("expected 1 URI event, got %d", len(events))
	}

	event := events[0]
	if event.URI != customURI {
		t.Errorf("expected event URI %s, got %s", customURI, event.URI)
	}
	if event.ID != tokenID {
		t.Errorf("expected event ID %d, got %d", tokenID, event.ID)
	}
}

// TestERC1155TokenPause tests token pausing functionality
func TestERC1155TokenPause(t *testing.T) {
	config := DefaultERC1155TokenConfig()
	config.Pausable = true
	token := NewERC1155Token("https://api.example.com/metadata/", generateRandomAddress(), config)
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

// TestERC1155TokenBlacklist tests token blacklisting functionality
func TestERC1155TokenBlacklist(t *testing.T) {
	config := DefaultERC1155TokenConfig()
	config.Blacklistable = true
	token := NewERC1155Token("https://api.example.com/metadata/", generateRandomAddress(), config)
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

// TestERC1155TokenClone tests token cloning functionality
func TestERC1155TokenClone(t *testing.T) {
	config := DefaultERC1155TokenConfig()
	config.Mintable = true
	config.Burnable = true
	config.Pausable = true
	config.Blacklistable = true
	token := NewERC1155Token("https://api.example.com/metadata/", generateRandomAddress(), config)

	// Add some state
	address1 := generateRandomAddress()
	tokenID := uint64(1)
	
	// Mint a token
	err := token.Mint(address1, tokenID, big.NewInt(100), nil, generateRandomHash(), 12345)
	if err != nil {
		t.Fatalf("failed to mint token: %v", err)
	}

	// Clone the token
	clone := token.Clone()

	if clone == token {
		t.Error("expected clone to be different from original")
	}

	// Check that values are copied
	if clone.GetURI() != token.GetURI() {
		t.Error("expected clone URI to match original")
	}
	if clone.GetOwner() != token.GetOwner() {
		t.Error("expected clone owner to match original")
	}

	// Check that token balance is copied
	if clone.GetBalance(address1, tokenID).Cmp(token.GetBalance(address1, tokenID)) != 0 {
		t.Error("expected clone token balance to match original")
	}
}

// TestERC1155TokenConcurrency tests concurrent access to token
func TestERC1155TokenConcurrency(t *testing.T) {
	token := NewERC1155Token("https://api.example.com/metadata/", generateRandomAddress(), DefaultERC1155TokenConfig())
	
	// Test concurrent reads
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = token.GetURI()
				_ = token.GetOwner()
				_ = token.GetTotalSupply(1)
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

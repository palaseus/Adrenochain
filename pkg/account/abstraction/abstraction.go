// Package abstraction implements ERC-4337 account abstraction building blocks:
// UserOperations, an EntryPoint-style validation interface, and a paymaster/gas
// abstraction model. It does not deploy any on-chain contracts; it provides the
// off-chain data structures and validation logic a bundler or wallet would use.
//
// Aligned with ERC-4337 (final Standards Track; EntryPoint deployed on mainnet
// March 2023, 26M+ smart accounts / 170M+ UserOperations as of 2026) and
// EIP-7702 (Pectra, May 2025) for in-place EOA delegation.
//
// See pkg/account/abstraction/README.md for design notes.

package abstraction

import (
	"crypto/sha256"
	"errors"
	"math/big"
)

// ---------------------------------------------------------------------------
// UserOperation (ERC-4337)
// ---------------------------------------------------------------------------

// UserOperation is a pseudo-transaction submitted to the alternative mempool.
// Fields match the ERC-4337 UserOperation struct.
type UserOperation struct {
	Sender               []byte   // contract account address
	Nonce                *big.Int // anti-replay
	InitCode             []byte   // factory + constructor calldata
	CallData             []byte   // execution calldata
	CallGasLimit         *big.Int
	VerificationGasLimit *big.Int
	PreVerificationGas   *big.Int
	MaxFeePerGas         *big.Int
	MaxPriorityFeePerGas *big.Int
	PaymasterAndData     []byte
	Signature            []byte
}

// Hash returns the UserOperation hash (keccak-style placeholder = sha256 here).
func (uo *UserOperation) Hash() []byte {
	if uo == nil {
		return nil
	}
	h := sha256.New()
	h.Write(uo.Sender)
	h.Write(uo.Nonce.Bytes())
	h.Write(uo.InitCode)
	h.Write(uo.CallData)
	h.Write(uo.CallGasLimit.Bytes())
	h.Write(uo.VerificationGasLimit.Bytes())
	h.Write(uo.PreVerificationGas.Bytes())
	h.Write(uo.MaxFeePerGas.Bytes())
	h.Write(uo.MaxPriorityFeePerGas.Bytes())
	h.Write(uo.PaymasterAndData)
	h.Write(uo.Signature)
	return h.Sum(nil)
}

// ValidateBasic checks structural invariants before signature verification.
func (uo *UserOperation) ValidateBasic() error {
	if uo == nil {
		return errors.New("nil user operation")
	}
	if len(uo.Sender) == 0 {
		return errors.New("empty sender")
	}
	if uo.Nonce == nil {
		return errors.New("nil nonce")
	}
	if uo.CallGasLimit == nil || uo.CallGasLimit.Sign() < 0 {
		return errors.New("invalid call gas limit")
	}
	if uo.VerificationGasLimit == nil || uo.VerificationGasLimit.Sign() < 0 {
		return errors.New("invalid verification gas limit")
	}
	if uo.PreVerificationGas == nil || uo.PreVerificationGas.Sign() < 0 {
		return errors.New("invalid pre verification gas")
	}
	if uo.MaxFeePerGas == nil || uo.MaxFeePerGas.Sign() < 0 {
		return errors.New("invalid max fee per gas")
	}
	if uo.MaxPriorityFeePerGas == nil || uo.MaxPriorityFeePerGas.Sign() < 0 {
		return errors.New("invalid max priority fee per gas")
	}
	if uo.MaxPriorityFeePerGas.Cmp(uo.MaxFeePerGas) > 0 {
		return errors.New("priority fee exceeds max fee")
	}
	return nil
}

// ---------------------------------------------------------------------------
// EntryPoint-style validation
// ---------------------------------------------------------------------------

// ValidationResult is the outcome of validating a UserOperation.
type ValidationResult struct {
	Valid          bool
	Reason         string
	GasUsed        *big.Int
	PaymasterStake bool
}

// Account is a smart contract account that can validate a UserOperation.
type Account interface {
	ValidateUserOp(uo *UserOperation, caller []byte) (*ValidationResult, error)
}

// EntryPoint is the on-chain hub that validates and executes UserOperations.
// Off-chain, it exposes the same validation interface a bundler uses.
type EntryPoint struct {
	accounts map[string]Account
}

// NewEntryPoint creates an EntryPoint.
func NewEntryPoint() *EntryPoint {
	return &EntryPoint{accounts: make(map[string]Account)}
}

// RegisterAccount maps a sender address to its account implementation.
func (ep *EntryPoint) RegisterAccount(sender []byte, acct Account) {
	ep.accounts[string(sender)] = acct
}

// ValidateUserOp validates a UserOperation against its sender account.
func (ep *EntryPoint) ValidateUserOp(uo *UserOperation) (*ValidationResult, error) {
	if uo == nil {
		return nil, errors.New("nil user operation")
	}
	if err := uo.ValidateBasic(); err != nil {
		return &ValidationResult{Valid: false, Reason: err.Error()}, nil
	}
	acct, ok := ep.accounts[string(uo.Sender)]
	if !ok {
		return &ValidationResult{Valid: false, Reason: "sender account not registered"}, nil
	}
	if acct == nil {
		return &ValidationResult{Valid: false, Reason: "sender account is nil"}, nil
	}
	return acct.ValidateUserOp(uo, uo.Sender)
}

// ---------------------------------------------------------------------------
// Paymaster (gas abstraction)
// ---------------------------------------------------------------------------

// Paymaster sponsors gas for a UserOperation (ERC-4337 paymaster).
type Paymaster interface {
	// ValidatePaymasterUserOp checks whether the paymaster will sponsor this op.
	ValidatePaymasterUserOp(uo *UserOperation) (*ValidationResult, error)
}

// SponsoredPaymaster is a paymaster that sponsors ops for a whitelist of senders.
type SponsoredPaymaster struct {
	Senders map[string]bool
}

// ValidatePaymasterUserOp returns valid if sender is whitelisted.
func (pm *SponsoredPaymaster) ValidatePaymasterUserOp(uo *UserOperation) (*ValidationResult, error) {
	if pm == nil || pm.Senders == nil {
		return &ValidationResult{Valid: false, Reason: "paymaster not configured"}, nil
	}
	if uo == nil {
		return &ValidationResult{Valid: false, Reason: "nil user operation"}, nil
	}
	if !pm.Senders[string(uo.Sender)] {
		return &ValidationResult{Valid: false, Reason: "sender not sponsored"}, nil
	}
	return &ValidationResult{Valid: true}, nil
}

// ---------------------------------------------------------------------------
// EIP-7702 (Pectra) delegation
// ---------------------------------------------------------------------------

// Authorization represents an EIP-7702 authorization for an EOA to delegate
// to a smart contract code (contract address, chain id, nonce).
type Authorization struct {
	ContractAddress []byte
	ChainID         *big.Int
	Nonce           *big.Int
	Signature       []byte
}

// SignHash returns the digest an EOA signs for the authorization.
func (auth *Authorization) SignHash() []byte {
	if auth == nil {
		return nil
	}
	if auth.ContractAddress == nil {
		return nil
	}
	if auth.ChainID == nil || auth.Nonce == nil {
		return nil
	}
	h := sha256.New()
	h.Write(auth.ContractAddress)
	h.Write(auth.ChainID.Bytes())
	h.Write(auth.Nonce.Bytes())
	return h.Sum(nil)
}

// ---------------------------------------------------------------------------
// Fee estimation
// ---------------------------------------------------------------------------

// GasEstimate estimates the total gas cost for a UserOperation.
type GasEstimate struct {
	PreVerificationGas *big.Int
	VerificationGas    *big.Int
	CallGas            *big.Int
	Total              *big.Int
}

// EstimateGas sums the gas components.
func EstimateGas(pre, verification, call *big.Int) *GasEstimate {
	if pre == nil {
		pre = big.NewInt(0)
	}
	if verification == nil {
		verification = big.NewInt(0)
	}
	if call == nil {
		call = big.NewInt(0)
	}
	total := new(big.Int).Add(pre, verification)
	total.Add(total, call)
	return &GasEstimate{
		PreVerificationGas: pre,
		VerificationGas:    verification,
		CallGas:            call,
		Total:              total,
	}
}

// MaxFeeForGas returns the max total fee (gas * maxFeePerGas) with cap.
func MaxFeeForGas(gas, maxFeePerGas *big.Int) *big.Int {
	if gas == nil || maxFeePerGas == nil {
		return big.NewInt(0)
	}
	return new(big.Int).Mul(gas, maxFeePerGas)
}

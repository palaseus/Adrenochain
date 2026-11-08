// Package xintent implements a Cross-Chain Intent framework inspired by
// ERC-7683 (Cross Chain Intents, Uniswap/Across Labs, Draft) and the broader
// intent-based architecture that dominated cross-chain design in 2026.
//
// Instead of users signing transactions that dictate every execution step,
// users sign intents — declarative, signed messages expressing an outcome
// (e.g. "deliver asset X on destination chain"). Competitive off-chain actors
// called solvers race to fulfill the intent, earning a spread or fee. A
// resolver normalizes protocol-specific orders into a common solver-facing
// representation.
//
// Architecture:
//   Intent         — user-signed declaration of a desired cross-chain outcome
//   Resolver       — normalizes a protocol order into solver-facing form
//   Solver         — watches intents, simulates fills, submits settlements
//   Settlement     — on-chain verification that an intent was fulfilled
//
// See pkg/xintent/README.md for the full design and ERC-7683 alignment.

package xintent

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"math/big"
	"sync"
	"time"
)

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const (
	// IntentVersion is the current intent format version.
	IntentVersion = 1

	// MaxDeadlineSeconds is the maximum allowed intent deadline (30 days).
	MaxDeadlineSeconds = 30 * 24 * 60 * 60

	// MinFillRewardBasisPoints is the minimum solver reward (1 bps = 0.01%).
	MinFillRewardBasisPoints = 1
)

// IntentType categorizes the desired cross-chain outcome.
type IntentType int

const (
	// IntentTypeTransfer is a cross-chain asset transfer.
	IntentTypeTransfer IntentType = iota
	// IntentTypeSwap is a cross-chain asset swap (different assets on each side).
	IntentTypeSwap
	// IntentTypeCall is a cross-chain contract call / message execution.
	IntentTypeCall
)

// String returns the intent type name.
func (t IntentType) String() string {
	names := map[IntentType]string{
		IntentTypeTransfer: "transfer",
		IntentTypeSwap:     "swap",
		IntentTypeCall:     "call",
	}
	if n, ok := names[t]; ok {
		return n
	}
	return "unknown"
}

// ---------------------------------------------------------------------------
// CrossChainOrder (ERC-7683 aligned)
// ---------------------------------------------------------------------------

// CrossChainOrder is the solver-facing order struct (ERC-7683 aligned).
// It is intentionally minimal: protocols wrap their execution logic in an
// opaque OrderData blob, and this struct provides the common fields solvers
// need to evaluate and fill the order.
type CrossChainOrder struct {
	// SettlementContract is the address of the on-chain settlement contract
	// that verifies and executes the fill.
	SettlementContract []byte // 20 bytes

	// Swapper is the user who originated the intent.
	Swapper []byte // 20 bytes

	// Nonce is a unique order nonce to prevent replay.
	Nonce *big.Int

	// OriginChain is the chain where the intent was initiated.
	OriginChain uint64

	// InitiationDeadline is the latest time the origin-chain action may occur.
	InitiationDeadline uint64

	// FillDeadline is the latest time a solver may fill the order.
	FillDeadline uint64

	// OrderData is protocol-specific execution payload (opaque to solvers).
	OrderData []byte

	// Signature is the swapper's signature over the order hash.
	Signature []byte
}

// OrderDataType identifies the interpretation of OrderData.
type OrderDataType int

const (
	// OrderDataTypeTransfer is a native asset transfer payload.
	OrderDataTypeTransfer OrderDataType = iota
	// OrderDataTypeSwap is a token swap payload (token in, token out).
	OrderDataTypeSwap
	// OrderDataTypeCall is an arbitrary contract call payload.
	OrderDataTypeCall
)

// Intent is a user-signed cross-chain intent declaration.
type Intent struct {
	// ID is the globally unique intent identifier (SHA-256 of the order).
	ID [32]byte

	// Order is the underlying cross-chain order.
	Order *CrossChainOrder

	// Type is the high-level intent category.
	Type IntentType

	// FillRewardBasisPoints is the solver reward in basis points.
	FillRewardBasisPoints uint16

	// Expiry is the absolute deadline after which the intent is invalid.
	Expiry time.Time

	// Status is the current lifecycle status.
	Status IntentStatus

	// CreatedAt records when the intent was created.
	CreatedAt time.Time

	// FilledBy is the solver address that filled the intent (if any).
	FilledBy []byte
}

// IntentStatus tracks the lifecycle of an intent.
type IntentStatus int

const (
	IntentStatusOpen IntentStatus = iota
	IntentStatusFilled
	IntentStatusExpired
	IntentStatusCancelled
)

// NewIntent creates a new cross-chain intent from an order.
func NewIntent(order *CrossChainOrder, itype IntentType, fillRewardBps uint16, expiry time.Time) (*Intent, error) {
	if order == nil {
		return nil, errors.New("nil order")
	}
	if order.Nonce == nil || order.Nonce.Sign() <= 0 {
		return nil, errors.New("invalid nonce")
	}
	if fillRewardBps < MinFillRewardBasisPoints {
		return nil, errors.New("fill reward below minimum")
	}
	if expiry.Before(time.Now().UTC()) {
		return nil, errors.New("intent already expired")
	}

	var id [32]byte
	h := sha256.New()
	h.Write(order.OrderData)
	h.Write(order.SettlementContract)
	h.Write(order.Swapper)
	h.Write(order.Nonce.Bytes())
	originChain := make([]byte, 8)
	binary.BigEndian.PutUint64(originChain, order.OriginChain)
	h.Write(originChain)
	initDeadline := make([]byte, 8)
	binary.BigEndian.PutUint64(initDeadline, order.InitiationDeadline)
	h.Write(initDeadline)
	fillDeadline := make([]byte, 8)
	binary.BigEndian.PutUint64(fillDeadline, order.FillDeadline)
	h.Write(fillDeadline)
	h.Write(order.Signature)
	copy(id[:], h.Sum(nil))

	return &Intent{
		ID:                    id,
		Order:                 order,
		Type:                  itype,
		FillRewardBasisPoints: fillRewardBps,
		Expiry:                expiry,
		Status:                IntentStatusOpen,
		CreatedAt:             time.Now().UTC(),
	}, nil
}

// Hash returns the intent identifier (SHA-256 of the order).
func (i *Intent) Hash() [32]byte {
	return i.ID
}

// IsExpired reports whether the intent is past its expiry.
func (i *Intent) IsExpired() bool {
	return time.Now().UTC().After(i.Expiry)
}

// FillReward returns the solver reward amount (placeholder: pro-rata of order
// value; real implementation computes from OrderData).
func (i *Intent) FillReward() *big.Int {
	// Placeholder: reward is fillRewardBps bps of a notional 1e18 unit.
	notional := big.NewInt(1e18)
	reward := new(big.Int).Mul(notional, big.NewInt(int64(i.FillRewardBasisPoints)))
	return new(big.Int).Div(reward, big.NewInt(10000))
}

// ValidateBasic checks structural invariants.
func (i *Intent) ValidateBasic() error {
	if i.Order == nil {
		return errors.New("nil order")
	}
	if i.IsExpired() {
		return errors.New("intent expired")
	}
	if i.Status == IntentStatusFilled && len(i.FilledBy) == 0 {
		return errors.New("filled intent missing solver")
	}
	return nil
}

// ---------------------------------------------------------------------------
// Resolver (ERC-7683 solver-facing normalization)
// ---------------------------------------------------------------------------

// Resolver normalizes a protocol-specific order into a solver-facing
// CrossChainOrder. Different intent protocols implement this interface.
type Resolver interface {
	// Resolve translates a raw protocol order into a CrossChainOrder.
	Resolve(rawOrder []byte) (*CrossChainOrder, error)
	// SupportedProtocol returns the protocol identifier this resolver handles.
	SupportedProtocol() string
}

// ---------------------------------------------------------------------------
// Solver
// ---------------------------------------------------------------------------

// SolverIntent is a solver's view of an open intent with evaluation metadata.
type SolverIntent struct {
	Intent    *Intent
	Resolver  Resolver
	Profit    *big.Int  // estimated solver profit
	RiskScore uint8     // 0-100 solver-assessed risk
	ExpiresAt time.Time // intent expiry
}

// Solver evaluates and fills intents.
type Solver interface {
	// Evaluate scores an intent for the solver (profit, risk, feasibility).
	Evaluate(intent *Intent) (*SolverIntent, error)
	// Fill attempts to settle the intent on behalf of a solver.
	Fill(intent *Intent, solverAddress []byte) (*FillResult, error)
}

// FillResult is the outcome of a solver fill attempt.
type FillResult struct {
	// IntentID is the filled intent identifier.
	IntentID [32]byte
	// Solver is the filling solver address.
	Solver []byte
	// SettlementTx is the on-chain settlement transaction hash (placeholder).
	SettlementTx []byte
	// Reward is the reward paid to the solver.
	Reward *big.Int
	// SettledAt is when the fill settled.
	SettledAt time.Time
	// Success indicates whether the fill succeeded.
	Success bool
}

// ---------------------------------------------------------------------------
// IntentRegistry
// ---------------------------------------------------------------------------

// IntentRegistry tracks open intents and their fills.
type IntentRegistry struct {
	mu      sync.RWMutex
	intents map[[32]byte]*Intent
	fills   map[[32]byte]*FillResult
}

// NewIntentRegistry creates an empty intent registry.
func NewIntentRegistry() *IntentRegistry {
	return &IntentRegistry{
		intents: make(map[[32]byte]*Intent),
		fills:   make(map[[32]byte]*FillResult),
	}
}

// Register adds an open intent to the registry.
func (r *IntentRegistry) Register(intent *Intent) error {
	if intent == nil {
		return errors.New("nil intent")
	}
	if err := intent.ValidateBasic(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.intents[intent.ID] = intent
	return nil
}

// Get returns an intent by ID, or nil if not found.
func (r *IntentRegistry) Get(id [32]byte) *Intent {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.intents[id]
}

// Fill records a fill result for an intent and marks it filled.
func (r *IntentRegistry) Fill(result *FillResult) error {
	if result == nil {
		return errors.New("nil fill result")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	intent, exists := r.intents[result.IntentID]
	if !exists {
		return errors.New("intent not found")
	}
	if intent.Status == IntentStatusFilled {
		return errors.New("intent already filled")
	}
	intent.Status = IntentStatusFilled
	intent.FilledBy = result.Solver
	r.fills[result.IntentID] = result
	return nil
}

// Cancel marks an intent as cancelled.
func (r *IntentRegistry) Cancel(id [32]byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	intent, exists := r.intents[id]
	if !exists {
		return errors.New("intent not found")
	}
	if intent.Status == IntentStatusFilled {
		return errors.New("cannot cancel filled intent")
	}
	intent.Status = IntentStatusCancelled
	return nil
}

// OpenIntents returns all currently open (not expired, not filled) intents.
func (r *IntentRegistry) OpenIntents() []*Intent {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*Intent
	for _, intent := range r.intents {
		if intent.Status == IntentStatusOpen && !intent.IsExpired() {
			out = append(out, intent)
		}
	}
	return out
}

// FillCount returns the number of filled intents.
func (r *IntentRegistry) FillCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.fills)
}

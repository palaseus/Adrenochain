// Package circuitbreaker implements the circuit breaker pattern for resilience
// against cascading failures. A circuit breaker wraps an operation and tracks
// its success/failure rate, tripping open after a configurable failure threshold
// to fail fast and allow the downstream system to recover.
//
// States:
//   - Closed: requests flow normally; failures are counted.
//   - Open: requests fail fast without calling the operation; after a cooldown
//     the circuit moves to half-open.
//   - Half-open: a limited number of trial requests are allowed; if they
//     succeed the circuit closes, otherwise it re-opens.
//
// See pkg/resilience/README.md for usage.

package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

// State is the current state of a circuit breaker.
type State int

const (
	// StateClosed means requests flow normally.
	StateClosed State = iota
	// StateOpen means requests fail fast.
	StateOpen
	// StateHalfOpen means a limited trial is in progress.
	StateHalfOpen
)

// String returns the state name.
func (s State) String() string {
	names := map[State]string{
		StateClosed:   "closed",
		StateOpen:     "open",
		StateHalfOpen: "half-open",
	}
	if n, ok := names[s]; ok {
		return n
	}
	return "unknown"
}

// ErrCircuitOpen is returned when a request is rejected because the circuit is open.
var ErrCircuitOpen = errors.New("circuit breaker is open")

// Config holds circuit breaker settings.
type Config struct {
	// FailureThreshold is the number of consecutive failures before tripping open.
	FailureThreshold int
	// SuccessThreshold is the number of successes in half-open needed to close.
	SuccessThreshold int
	// Timeout is how long to stay open before attempting half-open.
	Timeout time.Duration
	// MaxRequests is the max trial requests in half-open state.
	MaxRequests int
}

// DefaultConfig returns sensible production defaults.
func DefaultConfig() Config {
	return Config{
		FailureThreshold: 5,
		SuccessThreshold: 2,
		Timeout:          30 * time.Second,
		MaxRequests:      3,
	}
}

// Breaker is a circuit breaker.
type Breaker struct {
	mu sync.Mutex

	config Config

	state           State
	failures        int
	successes       int
	lastStateChange time.Time
}

// New creates a circuit breaker with the given config.
func New(config Config) *Breaker {
	return &Breaker{
		config:          config,
		state:           StateClosed,
		lastStateChange: time.Now(),
	}
}

// Execute runs fn if the circuit allows it, returning ErrCircuitOpen otherwise.
// On success the failure count resets; on failure it increments and may trip.
func (b *Breaker) Execute(fn func() error) error {
	b.mu.Lock()
	allowed, trial := b.allow()
	b.mu.Unlock()

	if !allowed {
		return ErrCircuitOpen
	}

	err := fn()

	b.mu.Lock()
	defer b.mu.Unlock()

	if err != nil {
		b.registerFailure()
	} else {
		b.registerSuccess(trial)
	}
	return err
}

// State returns the current breaker state.
func (b *Breaker) State() State {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state
}

// allow checks whether a request may proceed. Must be called with the lock held.
// trial reports whether this is a half-open trial request.
func (b *Breaker) allow() (allowed, trial bool) {
	now := time.Now()

	switch b.state {
	case StateClosed:
		return true, false

	case StateOpen:
		if now.Sub(b.lastStateChange) >= b.config.Timeout {
			b.state = StateHalfOpen
			b.successes = 0
			b.lastStateChange = now
			return true, true
		}
		return false, false

	case StateHalfOpen:
		if b.successes >= b.config.MaxRequests {
			return false, false
		}
		return true, true
	}
	return false, false
}

// registerFailure records a failed call. Must be called with the lock held.
func (b *Breaker) registerFailure() {
	b.failures++
	switch b.state {
	case StateClosed:
		if b.failures >= b.config.FailureThreshold {
			b.state = StateOpen
			b.lastStateChange = time.Now()
		}
	case StateHalfOpen:
		// Any failure in half-open re-opens immediately.
		b.state = StateOpen
		b.lastStateChange = time.Now()
	}
}

// registerSuccess records a successful call. Must be called with the lock held.
// trial reports whether this was a half-open trial.
func (b *Breaker) registerSuccess(trial bool) {
	if !trial {
		b.failures = 0
		return
	}
	b.successes++
	if b.state == StateHalfOpen && b.successes >= b.config.SuccessThreshold {
		b.state = StateClosed
		b.failures = 0
		b.successes = 0
		b.lastStateChange = time.Now()
	}
}

// Stats returns current breaker statistics.
type Stats struct {
	State           State
	Failures        int
	Successes       int
	LastStateChange time.Time
}

// Stats returns a snapshot of current breaker statistics.
func (b *Breaker) Stats() Stats {
	b.mu.Lock()
	defer b.mu.Unlock()
	return Stats{
		State:           b.state,
		Failures:        b.failures,
		Successes:       b.successes,
		LastStateChange: b.lastStateChange,
	}
}

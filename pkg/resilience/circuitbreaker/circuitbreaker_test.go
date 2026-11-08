package circuitbreaker

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBreakerTripsAndRecovers(t *testing.T) {
	cb := New(Config{
		FailureThreshold: 2,
		SuccessThreshold: 1,
		Timeout:          50 * time.Millisecond,
		MaxRequests:      1,
	})

	// Successes keep it closed.
	assert.NoError(t, cb.Execute(func() error { return nil }))
	assert.Equal(t, StateClosed, cb.State())

	// Two failures trip it open.
	assert.Error(t, cb.Execute(func() error { return errors.New("boom") }))
	assert.Error(t, cb.Execute(func() error { return errors.New("boom") }))
	assert.Equal(t, StateOpen, cb.State())

	// Open: fail fast.
	assert.Equal(t, ErrCircuitOpen, cb.Execute(func() error { return nil }))

	// Wait for timeout, half-open trial, success closes.
	time.Sleep(100 * time.Millisecond)
	assert.NoError(t, cb.Execute(func() error { return nil }))
	assert.Equal(t, StateClosed, cb.State())
}

func TestBreakerHalfOpenFailureReopens(t *testing.T) {
	cb := New(Config{
		FailureThreshold: 1,
		SuccessThreshold: 2,
		Timeout:          50 * time.Millisecond,
		MaxRequests:      2,
	})
	cb.Execute(func() error { return errors.New("boom") })
	assert.Equal(t, StateOpen, cb.State())
	time.Sleep(100 * time.Millisecond)
	// Half-open trial fails -> re-open.
	assert.Error(t, cb.Execute(func() error { return errors.New("boom") }))
	assert.Equal(t, StateOpen, cb.State())
}

func TestStats(t *testing.T) {
	cb := New(DefaultConfig())
	cb.Execute(func() error { return nil })
	st := cb.Stats()
	assert.Equal(t, StateClosed, st.State)
	assert.Equal(t, 0, st.Failures)
}

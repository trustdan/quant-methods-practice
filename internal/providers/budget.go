package providers

import (
	"errors"
	"fmt"
	"sync"
)

// Default budget constraints
const (
	DefaultMaxRequestsPerSession = 20
	DefaultMaxTokensPerRequest   = 1024
)

var ErrBudgetExceeded = errors.New("session AI request budget limit reached")

// BudgetTracker enforces session-level request and token boundaries for external providers.
type BudgetTracker struct {
	mu              sync.RWMutex
	maxRequests     int
	currentRequests int
	estimatedTokens int
}

// NewBudgetTracker initializes a tracker with specified or default constraints.
func NewBudgetTracker(maxRequests int) *BudgetTracker {
	if maxRequests <= 0 {
		maxRequests = DefaultMaxRequestsPerSession
	}
	return &BudgetTracker{
		maxRequests: maxRequests,
	}
}

// RecordRequest validates and tracks an outgoing external LLM call.
// Returns ErrBudgetExceeded if the session limit has been reached.
func (b *BudgetTracker) RecordRequest() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.currentRequests >= b.maxRequests {
		return fmt.Errorf("%w (%d/%d requests used). Switch to offline tutor or reset budget in Settings.",
			ErrBudgetExceeded, b.currentRequests, b.maxRequests)
	}

	b.currentRequests++
	return nil
}

// RecordTokens records token consumption for the session.
func (b *BudgetTracker) RecordTokens(tokens int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.estimatedTokens += tokens
}

// GetStatus returns the current budget usage and remaining allowances.
func (b *BudgetTracker) GetStatus() BudgetStatus {
	b.mu.RLock()
	defer b.mu.RUnlock()

	rem := b.maxRequests - b.currentRequests
	if rem < 0 {
		rem = 0
	}

	return BudgetStatus{
		MaxRequestsPerSession: b.maxRequests,
		CurrentRequests:       b.currentRequests,
		RemainingRequests:     rem,
		EstimatedTokens:       b.estimatedTokens,
		CapReached:            b.currentRequests >= b.maxRequests,
	}
}

// Reset clears the session usage counter.
func (b *BudgetTracker) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.currentRequests = 0
	b.estimatedTokens = 0
}

package tutor

import (
	"context"
	"errors"
	"sync/atomic"
	"time"
)

// FakeProvider simulates an asynchronous external provider for testing streaming,
// timeouts, cancellations, and fallback mechanisms.
type FakeProvider struct {
	ID             string
	ChunkDelay     time.Duration
	InitialDelay   time.Duration
	ShouldFail     bool
	FailMidStream  bool
	FailMessage    string
	SimulatedText  string
	CancelledCount int64
}

// NewFakeProvider creates a controllable mock provider.
func NewFakeProvider(id string) *FakeProvider {
	if id == "" {
		id = "fake_async"
	}
	return &FakeProvider{
		ID:            id,
		ChunkDelay:    10 * time.Millisecond,
		SimulatedText: "Simulated asynchronous provider response with LaTeX: $P(X=2) = 0.375$.",
	}
}

func (p *FakeProvider) ProviderID() string {
	return p.ID
}

func (p *FakeProvider) Capabilities() ProviderCapabilities {
	return ProviderCapabilities{
		SupportsStreaming:    true,
		SupportsFollowUps:    true,
		SupportsCancellation: true,
		IsOffline:            false,
	}
}

// Stream emits simulated chunks with delays, honoring context cancellations and failure flags.
func (p *FakeProvider) Stream(ctx context.Context, req TutorRequest) (<-chan TutorEvent, error) {
	if p.ShouldFail && !p.FailMidStream {
		msg := p.FailMessage
		if msg == "" {
			msg = "simulated provider connection failure"
		}
		return nil, errors.New(msg)
	}

	out := make(chan TutorEvent, 16)

	go func() {
		defer close(out)

		now := time.Now().UTC()

		if p.InitialDelay > 0 {
			select {
			case <-ctx.Done():
				atomic.AddInt64(&p.CancelledCount, 1)
				out <- TutorEvent{
					Type:      EventCancelled,
					RequestID: req.RequestID,
					Timestamp: time.Now().UTC(),
				}
				return
			case <-time.After(p.InitialDelay):
			}
		}

		// 1. Emit started event
		select {
		case <-ctx.Done():
			atomic.AddInt64(&p.CancelledCount, 1)
			out <- TutorEvent{
				Type:      EventCancelled,
				RequestID: req.RequestID,
				Timestamp: time.Now().UTC(),
			}
			return
		case out <- TutorEvent{
			Type:      EventStarted,
			RequestID: req.RequestID,
			Timestamp: now,
		}:
		}

		text := p.SimulatedText
		if text == "" {
			text = "Simulated async response."
		}

		chunkSize := 15
		runes := []rune(text)
		for i := 0; i < len(runes); i += chunkSize {
			if p.FailMidStream && i > len(runes)/2 {
				msg := p.FailMessage
				if msg == "" {
					msg = "simulated mid-stream provider error"
				}
				out <- TutorEvent{
					Type:      EventError,
					RequestID: req.RequestID,
					Error:     msg,
					Timestamp: time.Now().UTC(),
				}
				return
			}

			end := i + chunkSize
			if end > len(runes) {
				end = len(runes)
			}
			chunk := string(runes[i:end])

			select {
			case <-ctx.Done():
				atomic.AddInt64(&p.CancelledCount, 1)
				out <- TutorEvent{
					Type:      EventCancelled,
					RequestID: req.RequestID,
					Timestamp: time.Now().UTC(),
				}
				return
			case out <- TutorEvent{
				Type:      EventTextDelta,
				RequestID: req.RequestID,
				Delta:     chunk,
				Timestamp: time.Now().UTC(),
			}:
			}

			if p.ChunkDelay > 0 {
				select {
				case <-ctx.Done():
					atomic.AddInt64(&p.CancelledCount, 1)
					out <- TutorEvent{
						Type:      EventCancelled,
						RequestID: req.RequestID,
						Timestamp: time.Now().UTC(),
					}
					return
				case <-time.After(p.ChunkDelay):
				}
			}
		}

		// Complete event
		select {
		case <-ctx.Done():
			atomic.AddInt64(&p.CancelledCount, 1)
			out <- TutorEvent{
				Type:      EventCancelled,
				RequestID: req.RequestID,
				Timestamp: time.Now().UTC(),
			}
		case out <- TutorEvent{
			Type:      EventComplete,
			RequestID: req.RequestID,
			Text:      text,
			Timestamp: time.Now().UTC(),
		}:
		}
	}()

	return out, nil
}

package tutor

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"
)

// TutorManager orchestrates AI tutor requests, lifecycle management,
// streaming distribution, cancellations, and fallback to reviewed offline explanations.
type TutorManager struct {
	providers      map[string]TutorService
	offlineTutor   *OfflineTutor
	activeCancels  map[string]context.CancelFunc
	activeChannels map[string]chan TutorEvent
	mu             sync.RWMutex
}

// NewTutorManager initializes a manager with the canonical OfflineTutor.
func NewTutorManager(offlineTutor *OfflineTutor) *TutorManager {
	if offlineTutor == nil {
		offlineTutor = NewOfflineTutor(10*time.Millisecond, time.Now)
	}
	m := &TutorManager{
		providers:      make(map[string]TutorService),
		offlineTutor:   offlineTutor,
		activeCancels:  make(map[string]context.CancelFunc),
		activeChannels: make(map[string]chan TutorEvent),
	}
	m.providers["offline"] = offlineTutor
	return m
}

// RegisterProvider registers an external or mock tutor provider.
func (m *TutorManager) RegisterProvider(svc TutorService) {
	if svc == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.providers[svc.ProviderID()] = svc
}

// GenerateRequestID generates a unique, unguessable tutor request ID.
func GenerateRequestID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("req_%d_%s", time.Now().UnixNano(), hex.EncodeToString(b))
}

// StartRequest initiates a streaming tutor interaction, managing cancellation and fallback.
func (m *TutorManager) StartRequest(parentCtx context.Context, req TutorRequest) (string, <-chan TutorEvent, error) {
	if req.RequestID == "" {
		req.RequestID = GenerateRequestID()
	}

	reqCtx, cancel := context.WithCancel(parentCtx)

	out := make(chan TutorEvent, 32)

	m.mu.Lock()
	m.activeCancels[req.RequestID] = cancel
	m.activeChannels[req.RequestID] = out
	m.mu.Unlock()

	go func() {
		defer func() {
			m.mu.Lock()
			delete(m.activeCancels, req.RequestID)
			delete(m.activeChannels, req.RequestID)
			m.mu.Unlock()
			close(out)
		}()

		// Determine target provider
		providerID := req.Provider
		if providerID == "" {
			providerID = "offline"
		}

		m.mu.RLock()
		prov, ok := m.providers[providerID]
		m.mu.RUnlock()

		var streamChan <-chan TutorEvent
		var err error

		if ok && prov != nil {
			streamChan, err = prov.Stream(reqCtx, req)
		} else {
			err = fmt.Errorf("provider %q unavailable", providerID)
		}

		// If provider call failed to start, trigger offline fallback immediately
		if err != nil {
			select {
			case <-reqCtx.Done():
				out <- TutorEvent{
					Type:       EventCancelled,
					RequestID:  req.RequestID,
					SessionID:  req.SessionID,
					InstanceID: req.InstanceID,
					StageID:    req.StageID,
					Timestamp:  time.Now().UTC(),
				}
				return
			case out <- TutorEvent{
				Type:          EventFallback,
				RequestID:     req.RequestID,
				SessionID:     req.SessionID,
				InstanceID:    req.InstanceID,
				StageID:       req.StageID,
				FallbackLabel: fmt.Sprintf("Provider %s failed (%v); falling back to offline reviewed tutor.", providerID, err),
				Timestamp:     time.Now().UTC(),
			}:
			}

			// Stream from canonical offline tutor
			streamChan, _ = m.offlineTutor.Stream(reqCtx, req)
		}

		// Relay events from the provider stream
		sawTerminal := false
		for ev := range streamChan {
			if ev.Type == EventCancelled || ev.Type == EventComplete || ev.Type == EventError {
				sawTerminal = true
			}

			// If mid-stream error from non-offline provider, fallback to offline tutor
			if ev.Type == EventError && providerID != "offline" {
				select {
				case <-reqCtx.Done():
					out <- TutorEvent{
						Type:       EventCancelled,
						RequestID:  req.RequestID,
						SessionID:  req.SessionID,
						InstanceID: req.InstanceID,
						StageID:    req.StageID,
						Timestamp:  time.Now().UTC(),
					}
					return
				case out <- TutorEvent{
					Type:          EventFallback,
					RequestID:     req.RequestID,
					SessionID:     req.SessionID,
					InstanceID:    req.InstanceID,
					StageID:       req.StageID,
					FallbackLabel: fmt.Sprintf("Provider encountered an error (%s); falling back to offline reviewed explanation.", ev.Error),
					Timestamp:     time.Now().UTC(),
				}:
				}

				offlineChan, _ := m.offlineTutor.Stream(reqCtx, req)
				for offEv := range offlineChan {
					if offEv.Type == EventCancelled {
						out <- offEv
						return
					}
					select {
					case <-reqCtx.Done():
						out <- TutorEvent{
							Type:       EventCancelled,
							RequestID:  req.RequestID,
							SessionID:  req.SessionID,
							InstanceID: req.InstanceID,
							StageID:    req.StageID,
							Timestamp:  time.Now().UTC(),
						}
						return
					case out <- offEv:
					}
				}
				return
			}

			if ev.Type == EventCancelled {
				out <- ev
				return
			}

			select {
			case <-reqCtx.Done():
				out <- TutorEvent{
					Type:       EventCancelled,
					RequestID:  req.RequestID,
					SessionID:  req.SessionID,
					InstanceID: req.InstanceID,
					StageID:    req.StageID,
					Timestamp:  time.Now().UTC(),
				}
				return
			case out <- ev:
			}
		}

		if reqCtx.Err() != nil && !sawTerminal {
			out <- TutorEvent{
				Type:       EventCancelled,
				RequestID:  req.RequestID,
				SessionID:  req.SessionID,
				InstanceID: req.InstanceID,
				StageID:    req.StageID,
				Timestamp:  time.Now().UTC(),
			}
		}
	}()

	return req.RequestID, out, nil
}

// CancelRequest aborts an active in-flight tutor request.
func (m *TutorManager) CancelRequest(requestID string) bool {
	m.mu.Lock()
	cancel, ok := m.activeCancels[requestID]
	m.mu.Unlock()

	if ok && cancel != nil {
		cancel()
		return true
	}
	return false
}

// IsActive checks if a request is currently streaming.
func (m *TutorManager) IsActive(requestID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.activeCancels[requestID]
	return ok
}

// GetEventsChannel returns the broadcast channel for an active request if present.
func (m *TutorManager) GetEventsChannel(requestID string) (<-chan TutorEvent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	ch, ok := m.activeChannels[requestID]
	if !ok {
		return nil, errors.New("request not found or already completed")
	}
	return ch, nil
}

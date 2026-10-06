package providers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/trustdan/quant-methods-practice/internal/auth"
	"github.com/trustdan/quant-methods-practice/internal/tutor"
)

// AnthropicAdapter integrates with Anthropic's Messages API.
type AnthropicAdapter struct {
	Vault        auth.Vault
	Budget       *BudgetTracker
	BaseURL      string
	HTTPClient   *http.Client
	DefaultModel string
}

// NewAnthropicAdapter initializes an AnthropicAdapter.
func NewAnthropicAdapter(vault auth.Vault, budget *BudgetTracker) *AnthropicAdapter {
	if budget == nil {
		budget = NewBudgetTracker(DefaultMaxRequestsPerSession)
	}
	return &AnthropicAdapter{
		Vault:        vault,
		Budget:       budget,
		BaseURL:      "https://api.anthropic.com",
		HTTPClient:   &http.Client{Timeout: 65 * time.Second},
		DefaultModel: "claude-3-5-sonnet-20241022",
	}
}

func (a *AnthropicAdapter) ProviderID() string {
	return auth.RouteAnthropic
}

func (a *AnthropicAdapter) Capabilities() tutor.ProviderCapabilities {
	return tutor.ProviderCapabilities{
		SupportsStreaming:    true,
		SupportsFollowUps:    true,
		SupportsCancellation: true,
		IsOffline:            false,
	}
}

type anthropicMessageRequest struct {
	Model     string                    `json:"model"`
	MaxTokens int                       `json:"max_tokens"`
	Stream    bool                      `json:"stream"`
	System    string                    `json:"system,omitempty"`
	Messages  []anthropicMessageContent `json:"messages"`
}

type anthropicMessageContent struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicStreamEvent struct {
	Type  string `json:"type"`
	Delta *struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"delta,omitempty"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (a *AnthropicAdapter) Stream(ctx context.Context, req tutor.TutorRequest) (<-chan tutor.TutorEvent, error) {
	apiKey, err := a.Vault.Get(auth.RouteAnthropic)
	if err != nil || apiKey == "" {
		return nil, errors.New("Anthropic API key not configured")
	}

	if err := a.Budget.RecordRequest(); err != nil {
		return nil, err
	}

	model := a.DefaultModel
	out := make(chan tutor.TutorEvent, 32)

	go func() {
		defer close(out)

		now := time.Now().UTC()
		// 1. Emit started event
		select {
		case <-ctx.Done():
			out <- tutor.TutorEvent{
				Type:      tutor.EventCancelled,
				RequestID: req.RequestID,
				Timestamp: time.Now().UTC(),
			}
			return
		case out <- tutor.TutorEvent{
			Type:       tutor.EventStarted,
			RequestID:  req.RequestID,
			SessionID:  req.SessionID,
			InstanceID: req.InstanceID,
			StageID:    req.StageID,
			Timestamp:  now,
		}:
		}

		payload := anthropicMessageRequest{
			Model:     model,
			MaxTokens: DefaultMaxTokensPerRequest,
			Stream:    true,
			System:    BuildSystemPrompt(),
			Messages: []anthropicMessageContent{
				{Role: "user", Content: BuildUserPrompt(req)},
			},
		}

		bodyBytes, err := json.Marshal(payload)
		if err != nil {
			out <- tutor.TutorEvent{
				Type:      tutor.EventError,
				RequestID: req.RequestID,
				Error:     "failed to serialize Anthropic request",
				Timestamp: time.Now().UTC(),
			}
			return
		}

		url := fmt.Sprintf("%s/v1/messages", strings.TrimRight(a.BaseURL, "/"))
		var resp *http.Response
		var httpErr error

		// Retry with backoff for rate limiting (429 / 503)
		for attempt := 0; attempt < 3; attempt++ {
			httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
			if err != nil {
				httpErr = err
				break
			}

			httpReq.Header.Set("x-api-key", apiKey)
			httpReq.Header.Set("anthropic-version", "2023-06-01")
			httpReq.Header.Set("content-type", "application/json")

			resp, httpErr = a.HTTPClient.Do(httpReq)
			if httpErr != nil {
				if ctx.Err() != nil {
					out <- tutor.TutorEvent{
						Type:      tutor.EventCancelled,
						RequestID: req.RequestID,
						Timestamp: time.Now().UTC(),
					}
					return
				}
				break
			}

			if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusServiceUnavailable {
				resp.Body.Close()
				backoff := time.Duration(100*(1<<attempt)) * time.Millisecond
				select {
				case <-ctx.Done():
					out <- tutor.TutorEvent{
						Type:      tutor.EventCancelled,
						RequestID: req.RequestID,
						Timestamp: time.Now().UTC(),
					}
					return
				case <-time.After(backoff):
					continue
				}
			}
			break
		}

		if httpErr != nil {
			out <- tutor.TutorEvent{
				Type:      tutor.EventError,
				RequestID: req.RequestID,
				Error:     fmt.Sprintf("Anthropic connection failed: %v", httpErr),
				Timestamp: time.Now().UTC(),
			}
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			out <- tutor.TutorEvent{
				Type:      tutor.EventError,
				RequestID: req.RequestID,
				Error:     "Invalid or unauthorized Anthropic API key. Check settings.",
				Timestamp: time.Now().UTC(),
			}
			return
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			out <- tutor.TutorEvent{
				Type:      tutor.EventError,
				RequestID: req.RequestID,
				Error:     "Anthropic rate limit exceeded. Please wait or use offline tutor.",
				Timestamp: time.Now().UTC(),
			}
			return
		}

		if resp.StatusCode != http.StatusOK {
			errBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
			out <- tutor.TutorEvent{
				Type:      tutor.EventError,
				RequestID: req.RequestID,
				Error:     fmt.Sprintf("Anthropic API error (%d): %s", resp.StatusCode, string(errBytes)),
				Timestamp: time.Now().UTC(),
			}
			return
		}

		var fullText strings.Builder
		scanner := bufio.NewScanner(resp.Body)

		for scanner.Scan() {
			if ctx.Err() != nil {
				out <- tutor.TutorEvent{
					Type:      tutor.EventCancelled,
					RequestID: req.RequestID,
					Timestamp: time.Now().UTC(),
				}
				return
			}

			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			dataStr := strings.TrimPrefix(line, "data: ")
			if strings.TrimSpace(dataStr) == "[DONE]" {
				break
			}

			var ev anthropicStreamEvent
			if err := json.Unmarshal([]byte(dataStr), &ev); err != nil {
				continue
			}

			if ev.Error != nil {
				out <- tutor.TutorEvent{
					Type:      tutor.EventError,
					RequestID: req.RequestID,
					Error:     ev.Error.Message,
					Timestamp: time.Now().UTC(),
				}
				return
			}

			if ev.Type == "content_block_delta" && ev.Delta != nil && ev.Delta.Text != "" {
				chunk := ev.Delta.Text
				fullText.WriteString(chunk)

				select {
				case <-ctx.Done():
					out <- tutor.TutorEvent{
						Type:      tutor.EventCancelled,
						RequestID: req.RequestID,
						Timestamp: time.Now().UTC(),
					}
					return
				case out <- tutor.TutorEvent{
					Type:       tutor.EventTextDelta,
					RequestID:  req.RequestID,
					SessionID:  req.SessionID,
					InstanceID: req.InstanceID,
					StageID:    req.StageID,
					Delta:      chunk,
					Timestamp:  time.Now().UTC(),
				}:
				}
			}
		}

		finalStr := fullText.String()
		a.Budget.RecordTokens(len(finalStr) / 4) // rough token estimate

		select {
		case <-ctx.Done():
			out <- tutor.TutorEvent{
				Type:      tutor.EventCancelled,
				RequestID: req.RequestID,
				Timestamp: time.Now().UTC(),
			}
		case out <- tutor.TutorEvent{
			Type:       tutor.EventComplete,
			RequestID:  req.RequestID,
			SessionID:  req.SessionID,
			InstanceID: req.InstanceID,
			StageID:    req.StageID,
			Text:       finalStr,
			Timestamp:  time.Now().UTC(),
		}:
		}
	}()

	return out, nil
}

type anthropicModelCatalogResponse struct {
	Data []struct {
		ID          string `json:"id"`
		DisplayName string `json:"display_name"`
	} `json:"data"`
}

func (a *AnthropicAdapter) DiscoverModels(ctx context.Context, apiKey string) ([]ModelInfo, error) {
	if apiKey == "" {
		var err error
		apiKey, err = a.Vault.Get(auth.RouteAnthropic)
		if err != nil || apiKey == "" {
			return nil, errors.New("Anthropic API key required for model discovery")
		}
	}

	url := fmt.Sprintf("%s/v1/models", strings.TrimRight(a.BaseURL, "/"))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := a.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Anthropic models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Anthropic model discovery failed with status %d", resp.StatusCode)
	}

	var cat anthropicModelCatalogResponse
	if err := json.NewDecoder(resp.Body).Decode(&cat); err != nil {
		return nil, fmt.Errorf("failed to decode Anthropic models response: %w", err)
	}

	var models []ModelInfo
	for _, m := range cat.Data {
		name := m.DisplayName
		if name == "" {
			name = m.ID
		}
		models = append(models, ModelInfo{
			ID:                m.ID,
			Name:              name,
			Provider:          auth.RouteAnthropic,
			SupportsStreaming: true,
			ContextWindow:     200000,
			Description:       "Discovered Anthropic Claude model.",
			IsDefault:         m.ID == a.DefaultModel,
		})
	}

	return models, nil
}

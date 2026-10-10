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

// OpenAIAdapter integrates with OpenAI's Chat Completions API.
type OpenAIAdapter struct {
	Vault        auth.Vault
	Budget       *BudgetTracker
	BaseURL      string
	HTTPClient   *http.Client
	DefaultModel string
}

// NewOpenAIAdapter initializes an OpenAIAdapter.
func NewOpenAIAdapter(vault auth.Vault, budget *BudgetTracker) *OpenAIAdapter {
	if budget == nil {
		budget = NewBudgetTracker(DefaultMaxRequestsPerSession)
	}
	return &OpenAIAdapter{
		Vault:        vault,
		Budget:       budget,
		BaseURL:      "https://api.openai.com",
		HTTPClient:   &http.Client{Timeout: 65 * time.Second},
		DefaultModel: "gpt-4o",
	}
}

func (o *OpenAIAdapter) ProviderID() string {
	return auth.RouteOpenAI
}

func (o *OpenAIAdapter) Capabilities() tutor.ProviderCapabilities {
	return tutor.ProviderCapabilities{
		SupportsStreaming:    true,
		SupportsFollowUps:    true,
		SupportsCancellation: true,
		IsOffline:            false,
	}
}

type openAIMessageContent struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIChatRequest struct {
	Model               string                 `json:"model"`
	Stream              bool                   `json:"stream"`
	Messages            []openAIMessageContent `json:"messages"`
	MaxTokens           *int                   `json:"max_tokens,omitempty"`
	MaxCompletionTokens *int                   `json:"max_completion_tokens,omitempty"`
	Temperature         *float64               `json:"temperature,omitempty"`
}

type openAIStreamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

func (o *OpenAIAdapter) Stream(ctx context.Context, req tutor.TutorRequest) (<-chan tutor.TutorEvent, error) {
	apiKey, err := o.Vault.Get(auth.RouteOpenAI)
	if err != nil || apiKey == "" {
		return nil, errors.New("OpenAI API key not configured")
	}

	if err := o.Budget.RecordRequest(); err != nil {
		return nil, err
	}

	model := o.DefaultModel
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

		maxTok := DefaultMaxTokensPerRequest
		temp := 0.2

		payload := openAIChatRequest{
			Model:  model,
			Stream: true,
			Messages: []openAIMessageContent{
				{Role: "system", Content: RequestSystemPrompt(req)},
				{Role: "user", Content: BuildUserPrompt(req)},
			},
		}

		// Reasoning models (e.g. o1, o3) use max_completion_tokens and do not accept custom temperature
		if strings.HasPrefix(model, "o1") || strings.HasPrefix(model, "o3") {
			payload.MaxCompletionTokens = &maxTok
		} else {
			payload.MaxTokens = &maxTok
			payload.Temperature = &temp
		}

		bodyBytes, err := json.Marshal(payload)
		if err != nil {
			out <- tutor.TutorEvent{
				Type:      tutor.EventError,
				RequestID: req.RequestID,
				Error:     "failed to serialize OpenAI request",
				Timestamp: time.Now().UTC(),
			}
			return
		}

		url := fmt.Sprintf("%s/v1/chat/completions", strings.TrimRight(o.BaseURL, "/"))
		var resp *http.Response
		var httpErr error

		// Retry with backoff for rate limiting (429 / 503)
		for attempt := 0; attempt < 3; attempt++ {
			httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
			if err != nil {
				httpErr = err
				break
			}

			httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
			httpReq.Header.Set("content-type", "application/json")

			resp, httpErr = o.HTTPClient.Do(httpReq)
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
				Error:     fmt.Sprintf("OpenAI connection failed: %v", httpErr),
				Timestamp: time.Now().UTC(),
			}
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			out <- tutor.TutorEvent{
				Type:      tutor.EventError,
				RequestID: req.RequestID,
				Error:     "Invalid or unauthorized OpenAI API key. Check settings.",
				Timestamp: time.Now().UTC(),
			}
			return
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			out <- tutor.TutorEvent{
				Type:      tutor.EventError,
				RequestID: req.RequestID,
				Error:     "OpenAI rate limit exceeded. Please wait or use offline tutor.",
				Timestamp: time.Now().UTC(),
			}
			return
		}

		if resp.StatusCode != http.StatusOK {
			errBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
			out <- tutor.TutorEvent{
				Type:      tutor.EventError,
				RequestID: req.RequestID,
				Error:     fmt.Sprintf("OpenAI API error (%d): %s", resp.StatusCode, string(errBytes)),
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

			var chunk openAIStreamChunk
			if err := json.Unmarshal([]byte(dataStr), &chunk); err != nil {
				continue
			}

			if chunk.Error != nil {
				out <- tutor.TutorEvent{
					Type:      tutor.EventError,
					RequestID: req.RequestID,
					Error:     chunk.Error.Message,
					Timestamp: time.Now().UTC(),
				}
				return
			}

			for _, c := range chunk.Choices {
				if c.Delta.Content != "" {
					fullText.WriteString(c.Delta.Content)
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
						Delta:      c.Delta.Content,
						Timestamp:  time.Now().UTC(),
					}:
					}
				}
			}
		}

		finalStr := fullText.String()
		o.Budget.RecordTokens(len(finalStr) / 4)

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

type openAIModelCatalogResponse struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

func (o *OpenAIAdapter) DiscoverModels(ctx context.Context, apiKey string) ([]ModelInfo, error) {
	if apiKey == "" {
		var err error
		apiKey, err = o.Vault.Get(auth.RouteOpenAI)
		if err != nil || apiKey == "" {
			return nil, errors.New("OpenAI API key required for model discovery")
		}
	}

	url := fmt.Sprintf("%s/v1/models", strings.TrimRight(o.BaseURL, "/"))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	resp, err := o.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch OpenAI models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI model discovery failed with status %d", resp.StatusCode)
	}

	var cat openAIModelCatalogResponse
	if err := json.NewDecoder(resp.Body).Decode(&cat); err != nil {
		return nil, fmt.Errorf("failed to decode OpenAI models response: %w", err)
	}

	var models []ModelInfo
	for _, m := range cat.Data {
		// Filter relevant text/reasoning models
		if !strings.HasPrefix(m.ID, "gpt-") && !strings.HasPrefix(m.ID, "o1") && !strings.HasPrefix(m.ID, "o3") {
			continue
		}
		// Skip preview/deprecated snapshots with dates if base model exists
		if strings.Contains(m.ID, "instruct") || strings.Contains(m.ID, "realtime") || strings.Contains(m.ID, "audio") {
			continue
		}

		models = append(models, ModelInfo{
			ID:                m.ID,
			Name:              m.ID,
			Provider:          auth.RouteOpenAI,
			SupportsStreaming: true,
			ContextWindow:     128000,
			Description:       "Discovered OpenAI model.",
			IsDefault:         m.ID == o.DefaultModel,
		})
	}

	return models, nil
}

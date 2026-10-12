package providers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/trustdan/quant-methods-practice/internal/auth"
	"github.com/trustdan/quant-methods-practice/internal/tutor"
)

// DefaultLocalLLMBaseURL is the standard LM Studio loopback API port.
const DefaultLocalLLMBaseURL = "http://localhost:1234"

// OllamaDefaultBaseURL is the standard Ollama loopback API port.
const OllamaDefaultBaseURL = "http://localhost:11434"

// isLoopbackURL verifies that an endpoint hostname is strictly loopback (localhost, 127.0.0.1, or ::1),
// preventing any outbound network requests and preserving the offline privacy invariant.
func isLoopbackURL(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	hostname := strings.ToLower(parsed.Hostname())
	return hostname == "localhost" || hostname == "127.0.0.1" || hostname == "::1" || hostname == "[::1]"
}

// LocalLLMAdapter connects to a locally hosted OpenAI-compatible model server
// (such as LM Studio, Ollama, vLLM, or llama.cpp) over loopback HTTP.
// Guarantees zero external network access, full offline privacy, and zero API costs.
type LocalLLMAdapter struct {
	mu           sync.RWMutex
	BaseURL      string
	DefaultModel string
	HTTPClient   *http.Client
}

// NewLocalLLMAdapter creates a new local offline LLM adapter.
func NewLocalLLMAdapter(baseURL string) *LocalLLMAdapter {
	if baseURL == "" {
		baseURL = DefaultLocalLLMBaseURL
	}
	return &LocalLLMAdapter{
		BaseURL:      strings.TrimRight(baseURL, "/"),
		DefaultModel: "local-model",
		HTTPClient:   &http.Client{Timeout: 120 * time.Second},
	}
}

// ProviderID identifies this provider as the local route.
func (a *LocalLLMAdapter) ProviderID() string {
	return auth.RouteLocal
}

// Capabilities describes feature support for the local LLM.
func (a *LocalLLMAdapter) Capabilities() tutor.ProviderCapabilities {
	return tutor.ProviderCapabilities{
		SupportsStreaming:    true,
		SupportsFollowUps:    true,
		SupportsCancellation: true,
		IsOffline:            true,
	}
}

// GetBaseURL returns the configured local endpoint.
func (a *LocalLLMAdapter) GetBaseURL() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.BaseURL
}

// SetBaseURL configures the local endpoint, validating that it is on loopback.
func (a *LocalLLMAdapter) SetBaseURL(rawURL string) error {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		rawURL = DefaultLocalLLMBaseURL
	}
	if !isLoopbackURL(rawURL) {
		return fmt.Errorf("local LLM provider only permits loopback addresses (localhost, 127.0.0.1, [::1]); got %q", rawURL)
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.BaseURL = strings.TrimRight(rawURL, "/")
	return nil
}

type localModelCatalogResponse struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

// DiscoverModels queries GET /v1/models on the local loopback server.
func (a *LocalLLMAdapter) DiscoverModels(ctx context.Context, apiKey string) ([]ModelInfo, error) {
	baseURL := a.GetBaseURL()
	url := fmt.Sprintf("%s/v1/models", baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create discovery request: %w", err)
	}

	resp, err := a.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to local LLM server at %s. Please ensure your local server is running (LM Studio: click 'Start Server'; Ollama: run 'ollama serve' or 'ollama run <model>'): %w", baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("local LLM model discovery returned status %d: %s", resp.StatusCode, string(body))
	}

	var cat localModelCatalogResponse
	if err := json.NewDecoder(resp.Body).Decode(&cat); err != nil {
		return nil, fmt.Errorf("failed to decode local models response: %w", err)
	}

	var models []ModelInfo
	for _, m := range cat.Data {
		id := strings.TrimSpace(m.ID)
		if id == "" {
			continue
		}
		models = append(models, ModelInfo{
			ID:                id,
			Name:              id,
			Provider:          auth.RouteLocal,
			SupportsStreaming: true,
			ContextWindow:     8192,
			Description:       "Local offline LLM (LM Studio / Ollama)",
			IsDefault:         id == a.DefaultModel,
		})
	}

	if len(models) == 0 {
		return nil, fmt.Errorf("no loaded models found on local LLM server at %s (LM Studio: select a model in top dropdown; Ollama: pull a model with 'ollama run llama3.1')", baseURL)
	}

	a.mu.Lock()
	if a.DefaultModel == "local-model" || a.DefaultModel == "" {
		a.DefaultModel = models[0].ID
		models[0].IsDefault = true
	}
	a.mu.Unlock()

	return models, nil
}

// Stream executes a streaming chat completion against the local OpenAI-compatible endpoint.
func (a *LocalLLMAdapter) Stream(ctx context.Context, req tutor.TutorRequest) (<-chan tutor.TutorEvent, error) {
	baseURL := a.GetBaseURL()
	model := a.DefaultModel

	out := make(chan tutor.TutorEvent, 32)

	go func() {
		defer close(out)

		now := time.Now().UTC()
		// 1. Emit started event
		select {
		case <-ctx.Done():
			out <- tutor.TutorEvent{
				Type:       tutor.EventCancelled,
				RequestID:  req.RequestID,
				SessionID:  req.SessionID,
				InstanceID: req.InstanceID,
				StageID:    req.StageID,
				Timestamp:  time.Now().UTC(),
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

		// 2. Build messages payload (supporting multi-turn history)
		messages := a.buildMessages(req)

		payload := openAIChatRequest{
			Model:    model,
			Stream:   true,
			Messages: messages,
		}

		bodyBytes, err := json.Marshal(payload)
		if err != nil {
			out <- tutor.TutorEvent{
				Type:       tutor.EventError,
				RequestID:  req.RequestID,
				SessionID:  req.SessionID,
				InstanceID: req.InstanceID,
				StageID:    req.StageID,
				Error:      "failed to serialize local LLM request",
				Timestamp:  time.Now().UTC(),
			}
			return
		}

		url := fmt.Sprintf("%s/v1/chat/completions", baseURL)
		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
		if err != nil {
			out <- tutor.TutorEvent{
				Type:       tutor.EventError,
				RequestID:  req.RequestID,
				SessionID:  req.SessionID,
				InstanceID: req.InstanceID,
				StageID:    req.StageID,
				Error:      fmt.Sprintf("failed to construct request: %v", err),
				Timestamp:  time.Now().UTC(),
			}
			return
		}
		httpReq.Header.Set("content-type", "application/json")

		resp, err := a.HTTPClient.Do(httpReq)
		if err != nil {
			if ctx.Err() != nil {
				out <- tutor.TutorEvent{
					Type:       tutor.EventCancelled,
					RequestID:  req.RequestID,
					SessionID:  req.SessionID,
					InstanceID: req.InstanceID,
					StageID:    req.StageID,
					Timestamp:  time.Now().UTC(),
				}
				return
			}
			out <- tutor.TutorEvent{
				Type:       tutor.EventError,
				RequestID:  req.RequestID,
				SessionID:  req.SessionID,
				InstanceID: req.InstanceID,
				StageID:    req.StageID,
				Error:      fmt.Sprintf("Cannot connect to local LLM at %s: %v", baseURL, err),
				Timestamp:  time.Now().UTC(),
			}
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			errBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
			out <- tutor.TutorEvent{
				Type:       tutor.EventError,
				RequestID:  req.RequestID,
				SessionID:  req.SessionID,
				InstanceID: req.InstanceID,
				StageID:    req.StageID,
				Error:      fmt.Sprintf("Local LLM error (status %d): %s", resp.StatusCode, string(errBytes)),
				Timestamp:  time.Now().UTC(),
			}
			return
		}

		var fullText strings.Builder
		scanner := bufio.NewScanner(resp.Body)

		for scanner.Scan() {
			if ctx.Err() != nil {
				out <- tutor.TutorEvent{
					Type:       tutor.EventCancelled,
					RequestID:  req.RequestID,
					SessionID:  req.SessionID,
					InstanceID: req.InstanceID,
					StageID:    req.StageID,
					Timestamp:  time.Now().UTC(),
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
					Type:       tutor.EventError,
					RequestID:  req.RequestID,
					SessionID:  req.SessionID,
					InstanceID: req.InstanceID,
					StageID:    req.StageID,
					Error:      chunk.Error.Message,
					Timestamp:  time.Now().UTC(),
				}
				return
			}

			for _, c := range chunk.Choices {
				if c.Delta.Content != "" {
					fullText.WriteString(c.Delta.Content)
					select {
					case <-ctx.Done():
						out <- tutor.TutorEvent{
							Type:       tutor.EventCancelled,
							RequestID:  req.RequestID,
							SessionID:  req.SessionID,
							InstanceID: req.InstanceID,
							StageID:    req.StageID,
							Timestamp:  time.Now().UTC(),
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
		select {
		case <-ctx.Done():
			out <- tutor.TutorEvent{
				Type:       tutor.EventCancelled,
				RequestID:  req.RequestID,
				SessionID:  req.SessionID,
				InstanceID: req.InstanceID,
				StageID:    req.StageID,
				Timestamp:  time.Now().UTC(),
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

func (a *LocalLLMAdapter) buildMessages(req tutor.TutorRequest) []openAIMessageContent {
	sys := RequestSystemPrompt(req)
	var msgs []openAIMessageContent
	msgs = append(msgs, openAIMessageContent{Role: "system", Content: sys})

	// If multi-turn history exists, include up to the last 10 turns (sliding window)
	if len(req.History) > 0 {
		startIdx := 0
		if len(req.History) > 10 {
			startIdx = len(req.History) - 10
		}

		// First message in context includes problem scenario if available
		first := true
		for _, h := range req.History[startIdx:] {
			content := h.Content
			if first && h.Role == "user" && req.Instance != nil {
				content = fmt.Sprintf("Context:\n%s\n\nQuestion:\n%s", BuildUserPrompt(req), h.Content)
				first = false
			}
			msgs = append(msgs, openAIMessageContent{
				Role:    h.Role,
				Content: content,
			})
		}
		// If custom prompt is provided and not already the last history item
		if req.CustomPrompt != "" && (len(req.History) == 0 || req.History[len(req.History)-1].Content != req.CustomPrompt) {
			msgs = append(msgs, openAIMessageContent{
				Role:    "user",
				Content: req.CustomPrompt,
			})
		}
	} else {
		msgs = append(msgs, openAIMessageContent{
			Role:    "user",
			Content: BuildUserPrompt(req),
		})
	}

	return msgs
}

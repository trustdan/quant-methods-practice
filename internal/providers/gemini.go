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

// GeminiAdapter integrates with Google's Gemini API.
type GeminiAdapter struct {
	Vault        auth.Vault
	Budget       *BudgetTracker
	BaseURL      string
	HTTPClient   *http.Client
	DefaultModel string
}

// NewGeminiAdapter initializes a GeminiAdapter.
func NewGeminiAdapter(vault auth.Vault, budget *BudgetTracker) *GeminiAdapter {
	if budget == nil {
		budget = NewBudgetTracker(DefaultMaxRequestsPerSession)
	}
	return &GeminiAdapter{
		Vault:        vault,
		Budget:       budget,
		BaseURL:      "https://generativelanguage.googleapis.com",
		HTTPClient:   &http.Client{Timeout: 65 * time.Second},
		DefaultModel: "gemini-2.0-flash",
	}
}

func (g *GeminiAdapter) ProviderID() string {
	return auth.RouteGemini
}

func (g *GeminiAdapter) Capabilities() tutor.ProviderCapabilities {
	return tutor.ProviderCapabilities{
		SupportsStreaming:    true,
		SupportsFollowUps:    true,
		SupportsCancellation: true,
		IsOffline:            false,
	}
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiGenerateRequest struct {
	SystemInstruction *struct {
		Parts []geminiPart `json:"parts"`
	} `json:"system_instruction,omitempty"`
	Contents         []geminiContent `json:"contents"`
	GenerationConfig struct {
		MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
		Temperature     float64 `json:"temperature,omitempty"`
	} `json:"generationConfig"`
}

type geminiStreamChunk struct {
	Candidates []struct {
		Content struct {
			Parts []geminiPart `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error,omitempty"`
}

func (g *GeminiAdapter) Stream(ctx context.Context, req tutor.TutorRequest) (<-chan tutor.TutorEvent, error) {
	apiKey, err := g.Vault.Get(auth.RouteGemini)
	if err != nil || apiKey == "" {
		return nil, errors.New("Gemini API key not configured")
	}

	if err := g.Budget.RecordRequest(); err != nil {
		return nil, err
	}

	model := g.DefaultModel
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

		cleanModel := strings.TrimPrefix(model, "models/")
		payload := geminiGenerateRequest{
			SystemInstruction: &struct {
				Parts []geminiPart `json:"parts"`
			}{
				Parts: []geminiPart{{Text: BuildSystemPrompt()}},
			},
			Contents: []geminiContent{
				{
					Role:  "user",
					Parts: []geminiPart{{Text: BuildUserPrompt(req)}},
				},
			},
		}
		payload.GenerationConfig.MaxOutputTokens = DefaultMaxTokensPerRequest
		payload.GenerationConfig.Temperature = 0.2

		bodyBytes, err := json.Marshal(payload)
		if err != nil {
			out <- tutor.TutorEvent{
				Type:      tutor.EventError,
				RequestID: req.RequestID,
				Error:     "failed to serialize Gemini request",
				Timestamp: time.Now().UTC(),
			}
			return
		}

		url := fmt.Sprintf("%s/v1beta/models/%s:streamGenerateContent?alt=sse",
			strings.TrimRight(g.BaseURL, "/"), cleanModel)

		var resp *http.Response
		var httpErr error

		// Retry with backoff for 429
		for attempt := 0; attempt < 3; attempt++ {
			httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
			if err != nil {
				httpErr = err
				break
			}

			httpReq.Header.Set("x-goog-api-key", apiKey)
			httpReq.Header.Set("content-type", "application/json")

			resp, httpErr = g.HTTPClient.Do(httpReq)
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
				Error:     fmt.Sprintf("Gemini connection failed: %v", httpErr),
				Timestamp: time.Now().UTC(),
			}
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			out <- tutor.TutorEvent{
				Type:      tutor.EventError,
				RequestID: req.RequestID,
				Error:     "Invalid or unauthorized Gemini API key. Check settings.",
				Timestamp: time.Now().UTC(),
			}
			return
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			out <- tutor.TutorEvent{
				Type:      tutor.EventError,
				RequestID: req.RequestID,
				Error:     "Gemini rate limit exceeded. Please wait or use offline tutor.",
				Timestamp: time.Now().UTC(),
			}
			return
		}

		if resp.StatusCode != http.StatusOK {
			errBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
			out <- tutor.TutorEvent{
				Type:      tutor.EventError,
				RequestID: req.RequestID,
				Error:     fmt.Sprintf("Gemini API error (%d): %s", resp.StatusCode, string(errBytes)),
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

			var chunk geminiStreamChunk
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

			for _, c := range chunk.Candidates {
				for _, p := range c.Content.Parts {
					if p.Text != "" {
						fullText.WriteString(p.Text)
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
							Delta:      p.Text,
							Timestamp:  time.Now().UTC(),
						}:
						}
					}
				}
			}
		}

		finalStr := fullText.String()
		g.Budget.RecordTokens(len(finalStr) / 4)

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

type geminiModelCatalogResponse struct {
	Models []struct {
		Name                       string   `json:"name"`
		DisplayName                string   `json:"displayName"`
		Description                string   `json:"description"`
		InputTokenLimit            int      `json:"inputTokenLimit"`
		SupportedGenerationMethods []string `json:"supportedGenerationMethods"`
	} `json:"models"`
}

func (g *GeminiAdapter) DiscoverModels(ctx context.Context, apiKey string) ([]ModelInfo, error) {
	if apiKey == "" {
		var err error
		apiKey, err = g.Vault.Get(auth.RouteGemini)
		if err != nil || apiKey == "" {
			return nil, errors.New("Gemini API key required for model discovery")
		}
	}

	url := fmt.Sprintf("%s/v1beta/models", strings.TrimRight(g.BaseURL, "/"))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("x-goog-api-key", apiKey)

	resp, err := g.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Gemini models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Gemini model discovery failed with status %d", resp.StatusCode)
	}

	var cat geminiModelCatalogResponse
	if err := json.NewDecoder(resp.Body).Decode(&cat); err != nil {
		return nil, fmt.Errorf("failed to decode Gemini models response: %w", err)
	}

	var models []ModelInfo
	for _, m := range cat.Models {
		// Filter by generateContent capability
		canGen := false
		for _, meth := range m.SupportedGenerationMethods {
			if meth == "generateContent" {
				canGen = true
				break
			}
		}
		if !canGen {
			continue
		}

		cleanID := strings.TrimPrefix(m.Name, "models/")
		name := m.DisplayName
		if name == "" {
			name = cleanID
		}

		models = append(models, ModelInfo{
			ID:                cleanID,
			Name:              name,
			Provider:          auth.RouteGemini,
			SupportsStreaming: true,
			ContextWindow:     m.InputTokenLimit,
			Description:       m.Description,
			IsDefault:         cleanID == g.DefaultModel,
		})
	}

	return models, nil
}

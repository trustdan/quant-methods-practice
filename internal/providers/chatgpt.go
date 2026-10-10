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

// PlanTokenSource supplies ChatGPT-plan access tokens (implemented by siwc.Client).
type PlanTokenSource interface {
	AccessToken(ctx context.Context) (string, error)
	MarkSelectedReauth()
}

// ChatGPTPlanAdapter sends tutor requests through a signed-in ChatGPT plan
// using the Responses API. It is a separate billing route from the OpenAI
// API-key adapter and never reads or falls back to an API key.
type ChatGPTPlanAdapter struct {
	Tokens       PlanTokenSource
	Budget       *BudgetTracker
	BaseURL      string
	HTTPClient   *http.Client
	DefaultModel string
}

// NewChatGPTPlanAdapter initializes the plan adapter. No default model is
// assumed: plan model availability is account-specific and must be discovered.
func NewChatGPTPlanAdapter(tokens PlanTokenSource, budget *BudgetTracker) *ChatGPTPlanAdapter {
	if budget == nil {
		budget = NewBudgetTracker(DefaultMaxRequestsPerSession)
	}
	return &ChatGPTPlanAdapter{
		Tokens:     tokens,
		Budget:     budget,
		BaseURL:    "https://api.openai.com",
		HTTPClient: &http.Client{Timeout: 120 * time.Second},
	}
}

func (c *ChatGPTPlanAdapter) ProviderID() string { return auth.RouteChatGPT }

func (c *ChatGPTPlanAdapter) Capabilities() tutor.ProviderCapabilities {
	return tutor.ProviderCapabilities{
		SupportsStreaming:    true,
		SupportsFollowUps:    true,
		SupportsCancellation: true,
	}
}

// planResponsesRequest contains only fields the plan route accepts. The
// preview disallows max_output_tokens, temperature, top_p, metadata, user,
// previous_response_id and others, so they are deliberately absent.
type planResponsesRequest struct {
	Model        string             `json:"model"`
	Instructions string             `json:"instructions"`
	Input        []planInputMessage `json:"input"`
	Store        bool               `json:"store"`
	Stream       bool               `json:"stream"`
}

type planInputMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type planAPIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type planStreamEvent struct {
	Type     string `json:"type"`
	Delta    string `json:"delta"`
	Response *struct {
		Error             *planAPIError `json:"error"`
		IncompleteDetails *struct {
			Reason string `json:"reason"`
		} `json:"incomplete_details"`
	} `json:"response"`
}

// planErrorMessage maps documented error codes to route-appropriate actions.
func planErrorMessage(status int, code, msg string) string {
	switch code {
	case "subscription_sharing_usage_limit_exceeded":
		return "ChatGPT plan usage limit reached. Requests are paused; check ChatGPT Settings > Usage, or switch to the offline tutor."
	case "subscription_sharing_user_not_eligible":
		return "This ChatGPT account is not eligible for plan usage in third-party apps."
	case "subscription_sharing_invalid_user":
		return "ChatGPT sign-in is no longer valid. Sign in again in AI Providers settings."
	case "subscription_sharing_unsupported_capability":
		return "The ChatGPT plan route does not support this request or model. Choose another discovered model."
	case "subscription_sharing_usage_unavailable", "subscription_sharing_user_unavailable":
		return "ChatGPT plan usage is temporarily unavailable. Try again shortly."
	case "subscription_sharing_route_not_supported", "chatpass_v2_scope_not_authorized":
		return fmt.Sprintf("ChatGPT plan route rejected the request (%s).", code)
	}
	switch status {
	case http.StatusUnauthorized:
		return "ChatGPT sign-in was rejected. Check the selected account and sign in again."
	case http.StatusForbidden:
		return "ChatGPT plan access was refused for this account or integration."
	case http.StatusServiceUnavailable:
		return "ChatGPT plan routing is temporarily unavailable. Try again shortly."
	}
	if msg != "" {
		return fmt.Sprintf("ChatGPT plan error (%d): %s", status, msg)
	}
	return fmt.Sprintf("ChatGPT plan error (%d)", status)
}

func decodePlanError(body []byte) (code, msg string) {
	var env struct {
		Error *planAPIError `json:"error"`
	}
	if json.Unmarshal(body, &env) == nil && env.Error != nil {
		return env.Error.Code, env.Error.Message
	}
	return "", strings.TrimSpace(string(body))
}

func tokenErrorMessage(err error) string {
	switch {
	case errors.Is(err, context.Canceled):
		return "request cancelled"
	case err != nil:
		return err.Error()
	}
	return ""
}

func (c *ChatGPTPlanAdapter) Stream(ctx context.Context, req tutor.TutorRequest) (<-chan tutor.TutorEvent, error) {
	if c.DefaultModel == "" {
		return nil, errors.New("no ChatGPT plan model selected; refresh the model list in AI Providers settings")
	}
	token, err := c.Tokens.AccessToken(ctx)
	if err != nil {
		return nil, errors.New(tokenErrorMessage(err))
	}
	if err := c.Budget.RecordRequest(); err != nil {
		return nil, err
	}

	payload, err := json.Marshal(planResponsesRequest{
		Model:        c.DefaultModel,
		Instructions: RequestSystemPrompt(req),
		Input:        []planInputMessage{{Role: "user", Content: BuildUserPrompt(req)}},
		Store:        false,
		Stream:       true,
	})
	if err != nil {
		return nil, err
	}

	out := make(chan tutor.TutorEvent, 32)
	go func() {
		defer close(out)
		ev := func(e tutor.TutorEvent) bool {
			e.RequestID, e.SessionID, e.InstanceID, e.StageID = req.RequestID, req.SessionID, req.InstanceID, req.StageID
			e.Timestamp = time.Now().UTC()
			select {
			case <-ctx.Done():
				return false
			case out <- e:
				return true
			}
		}
		cancelled := func() {
			select { // never block a goroutine whose consumer has gone away
			case out <- tutor.TutorEvent{Type: tutor.EventCancelled, RequestID: req.RequestID, Timestamp: time.Now().UTC()}:
			default:
			}
		}
		fail := func(msg string) { ev(tutor.TutorEvent{Type: tutor.EventError, Error: msg}) }

		if !ev(tutor.TutorEvent{Type: tutor.EventStarted}) {
			cancelled()
			return
		}

		url := strings.TrimRight(c.BaseURL, "/") + "/v1/responses"
		var resp *http.Response
		// Only 503 "temporarily unavailable" is retried, with bounded backoff.
		// A 429 usage limit pauses: retrying would just burn the plan's limit.
		for attempt := 0; attempt < 3; attempt++ {
			httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
			if err != nil {
				fail(err.Error())
				return
			}
			httpReq.Header.Set("Authorization", "Bearer "+token)
			httpReq.Header.Set("Content-Type", "application/json")
			httpReq.Header.Set("Accept", "text/event-stream")
			resp, err = c.HTTPClient.Do(httpReq)
			if err != nil {
				if ctx.Err() != nil {
					cancelled()
					return
				}
				fail("ChatGPT plan connection failed: " + err.Error())
				return
			}
			if resp.StatusCode != http.StatusServiceUnavailable || attempt == 2 {
				break
			}
			resp.Body.Close()
			select {
			case <-ctx.Done():
				cancelled()
				return
			case <-time.After(time.Duration(250<<attempt) * time.Millisecond):
			}
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
			code, msg := decodePlanError(body)
			if code == "subscription_sharing_invalid_user" || (resp.StatusCode == http.StatusUnauthorized && code == "") {
				c.Tokens.MarkSelectedReauth()
			}
			fail(planErrorMessage(resp.StatusCode, code, msg))
			return
		}

		var full strings.Builder
		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 0, 64*1024), 1<<20)
		completed := false
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data:") {
				continue
			}
			var e planStreamEvent
			if json.Unmarshal([]byte(strings.TrimSpace(strings.TrimPrefix(line, "data:"))), &e) != nil {
				continue
			}
			switch e.Type {
			case "response.output_text.delta":
				full.WriteString(e.Delta)
				if !ev(tutor.TutorEvent{Type: tutor.EventTextDelta, Delta: e.Delta}) {
					cancelled()
					return
				}
			case "response.completed":
				completed = true
			case "response.failed":
				code, msg := "", ""
				if e.Response != nil && e.Response.Error != nil {
					code, msg = e.Response.Error.Code, e.Response.Error.Message
				}
				fail(planErrorMessage(http.StatusOK, code, msg))
				return
			case "response.incomplete":
				reason := "unknown reason"
				if e.Response != nil && e.Response.IncompleteDetails != nil {
					reason = e.Response.IncompleteDetails.Reason
				}
				fail("ChatGPT plan response was incomplete (" + reason + ").")
				return
			}
			if completed {
				break
			}
		}
		if ctx.Err() != nil {
			cancelled()
			return
		}
		if !completed {
			// Only response.completed marks success; a dropped stream is an error.
			fail("ChatGPT plan stream ended before completion.")
			return
		}

		text := full.String()
		c.Budget.RecordTokens(len(text) / 4)
		if !ev(tutor.TutorEvent{Type: tutor.EventComplete, Text: text}) {
			cancelled()
		}
	}()
	return out, nil
}

type planModelsResponse struct {
	Models []struct {
		Slug        string `json:"slug"`
		DisplayName string `json:"display_name"`
		Visibility  string `json:"visibility"`
	} `json:"models"`
}

// DiscoverModels lists plan models with the access token. The apiKey
// argument is ignored: this route never uses API keys.
func (c *ChatGPTPlanAdapter) DiscoverModels(ctx context.Context, _ string) ([]ModelInfo, error) {
	token, err := c.Tokens.AccessToken(ctx)
	if err != nil {
		return nil, errors.New(tokenErrorMessage(err))
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(c.BaseURL, "/")+"/v1/models", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ChatGPT plan models: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		code, msg := decodePlanError(body)
		return nil, errors.New(planErrorMessage(resp.StatusCode, code, msg))
	}
	var cat planModelsResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(&cat); err != nil {
		return nil, fmt.Errorf("failed to decode ChatGPT plan models: %w", err)
	}
	models := make([]ModelInfo, 0, len(cat.Models))
	for _, m := range cat.Models {
		if m.Visibility != "list" || m.Slug == "" {
			continue
		}
		name := m.DisplayName
		if name == "" {
			name = m.Slug
		}
		models = append(models, ModelInfo{
			ID:                m.Slug,
			Name:              name,
			Provider:          auth.RouteChatGPT,
			SupportsStreaming: true,
			Description:       "Available through your ChatGPT plan (account-specific).",
			IsDefault:         m.Slug == c.DefaultModel,
		})
	}
	return models, nil
}

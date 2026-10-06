package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/trustdan/quant-methods-practice/internal/auth"
	"github.com/trustdan/quant-methods-practice/internal/siwc"
)

type chatGPTSignInRequestDTO struct {
	// AccountKey reauthorizes a saved account; empty registers a new one.
	AccountKey string `json:"account_key,omitempty"`
}

type chatGPTSignInResponseDTO struct {
	AuthorizeURL string            `json:"authorize_url"`
	Status       siwc.SignInStatus `json:"status"`
}

type chatGPTSignOutResponseDTO struct {
	Revoked  bool                  `json:"revoked"`
	Accounts []siwc.AccountSummary `json:"accounts"`
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

// handleChatGPT serves /api/providers/chatgpt/{signin,accounts,...}. It
// reports whether it handled the request so model routes fall through to
// the generic handler. The OAuth callback itself is not served here: it
// runs on a separate one-shot loopback listener owned by siwc.Client.
func (s *Server) handleChatGPT(w http.ResponseWriter, r *http.Request, parts []string) bool {
	client := s.providerManager.ChatGPT()
	if len(parts) == 0 {
		return false
	}

	switch {
	case parts[0] == "credentials":
		http.Error(w, "the ChatGPT plan route uses Sign in with ChatGPT; tokens cannot be pasted", http.StatusBadRequest)

	case parts[0] == "disconnect":
		http.Error(w, "sign out a specific ChatGPT account instead", http.StatusBadRequest)

	case parts[0] == "signin" && len(parts) == 1:
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, client.Status())
		case http.MethodPost:
			var dto chatGPTSignInRequestDTO
			if r.ContentLength != 0 {
				if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&dto); err != nil {
					http.Error(w, "invalid request body", http.StatusBadRequest)
					return true
				}
			}
			authURL, st, err := client.BeginSignIn(dto.AccountKey)
			if err != nil {
				http.Error(w, fmt.Sprintf("could not start ChatGPT sign-in: %v", err), http.StatusBadRequest)
				return true
			}
			writeJSON(w, chatGPTSignInResponseDTO{AuthorizeURL: authURL, Status: st})
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}

	case parts[0] == "signin" && len(parts) == 2 && parts[1] == "cancel":
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return true
		}
		writeJSON(w, client.CancelSignIn())

	case parts[0] == "accounts" && len(parts) == 1:
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return true
		}
		accts, err := client.Accounts()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return true
		}
		writeJSON(w, accts)

	case parts[0] == "accounts" && len(parts) == 3 && parts[2] == "select":
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return true
		}
		if err := client.Select(parts[1]); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return true
		}
		accts, _ := client.Accounts()
		writeJSON(w, accts)

	case parts[0] == "accounts" && len(parts) == 3 && parts[2] == "signout":
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return true
		}
		revoked, err := client.SignOut(r.Context(), parts[1])
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return true
		}
		// Leaving the plan route without a usable account falls back to the
		// visible offline default, never to a different billing route.
		if s.providerManager.GetActiveRoute() == auth.RouteChatGPT {
			if acct, ok := client.SelectedAccount(); !ok || acct.NeedsReauth {
				_ = s.providerManager.SetActive(auth.RouteOffline, "offline-curriculum")
			}
		}
		accts, _ := client.Accounts()
		writeJSON(w, chatGPTSignOutResponseDTO{Revoked: revoked, Accounts: accts})

	default:
		return false
	}
	return true
}

package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net"
	"net/http"
	"os"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/trustdan/quant-methods-practice/internal/auth"
	"github.com/trustdan/quant-methods-practice/internal/bank"
	"github.com/trustdan/quant-methods-practice/internal/domain"
	"github.com/trustdan/quant-methods-practice/internal/drill"
	"github.com/trustdan/quant-methods-practice/internal/mastery"
	"github.com/trustdan/quant-methods-practice/internal/providers"
	"github.com/trustdan/quant-methods-practice/internal/tutor"
)

type Config struct {
	Addr              string
	AssetsFS          fs.FS
	Version           string
	AllowedDevOrigins []string
	Bank              *bank.Bank
	SessionManager    *drill.SessionManager
	Store             drill.SessionStore
	NoteStore         NoteStore
	TutorManager      *tutor.TutorManager
	Vault             auth.Vault
	ProviderManager   *providers.ProviderManager
	CandidateStore    CandidateStore
	WorksheetStore    WorksheetStore
}

type Server struct {
	config            Config
	listener          net.Listener
	httpServer        *http.Server
	bootstrapToken    string
	bootstrapConsumed bool
	sessionToken      string
	bank              *bank.Bank
	sessionManager    *drill.SessionManager
	noteStore         NoteStore
	tutorManager      *tutor.TutorManager
	vault             auth.Vault
	providerManager   *providers.ProviderManager
	candidateStore    CandidateStore
	candidateState    candidateState
	worksheetStore    WorksheetStore
	lastSessionID     string
	mu                sync.RWMutex
}

func NewServer(cfg Config) (*Server, error) {
	if cfg.Addr == "" {
		cfg.Addr = "127.0.0.1:0"
	}
	if cfg.Version == "" {
		cfg.Version = "0.1.0-dev"
	}
	if cfg.SessionManager == nil {
		if cfg.Store != nil {
			cfg.SessionManager = drill.NewSessionManagerWithStore(cfg.Store, nil)
		} else {
			cfg.SessionManager = drill.NewSessionManager(nil)
		}
	}
	if cfg.Bank == nil {
		bankDirs := []string{"curriculum/approved", "../curriculum/approved", "../../curriculum/approved"}
		for _, dir := range bankDirs {
			if fi, err := os.Stat(dir); err == nil && fi.IsDir() {
				b, err := bank.LoadActiveBank(dir, nil)
				if err == nil && b.Count() > 0 {
					cfg.Bank = b
					break
				}
			}
		}
		if cfg.Bank == nil {
			cfg.Bank = bank.NewBank()
		}
	}

	tokenBytes := make([]byte, 24)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random bootstrap token: %w", err)
	}
	bootstrapToken := hex.EncodeToString(tokenBytes)

	noteStore := cfg.NoteStore
	if noteStore == nil && cfg.Store != nil {
		if ns, ok := cfg.Store.(NoteStore); ok {
			noteStore = ns
		}
	}

	tutorManager := cfg.TutorManager
	if tutorManager == nil {
		tutorManager = tutor.NewTutorManager(nil)
	}

	vault := cfg.Vault
	if vault == nil {
		vault = auth.NewMemoryVault()
	}

	providerManager := cfg.ProviderManager
	if providerManager == nil {
		budget := providers.NewBudgetTracker(providers.DefaultMaxRequestsPerSession)
		catalog := providers.NewCatalogCache()
		providerManager = providers.NewProviderManager(vault, budget, catalog)
	}

	// Register external providers in tutorManager
	for _, route := range []string{auth.RouteAnthropic, auth.RouteGemini, auth.RouteOpenAI, auth.RouteChatGPT} {
		if svc, err := providerManager.GetProvider(route); err == nil && svc != nil {
			tutorManager.RegisterProvider(svc)
		}
	}

	s := &Server{
		config:          cfg,
		bootstrapToken:  bootstrapToken,
		bank:            cfg.Bank,
		sessionManager:  cfg.SessionManager,
		noteStore:       noteStore,
		tutorManager:    tutorManager,
		vault:           vault,
		providerManager: providerManager,
	}

	s.candidateStore = cfg.CandidateStore
	s.worksheetStore = cfg.WorksheetStore
	if s.worksheetStore == nil && cfg.Store != nil {
		s.worksheetStore, _ = cfg.Store.(WorksheetStore)
	}
	if s.candidateStore == nil && cfg.Store != nil {
		s.candidateStore, _ = cfg.Store.(CandidateStore)
	}
	if s.candidateStore != nil {
		records, err := s.candidateStore.ListCandidates(context.Background())
		if err != nil {
			return nil, fmt.Errorf("loading candidate bank: %w", err)
		}
		for _, rec := range records {
			if rec.Status != "approved" {
				continue
			}
			if _, exists := s.bank.Get(rec.ID); exists {
				return nil, fmt.Errorf("duplicate approved candidate id %q", rec.ID)
			}
			if err := bank.ValidateTemplate(rec.Template, nil); err != nil {
				return nil, fmt.Errorf("invalid approved candidate: %w", err)
			}
			s.bank.Add(rec.Template)
		}
	}

	if s.sessionManager != nil && s.bank != nil {
		s.sessionManager.SetContrastFinder(func(originID, misID string) (*domain.QuestionTemplate, bool) {
			partnerID, found := mastery.FindContrastPartner(originID, misID)
			if !found {
				return nil, false
			}
			partnerTmpl, ok := s.bank.Get(partnerID)
			if !ok || partnerTmpl.Status != domain.StatusActive {
				return nil, false
			}
			return partnerTmpl, true
		})
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/api/local-session", s.handleLocalSession)
	mux.HandleFunc("/api/settings", s.handleSettings)
	mux.HandleFunc("/api/bank", s.handleBank)
	mux.HandleFunc("/api/mastery", s.handleMastery)
	mux.HandleFunc("/api/practice/sessions", s.handlePracticeSessions)
	mux.HandleFunc("/api/practice/sessions/", s.handlePracticeSessionByID)
	mux.HandleFunc("/api/tutor/requests", s.handleTutorRequests)
	mux.HandleFunc("/api/tutor/requests/", s.handleTutorRequestByID)
	mux.HandleFunc("/api/candidates", s.handleCandidates)
	mux.HandleFunc("/api/candidates/", s.handleCandidates)
	mux.HandleFunc("/api/worksheets", s.handleWorksheets)
	mux.HandleFunc("/api/worksheets/", s.handleWorksheets)
	mux.HandleFunc("/api/notes", s.handleNotes)
	mux.HandleFunc("/api/notes/", s.handleNoteByID)
	mux.HandleFunc("/api/exports", s.handleExports)
	mux.HandleFunc("/api/tutor/drafts/", s.handleTutorDrafts)
	mux.HandleFunc("/api/providers", s.handleProviders)
	mux.HandleFunc("/api/providers/", s.handleProviders)
	mux.HandleFunc("/", s.handleStaticOrSPA)

	wrapped := s.securityMiddleware(mux)

	s.httpServer = &http.Server{
		Handler:      wrapped,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s, nil
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.config.Addr)
	if err != nil {
		return fmt.Errorf("failed to bind loopback listener: %w", err)
	}
	s.listener = ln

	go func() {
		_ = s.httpServer.Serve(ln)
	}()

	return nil
}

func (s *Server) Close() error {
	if s.httpServer != nil {
		return s.httpServer.Close()
	}
	return nil
}

func (s *Server) Addr() string {
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return s.config.Addr
}

func (s *Server) BootstrapURL() string {
	return fmt.Sprintf("http://%s/#bootstrap=%s", s.Addr(), s.bootstrapToken)
}

func (s *Server) BootstrapToken() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.bootstrapToken
}

func (s *Server) securityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Host validation: must be loopback to prevent DNS rebinding attacks
		host := r.Host
		if colon := strings.LastIndex(host, ":"); colon != -1 {
			host = host[:colon]
		}
		if host != "127.0.0.1" && host != "localhost" && host != "::1" && host != "[::1]" {
			http.Error(w, "invalid host header: loopback access only", http.StatusBadRequest)
			return
		}

		// Origin validation for state-changing requests
		if r.Method == http.MethodPost || r.Method == http.MethodPatch || r.Method == http.MethodDelete || r.Method == http.MethodPut {
			origin := r.Header.Get("Origin")
			if origin != "" && !s.isAllowedOrigin(origin) {
				http.Error(w, "cross-origin mutation rejected", http.StatusForbidden)
				return
			}
		}

		// Security headers
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; font-src 'self'; img-src 'self' data:; connect-src 'self'; frame-ancestors 'none';")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")

		next.ServeHTTP(w, r)
	})
}

func (s *Server) isAllowedOrigin(origin string) bool {
	// Strip scheme
	trimmed := strings.TrimPrefix(origin, "http://")
	trimmed = strings.TrimPrefix(trimmed, "https://")

	host := trimmed
	if colon := strings.LastIndex(host, ":"); colon != -1 {
		host = host[:colon]
	}

	if host == "127.0.0.1" || host == "localhost" || host == "::1" || host == "[::1]" {
		return true
	}

	for _, allowed := range s.config.AllowedDevOrigins {
		if origin == allowed {
			return true
		}
	}

	return false
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":    "ok",
		"version":   s.config.Version,
		"toolchain": "go1.27.1",
	})
}

type localSessionRequest struct {
	BootstrapToken string `json:"bootstrap_token"`
}

type localSessionResponse struct {
	Status       string `json:"status"`
	SessionToken string `json:"session_token"`
}

func (s *Server) handleLocalSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req localSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.bootstrapConsumed || req.BootstrapToken != s.bootstrapToken {
		http.Error(w, "invalid or expired bootstrap token", http.StatusUnauthorized)
		return
	}

	// Consume bootstrap token
	s.bootstrapConsumed = true

	// Generate persistent application session token
	sessionBytes := make([]byte, 24)
	_, _ = rand.Read(sessionBytes)
	s.sessionToken = hex.EncodeToString(sessionBytes)

	// Set HttpOnly SameSite cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "quant_session",
		Value:    s.sessionToken,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(localSessionResponse{
		Status:       "ok",
		SessionToken: s.sessionToken,
	})
}

type createSessionRequest struct {
	TemplateID    string   `json:"template_id,omitempty"`
	QuestionCount int      `json:"question_count,omitempty"`
	ModuleIDs     []string `json:"module_ids,omitempty"`
	Intensity     string   `json:"intensity,omitempty"`
	Seed          int64    `json:"seed,omitempty"`
}

func (s *Server) handlePracticeSessions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var req createSessionRequest
		if r.Body != nil && r.ContentLength > 0 {
			_ = json.NewDecoder(r.Body).Decode(&req)
		}

		s.mu.Lock()
		defer s.mu.Unlock()

		if s.bank == nil || s.bank.Count() == 0 {
			http.Error(w, "no approved question templates available", http.StatusBadRequest)
			return
		}

		var sess *drill.DrillSession
		var err error

		if req.TemplateID != "" {
			tmpl, ok := s.bank.Get(req.TemplateID)
			if !ok {
				http.Error(w, fmt.Sprintf("template %q not found", req.TemplateID), http.StatusBadRequest)
				return
			}
			sess, err = s.sessionManager.CreateSession(tmpl, req.Seed)
		} else {
			allTemplates := s.bank.List()
			var filtered []*domain.QuestionTemplate
			if len(req.ModuleIDs) > 0 {
				modMap := make(map[string]bool)
				for _, m := range req.ModuleIDs {
					modMap[m] = true
				}
				for _, tmpl := range allTemplates {
					if modMap[tmpl.ModuleID] {
						filtered = append(filtered, tmpl)
					}
				}
			}
			if len(filtered) == 0 {
				filtered = allTemplates
			}

			// Ensure canonical introductory template is placed first if present
			sort.SliceStable(filtered, func(i, j int) bool {
				if filtered[i].ID == "binomial_fair_coin_exactly_two" {
					return true
				}
				if filtered[j].ID == "binomial_fair_coin_exactly_two" {
					return false
				}
				return filtered[i].ID < filtered[j].ID
			})

			qCount := req.QuestionCount
			if qCount <= 0 {
				if s.bank.Count() == 1 {
					qCount = 1
				} else {
					qCount = 10
				}
			}

			// Load learner's mastery profile from store if available
			var masteryMap map[string]*mastery.ConceptMastery
			if s.config.Store != nil {
				if storeWithMastery, ok := s.config.Store.(interface {
					GetAllMasteryProjections(ctx context.Context) (map[string]*mastery.ConceptMastery, error)
				}); ok {
					masteryMap, _ = storeWithMastery.GetAllMasteryProjections(r.Context())
				}
			}

			// Seeded weighted selection with floor, mixed review, and anti-repeat penalties
			selectionOpts := mastery.SelectionOptions{
				CandidateTemplates: filtered,
				MasteryMap:         masteryMap,
				QuestionCount:      qCount,
				Intensity:          req.Intensity,
				Seed:               req.Seed,
				Now:                time.Now(),
			}
			selected := mastery.SelectQuestions(selectionOpts)
			if len(selected) == 0 {
				for i := 0; i < qCount; i++ {
					selected = append(selected, filtered[i%len(filtered)])
				}
			}

			// Determine scaffold level for each problem
			scaffolds := make([]mastery.ScaffoldLevel, len(selected))
			for i, tmpl := range selected {
				scaffolds[i] = mastery.DetermineTemplateScaffold(tmpl, masteryMap)
			}

			settings := domain.SessionSettings{
				QuestionCount: qCount,
				ModuleIDs:     req.ModuleIDs,
				Intensity:     req.Intensity,
			}
			if settings.Intensity == "" {
				settings.Intensity = "standard"
			}

			sess, err = s.sessionManager.CreateMultiQuestionSessionWithScaffolds(selected, scaffolds, settings, req.Seed)
		}

		if err != nil {
			http.Error(w, fmt.Sprintf("failed to create session: %v", err), http.StatusInternalServerError)
			return
		}

		s.lastSessionID = sess.ID
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(sess.ToPublicView())

	case http.MethodGet:
		s.mu.RLock()
		lastID := s.lastSessionID
		s.mu.RUnlock()

		if lastID != "" {
			if sess, ok := s.sessionManager.GetSession(lastID); ok {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(sess.ToPublicView())
				return
			}
		}

		// Check persistent storage for an active session
		if sess, ok := s.sessionManager.GetActiveSession(); ok {
			s.mu.Lock()
			s.lastSessionID = sess.ID
			s.mu.Unlock()
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(sess.ToPublicView())
			return
		}

		// If no active session, create one
		s.mu.Lock()
		defer s.mu.Unlock()

		if s.bank == nil || s.bank.Count() == 0 {
			http.Error(w, "no approved question templates available", http.StatusNotFound)
			return
		}

		allTemplates := s.bank.List()
		sort.SliceStable(allTemplates, func(i, j int) bool {
			if allTemplates[i].ID == "binomial_fair_coin_exactly_two" {
				return true
			}
			if allTemplates[j].ID == "binomial_fair_coin_exactly_two" {
				return false
			}
			return allTemplates[i].ID < allTemplates[j].ID
		})
		qCount := 10
		if s.bank.Count() < 10 {
			qCount = s.bank.Count()
		}
		selected := make([]*domain.QuestionTemplate, 0, qCount)
		for i := 0; i < qCount; i++ {
			selected = append(selected, allTemplates[i%len(allTemplates)])
		}

		settings := domain.SessionSettings{
			QuestionCount: qCount,
			Intensity:     "standard",
		}
		sess, err := s.sessionManager.CreateMultiQuestionSession(selected, settings, 0)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to create session: %v", err), http.StatusInternalServerError)
			return
		}

		s.lastSessionID = sess.ID
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(sess.ToPublicView())

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		if s.config.Store != nil {
			if val, err := s.config.Store.GetSettings(r.Context(), "user_preferences"); err == nil && val != "" {
				_, _ = w.Write([]byte(val))
				return
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"question_count": 10,
			"module_ids":     []string{},
			"intensity":      "standard",
		})

	case http.MethodPost:
		body, err := io.ReadAll(io.LimitReader(r.Body, 64*1024))
		if err != nil {
			http.Error(w, "failed to read request body", http.StatusBadRequest)
			return
		}
		if s.config.Store != nil {
			_ = s.config.Store.SaveSettings(r.Context(), "user_preferences", string(body))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

type templateSummary struct {
	ID          string   `json:"id"`
	Version     int      `json:"version"`
	Title       string   `json:"title"`
	ModuleID    string   `json:"module_id"`
	FamilyID    string   `json:"family_id"`
	ConceptIDs  []string `json:"concept_ids"`
	StagesCount int      `json:"stages_count"`
}

func (s *Server) handleBank(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if s.bank == nil {
		_ = json.NewEncoder(w).Encode([]templateSummary{})
		return
	}

	list := s.bank.List()
	summaries := make([]templateSummary, len(list))
	for i, tmpl := range list {
		summaries[i] = templateSummary{
			ID:          tmpl.ID,
			Version:     tmpl.Version,
			Title:       tmpl.Title,
			ModuleID:    tmpl.ModuleID,
			FamilyID:    tmpl.FamilyID,
			ConceptIDs:  tmpl.ConceptIDs,
			StagesCount: len(tmpl.Stages),
		}
	}
	_ = json.NewEncoder(w).Encode(summaries)
}

func (s *Server) handleMastery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if s.config.Store != nil {
		if storeWithMastery, ok := s.config.Store.(interface {
			GetMasterySummary(ctx context.Context, ledger *mastery.EvidenceLedger, now time.Time) (*mastery.MasterySummary, error)
		}); ok {
			summary, err := storeWithMastery.GetMasterySummary(r.Context(), nil, time.Now())
			if err == nil && summary != nil {
				_ = json.NewEncoder(w).Encode(summary)
				return
			}
		}
	}

	// Fallback to empty summary if store not available or error
	_ = json.NewEncoder(w).Encode(map[string]any{
		"policy_version":     mastery.EvidencePolicyVersion,
		"overall_score":      0.0,
		"total_mastered":     0,
		"total_transferring": 0,
		"total_learning":     0,
		"total_new":          0,
		"concepts":           []any{},
		"generated_at":       time.Now(),
	})
}

func (s *Server) handlePracticeSessionByID(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(r.URL.Path, "/api/practice/sessions/")
	parts := strings.Split(strings.Trim(rest, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}

	sessionID := parts[0]
	sess, ok := s.sessionManager.GetSession(sessionID)
	if !ok {
		http.Error(w, fmt.Sprintf("session %q not found", sessionID), http.StatusNotFound)
		return
	}

	if len(parts) == 1 {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(sess.ToPublicView())
		return
	}

	if len(parts) == 2 && parts[1] == "commands" {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var cmd drill.SessionCommand
		if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
			http.Error(w, fmt.Sprintf("invalid command body: %v", err), http.StatusBadRequest)
			return
		}

		res, err := sess.ExecuteCommand(cmd, nil)
		w.Header().Set("Content-Type", "application/json")
		if err == drill.ErrRevisionConflict {
			w.WriteHeader(http.StatusConflict)
		} else if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		if res != nil {
			_ = json.NewEncoder(w).Encode(res)
		} else {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"success":       false,
				"error_message": err.Error(),
			})
		}
		return
	}

	http.NotFound(w, r)
}

func (s *Server) handleStaticOrSPA(w http.ResponseWriter, r *http.Request) {
	if s.config.AssetsFS == nil {
		http.NotFound(w, r)
		return
	}

	reqPath := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
	if reqPath == "" || reqPath == "." {
		reqPath = "index.html"
	}

	file, err := s.config.AssetsFS.Open(reqPath)
	if err != nil {
		// SPA fallback: return index.html for non-API client-side routes
		if !strings.HasPrefix(r.URL.Path, "/api/") {
			indexFile, indexErr := s.config.AssetsFS.Open("index.html")
			if indexErr == nil {
				defer indexFile.Close()
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				_, _ = io.Copy(w, indexFile)
				return
			}
		}
		http.NotFound(w, r)
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil || stat.IsDir() {
		// If requesting a directory, serve index.html
		indexFile, indexErr := s.config.AssetsFS.Open("index.html")
		if indexErr == nil {
			defer indexFile.Close()
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = io.Copy(w, indexFile)
			return
		}
		http.NotFound(w, r)
		return
	}

	ext := path.Ext(reqPath)
	ctype := mime.TypeByExtension(ext)
	if ctype == "" {
		switch ext {
		case ".html":
			ctype = "text/html; charset=utf-8"
		case ".js", ".mjs":
			ctype = "text/javascript; charset=utf-8"
		case ".css":
			ctype = "text/css; charset=utf-8"
		case ".svg":
			ctype = "image/svg+xml"
		case ".json":
			ctype = "application/json"
		case ".woff2":
			ctype = "font/woff2"
		case ".woff":
			ctype = "font/woff"
		default:
			ctype = "application/octet-stream"
		}
	}

	w.Header().Set("Content-Type", ctype)
	_, _ = io.Copy(w, file)
}

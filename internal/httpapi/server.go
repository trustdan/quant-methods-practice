package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"
)

type Config struct {
	Addr              string
	AssetsFS          fs.FS
	Version           string
	AllowedDevOrigins []string
}

type Server struct {
	config            Config
	listener          net.Listener
	httpServer        *http.Server
	bootstrapToken    string
	bootstrapConsumed bool
	sessionToken      string
	mu                sync.RWMutex
}

func NewServer(cfg Config) (*Server, error) {
	if cfg.Addr == "" {
		cfg.Addr = "127.0.0.1:0"
	}
	if cfg.Version == "" {
		cfg.Version = "0.1.0-dev"
	}

	tokenBytes := make([]byte, 24)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random bootstrap token: %w", err)
	}
	bootstrapToken := hex.EncodeToString(tokenBytes)

	s := &Server{
		config:         cfg,
		bootstrapToken: bootstrapToken,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/api/local-session", s.handleLocalSession)
	mux.HandleFunc("/", s.handleStaticOrSPA)

	wrapped := s.securityMiddleware(mux)

	s.httpServer = &http.Server{
		Handler:      wrapped,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
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

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/trustdan/quant-methods-practice/internal/storage"
)

// ArcadeStore defines the persistence contract for arcade scores.
type ArcadeStore interface {
	RecordArcadeScore(ctx context.Context, rec storage.ArcadeScoreRecord) error
	GetTopArcadeScores(ctx context.Context, limit int) ([]storage.ArcadeScoreRecord, error)
	GetHighScore(ctx context.Context) (int, error)
}

type arcadeScoresResponse struct {
	Scores    []storage.ArcadeScoreRecord `json:"scores"`
	HighScore int                         `json:"high_score"`
}

type submitArcadeScoreRequest struct {
	GameVersion     string  `json:"game_version"`
	Seed            int64   `json:"seed"`
	IsDemo          bool    `json:"is_demo"`
	DurationSeconds float64 `json:"duration_seconds"`
	SectorsCleared  int     `json:"sectors_cleared"`
	Score           int     `json:"score"`
	Initials        string  `json:"initials"`
}

func (s *Server) handleArcadeScores(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		if s.arcadeStore == nil {
			_ = json.NewEncoder(w).Encode(arcadeScoresResponse{
				Scores:    []storage.ArcadeScoreRecord{},
				HighScore: 0,
			})
			return
		}

		scores, err := s.arcadeStore.GetTopArcadeScores(r.Context(), 10)
		if err != nil {
			scores = []storage.ArcadeScoreRecord{}
		}
		highScore, err := s.arcadeStore.GetHighScore(r.Context())
		if err != nil {
			highScore = 0
		}

		_ = json.NewEncoder(w).Encode(arcadeScoresResponse{
			Scores:    scores,
			HighScore: highScore,
		})

	case http.MethodPost:
		var req submitArcadeScoreRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid request json"}`, http.StatusBadRequest)
			return
		}

		// Re-enforce invariant: Demo/attract play can never enter the human leaderboard
		if req.IsDemo {
			http.Error(w, `{"error":"demo scores cannot enter human leaderboard"}`, http.StatusBadRequest)
			return
		}
		if req.Score < 0 {
			http.Error(w, `{"error":"score cannot be negative"}`, http.StatusBadRequest)
			return
		}

		initials := strings.ToUpper(strings.TrimSpace(req.Initials))
		if initials == "" {
			initials = "PIL"
		}
		if len(initials) > 3 {
			initials = initials[:3]
		}

		version := req.GameVersion
		if version == "" {
			version = s.config.Version
		}

		rec := storage.ArcadeScoreRecord{
			GameVersion:     version,
			Seed:            req.Seed,
			IsDemo:          false,
			DurationSeconds: req.DurationSeconds,
			SectorsCleared:  req.SectorsCleared,
			Score:           req.Score,
			Initials:        initials,
			CreatedAt:       time.Now().UTC(),
		}

		if s.arcadeStore != nil {
			if err := s.arcadeStore.RecordArcadeScore(r.Context(), rec); err != nil {
				http.Error(w, `{"error":"failed to record score"}`, http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"record":  rec,
		})

	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

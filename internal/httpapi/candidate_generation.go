package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/trustdan/quant-methods-practice/internal/candidates"
	"github.com/trustdan/quant-methods-practice/internal/tutor"
	"strings"
)

// generateCandidateWording consumes bounded provider output and requires a completed stream.
// It has no store, bank, or learning-state access.
func generateCandidateWording(ctx context.Context, svc tutor.TutorService, base candidates.Proposal) (candidates.Proposal, error) {
	data, _ := json.Marshal(base)
	req := tutor.TutorRequest{RequestID: tutor.GenerateRequestID(), Action: tutor.ActionCandidate, CustomPrompt: string(data)}
	stream, err := svc.Stream(ctx, req)
	if err != nil {
		return base, err
	}
	var body strings.Builder
	complete := false
	for {
		select {
		case <-ctx.Done():
			return base, ctx.Err()
		case e, ok := <-stream:
			if !ok {
				if !complete {
					return base, errors.New("candidate generation did not complete")
				}
				p, err := candidates.Decode([]byte(body.String()))
				if err != nil {
					return base, err
				}
				if p.FamilyID != base.FamilyID || p.N != base.N || p.P != base.P || p.K != base.K {
					return base, errors.New("AI changed fixed parameters; candidate rejected")
				}
				return p, nil
			}
			switch e.Type {
			case tutor.EventTextDelta:
				if body.Len()+len(e.Delta) > 12000 {
					return base, errors.New("candidate output exceeds 12000 bytes")
				}
				body.WriteString(e.Delta)
			case tutor.EventComplete:
				complete = true
			case tutor.EventError, tutor.EventFallback, tutor.EventCancelled:
				return base, errors.New("candidate generation failed; no fallback was published")
			}
		}
	}
}

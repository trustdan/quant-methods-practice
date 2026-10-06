package drill

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/trustdan/quant-methods-practice/internal/domain"
	"github.com/trustdan/quant-methods-practice/internal/mastery"
)

var (
	ErrRevisionConflict   = errors.New("revision conflict: stale session revision")
	ErrStageNotActive     = errors.New("stage is not active for submission")
	ErrStageAlreadyDone   = errors.New("stage is already completed")
	ErrInvalidStageIndex  = errors.New("invalid or ineligible stage index")
	ErrUnknownCommandType = errors.New("unknown command type")
)

// SessionManager manages active drill sessions thread-safely.
type SessionManager struct {
	mu             sync.RWMutex
	sessions       map[string]*DrillSession
	store          SessionStore
	clock          func() time.Time
	contrastFinder func(originID, misID string) (*domain.QuestionTemplate, bool)
}

// NewSessionManager creates a manager for drill sessions.
func NewSessionManager(clock func() time.Time) *SessionManager {
	if clock == nil {
		clock = time.Now
	}
	return &SessionManager{
		sessions: make(map[string]*DrillSession),
		clock:    clock,
	}
}

// NewSessionManagerWithStore creates a manager with a persistent storage backend.
func NewSessionManagerWithStore(store SessionStore, clock func() time.Time) *SessionManager {
	sm := NewSessionManager(clock)
	sm.store = store
	return sm
}

// SetStore assigns a persistence store to the session manager.
func (sm *SessionManager) SetStore(store SessionStore) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.store = store
}

// SetContrastFinder registers a contrast partner callback across managed sessions.
func (sm *SessionManager) SetContrastFinder(fn func(originID, misID string) (*domain.QuestionTemplate, bool)) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.contrastFinder = fn
}

// CreateMultiQuestionSession instantiates a practice session with multiple question templates.
func (sm *SessionManager) CreateMultiQuestionSession(tmpls []*domain.QuestionTemplate, settings domain.SessionSettings, seed int64) (*DrillSession, error) {
	return sm.CreateMultiQuestionSessionWithScaffolds(tmpls, nil, settings, seed)
}

// CreateMultiQuestionSessionWithScaffolds instantiates a practice session applying specified scaffold levels.
func (sm *SessionManager) CreateMultiQuestionSessionWithScaffolds(
	tmpls []*domain.QuestionTemplate,
	scaffolds []mastery.ScaffoldLevel,
	settings domain.SessionSettings,
	seed int64,
) (*DrillSession, error) {
	if len(tmpls) == 0 {
		return nil, errors.New("at least one template is required")
	}

	sessionID := fmt.Sprintf("drill_%d_s%d", sm.clock().UnixNano(), seed)
	qStates := make([]QuestionState, len(tmpls))

	for idx, tmpl := range tmpls {
		qSeed := seed
		if seed != 0 {
			qSeed = seed + int64(idx*1000)
		}
		scaffLevel := mastery.ScaffoldFull
		if idx < len(scaffolds) && scaffolds[idx] != "" {
			scaffLevel = scaffolds[idx]
		}
		rawStages := mastery.DegradeStages(tmpl, scaffLevel)
		if len(rawStages) == 0 {
			rawStages = tmpl.Stages
		}

		qInst := domain.QuestionInstance{
			ID:               fmt.Sprintf("inst_%s_%d", tmpl.ID, qSeed),
			TemplateID:       tmpl.ID,
			TemplateVersion:  tmpl.Version,
			Seed:             qSeed,
			Parameters:       tmpl.Parameters,
			Title:            tmpl.Title,
			ScenarioMarkdown: tmpl.ScenarioMarkdown,
			Assumptions:      tmpl.Assumptions,
			SettingGroup:     tmpl.SettingGroup,
			Stages:           make([]domain.StageInstance, len(rawStages)),
		}

		var rng *rand.Rand
		if qSeed != 0 {
			rng = rand.New(rand.NewSource(qSeed))
		}

		stages := make([]StageState, len(rawStages))
		for i, st := range rawStages {
			stageInst := domain.StageInstance{
				ID:                  st.ID,
				Kind:                st.Kind,
				PromptMarkdown:      st.PromptMarkdown,
				Options:             make([]domain.Option, len(st.Options)),
				ExpectedAnswer:      st.ExpectedAnswer,
				EvidenceConceptIDs:  st.EvidenceConceptIDs,
				ExplanationMarkdown: st.ExplanationMarkdown,
				NumericPolicy:       st.NumericPolicy,
			}
			copy(stageInst.Options, st.Options)

			if rng != nil && len(stageInst.Options) > 1 {
				rng.Shuffle(len(stageInst.Options), func(a, b int) {
					stageInst.Options[a], stageInst.Options[b] = stageInst.Options[b], stageInst.Options[a]
				})
			}

			qInst.Stages[i] = stageInst

			status := StageStatusUnvisited
			if i == 0 {
				status = StageStatusActive
			}

			stages[i] = StageState{
				Instance: stageInst,
				Status:   status,
				Attempts: make([]domain.StageAttempt, 0, 2),
			}
		}

		qStates[idx] = QuestionState{
			Index:             idx,
			TemplateID:        tmpl.ID,
			TemplateVersion:   tmpl.Version,
			QuestionInstance:  qInst,
			Stages:            stages,
			CurrentStageIndex: 0,
			Completed:         false,
			ScaffoldLevel:     string(scaffLevel),
		}
	}

	firstQ := qStates[0]
	s := &DrillSession{
		ID:                   sessionID,
		TemplateID:           firstQ.TemplateID,
		TemplateVersion:      firstQ.TemplateVersion,
		Seed:                 seed,
		Revision:             1,
		QuestionInstance:     firstQ.QuestionInstance,
		Stages:               firstQ.Stages,
		CurrentStageIndex:    0,
		Completed:            false,
		CreatedAt:            sm.clock(),
		UpdatedAt:            sm.clock(),
		CurrentQuestionIndex: 0,
		Questions:            qStates,
		Settings:             settings,
		commandCache:         make(map[string]*CommandResult),
		store:                sm.store,
		contrastFinder:       sm.contrastFinder,
	}

	if sm.store != nil {
		if err := sm.store.SaveSession(context.Background(), s); err != nil {
			return nil, fmt.Errorf("failed to persist initial session: %w", err)
		}
	}

	sm.mu.Lock()
	sm.sessions[s.ID] = s
	sm.mu.Unlock()

	return s, nil
}

// CreateSession instantiates a new drill session from a single approved template.
func (sm *SessionManager) CreateSession(tmpl *domain.QuestionTemplate, seed int64) (*DrillSession, error) {
	if tmpl == nil {
		return nil, errors.New("template cannot be nil")
	}
	settings := domain.SessionSettings{
		QuestionCount: 1,
		ModuleIDs:     []string{tmpl.ModuleID},
		Intensity:     "standard",
	}
	return sm.CreateMultiQuestionSession([]*domain.QuestionTemplate{tmpl}, settings, seed)
}

// GetSession returns a session by ID, querying memory or persistent store.
func (sm *SessionManager) GetSession(id string) (*DrillSession, bool) {
	sm.mu.RLock()
	s, ok := sm.sessions[id]
	sm.mu.RUnlock()
	if ok {
		return s, true
	}

	if sm.store != nil {
		sess, err := sm.store.GetSession(context.Background(), id)
		if err == nil && sess != nil {
			sess.SetStore(sm.store)
			sm.mu.Lock()
			sm.sessions[sess.ID] = sess
			sm.mu.Unlock()
			return sess, true
		}
	}

	return nil, false
}

// GetActiveSession retrieves the current active drill session from memory or store.
func (sm *SessionManager) GetActiveSession() (*DrillSession, bool) {
	if sm.store != nil {
		sess, err := sm.store.GetLatestActiveSession(context.Background())
		if err == nil && sess != nil {
			sess.SetStore(sm.store)
			sm.mu.Lock()
			sm.sessions[sess.ID] = sess
			sm.mu.Unlock()
			return sess, true
		}
	}

	sm.mu.RLock()
	defer sm.mu.RUnlock()
	var latest *DrillSession
	for _, s := range sm.sessions {
		if !s.Completed {
			if latest == nil || s.UpdatedAt.After(latest.UpdatedAt) {
				latest = s
			}
		}
	}
	if latest != nil {
		return latest, true
	}
	for _, s := range sm.sessions {
		if latest == nil || s.UpdatedAt.After(latest.UpdatedAt) {
			latest = s
		}
	}
	if latest != nil {
		return latest, true
	}
	return nil, false
}

// ExecuteCommand executes a command against a session, ensuring idempotency and revision checks.
func (s *DrillSession) ExecuteCommand(cmd SessionCommand, clock func() time.Time) (*CommandResult, error) {
	if clock == nil {
		clock = time.Now
	}

	// 1. Check idempotency cache (memory and persistent store)
	if cmd.CommandID != "" {
		if cached, ok := s.GetCachedResult(cmd.CommandID); ok {
			return cached, nil
		}
		if s.store != nil {
			if cached, err := s.store.GetCommandResult(context.Background(), cmd.CommandID); err == nil && cached != nil {
				s.SetCachedResult(cmd.CommandID, cached)
				return cached, nil
			}
		}
	}

	// 2. Concurrency revision check
	if cmd.ExpectedRevision > 0 && cmd.ExpectedRevision != s.Revision {
		return &CommandResult{
			Success:      false,
			CommandID:    cmd.CommandID,
			ErrorMessage: fmt.Sprintf("revision conflict: expected %d, current is %d", cmd.ExpectedRevision, s.Revision),
			SessionState: s.ToPublicView(),
		}, ErrRevisionConflict
	}

	var res *CommandResult
	var err error

	switch cmd.Type {
	case CmdSubmitAnswer:
		res, err = s.handleSubmit(cmd, clock)
	case CmdRequestHint:
		res, err = s.handleRequestHint(cmd, clock)
	case CmdNavigateStage:
		res, err = s.handleNavigate(cmd, clock)
	case CmdNavigateQuestion:
		res, err = s.handleNavigateQuestion(cmd, clock)
	case CmdResetDrill:
		res, err = s.handleReset(cmd, clock)
	case CmdSaveDraft:
		res, err = s.handleSaveDraft(cmd, clock)
	case CmdClearDraft:
		res, err = s.handleClearDraft(cmd, clock)
	default:
		return nil, fmt.Errorf("%w: %q", ErrUnknownCommandType, cmd.Type)
	}

	if err != nil {
		return res, err
	}

	// Synchronize current question state
	s.syncCurrentQuestion()

	// Persist session mutation and command outcome
	if s.store != nil {
		if saveErr := s.store.SaveSession(context.Background(), s); saveErr != nil {
			return nil, fmt.Errorf("failed to persist session to store: %w", saveErr)
		}
		if cmd.CommandID != "" && res != nil {
			_ = s.store.SaveCommandResult(context.Background(), s.ID, cmd.CommandID, cmd.Type, res)
		}
	}

	// Cache command result if command ID was provided
	if cmd.CommandID != "" && res != nil {
		s.SetCachedResult(cmd.CommandID, res)
	}

	return res, nil
}

func (s *DrillSession) handleSubmit(cmd SessionCommand, clock func() time.Time) (*CommandResult, error) {
	if s.CurrentStageIndex < 0 || s.CurrentStageIndex >= len(s.Stages) {
		return nil, ErrInvalidStageIndex
	}

	stage := &s.Stages[s.CurrentStageIndex]

	if stage.Status == StageStatusCompleted {
		return &CommandResult{
			Success:      false,
			CommandID:    cmd.CommandID,
			ErrorMessage: "Stage is already completed",
			SessionState: s.ToPublicView(),
		}, ErrStageAlreadyDone
	}

	if cmd.Answer == nil {
		return &CommandResult{
			Success:      false,
			CommandID:    cmd.CommandID,
			InvalidInput: true,
			ErrorMessage: "Answer is required",
			SessionState: s.ToPublicView(),
		}, nil
	}

	// Grade submission
	att, invalid, diag, err := GradeSubmission(s.ID, s.QuestionInstance.ID, stage, *cmd.Answer, clock)
	if err != nil {
		return nil, err
	}

	if invalid {
		stage.InvalidInputNotice = diag
		return &CommandResult{
			Success:      false,
			CommandID:    cmd.CommandID,
			InvalidInput: true,
			ErrorMessage: diag,
			SessionState: s.ToPublicView(),
		}, nil
	}

	// Valid input: clear any previous notice
	stage.InvalidInputNotice = ""
	stage.Attempts = append(stage.Attempts, att)

	if att.IsCorrect {
		if len(stage.Attempts) == 1 {
			stage.FirstTryCorrect = true
		} else {
			stage.SolvedOnRetry = true
			stage.Assistance = append(stage.Assistance, domain.AssistanceRetry)
		}
		stage.Status = StageStatusCompleted
		stage.LastFeedback = att.FeedbackMarkdown

		// Check overall drill completion
		if s.checkAllCompleted() {
			s.Completed = true
		} else {
			// Advance to next uncompleted stage automatically if current
			s.advanceToNextStage()
		}
	} else {
		// Incorrect
		if len(stage.Attempts) == 1 {
			// First error: provide causal hint, transition to retry
			stage.Status = StageStatusRetry
			stage.Assistance = append(stage.Assistance, domain.AssistanceHint)
			stage.LastFeedback = att.FeedbackMarkdown

			// Check for eligible misconception to queue contrast partner
			if stage.MisconceptionID != nil && !s.CurrentQuestionIsContrast() && s.ContrastCount == 0 && s.contrastFinder != nil {
				if partnerTmpl, ok := s.contrastFinder(s.TemplateID, *stage.MisconceptionID); ok && partnerTmpl != nil {
					s.queueContrastPartner(partnerTmpl, s.Seed+999)
				}
			}
		} else {
			// Second error: reveal solution, mark completed
			stage.Revealed = true
			stage.Status = StageStatusCompleted
			stage.Assistance = append(stage.Assistance, domain.AssistanceSolutionReveal)
			stage.LastFeedback = att.FeedbackMarkdown

			// Check for eligible misconception if not queued yet
			if stage.MisconceptionID != nil && !s.CurrentQuestionIsContrast() && s.ContrastCount == 0 && s.contrastFinder != nil {
				if partnerTmpl, ok := s.contrastFinder(s.TemplateID, *stage.MisconceptionID); ok && partnerTmpl != nil {
					s.queueContrastPartner(partnerTmpl, s.Seed+999)
				}
			}

			if s.checkAllCompleted() {
				s.Completed = true
			} else {
				s.advanceToNextStage()
			}
		}
	}

	stage.DraftAnswer = nil
	if s.store != nil {
		_ = s.store.ClearSessionDraft(context.Background(), s.ID, stage.Instance.ID)
	}

	s.Revision++
	s.UpdatedAt = clock()

	return &CommandResult{
		Success:      true,
		CommandID:    cmd.CommandID,
		SessionState: s.ToPublicView(),
	}, nil
}

func (s *DrillSession) handleRequestHint(cmd SessionCommand, clock func() time.Time) (*CommandResult, error) {
	if s.CurrentStageIndex < 0 || s.CurrentStageIndex >= len(s.Stages) {
		return nil, ErrInvalidStageIndex
	}

	stage := &s.Stages[s.CurrentStageIndex]
	if stage.Status == StageStatusCompleted {
		return &CommandResult{
			Success:      false,
			CommandID:    cmd.CommandID,
			ErrorMessage: "Stage is already completed",
			SessionState: s.ToPublicView(),
		}, nil
	}

	// Record hint assistance if not already recorded
	hasHint := false
	for _, a := range stage.Assistance {
		if a == domain.AssistanceHint {
			hasHint = true
			break
		}
	}
	if !hasHint {
		stage.Assistance = append(stage.Assistance, domain.AssistanceHint)
	}

	// Generate offline hint text
	if stage.ActiveHint == "" {
		if stage.Instance.Kind == domain.StageKindChoice {
			stage.ActiveHint = "Focus on what distinguishes the requested event from alternative possibilities."
		} else {
			stage.ActiveHint = "Substitute $n=4, k=2, p=0.5$ into $P(X=2) = \\binom{4}{2}(0.5)^2(0.5)^2 = 6 \\times (0.5)^4$."
		}
	}

	s.Revision++
	s.UpdatedAt = clock()

	return &CommandResult{
		Success:      true,
		CommandID:    cmd.CommandID,
		SessionState: s.ToPublicView(),
	}, nil
}

// RecordTutorAssistance records tutor assistance on the specified stage if it is active or unresolved.
func (s *DrillSession) RecordTutorAssistance(stageID string, clock func() time.Time) error {
	if clock == nil {
		clock = time.Now
	}

	for i := range s.Stages {
		if s.Stages[i].Instance.ID == stageID {
			if s.Stages[i].Status != StageStatusCompleted {
				hasTutor := false
				for _, a := range s.Stages[i].Assistance {
					if a == domain.AssistanceTutor {
						hasTutor = true
						break
					}
				}
				if !hasTutor {
					s.Stages[i].Assistance = append(s.Stages[i].Assistance, domain.AssistanceTutor)
					s.Revision++
					s.UpdatedAt = clock()
					if s.store != nil {
						_ = s.store.SaveSession(context.Background(), s)
					}
				}
			}
			break
		}
	}
	return nil
}

func (s *DrillSession) handleNavigate(cmd SessionCommand, clock func() time.Time) (*CommandResult, error) {
	if cmd.TargetStageIndex == nil {
		return nil, errors.New("target_stage_index required")
	}
	target := *cmd.TargetStageIndex
	if target < 0 || target >= len(s.Stages) {
		return nil, ErrInvalidStageIndex
	}

	// Check eligibility: learner cannot jump ahead to an unvisited stage beyond current highest active/completed
	maxReachable := 0
	for i, st := range s.Stages {
		if st.Status != StageStatusUnvisited {
			maxReachable = i
		}
	}

	if target > maxReachable {
		return &CommandResult{
			Success:      false,
			CommandID:    cmd.CommandID,
			ErrorMessage: "Cannot jump ahead to an unvisited stage",
			SessionState: s.ToPublicView(),
		}, ErrInvalidStageIndex
	}

	s.CurrentStageIndex = target
	s.Revision++
	s.UpdatedAt = clock()

	return &CommandResult{
		Success:      true,
		CommandID:    cmd.CommandID,
		SessionState: s.ToPublicView(),
	}, nil
}

func (s *DrillSession) handleReset(cmd SessionCommand, clock func() time.Time) (*CommandResult, error) {
	for i := range s.Stages {
		status := StageStatusUnvisited
		if i == 0 {
			status = StageStatusActive
		}
		s.Stages[i].Status = status
		s.Stages[i].Attempts = nil
		s.Stages[i].Assistance = nil
		s.Stages[i].FirstTryCorrect = false
		s.Stages[i].SolvedOnRetry = false
		s.Stages[i].Revealed = false
		s.Stages[i].ActiveHint = ""
		s.Stages[i].LastFeedback = ""
		s.Stages[i].MisconceptionID = nil
		s.Stages[i].InvalidInputNotice = ""
		s.Stages[i].DraftAnswer = nil
		if s.store != nil {
			_ = s.store.ClearSessionDraft(context.Background(), s.ID, s.Stages[i].Instance.ID)
		}
	}
	s.CurrentStageIndex = 0
	s.Completed = false
	s.Revision++
	s.UpdatedAt = clock()

	return &CommandResult{
		Success:      true,
		CommandID:    cmd.CommandID,
		SessionState: s.ToPublicView(),
	}, nil
}

func (s *DrillSession) handleSaveDraft(cmd SessionCommand, clock func() time.Time) (*CommandResult, error) {
	targetIdx := s.CurrentStageIndex
	if cmd.StageID != "" {
		for i := range s.Stages {
			if s.Stages[i].Instance.ID == cmd.StageID {
				targetIdx = i
				break
			}
		}
	}
	if targetIdx < 0 || targetIdx >= len(s.Stages) {
		return nil, ErrInvalidStageIndex
	}
	stage := &s.Stages[targetIdx]
	if stage.Status == StageStatusCompleted {
		return &CommandResult{
			Success:      false,
			CommandID:    cmd.CommandID,
			ErrorMessage: "Cannot save draft for completed stage",
			SessionState: s.ToPublicView(),
		}, nil
	}

	stage.DraftAnswer = cmd.Answer
	if s.store != nil && cmd.Answer != nil {
		_ = s.store.SaveDraftAnswer(context.Background(), s.ID, stage.Instance.ID, *cmd.Answer)
	}

	return &CommandResult{
		Success:      true,
		CommandID:    cmd.CommandID,
		SessionState: s.ToPublicView(),
	}, nil
}

func (s *DrillSession) handleClearDraft(cmd SessionCommand, clock func() time.Time) (*CommandResult, error) {
	targetIdx := s.CurrentStageIndex
	if cmd.StageID != "" {
		for i := range s.Stages {
			if s.Stages[i].Instance.ID == cmd.StageID {
				targetIdx = i
				break
			}
		}
	}
	if targetIdx < 0 || targetIdx >= len(s.Stages) {
		return nil, ErrInvalidStageIndex
	}
	stage := &s.Stages[targetIdx]
	stage.DraftAnswer = nil
	if s.store != nil {
		_ = s.store.ClearSessionDraft(context.Background(), s.ID, stage.Instance.ID)
	}

	return &CommandResult{
		Success:      true,
		CommandID:    cmd.CommandID,
		SessionState: s.ToPublicView(),
	}, nil
}

func (s *DrillSession) handleNavigateQuestion(cmd SessionCommand, clock func() time.Time) (*CommandResult, error) {
	if cmd.TargetQuestionIndex == nil {
		return nil, errors.New("target_question_index is required")
	}
	target := *cmd.TargetQuestionIndex
	if target < 0 || target >= len(s.Questions) {
		return &CommandResult{
			Success:      false,
			CommandID:    cmd.CommandID,
			ErrorMessage: fmt.Sprintf("invalid question index %d", target),
			SessionState: s.ToPublicView(),
		}, errors.New("invalid question index")
	}

	if target == s.CurrentQuestionIndex {
		return &CommandResult{
			Success:      true,
			CommandID:    cmd.CommandID,
			SessionState: s.ToPublicView(),
		}, nil
	}

	// 1. Sync current question
	s.syncCurrentQuestion()

	// 2. Switch to target question
	s.CurrentQuestionIndex = target
	qState := &s.Questions[target]
	s.TemplateID = qState.TemplateID
	s.TemplateVersion = qState.TemplateVersion
	s.QuestionInstance = qState.QuestionInstance
	s.Stages = qState.Stages
	s.CurrentStageIndex = qState.CurrentStageIndex

	// Ensure the first stage of the target question is active if unvisited
	if len(s.Stages) > 0 && s.Stages[0].Status == StageStatusUnvisited {
		s.Stages[0].Status = StageStatusActive
	}

	s.Revision++
	s.UpdatedAt = clock()

	return &CommandResult{
		Success:      true,
		CommandID:    cmd.CommandID,
		SessionState: s.ToPublicView(),
	}, nil
}

func (s *DrillSession) SyncCurrentQuestion() {
	s.syncCurrentQuestion()
}

func (s *DrillSession) syncCurrentQuestion() {
	if len(s.Questions) > 0 && s.CurrentQuestionIndex >= 0 && s.CurrentQuestionIndex < len(s.Questions) {
		s.Questions[s.CurrentQuestionIndex].Stages = s.Stages
		s.Questions[s.CurrentQuestionIndex].CurrentStageIndex = s.CurrentStageIndex
		s.Questions[s.CurrentQuestionIndex].Completed = s.checkCurrentQuestionCompleted()
	}

	// Overall session completed if all questions are completed
	if len(s.Questions) > 0 {
		allDone := true
		for _, q := range s.Questions {
			if !q.Completed {
				allDone = false
				break
			}
		}
		s.Completed = allDone
	} else {
		s.Completed = s.checkCurrentQuestionCompleted()
	}
}

func (s *DrillSession) checkCurrentQuestionCompleted() bool {
	if len(s.Stages) == 0 {
		return false
	}
	for _, st := range s.Stages {
		if st.Status != StageStatusCompleted {
			return false
		}
	}
	return true
}

func (s *DrillSession) checkAllCompleted() bool {
	return s.checkCurrentQuestionCompleted()
}

func (s *DrillSession) advanceToNextStage() {
	for i := s.CurrentStageIndex + 1; i < len(s.Stages); i++ {
		if s.Stages[i].Status == StageStatusUnvisited {
			s.Stages[i].Status = StageStatusActive
			s.CurrentStageIndex = i
			return
		}
		if s.Stages[i].Status != StageStatusCompleted {
			s.CurrentStageIndex = i
			return
		}
	}
}

// ToPublicView constructs the sanitized projection for the client.
func (s *DrillSession) ToPublicView() PublicSessionView {
	stages := make([]PublicStageView, len(s.Stages))
	for i, st := range s.Stages {
		pv := PublicStageView{
			ID:                 st.Instance.ID,
			Number:             i + 1,
			Label:              StageLabel(st.Instance.ID, i+1),
			Kind:               st.Instance.Kind,
			PromptMarkdown:     st.Instance.PromptMarkdown,
			Status:             st.Status,
			AttemptCount:       len(st.Attempts),
			MaxAttempts:        2,
			SolvedOnRetry:      st.SolvedOnRetry,
			Revealed:           st.Revealed,
			ActiveHint:         st.ActiveHint,
			LastFeedback:       st.LastFeedback,
			Attempts:           st.Attempts,
			InvalidInputNotice: st.InvalidInputNotice,
			DraftAnswer:        st.DraftAnswer,
		}

		if st.FirstTryCorrect || st.SolvedOnRetry {
			corr := true
			pv.IsCorrect = &corr
		} else if st.Revealed {
			corr := false
			pv.IsCorrect = &corr
		}

		// Options: public projection strips misconception IDs and hints before stage completion
		if st.Instance.Kind == domain.StageKindChoice {
			pv.Options = make([]PublicOptionView, len(st.Instance.Options))
			for j, opt := range st.Instance.Options {
				pv.Options[j] = PublicOptionView{
					ID:           opt.ID,
					TextMarkdown: opt.TextMarkdown,
				}
			}
		}

		if st.Instance.Kind == domain.StageKindNumeric && st.Instance.NumericPolicy != nil {
			forms := make([]string, len(st.Instance.NumericPolicy.AllowedForms))
			for j, f := range st.Instance.NumericPolicy.AllowedForms {
				forms[j] = string(f)
			}
			pv.NumericPolicy = &PublicNumericPolicyView{
				Version:         st.Instance.NumericPolicy.Version,
				AllowedForms:    forms,
				DisplayDecimals: st.Instance.NumericPolicy.DisplayDecimals,
			}
		}

		// If stage is completed, safely expose explanation and expected answer
		if st.Status == StageStatusCompleted {
			pv.ExplanationMarkdown = st.Instance.ExplanationMarkdown
			pv.ExpectedAnswer = &st.Instance.ExpectedAnswer
		}

		stages[i] = pv
	}

	totalQ := len(s.Questions)
	if totalQ == 0 {
		totalQ = 1
	}

	qInfos := make([]PublicQuestionInfo, totalQ)
	if len(s.Questions) > 0 {
		for i, q := range s.Questions {
			status := "pending"
			if q.Completed {
				status = "completed"
			} else if i == s.CurrentQuestionIndex {
				status = "in_progress"
			} else if q.CurrentStageIndex > 0 || (len(q.Stages) > 0 && q.Stages[0].Status == StageStatusCompleted) {
				status = "in_progress"
			}
			scaff := q.ScaffoldLevel
			if scaff == "" {
				scaff = "full"
			}
			qInfos[i] = PublicQuestionInfo{
				Index:         i,
				Title:         q.QuestionInstance.Title,
				Status:        status,
				ScaffoldLevel: scaff,
				IsContrast:    q.IsContrast,
			}
		}
	} else {
		status := "in_progress"
		if s.Completed {
			status = "completed"
		}
		qInfos[0] = PublicQuestionInfo{
			Index:         0,
			Title:         s.QuestionInstance.Title,
			Status:        status,
			ScaffoldLevel: "full",
			IsContrast:    false,
		}
	}

	allCompleted := s.Completed
	if len(s.Questions) > 0 {
		allCompleted = true
		for _, q := range s.Questions {
			if !q.Completed {
				allCompleted = false
				break
			}
		}
	}

	assumptions := s.QuestionInstance.Assumptions
	if len(assumptions) == 0 {
		assumptions = []string{"Standard assumptions apply"}
	}

	activeScaffold := "full"
	activeContrast := false
	if len(s.Questions) > 0 && s.CurrentQuestionIndex >= 0 && s.CurrentQuestionIndex < len(s.Questions) {
		if s.Questions[s.CurrentQuestionIndex].ScaffoldLevel != "" {
			activeScaffold = s.Questions[s.CurrentQuestionIndex].ScaffoldLevel
		}
		activeContrast = s.Questions[s.CurrentQuestionIndex].IsContrast
	}

	view := PublicSessionView{
		ID:                   s.ID,
		TemplateID:           s.TemplateID,
		Revision:             s.Revision,
		Title:                s.QuestionInstance.Title,
		ScenarioMarkdown:     s.QuestionInstance.ScenarioMarkdown,
		Parameters:           s.QuestionInstance.Parameters,
		Assumptions:          assumptions,
		CurrentStageIndex:    s.CurrentStageIndex,
		Completed:            s.checkCurrentQuestionCompleted(),
		Stages:               stages,
		UpdatedAt:            s.UpdatedAt,
		CurrentQuestionIndex: s.CurrentQuestionIndex,
		TotalQuestions:       totalQ,
		Questions:            qInfos,
		AllCompleted:         allCompleted,
		ScaffoldLevel:        activeScaffold,
		IsContrast:           activeContrast,
	}

	if s.checkCurrentQuestionCompleted() {
		if recap, err := BuildRecap(s); err == nil {
			view.Recap = recap
		}
	}

	return view
}

// queueContrastPartner instantiates a contrast partner question with full guidance and contrast assistance,
// inserting it as the next problem in the session without allowing unbounded session chaining.
func (s *DrillSession) queueContrastPartner(partnerTmpl *domain.QuestionTemplate, seed int64) {
	if partnerTmpl == nil {
		return
	}

	qSeed := seed
	if qSeed == 0 {
		qSeed = 999
	}

	qInst := domain.QuestionInstance{
		ID:               fmt.Sprintf("inst_%s_contrast_%d", partnerTmpl.ID, qSeed),
		TemplateID:       partnerTmpl.ID,
		TemplateVersion:  partnerTmpl.Version,
		Seed:             qSeed,
		Parameters:       partnerTmpl.Parameters,
		Title:            "[Contrast] " + partnerTmpl.Title,
		ScenarioMarkdown: partnerTmpl.ScenarioMarkdown,
		Assumptions:      partnerTmpl.Assumptions,
		SettingGroup:     partnerTmpl.SettingGroup,
		Stages:           make([]domain.StageInstance, len(partnerTmpl.Stages)),
	}

	stages := make([]StageState, len(partnerTmpl.Stages))
	for i, st := range partnerTmpl.Stages {
		stageInst := domain.StageInstance{
			ID:                  st.ID,
			Kind:                st.Kind,
			PromptMarkdown:      st.PromptMarkdown,
			Options:             make([]domain.Option, len(st.Options)),
			ExpectedAnswer:      st.ExpectedAnswer,
			EvidenceConceptIDs:  st.EvidenceConceptIDs,
			ExplanationMarkdown: st.ExplanationMarkdown,
			NumericPolicy:       st.NumericPolicy,
		}
		copy(stageInst.Options, st.Options)
		qInst.Stages[i] = stageInst

		status := StageStatusUnvisited
		if i == 0 {
			status = StageStatusActive
		}

		stages[i] = StageState{
			Instance:   stageInst,
			Status:     status,
			Attempts:   make([]domain.StageAttempt, 0, 2),
			Assistance: []domain.AssistanceType{domain.AssistanceGuidedContrast},
		}
	}

	contrastQ := QuestionState{
		TemplateID:        partnerTmpl.ID,
		TemplateVersion:   partnerTmpl.Version,
		QuestionInstance:  qInst,
		Stages:            stages,
		CurrentStageIndex: 0,
		Completed:         false,
		ScaffoldLevel:     "full",
		IsContrast:        true,
		ContrastPartnerID: partnerTmpl.ID,
	}

	nextIdx := s.CurrentQuestionIndex + 1
	if nextIdx < len(s.Questions) {
		// If the next question has not been started, replace it to preserve session bounds
		if len(s.Questions[nextIdx].Stages) > 0 && len(s.Questions[nextIdx].Stages[0].Attempts) == 0 {
			contrastQ.Index = nextIdx
			s.Questions[nextIdx] = contrastQ
		} else {
			// Otherwise insert right after current question
			contrastQ.Index = nextIdx
			newQuestions := make([]QuestionState, 0, len(s.Questions)+1)
			newQuestions = append(newQuestions, s.Questions[:nextIdx]...)
			newQuestions = append(newQuestions, contrastQ)
			newQuestions = append(newQuestions, s.Questions[nextIdx:]...)
			for idx := range newQuestions {
				newQuestions[idx].Index = idx
			}
			s.Questions = newQuestions
		}
	} else {
		contrastQ.Index = nextIdx
		s.Questions = append(s.Questions, contrastQ)
	}

	s.ContrastCount++
}

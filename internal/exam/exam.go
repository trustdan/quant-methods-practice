package exam

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/trustdan/quant-methods-practice/internal/domain"
	"github.com/trustdan/quant-methods-practice/internal/mathengine"
)

var (
	ErrConflict      = errors.New("exam state modified; reload before retrying")
	ErrInvalidInput  = errors.New("invalid exam input")
	ErrExamFinalized = errors.New("exam is already finalized")
	safeID           = regexp.MustCompile(`^[a-zA-Z0-9_\-\.]{1,80}$`)
)

func invalidInput(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalidInput, fmt.Sprintf(format, args...))
}

// New creates an immutable snapshot-based exam session.
// Questions and option ordering are fixed and detached from subsequent bank changes.
func New(id string, cfg ExamConfig, tmpls []*domain.QuestionTemplate, now time.Time) (*ExamRecord, error) {
	if !safeID.MatchString(id) {
		return nil, invalidInput("exam id must be a valid identifier (1-80 chars)")
	}
	if len(tmpls) == 0 {
		return nil, errors.New("at least one approved question template is required")
	}

	for _, t := range tmpls {
		if t.Status != domain.StatusActive || t.Approval == nil {
			return nil, fmt.Errorf("template %q is not an approved active question", t.ID)
		}
	}

	qCount := cfg.QuestionCount
	if qCount <= 0 {
		qCount = len(tmpls)
	}
	if qCount > len(tmpls) {
		qCount = len(tmpls)
	}

	// Filter by module if requested
	selectedTmpls := make([]*domain.QuestionTemplate, 0, len(tmpls))
	if len(cfg.ModuleIDs) > 0 {
		modMap := make(map[string]bool)
		for _, m := range cfg.ModuleIDs {
			modMap[m] = true
		}
		for _, t := range tmpls {
			if modMap[t.ModuleID] {
				selectedTmpls = append(selectedTmpls, t)
			}
		}
	}
	if len(selectedTmpls) == 0 {
		selectedTmpls = tmpls
	}

	// Shuffle templates deterministically using explicit seed or clock-based seed
	effSeed := cfg.Seed
	if effSeed == 0 {
		effSeed = now.UnixNano()
	}

	available := make([]*domain.QuestionTemplate, len(selectedTmpls))
	copy(available, selectedTmpls)
	r := rand.New(rand.NewSource(effSeed))
	r.Shuffle(len(available), func(i, j int) {
		available[i], available[j] = available[j], available[i]
	})

	chosen := available
	if len(chosen) > qCount {
		chosen = chosen[:qCount]
	}

	questions := make([]domain.QuestionInstance, len(chosen))
	for idx, tmpl := range chosen {
		qSeed := effSeed + int64((idx+1)*1000)
		rng := rand.New(rand.NewSource(qSeed))

		stages := make([]domain.StageInstance, len(tmpl.Stages))
		for sIdx, st := range tmpl.Stages {
			opts := make([]domain.Option, len(st.Options))
			copy(opts, st.Options)

			// Shuffle options deterministically
			if rng != nil && len(opts) > 1 {
				rng.Shuffle(len(opts), func(a, b int) {
					opts[a], opts[b] = opts[b], opts[a]
				})
			}

			stages[sIdx] = domain.StageInstance{
				ID:                  st.ID,
				Kind:                st.Kind,
				PromptMarkdown:      st.PromptMarkdown,
				Options:             opts,
				ExpectedAnswer:      st.ExpectedAnswer,
				EvidenceConceptIDs:  st.EvidenceConceptIDs,
				ExplanationMarkdown: st.ExplanationMarkdown,
				NumericPolicy:       st.NumericPolicy,
			}
		}

		questions[idx] = domain.QuestionInstance{
			ID:               fmt.Sprintf("exam_inst_%s_%d", tmpl.ID, qSeed),
			TemplateID:       tmpl.ID,
			TemplateVersion:  tmpl.Version,
			Seed:             qSeed,
			Parameters:       tmpl.Parameters,
			Title:            tmpl.Title,
			ScenarioMarkdown: tmpl.ScenarioMarkdown,
			Assumptions:      tmpl.Assumptions,
			SettingGroup:     tmpl.SettingGroup,
			Stages:           stages,
		}
	}

	title := cfg.Title
	if title == "" {
		if cfg.DurationMinutes > 0 {
			title = fmt.Sprintf("Timed Assessment (%d min)", cfg.DurationMinutes)
		} else {
			title = "Untimed Assessment"
		}
	}

	rec := &ExamRecord{
		ID:                   id,
		Title:                title,
		Revision:             1,
		Status:               StatusActive,
		DurationMinutes:      cfg.DurationMinutes,
		AllowedAid:           cfg.AllowedAid,
		Seed:                 cfg.Seed,
		CreatedAt:            now.UTC(),
		UpdatedAt:            now.UTC(),
		StartedAt:            now.UTC(),
		CurrentQuestionIndex: 0,
		Questions:            questions,
		Responses:            make(map[string]domain.SubmittedAnswer),
		BankTotalQuestions:   len(tmpls),
	}

	if cfg.DurationMinutes > 0 {
		deadline := now.UTC().Add(time.Duration(cfg.DurationMinutes) * time.Minute)
		rec.DeadlineAt = &deadline
	}

	return clone(rec)
}

func clone(r *ExamRecord) (*ExamRecord, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	var copy ExamRecord
	err = json.Unmarshal(b, &copy)
	return &copy, err
}

// Apply executes an exam command atomically against the recorded state.
func Apply(old *ExamRecord, cmd ExamCommand, now time.Time) (*ExamRecord, error) {
	if !safeID.MatchString(cmd.CommandID) {
		return nil, invalidInput("command_id must be a safe identifier of 1-80 characters")
	}
	if cmd.ExpectedRevision != old.Revision {
		return nil, ErrConflict
	}

	r, err := clone(old)
	if err != nil {
		return nil, err
	}

	nowUTC := now.UTC()

	// Check if already finalized
	if r.Status != StatusActive {
		return nil, fmt.Errorf("%w (current status: %s)", ErrExamFinalized, r.Status)
	}

	// Check wall-clock deadline expiration
	if r.DeadlineAt != nil && nowUTC.After(*r.DeadlineAt) {
		r.Status = StatusExpired
		r.CompletedAt = r.DeadlineAt
		r.Report = EvaluateExam(r, r.DeadlineAt)
		r.Revision++
		r.UpdatedAt = nowUTC
		return r, nil
	}

	switch cmd.Type {
	case "save_answer":
		if cmd.StageID == "" {
			return nil, invalidInput("stage_id is required to save an answer")
		}

		qIdx := r.CurrentQuestionIndex
		if cmd.QuestionIndex != nil {
			qIdx = *cmd.QuestionIndex
		}
		if qIdx < 0 || qIdx >= len(r.Questions) {
			return nil, invalidInput("question index %d out of bounds", qIdx)
		}

		targetQ := r.Questions[qIdx]
		var targetStage *domain.StageInstance
		for i := range targetQ.Stages {
			if targetQ.Stages[i].ID == cmd.StageID {
				targetStage = &targetQ.Stages[i]
				break
			}
		}
		if targetStage == nil {
			return nil, invalidInput("stage %q not found in question %d", cmd.StageID, qIdx)
		}

		respKey := fmt.Sprintf("%s:%s", targetQ.ID, targetStage.ID)
		if cmd.Answer == nil {
			delete(r.Responses, respKey)
		} else {
			ans := *cmd.Answer
			if len(ans.NumericRaw) > 200 || len(ans.OptionID) > 100 {
				return nil, invalidInput("answer string too long")
			}
			if ans.Kind != targetStage.Kind {
				return nil, invalidInput("answer kind mismatch for stage %q", targetStage.ID)
			}
			r.Responses[respKey] = ans
		}

	case "navigate":
		if cmd.QuestionIndex == nil {
			return nil, invalidInput("question_index required for navigate")
		}
		target := *cmd.QuestionIndex
		if target < 0 || target >= len(r.Questions) {
			return nil, invalidInput("target question_index %d out of bounds", target)
		}
		r.CurrentQuestionIndex = target

	case "submit_exam":
		r.Status = StatusCompleted
		r.CompletedAt = &nowUTC
		r.Report = EvaluateExam(r, &nowUTC)

	case "abandon_exam":
		r.Status = StatusAbandoned
		r.CompletedAt = &nowUTC
		r.Report = EvaluateExam(r, &nowUTC)

	default:
		return nil, invalidInput("unknown exam command type %q", cmd.Type)
	}

	r.Revision++
	r.UpdatedAt = nowUTC
	return r, nil
}

// EvaluateExam derives deterministic grades, scores, and feedback for all questions.
func EvaluateExam(r *ExamRecord, atTime *time.Time) *ExamReport {
	timeRef := r.StartedAt
	if atTime != nil {
		timeRef = *atTime
	}
	timeSpent := int(timeRef.Sub(r.StartedAt).Seconds())
	if timeSpent < 0 {
		timeSpent = 0
	}

	var (
		totalStages   int
		correctStages int
		qReports      = make([]ExamQuestionReport, len(r.Questions))
	)

	for qIdx, q := range r.Questions {
		stageReports := make([]ExamStageReport, len(q.Stages))
		qCorrect := 0

		for sIdx, st := range q.Stages {
			totalStages++
			respKey := fmt.Sprintf("%s:%s", q.ID, st.ID)
			sub, answered := r.Responses[respKey]

			report := ExamStageReport{
				ID:                  st.ID,
				Kind:                st.Kind,
				PromptMarkdown:      st.PromptMarkdown,
				Options:             st.Options,
				ExpectedAnswer:      st.ExpectedAnswer,
				ExplanationMarkdown: st.ExplanationMarkdown,
				NumericPolicy:       st.NumericPolicy,
			}

			if answered {
				report.SubmittedAnswer = &sub
				switch st.Kind {
				case domain.StageKindChoice:
					if sub.OptionID == st.ExpectedAnswer.OptionID {
						report.IsCorrect = true
						report.FeedbackMarkdown = fmt.Sprintf("Correct! %s", st.ExplanationMarkdown)
						qCorrect++
						correctStages++
					} else {
						report.IsCorrect = false
						report.FeedbackMarkdown = fmt.Sprintf("Incorrect. The expected choice is: **%s**.\n\n%s",
							getOptionText(st.Options, st.ExpectedAnswer.OptionID), st.ExplanationMarkdown)
					}

				case domain.StageKindNumeric:
					var allowedForms []domain.NumericForm
					if st.NumericPolicy != nil {
						allowedForms = st.NumericPolicy.AllowedForms
					}
					parsed, parseErr := mathengine.ParseNumericInput(sub.NumericRaw, allowedForms)
					if parseErr != nil {
						report.IsCorrect = false
						report.FeedbackMarkdown = fmt.Sprintf("Invalid numeric input (%s). Expected: **%.6g %s**.",
							parseErr.Error(), *st.ExpectedAnswer.Value, st.ExpectedAnswer.Units)
					} else if st.ExpectedAnswer.Value != nil {
						policy := domain.NumericPolicy{
							Version:           1,
							AbsoluteTolerance: 1e-6,
							RelativeTolerance: 1e-6,
						}
						if st.NumericPolicy != nil {
							policy = *st.NumericPolicy
						}
						grade := mathengine.GradeNumericAnswer(parsed, *st.ExpectedAnswer.Value, policy)
						report.IsCorrect = grade.Correct
						if grade.Correct {
							report.FeedbackMarkdown = fmt.Sprintf("Correct! %s", st.ExplanationMarkdown)
							qCorrect++
							correctStages++
						} else {
							report.FeedbackMarkdown = fmt.Sprintf("Incorrect. Expected: **%.6g %s**.\n\n%s",
								*st.ExpectedAnswer.Value, st.ExpectedAnswer.Units, st.ExplanationMarkdown)
						}
					}
				}
			} else {
				report.IsCorrect = false
				switch st.Kind {
				case domain.StageKindChoice:
					report.FeedbackMarkdown = fmt.Sprintf("Unanswered. Expected choice: **%s**.\n\n%s",
						getOptionText(st.Options, st.ExpectedAnswer.OptionID), st.ExplanationMarkdown)
				case domain.StageKindNumeric:
					if st.ExpectedAnswer.Value != nil {
						report.FeedbackMarkdown = fmt.Sprintf("Unanswered. Expected value: **%.6g %s**.\n\n%s",
							*st.ExpectedAnswer.Value, st.ExpectedAnswer.Units, st.ExplanationMarkdown)
					} else {
						report.FeedbackMarkdown = "Unanswered."
					}
				}
			}

			stageReports[sIdx] = report
		}

		qReports[qIdx] = ExamQuestionReport{
			Index:            qIdx,
			QuestionID:       q.ID,
			TemplateID:       q.TemplateID,
			Title:            q.Title,
			ScenarioMarkdown: q.ScenarioMarkdown,
			Assumptions:      q.Assumptions,
			StageReports:     stageReports,
			CorrectStages:    qCorrect,
			TotalStages:      len(q.Stages),
			IsFullyCorrect:   qCorrect == len(q.Stages) && len(q.Stages) > 0,
		}
	}

	scorePct := 0.0
	if totalStages > 0 {
		scorePct = float64(correctStages) / float64(totalStages) * 100.0
	}

	return &ExamReport{
		TotalQuestions:   len(r.Questions),
		TotalStages:      totalStages,
		CorrectStages:    correctStages,
		ScorePercent:     scorePct,
		TimeSpentSeconds: timeSpent,
		Expired:          r.Status == StatusExpired,
		QuestionReports:  qReports,
	}
}

func getOptionText(options []domain.Option, optionID string) string {
	for _, opt := range options {
		if opt.ID == optionID {
			return opt.TextMarkdown
		}
	}
	return optionID
}

// Public constructs the client projection.
// When active, expected answers, options hint markdown, feedback, and solutions are strictly withheld.
func Public(r *ExamRecord, now time.Time) PublicExamView {
	nowUTC := now.UTC()
	effectiveStatus := r.Status

	// Check deadline expiration dynamically for read projection
	if effectiveStatus == StatusActive && r.DeadlineAt != nil && nowUTC.After(*r.DeadlineAt) {
		effectiveStatus = StatusExpired
	}

	var remainingSec *int
	if effectiveStatus == StatusActive && r.DeadlineAt != nil {
		rem := int(r.DeadlineAt.Sub(nowUTC).Seconds())
		if rem < 0 {
			rem = 0
		}
		remainingSec = &rem
	}

	answeredQuestionsCount := 0
	publicQuestions := make([]PublicQuestionView, len(r.Questions))

	for qIdx, q := range r.Questions {
		qAnswered := true
		publicStages := make([]PublicStageView, len(q.Stages))

		for sIdx, st := range q.Stages {
			respKey := fmt.Sprintf("%s:%s", q.ID, st.ID)
			sub, hasSub := r.Responses[respKey]

			var subAns *domain.SubmittedAnswer
			if hasSub {
				subCopy := sub
				subAns = &subCopy
			} else {
				qAnswered = false
			}

			// Clean options: strip HintMarkdown and MisconceptionID so no hint leaks!
			cleanOpts := make([]PublicOptionView, len(st.Options))
			for oIdx, opt := range st.Options {
				cleanOpts[oIdx] = PublicOptionView{
					ID:           opt.ID,
					TextMarkdown: opt.TextMarkdown,
				}
			}

			var policy *domain.NumericPolicy
			if st.NumericPolicy != nil {
				policy = &domain.NumericPolicy{
					Version:           st.NumericPolicy.Version,
					AbsoluteTolerance: st.NumericPolicy.AbsoluteTolerance,
					RelativeTolerance: st.NumericPolicy.RelativeTolerance,
					AllowedForms:      st.NumericPolicy.AllowedForms,
				}
			}

			publicStages[sIdx] = PublicStageView{
				ID:              st.ID,
				Number:          sIdx + 1,
				Kind:            st.Kind,
				PromptMarkdown:  st.PromptMarkdown,
				Options:         cleanOpts,
				NumericPolicy:   policy,
				SubmittedAnswer: subAns,
				IsAnswered:      hasSub,
			}
		}

		if qAnswered && len(q.Stages) > 0 {
			answeredQuestionsCount++
		}

		publicQuestions[qIdx] = PublicQuestionView{
			Index:            qIdx,
			QuestionID:       q.ID,
			Title:            q.Title,
			ScenarioMarkdown: q.ScenarioMarkdown,
			Assumptions:      q.Assumptions,
			Stages:           publicStages,
			IsAnswered:       qAnswered && len(q.Stages) > 0,
		}
	}

	view := PublicExamView{
		ID:                   r.ID,
		Title:                r.Title,
		Revision:             r.Revision,
		Status:               effectiveStatus,
		DurationMinutes:      r.DurationMinutes,
		AllowedAid:           r.AllowedAid,
		StartedAt:            r.StartedAt,
		DeadlineAt:           r.DeadlineAt,
		CompletedAt:          r.CompletedAt,
		TimeRemainingSeconds: remainingSec,
		CurrentQuestionIndex: r.CurrentQuestionIndex,
		TotalQuestions:       len(r.Questions),
		BankTotalQuestions:   r.BankTotalQuestions,
		AnsweredQuestions:    answeredQuestionsCount,
		Questions:            publicQuestions,
	}

	// Reveal full report ONLY if not active
	if effectiveStatus != StatusActive {
		if r.Report != nil {
			view.Report = r.Report
		} else {
			view.Report = EvaluateExam(r, r.CompletedAt)
		}
	}

	return view
}

// Summary constructs a concise overview of the exam for history tables.
func Summary(r *ExamRecord) ExamSummary {
	var (
		scorePct *float64
		corr     *int
		tot      *int
	)
	if r.Report != nil {
		p := r.Report.ScorePercent
		scorePct = &p
		c := r.Report.CorrectStages
		corr = &c
		t := r.Report.TotalStages
		tot = &t
	}

	return ExamSummary{
		ID:              r.ID,
		Title:           r.Title,
		Status:          r.Status,
		DurationMinutes: r.DurationMinutes,
		QuestionCount:   len(r.Questions),
		CreatedAt:       r.CreatedAt,
		CompletedAt:     r.CompletedAt,
		ScorePercent:    scorePct,
		CorrectStages:   corr,
		TotalStages:     tot,
	}
}

// FormatMarkdownReport generates a clean UTF-8 Markdown evaluation report.
func FormatMarkdownReport(r *ExamRecord) string {
	report := r.Report
	if report == nil {
		report = EvaluateExam(r, r.CompletedAt)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Assessment Report: %s\n\n", r.Title))
	sb.WriteString(fmt.Sprintf("- **Exam ID:** `%s`\n", r.ID))
	sb.WriteString(fmt.Sprintf("- **Status:** `%s`\n", r.Status))
	sb.WriteString(fmt.Sprintf("- **Started:** %s\n", r.StartedAt.Format("2006-01-02 15:04:05 UTC")))
	if r.CompletedAt != nil {
		sb.WriteString(fmt.Sprintf("- **Completed:** %s\n", r.CompletedAt.Format("2006-01-02 15:04:05 UTC")))
	}
	sb.WriteString(fmt.Sprintf("- **Time Spent:** %d min %d sec\n", report.TimeSpentSeconds/60, report.TimeSpentSeconds%60))
	sb.WriteString(fmt.Sprintf("- **Overall Score:** **%.1f%%** (%d / %d stages correct)\n\n",
		report.ScorePercent, report.CorrectStages, report.TotalStages))

	sb.WriteString("## Question Performance\n\n")
	for _, qr := range report.QuestionReports {
		statusMark := "❌"
		if qr.IsFullyCorrect {
			statusMark = "✅"
		}
		sb.WriteString(fmt.Sprintf("### %s Question %d: %s\n\n", statusMark, qr.Index+1, qr.Title))
		sb.WriteString(fmt.Sprintf("**Scenario:** %s\n\n", qr.ScenarioMarkdown))

		for _, sr := range qr.StageReports {
			sMark := "❌"
			if sr.IsCorrect {
				sMark = "✅"
			}
			sb.WriteString(fmt.Sprintf("#### %s Stage: %s\n\n", sMark, sr.ID))
			sb.WriteString(fmt.Sprintf("**Prompt:** %s\n\n", sr.PromptMarkdown))

			if sr.SubmittedAnswer != nil {
				if sr.Kind == domain.StageKindChoice {
					sb.WriteString(fmt.Sprintf("- **Your Answer:** `%s` (%s)\n",
						sr.SubmittedAnswer.OptionID, getOptionText(sr.Options, sr.SubmittedAnswer.OptionID)))
				} else {
					sb.WriteString(fmt.Sprintf("- **Your Answer:** `%s`\n", sr.SubmittedAnswer.NumericRaw))
				}
			} else {
				sb.WriteString("- **Your Answer:** *(no answer submitted)*\n")
			}

			if sr.Kind == domain.StageKindChoice {
				sb.WriteString(fmt.Sprintf("- **Expected Answer:** `%s` (%s)\n",
					sr.ExpectedAnswer.OptionID, getOptionText(sr.Options, sr.ExpectedAnswer.OptionID)))
			} else if sr.ExpectedAnswer.Value != nil {
				sb.WriteString(fmt.Sprintf("- **Expected Answer:** `%.6g %s`\n",
					*sr.ExpectedAnswer.Value, sr.ExpectedAnswer.Units))
			}

			sb.WriteString(fmt.Sprintf("\n> %s\n\n", sr.FeedbackMarkdown))
		}
		sb.WriteString("---\n\n")
	}

	return sb.String()
}

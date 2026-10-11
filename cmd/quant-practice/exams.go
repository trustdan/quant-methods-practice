package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/trustdan/quant-methods-practice/internal/bank"
	"github.com/trustdan/quant-methods-practice/internal/exam"
	"github.com/trustdan/quant-methods-practice/internal/storage"
)

func runExamHistory(store *storage.Store) {
	ctx := context.Background()
	summaries, err := store.ListExams(ctx)
	if err != nil {
		log.Fatalf("failed to list exams: %v", err)
	}

	fmt.Printf("=== Assessment History (%d assessment(s)) ===\n\n", len(summaries))
	if len(summaries) == 0 {
		fmt.Println("No assessments taken yet. Use -exam to start an assessment.")
		return
	}

	fmt.Printf("%-20s %-12s %-12s %-10s %-10s %s\n",
		"EXAM ID", "STATUS", "DURATION", "QUESTIONS", "SCORE", "CREATED AT")
	fmt.Println(strings.Repeat("-", 80))

	for _, s := range summaries {
		durStr := "untimed"
		if s.DurationMinutes > 0 {
			durStr = fmt.Sprintf("%d min", s.DurationMinutes)
		}
		scoreStr := "-"
		if s.ScorePercent != nil {
			scoreStr = fmt.Sprintf("%.1f%%", *s.ScorePercent)
		}
		fmt.Printf("%-20s %-12s %-12s %-10d %-10s %s\n",
			s.ID,
			string(s.Status),
			durStr,
			s.QuestionCount,
			scoreStr,
			s.CreatedAt.Format("2006-01-02 15:04"),
		)
	}
}

func runExamReport(store *storage.Store, examID string) {
	if examID == "" {
		fmt.Fprintf(os.Stderr, "Error: -exam-report requires an exam ID\n")
		os.Exit(1)
	}

	ctx := context.Background()
	rec, err := store.GetExam(ctx, examID)
	if err != nil {
		log.Fatalf("failed to fetch exam %q: %v", examID, err)
	}

	if rec.Status == exam.StatusActive {
		fmt.Printf("Assessment %q is currently in progress. Score report is withheld until completion.\n", examID)
		return
	}

	md := exam.FormatMarkdownReport(rec)
	fmt.Println(md)
}

func runResumeExam(store *storage.Store, examID string) {
	if examID == "" {
		fmt.Fprintf(os.Stderr, "Error: -resume-exam requires an exam ID\n")
		os.Exit(1)
	}

	ctx := context.Background()
	rec, err := store.GetExam(ctx, examID)
	if err != nil {
		log.Fatalf("failed to fetch exam %q: %v", examID, err)
	}

	v := exam.Public(rec, time.Now())
	fmt.Printf("=== Assessment Status: %s ===\n", v.Title)
	fmt.Printf("  ID:                 %s\n", v.ID)
	fmt.Printf("  Status:             %s\n", v.Status)
	fmt.Printf("  Started At:         %s\n", v.StartedAt.Format("2006-01-02 15:04:05 UTC"))
	if v.DeadlineAt != nil {
		fmt.Printf("  Deadline:           %s\n", v.DeadlineAt.Format("2006-01-02 15:04:05 UTC"))
	}
	if v.TimeRemainingSeconds != nil {
		rem := *v.TimeRemainingSeconds
		fmt.Printf("  Time Remaining:     %d min %d sec\n", rem/60, rem%60)
	}
	fmt.Printf("  Progress:           %d / %d questions answered\n", v.AnsweredQuestions, v.TotalQuestions)
	if v.Report != nil {
		fmt.Printf("  Final Score:        %.1f%% (%d / %d stages correct)\n",
			v.Report.ScorePercent, v.Report.CorrectStages, v.Report.TotalStages)
	}
}

func runStartExam(store *storage.Store, b *bank.Bank, qCount, duration int, seed int64, moduleFilter string) {
	if b == nil || b.Count() == 0 {
		log.Fatalf("no approved questions available in curriculum bank")
	}

	ctx := context.Background()
	allTmpls := b.List()

	var modIDs []string
	if moduleFilter != "" {
		for _, m := range strings.Split(moduleFilter, ",") {
			trimmed := strings.TrimSpace(m)
			if trimmed != "" {
				modIDs = append(modIDs, trimmed)
			}
		}
	}

	if qCount <= 0 {
		qCount = 10
	}

	cfg := exam.ExamConfig{
		QuestionCount:   qCount,
		DurationMinutes: duration,
		Seed:            seed,
		ModuleIDs:       modIDs,
	}

	id := storage.GenerateExamID()
	rec, err := exam.New(id, cfg, allTmpls, time.Now())
	if err != nil {
		log.Fatalf("failed to initialize assessment: %v", err)
	}

	if err := store.CreateExam(ctx, rec); err != nil {
		log.Fatalf("failed to store assessment: %v", err)
	}

	fmt.Printf("=== Assessment Initialized: %s ===\n", rec.Title)
	fmt.Printf("  Exam ID:            %s\n", rec.ID)
	fmt.Printf("  Questions:          %d\n", len(rec.Questions))
	if rec.DurationMinutes > 0 {
		fmt.Printf("  Duration:           %d minutes (Deadline: %s)\n",
			rec.DurationMinutes, rec.DeadlineAt.Format("15:04:05 UTC"))
	} else {
		fmt.Println("  Duration:           Untimed")
	}
	fmt.Println("  Aid Policy:         No hints or answers revealed during assessment")
	fmt.Printf("\nOpen the web interface and press Shift+E to take this assessment.\n")
}

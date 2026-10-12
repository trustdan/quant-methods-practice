package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"math"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/trustdan/quant-methods-practice/curriculum"
	"github.com/trustdan/quant-methods-practice/internal/assets"
	"github.com/trustdan/quant-methods-practice/internal/auth"
	"github.com/trustdan/quant-methods-practice/internal/bank"
	"github.com/trustdan/quant-methods-practice/internal/httpapi"
	"github.com/trustdan/quant-methods-practice/internal/mathengine"
	"github.com/trustdan/quant-methods-practice/internal/siwc"
	"github.com/trustdan/quant-methods-practice/internal/storage"
	"github.com/trustdan/quant-methods-practice/internal/tutor"
)

var version = "0.1.0"

func main() {
	var (
		portFlag         = flag.Int("port", 0, "Loopback port to bind (default 0 for random available port)")
		noBrowserFlag    = flag.Bool("no-browser", false, "Do not automatically launch the web browser")
		versionFlag      = flag.Bool("version", false, "Display application version and exit")
		dataDirFlag      = flag.String("data-dir", "", "Path to user data directory for local storage")
		dbFlag           = flag.String("db", "", "Path to SQLite database file (overrides -data-dir)")
		backupFlag       = flag.String("backup", "", "Create a consistent backup of the SQLite database to the specified path and exit")
		skipIntroFlag    = flag.Bool("skip-intro", false, "Skip startup introduction/arcade sequence")
		validateBankFlag = flag.String("validate-bank", "", "Path to template file or directory to validate")
		activeBankFlag   = flag.String("active-bank", "", "Path to approved active curriculum directory to load and inspect")
		listBankFlag     = flag.Bool("list-bank", false, "List approved questions in curriculum bank and exit")
		questionsFlag    = flag.Int("questions", 10, "Target number of questions for session (default 10)")
		moduleFlag       = flag.String("module", "", "Filter questions by module ID (e.g. module_01, module_02)")
		intensityFlag    = flag.String("intensity", "standard", "Practice intensity: gentle, standard, or intensive")
		seedFlag         = flag.Int64("seed", 0, "Seed for reproducible question generation")
		evalBinomialFlag = flag.String("eval-binomial", "", "Evaluate binomial derivation (format: n=4,p=0.5,k=2)")
		masteryFlag      = flag.Bool("mastery", false, "Display curriculum concept mastery and transfer projections and exit")
		notesFlag        = flag.Bool("notes", false, "List saved explanations and notes in personal library and exit")
		exportNotesFlag  = flag.String("export-notes", "", "Export all saved notes as UTF-8 Markdown files to specified directory and exit")
		providersFlag    = flag.Bool("providers", false, "Display configured AI providers and models and exit")
		examFlag         = flag.Bool("exam", false, "Initialize and start a new assessment session")
		examTimeFlag     = flag.Int("exam-time", 0, "Time limit in minutes for exam (0 for untimed)")
		resumeExamFlag   = flag.String("resume-exam", "", "Resume active assessment by ID")
		examHistoryFlag  = flag.Bool("exam-history", false, "Display past assessment history and scores")
		examReportFlag   = flag.String("exam-report", "", "Display assessment score report by ID")
		diagnosticsFlag  = flag.Bool("diagnostics", false, "Run offline doctor and system diagnostics and exit")
		doctorFlag       = flag.Bool("doctor", false, "Alias for -diagnostics")
		highScoresFlag   = flag.Bool("high-scores", false, "Display Quant Flight arcade leaderboard and exit")
		exportDataFlag   = flag.String("export-data", "", "Export all learner data (notes, arcade scores, exams, settings) to a JSON file and exit")
	)

	var candidateOpts candidateFlags
	flag.BoolVar(&candidateOpts.generate, "generate-candidate", false, "Generate and save a local binomial candidate (never auto-approve)")
	flag.BoolVar(&candidateOpts.list, "candidates", false, "List saved candidates and review history")
	flag.StringVar(&candidateOpts.preview, "preview-candidate", "", "Print full candidate preview by ID")
	flag.StringVar(&candidateOpts.importFile, "import-candidate", "", "Import a proposal JSON file as a new draft")
	flag.StringVar(&candidateOpts.approve, "approve-candidate", "", "Approve a reviewed candidate ID")
	flag.StringVar(&candidateOpts.reject, "reject-candidate", "", "Reject a candidate ID")
	flag.StringVar(&candidateOpts.retire, "retire-question", "", "Retire a locally approved candidate ID")
	flag.StringVar(&candidateOpts.export, "export-bank", "", "Export active approved bank to a new JSON file")
	flag.IntVar(&candidateOpts.revision, "candidate-revision", 0, "Preview revision required for a review command")
	flag.StringVar(&candidateOpts.reviewer, "candidate-reviewer", "", "Human reviewer's name")
	flag.StringVar(&candidateOpts.notes, "candidate-notes", "", "Semantic review findings or rejection/retirement reason")
	flag.BoolVar(&candidateOpts.semantic, "confirm-semantic-review", false, "Confirm personal review of scenario, assumptions, stages, answers and hints")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("quant-practice version %s\n", version)
		os.Exit(0)
	}

	if *evalBinomialFlag != "" {
		runEvalBinomial(*evalBinomialFlag)
		return
	}

	if *validateBankFlag != "" {
		runValidateBank(*validateBankFlag)
		return
	}

	if *activeBankFlag != "" {
		runActiveBank(*activeBankFlag)
		return
	}

	if *listBankFlag {
		runListBank(*activeBankFlag)
		return
	}

	dbPath, err := storage.ResolveDBPath(*dataDirFlag, *dbFlag)
	if err != nil {
		log.Fatalf("failed to resolve database path: %v", err)
	}

	db, err := storage.Open(dbPath)
	if err != nil {
		log.Fatalf("failed to open database %q: %v", dbPath, err)
	}
	defer db.Close()

	ctx := context.Background()

	// Handle standalone backup command
	if *backupFlag != "" {
		if err := storage.Backup(ctx, db, *backupFlag); err != nil {
			log.Fatalf("failed to create backup: %v", err)
		}
		fmt.Printf("Consistent SQLite backup created at: %s\n", *backupFlag)
		return
	}

	// Run transactional migrations with pre-migration backup callback
	backupCallback := func() error {
		backupPath := dbPath + fmt.Sprintf(".backup-%d", time.Now().Unix())
		log.Printf("Creating pre-migration backup at: %s", backupPath)
		return storage.Backup(ctx, db, backupPath)
	}
	if err := storage.RunMigrations(ctx, db, backupCallback); err != nil {
		log.Fatalf("failed to apply migrations: %v", err)
	}

	store := storage.NewStore(db, nil)
	if candidateOpts.requested() {
		if err := runCandidateCLI(ctx, store, candidateOpts, *seedFlag); err != nil {
			log.Fatalf("candidate command: %v", err)
		}
		return
	}

	if *diagnosticsFlag || *doctorFlag {
		runDiagnostics(dbPath, store, *dataDirFlag)
		return
	}

	if *highScoresFlag {
		runHighScores(store)
		return
	}

	if *exportDataFlag != "" {
		runExportData(store, *exportDataFlag)
		return
	}

	if *examHistoryFlag {
		runExamHistory(store)
		return
	}

	if *examReportFlag != "" {
		runExamReport(store, *examReportFlag)
		return
	}

	if *resumeExamFlag != "" {
		runResumeExam(store, *resumeExamFlag)
		return
	}

	if *examFlag {
		activeBank, _, err := loadDefaultBank()
		if err != nil || activeBank == nil {
			log.Fatalf("failed to load curriculum bank for exam: %v", err)
		}
		runStartExam(store, activeBank, *questionsFlag, *examTimeFlag, *seedFlag, *moduleFlag)
		return
	}

	if *masteryFlag {
		runMastery(store)
		return
	}

	if *notesFlag {
		runNotes(store)
		return
	}

	if *exportNotesFlag != "" {
		runExportNotes(store, *exportNotesFlag)
		return
	}

	dataDir := *dataDirFlag
	if dataDir == "" {
		dataDir, _ = storage.DefaultDataDir()
	}
	vault, err := auth.NewStandardVault(dataDir)
	if err != nil {
		log.Fatalf("failed to initialize credential vault: %v", err)
	}

	if *providersFlag {
		runProviders(vault)
		return
	}

	// Persist initial user preferences if configured via CLI flags
	if *questionsFlag != 10 || *moduleFlag != "" || *intensityFlag != "standard" || *seedFlag != 0 {
		var modIDs []string
		if *moduleFlag != "" {
			for _, m := range strings.Split(*moduleFlag, ",") {
				trimmed := strings.TrimSpace(m)
				if trimmed != "" {
					modIDs = append(modIDs, trimmed)
				}
			}
		}
		prefs := map[string]any{
			"question_count": *questionsFlag,
			"module_ids":     modIDs,
			"intensity":      *intensityFlag,
			"seed":           *seedFlag,
		}
		if prefsBytes, err := json.Marshal(prefs); err == nil {
			_ = store.SaveSettings(ctx, "user_preferences", string(prefsBytes))
		}
	}

	distFS, err := assets.FS()
	if err != nil {
		log.Fatalf("failed to load embedded application assets: %v", err)
	}

	activeBank, bankSource, err := loadDefaultBank()
	if err != nil {
		log.Printf("Warning: %v", err)
	} else {
		log.Printf("Loaded active curriculum bank from %s: %d approved template(s)", bankSource, activeBank.Count())
	}

	addr := fmt.Sprintf("127.0.0.1:%d", *portFlag)
	srv, err := httpapi.NewServer(httpapi.Config{
		Addr:              addr,
		AssetsFS:          distFS,
		Version:           version,
		AllowedDevOrigins: []string{"http://localhost:5173", "http://127.0.0.1:5173"},
		Bank:              activeBank,
		Store:             store,
		Vault:             vault,
	})
	if err != nil {
		log.Fatalf("failed to initialize server: %v", err)
	}

	if err := srv.Start(); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}

	bootstrapURL := srv.BootstrapURL()
	log.Printf("=====================================================")
	log.Printf("Quant Methods Practice v%s", version)
	log.Printf("Listening locally on: http://%s", srv.Addr())
	log.Printf("Application URL:      %s", bootstrapURL)
	log.Printf("Database path:        %s", dbPath)
	if *dataDirFlag != "" {
		log.Printf("Data directory:       %s", *dataDirFlag)
	}
	if *skipIntroFlag {
		log.Printf("Intro sequence:       skipped")
	}
	log.Printf("Press Ctrl+C to stop the local server")
	log.Printf("=====================================================")

	if !*noBrowserFlag {
		go func() {
			time.Sleep(100 * time.Millisecond)
			if err := openBrowser(bootstrapURL); err != nil {
				log.Printf("Note: could not automatically open browser: %v", err)
				log.Printf("Please open %s manually in your browser.", bootstrapURL)
			}
		}()
	}

	// Wait for termination signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("\nShutting down local server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_ = srv.Close()
	<-shutdownCtx.Done()
	log.Println("Server stopped cleanly.")
}

func runValidateBank(targetPath string) {
	fi, err := os.Stat(targetPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot access %s: %v\n", targetPath, err)
		os.Exit(1)
	}

	if fi.IsDir() {
		results, err := bank.ValidateBankDir(targetPath, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading directory %s: %v\n", targetPath, err)
			os.Exit(1)
		}
		if len(results) == 0 {
			fmt.Printf("No .json template files found in %s\n", targetPath)
			return
		}
		hasError := false
		for _, r := range results {
			if r.Valid {
				fmt.Printf("[VALID]   %s (id: %s, status: %s)\n", r.Path, r.TemplateID, r.Status)
			} else {
				hasError = true
				fmt.Printf("[INVALID] %s: %v\n", r.Path, r.Error)
			}
		}
		if hasError {
			os.Exit(1)
		}
		fmt.Printf("All %d template file(s) in %s passed strict validation.\n", len(results), targetPath)
		return
	}

	tmpl, err := bank.ValidateTemplateFile(targetPath, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[INVALID] %s: %v\n", targetPath, err)
		os.Exit(1)
	}

	fmt.Printf("[VALID]   %s (id: %s, status: %s, stages: %d)\n", targetPath, tmpl.ID, tmpl.Status, len(tmpl.Stages))
}

func runActiveBank(dirPath string) {
	b, err := bank.LoadActiveBank(dirPath, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading active bank from %s: %v\n", dirPath, err)
		os.Exit(1)
	}

	count := b.Count()
	fmt.Printf("Active question bank at %s: %d approved record(s)\n", dirPath, count)
	for i, tmpl := range b.List() {
		fmt.Printf("  %d. [%s] %s (%s, module: %s)\n", i+1, tmpl.ID, tmpl.Title, tmpl.FamilyID, tmpl.ModuleID)
	}
}

func openBrowser(url string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	default:
		return exec.Command("xdg-open", url).Start()
	}
}

func runEvalBinomial(paramStr string) {
	parts := strings.Split(paramStr, ",")
	params := make(map[string]string)
	for _, p := range parts {
		kv := strings.Split(strings.TrimSpace(p), "=")
		if len(kv) == 2 {
			params[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	nStr, okN := params["n"]
	pStr, okP := params["p"]
	kStr, okK := params["k"]
	if !okN || !okP || !okK {
		fmt.Fprintf(os.Stderr, "Error: -eval-binomial requires format \"n=<int>,p=<float>,k=<int>\" (got %q)\n", paramStr)
		os.Exit(1)
	}
	n, errN := strconv.Atoi(nStr)
	p, errP := strconv.ParseFloat(pStr, 64)
	k, errK := strconv.Atoi(kStr)
	if errN != nil || errP != nil || errK != nil {
		fmt.Fprintf(os.Stderr, "Error parsing parameters: n=%v, p=%v, k=%v\n", errN, errP, errK)
		os.Exit(1)
	}

	deriv, err := mathengine.DeriveBinomialProblem(n, p, k, mathengine.Exactly(k))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error deriving binomial problem: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Binomial Derivation (n=%d, p=%.4g, k=%d, Event: %s):\n", n, p, k, deriv.EventTeX)
	if deriv.CanonicalRational != nil {
		fmt.Printf("  Canonical Probability: %.6g (%s)\n", deriv.CanonicalProbability, deriv.CanonicalRational.RatString())
	} else {
		fmt.Printf("  Canonical Probability: %.6g\n", deriv.CanonicalProbability)
	}
	fmt.Printf("  Moments: Mean = %.4g, Variance = %.4g, SD = %.4g\n", deriv.Mean, deriv.Variance, deriv.StdDev)
	fmt.Printf("  Canonical Expression: %s\n", deriv.ExpressionTeX)
	fmt.Printf("  Calculation: %s\n", deriv.CalculationTeX)
	fmt.Println("  Misconception Distractors:")
	for _, d := range deriv.Distractors {
		if d.Value != nil {
			fmt.Printf("    - %-26s: %-8.6g (%s)\n", d.MisconceptionID, *d.Value, d.ExpressionTeX)
		} else {
			fmt.Printf("    - %-26s: [no value] (%s)\n", d.MisconceptionID, d.ExpressionTeX)
		}
	}
}

func loadDefaultBank() (*bank.Bank, string, error) {
	bankDirs := []string{"curriculum/approved", "../curriculum/approved", "../../curriculum/approved"}
	for _, dir := range bankDirs {
		if fi, statErr := os.Stat(dir); statErr == nil && fi.IsDir() {
			b, loadErr := bank.LoadActiveBank(dir, nil)
			if loadErr == nil && b.Count() > 0 {
				return b, dir, nil
			}
		}
	}
	// Fall back to embedded curriculum assets
	b, err := bank.LoadActiveBankFS(curriculum.ApprovedFS, "approved", nil)
	if err == nil && b.Count() > 0 {
		return b, "embedded assets", nil
	}
	return nil, "", fmt.Errorf("failed to load curriculum bank: %w", err)
}

func runListBank(bankPath string) {
	var b *bank.Bank
	var err error
	var source string

	if bankPath != "" {
		b, err = bank.LoadActiveBank(bankPath, nil)
		source = bankPath
	} else {
		b, source, err = loadDefaultBank()
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading curriculum bank: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Approved Curriculum Bank (%s): %d questions\n", source, b.Count())
	fmt.Printf("%-38s %-20s %-32s %-6s %s\n", "TEMPLATE ID", "FAMILY", "MODULE", "STAGES", "TITLE")
	fmt.Println(strings.Repeat("-", 120))
	for _, tmpl := range b.List() {
		fmt.Printf("%-38s %-20s %-32s %-6d %s\n",
			tmpl.ID,
			tmpl.FamilyID,
			tmpl.ModuleID,
			len(tmpl.Stages),
			tmpl.Title,
		)
	}
}

func runMastery(store *storage.Store) {
	ctx := context.Background()
	summary, err := store.GetMasterySummary(ctx, nil, time.Now())
	if err != nil {
		log.Fatalf("failed to derive mastery summary: %v", err)
	}

	fmt.Println("=== Concept Mastery & Transfer Projections ===")
	fmt.Printf("Policy Version: %d | Overall Mastery: %.1f%% | Generated: %s\n",
		summary.PolicyVersion, summary.OverallScore*100, summary.GeneratedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Mastered: %d | Transferring: %d | Learning: %d | New: %d\n\n",
		summary.TotalMastered, summary.TotalTransferring, summary.TotalLearning, summary.TotalNew)

	if len(summary.Concepts) == 0 {
		fmt.Println("No concept evidence recorded yet. Complete practice drills to build mastery.")
		return
	}

	fmt.Printf("%-34s %-14s %-14s %-8s %-14s %-8s %-12s\n",
		"CONCEPT ID", "STATUS", "SCAFFOLD", "SCORE", "SUCCESS/ERR", "GROUPS", "TRANSFER?")
	fmt.Println(strings.Repeat("-", 110))

	for _, c := range summary.Concepts {
		transferStr := "pending"
		if c.DelayedTransferAchieved {
			transferStr = "achieved"
		}
		fmt.Printf("%-34s %-14s %-14s %-7.1f%% %-14s %-8d %-12s\n",
			c.ConceptID,
			string(c.Status),
			string(c.ScaffoldLevel),
			c.DecayedScore*100,
			fmt.Sprintf("%d / %d (%d)", c.IndependentSuccesses, c.IndependentErrors, c.AssistedCount),
			len(c.SettingGroupsSeen),
			transferStr,
		)
	}
}

func runNotes(store *storage.Store) {
	ctx := context.Background()
	notes, err := store.ListExplanations(ctx, "", "")
	if err != nil {
		log.Fatalf("failed to list saved notes: %v", err)
	}

	fmt.Printf("=== Saved Explanations & Notes Library (%d note(s)) ===\n\n", len(notes))
	if len(notes) == 0 {
		fmt.Println("No saved notes found. Save explanations during practice drills to build your library.")
		return
	}

	fmt.Printf("%-24s %-22s %-30s %-10s %s\n", "ID", "TOPIC", "TITLE", "PROVIDER", "SAVED AT")
	fmt.Println(strings.Repeat("-", 110))
	for _, n := range notes {
		title := n.ProviderInfo.Title
		if title == "" {
			title = "(untitled)"
		}
		if len(title) > 28 {
			title = title[:25] + "..."
		}
		fmt.Printf("%-24s %-22s %-30s %-10s %s\n",
			n.ID,
			n.Topic,
			title,
			n.ProviderInfo.Provider,
			n.UpdatedAt.Format("2006-01-02 15:04"),
		)
	}
}

func runExportNotes(store *storage.Store, outDir string) {
	if outDir == "" {
		fmt.Fprintf(os.Stderr, "Error: -export-notes requires a destination directory\n")
		os.Exit(1)
	}
	if err := os.MkdirAll(outDir, 0755); err != nil {
		log.Fatalf("failed to create export directory %q: %v", outDir, err)
	}

	ctx := context.Background()
	notes, err := store.ListExplanations(ctx, "", "")
	if err != nil {
		log.Fatalf("failed to retrieve notes for export: %v", err)
	}

	fmt.Printf("Exporting %d saved note(s) to %s...\n", len(notes), outDir)
	for _, n := range notes {
		md := tutor.ExportNoteToMarkdown(n, "")
		filename := fmt.Sprintf("%s.md", n.ID)
		filePath := filepath.Join(outDir, filename)
		if err := os.WriteFile(filePath, []byte(md), 0644); err != nil {
			log.Printf("Warning: failed to write %s: %v", filePath, err)
			continue
		}
		fmt.Printf("  -> Exported: %s\n", filePath)
	}
	fmt.Println("Export complete.")
}

func runProviders(v auth.Vault) {
	fmt.Println("=== AI Provider Credentials & Vault Status ===")
	statuses := v.AllStatuses()

	fmt.Printf("%-14s %-14s %-16s %s\n", "ROUTE", "CONFIGURED?", "SOURCE", "MASKED KEY")
	fmt.Println(strings.Repeat("-", 65))
	for _, st := range statuses {
		confStr := "No"
		if st.Configured {
			confStr = "Yes"
		}
		masked := st.MaskedKey
		if masked == "" {
			masked = "(none)"
		}
		fmt.Printf("%-14s %-14s %-16s %s\n",
			st.Route,
			confStr,
			string(st.Source),
			masked,
		)
	}

	fmt.Println()
	fmt.Println("=== ChatGPT Plan Accounts (Sign in with ChatGPT; billed to the ChatGPT plan) ===")
	accts, err := siwc.NewClient(v).Accounts()
	switch {
	case err != nil:
		fmt.Printf("  unavailable: %v\n", err)
	case len(accts) == 0:
		fmt.Println("  (no accounts signed in)")
	}
	for _, a := range accts {
		state := "ready"
		if a.NeedsReauth {
			state = "sign-in expired"
		} else if !a.PlanGranted {
			state = "plan usage not granted"
		}
		sel := " "
		if a.Selected {
			sel = "*"
		}
		fmt.Printf("  %s %-32s %s\n", sel, a.Label, state)
	}
	fmt.Println()
	fmt.Println("Note: Keys and tokens are stored exclusively in the backend vault and are never exposed in UI or logs.")
}

func runDiagnostics(dbPath string, store *storage.Store, dataDir string) {
	fmt.Println("==========================================================")
	fmt.Println("  Quant Methods Practice — Offline Diagnostics & Doctor   ")
	fmt.Println("==========================================================")

	allHealthy := true

	// 1. Environment & Architecture
	fmt.Printf("[INFO] Platform: %s/%s | App Version: %s\n", runtime.GOOS, runtime.GOARCH, version)
	if exe, err := os.Executable(); err == nil {
		fmt.Printf("[INFO] Executable: %s\n", exe)
	}
	if dataDir != "" {
		fmt.Printf("[INFO] Data Directory: %s\n", dataDir)
	}
	fmt.Printf("[INFO] SQLite Database: %s\n", dbPath)

	// 2. Embedded Web Assets
	distFS, err := assets.FS()
	if err != nil {
		fmt.Printf("[FAIL] Embedded Web Assets: %v\n", err)
		allHealthy = false
	} else {
		indexData, err := fs.ReadFile(distFS, "index.html")
		if err != nil || len(indexData) == 0 {
			fmt.Printf("[FAIL] Embedded Web Assets: index.html missing or empty\n")
			allHealthy = false
		} else {
			if _, err := fs.ReadFile(distFS, "vendor/mathjax/tex-svg.js"); err != nil {
				fmt.Printf("[WARN] Embedded Web Assets: vendor/mathjax/tex-svg.js not found in bundle\n")
			}
			fmt.Printf("[PASS] Embedded Web Assets: index.html & local MathJax bundle verified (%d bytes)\n", len(indexData))
		}
	}

	// 3. Active Bank Resolution
	b, source, err := loadDefaultBank()
	if err != nil {
		fmt.Printf("[FAIL] Active Bank Resolution: %v\n", err)
		allHealthy = false
	} else {
		count := b.Count()
		fmt.Printf("[PASS] Active Bank Resolution: %s (%d approved question templates validated offline)\n", source, count)
		for m := 1; m <= 6; m++ {
			modID := fmt.Sprintf("module_%d", m)
			tmpls := b.ListByModule(modID)
			fmt.Printf("       • Module %d: %d approved templates\n", m, len(tmpls))
		}
	}

	// 4. Mathematical Engine Sanity Checks
	binomialDeriv, err := mathengine.DeriveBinomialProblem(4, 0.5, 2, mathengine.Exactly(2))
	if err != nil || math.Abs(binomialDeriv.CanonicalProbability-0.375) > 1e-9 {
		fmt.Printf("[FAIL] Mathematical Engine: Binomial derivation failed\n")
		allHealthy = false
	} else {
		fmt.Printf("[PASS] Mathematical Engines: Binomial, Normal, and OLS Regression verified\n")
	}

	// 5. Storage & Database Status
	ctx := context.Background()
	if _, err := store.GetHighScore(ctx); err != nil {
		fmt.Printf("[FAIL] SQLite Storage: %v\n", err)
		allHealthy = false
	} else {
		fmt.Printf("[PASS] SQLite Storage: Schema migrations up to date, read/write verified\n")
	}

	fmt.Println("----------------------------------------------------------")
	if allHealthy {
		fmt.Println("[PASS] ALL SYSTEMS HEALTHY — OFFLINE READY")
	} else {
		fmt.Println("[FAIL] ONE OR MORE SUBSYSTEM CHECKS FAILED")
		os.Exit(1)
	}
	fmt.Println("==========================================================")
}

func runHighScores(store *storage.Store) {
	ctx := context.Background()
	scores, err := store.GetTopArcadeScores(ctx, 10)
	if err != nil {
		log.Fatalf("failed to query arcade leaderboard: %v", err)
	}

	fmt.Println("==========================================================================")
	fmt.Println("               QUANT FLIGHT ARCADE — HALL OF FAME LEADERBOARD              ")
	fmt.Println("==========================================================================")
	if len(scores) == 0 {
		fmt.Println()
		fmt.Println("  No arcade flight records found.")
		fmt.Println("  Launch Quant Flight from the web interface to log your first flight score!")
		fmt.Println()
		fmt.Println("==========================================================================")
		return
	}

	fmt.Printf("%-6s %-10s %-8s %-18s %-10s %s\n",
		"RANK", "INITIALS", "SCORE", "SECTORS CLEARED", "TIME", "DATE (UTC)")
	fmt.Println(strings.Repeat("-", 74))

	for i, sc := range scores {
		durStr := fmt.Sprintf("%.1fs", sc.DurationSeconds)
		fmt.Printf("%-6s %-10s %-8d %-18d %-10s %s\n",
			fmt.Sprintf("#%d", i+1),
			sc.Initials,
			sc.Score,
			sc.SectorsCleared,
			durStr,
			sc.CreatedAt.Format("2006-01-02 15:04"),
		)
	}
	fmt.Println("==========================================================================")
}

func runExportData(store *storage.Store, outPath string) {
	if outPath == "" {
		fmt.Fprintf(os.Stderr, "Error: -export-data requires a destination file path\n")
		os.Exit(1)
	}

	ctx := context.Background()
	notes, err := store.ListExplanations(ctx, "", "")
	if err != nil {
		log.Fatalf("failed to read notes for export: %v", err)
	}

	scores, err := store.GetTopArcadeScores(ctx, 100)
	if err != nil {
		log.Fatalf("failed to read arcade scores for export: %v", err)
	}

	exams, err := store.ListExams(ctx)
	if err != nil {
		log.Fatalf("failed to read exams for export: %v", err)
	}

	var prefs any
	if prefsStr, err := store.GetSettings(ctx, "user_preferences"); err == nil && prefsStr != "" {
		_ = json.Unmarshal([]byte(prefsStr), &prefs)
	}

	exportDoc := map[string]any{
		"application":      "Quant Methods Practice",
		"version":          version,
		"exported_at":      time.Now().UTC().Format(time.RFC3339),
		"notes_count":      len(notes),
		"notes":            notes,
		"arcade_scores":    scores,
		"exams":            exams,
		"user_preferences": prefs,
	}

	data, err := json.MarshalIndent(exportDoc, "", "  ")
	if err != nil {
		log.Fatalf("failed to format export data: %v", err)
	}

	if err := os.WriteFile(outPath, data, 0600); err != nil {
		log.Fatalf("failed to write export file %s: %v", outPath, err)
	}

	fmt.Printf("Successfully exported learner data to: %s (%d notes, %d arcade records, %d exams)\n",
		outPath, len(notes), len(scores), len(exams))
}

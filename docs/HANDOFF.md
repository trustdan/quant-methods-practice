# Handoff

## Current state - October 5, 2026

Stage 00 through Stage 10 are complete and verified. Stage 11 (ChatGPT plan sign-in) is **implemented and mock-verified; live verification is pending** an eligible ChatGPT account sign-in, inference and sign-out (the stage's exit gate). Stage 10 and Stage 11 work is uncommitted in the working tree. The standalone repository is active on branch `main` at `https://github.com/trustdan/quant-methods-practice.git`.

### Stage 00 Record
- Toolchain versions: Go `go1.27.1` (windows/amd64), Node `v22.21.1`, npm `10.9.4`, Python `3.13.15`, Windows 11 amd64 (Microsoft Edge available).
- Module path: `github.com/trustdan/quant-methods-practice`.
- Course material inventory: Confirmed `course-materials/private` is empty (`.gitkeep` only). Baseline learner-reported topics remain recorded in [docs/COURSE-MAP.md](COURSE-MAP.md).
- Stale move/setup instructions updated across `README.md`, `PLAN.md`, `docs/BOOTSTRAP.md`, `docs/COURSE-MAP.md`, and `docs/DECISIONS.md`.

### Stage 01 Implementation
- Go backend: Loopback HTTP server in [internal/httpapi](file:///internal/httpapi/server.go) and CLI entry point in [cmd/quant-practice](file:///cmd/quant-practice/main.go) with embedded production frontend assets in [internal/assets](file:///internal/assets/assets.go).
- Security boundaries: Loopback-only binding, Host header validation (DNS rebinding prevention), Origin validation on mutating requests, HttpOnly SameSite session cookies, single-use bootstrap tokens via URL fragment `#bootstrap=<token>`, strict Content-Security-Policy headers, and SPA route fallback.
- Web frontend: React 18, TypeScript strict, Vite, Vanilla CSS design system, [MathMarkdown](file:///web/src/components/MathMarkdown.tsx) component with DOMPurify sanitization and MathJax typeset lifecycle, [StageStrip](file:///web/src/components/StageStrip.tsx) (7 stages), [QuestionStrip](file:///web/src/components/QuestionStrip.tsx), central keyboard command resolver [keymap.ts](file:///web/src/navigation/keymap.ts), and static ungraded math demonstration shell.
- MathJax bundling: MathJax v3 TeX-SVG bundled completely locally in `web/public/vendor/mathjax/` via `web/scripts/copy-mathjax.js` with deterministic prebuild hook. Zero external CDN or network requests required.
- Build & test pipeline: `scripts/build.ps1`, `scripts/build.sh`, `scripts/test.ps1`, `scripts/test.sh`, and GitHub Actions CI in `.github/workflows/ci.yml`.

### Stage 02 Implementation
- Domain contracts ([internal/domain](file:///internal/domain)):
  - Stable IDs: Regex validation (`^[a-z][a-z0-9_]*$`) in `id.go`.
  - Question template: `QuestionTemplate`, `StageTemplate`, `Option`, `ApprovalRecord`, `NumericPolicy`, `NumericForm`, `TemplateStatus`, `TemplateKind` in `template.go`.
  - Strict answer types: `ExpectedAnswer` with custom JSON unmarshaling that strictly separates choice and numeric expected answers and disallows unknown fields per variant.
  - Instances & attempts: `QuestionInstance`, `StageInstance` in `instance.go`, plus `StageAttempt`, `SubmittedAnswer`, `AssistanceType`, `SessionSettings`, and `PracticeSession` in `session.go`. Pure domain types with no HTTP, UI, or storage imports.
- Bank contracts & parsing ([internal/bank](file:///internal/bank)):
  - Family registry: `Family` interface, `Registry`, and `BinomialFamily` (`binomial_pmf`, rule version 1) with parameter validation: integer $n \ge 0$, $0 \le p \le 1$, integer $0 \le k \le n$, and rejection of unknown parameters.
  - Strict parser & semantic validator: `StrictDecodeTemplate` (using `DisallowUnknownFields`), `ValidateTemplate` checking ID patterns, stage count (1-7), choice option bounds (2-4), option uniqueness, numeric policy invariants, concept-to-stage 1-to-1 evidence mapping, and human approval requirement for active status.
  - Active bank loading & directory validation: `LoadActiveBank` enforcing admission of active approved records only (rejecting drafts/retired), `ValidateTemplateFile`, and `ValidateBankDir`.
- CLI integration ([cmd/quant-practice](file:///cmd/quant-practice/main.go)):
  - Added `-validate-bank` to validate template files or directories.
  - Added `-active-bank` to inspect approved active templates in a bank directory.

### Stage 03 Implementation
- Mathematical engine ([internal/mathengine](file:///internal/mathengine)):
  - Combinatorics ([internal/mathengine/combinatorics.go](file:///internal/mathengine/combinatorics.go)): Exact arbitrary-precision combinations `ChooseBigInt`, factorials `FactorialBigInt`, stable log-gamma combinations `LogChoose` via `math.Lgamma`, float combinations `Choose` (exact conversion for $n \le 60$, stable exponentiation for larger $n$), and LaTeX formatting `FormatCombinationLaTeX`.
  - Binomial distribution ([internal/mathengine/binomial.go](file:///internal/mathengine/binomial.go)): `BinomialDistribution` with $n \ge 0, 0 \le p \le 1$; moments (`Mean = np`, `Variance = np(1-p)`, `StdDev = sqrt(np(1-p))`); `PMF` preserving exact dyadic rationals for $n \le 30$ and stable log-space via `math.Log1p(-p)` for $n > 30$; `ExactPMFRat` returning exact `*big.Rat` (e.g. $3/8$ for $n=4, p=1/2, k=2$); `CDF` and `Survival` computing tails directly when $k > n/2$ to prevent catastrophic cancellation; `SumIndependentBinomials` enforcing that sum is binomial only when trial success probabilities are equal.
  - Canonical discrete events ([internal/mathengine/events.go](file:///internal/mathengine/events.go)): `DiscreteEvent` with relation enum (`equal`, `at_most`, `less_than`, `at_least`, `greater_than`, `between_inclusive`); canonicalization ($X < k \to X \le k-1$, $X > k \to X \ge k+1$); LaTeX and description formatters; `ParseEventRelation`; `EvaluateBinomial` and `EvaluateBinomialRat`.
  - Poisson distribution ([internal/mathengine/poisson.go](file:///internal/mathengine/poisson.go)): `PoissonDistribution` with $\lambda \ge 0$; moments (`Mean = lambda`, `Variance = lambda`); `PMF` and `LogPMF`; `CDF` and `Survival`; explicit $\lambda = 0$ handling ($P(X=0)=1, P(X>0)=0$); `SumIndependentPoissons` ($\lambda_{\text{sum}} = \lambda_1 + \lambda_2$).
  - Set probability rules ([internal/mathengine/setprob.go](file:///internal/mathengine/setprob.go)): `ValidateSetProbabilities` with Fréchet inequality bounds ($\max(0, P(A)+P(B)-1) \le P(A \cap B) \le \min(P(A), P(B))$); `Union`; `Complement`; `Conditional`; `AreIndependent`; `AreDisjoint`; enforcement of invariant that disjointness does not imply independence.
  - Linear combinations & sums ([internal/mathengine/linearcomb.go](file:///internal/mathengine/linearcomb.go)): `LinearCombMean` ($E[aX + bY + c] = aE[X] + bE[Y] + c$, holds universally); `LinearCombVariance` ($a^2 Var(X) + b^2 Var(Y) + 2ab Cov(X, Y)$, constant contributes 0 variance); Cauchy-Schwarz covariance validation; `LinearCombVarianceIndependent`; `CovarianceFromCorrelation`; verification that variance distinguishes equal means under different dependence.
  - Numeric parsing and policy grading ([internal/mathengine/numeric.go](file:///internal/mathengine/numeric.go)): `ParseNumericInput` supporting decimal, explicit percent, and simple fractions; strict rejection of ambiguous commas with clear actionable guidance; rejection of unevaluated expressions/LaTeX; strict enforcement that bare `37.5` is not silently interpreted as `37.5%`; `GradeNumericAnswer` using per-stage `NumericPolicy` tolerance inequality $|x - y| \le \max(absTol, relTol \times |y|)$, stated rounding intervals with tie-breaking conventions, and diagnostic feedback for bare percentages.
  - Canonical derivation & distractors ([internal/mathengine/derivation.go](file:///internal/mathengine/derivation.go)): `DeriveBinomialProblem` deterministically computing canonical probabilities, exact rationals, LaTeX expressions, calculation steps, and 8 reviewed misconception distractors (`missing_combination`, `missing_failure_factor`, `exactly_as_at_most`, `exactly_as_at_least`, `complement_event`, `experiment_as_single_trial`, `target_as_trial_count`, `count_as_probability`); `CheckDistractorCollisions` detecting parameter combinations where distractors collide with the key under policy tolerance.
  - CLI integration ([cmd/quant-practice](file:///cmd/quant-practice/main.go)):
    - Added `-eval-binomial` CLI flag to inspect canonical derivation and distractors (e.g. `.\bin\quant-practice.exe -eval-binomial "n=4,p=0.5,k=2"`).

### Stage 04 Implementation
- Approved curriculum & review provenance:
  - [curriculum/approved/binomial-fair-coin-exactly-two.json](file:///curriculum/approved/binomial-fair-coin-exactly-two.json): Active approved 7-stage template ($n=4, p=0.5, k=2$) with review record, delegation basis, 1-to-1 concept evidence mapping, and reviewed distractors.
  - [curriculum/reviews/binomial-fair-coin-exactly-two.review.json](file:///curriculum/reviews/binomial-fair-coin-exactly-two.review.json): Review audit record documenting mathematical derivation ($P(X=2)=0.375=3/8$), moment checks ($\mu=2, \sigma^2=1, \sigma=1$), 0 distractor collisions under tolerance, and pedagogical constraints.
- Drill orchestration engine ([internal/drill](file:///internal/drill)):
  - [internal/drill/types.go](file:///internal/drill/types.go): Core types for stage progress status, session state machine, session commands, public projections, and drill recap.
  - [internal/drill/engine.go](file:///internal/drill/engine.go): `DrillSession` and `SessionManager` managing command execution (`submit_answer`, `request_hint`, `navigate_stage`, `reset_drill`), idempotency caching, revision concurrency checks, eligibility constraints, and answer key withholding on unresolved stages.
  - [internal/drill/grading.go](file:///internal/drill/grading.go): Choice option matching, misconception targeting with causal hints, numeric policy grading via `mathengine`, and invalid input non-grading.
  - [internal/drill/recap.go](file:///internal/drill/recap.go): Post-drill recap generator integrating `mathengine.DeriveBinomialProblem` and per-stage decision outcomes (`first_try`, `retry`, `revealed`).
- HTTP API integration ([internal/httpapi](file:///internal/httpapi)):
  - Added endpoints: `POST /api/practice/sessions` (create drill session), `GET /api/practice/sessions/{id}` (public session projection), `POST /api/practice/sessions/{id}/commands` (command mutations).
  - Robust active bank discovery across working directory structures.
- Web UI & navigation:
  - [web/src/features/practice/PracticeDrill.tsx](file:///web/src/features/practice/PracticeDrill.tsx): Full 7-stage interactive drill workspace with scenario context, responsive stage strip, choice options, numeric calculation field, causal hints, solution reveals, and keyboard navigation.
  - [web/src/features/practice/DrillRecapView.tsx](file:///web/src/features/practice/DrillRecapView.tsx): Worked solution recap with analytical derivation, moments, and stage-by-stage decision summaries.
  - [web/src/app/App.tsx](file:///web/src/app/App.tsx): Connected active drill with loopback server and offline fallback simulation.
  - [web/src/types/practice.ts](file:///web/src/types/practice.ts): Strict TypeScript DTO and domain interfaces.

### Stage 05 Implementation
- Durable storage engine ([internal/storage](file:///internal/storage)):
  - [internal/storage/migrations/001_initial_schema.sql](file:///internal/storage/migrations/001_initial_schema.sql): Complete transactional schema covering all 16 tables required by [docs/STORAGE.md](STORAGE.md) (`schema_migrations`, `settings`, `sessions`, `question_instances`, `drill_stage_states`, `attempts`, `assistance_events`, `session_drafts`, `command_idempotency`, `mastery_projections`, `candidate_questions`, `content_approval_events`, `saved_explanations`, `tutor_drafts`, `exam_responses`, `arcade_scores`).
  - [internal/storage/migrations.go](file:///internal/storage/migrations.go): Migration runner with SHA256 checksum verification, pre-migration backup trigger on schema upgrades, and transactional rollback on failures.
  - [internal/storage/db.go](file:///internal/storage/db.go): SQLite connection initializer using pure Go `modernc.org/sqlite` driver, configuring WAL journal mode, `foreign_keys=ON`, `busy_timeout=5000`, single-writer connection pooling, and in-memory test databases.
  - [internal/storage/paths.go](file:///internal/storage/paths.go): Platform-specific user data directory resolver (Windows `%APPDATA%\quant-methods-practice`, macOS `Library/Application Support/quant-methods-practice`, Linux `$XDG_DATA_HOME/quant-methods-practice`) with `--data-dir`, `--db`, and `QUANT_DATA_DIR` override support.
  - [internal/storage/backup.go](file:///internal/storage/backup.go): Point-in-time consistent SQLite backup using native `VACUUM INTO`, safe for concurrent WAL mode writes.
  - [internal/storage/store.go](file:///internal/storage/store.go): Transactional persistence store implementing `drill.SessionStore`, saving sessions, immutable `QuestionInstance` snapshots, stage states, attempts, assistance events, session drafts, and command idempotency results.
- Drill engine persistence integration ([internal/drill](file:///internal/drill)):
  - Inverted dependency interface `drill.SessionStore` ensuring `internal/drill` and `internal/domain` have zero storage or driver dependencies.
  - In `DrillSession.ExecuteCommand`: durable idempotency lookup checks `command_idempotency` before command evaluation; prevents duplicate attempts, duplicate assistance, and spurious revision bumps.
  - Session state mutations, attempts, and command outcomes atomically persist to SQLite.
  - `SessionManager.GetActiveSession` recovers active/completed drills on server startup.
- HTTP API & CLI integration:
  - [internal/httpapi/server.go](file:///internal/httpapi/server.go): Added `Store` to server config; `GET /api/practice/sessions` recovers the stored active drill on page reload or server restart.
  - [cmd/quant-practice/main.go](file:///cmd/quant-practice/main.go): Added `--db` and `--backup` CLI flags, automated pre-migration backups, and graceful SQLite connection teardown.
- End-to-end automated replay verification:
  - [internal/storage/replay_test.go](file:///internal/storage/replay_test.go): Comprehensive test suite covering the 3 critical Stage 05 invariants:
    1. Close/reopen restores drill (`TestCloseReopenRestoresDrill`).
    2. Duplicate command cannot inflate evidence (`TestDuplicateCommandCannotInflateEvidence`).
    3. Changed bank cannot change historical content/grades (`TestChangedBankCannotChangeHistoricalContentOrGrades`).
    4. Multi-tab concurrency conflict rejection (`TestMultiTabRevisionConflictRejection`).
  - [tests/e2e/foundation.spec.ts](file:///tests/e2e/foundation.spec.ts): Updated Playwright E2E browser test verifying that an active session survives server process termination and reload, restoring at Stage 2 with Stage 1 completed.

### Stage 06 Implementation
- Keyboard navigation engine ([web/src/navigation/keymap.ts](file:///web/src/navigation/keymap.ts)):
  - Pure central keyboard resolver implementing the canonical [docs/NAVIGATION.md](NAVIGATION.md) keymap.
  - Shortcuts:
    - Stage navigation: `h`/`l` and `ArrowLeft`/`ArrowRight`.
    - Problem navigation: `Ctrl+ArrowLeft`/`Ctrl+ArrowRight` and `Shift+H`/`Shift+L`.
    - Choice selection: `1`-`4` or `a`-`d` in choice stages; strictly ignored in numeric mode.
    - Choice movement: `j`/`k` and `ArrowDown`/`ArrowUp`.
    - Submission / retry / advance: `Enter`, `Space`.
    - Assistance: `?`, `e`, `F1`.
    - Reading / notes scroll: `j`/`k` (scroll step), `u`/`d` or `PageUp`/`PageDown` (half page), `gg` (top with 500ms timeout), `G` (bottom).
    - View toggles: `s` (mastery), `V` (notes), `t` (settings), `p` (candidates), `n` (ai), `J` (full solution), `F` (cases), `E` (exam), `A` (arcade), `L` (leaderboard), `i` (intensity), `[`/`]` (session size).
    - Leave intent: `q` (prompt), `y` (confirm save & exit), `n` (discard & exit), `Esc` (stay).
  - Safety constraints:
    - Native editable fields (`<input>`, `<textarea>`, `[contenteditable]`) are strictly protected from single-character shortcuts; only `Enter` and `Esc` resolve.
    - Browser reserved shortcuts (`Ctrl+C`, `Ctrl+L`, `Ctrl+W`, `Ctrl+R`, `F5`, `Ctrl +/-`) pass through unconditionally.
    - IME composition (`isComposing`) cancels shortcuts immediately.
- Usable stage strip & question list:
  - [web/src/components/StageStrip.tsx](file:///web/src/components/StageStrip.tsx): Visual status indicators (`completed` `✓`, `revealed` `rev`, `locked` `🔒`), accessible descriptions, and informative click-notice on locked stages explaining prerequisite completion.
  - [web/src/components/QuestionStrip.tsx](file:///web/src/components/QuestionStrip.tsx): Horizontal problem selector with status markers (`completed`, `in_progress`, `skipped`), and `< Prev Problem` / `Next Problem >` buttons with shortcut hints.
- Unsent answer preservation & draft persistence:
  - [internal/drill/types.go](file:///internal/drill/types.go) & [internal/drill/engine.go](file:///internal/drill/engine.go): Added `CmdSaveDraft` and `CmdClearDraft` to `DrillSession`. Drafts save to `session_drafts` in SQLite without creating attempts or inflating evidence; drafts are automatically purged on stage submission or drill reset.
  - [internal/storage/store.go](file:///internal/storage/store.go): Implemented `SaveDraftAnswer` and populated `DraftAnswer` upon session retrieval from `session_drafts`.
  - [web/src/features/practice/PracticeDrill.tsx](file:///web/src/features/practice/PracticeDrill.tsx): In-memory draft tracking across stage visits with debounced background synchronization to the server; unsubmitted answers remain intact when navigating backward or forward.
- Browser usability & multi-tab concurrency:
  - [web/src/app/App.tsx](file:///web/src/app/App.tsx):
    - Concurrency conflict handling (HTTP 409): displays clear conflict notification and synchronizes to latest server revision.
    - Window focus listener checks server revision to detect external changes.
    - Dirty draft leave-intent modal with `beforeunload` browser unload guard.
    - Context-sensitive Help modal with focus trapping and previous element focus restore.
    - Reading mode keyboard scrolling container.

### Stage 07 Implementation
- Reviewed Small Bank Across Learner's Covered Topics ([curriculum/approved](file:///curriculum/approved) & [curriculum/reviews](file:///curriculum/reviews)):
  - Expanded approved curriculum bank from 1 to 10 verified, human-reviewed questions covering foundational probability, discrete distributions, and moment operations:
    1. `binomial_fair_coin_exactly_two` (Module 2, 7 stages): Canonical coin tossing derivation ($n=4, p=0.5, k=2$).
    2. `binomial_defective_at_most_one` (Module 2, 4 stages): Defective component batch with complement ($n=5, p=0.10, k \le 1$).
    3. `binomial_service_calls_zero` (Module 2, 3 stages): Zero service requests boundary calculation ($n=3, p=0.20, k=0$).
    4. `poisson_call_center_arrivals` (Module 2, 4 stages): Call center arrivals ($k=2, \lambda=3.0$).
    5. `poisson_rare_event_one` (Module 2, 3 stages): Single fabric defect ($k=1, \lambda=0.5$).
    6. `poisson_server_failures_zero` (Module 2, 3 stages): Zero server alert boundary ($k=0, \lambda=1.2$).
    7. `set_probability_complement_rule` (Module 1, 2 stages): Complement rule on market surge ($P(A)=0.70 \implies P(A^c)=0.30$).
    8. `set_probability_union_rule` (Module 1, 3 stages): Addition rule union ($P(A)=0.4, P(B)=0.5, P(A \cap B)=0.2 \implies P(A \cup B)=0.7$).
    9. `set_probability_conditional` (Module 1, 3 stages): Conditional probability definition & independence check ($P(A|B)=0.625$).
    10. `linear_comb_sum_variance_independent` (Module 2, 4 stages): Sum and variance of independent indicators ($Var(X_1+X_2)=0.42$).
  - Full review provenance audit files created in `curriculum/reviews/*.review.json` documenting mathematical verification, 0 distractor collisions, and pedagogical invariants.
- Multi-Question Storage & Replay ([internal/storage](file:///internal/storage)):
  - Updated [internal/storage/store.go](file:///internal/storage/store.go) `SaveSession` and `GetSession` to persist multi-question metadata (`SessionSettings`, question instances, and per-question stage states).
  - Implemented scoped stage keys (`scopedStageKey(instID, stageID, isMulti)`) ensuring multiple questions in the same session maintain independent stage records, attempts, assistance, and drafts while retaining 100% backward compatibility with single-question sessions.
  - Added [internal/storage/replay_test.go](file:///internal/storage/replay_test.go) `TestMultiQuestionSessionPersistenceAndReplay` verifying a 10-question drill persists and recovers cleanly across process shutdown.
- Multi-Question Engine Orchestration ([internal/drill](file:///internal/drill)):
  - Exported `SyncCurrentQuestion` and updated `handleNavigateQuestion` to atomically persist active question state transitions.
  - Added user settings persistence (`SaveSettings` / `GetSettings`) to the `drill.SessionStore` interface.
- HTTP API Endpoints ([internal/httpapi](file:///internal/httpapi)):
  - `POST /api/practice/sessions`: supports `question_count`, `module_ids`, `intensity`, `seed`, and `template_id`.
  - `GET /api/settings` and `POST /api/settings`: persistent user preference storage in SQLite.
  - `GET /api/bank`: lists approved templates with families, modules, titles, and stage counts.
  - Guaranteed stable sorting placing canonical `binomial_fair_coin_exactly_two` first.
- CLI Essentials & Launch Scripts:
  - Added CLI flags to [cmd/quant-practice/main.go](file:///cmd/quant-practice/main.go): `-questions`, `-module`, `-intensity`, `-seed`, and `-list-bank`.
  - Created standalone launch scripts: [scripts/start.ps1](file:///scripts/start.ps1) (PowerShell), [scripts/start.bat](file:///scripts/start.bat) (Windows Command Prompt), [scripts/start.sh](file:///scripts/start.sh) (Linux/macOS Bash).
- Web Frontend: Reference Library, Settings Modal, and Responsive Math:
  - [web/src/features/reference/ReferenceLibrary.tsx](file:///web/src/features/reference/ReferenceLibrary.tsx): Offline reference library with search, category filtering (Foundations, Distributions, Moments), LaTeX MathJax rendering, and exam pitfalls (`r` key shortcut).
  - [web/src/features/settings/SettingsModal.tsx](file:///web/src/features/settings/SettingsModal.tsx): Interactive settings modal (`t` key shortcut) for question count selection (5, 10, 15, 20), curriculum module filtering, drill intensity selection, and seed inputs.
  - [web/src/components/QuestionStrip.tsx](file:///web/src/components/QuestionStrip.tsx) & [web/src/index.css](file:///web/src/index.css): Responsive layout and math rendering for narrow screens (e.g. 375px mobile viewport); flex wrapping and SVG scaling ensure zero horizontal blowout (`scrollWidth <= 376`).
  - [web/src/app/App.tsx](file:///web/src/app/App.tsx): Integrated reference library tab, settings modal, question navigation, and preference persistence.
- E2E Replay & Usability Browser Verification:
  - [tests/e2e/foundation.spec.ts](file:///tests/e2e/foundation.spec.ts): Added dedicated Stage 07 test verifying:
    1. 10-question session with 10 pills in `QuestionStrip`.
    2. Problem 1 -> Problem 2 -> Problem 3 navigation via pill clicks and `Control+ArrowRight` / `Control+ArrowLeft` shortcuts.
    3. Offline `ReferenceLibrary` opening, LaTeX math rendering, category filtering, and return to practice.
    4. `SettingsModal` opening and `Escape` key dismissal.
    5. Narrow viewport responsive layout test (375px wide) verifying no horizontal overflow.
    6. Complete process termination, loopback server restart on a new port, page reload, and SQLite replay of the 10-question session with active Problem 2 restored.
    7. Captured verified screenshot [tests/e2e/screenshots/stage07_verified.png](file:///tests/e2e/screenshots/stage07_verified.png).

### Stage 08 Implementation
- Mastery & Scheduling Engine ([internal/mastery](file:///internal/mastery)):
  - Domain types ([internal/mastery/types.go](file:///internal/mastery/types.go)): `ScaffoldLevel` (`full`, `intermediate`, `faded`), `MasteryStatus` (`new`, `learning`, `transferring`, `mastered`), `EvidenceOutcome`, `ConceptMastery`, `RawExposure`, `MasterySummary`, `EvidencePolicyVersion = 1`.
  - Bayesian decay & scheduling policy ([internal/mastery/policy.go](file:///internal/mastery/policy.go)): Informative uniform prior ($\alpha=1, \beta=1$), Bayesian mean $(S+1)/(S+E+2)$, 3-day exponential half-life decay computed strictly at read time, clock rollback defense ($t_{\text{now}} < t_{\text{last}}$), scaffold level derivation with decay margin tolerance, status determination, and priority scoring.
  - Evidence ledger ([internal/mastery/ledger.go](file:///internal/mastery/ledger.go)): Strictly enforces **one contribution per concept per instance** (subsequent attempts on the same instance are ignored), preserves first error on retry, categorizes assistance (`independent_first`, `hinted_retry`, `reference_used`, `revealed`, `guided_contrast`), tracks distinct setting groups, and requires $\ge 2$ distinct setting groups plus $\ge 10$ minutes delayed transfer for graduation.
  - Scaffold degradation policy ([internal/mastery/scaffold.go](file:///internal/mastery/scaffold.go)): `DegradeStages` degrades 7-stage templates to 4 stages (Intermediate) or 2 stages (Faded) while preserving concept evidence attribution. `DetermineTemplateScaffold` evaluates all template concepts against the learner mastery profile.
  - Seeded weighted selection ([internal/mastery/selector.go](file:///internal/mastery/selector.go)): Implements non-zero eligibility floor (0.1), concept need weighting, ~20% mixed review allocation, and anti-repeat penalties (exact template and setting group).
  - Contrast registry & bounded queue ([internal/mastery/contrast.go](file:///internal/mastery/contrast.go)): Pairs eligible misconceptions (`exactly_as_at_most`, `missing_combination`, `independence_concept`, etc.) with approved contrast partners.
  - Mastery tests ([internal/mastery/mastery_test.go](file:///internal/mastery/mastery_test.go)): 8/8 comprehensive unit tests passing.
- Durable Mastery Storage & Rebuild ([internal/storage](file:///internal/storage)):
  - [internal/storage/mastery.go](file:///internal/storage/mastery.go): Implemented `GetHistoricalExposures` extracting attempts and assistance from SQLite, `SaveMasteryProjection`, `GetAllMasteryProjections`, `RebuildMastery` reconstructing projections from historical database records, and `GetMasterySummary`.
  - [internal/storage/mastery_storage_test.go](file:///internal/storage/mastery_storage_test.go): Tested projection roundtrip persistence and rebuild from historical attempts.
- Drill Orchestration & Bounded Contrast Queue ([internal/drill](file:///internal/drill)):
  - [internal/drill/engine.go](file:///internal/drill/engine.go): Added `CreateMultiQuestionSessionWithScaffolds`, contrast partner detection on misconception submissions, and bounded queuing with strict single-contrast limit (`!CurrentQuestionIsContrast() && ContrastCount == 0`), ensuring contrasts **cannot chain or cause unbounded session expansion**.
  - Dynamic scaffold indicators and contrast badges surfaced in `ToPublicView`.
  - [internal/drill/drill_test.go](file:///internal/drill/drill_test.go): Added tests verifying contrast partner insertion and scaffold degradation.
- HTTP API & CLI:
  - [internal/httpapi/server.go](file:///internal/httpapi/server.go): Added `GET /api/mastery` endpoint returning `MasterySummary`, wired contrast finder and weighted selection in `POST /api/practice/sessions`.
  - [cmd/quant-practice/main.go](file:///cmd/quant-practice/main.go): Added `-mastery` CLI flag displaying formatted mastery table with status, scaffold level, decayed score, success/error counts, groups, and transfer status.
- Web UI & E2E Browser Verification:
  - [web/src/features/mastery/MasteryView.tsx](file:///web/src/features/mastery/MasteryView.tsx): Interactive mastery and transfer dashboard with overall retention score, status filter pills, search input, decayed score bars, and evidence metrics.
  - [web/src/features/mastery/MasteryView.test.tsx](file:///web/src/features/mastery/MasteryView.test.tsx): 4 Vitest unit tests verifying API fetch, rendering, filtering, search, and error notices.
  - [web/src/features/practice/PracticeDrill.tsx](file:///web/src/features/practice/PracticeDrill.tsx): Added `#contrast-badge` and `#scaffold-indicator` indicating Full Guidance (7 stages), Intermediate (4 stages), or Faded (2 stages).
  - [web/src/app/App.tsx](file:///web/src/app/App.tsx): Wired `activeTab === 'mastery'`, `s` key shortcut, `Esc` return to practice, and header Mastery tab.
  - [tests/e2e/foundation.spec.ts](file:///tests/e2e/foundation.spec.ts): Added dedicated Stage 08 Playwright E2E browser test verifying mastery dashboard, filter pills, search input, scaffold indicators, keyboard shortcuts, and captured verified screenshot [tests/e2e/screenshots/stage08_verified.png](file:///tests/e2e/screenshots/stage08_verified.png).

### Stage 09 Implementation
- Durable storage engine ([internal/storage](file:///internal/storage)):
  - [internal/storage/notes.go](file:///internal/storage/notes.go): Implemented persistence methods for `saved_explanations` and `tutor_drafts` tables (`SaveExplanation`, `GetExplanation`, `ListExplanations`, `DeleteExplanation`, `SaveTutorDraft`, `GetTutorDraft`, `ClearTutorDraft`).
  - [internal/storage/notes_test.go](file:///internal/storage/notes_test.go): 5 unit tests verifying CRUD roundtrip, topic/search filtering, and draft management.
- Domain & Drill Engine ([internal/domain](file:///internal/domain), [internal/drill](file:///internal/drill)):
  - Added `AssistanceTutor` to `domain.AssistanceType` in [internal/domain/session.go](file:///internal/domain/session.go).
  - Integrated `AssistanceTutor` in [internal/mastery/ledger.go](file:///internal/mastery/ledger.go) assistance classification.
  - Added `RecordTutorAssistance` method to `DrillSession` in [internal/drill/engine.go](file:///internal/drill/engine.go) to track tutor exposure on unresolved stages without modifying grades or answer keys.
- Tutor Engine ([internal/tutor](file:///internal/tutor)):
  - [internal/tutor/types.go](file:///internal/tutor/types.go): `TutorAction` (`hint`, `explain`, `follow_up`), `FollowUpKind` (`explain_differently`, `worked_example`, `why_condition_matters`, `compare_concepts`, `custom`), `TutorRequest`, `TutorEvent`, `TutorEventType` (`started`, `text_delta`, `complete`, `cancelled`, `fallback`, `error`), and `ProviderCapabilities`.
  - [internal/tutor/offline.go](file:///internal/tutor/offline.go): Curriculum-grounded `OfflineTutor` with causal hints, analytical step-by-step solutions (binomial $P(X=2)=0.375=3/8$, moments, Poisson, sets, sums), all 4 follow-up modes, and cancellable chunk streaming.
  - [internal/tutor/fake_provider.go](file:///internal/tutor/fake_provider.go): Controllable mock provider for simulating async delays, cancellation, and mid-stream error fallback.
  - [internal/tutor/manager.go](file:///internal/tutor/manager.go): `TutorManager` managing lifecycle, in-flight cancellations, and fallback to `OfflineTutor`.
  - [internal/tutor/export.go](file:///internal/tutor/export.go): `ExportNoteToMarkdown` formatting clean UTF-8 markdown with YAML frontmatter, advisory warning, problem context, and LaTeX math compatible with Obsidian and Typora.
  - [internal/tutor/tutor_test.go](file:///internal/tutor/tutor_test.go): 10 unit tests covering hints, math derivations, follow-ups, streaming chunks, cancellations, provider fallback, markdown export, and grade mutation protection.
- HTTP API ([internal/httpapi](file:///internal/httpapi)):
  - [internal/httpapi/tutor_notes.go](file:///internal/httpapi/tutor_notes.go): Added endpoints `POST /api/tutor/requests`, SSE `GET /api/tutor/requests/{id}/events`, `DELETE /api/tutor/requests/{id}`, `GET/POST /api/notes`, `GET/DELETE /api/notes/{id}`, `GET /api/notes/{id}/export`, `POST /api/exports`, `GET/POST/DELETE /api/tutor/drafts/{id}`.
  - Updated [internal/httpapi/server.go](file:///internal/httpapi/server.go): Added `NoteStore` and `TutorManager` to `Config`/`Server`, set `WriteTimeout: 120 * time.Second` for SSE.
- CLI Integration ([cmd/quant-practice](file:///cmd/quant-practice/main.go)):
  - Added `-notes` (list saved notes in personal library) and `-export-notes <dir>` (export all notes to markdown files).
- Web Frontend ([web/src](file:///web/src)):
  - [web/src/types/tutor.ts](file:///web/src/types/tutor.ts): Strict TypeScript interfaces for tutor requests, events, notes, and drafts.
  - [web/src/features/tutor/AITutorPanel.tsx](file:///web/src/features/tutor/AITutorPanel.tsx): Interactive tutor panel with streaming MathMarkdown, causal hints, step-by-step solutions, 4 follow-up prompts, custom query input, cancellation, fallback notice, advisory warning banner, save-to-notes button, and leave-intent modal guard (`y`/`n`/`Esc`).
  - [web/src/features/notes/NoteLibraryView.tsx](file:///web/src/features/notes/NoteLibraryView.tsx): Dual-pane note library with search input, topic filters, keyboard navigation (`j`/`k`/`n`/`p`), MathMarkdown preview, markdown download export, and deletion.
  - [web/src/features/practice/PracticeDrill.tsx](file:///web/src/features/practice/PracticeDrill.tsx): Added `onOpenTutor` prop and "💡 AI Tutor (n)" button.
  - [web/src/app/App.tsx](file:///web/src/app/App.tsx): Added `'notes'` tab, `showTutorModal` state, header tabs for Notes (`V`) and AI Tutor (`n`), shortcuts (`VIEW_SAVED_NOTES`, `VIEW_AI_REQUEST`, `ESCAPE`), and integrated both views.
  - [web/src/features/tutor/AITutorPanel.test.tsx](file:///web/src/features/tutor/AITutorPanel.test.tsx) & [web/src/features/notes/NoteLibraryView.test.tsx](file:///web/src/features/notes/NoteLibraryView.test.tsx): 9 Vitest unit tests added. All 54 frontend unit tests pass (`npm run test`). `npm run typecheck` passes with 0 errors.
- End-to-end Automated Verification:
  - [tests/e2e/foundation.spec.ts](file:///tests/e2e/foundation.spec.ts): Added dedicated Stage 09 Playwright E2E browser test verifying AI tutor streaming with offline math rendering, leave-intent protection modal (`y` Save, `n` Discard, `Esc` Stay), streaming cancellation, note library viewing, searching, LaTeX math rendering in detail pane, and note markdown export. Captured verified screenshot [tests/e2e/screenshots/stage09_verified.png](file:///tests/e2e/screenshots/stage09_verified.png).

### Stage 10 Implementation
- Backend Credential Vault ([internal/auth](file:///internal/auth)):
  - [internal/auth/types.go](file:///internal/auth/types.go): `Route` constants (`offline`, `anthropic`, `gemini`, `openai`), `CredentialStatus`, `Vault` interface (`SaveKey`, `GetKey`, `DeleteKey`, `Status`), and `MaskKey` (e.g. `AIzaSy...9988`).
  - [internal/auth/memory_vault.go](file:///internal/auth/memory_vault.go): Concurrent thread-safe in-memory vault for session-only/test storage.
  - [internal/auth/file_vault.go](file:///internal/auth/file_vault.go): AES-256-GCM encrypted persistent file store with 0600 file permissions.
  - [internal/auth/dpapi_vault_windows.go](file:///internal/auth/dpapi_vault_windows.go): Windows DPAPI user-credential encryption leveraging native `CryptProtectData` and `CryptUnprotectData`.
  - [internal/auth/dpapi_vault_other.go](file:///internal/auth/dpapi_vault_other.go): Cross-platform fallback stub.
  - [internal/auth/vault.go](file:///internal/auth/vault.go): `StandardVault` coordinating persistent encrypted storage with backend environment variables (`ANTHROPIC_API_KEY`, `GEMINI_API_KEY`/`GOOGLE_API_KEY`, `OPENAI_API_KEY`).
  - [internal/auth/vault_test.go](file:///internal/auth/vault_test.go): 4 unit tests verifying memory vault, key masking, AES file encryption, and DPAPI persistence.
- Provider Adapters & Discovery ([internal/providers](file:///internal/providers)):
  - [internal/providers/types.go](file:///internal/providers/types.go): `ModelInfo`, `ProviderSummary`, `BudgetStatus`, and `ModelDiscoverer` interface.
  - [internal/providers/budget.go](file:///internal/providers/budget.go): `BudgetTracker` enforcing 20 external requests/session, input/output token tracking, and thread-safe resets.
  - [internal/providers/catalog.go](file:///internal/providers/catalog.go): `CatalogCache` with curated defaults for offline, Anthropic (`claude-3-5-sonnet-latest`, `claude-3-5-haiku-latest`), Gemini (`gemini-1.5-pro-latest`, `gemini-1.5-flash-latest`), and OpenAI (`gpt-4o`, `gpt-4o-mini`), 24-hour TTL caching, and custom model ID support.
  - [internal/providers/prompt.go](file:///internal/providers/prompt.go): Strict pedagogical system prompt preventing answer reveals on unresolved stages, causal hint instructions, and structured context assembly.
  - [internal/providers/anthropic.go](file:///internal/providers/anthropic.go): Anthropic Claude Messages API adapter with SSE streaming and `/v1/models` discovery.
  - [internal/providers/gemini.go](file:///internal/providers/gemini.go): Google Gemini API adapter with SSE streaming and `/v1beta/models` discovery.
  - [internal/providers/openai.go](file:///internal/providers/openai.go): OpenAI Chat Completions API adapter with SSE streaming and `/v1/models` discovery.
  - [internal/providers/manager.go](file:///internal/providers/manager.go): `ProviderManager` coordinating active routes, models, catalog caches, and automatic fallback to `OfflineTutor`.
  - [internal/providers/providers_test.go](file:///internal/providers/providers_test.go): 8 unit tests covering budget enforcement, catalog cache, Anthropic/Gemini/OpenAI mock streaming and discovery, manager coordination, offline zero-network invariant, and rate limit exponential backoff.
- HTTP API & CLI Integration:
  - [internal/httpapi/providers.go](file:///internal/httpapi/providers.go): Endpoints for `/api/providers`, `/api/providers/active`, `/api/providers/{route}/credentials`, `/api/providers/{route}/disconnect`, `/api/providers/{route}/models`, `/api/providers/{route}/models/refresh`, `/api/providers/{route}/models/custom`, `/api/providers/budget`, and `/api/providers/budget/reset`.
  - [internal/httpapi/providers_test.go](file:///internal/httpapi/providers_test.go): 15 unit tests verifying provider listing, credential save/disconnect without secret leakage in responses or logs, model switching, custom model additions, budget consumption, and zero network calls when offline.
  - [internal/httpapi/server.go](file:///internal/httpapi/server.go) & [internal/httpapi/tutor_notes.go](file:///internal/httpapi/tutor_notes.go): Integrated `Vault` and `ProviderManager` into server runtime and connected tutor request handler to route to active provider.
  - [cmd/quant-practice/main.go](file:///cmd/quant-practice/main.go): Added `-providers` CLI flag to inspect active provider and vault credential status.
- Web Frontend ([web/src](file:///web/src)):
  - [web/src/types/providers.ts](file:///web/src/types/providers.ts): TypeScript DTO interfaces.
  - [web/src/features/providers/ProviderSettings.tsx](file:///web/src/features/providers/ProviderSettings.tsx): UI component with provider cards, masked password input, model selection dropdown, dynamic model refresh, custom model input, and session budget meter.
  - [web/src/features/settings/SettingsModal.tsx](file:///web/src/features/settings/SettingsModal.tsx): Dual-tab navigation for Session Settings vs AI Providers & Credentials.
  - [web/src/features/tutor/AITutorPanel.tsx](file:///web/src/features/tutor/AITutorPanel.tsx): Header badge reflecting active provider and model, linking directly to settings.
  - [web/src/features/providers/ProviderSettings.test.tsx](file:///web/src/features/providers/ProviderSettings.test.tsx): 4 Vitest unit tests verifying provider listing, API key save, model select, and budget meter rendering.
- End-to-end Automated Verification:
  - [tests/e2e/foundation.spec.ts](file:///tests/e2e/foundation.spec.ts): Added dedicated Stage 10 Playwright E2E browser test verifying provider settings tab navigation, key saving, masked display, active provider toggling, and AI tutor provider badge rendering with zero network requests. Captured verified screenshot [tests/e2e/screenshots/stage10_verified.png](file:///tests/e2e/screenshots/stage10_verified.png).

### Stage 11 Implementation (live verification pending)
- Protocol reverified on October 5, 2026 against the official docs ([sign-in](https://developers.openai.com/siwc/token-sharing-open-source/sign-in), [accounts and sessions](https://developers.openai.com/siwc/token-sharing-open-source/profiles-and-sessions), [models and inference](https://developers.openai.com/siwc/token-sharing-open-source/models-and-inference), [token reference](https://developers.openai.com/siwc/token-sharing-open-source/token-reference), [errors and recovery](https://developers.openai.com/siwc/token-sharing-open-source/errors-and-recovery), [preview limitations](https://developers.openai.com/siwc/token-sharing-open-source/preview-limitations)) and the live `https://auth.openai.com/.well-known/openid-configuration` (issuer `https://auth.openai.com`, RS256 JWKS, revocation endpoint `/api/accounts/oauth/revoke`, public client auth `none`). Details recorded in [docs/PROVIDERS.md](PROVIDERS.md).
- [internal/siwc/oidc.go](../internal/siwc/oidc.go): endpoints, PKCE S256, stdlib RS256 ID-token verification (signature via JWKS, issuer, audience = issued client ID, expiry, nonce, subject). No new dependencies.
- [internal/siwc/client.go](../internal/siwc/client.go): `Client` with dynamic registration via `dynamic_agent_client` (issued `client_id` read from the callback and saved per account), persistent `urn:uuid:` host identifier (`ext_agent_host_id`), `agent_name_hint` only on new registration, `id_token_hint` on reauthorization, fresh state/nonce/PKCE per attempt, one-shot callback listener on `127.0.0.1:<ephemeral>/auth/callback` separate from the app server (Host-checked; forged state rejected without consuming the attempt; 10-minute timeout; graceful shutdown), `chatgpt.tokens.use.direct` grant check, refresh with rotation and the documented reauth error codes, revoke-then-delete sign-out, multiple accounts with selection, identity check on reauth (a different account replaces nothing). Accounts and tokens are stored as one record in the existing encrypted vault (DPAPI on Windows) under `chatgpt_plan_accounts`; the UI only receives redacted `AccountSummary` values.
- [internal/providers/chatgpt.go](../internal/providers/chatgpt.go): `ChatGPTPlanAdapter` (route `chatgpt`) calling `POST /v1/responses` with only `model`, `instructions`, `input` (array), `store:false`, `stream:true`; parses `response.output_text.delta`, treats only `response.completed` as success, surfaces `response.failed`/`response.incomplete`; maps documented `subscription_sharing_*` codes; a 429 usage limit is never retried, 503 is retried with bounded backoff; `subscription_sharing_invalid_user` marks the account for reauth. Model discovery via `GET /v1/models` filtered to `visibility: "list"`. No default model is assumed and it never reads an API key.
- [internal/providers/manager.go](../internal/providers/manager.go) / [types.go](../internal/providers/types.go): route registered; summaries now carry `auth_kind`, `billing` (`none` / `api_usage` / `chatgpt_plan`) and `account_label`; the plan route cannot be activated without a signed-in account that granted plan usage and a selected model. Startup still always begins offline.
- [internal/httpapi/chatgpt.go](../internal/httpapi/chatgpt.go): `POST/GET /api/providers/chatgpt/signin`, `POST .../signin/cancel`, `GET .../accounts`, `POST .../accounts/{key}/select`, `POST .../accounts/{key}/signout` (falls back visibly to offline if the active plan account goes away). Pasting tokens into `.../credentials` is rejected.
- [cmd/quant-practice/main.go](../cmd/quant-practice/main.go): `-providers` lists ChatGPT plan accounts and their state.
- Web: [ChatGPTPlanPanel.tsx](../web/src/features/providers/ChatGPTPlanPanel.tsx) (sign-in opens OpenAI's page with `noopener`, status polling, cancel, account select/renew/sign-out, billing explanation); [ProviderSettings.tsx](../web/src/features/providers/ProviderSettings.tsx) adds the `ChatGPT Plan` tab, renames `OpenAI` to `OpenAI API`, shows a billing label on every route, no longer carries one route's model into another, explains an empty catalog, and disables Set as Active without a model; [AITutorPanel.tsx](../web/src/features/tutor/AITutorPanel.tsx) badge shows the billing route.
- Verification tooling: [scripts/test.ps1](../scripts/test.ps1) now builds before browser tests and fails on any non-zero native exit code and on gofmt drift. **Previously it ignored npm/go exit codes and ran E2E against the old binary before rebuilding, so earlier HANDOFF entries saying "test.ps1 passed" did not prove every step passed.**

## Checks

Stage 11 session (October 5, 2026), commands actually run:
- `go test -count=1 ./internal/siwc/` and `go test -race -count=1 ./internal/siwc/`: Passed (8 tests, one with 5 subtests, against a fake auth server signing real RS256 tokens: RFC 7636 PKCE vector; new-registration flow and authorize parameters; forged-state rejection; nonce/audience/expiry/signature/missing-plan-scope rejection; authorization error callback; refresh rotation then `refresh_token_reused` reauth; reauth-as-different-account rejection; revoke-then-delete sign-out; cancelled listener closes).
- `go test -count=1 ./internal/...`: Passed (12 packages). Includes 8 new plan-adapter tests (permitted payload keys only, missing completion is an error, 429 not retried, 503 bounded retry, invalid user marks reauth, no API-key fallback and no network/budget use without a plan token, model visibility filtering, plan/API routes distinct) and 3 new HTTP tests (pasted tokens rejected, activation requires sign-in, sign-in start/cancel with the main server not serving the callback).
- `go vet ./internal/... ./cmd/...` and `gofmt -s -l cmd internal`: clean.
- `npm run typecheck`: Passed. `npm run test`: Passed (62 tests in 13 files; 4 new plan-panel tests).
- `npm run lint`: was not working (pre-existing; ESLint 9.39.5 found no `eslint.config.js`). Restored in the follow-up below.
- `scripts/build.ps1`: Passed. `npm run test:e2e` after the build: Passed (8 tests, including the new Stage 11 test: network blocked, `window.open` stubbed, asserts authorize URL/PKCE/loopback redirect, billing labels, no key field, cancel, zero external requests). Passed 5 more consecutive runs. One earlier run inside `test.ps1` had 3 failures (`toBeVisible`, 17.5 s vs. the usual ~6 s) that did not reproduce; cause not identified.
- Fixed `scripts/test.ps1`: final full run exited 0 with every step passing; a deliberately failing command was confirmed to stop it with a non-zero exit.
- `bin/quant-practice.exe -providers` (fresh data dir): prints the API-key table plus "ChatGPT Plan Accounts: (no accounts signed in)".
- **Not run:** any live sign-in, token exchange, inference, model listing or revocation against OpenAI. No eligible account was used, and Stage 11 live support is not claimed.

Lint restoration (October 5, 2026, follow-up session):
- Added [web/eslint.config.js](../web/eslint.config.js) (flat config: `@eslint/js` recommended + `@typescript-eslint` `flat/recommended`, browser globals, `_`-prefixed unused names allowed). Declared `@eslint/js` (^9.39.5, already installed transitively) as a direct dev dependency. Dropped the `--ext` flag, which flat config rejects.
- The first real lint run found 10 errors, all fixed without disabling rules: lexical declarations in a `switch` case in `App.tsx` (wrapped in a block); `any` replaced with `SubmittedAnswerDTO` in `PracticeDrill.tsx` draft/submit payloads, `ReturnType<typeof setTimeout>` for the debounce ref, `unknown` for `PublicSessionView.parameters`, an `instanceof Error` check in `MasteryView.tsx`, and `RequestInit` / `typeof fetch` in two test files.
- Lint added to [scripts/test.ps1](../scripts/test.ps1) and the CI workflow.
- `npm run lint` and `npm run typecheck`: clean. Full `scripts/test.ps1`: passed (scaffold, typecheck, lint, 62 unit tests, build, 8 browser tests, gofmt, go test, go vet).

Stage 11 live check and tutor rendering fix (October 5, 2026, same follow-up session):
- **Live (user-run, partial):** the user signed in with an eligible ChatGPT account and received a streamed tutor answer over the `chatgpt` plan route in the real app. `bin\quant-practice.exe -providers` afterwards lists the account as `ready`. Not yet confirmed live: sign-out/revocation. The model used was not recorded.
- The live answer exposed a rendering bug in [MathMarkdown.tsx](../web/src/components/MathMarkdown.tsx): a `$$` block spread over several lines was split into separate paragraphs, so MathJax could not match its delimiters, and only `#`/`##`/`### ` at column 0 became headings. The block parser now keeps multi-line `$$ … $$` and `\[ … \]` blocks in one element (an unterminated block stays plain text while streaming), handles `#`–`######` with up to three leading spaces, numbered lists, horizontal rules and fenced code blocks, protects `\( … \)` inline math, and HTML-escapes math text, which MathJax still reads as TeX. Styles for headings, lists, rules and code blocks were added to `index.css`.
- **Test isolation bug fixed:** the browser tests started the server with only `--db`, so the credential vault resolved to the user's real data directory. The Stage 11 test failed once the user had really signed in, and the Stage 10 test's fake Gemini key (`AIzaSy...9988`) is still in the user's real vault. All three server launches in `tests/e2e/foundation.spec.ts` now pass `--data-dir` to the temporary directory. The leftover fake key was not removed; it is the user's to remove in Settings.
- Added 4 MathMarkdown unit tests and 1 browser test that replays a live-shaped answer through the real panel and asserts real MathJax typesets the display block and headings.
- Full `scripts/test.ps1`: passed (lint clean, 66 unit tests, build, 9 browser tests, gofmt, go test, go vet).

Earlier Stage 10 record:
- `go test -v ./internal/auth/...`: Passed (4 comprehensive unit tests covering memory vault, key masking, AES file encryption, and DPAPI persistence).
- `go test -v ./internal/providers/...`: Passed (8 comprehensive unit tests covering budgets, catalog cache, Anthropic/Gemini/OpenAI mock streaming/discovery, provider manager, offline zero-network verification, and rate-limit backoff).
- `go test -v ./internal/httpapi/...`: Passed (15 unit tests covering provider and tutor endpoints with zero secret leakage).
- `go test -v ./internal/tutor/...`: Passed (10 unit tests).
- `go test -v ./internal/storage/...`: Passed (25 unit tests).
- `go test -v ./internal/mastery/...`: Passed (8 unit tests).
- `go test -v ./internal/drill/...`: Passed (11 unit tests).
- `go test -count=1 ./internal/...`: Passed (11/11 packages passed across assets, auth, bank, domain, drill, httpapi, mastery, mathengine, providers, storage, and tutor).
- `go vet ./internal/... ./cmd/...`: Passed (0 warnings).
- `gofmt -s -l cmd internal`: Passed (clean formatting).
- `python scripts/validate_scaffold.py`: Passed (34 Markdown files, 51 local links, 41 JSON files, 35 source directories; draft basics checked).
- `npm run typecheck` (in `web/`): Passed (strict TypeScript, 0 errors).
- `npm run test` (in `web/`): Passed (58 unit tests passed in Vitest across 12 files).
- `npm run test:e2e` (in `web/`): Passed (7 Playwright end-to-end browser tests passed with network disabled, verifying offline math/keyboard navigation, persistent drill recovery, Stage 06 navigation/draft preservation, Stage 07 multi-question navigation/reference library/settings/375px responsive math, Stage 08 concept mastery/scaffolds, Stage 09 read-only AI tutor streaming/leave protection/notes library/export, and Stage 10 provider settings/vault credential masking/AI tutor badge).
- `powershell -ExecutionPolicy Bypass -File scripts/test.ps1`: Passed (all verification checks passed end-to-end).
- `powershell -ExecutionPolicy Bypass -File scripts/build.ps1`: Passed (frontend assets bundled with local MathJax and standalone Go binary compiled to `bin\quant-practice.exe`).
- `.\bin\quant-practice.exe -providers`: Passed (prints formatted provider and vault credential table).

### Backlog Additions Recorded
- Formally scheduled four key features in `PLAN.md`, `REQUIREMENTS.md`, `docs/PROVIDERS.md`, `docs/AI-TUTOR.md`, and `docs/DECISIONS.md`:
  1. **Excel formula equivalents** (R29): Engine derivations, stage explanations, recaps, and reference library supply standard Excel functions (`=BINOM.DIST`, `=POISSON.DIST`, `=COMBIN`, `=NORM.DIST`).
  2. **Interactive multi-turn AI tutor chat** (R30): Conversational threads in SQLite, sliding-window budgets, strict answer-withholding guardrails, and flexible note saving.
  3. **ChatGPT plan sign-in** (R14, Stage 11 [IMPLEMENTED, live verification pending]): Verified local/open-source OAuth integration with PKCE `S256`, loopback redirect callback, dynamic registration, and plan Responses adapter.
  4. **Local offline LLM provider** (R31): OpenAI-compatible loopback adapter for LM Studio (`localhost:1234`) and Ollama (`localhost:11434`), zero external network calls, true offline generative AI.

## Next action

1. **Finish the live Stage 11 gate (needs the user):** sign-in and a live tutor answer are confirmed. Still needed: sign out from Settings > AI Providers > ChatGPT Plan and confirm the account disappears, and note the model used. Record the date, route, model and redacted outcome here, separately from mock checks. If OpenAI rejects the registration or any parameter, record the exact error and fix it against the docs; never borrow another app's client ID.
2. Commit the Stage 10 and Stage 11 work in reviewable commits once the user approves.
3. Then Stage 12 (creative candidate workflow) per [PLAN.md](../PLAN.md), or a backlog item if the user reprioritizes (the local LLM provider, R31, can reuse the existing SSE parsing). The curated API-key model defaults in `internal/providers/catalog.go` (Claude 3.5, Gemini 1.5, GPT-4o) are dated and should be refreshed against current catalogs.

## Unresolved external gates

Actual course syllabus/slides/notation; ChatGPT-plan sign-in eligible live account access; live provider API keys for optional non-mock testing; continuous distribution/test/regression course details.

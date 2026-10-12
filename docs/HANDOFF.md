# Handoff

## Fix GitHub CI verify build failure - October 12, 2026 (Windows)

- **Root Cause**: `internal/httpapi/arcade.go` was committed without a `package httpapi` declaration (starting directly with `import (`) and had not been wired into `internal/httpapi/server.go` (`ArcadeStore` field on `Config` and `Server`, `/api/arcade/scores` route registration).
- **Fixes Applied**:
  - Added `package httpapi` to [arcade.go](file:///c:/Users/Dan/quant-methods-practice/internal/httpapi/arcade.go).
  - Wired `ArcadeStore` into `Config` and `Server` in [server.go](file:///c:/Users/Dan/quant-methods-practice/internal/httpapi/server.go), bound it in `NewServer`, and registered `mux.HandleFunc("/api/arcade/scores", s.handleArcadeScores)`.
  - Added unit test suite in [arcade_test.go](file:///c:/Users/Dan/quant-methods-practice/internal/httpapi/arcade_test.go) covering GET/POST, demo score rejection, negative score rejection, malformed payload handling, and status codes.
- **Verification Run**:
  - `python scripts/validate_scaffold.py`: Scaffold OK.
  - `gofmt -d cmd internal`: Clean diff.
  - `npm run typecheck`, `npm run lint`, `npm run test` (19 test files, 100/100 passed), and `npm run build`: All passed cleanly.
  - `go vet ./internal/... ./cmd/...`: Clean.
  - `go test -v ./internal/... ./cmd/...`: All Go packages passed cleanly.
  - `go build ./cmd/quant-practice`: Compiled successfully.

## GitHub README Redesign & Demo Tour Embedding - October 12, 2026 (Windows)

- **README.md Overhaul**: Completely redesigned `README.md` into a modern, beautifully structured GitHub presentation document.
  - Centered hero section with dynamic SVG shields (Offline-First architecture, 137 questions, 22 objectives, Go numerics engine, React/TypeScript/Vite, 16/16 Playwright E2E passed).
  - Embedded high-definition walkthrough video (`recordings/quant_methods_demo_tour.webm`, 720p 60 FPS) directly near the top with fallback links and an Act 1–Act 3 chapter breakdown table.
  - Formatted feature cards detailing offline privacy, deterministic Go numerics, 7-stage pedagogical scaffolding, and keyboard fluency.
  - Complete curriculum coverage hierarchy and objective table spanning all 6 modules (137 approved templates).
  - Visuals Lab layout and guide covering Normal, Uniform, Exponential, Poisson, CLT, Confidence Intervals, Hypothesis Testing, and Regression.
  - Quick-start guide, CLI reference commands, complete keyboard shortcut matrix, ASCII system architecture diagram, test commands, and comprehensive 29-document documentation map.
- **Verification**: `python scripts/validate_scaffold.py` passes cleanly (34 Markdown files, 87 local links verified). `python scripts/check_content_coverage.py` passes across all 22 objectives. All 15 Go packages pass unit tests (`go test -count=1 ./internal/...`).

## Final Wave 4 + Wave 5 complete - October 11, 2026 (Windows)

Completed the remaining **20 questions** (7 from Wave 4, 13 from Wave 5), with **60 stages**, 20 matching review records and explicit user-delegated approval. The active bank is now **137 questions: 16 original + all 121 planned additions**, covering all **22 published objective targets**. Current approval timestamps are UTC October 12, corresponding to the October 11 client date in America/Chicago. The user's explicit delegation was: "Yes—delegate review and activate after checks pass." New records identify `codex_delegated_final_waves`; they do not claim personal user reading or rely on the earlier vague roadmap-authorization notes.

### Content and implementation

- Wave 4: `twosample-read-excel-output`, `regout-compare-models-fit`, `mreg-adjusted-r2-weak-predictor`, `mreg-overall-f-vs-individual-t`, `mreg-dummy-variable`, `regprob-nonlinearity`, `regprob-autocorrelation`.
- Wave 5: `desc-zscore-empirical-rule`, `desc-weighted-mean-price`, `desc-quartiles-iqr-fences`, `drv-compare-projects-risk`, `drv-insurance-break-even`, `drv-cumulative-from-table`, `lincomb-weekly-demand-sum-vs-scale`, `lincomb-portfolio-correlated`, `twosample-paired-training`, `twosample-independent-means-layouts`, `twosample-two-proportion-ads`, `regprob-outlier-influence`, `regprob-omitted-variable-causation`.
- New JSON lives under `curriculum/approved/` with matching `curriculum/reviews/*.review.json`. [Full worked preview](REMAINING-CONTENT-REVIEW.md) includes every scenario, assumption, choice, hint, key and explanation.
- `internal/mathengine/curriculum_calculations.go` adds reviewed rule-v2, input-only operations over the existing engines, inclusive interpolated quantiles, paired differences from raw arrays, signed directional p-values, strict PMF checks and pooled normal-approximation checks. `internal/bank/calculations.go`, the family registry, parser, optional domain fields and illustrative schema bind calculation stages to Go-derived quantities and reject key/input/label drift and colliding distractors. Original rule-v1 content is preserved; no historical learner snapshots are changed.
- `internal/mathengine/curriculum_calculations_test.go` uses hand-derived fixtures, including exact Cauchy directional tails. `internal/bank/calculations_test.go` grades every option in all 20 templates after snapshot serialization and checks unresolved public projections withhold calculation metadata and keys. Existing legacy family work from the prior session remains intact.
- `scripts/content-audit/main.go` validates final-sprint content and regenerates its complete preview; `-reviews` exports matching already-approved review records and refuses different overwrites. It never grants approval. `scripts/check_content_coverage.py` checks every catalog slug, all 22 objective counts, review identity/status/version and the 137 unique bank IDs; older review formats are accepted only for original v1 templates.
- `tests/e2e/worksheets.spec.ts` now exercises all 20 new questions and 60 stages through the native offline browser UI, checks persisted correct attempts and MathJax errors, and verifies quartile/portfolio content fits at 375px width. Test databases and credential stores are isolated temporary directories.
- Updated README, PLAN, CONTENT-BACKLOG, QUESTION-BANK and NUMERICS. Applied Go formatting; pre-existing extra blank lines and CRLF formatting were normalized as required by the repository check, without unrelated logic changes.

### Verification actually run

- `go test -count=1 ./internal/mathengine ./internal/bank`: passed, including independent fixtures, invalid-input checks, wrong-direction tails, parameter/key drift rejection, every distractor grade and snapshot round-trips.
- `go run ./scripts/content-audit -derive` while still unapproved drafts, then `go run ./scripts/content-audit -dir curriculum/approved -reviewer codex_delegated_final_waves -reviews`: all 20 passed; review records contain canonical engine quantities rather than invented probability/moment placeholders.
- `powershell -ExecutionPolicy Bypass -File scripts/test.ps1`: passed after allowing subprocess launch outside the restricted sandbox. Includes scaffold validation, TypeScript, lint, **100/100 frontend unit tests**, production UI/native executable build, **15/15 existing browser tests**, Go formatting, all 15 internal package tests and vet. The first restricted attempt stopped at Vitest/esbuild `spawn EPERM`; it did not indicate a test assertion failure.
- `npm run test:e2e -- worksheets.spec.ts` after adding the focused content check: **3/3 passed**, including all 60 new stages, no external requests or MathJax errors. One initial test locator timed out because an exact implicit-label lookup did not match the label containing a select; corrected to scope the select within its visible label and reran successfully.
- `python scripts/check_content_coverage.py`: passed - **121/121 additions, 22/22 objectives, 137 unique active templates**.
- Final complete `scripts/test.ps1` rerun after adding the new browser test and correcting approval text encoding: **passed with 100 frontend tests and all 16 browser tests**, native rebuild, Go formatting/tests/vet. `go test ./scripts/content-audit` compiles the audit command (no test files); `go vet ./scripts/content-audit` passes. Repeated approved review export passes with CRLF/LF normalization. CLI `bin/quant-practice.exe -data-dir .cache/final-bank-smoke -list-bank` loads the final bank from the rebuilt executable in isolated local data. Generated changes to existing screenshot fixtures were removed; the new browser assertions remain.

### Remaining scope and next work

No catalog questions remain to author. Inclusive quartiles, pooled versus Welch and dummy-variable assessment coverage remain unconfirmed course conventions, explicitly labeled in the affected exercises. No private course files were supplied (inventory remains `.gitkeep` only), and no new claim of instructor-specific alignment, live provider verification or learning effectiveness is made.

The local content changes are not committed. Per the backlog's delivery rule, the final 20 checklist boxes remain open until their files are committed; authoring, delegated approval, activation and validation are complete. Preserve all existing local sprint work when preparing that commit. Stage 18 arcade parity and Stage 19 release hardening remain the next implementation stages; Stage 20 ongoing transfer-quality review is not finished merely by meeting counts.

---

## Dispersed Wave 4 Sprint (13 Questions Completed + All 4 Engines Implemented — Bank Reaches 117 Questions!) — October 11, 2026 (Windows)

Successfully implemented all remaining engine prerequisites, authored, audited, and approved **13 new questions** in **Dispersed Wave 4**, expanding the active approved repository from 104 to **117 approved questions** (16 baseline + 101 new), and achieving **100% completion of Row 1.3 (6/6) and Row 5.3 (6/6)**!

- **Engine Prerequisites Implemented & Verified**:
  - `ENG-DESC` (`descriptive_stats`): `internal/mathengine/descriptive.go` with mean, median, mode, range, sample/population variance and standard deviation, coefficient of variation ($CV$), $z$-scores, and weighted means. Unit tests in `internal/mathengine/descriptive_test.go`.
  - `ENG-PMF` (`discrete_rv`): `internal/mathengine/discrete_rv.go` with discrete probability distribution validation, solving missing probabilities, expected value ($E[X]$), second moment ($E[X^2]$), variance ($\mathrm{Var}(X) = E[X^2] - \mu^2$), standard deviation, and interval/tail probabilities. Unit tests in `internal/mathengine/discrete_rv_test.go`.
  - `ENG-AFFINE` (`linear_combination` extension): `internal/mathengine/linearcomb.go` extended with affine transformations ($E[aX+c] = aE[X]+c, \mathrm{Var}(aX+c) = a^2\mathrm{Var}(X), \mathrm{SD}(aX+c) = |a|\mathrm{SD}(X)$) where constant $c$ contributes zero variance, plus sum and difference of independent random variables ($\mathrm{Var}(X \pm Y) = \mathrm{Var}(X) + \mathrm{Var}(Y)$). Unit tests in `internal/mathengine/linearcomb_test.go`.
  - `ENG-TWOSAMPLE` (`two_sample_test`): `internal/mathengine/two_sample.go` with paired samples $t$-tests ($df = n - 1$), independent two-sample pooled $t$-tests ($df = n_1 + n_2 - 2$), and two-sample pooled proportion $z$-tests. Unit tests in `internal/mathengine/two_sample_test.go`.
  - `ENG-PROB-EXT` (`set_probability` extension): `internal/mathengine/setprob.go` extended with Law of Total Probability and Bayes' Theorem for 2-partitions. Unit tests in `internal/mathengine/setprob_test.go`.
  - Registered all new families and parameter rules in `internal/bank/family.go` and verified in `internal/bank/bank_test.go`.

- **Module 1 Completed in Wave 4 (4 questions — Row 1.3 reaches 6/6 milestone!)**:
  - `curriculum/approved/desc-mean-median-outlier.json` + review: monthly sales of $7$ retail branches ($\{18, 20, 22, 24, 25, 26, 120\}$ in thousands), computing mean $\bar{x} = 255/7 \approx \$36.43\text{k}$ vs median $\$24.00\text{k}$, showing how the $\$120\text{k}$ flagship outlier inflates the mean above $85.7\%$ of branches, and establishing the median as the robust representative measure.
  - `curriculum/approved/desc-sample-sd-delivery-times.json` + review: $6$ courier delivery times ($\{12, 16, 18, 20, 22, 32\}$ min), computing sum of squared deviations $SS = 232.0$, sample variance $s^2 = 232/5 = 46.40$, sample SD $s = \sqrt{46.4} \approx 6.81\text{ min}$, and contrasting with population SD $\sigma = \sqrt{232/6} \approx 6.22\text{ min}$. Excel: `=STDEV.S` vs `=STDEV.P` and Bessel's correction ($n - 1$).
  - `curriculum/approved/desc-cv-compare-stocks.json` + review: comparing Tech Growth Fund A ($\bar{x}_A = 12.0\%, s_A = 4.0\%$) vs Stable Income Fund B ($\bar{x}_B = 4.0\%, s_B = 2.0\%$), computing $CV_A = 4/12 = 33.3\%$ vs $CV_B = 2/4 = 50.0\%$, and showing Fund B carries higher relative risk per unit of return despite having half the absolute standard deviation.
  - `curriculum/approved/prob-two-plant-defect-tree.json` + review: automotive microchips from Plant 1 ($60\%$, $2\%$ defect) and Plant 2 ($40\%$, $5\%$ defect), applying Total Probability $P(D) = 0.60(0.02) + 0.40(0.05) = 0.032$ ($3.2\%$) and Bayes' Theorem $P(\text{Plant 2} \mid D) = 0.020 / 0.032 = 0.625$ ($62.5\%$), showing how prior $40\%$ updates to $62.5\%$ posterior. **Completes Row 1.3 at 6/6!**

- **Module 2 Completed in Wave 4 (7 questions — Rows 2.1 at 3/6, Row 2.2 at 4/6)**:
  - `curriculum/approved/drv-valid-pmf-missing-value.json` + review: dealership daily car sales $X \in \{0, 1, 2, 3, 4\}$, solving missing $P(X = 3) = 1.00 - 0.80 = 0.20$, verifying PMF validity axioms ($0 \le P(x) \le 1, \sum P(x) = 1$), and computing tail $P(X \ge 2) = 0.35 + 0.20 + 0.10 = 0.65$ ($65\%$).
  - `curriculum/approved/drv-expected-cars-sold.json` + review: dealership daily sales, computing $E[X] = \mu = 0(0.10) + 1(0.25) + 2(0.35) + 3(0.20) + 4(0.10) = 1.95\text{ cars/day}$, identifying Excel formula `=SUMPRODUCT(A2:A6, B2:B6)`, and interpreting long-run daily sales expectation.
  - `curriculum/approved/drv-variance-cars-sold.json` + review: dealership sales, computing second moment $E[X^2] = 5.05$, variance $\mathrm{Var}(X) = 5.05 - (1.95)^2 = 1.2475\text{ cars}^2$, standard deviation $\sigma = \sqrt{1.2475} \approx 1.1169\text{ cars}$, and dispelling the common trap of forgetting to subtract $\mu^2$.
  - `curriculum/approved/affine-profit-from-units.json` + review: bakery custom cake sales ($E[X] = 30, \mathrm{SD}(X) = 6$), profit equation $Y = 50X - 200$, computing $E[Y] = 50(30) - 200 = \$1{,}300$, variance $\mathrm{Var}(Y) = 50^2(36) = 90{,}000$, and standard deviation $\mathrm{SD}(Y) = 50 \times 6 = \$300$, proving fixed cost $-\$200$ shifts the mean but has zero effect on variance.
  - `curriculum/approved/affine-currency-conversion.json` + review: European branch expenses $E[X] = 40{,}000\text{ EUR}, \mathrm{SD}(X) = 5{,}000\text{ EUR}$, converted to USD at $1.10$, computing $E[Y] = \$44{,}000$, $\mathrm{SD}(Y) = \$5{,}500$, and $\mathrm{Var}(Y) = 1.10^2(25\text{M}) = 30{,}250{,}000\text{ dollars}^2$, contrasting linear SD scaling ($1.10\times$) with quadratic variance scaling ($1.21\times$).
  - `curriculum/approved/lincomb-total-two-product-lines.json` + review: independent daily revenues Line A ($E = \$12\text{k}, \mathrm{SD} = \$1.5\text{k}$) and Line B ($E = \$8\text{k}, \mathrm{SD} = \$2.0\text{k}$), computing expected total $E[T] = \$20{,}000$, variance $\mathrm{Var}(T) = 2.25\text{M} + 4.0\text{M} = 6.25\text{M}$, and $\mathrm{SD}(T) = \sqrt{6.25\text{M}} = \$2{,}500$, disproving naive addition of standard deviations ($\$3{,}500$).
  - `curriculum/approved/lincomb-difference-independent.json` + review: independent customer traffic Store 1 ($E = 150, \mathrm{Var} = 160$) and Store 2 ($E = 120, \mathrm{Var} = 90$), expected difference $E[D] = 150 - 120 = 30$, variance $\mathrm{Var}(D) = 160 + 90 = 250$ (never $160 - 90 = 70$), and standard deviation $\mathrm{SD}(D) = \sqrt{250} \approx 15.81$, establishing why variances always add for independent variables.

- **Module 5 Completed in Wave 4 (2 questions — Row 5.3 reaches 6/6 milestone!)**:
  - `curriculum/approved/hyp-excel-tail-function.json` + review: selecting Excel functions for $t$-tests ($df = 24$): right-tailed ($H_1: \mu > \mu_0$) uses `=T.DIST.RT(2.15, 24)`; left-tailed ($H_1: \mu < \mu_0$) uses `=T.DIST(-2.15, 24, TRUE)`; two-tailed ($H_1: \mu \ne \mu_0$) uses `=T.DIST.2T(ABS(t), 24)` and explaining the $x > 0$ requirement. **Completes Row 5.3 at 6/6!**
  - `curriculum/approved/twosample-choose-design.json` + review: classifying experimental setups: before/after sales training on same reps (paired $t$-test, $df = 24$), open office vs cubicles on separate workers (independent two-sample $t$-test, $df = 58$), support resolution times vs $15.0\text{ min}$ SLA benchmark (one-sample $t$-test, $df = 39$), and explaining how paired designs control for subject heterogeneity to maximize power.

- **Verification Status**:
  - `python scripts/validate_scaffold.py`: Pass (0 errors across 257 JSON files, 33 Markdown files, 84 local links).
  - `go test -count=1 ./...`: Pass across all 15 packages.
  - Active curriculum bank: Loads cleanly with **117 approved records** (16 baseline + 101 new).
  - Vitest: 19 test files pass, 100/100 tests pass.
  - Total remaining across entire 121-question backlog: **20 questions** (7 in Wave 4, 13 in Wave 5).

- **Next Concrete Target (When Resuming)**:
  - Finish remaining 7 questions of Wave 4:
    - Row 5.4: `twosample-read-excel-output`
    - Row 6.2: `regout-compare-models-fit` (completes Row 6.2 at 5/5)
    - Row 6.4: `mreg-adjusted-r2-weak-predictor`, `mreg-overall-f-vs-individual-t`, `mreg-dummy-variable` (completes Row 6.4 at 6/6)
    - Row 6.5: `regprob-nonlinearity`, `regprob-autocorrelation`

---

## Dispersed Wave 3 Sprint (All 20 Questions Completed — Bank Hits 104 Questions!) — October 11, 2026 (Windows)


Successfully authored, audited, and added all **20 questions** in **Dispersed Wave 3**, expanding the active approved repository from 84 to **104 approved questions** (16 baseline + 88 new), achieving **100% completion of Module 4 (22/22 questions)** and completing Rows 1.1, 4.3, 4.4, 5.1, 5.2, 6.1, and 6.3!

- **Module 4 Completed in Wave 3 (3/3 questions — Module 4 reaches 22/22 milestone!)**:
  - `curriculum/approved/ci-level-vs-width.json` + review: silicon wafer thickness ($n = 36, SE = 2.0\\ \\mu\\text{m}$), contrasting 90% ($6.58\\ \\mu\\text{m}$), 95% ($7.84\\ \\mu\\text{m}$), and 99% ($10.30\\ \\mu\\text{m}$) intervals, proving that higher confidence requires wider intervals (the precision vs certainty trade-off), and demonstrating that increasing sample size $n$ is the sole way to narrow width without reducing confidence.
  - `curriculum/approved/ci-plausible-claim.json` + review: cloud latency $[146.6\\text{ ms}, 153.4\\text{ ms}]$, evaluating vendor claim $\\mu_0 = 155.0\\text{ ms}$ (strictly outside, not plausible, rejected at $\\alpha = 0.05$ via duality) vs internal target $\\mu_0 = 150.0\\text{ ms}$ (inside, plausible, fail to reject at $\\alpha = 0.05$).
  - `curriculum/approved/n-round-up-misconception.json` + review: hospital wait times survey, showing continuous formula $n = 96.04$ must round up to ceiling integer $n = 97$, proving why rounding down to $96$ gives $E = 5.0010 > 5.0\\text{ min}$ violating target tolerance, and establishing the universal ceiling rule. Excel: `=ROUNDUP((1.96*25/5)^2, 0)`.

- **Module 5 Completed in Wave 3 (6/6 questions — Rows 5.1 & 5.2 reach 5/5 each!)**:
  - `curriculum/approved/hyp-formulate-two-tailed-calibration.json` + review: CNC bushing milling target $\\mu_0 = 2.000\\text{ cm}$, formulating $H_0: \\mu = 2.000$ vs $H_1: \\mu \\ne 2.000$, establishing two-tailed rejection region split $\\alpha/2 = 0.025$ in each tail, and setting the burden of proof.
  - `curriculum/approved/hyp-formulate-proportion.json` + review: juice bottle preference test $p_0 = 0.40$, directional claim formulating $H_0: p \\le 0.40$ vs $H_1: p > 0.40$, upper right-tailed rejection region, and dispelling confusion between population parameter $p$ and sample statistic $\\hat{p}$.
  - `curriculum/approved/errors-larger-n-reduces-beta.json` + review: turbine fasteners tensile test, proving increasing $n$ from $25$ to $100$ shrinks $SE$, separating the null and alternative distributions, driving $\\beta$ down from $0.28$ to $< 0.05$ and boosting power $1 - \\beta > 0.95$ at fixed $\\alpha = 0.05$.
  - `curriculum/approved/errors-identify-after-the-fact.json` + review: retrospective audit of fraud suspension (rejected true $H_0 \\implies$ Type I error / false alarm) vs contaminated vaccine release (failed to reject false $H_0 \\implies$ Type II error / missed defect).
  - `curriculum/approved/hyp-pvalue-interpretation.json` + review: e-commerce A/B checkout test ($p = 0.030$), establishing the formal frequentist definition $P(\\text{data at least as extreme} \\mid H_0)$, and debunking both the inverse probability fallacy ($P(H_0) = 0.03$) and the chance fallacy.
  - `curriculum/approved/hyp-alpha-changes-decision.json` + review: clinical blood pressure trial ($p = 0.032$), demonstrating reject $H_0$ at $\\alpha = 0.05$ ($0.032 \\le 0.05$) vs fail to reject at $\\alpha = 0.01$ ($0.032 > 0.01$), and establishing why $\\alpha$ must be specified *a priori* to prevent $p$-hacking.

- **Module 6 Completed in Wave 3 (8/8 questions — Rows 6.1 & 6.3 reach 5/5 each!)**:
  - `curriculum/approved/reg-extrapolation-warning.json` + review: chemical yield regression ($20^\\circ\\text{C} \\le x \\le 60^\\circ\\text{C}$), critiquing prediction at $x = 120^\\circ\\text{C}$ as severe extrapolation, explaining non-linear curvature and expanding prediction variance.
  - `curriculum/approved/reg-predict-from-excel-coefficients.json` + review: delivery van maintenance ($n = 24, k = 1$), reading Intercept $b_0 = 20.000$ and Mileage slope $b_1 = 1.200$, model equation $\\hat{y} = 20.0 + 1.2x$, predicting $\\hat{y} = 56.0\\text{ hundred dollars}$ ($\\\\$5{,}600$) at $x = 30.0\\text{k miles}$, and interpreting incremental cost of $\\\\$120$ per $1{,}000$ miles.
  - `curriculum/approved/regout-significance-f.json` + review: commercial loan default ANOVA ($n = 26, k = 1, F = 16.00$), defining Significance $F = 0.0005$ as the model $p$-value, rejecting $H_0$ at $\\alpha = 0.05$, and linking $F = t^2 = 4^2 = 16$ with identical $p$-values. Excel: `=F.DIST.RT(16, 1, 24)`.
  - `curriculum/approved/regout-r-from-r2-sign.json` + review: used car depreciation ($n = 11, k = 1, R^2 = 0.640$, slope $b_1 = -2.000 < 0$), deriving $r = -\\sqrt{0.640} = -0.800$, and explaining why $R^2$ is strictly non-negative while $r$ inherits the sign of the slope.
  - `curriculum/approved/regout-t-squared-equals-f.json` + review: ER nurse staffing regression ($n = 22, k = 1, df = 20$), verifying identity $t_1^2 = 3.000^2 = 9.000 = F$, identical two-tailed $p$-values ($0.0071$), and explaining why this identity is restricted exclusively to single-predictor models ($k = 1$).
  - `curriculum/approved/regout-compute-slope-ci.json` + review: commercial real estate ($n = 27, k = 1, df = 25$), slope $b_1 = 120.000, s_{b_1} = 15.000$, critical $t_{0.025, 25} = 2.0595$, margin of error $E = 30.893$, $95\\%$ CI $[89.107, 150.893]$ strictly positive, and Excel formula `=120.0 - T.INV.2T(0.05, 25)*15.0`.
  - `curriculum/approved/mreg-interpret-coefficient-ceteris-paribus.json` + review: residential property model ($n = 35, k = 3$), interpreting commute distance slope $b_2 = -12.0$ as an estimated $\\\\$12{,}000$ price reduction per mile strictly holding square footage and school rating constant (*ceteris paribus*), and explaining omitted variable bias.
  - `curriculum/approved/regprob-heteroscedasticity.json` + review: corporate R&D vs revenue model ($n = 60$), diagnosing fan-shaped residual pattern as heteroscedasticity (violation of constant error variance), analyzing unbiased point estimates vs biased standard errors, and prescribing log-log transformations or robust standard errors.

- **Module 1 Completed in Wave 3 (3/3 questions — Row 1.1 reaches 5/5!)**:
  - `curriculum/approved/concept-cross-sectional-vs-time-series.json` + review: retail enterprise BI comparing 50-store Q4 snapshot (cross-sectional) vs single-store 36-month tracking (time-series), analyzing row reordering invariance vs temporal autocorrelation, and matching research questions.
  - `curriculum/approved/prob-contingency-age-purchase.json` + review: promotional campaign contingency table ($n = 500$), computing marginal $P(B) = 180/500 = 0.360$, joint $P(A_1 \\cap B) = 60/500 = 0.120$, conditional $P(B \\mid A_1) = 60/200 = 0.300$, and proving statistical dependence ($P(B \\mid A_1) \\ne P(B)$).
  - `curriculum/approved/prob-disjoint-vs-independent.json` + review: checkout payment methods ($P(A) = 0.60, P(B) = 0.30$), proving mathematically that mutually exclusive events ($P(A \\cap B) = 0$) cannot be independent because $0 \\ne P(A)P(B) = 0.180$, and showing conditional probability $P(B \\mid A) = 0 \\ne 0.30$ represents maximum dependence.

- **Verification Status**:
  - `python scripts/validate_scaffold.py`: Pass (0 errors across 231 JSON files, 33 Markdown files, 84 local links).
  - `go test -count=1 ./...`: Pass across all 15 packages.
  - Active curriculum bank: Loads cleanly with **104 approved records** (16 baseline + 88 new).
  - Vitest: 19 test files pass, 100/100 tests pass.

- **Next Concrete Target: Dispersed Wave 4 (20 Questions)**:
  - Prerequisite Engines: ENG-TWOSAMPLE (`two_sample_test`), ENG-DESC (`descriptive_stats`), ENG-PMF (`discrete_rv`), ENG-AFFINE (`linear_combination` extension).
  - Wave 4 items will complete Rows 5.3, 6.2, 6.4, and 1.3!

---

## Dispersed Wave 2 Sprint (All 20 Questions Completed) — October 11, 2026 (Windows)

Successfully authored, reviewed, and validated all **20 questions** in **Dispersed Wave 2**, expanding the active approved question bank from 64 to **84 approved questions** (16 baseline + 68 new) across all modules:

- **Module 6 Completed in Wave 2 (6/6 questions — Rows 6.1 at 3/5, Row 6.2 at 2/5, Row 6.3 at 3/5, Row 6.4 at 2/6)**:
  - `curriculum/approved/reg-interpret-slope-intercept.json` + review: commercial electricity cost model $\hat{y} = 450 + 38.5x$ ($42^\circ\text{F} \le x \le 94^\circ\text{F}$), interpreting slope as $\$38.50$ per $1^\circ\text{F}$ increase, and critiquing the intercept $\$450$ as mathematical line placement that should not be given physical meaning because $0^\circ\text{F}$ is far outside the observed sample range.
  - `curriculum/approved/reg-residual-observed-point.json` + review: retail advertising regression ($n = 5$), at Store 3 ($x = 20, y = 60$), predicting $\hat{y} = 17.2 + 2.24(20) = 62.0$, and computing residual $e = y - \hat{y} = 60.0 - 62.0 = -2.0$ (underperformed model benchmark by $\$2,000$).
  - `curriculum/approved/regout-se-from-sse.json` + review: ANOVA maintenance overhead model ($n = 20, k = 1, SSR = 1024.0, SSE = 288.0, SST = 1312.0$), residual $df_{\text{res}} = 18$, $MSE = 288/18 = 16.0$, standard error of estimate $s_e = \sqrt{16.0} = 4.0$ ($\$4,000$ typical prediction dispersion), and verifying $(b_1/s_{b_1})^2 = 8^2 = 64.0 = F$ consistency.
  - `curriculum/approved/regout-slope-ci-includes-zero.json` + review: express freight delivery duration regression ($n = 20, k = 1, df = 18$), cargo weight $b_1 = 0.600, s_{b_1} = 0.500$, critical $t_{0.025, 18} \approx 2.1009$, $95\%$ confidence interval $[-0.450, +1.650]$ containing zero, testing duality showing $t = 1.20, p = 0.2457 > 0.05$, and failing to reject $H_0: \beta_1 = 0$ (not statistically significant). Excel: `=T.INV.2T(0.05, 18)`.
  - `curriculum/approved/regout-slope-ci-interpretation.json` + review: commercial solar power generation ($n = 22, k = 1, df = 20$), sunlight hours $b_1 = 35.000\text{ kWh}, s_{b_1} = 4.000$, $95\%$ confidence interval $[26.656, 43.344]$ strictly positive and excluding zero ($t = 8.75, p < 0.0001$), establishing frequentist parameter interpretation and Excel formula `=35.0 - T.INV.2T(0.05, 20)*4.0`.
  - `curriculum/approved/mreg-write-equation-predict.json` + review: retail sales volume multiple regression ($n = 30, k = 3, df_{\text{res}} = 26$), model equation $\hat{Y} = 50.0 + 4.2 X_1 - 2.5 X_2 + 0.8 X_3$, point forecast at $X_1 = 10, X_2 = 20, X_3 = 25$ yielding $\hat{Y} = 62.0\text{ hundred units}$ ($6{,}200$ units), and partial slope interpretation under *ceteris paribus*.

- **Module 1 Completed in Wave 2 (3/3 questions — Row 1.1 at 4/5, Row 1.3 at 3/6)**:
  - `curriculum/approved/concept-sampling-bias-online-poll.json` + review: retail loyalty app voluntary receipt poll ($n = 12{,}000$, $88\%$ in favor), diagnosing voluntary response bias and undercoverage of non-app customers, proving that large sample size cannot cure systematic bias, and prescribing probability sampling.
  - `curriculum/approved/concept-data-types-business.json` + review: e-commerce vendor data dictionary classifying `Store_Region` (categorical nominal), `Satisfaction_Rating` (categorical ordinal), `Units_Returned` (numerical discrete count, ratio), and `Monthly_Revenue` (numerical continuous monetary flow, ratio).
  - `curriculum/approved/prob-independence-check.json` + review: cross-channel conversion tracking with marginals $P(A) = 0.40, P(B) = 0.50$, testing multiplication rule $P(A \cap B) = 0.20 = 0.40 \times 0.50$ (Cohort 1 independent with $P(A \mid B) = P(A) = 0.40$) vs Cohort 2 $P(A \cap B) = 0.25 \ne 0.20$ (dependent, positive synergy $P(A \mid B) = 0.50 > 0.40$), and dispelling confusion with mutual exclusivity.

- **Module 2 Completed in Wave 2 (2/2 questions — Row 2.3 completes at 5/5!)**:
  - `curriculum/approved/binomial-mean-sd-shipments.json` + review: microcontroller receiving inspection ($n = 200, p = 0.02$), computing expected defects $E[X] = \mu = 200 \times 0.02 = 4.0$ units, variance $\sigma^2 = 3.92$, standard deviation $\sigma = \sqrt{3.92} \approx 1.98$ units, avoiding Poisson $\sqrt{np}$ distractor, and interpreting lot-to-lot dispersion. Excel: `=SQRT(200*0.02*(1-0.02))`.
  - `curriculum/approved/binomial-condition-violated-small-lot.json` + review: aerospace fuel valves lot inspection ($N = 20, M = 4, n = 5$ without replacement), proving violation of independent trials and constant $p$ due to excessive sampling fraction $n/N = 25\% > 5\%$, identifying exact Hypergeometric model ($P(X = 1) = 0.4696$ vs Binomial $0.4096$), and completing Row 2.3 at 5/5 approved questions.

- **Module 4 Completed in Wave 2 (4/4 questions — Rows 4.3 & 4.4 at 4/5 each)**:
  - `curriculum/approved/ci-mean-t-small-sample-90.json` + review: machine cycle times $n = 10, \bar{x} = 42.3\text{ min}, s = 5.1\text{ min}$, degrees of freedom $df = 9$, $SE = 5.1/\sqrt{10} \approx 1.6128\text{ min}$, critical $t_{0.05, 9} = 1.8331$, margin of error $E = 2.96\text{ min}$, and $90\%$ confidence interval $[39.34, 45.26]\text{ min}$. Excel: `=CONFIDENCE.T(0.10, 5.1, 10)`.
  - `curriculum/approved/ci-interpretation-statement.json` + review: streaming subscription spending $[\\$45.20, \\$52.80]$, evaluating frequentist interpretation (long-run capture frequency over repeated independent samples) and dispelling the widespread "$P(45.20 \le \mu \le 52.80) = 0.95$" misconception regarding a fixed population parameter.
  - `curriculum/approved/n-proportion-pilot.json` + review: telecom churn survey with pilot estimate $p^* = 0.20$, target $E = 0.03$ at $95\%$ confidence, unrounded $n = 682.93 \implies$ ceiling integer $n = 683$, saving 385 respondents compared to conservative planning $p^* = 0.50$ ($n = 1{,}068$). Excel: `=ROUNDUP((NORM.S.INV(0.975)^2 * 0.20 * 0.80)/(0.03^2), 0)`.
  - `curriculum/approved/n-halve-margin-quadruple.json` + review: healthcare wait times survey, proving the inverse-square precision law $n \propto 1/E^2$—halving margin of error ($E/2$) requires quadrupling data from $n = 100$ to $n = 400$, derived from the square-root decay of standard error ($SE = \sigma/\sqrt{n}$).

- **Module 5 Completed in Wave 2 (5/5 questions — Rows 5.1 & 5.2 at 3/5, Row 5.3 at 3/6)**:
  - `curriculum/approved/hyp-formulate-right-tailed-sales.json` + review: retail endcap display raising average daily sales above baseline $\mu_0 = 50$, formulating $H_0: \mu \le 50$ (or $\mu = 50$) vs directional alternative $H_1: \mu > 50$, and identifying the upper right-tailed rejection region.
  - `curriculum/approved/hyp-steps-and-conclusion.json` + review: loan origination processing audit ($p = 0.18 > 0.05$), clarifying why we "fail to reject $H_0$" rather than "accept $H_0$ as proven fact," and verifying the standard 5-step hypothesis testing sequence.
  - `curriculum/approved/errors-which-is-worse.json` + review: contrasting baby formula pathogen testing ($H_0: \text{safe}$ vs $H_1: \text{contaminated}$) with marketing ad testing to show how catastrophic Type II error (shipping contaminated product) drives test power ($1 - \beta \ge 0.99$), larger sample sizes, and adjusted $\alpha$.
  - `curriculum/approved/errors-alpha-beta-tradeoff.json` + review: cleanroom wafer resistance testing, evaluating the inescapable inverse trade-off where lowering $\alpha$ from $0.05$ to $0.01$ with fixed $n = 50$ increases $\beta$ and lowers power, and confirming that increasing sample size $n$ is the only method to reduce both $\alpha$ and $\beta$ simultaneously.
  - `curriculum/approved/hyp-proportion-left-tailed.json` + review: vendor switch defect inspection $p_0 = 0.20, n = 200, x = 28 \implies \hat{p} = 0.14$, standard error under the null $SE_0 = \sqrt{0.20(0.80)/200} \approx 0.02828$, test statistic $Z = -2.1213$, left-tailed $p$-value $= 0.01695 \approx 0.0170 < 0.05$, rejecting $H_0$. Excel: `=NORM.S.DIST((0.14 - 0.20)/SQRT(0.20*0.80/200), TRUE)`.

- **Engine Extensions**:
  - `SetProbabilityFamily`: Registered `"independence_check"` operation and required `p_intersection` validation.
  - `BinomialFamily`: Extended `ValidateParameters` to admit `operation: "mean_sd"` along with `mean`, `sd`, and `var` validation against theoretical moments ($E[X] = np, \sigma = \sqrt{np(1-p)}$).

- **Verification Status**:
  - `python scripts/validate_scaffold.py`: Pass (0 errors across 191 JSON files, 33 Markdown files, 84 local links).
  - `go test -count=1 ./...`: Pass across all 15 packages.
  - Active curriculum bank: Loads cleanly with **84 approved records** (16 baseline + 68 new).
  - Vitest: 19 test files pass, 100/100 tests pass.

- **Next Concrete Target: Dispersed Wave 3 (20 Questions)**:
  - Module 4 (3 questions — completes Module 4 at 22/22): `ci-level-vs-width`, `ci-plausible-claim`, `n-round-up-misconception`
  - Module 5 (6 questions — completes Rows 5.1 & 5.2 at 5/5): `hyp-formulate-two-tailed-calibration`, `hyp-formulate-proportion`, `errors-larger-n-reduces-beta`, `errors-identify-after-the-fact`, `hyp-pvalue-interpretation`, `hyp-alpha-changes-decision`
  - Module 6 (8 questions — completes Rows 6.1 & 6.3 at 5/5): `reg-extrapolation-warning`, `reg-predict-from-excel-coefficients`, `regout-significance-f`, `regout-r-from-r2-sign`, `regout-t-squared-equals-f`, `regout-compute-slope-ci`, `mreg-interpret-coefficient-ceteris-paribus`, `regprob-heteroscedasticity`
  - Module 1 (3 questions — completes Row 1.1 at 5/5): `concept-cross-sectional-vs-time-series`, `prob-contingency-age-purchase`, `prob-disjoint-vs-independent`.

---

Successfully authored, reviewed, and validated all 20 questions in **Dispersed Wave 1**, expanding the active approved question bank from 44 to **64 approved questions** (16 baseline + 48 new):

- **Module 4 (4 questions)**:
  - `curriculum/approved/ci-mean-z-known-sigma.json` + review: polymer tensile strength $\sigma = 12.0\text{ MPa}, n = 49, \bar{x} = 150.0\text{ MPa}$, standard error $SE = 12/7 \approx 1.7143\text{ MPa}$, critical value $Z^* = 1.960$, margin of error $E = 3.360\text{ MPa}$, and $95\%$ interval $[146.64, 153.36]\text{ MPa}$. Excel: `=CONFIDENCE.NORM(0.05, 12, 49)`.
  - `curriculum/approved/ci-proportion-recommend.json` + review: customer recommendation survey $x = 132, n = 400 \implies \hat{p} = 0.330$, normality verification $n\hat{p} = 132 \ge 10$ and $n(1-\hat{p}) = 268 \ge 10$, $SE_{\hat{p}} = \sqrt{0.33(0.67)/400} \approx 0.02351$, margin of error $E = 0.0461$ ($4.61\%$), and $95\%$ interval $[0.2839, 0.3761]$ ($28.4\%$ to $37.6\%$).
  - `curriculum/approved/n-mean-margin-five.json` + review: urgent care wait times $\sigma = 25.0\text{ min}$, target margin $E \le 5.0\text{ min}$ at $95\%$ confidence, continuous bound $(1.95996 \times 25 / 5)^2 \approx 96.0365 \implies$ ceiling integer $n = 97$, proving why rounding down to $96$ gives $E \approx 5.0009 > 5.0$, and demonstrating the quadratic precision law ($2\times$ tighter margin requires $4\times$ data to $n = 385$). Excel: `=ROUNDUP((NORM.S.INV(0.975)*25/5)^2, 0)`.
  - `curriculum/approved/n-proportion-conservative.json` + review: statewide voter polling with no prior benchmark, conservative $p^* = 0.50$ maximizing $p(1-p) = 0.25$, target $E = 0.04$ at $90\%$ confidence ($Z^* = 1.645$), continuous bound $422.741 \implies$ ceiling integer $n = 423$. Contrasted with prior study $p \approx 0.20$ yielding $n = 271$ (saving 152 respondents). Excel: `=ROUNDUP((NORM.S.INV(0.95)^2)*0.25/(0.04^2), 0)`.

- **Module 5 (4 questions)**:
  - `curriculum/approved/hyp-formulate-underfill.json` + review: olive oil 16 oz labeled containers with suspected underfill, formulating alternative $H_1: \mu < 16.0\text{ fl oz}$ as the directional research claim carrying burden of proof, status quo null $H_0: \mu = 16.0\text{ fl oz}$ (or $\mu \ge 16.0$), and establishing the left-tailed rejection region.
  - `curriculum/approved/errors-define-in-context.json` + review: semiconductor quality control ($H_0: \text{conforming lot}$ vs $H_1: \text{defective lot}$), defining Type I error as scrapping a conforming lot (Producer's risk / false alarm) and Type II error as shipping defective chips to automakers (Consumer's risk / missed defect), plus analyzing how reducing $\alpha$ increases $\beta$ for fixed $n$ and why expanding $n$ is required to shrink both risks.
  - `curriculum/approved/hyp-z-mean-two-tailed.json` + review: beverage bottling line $\mu_0 = 2.00\text{ L}, \sigma = 0.05\text{ L}, n = 36, \bar{x} = 2.02\text{ L}$, $SE = 0.05/6 \approx 0.008333\text{ L}$, standardized test statistic $Z = +2.40$, two-tailed $p$-value $= 2(1 - \Phi(2.40)) = 0.0164$, rejecting $H_0$ at $\alpha = 0.05$ since $0.0164 < 0.05$ (and $|Z| = 2.40 > 1.960$). Excel: `=2*(1 - NORM.S.DIST(2.40, TRUE))`.
  - `curriculum/approved/hyp-t-mean-right-tailed.json` + review: CRM customer support handling duration $\mu_0 = 50.0\text{ s}, n = 20, \bar{x} = 53.1\text{ s}, s = 6.2\text{ s}$, degrees of freedom $df = 19$, $SE = 6.2/\sqrt{20} \approx 1.3864\text{ s}$, test statistic $t = +2.236$, right-tailed $p$-value $= P(T_{19} \ge 2.236) \approx 0.0188$, rejecting $H_0$ at $\alpha = 0.05$ since $2.236 > t_{0.05, 19} = 1.729$. Excel: `=T.DIST.RT(2.236, 19)`.

- **Module 6 (5 questions)**:
  - `curriculum/approved/reg-fit-raw-points-price-units.json` + review: 6 empirical $(x, y)$ coordinate pairs across price and units sold ($n = 6$), $\bar{x} = 15.0, \bar{y} = 60.0, SS_{xx} = 70.0, SS_{xy} = -280.0$, OLS slope $b_1 = -4.00$, intercept $b_0 = 120.00$, fitted model $\hat{y} = 120.0 - 4.0x$, point prediction $\hat{y} = 52.0$ units at $x = \$17.00$, and $R^2 = 0.9756$. Excel: `=SLOPE`, `=INTERCEPT`, `=FORECAST.LINEAR`.
  - `curriculum/approved/regout-r2-from-anova.json` + review: e-commerce ad spend vs revenue regression ($n = 25, k = 1$), ANOVA table $SSR = 450.0, SSE = 150.0, SST = 600.0$, computing $R^2 = 450/600 = 0.750$ ($75.0\%$ variance explained), Multiple $R = \sqrt{0.75} = 0.8660$, verifying $t_1^2 = 69.00 = F = 450 / (150/23)$.
  - `curriculum/approved/regout-slope-t-pvalue.json` + review: engineer compensation vs experience ($n = 25, k = 1, df = 23$), reading slope $b_1 = 2.400, s_{b_1} = 0.800$, test statistic $t = 2.4 / 0.8 = +3.000$, two-tailed $p$-value $= 0.0064$, rejecting $H_0: \beta_1 = 0$ at $\alpha = 0.05$ ($0.0064 < 0.05$). Excel: `=T.DIST.2T(3.0, 23)`.
  - `curriculum/approved/mreg-which-predictors-significant.json` + review: retail store revenue multiple regression ($n = 30, k = 3, df = 26$), overall model $SSR = 1200.0, SSE = 400.0, F = 26.00$ ($p < 0.0001$), individual predictors `Floor_Space` ($p = 0.0006 < 0.05$) and `Competitor_Distance` ($p = 0.0384 < 0.05$) significant, `Local_Population` ($p = 0.6213 > 0.05$) not significant, demonstrating model parsimony.
  - `curriculum/approved/regprob-multicollinearity.json` + review: digital advertising multichannel model diagnosing severe multicollinearity from high $R^2 = 0.88$ and significant overall $F$ ($p < 0.0001$) accompanied by insignificant individual $t$-tests ($p > 0.20$), explaining variance inflation of $s_{b_j}$ and VIF diagnostics.

- **Module 1 (4 questions)**:
  - `curriculum/approved/concept-descriptive-vs-inferential.json` + review: classifying census audit of all 120 stores as descriptive statistics (zero sampling error) vs 40-store customer dwell time survey as inferential statistics (sample estimation under sampling uncertainty).
  - `curriculum/approved/concept-parameter-vs-statistic.json` + review: SaaS customer satisfaction survey ($N = 20{,}000$ active users, $n = 500$ sample, $62\%$ satisfied), identifying population, sample, parameter ($p$), statistic ($\hat{p} = 0.62$), and frequentist principle of fixed parameter constant vs random variable sample statistic.
  - `curriculum/approved/prob-multiplication-independent-suppliers.json` + review: independent suppliers A and B with $P(A) = 0.90, P(B) = 0.85$, special multiplication rule $P(A \cap B) = 0.90 \times 0.85 = 0.765$ ($76.5\%$), and conditional invariance $P(A \mid B) = P(A) = 0.90$.
  - `curriculum/approved/prob-at-least-one-server-failure.json` + review: 3 independent cloud servers with outage probability $p = 0.05$, applying complement strategy $P(\text{at least one failure}) = 1 - P(\text{zero failures}) = 1 - (0.95)^3 = 1 - 0.857375 = 0.142625$ ($14.26\%$), and debunking the naive addition trap ($0.05 \times 3 = 0.15$).

- **Module 2 (3 questions)**:
  - `curriculum/approved/choose-binomial-or-poisson.json` + review: evaluating four operational count scenarios to classify fixed-trial counts ($n = 50$ boards, $n = 200$ visitors $\implies$ Binomial) vs continuous-interval occurrence rates ($\lambda = 14$ calls/hr, $1.8$ scratches/sheet $\implies$ Poisson).
  - `curriculum/approved/binomial-at-least-three-defects.json` + review: sensor batch manufacturing $n = 12, p = 0.15$, cumulative lower sum $P(X \le 2) = P(0) + P(1) + P(2) = 0.1422 + 0.3012 + 0.2924 = 0.7358$, upper tail $P(X \ge 3) = 1 - 0.73582 = 0.2642$ ($26.42\%$), and avoiding off-by-one boundary subtraction. Excel: `=1 - BINOM.DIST(2, 12, 0.15, TRUE)`.
  - `curriculum/approved/poisson-rate-rescaling-quarter-hour.json` + review: IT help desk call arrivals at $6.0\text{ calls/hr}$, rescaling to 15-minute window $\lambda = 6.0 \times (15/60) = 1.50\text{ calls}$, computing $P(X \le 1) = 2.5 e^{-1.5} \approx 0.5578$ ($55.78\%$), and demonstrating why unscaled $\lambda = 6.0$ underestimates the true probability by $32\times$. Excel: `=POISSON.DIST(1, 1.5, TRUE)`.

- **Engine Extension**:
  - Formalized `multiplication` operation in `SetProbabilityFamily` (`internal/bank/family.go`) and added `mathengine.Multiplication(pA, pB)` with unit test coverage in `internal/mathengine/setprob_test.go` and `internal/bank/bank_test.go`.

- **Verification**:
  - `python scripts/validate_scaffold.py` passes with 0 errors across 151 JSON files.
  - All 15 Go packages pass `go test -count=1 ./...`.
  - Active curriculum bank loads cleanly with **64 approved templates**.
  - All 19 Vitest test suites pass (100/100 tests).

**Next concrete task:** Author **Dispersed Wave 2 (20 questions)** in [CONTENT-BACKLOG.md](CONTENT-BACKLOG.md):
- Module 4 (4 questions): `ci-mean-t-small-sample-90`, `ci-interpretation-statement`, `n-proportion-pilot`, `n-halve-margin-quadruple`
- Module 5 (5 questions): `hyp-formulate-right-tailed-sales`, `hyp-steps-and-conclusion`, `errors-which-is-worse`, `errors-alpha-beta-tradeoff`, `hyp-proportion-left-tailed`
- Module 6 (6 questions): `reg-interpret-slope-intercept`, `reg-residual-observed-point`, `regout-se-from-sse`, `regout-slope-ci-includes-zero`, `regout-slope-ci-interpretation`, `mreg-write-equation-predict`
- Module 1 (3 questions): `concept-sampling-bias-online-poll`, `concept-data-types-business`, `prob-independence-check`
- Module 2 (2 questions): `binomial-mean-sd-shipments`, `binomial-condition-violated-small-lot` (completes Row 2.3 at 5/5).

---

## Session Randomization & Curriculum Module Filtering Completed — October 11, 2026 (Windows)

Completed session randomization mechanism, module filtering across all 6 curriculum modules, and interactive controls:
- **Session Randomization Mechanism**:
  - Implemented clock-based PRNG seeding (`time.Now().UnixNano()` / `Date.now()`) for practice drills (`internal/httpapi/server.go`) and exams (`internal/exam/exam.go`).
  - Added option shuffling and mastery-weighted sampling across active templates so learners get different problems and shuffled answer choices across sessions.
  - Preserved snapshot isolation and deterministic introductory walkthrough for clean database test runs (`binomial_fair_coin_exactly_two` at position 1 and `binomial_defective_at_most_one` at position 2 when unconstrained).
- **Curriculum Module Filter**:
  - Configured full support for Modules 1–6 in Practice Drills (`web/src/features/settings/SettingsModal.tsx`): individual checkboxes with descriptions, "Select All", "Clear All", question count selection (5, 10, 15, 20, or All 44), intensity modes, and settings persistence (`/api/settings`).
  - Added module dropdown in Assessments (`web/src/features/exams/Exams.tsx`) allowing full mock exams targeted to any individual module or comprehensive across all modules.
- **Quick Randomization Actions**:
  - Added front-and-center "🎲 New Drill" buttons in both the main application header and the practice drill action bar (`web/src/app/App.tsx`, `web/src/features/practice/PracticeDrill.tsx`), allowing immediate clock-seeded problem re-shuffling at any point.
- **Verification**:
  - All 15 Playwright E2E test suites passed (15/15) including arcade, candidates, exams, foundation 01–12, visuals, and worksheets.
  - All 19 Vitest test files passed (100/100 tests).
  - All 15 Go packages passed (`go test -count=1 ./...`).
  - `python scripts/validate_scaffold.py` passed with 0 errors.

## Batch 2: Row 4.2 completed — October 11, 2026 (Windows)

Completed all 5 new approved questions and reviews for **Row 4.2 (Sample size and standard error)** in `sample_size`, `sampling_distribution`, and `concept_check` families:
- `curriculum/approved/se-quadruple-n-halves.json` + review: aerospace titanium pin diameters $\sigma = 20.0\\ \mu\text{m}$, contrasting initial $n_1 = 25$ ($SE_1 = 20/5 = 4.0\\ \mu\text{m}$) against quadrupled $n_2 = 100$ ($SE_2 = 20/10 = 2.0\\ \mu\text{m}$), and demonstrating the inverse square-root law where a $4\times$ data increase cuts standard error in half. Excel: `=20/SQRT(25)` and `=20/SQRT(100)`.
- `curriculum/approved/se-n-for-target-se.json` + review: emergency department wait times $\sigma = 15.0$ minutes, target standard error $SE \le 2.0$ minutes, deriving continuous bound $n \ge (15/2)^2 = 56.25$, proving why conventional round-down ($n = 56 \implies SE \approx 2.0045 > 2.0$) violates constraint, requiring ceiling integer $n = 57$ ($SE \approx 1.9869 \le 2.0$). Excel: `=ROUNDUP((15/2)^2, 0)`.
- `curriculum/approved/se-proportion-depends-on-p.json` + review: comparing statewide referendum $p_A = 0.50$ against niche EV ownership $p_B = 0.10$ at identical sample sizes $n = 400$, computing $SE_A = \sqrt{0.25/400} = 0.025$ ($2.5\%$) vs $SE_B = \sqrt{0.09/400} = 0.015$ ($1.5\%$), and proving parabolic variance peak at $p = 0.50$ explaining why split elections have maximum sampling uncertainty. Excel: `=SQRT(0.50*0.50/400)` and `=SQRT(0.10*0.90/400)`.
- `curriculum/approved/se-what-n-changes.json` + review: enterprise HR salary distribution ($\mu = \$92\text{k}, \sigma = \$24\text{k}$), establishing that expanding sample size from $n = 36$ to $n = 144$ leaves true population parameters $\mu$ and $\sigma$ completely unchanged, while narrowing the sampling distribution spread from $\sigma_{\bar{x}} = \$4{,}000$ to $\$2{,}000$, and debunking ecological fallacies that sample size compresses individual employee wage dispersion.
- `curriculum/approved/se-survey-design-compare.json` + review: telecom Net Promoter Score survey designs ($n_1 = 100$ vs $n_2 = 900$), computing error ratio $E_1 / E_2 = \sqrt{900/100} = 3.0$, analyzing the $66.7\%$ error reduction vs $9\times$ data collection cost, and deriving the general quadratic scaling law requiring $k^2 = 25\times$ sample size to achieve a $k = 5\times$ precision gain. Excel: `=SQRT(900/100)` and `=5^2`.

Engine prerequisite (ENG-SAMPLESIZE):
- Implemented `internal/mathengine/sample_size.go` with domain models and functions: `ComputeSampleSizeMean`, `ComputeSampleSizeProportion`, `ComputeSampleSizeTargetSE`, `StandardErrorMean`, and `StandardErrorProportion`.
- Registered `sample_size` family with rule version 1 and parameter validation in `internal/bank/family.go` and `DefaultRegistry()`.
- Added unit test coverage in `internal/mathengine/sample_size_test.go` and `internal/bank/bank_test.go`.

Active approved templates expanded from 39 to 44 (16 existing + 17 Batch 1 + 6 Row 4.1 + 5 Row 4.2). All 15 Go packages pass `go test -count=1 ./...`, `go vet`, and `python scripts/validate_scaffold.py` passes with 0 errors.

**Next concrete task:** Batch 2: Row 4.3 (Confidence intervals for mean and proportion, 6 questions) or Row 4.4 (Required sample size for margin of error, 5 questions).

## Batch 2: Row 4.1 completed — October 11, 2026 (Windows)

Completed all 6 new approved questions and reviews for **Row 4.1 (Central Limit Theorem — mean and proportion)** in `sampling_distribution` (and `concept_check`) family:
- `curriculum/approved/clt-mean-order-value.json` + review: order values $\mu = \$80.00, \sigma = \$24.00, n = 36 \ge 30$, standard error $\sigma_{\bar{x}} = 24 / 6 = \$4.00$, $Z = (86 - 80) / 4 = 1.50$, upper tail $P(\bar{X} > 86) = 1 - \Phi(1.50) = 0.0668$, and Excel formula `=1 - NORM.DIST(86, 80, 24/SQRT(36), TRUE)`.
- `curriculum/approved/clt-skewed-service-times.json` + review: technical support call durations following a heavily right-skewed exponential distribution $\mu = 6.0, \sigma = 6.0, n = 64 \ge 30$, normal approximation justification by CLT, $\sigma_{\bar{x}} = 6.0 / 8 = 0.75$, $Z = (4.8 - 6.0) / 0.75 = -1.60$, lower tail $P(\bar{X} < 4.8) = \Phi(-1.60) = 0.0548$, and Excel formula `=NORM.DIST(4.8, 6.0, 6.0/SQRT(64), TRUE)`.
- `curriculum/approved/clt-individual-vs-mean.json` + review: baggage weights $N(40.0, 10.0^2)$, contrasting single passenger tail $P(X > 45.0) = 1 - \Phi(0.50) = 0.3085$ against 25-passenger sample mean tail $P(\bar{X} > 45.0) = 1 - \Phi(2.50) = 0.0062$, demonstrating the $\sqrt{n} = 5\times$ standard error reduction ($2.0$ lbs vs $10.0$ lbs) and resulting $50\times$ tail probability drop.
- `curriculum/approved/clt-small-n-not-justified.json` + review: insurance flood claims ($\mu = \$25\text{k}, \sigma = \$18\text{k}$, heavily right-skewed), evaluating intern's small sample $n = 9 < 30$, identifying that CLT requires $n \ge 30$ when population is non-normal, showing exact small-$n$ normality requires an underlying normal population, and analyzing why symmetric normal tables underestimate right-tail catastrophe risk.
- `curriculum/approved/clt-proportion-coupon-users.json` + review: promotional coupon redemptions $p = 0.30, n = 200$, verifying success-failure conditions $np = 60 \ge 10$ and $n(1-p) = 140 \ge 10$, standard error $\sigma_{\hat{p}} = \sqrt{0.30(0.70)/200} \approx 0.0324$, $Z = (0.35 - 0.30) / 0.0324 = +1.54$, tail probability $P(\hat{p} > 0.35) = 1 - \Phi(1.54) = 0.0618$, and Excel formula `=1 - NORM.DIST(0.35, 0.30, SQRT(0.30*0.70/200), TRUE)`.
- `curriculum/approved/clt-finite-population-correction.json` + review: commercial loan balances $N = 400, n = 64$, verifying sampling fraction $n/N = 0.16 > 0.05$ requires FPC, computing FPC factor $\sqrt{(400-64)/399} = \sqrt{336/399} \approx 0.9177$, reducing uncorrected standard error $\$1{,}000.00$ to corrected $\sigma_{\bar{x}} = \$917.66$, and explaining uncertainty reduction from sampling without replacement.

Engine extension (ENG-PROP-SAMPLING):
- Added `SamplingDistributionProportion` domain type, `ComputeSamplingDistributionProportion(p, n, popSize)`, `ZScore`, `ProbabilityBelow`, and `ProbabilityAbove` to `internal/mathengine/inference.go`.
- Added `ZScore`, `ProbabilityBelow`, and `ProbabilityAbove` helper methods to `SamplingDistributionMean`.
- Extended `SamplingDistributionFamily.ValidateParameters` in `internal/bank/family.go` to support optional parameters `x_bar`, `p_hat`, `x`, `pop_shape`, `pop_is_normal`, and `pop_size` with strict bounds checking ($n \le pop\_size$).
- Added unit tests in `internal/mathengine/inference_test.go` and `internal/bank/bank_test.go`.

Active approved templates expanded from 33 to 39 (16 existing + 17 Batch 1 + 6 Row 4.1). All 15 Go packages pass `go test -count=1 ./...`, `go vet`, and `python scripts/validate_scaffold.py` passes with 0 errors.

**Next concrete task:** Batch 2: Row 4.2 (Sample size and standard error, 5 questions) or Row 4.3 (Confidence interval for mean and proportion, 6 questions).

## Batch 1: Row 3.3 completed — October 11, 2026 (Windows)

Completed all 6 new approved questions and reviews for **Row 3.3 (Exponential distribution)** in `exponential_dist` family:
- `curriculum/approved/exponential-arrival-within-two.json` + review: bank teller arrivals $\text{Exp}(\lambda = 0.25)$ ($\mu = 4$ min), $P(X \le 2) = 1 - e^{-0.5} = 0.3935$, and Excel function `=EXPON.DIST(2, 0.25, TRUE)`.
- `curriculum/approved/exponential-memoryless-wait.json` + review: tech support queue wait $\mu = 6$ min ($\lambda = 1/6$), memoryless conditional wait $P(X > 8 \mid X > 5) = P(X > 3) = e^{-0.5} = 0.6065$, and managerial invariance of past wait time on remaining duration.
- `curriculum/approved/exponential-rate-vs-mean-param.json` + review: 911 dispatch interval $\mu = 10$ min ($\lambda = 0.10$), $P(X \le 5) = 1 - e^{-0.5} = 0.3935$, and analysis of the classic software trap where entering mean $10$ into `=EXPON.DIST` inflates the probability to $1 - e^{-50} \approx 1.000$.
- `curriculum/approved/exponential-percentile-service.json` + review: cloud infrastructure SLA ticket resolution $\lambda = 0.40$/hr ($\mu = 2.5$ hrs), solving 90th percentile $x^* = -\frac{\ln(0.10)}{0.40} = 5.76$ hrs, absence of `EXPON.INV` in Excel, and right-skewed tail elongation ($5.76$ hrs $> 2 \times \mu$).
- `curriculum/approved/exponential-poisson-link.json` + review: e-commerce payments at 12 transactions/hour ($\lambda = 0.20$/min), interarrival gap $P(X > 10) = e^{-2} = 0.1353$, and exact mathematical duality with discrete Poisson zero-arrival probability $P(Y = 0) = e^{-2}$.
- `curriculum/approved/exponential-between-two-times.json` + review: EV inverter module lifetime $\mu = 1{,}000$ hrs ($\lambda = 0.001$), interval probability $P(500 < X < 1500) = e^{-0.5} - e^{-1.5} = 0.3834$, and continuous endpoint boundary invariance $P(500 < X < 1500) = P(500 \le X \le 1500)$.

Extended `ExponentialFamily.ValidateParameters` in `internal/bank/family.go` to support optional parameters `mean`, `p` ($0 < p < 1$), `s` ($\ge 0$), `t` ($> 0$), `x` ($\ge 0$), and `x1, x2` ($x_1 \le x_2$). Added `ExcelQuantile` helper to `ExponentialDistribution` in `internal/mathengine/continuous.go`. Added unit test coverage in `internal/bank/bank_test.go` and `internal/mathengine/continuous_test.go`.

Batch 1 (Module 3: Continuous distributions) is now fully complete (Rows 3.1, 3.2, and 3.3). Total active approved templates expanded from 27 to 33 (1 existing + 6 new for Row 3.3). All 16 Go packages pass `go test -count=1 ./...`, `go vet`, and `python scripts/validate_scaffold.py` passes with 0 errors.

**Next concrete task:** Batch 2: Row 4.1 (Central Limit Theorem — mean and proportion, 6 questions) in `curriculum/approved/` using `sampling_distribution` (+ ENG-PROP-SAMPLING).

## Batch 1: Row 3.2 completed — October 11, 2026 (Windows)

Completed all 6 new approved questions and reviews for **Row 3.2 (Normal distribution)** in `normal_dist` family:
- `curriculum/approved/normal-between-package-weight.json` + review: packaging net weight $N(500, 8^2)$, standardizing bounds $Z_1 = -1.25, Z_2 = +1.25$, interval probability $P(490 < X < 510) = \Phi(1.25) - \Phi(-1.25) = 0.7887$, and endpoint boundary invariance $P(490 < X < 510) = P(490 \le X \le 510)$.
- `curriculum/approved/normal-inverse-bonus-cutoff.json` + review: executive sales bonus $N(80000, 12000^2)$, top-5% critical score $Z^* = +1.645$, unstandardized threshold $X^* = 80000 + 1.645(12000) = \$99,740$, and Excel function `=NORM.INV(0.95, 80000, 12000)`.
- `curriculum/approved/normal-warranty-lower-percentile.json` + review: commercial tire tread lifespan $N(60000, 8000^2)$, lower 2% tail quantile $Z^* \approx -2.054$, warranty limit $X_w = 60000 - 2.054(8000) = 43,570\text{ miles}$, and cost analysis showing replacement rate jumps from 2% to 10.56% if set to 50k miles.
- `curriculum/approved/normal-compare-standardized-scores.json` + review: cross-territory manager evaluation $N(500000, 50000^2)$ ($x_A = \$580k \implies Z_A = +1.60$) vs $N(300000, 40000^2)$ ($x_B = \$370k \implies Z_B = +1.75$), demonstrating Manager B's superior relative achievement despite lower raw volume.
- `curriculum/approved/normal-inventory-service-level.json` + review: hospital demand $N(200, 30^2)$, 90% cycle service level $Z^* \approx +1.282$, safety stock $38.45$ kits, target stock level $s = 238.45$ kits, and complement stockout probability $P(X > 238.45) = 0.1000$.
- `curriculum/approved/normal-empirical-vs-exact.json` + review: roller bearing diameters $N(100, 15^2)$, comparing Empirical Rule's $\approx 95\%$ rule-of-thumb against exact cumulative normal table value $\Phi(2.00) - \Phi(-2.00) = 0.9545$, and bell-shaped normality prerequisites vs. Chebyshev bounds.

Extended `NormalFamily.ValidateParameters` in `internal/bank/family.go` to support optional quantile parameter `p` ($0 < p < 1$), multiplier `k` ($k > 0$), and comparison parameters `mu_b`, `sigma_b`, `x_b`. Added unit tests in `internal/bank/bank_test.go`.

Total active approved templates expanded from 21 to 27 (1 existing + 6 new for Row 3.2). All 15 Go packages pass `go test -count=1 ./...`, `go vet`, and `python scripts/validate_scaffold.py` passes with 0 errors.

**Next concrete task:** Row 3.3 (Exponential distribution — 6 questions, 1 existing + 5/6 new) in `curriculum/approved/` using `exponential_dist` family.

## Batch 1: Row 3.1 completed — October 11, 2026 (Windows)

Completed all 5 new approved questions and reviews for **Row 3.1 (Discrete vs. continuous distributions)**:
- `curriculum/approved/contvdisc-classify-variables.json` + review: classifies wait time (continuous), complaints (discrete), package weight (continuous), defects (discrete).
- `curriculum/approved/contvdisc-point-probability-zero.json` + review: continuous $P(X = 5) = 0$ vs. discrete Poisson $P(Y = 5) \approx 0.1755$, explaining why continuous probabilities represent intervals/areas under curves.
- `curriculum/approved/contvdisc-area-is-probability.json` + review: uniform delivery window $U(10, 40)$, density height $1/30$, $P(15 < X < 25) = 1/3 \approx 0.3333$, mean $E[X] = 25$.
- `curriculum/approved/contvdisc-strict-vs-nonstrict.json` + review: boundary invariance $P(X \le c) = P(X < c)$ for continuous Normal vs. discrete Binomial inequality differences $P(Y < 2) \ne P(Y \le 2)$.
- `curriculum/approved/contvdisc-match-model.json` + review: taxonomy matching business scenarios to Exponential, Normal, Binomial, and Poisson distributions.

Total active approved templates expanded from 16 to 21. All Go tests, bank validation (`TestLoadActiveBankAdmitsApproved`), and scaffold checks pass.

**Next concrete task:** Row 3.2 (Normal distribution — 6 questions, 1 existing + 5 new) in `curriculum/approved/` using `normal_dist` family.

## Content backlog engine prerequisites (Batch 1) — October 11, 2026 (Windows)

Implemented and verified the two engine prerequisites:
- `internal/mathengine/f_dist.go`: Snedecor's $F$-distribution $F(d_1, d_2)$ with exact CDF and survival via regularized incomplete beta $I_x(a, b)$, mean, variance, Excel formulas (`=F.DIST.RT`), and equivalence testing with Student's $t$ ($F(1, \nu) = t^2(\nu)$).
- `internal/mathengine/regression_output.go`: Excel-style regression output summary engine supporting $k \in [1, 4]$ predictors, evaluating ANOVA ($MSR, MSE, F, \text{Significance } F$), model fit ($R^2, \text{adj } R^2, s_e$), predictor coefficients ($t$, $p$-values, $(1-\alpha)\%$ CIs, $\hat{y}$ predictions), and consistency checks ($SST = SSR + SSE, t^2 = F$ when $k = 1$).
- `internal/bank/family.go`: Registered `concept_check` and `regression_output` problem families; added `TemplateValidator` interface enforcing choice-only stages, required misconception IDs, and hint markdown on all distractors.
- `internal/bank/parser.go`: Connected `TemplateValidator` into `ValidateTemplate`.
- Unit tests: Added comprehensive test coverage in `f_dist_test.go`, `regression_output_test.go`, and `new_families_test.go`. All 15 Go packages pass `go test -count=1 ./...` and `go vet`. Frontend test suite passed (99/99 tests).

## Content backlog plan — October 11, 2026 (Windows)

Added [CONTENT-BACKLOG.md](CONTENT-BACKLOG.md): 22 objective rows from the published module objectives, 121 proposed questions (at least 5 new per row), 9 engine prerequisites, batch order and open course conventions. Recorded the objectives in [COURSE-MAP.md](COURSE-MAP.md) and removed multiple regression from its exclusions.

## Stage 17 — October 10, 2026 (macOS)

Stage 00 through Stage 17 are complete within their documented scope. Stage 18 (Arcade complete parity) is next. All changes from this Mac session remain local and uncommitted.

### Implemented

- `web/src/features/arcade/types.ts`: Domain models, geometry constants, and configuration for Quant Flight space arcade:
  - Logical coordinate bounds (800x480), 60Hz fixed simulation step (`FIXED_DT = 1/60`), bounded entity pools (max 40 projectiles, 15 hazards, 70 particles).
  - Atmospheric statistical hazard labels (`Selection Bias`, `Confounding`, `Bad Denominator`, `Spurious Correlation`, `Missing Covariance`).
  - Full typing for ship physics, projectiles, hazards, particle sparks, stars, held input state, and game state.
- `web/src/features/arcade/simulation.ts`: Pure `step(state, input, dt, rng)` and `createInitialState(seed, mode)` fixed-step simulation engine:
  - Deterministic pseudo-random number generator using Mulberry32 for seedable runs.
  - Attract autopilot mode with sinusoidal cruising and line-of-sight targeting.
  - Immediate manual takeover upon any player movement or fire input (`takeoverManual`), resetting score, timer, and targets to 0 so demo play never contaminates human records (enforcing Invariant 12).
  - Velocity normalization on diagonal steering, smooth acceleration, banking roll tilt, weapon cooldown, circle-to-circle collision detection for projectiles and ship shields against hazards, particle bursts, and pause/reduced-motion handling.
- `web/src/features/arcade/renderer.ts`: Pure Canvas 2D renderer with original assets and geometry:
  - Deep space gradient and 3-layer parallax starfield (slow, mid, fast stars).
  - Original vector interceptor ship geometry with cockpit canopy, wing cannons, and dynamic dual engine exhaust glow.
  - Faceted hexagonal hazard bodies with glowing danger cores and statistical labels.
  - Projectile capsule energy bolts and particle spark explosion systems.
  - Heads-Up Display (HUD) showing pilot mode badge, score, hazards cleared count, shields meter, and shortcut hints.
  - Attract banner call-to-action and paused simulation overlay.
- `web/src/features/arcade/ArcadeGame.tsx`: React component managing fixed-step accumulator game loop via `requestAnimationFrame`:
  - High-DPI crisp canvas scaling with `devicePixelRatio`.
  - Simultaneous held-key input tracking for Arrows, WASD, j/k, Space, and F.
  - Immediate keydown takeover latching and canvas click takeover.
  - Window blur and document visibility key clearance and auto-pausing when unfocused.
  - Pause toggle button and shortcut (`p`).
  - Skip-to-practice button (`#btn-skip-arcade`) and shortcuts (`Escape`, `Enter`) providing immediate return to drills from any state.
  - Reduced motion toggle and system media-query detection.
  - Screen-reader accessible live announcement region (`#arcade-status-announcement`) with `aria-live="polite"`.
- `web/src/app/App.tsx` and `web/src/navigation/keymap.ts`:
  - Header tab navigation button (`#tab-arcade`, `Shift+A` / `A`).
  - Routed `VIEW_ARCADE` command.
  - Yields keyboard ownership to the arcade when active, and restores drill ownership immediately upon exit.
  - Complete isolation: zero arcade actions alter practice sessions, mastery records, or exams.
- `web/src/features/arcade/simulation.test.ts` & `ArcadeGame.test.tsx`:
  - 15 unit and component tests covering seed determinism, fixed-step simulation, attract score isolation, manual takeover, canvas boundary clamping, continuous fire cooldown, collisions, pause freezing, reduced-motion suppression, skip triggers, and pause toggles.
- `tests/e2e/arcade.spec.ts`:
  - End-to-end Playwright browser suite verifying arcade opening, attract mode presentation, manual takeover on keypress, pause/resume, reduced motion toggle, 375px mobile viewport responsiveness, immediate return to practice via button and Escape, and verified screenshot capture (`tests/e2e/screenshots/stage17_verified.png`).

### Actual checks

- Native `./scripts/test.sh` passed on macOS:
  - Documentation and scaffold check: 32 Markdown files, 74 local links, 54 JSON files, 35 source directories (0 errors).
  - Frontend typecheck (`tsc --noEmit`): 0 errors.
  - Frontend lint (`eslint src`): 0 errors.
  - Frontend unit tests (Vitest): 99 passed across 19 test files.
  - Frontend production build (`tsc && vite build`): bundle compiled cleanly.
  - Native Go binary compilation: `bin/quant-practice` compiled with embedded UI and bank.
  - Playwright E2E browser tests: 15/15 passed across all suites (`arcade.spec.ts`, `visuals.spec.ts`, `candidates.spec.ts`, `worksheets.spec.ts`, `exams.spec.ts`, and `foundation.spec.ts`).
  - Go formatting check (`gofmt -s -l cmd internal`): 100% formatted.
  - Go unit tests (`go test -count=1 ./internal/... ./cmd/...`): all 16 packages passed.
  - Go vet (`go vet ./internal/... ./cmd/...`): 0 issues.
- Verified screenshot generated at `tests/e2e/screenshots/stage17_verified.png`.

## Stage 16 — October 10, 2026 (macOS)

Stage 00 through Stage 16 are complete within their documented scope. Stage 17 (Arcade first playable) is next. All changes from this Mac session remain local and uncommitted.

### Implemented

- `internal/mathengine/`: Advanced course-grounded statistical inference and regression engines:
  - `student_t.go`: Student's $t$-distribution $T \sim t(\nu)$ with exact PDF via log-gamma, CDF and survival via Lentz's continued fraction for regularized incomplete beta $I_x(a, b)$, and quantile (inverse CDF) via Halley-Cornish-Fisher iteration verified against textbook statistical tables ($\nu = 9, 19, 29, 20$).
  - `inference.go`: Sampling distribution of the mean $\bar{X} \sim N(\mu, \sigma^2/n)$ (with Finite Population Correction $\sqrt{(N-n)/(N-1)}$ when $n/N > 0.05$), proportion standard error $\sqrt{p(1-p)/n}$, confidence intervals for mean ($Z$ and $t$) and proportion, and Monte Carlo confidence interval coverage simulation.
  - `hypothesis.go`: One-sample $z$ and $t$ tests for mean, $z$-test for proportion; two-tailed, left-tailed, and right-tailed tests; critical value cutoffs, exact $p$-values, decision logic, and enforcement of the frequentist invariant that failing to reject $H_0$ never proves $H_0$ is true.
  - `regression.go`: Ordinary Least Squares (OLS) regression for paired bivariate data $(x_i, y_i)$, slope $b_1 = SS_{xy}/SS_{xx}$, intercept $b_0 = \bar{y} - b_1 \bar{x}$, residuals $e_i = y_i - \hat{y}_i$, sample correlation $r$, coefficient of determination $R^2 = r^2$, standard error of estimate $s_e = \sqrt{SSE/(n-2)}$, prediction at $x^*$, and core invariant that association is not causation.
- `internal/bank/family.go`: Registered 4 new problem families: `sampling_distribution`, `confidence_interval`, `hypothesis_test`, and `linear_regression` with parameter validation, default instances, and stage derivation.
- `curriculum/approved/` & `curriculum/reviews/`: Added 3 new approved questions with human/delegated reviews:
  - `ci-mean-sample-estimate.json` (Module 4): 95% confidence interval for sample mean thickness with unknown $\sigma$ ($df = 15, t^* \approx 2.131$).
  - `hypothesis-mean-one-sample.json` (Module 5): Left-tailed one-sample $t$-test for product packaging underfill ($df = 24, t = -3.00, p = 0.0031$).
  - `regression-sales-advertising.json` (Module 6): OLS regression of retail sales on advertising ($b_1 = 1.80, b_0 = 12.80, R^2 = 0.81$, prediction at $x = 6.0$, association vs causation invariant).
  Active bank count expanded to 16 approved templates.
- `internal/httpapi/visuals.go`: Implemented 4 new statistical visualization endpoints:
  - `GET /api/visuals/sampling`: Central Limit Theorem simulation, empirical vs theoretical standard error, normal density curve overlay, and Excel formula `=NORM.DIST(x_bar, mu, sigma/SQRT(n), TRUE)`.
  - `GET /api/visuals/confidence_interval`: Confidence interval simulation, stacked whiskers marking capture vs miss of true $\mu$, empirical coverage scorecard, and Excel formulas `=CONFIDENCE.T`, `=CONFIDENCE.NORM`.
  - `GET /api/visuals/hypothesis_test`: Shaded critical rejection regions, test statistic position marker, critical value cutoff, $p$-value decision card, and Excel formulas `=T.DIST.2T`, `=T.DIST.RT`, `=NORM.S.DIST`.
  - `GET /api/visuals/regression`: Paired data scatter points, OLS fitted line, vertical residuals $e_i$, prediction at $x^*$, correlation & $R^2$, and Excel formulas `=SLOPE`, `=INTERCEPT`, `=RSQ`.
- `web/src/visuals/`: Interactive accessible SVG Visuals Lab:
  - `SamplingVisual.tsx`: Central Limit Theorem interactive simulator with population shape selector (skewed chi-square, uniform, exponential, normal), sample size slider, trial count selector, theoretical vs empirical standard error cards, and accessible SVG chart.
  - `ConfidenceIntervalVisual.tsx`: Confidence interval coverage simulator with confidence level toggle (90%, 95%, 99%), known vs unknown $\sigma$ toggle, sample size slider, stacked interval whiskers, empirical coverage rate card, and frequentist invariant explanation.
  - `HypothesisTestVisual.tsx`: Hypothesis test explorer with test type selector (mean $t$, mean $z$, proportion $z$), tail selector, significance level $\alpha$ selector, sample mean slider, rejection region shading, statistical decision card, and failure-to-reject caution.
  - `RegressionVisual.tsx`: Simple linear regression and residuals explorer with dataset selector, evaluated $x^*$ input, scatter plot with OLS line and vertical residual bars, $R^2$ decomposition, and prominent "Association is not Causation" warning card.
  - `VisualsExplorer.tsx`: Unified tabbed container updated with all 9 visual explorers and keyboard navigation.
  - `ReferenceLibrary.tsx`: Reference library cards updated with Modules 4–6 inference and regression reference cards and filter.
- `tests/e2e/visuals.spec.ts`: Expanded Playwright E2E browser tests covering all 9 visual tabs, CLT simulation, CI coverage, hypothesis testing decision badge, linear regression residuals and association warning, 375px mobile responsiveness, and verified screenshot capture (`tests/e2e/screenshots/stage16_verified.png`).

### Actual checks

- Native `./scripts/test.sh` passed on macOS:
  - Documentation and scaffold check: 43 Markdown files, 74 local links, 54 JSON files, 35 source directories (0 errors).
  - Frontend typecheck (`tsc --noEmit`): 0 errors.
  - Frontend lint (`eslint src`): 0 errors.
  - Frontend unit tests (Vitest): 84 passed across 17 test files.
  - Frontend production build (`tsc && vite build`): bundle compiled cleanly.
  - Native Go binary compilation: `bin/quant-practice` built with embedded UI and bank.
  - Playwright E2E browser tests: 14/14 passed across all suites (`visuals.spec.ts`, `candidates.spec.ts`, `worksheets.spec.ts`, `exams.spec.ts`, and `foundation.spec.ts`).
  - Go formatting check (`gofmt -s -l cmd internal`): 100% formatted.
  - Go unit tests (`go test -count=1 ./internal/... ./cmd/...`): all 16 packages passed.
  - Go race detector (`go test -race ./internal/...`): all packages passed with 0 data races.
  - Go vet (`go vet ./internal/... ./cmd/...`): 0 issues.
- Verified screenshot generated at `tests/e2e/screenshots/stage16_verified.png`.

## Stage 15 — October 10, 2026 (macOS)

Stage 00 through Stage 15 are complete within their documented scope. Stage 16 (sampling, confidence intervals, hypothesis testing and regression) is next. All changes from this Mac session remain local and uncommitted.

### Implemented

- `internal/mathengine/continuous.go`: Continuous distribution families and evaluations:
  - `NormalDistribution`: PDF $\phi(x; \mu, \sigma)$, CDF $\Phi(x; \mu, \sigma)$ via `math.Erf`/`math.Erfc` with stable extreme-tail bounds, survival function $1 - \Phi(x)$, and high-precision quantile (inverse CDF) via Peter John Acklam's rational probit approximation with Halley refinement.
  - `UniformDistribution`: Constant density $1/(b-a)$, linear CDF $(x-a)/(b-a)$, exact moments ($\mu = (a+b)/2, \sigma^2 = (b-a)^2/12$), and support validation $b > a$.
  - `ExponentialDistribution`: Decaying density $\lambda e^{-\lambda x}$, cumulative distribution $1 - e^{-\lambda x}$, survival $e^{-\lambda x}$, memoryless property preservation, and support $x \ge 0$.
  - Continuous event evaluator `EvaluateContinuousEvent`: exact point mass probability $P(X = c) = 0$, intervals $P(x_1 \le X \le x_2)$, lower tails $P(X \le x)$, upper survival tails $P(X \ge x)$, and symmetric two-tailed bounds.
- `internal/bank/family.go`: Registered `normal_dist`, `uniform_dist`, and `exponential_dist` in `DefaultRegistry()`, binding continuous parameter validation, default instances, and stage derivation.
- `internal/httpapi/visuals.go`: Reconciled visual math endpoints:
  - `GET /api/visuals/normal`: Returns curve points (PDF and CDF modes), $z$-score, shaded probability, 68-95-99.7 empirical rule values, and exact Excel formulas (`=NORM.DIST(x, mu, sigma, TRUE)`).
  - `GET /api/visuals/uniform`: Returns constant density height, interval bounds, shaded probability, moments, and Excel formulas.
  - `GET /api/visuals/exponential`: Returns decaying density points, cutoff height, survival probability, memoryless property notes, and Excel formulas (`=1 - EXPON.DIST(x, lambda, TRUE)`).
  - `GET /api/visuals/discrete`: Bar heights for Binomial and Poisson, event highlighting, and Excel formulas (`=BINOM.DIST`, `=POISSON.DIST`).
  - `GET /api/visuals/venn`: Set operations with Fréchet inequality validation ($P(A \cap B) \in [\max(0, P(A)+P(B)-1), \min(P(A), P(B))]$), analytical formulas, and independence checking.
- `curriculum/approved/`: Added 3 Module 3 approved questions:
  - `normal-standard-upper-tail.json`: Standard normal tail probability $P(Z \ge 1.645) \approx 0.05$.
  - `exponential-service-lifetime.json`: Service equipment survival with rate $\lambda = 0.5$.
  - `uniform-commute-wait-time.json`: Wait time between 0 and 15 minutes.
  Audit reviews added in `curriculum/reviews/`. Scaffold validator passes with 0 errors.
- `web/src/visuals/`: Interactive accessible SVG Visuals Lab:
  - `NormalVisual.tsx`: Interactive sliders for $\mu, \sigma$, tail region selector, PDF/CDF toggle, empirical rule standard deviation guidelines ($\pm 1\sigma, \pm 2\sigma, \pm 3\sigma$), accessible SVG diagonal hatching pattern, live Excel formula, and ARIA polite text summary.
  - `UniformVisual.tsx`: Sliders for bounds $[a, b]$, interval selector, rectangle shaded area, density height indicator.
  - `ExponentialVisual.tsx`: Rate parameter $\lambda$ slider, survival/lower/interval events, decaying curve with diagonal shading, memoryless property explanation.
  - `DiscreteVisual.tsx`: Binomial ($n, p$) and Poisson ($\lambda$) distribution bar charts, event highlighting, cumulative vs point mass views.
  - `VennVisual.tsx`: Interactive 2-set Euler/Venn diagram with Fréchet inequality slider enforcement, boolean region buttons (intersection, union, set differences, complement, neither), SVG clip paths and diagonal patterns, independence tester.
  - `VisualsExplorer.tsx`: Unified tabbed container with keyboard accessibility and back-to-practice button.
- `web/src/features/reference/ReferenceLibrary.tsx`: Reference library cards updated with Module 3 continuous distributions: formulas, properties, Excel functions, and exam pitfalls.
- `web/src/app/App.tsx` and `keymap.ts`: Wired `v` shortcut (`VIEW_VISUALS`) to toggle Visuals Lab, `Esc` to return to practice drill, tab button in header (`#tab-visuals`), and responsive mobile width constraints ($\le 376$px on 375px viewport).
- `tests/e2e/visuals.spec.ts`: End-to-end browser tests verifying all 5 visual explorers, dynamic curve updates, spreadsheet formula displays, Venn set operations, mobile responsiveness at 375px viewport, verified screenshot capture (`tests/e2e/screenshots/stage15_verified.png`), and `Escape` return to practice.

### Actual checks

- Native `./scripts/test.sh` passed on macOS:
  - Documentation and scaffold check: 32 Markdown files, 74 local links, 51 JSON files, 35 source directories (0 errors).
  - Frontend typecheck (`tsc --noEmit`): 0 errors.
  - Frontend lint (`eslint src`): 0 errors, 0 warnings.
  - Frontend unit tests (Vitest): 80 passed across 17 test files.
  - Frontend production build (`tsc && vite build`): successful bundle.
  - Native Go binary compilation: `bin/quant-practice` compiled cleanly.
  - Playwright E2E browser tests: 14/14 passed across all suites (including `visuals.spec.ts`, `candidates.spec.ts`, `worksheets.spec.ts`, `exams.spec.ts`, and `foundation.spec.ts`).
  - Go formatting check (`gofmt -l .`): 100% formatted.
  - Go unit tests with race detector (`go test -race ./...`): all 16 packages passed.
  - Go vet (`go vet ./...`): 0 issues.
- Verified screenshot generated at `tests/e2e/screenshots/stage15_verified.png` (182 KiB).

## Stage 14 — October 10, 2026 (macOS)

Stage 00 through Stage 14 are complete within their documented scope. Stage 15 (continuous distributions and math visuals) is next. All changes from this Mac session remain local and uncommitted.

### Implemented

- `internal/exam`: Domain engine for timed and untimed assessments with strict aid withholding. `New` constructs frozen question snapshots from approved content with deterministic per-question option shuffles. `Public` strips all hints, solutions, explanations, and expected keys via `PublicExamView` and `PublicOptionView`. Strict deadline enforcement uses UTC wall-clock time (`deadline_at`). Automated evaluation via mathematical engine (`EvaluateExam`), returning points, percentage score, stage breakdown, full mathematical derivations, and standalone Markdown export (`FormatMarkdownReport`).
- `internal/storage/exams.go`: Persistent exam storage implementing `CreateExam`, `GetExam`, `ListExams`, `CommandExam`, `GetExamReport`, and `GetHistory` over existing migration 001 SQLite tables (`sessions`, `question_instances`, `exam_responses`, `command_idempotency`). Replay idempotency via request hash, optimistic revision lock, and dynamic deadline expiration on read. Zero mastery writes (assessments remain strictly separate from independent learning ledger).
- `internal/httpapi/exams.go`: Session-authenticated and loopback/exact-Origin protected endpoints (`POST /api/exams`, `GET /api/exams`, `GET /api/exams/{id}`, `POST /api/exams/{id}/commands`, `GET /api/exams/{id}/report` with JSON and `?format=markdown`, `GET /api/history`). Unknown JSON fields and bodies > 128 KiB are rejected. Active exams return HTTP 403 on report retrieval.
- `cmd/quant-practice`: Added CLI assessment commands (`-exam [count]`, `-exam-time [minutes]`, `-resume-exam [id]`, `-exam-history`, `-exam-report [id]`) in `cmd/quant-practice/exams.go`.
- `web/src/features/exams`, `App.tsx`, and `keymap.ts`: Full React frontend in `web/src/features/exams/Exams.tsx` featuring setup dashboard, question count and time limit selectors, active countdown timer (`aria-label="Time Remaining"`), question strip, radio choice and numeric inputs, submission and abandonment confirmation modals, score hero card, detailed question review with revealed derivations, and Markdown report download link. Integrated into `App.tsx` header tabs and keymap (`Shift+E` / `E` for `VIEW_EXAM`), with automatic active exam recovery across server/browser restarts.
- Docs: Updated `API.md`, `PLAN.md`, `README.md`, and `HANDOFF.md` with Stage 14 details.

### Actual checks

- Native `./scripts/test.sh` passed on macOS: scaffold validation (`scripts/validate_scaffold.py`), frontend typecheck (`tsc --noEmit`), lint (`eslint src`), 73 frontend unit tests across 16 files, production build (`vite build` + `go build`), all 13 Playwright E2E browser tests (including `tests/e2e/exams.spec.ts`), gofmt check, unit tests across all 15 Go packages, and `go vet`.
- `tests/e2e/exams.spec.ts` verified complete lifecycle: starting a 30-minute exam, strict aid withholding during active exam, answering choice and numeric stages, native server stop and restart with wall-clock deadline preservation and answer recovery, submission confirmation modal, post-completion score report with unlocked keys/derivations, Markdown export link, history table update, and 375px mobile viewport responsiveness.
- `git diff --check` and `python3 scripts/validate_scaffold.py` passed with 0 errors.

## Stage 13 — October 9, 2026 (macOS)

Stage 00 through Stage 13 are complete within their documented scope. Stage 14 (exam mode) is next. All Mac-session changes, including Stage 12, remain local and uncommitted.

### Implemented

- `internal/worksheets`: approved-bank full-form snapshots using the progressive instantiator and `drill.GradeSubmission`; complete-form validation without partial attempts, one causal hint/retry per incorrect field, read-only completed review, public key withholding and Markdown worksheets. Approved source template/rule/approval, seed, parameters, option order and numeric policies are preserved independently of current bank state.
- `internal/mathengine/frequency.go` and worksheet CSV/case builder: generic relative frequency for recorded four-toss experiments, compared with the existing fair-independent-coin reference. Strict UTF-8/two-column/64 KiB/500-row import, unique safe IDs and counts 0–4; no expressions, file paths or inferred schemas. Hand-derived fixtures and mathematical references are recorded in `docs/WORKSHEETS.md`. Actual course sources remain absent (`course-materials/private` contains only `.gitkeep`).
- Case creation requires human review of the complete author preview (wording, options, hints, assumptions, keys and numeric policies), reviewer/source note and explicit confirmation. The dataset and semantic approval record are immutable. Full-form attempts are `full_solution_form`; cases also record `reference`, retries `hint`/`retry`, and reveals are on stage state. No worksheet evidence enters independent mastery/transfer, and no case enters the approved bank automatically.
- `internal/storage/{worksheets.go,migrations/002_worksheets.sql}`: separate worksheet state and command results, atomic revision checks, request hashes, exact idempotent replay after restart. Existing migration 001 is unchanged. Launch upgrade uses the existing consistent pre-migration backup. No drill/mastery table is written by worksheet commands.
- `internal/httpapi/worksheets.go`: authenticated/exact-Origin, strictly decoded bounded create/list/get/preview/command/Markdown-export endpoints. Shared app-session guard replaces duplicate candidate guard. Storage failures return a redacted 500 so the browser retains the exact pending request; validation errors consume no attempts.
- `web/src/features/worksheets`, App and CSS: Worksheets tab, Shift+J/Shift+F, all fields together, saved/history selection, bounded CSV upload/paste and semantic preview, durable draft, full submission/retry, solution review/download, navigation/stale-response handling, explicit uncertain-command retry, editable-field protection and narrow-screen math. Updated shortcut help. Removed obsolete “Stage 04 Drill” badge assertions; actual shell/drill/navigation checks remain.
- `internal/drill/grading.go`: incorrect numeric feedback now reveals the saved problem's own key instead of hardcoded `0.375`, with no hardcoded original-coin arithmetic in the first hint. Empirical numeric errors explain the experiment denominator.
- Docs: WORKSHEETS, API, STORAGE, TESTING, README, REQUIREMENTS and PLAN describe scope and evidence. Free-prose grading, arbitrary CSV schemas, other dataset families, course-specific cases, worksheet CLI and mastery contribution remain outside this initial scope.

### Actual checks

- Final `./scripts/test.sh` passed natively on macOS: scaffold, frontend typecheck/lint, 70 unit tests in 15 files, production frontend/native Go build, all 12 offline Playwright tests, gofmt, uncached Go tests across 14 internal packages and command package, and go vet. Older unrelated React `act(...)` warnings remain; new worksheet tests pass without them.
- Final `go test -race ./internal/worksheets ./internal/storage ./internal/httpapi` passed. Coverage includes every approved template's progressive/full agreement, safe CSV boundaries, snapshot isolation, own-key numeric reveal, hint/assistance, no mastery writes, restart and duplicate/stale/changed commands, injected transaction failure, 001→002 backup preservation, authenticated/exact-Origin endpoints, hidden keys and redacted save failure/retry.
- Focused worksheet browser run passed both flows: full-form draft and completed review across native restart, invalid input consumes no attempts, hinted numeric retry, Markdown export, editable shortcuts and 375px math; offline CSV file upload, explicit semantic approval, `2/5` observed versus `3/8` theoretical, preserved dataset/review after restart.
- Initial sandboxed browser launch could not start its loopback child; the permitted native run resolved that environment boundary. Early shortcut tests sent keys before app initialization; they now wait for authenticated readiness. Full-suite obsolete badge assertions blocked the first drill test and caused dependent restart/note failures; removing the obsolete assertions produced a clean full run. No unrelated note behavior was changed.
- No live provider calls or real-user content approvals were made. Synthetic case approvals use temporary test databases. No native Windows/Linux run is claimed.
- Final scaffold validation and `git diff --check` passed; the seven older browser verification screenshots were restored to committed versions. Stopped the older Stage 12 server and launched the verified app with `./scripts/start.sh`. The existing real database was backed up automatically (`quant-methods.db.backup-1791603925`) before migration 002. The app is left running at `http://127.0.0.1:64203/` (homepage HTTP 200), opened for user review. No real worksheet was submitted or case approved by the agent.

## Stage 12 — October 9, 2026 (macOS)

Stage 00 through Stage 12 are complete within the documented binomial candidate scope. Stage 13 (full solutions and cases) is next. All changes from this Mac session remain local and uncommitted.

### Implemented

- `internal/candidates`: strict, key-free proposal contract; four deterministic finite `(n,p,k)` sets; seven-stage binomial exactly-k builder deriving numeric truth from Go; semantic-review attestation and immutable approval/rejection/retirement transitions. Wording variants retain the existing setting group. Other families/parameters fail closed.
- `internal/storage/candidates.go`: durable candidate records and atomic revision-checked approval audit using existing Stage 05 tables; no migration needed. Existing snapshots/attempts/mastery are untouched.
- `internal/httpapi/candidates.go` and `candidate_generation.go`: session-authenticated, exact-Origin, bounded create/list/review/export endpoints; provider output capped at 12 KB with a 60-second timeout, completion requirement, fixed-parameter enforcement and cancellation. No automatic AI call or activation.
- `internal/httpapi/server.go`, `internal/bank/bank.go`: approved local content reloads at startup and joins future selection; retirement removes all selection indices. Duplicate replacement no longer duplicates module/family index entries.
- `internal/providers/{prompt,anthropic,gemini,openai,chatgpt}.go`, `internal/tutor/types.go`: separate candidate wording prompt over existing adapters/budgets. Provider protocols/billing routes are unchanged; no live AI-generation request was made.
- `cmd/quant-practice/{main,candidates}.go`: generation/import/list/preview/approve/reject/retire/export CLI equivalents, with review revision/name/notes/semantic attestation. CLI changes require an already running app to restart its in-memory bank.
- `web/src/features/candidates`: complete sanitized math preview, source/seed, review findings, approval/rejection/retirement, separate immutable wording drafts, approved-bank download, cancellation/stale response protection and preserved review input on failure. App tab/`p` shortcut; settings close refreshes the selected AI route. No generated content was approved in the user's real database.
- `internal/drill/recap.go`: removed a hardcoded “3/8 or 37.5%” suffix that incorrectly appeared for every numeric answer.
- `web/src/features/practice/PracticeDrill.tsx`: wrap action buttons on narrow screens; the Mac browser check found 381 px content at a 375 px viewport.
- `scripts/test.sh`, `web/playwright.config.ts`, `web/scripts/test-e2e.js`, `tests/e2e/{foundation,candidates}.spec.ts`: native Mac binary/Chromium support; Python 3, lint and gofmt checks; fresh build before browser tests; focused test arguments; isolated data for candidate approval/restart checks.
- Docs: README, REQUIREMENTS, PLAN, QUESTION-BANK, API, TESTING and `schemas/candidate-proposal.schema.json` describe implementation and limits. Approved candidates are immutable; changing wording creates a separate draft/content ID at version 1 with fresh approval. In-place version editing, new-family generation and retirement of source-controlled bundled curriculum are outside this initial workflow.

### Actual checks

- `./scripts/test.sh`: passed natively on macOS (frontend typecheck/lint, 68 unit tests in 14 files, production frontend + native Go build, all 10 Playwright browser tests, gofmt, uncached Go tests across 13 internal packages, go vet). Existing unrelated React `act(...)` warnings remain in older tests; assertions passed.
- `go test -race ./internal/candidates ./internal/storage ./internal/httpapi`: passed. Coverage includes structural/forged key rejection, wrong-story approval attestation, canonical numeric fixtures, stale/duplicate review, persisted audit/restart, immutable old snapshots, no learning writes, session/Origin checks, bounded/incomplete/changed-parameter provider output and cancellation.
- CLI smoke in an automatically cleaned temporary directory: generate, preview, approve synthetic fixture, reject stale approval, export 11 active templates, retire and read persisted audit all passed. No real learner or provider data used.
- Browser failures found during implementation were fixed: textarea's wrapping label accumulated its value and stopped exact label matching; labels now use explicit `for`/`id`. Narrow practice-button overflow fixed. The notes search failure caused by the aborted earlier restart test disappeared on the successful full suite.
- Final `python3 scripts/validate_scaffold.py` and `git diff --check` passed. Browser-generated replacements of older verification screenshots were restored to their committed versions.
- Stopped the earlier local server and reopened the fresh build via `./scripts/start.sh`; homepage returned HTTP 200 at `http://127.0.0.1:63632/`. The app is left running for user review.
- No native Windows/Linux run or live AI-generation test was performed in this Mac session. Existing sign-in/inference evidence remains recorded separately below.

## macOS continuation — October 9, 2026

- Working tree was clean on arrival; Stage 10/11 changes are already committed (including `99114fa`, titled `stage 11 complete`). The older uncommitted-work statement below is stale.
- Installed Go 1.27.2 with Homebrew; existing Node v26.7.0/npm 11.19.0. Installed frontend dependencies with `npm ci`.
- Made `scripts/start.sh`, `scripts/build.sh`, and `scripts/test.sh` executable. Removed an unused `errors` import from `internal/auth/dpapi_vault_other.go` that prevented non-Windows compilation.
- Frontend production build and native `go build -o bin/quant-practice ./cmd/quant-practice` passed. `npm run lint` passed; targeted MathMarkdown/ChatGPTPlanPanel tests passed (11 tests). `go test ./internal/auth ./internal/siwc` passed natively on macOS.
- Launched via `./scripts/start.sh`, which opened the browser and loaded 10 approved templates. Mac runtime data lives in `~/Library/Application Support/quant-methods-practice`.
- **Live (user-run, October 9):** user reports successful ChatGPT sign-in and sign-out on macOS. Together with the October 5 user-run streamed tutor answer over the `chatgpt` plan route, this closes Stage 11's sign-in/inference/sign-out gate. Model used was not recorded; no additional Mac inference or independent server-side revocation audit is claimed. Full verification suite was not rerun in this session.
- At this launch checkpoint Stage 00 through Stage 11 were complete; Stage 12 was next and is now implemented in the record above.
- Updated README, PLAN, PROVIDERS and this handoff to reflect the user-run gate. `python3 scripts/validate_scaffold.py` and `git diff --check` passed after these documentation changes.

## Historical state - October 5, 2026

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

1. Continue with Stage 14 (exam mode) per [PLAN.md](../PLAN.md): inspect immutable snapshots, deadlines, aid/feedback withholding, restart and report policy before implementation. Stage 13 workflow and limits are in [WORKSHEETS.md](WORKSHEETS.md).
2. Mac launch fixes, Stage 12/13 implementation and October 9 documentation remain local changes; Stage 10/11 implementation is already committed.
3. The curated API-key model defaults in `internal/providers/catalog.go` (Claude 3.5, Gemini 1.5, GPT-4o) are dated and should be refreshed against current catalogs when that work is taken up.

## Unresolved external gates

Actual course syllabus/slides/notation; live provider API keys for optional non-mock testing; continuous distribution/test/regression course details.

# Content backlog: question coverage by learning objective

Created October 11, 2026. This is the durable, machine-independent plan for broadening the approved bank so that every published module learning objective has **at least five new approved questions**. It implements PLAN Stage 20 (content quality and ongoing breadth) and is driven by the learner-supplied module objectives recorded in [COURSE-MAP.md](COURSE-MAP.md).

## Rules for every item

- Each item follows the [QUESTION-BANK.md](QUESTION-BANK.md) contract: strict template JSON in `curriculum/approved/`, a matching `curriculum/reviews/*.review.json`, a registered family with reviewed rules and tests, and the Go engine as the only source of answer keys (Invariant 2).
- Parameters below are **proposals**. The engine derives every key; if a proposed number turns out awkward (unrounded key, impossible event, violated condition), change the parameters and record that in the review notes.
- Assistant review counts as approval only under an explicit user delegation, recorded honestly in `approval.delegation_basis` (Invariant 5).
- Vary `setting_group` and wording within each row so the five questions are not cosmetic copies of each other (Invariant 8). Each row should include at least one misconception-driven distractor set and, where Excel applies, the matching Excel formula in the explanation (PLAN backlog item 1).
- Tick a box only when the approved JSON and review file are committed and `scripts/test.sh` (or `test.ps1`) passes. Update the status table counts in the same change.
- Items marked **[confirm]** depend on a course convention that is not yet verified. Author them with the most common introductory-textbook convention, state it in `assumptions`, and add the question to the "Open course conventions" list below if it isn't already there.

## Deadlines that shape the order

| Due | Item |
|---|---|
| Sun Nov 1, 2026 | Module 6 quiz |
| Sun Nov 8, 2026 | Final exam (cumulative) |

Work in the batch order below: Module 6 first, then cheap rows where the engine is already ready, then the larger engine additions.

## Status

"Existing" counts approved questions present on October 11, 2026. "New" counts questions approved under this backlog. Target is ≥ 5 new per row.

| Row | Objective | Family | Engine | Existing | New |
|---|---|---|---|---|---|
| 1.1 | Role of statistics in business | `concept_check` | ready (ENG-CONCEPT) | 0 | 5/5 |
| 1.2 | Measures of center and variability | `descriptive_stats` | ready (ENG-DESC) | 0 | 6/6 |
| 1.3 | Probability rules | `set_probability` (+ ext) | ready (ENG-PROB-EXT) | 3 | 6/6 |
| 2.1 | Discrete RV and its probability distribution | `discrete_rv` | ready (ENG-PMF) | 0 | 6/6 |
| 2.2 | E[X], Var(X), and transformations | `linear_combination` (+ affine) | ready (ENG-AFFINE) | 1 | 6/6 |
| 2.3 | Choosing and using binomial vs. Poisson | `binomial_pmf`, `poisson_pmf`, `concept_check` | ready | 6 | 5/5 |
| 3.1 | Discrete vs. continuous distributions | `concept_check`, `uniform_dist` | ready | 1 | 5/5 |
| 3.2 | Normal distribution | `normal_dist` | ready | 1 | 6/6 |
| 3.3 | Exponential distribution | `exponential_dist` | ready | 1 | 6/6 |
| 4.1 | Central Limit Theorem (mean and proportion) | `sampling_distribution` | ready (ENG-PROP-SAMPLING) | 0 | 6/6 |
| 4.2 | Sample size and standard error | `sampling_distribution`, `sample_size` | ready (ENG-SAMPLESIZE) | 0 | 5/5 |
| 4.3 | Confidence interval for mean and proportion | `confidence_interval`, `concept_check` | ready | 1 | 6/6 |
| 4.4 | Required sample size for a margin of error | `sample_size`, `concept_check` | ready (ENG-SAMPLESIZE) | 0 | 5/5 |
| 5.1 | Hypothesis-testing steps; formulating H₀/H₁ | `hypothesis_test`, `concept_check` | ready | 0 | 5/5 |
| 5.2 | Type I and Type II errors | `concept_check` | ready (ENG-CONCEPT) | 0 | 5/5 |
| 5.3 | Test for a mean or proportion with p-value | `hypothesis_test` | ready | 1 | 6/6 |
| 5.4 | One-sample vs. two-sample tests | `two_sample_test`, `concept_check` | ready (ENG-TWOSAMPLE) | 0 | 5/5 |
| 6.1 | Regression equation and prediction | `linear_regression`, `concept_check` | ready | 1 | 5/5 |
| 6.2 | Model significance and fit (p-value, R², s_e) | `regression_output` | ready (ENG-REGOUT) | 0 | 5/5 |
| 6.3 | Slope confidence interval and predictor significance | `regression_output` | ready (ENG-REGOUT) | 0 | 5/5 |
| 6.4 | Multiple regression and predictor selection | `regression_output` | ready (ENG-REGOUT) | 0 | 6/6 |
| 6.5 | Common regression problems | `concept_check` | ready (ENG-CONCEPT) | 0 | 6/6 |

Total planned: **121 new questions, all approved**, plus the 16 original questions. Active approved bank: **137 questions**. No catalog questions remain to be authored.


## Dispersed execution strategy & wave schedule

Rather than a waterfall order that finishes one module entirely before touching later ones (which would starve Modules 5, 6, 1, and 2), question development is organized into **interleaved, round-robin waves**.

Building in dispersed waves ensures that:
1. **Immediate breadth**: Every authoring sprint distributes new questions across all active and upcoming modules rather than clustering in a single topic. Wave 1 spread 20 questions across Modules 4, 5, 6, 1, and 2.
2. **Zero engine blockers**: Waves 1–3 drew exclusively on engines and families that are already fully implemented, tested, and registered (`confidence_interval`, `sample_size`, `concept_check`, `hypothesis_test`, `linear_regression`, `regression_output`, `set_probability`, `binomial_pmf`, `poisson_pmf`).
3. **Cumulative exam readiness**: The learner gets varied question styles across the entire course syllabus right away, supporting comprehensive practice drills and mock exams.

### Wave overview

**Current completion: Wave 4 20/20, Wave 5 13/13; all 121 planned additions are approved and all 22 objective targets are met.** The earlier-wave row below is a historical milestone. [Complete final-sprint review](REMAINING-CONTENT-REVIEW.md). Run `python scripts/check_content_coverage.py` to verify the catalog against approved templates and review records.

The final 20 checklist boxes remain open until this local work is committed, as required by the checklist rule above. Content authoring, delegated approval and activation are complete; this checkbox distinction concerns repository delivery only.

| Wave | Questions | Module distribution | Primary families | Milestone / Goal |
|---|---|---|---|---|
| **Completed** | 88 | Mod 3: 17, Mod 4: 22, Mod 5: 15, Mod 6: 19, Mod 1: 9, Mod 2: 6 | `concept_check`, `normal_dist`, `exponential_dist`, `uniform_dist`, `sampling_distribution`, `sample_size`, `confidence_interval`, `hypothesis_test`, `linear_regression`, `regression_output`, `set_probability`, `binomial_pmf`, `poisson_pmf` | **Waves 1, 2, & 3 complete (60/60 wave questions)**; **Module 3 complete (17/17)**; **Module 4 complete (22/22)**; Rows 1.1 (5/5), 2.3 (5/5), 4.1 (6/6), 4.2 (5/5), 4.3 (6/6), 4.4 (5/5), 5.1 (5/5), 5.2 (5/5), 6.1 (5/5), 6.3 (5/5) complete. Bank at **104 approved questions**. |
| **Wave 4** | 20 | Mod 5: 3, Mod 6: 6, Mod 1: 4, Mod 2: 7 | `two_sample_test`, `regression_output`, `descriptive_stats`, `discrete_rv`, `linear_combination` | Implements ENG-TWOSAMPLE, ENG-DESC, ENG-PMF, ENG-AFFINE. Rows 5.3, 6.2, 6.4, 1.3 complete. |
| **Wave 5** | 13 | Mod 5: 3, Mod 6: 2, Mod 1: 3, Mod 2: 5 | Same as Wave 4 | **All modules 100% complete** (121 new questions approved). Final exam bank ready. |

---

### Tonight's 20-question target: Dispersed Wave 1

Author these 20 items across the 5 modules tonight. Every family needed is already active and verified in the Go engine.

| # | Mod | Row | File Slug | Family | Topic / Core Concept |
|---|---|---|---|---|---|
| 1 | 4 | 4.3 | `ci-mean-z-known-sigma` | `confidence_interval` | $\sigma = 12$, $n = 49$, $\bar{x} = 150$, 95% $z$-confidence interval for mean; `=CONFIDENCE.NORM` |
| 2 | 4 | 4.3 | `ci-proportion-recommend` | `confidence_interval` | $x = 132$, $n = 400$, 95% Wilson/Wald confidence interval for proportion |
| 3 | 4 | 4.4 | `n-mean-margin-five` | `sample_size` | Required $n = \lceil (z\sigma/E)^2 \rceil$ for target mean margin ($\sigma = 25, E = 5, 95\% \implies n = 97$) |
| 4 | 4 | 4.4 | `n-proportion-conservative` | `sample_size` | Conservative proportion sample size using $p = 0.50$, $E = 0.04$, 90% confidence |
| 5 | 5 | 5.1 | `hyp-formulate-underfill` | `concept_check` | Product fill claim 16 oz; suspect underfill; formulate $H_0: \mu = 16$ vs $H_1: \mu < 16$ |
| 6 | 5 | 5.2 | `errors-define-in-context` | `concept_check` | Quality control rejecting good lot (Type I) vs accepting bad lot (Type II) |
| 7 | 5 | 5.3 | `hyp-z-mean-two-tailed` | `hypothesis_test` | Known $\sigma = 0.05$, $n = 36$, $\bar{x} = 2.02$, two-tailed test $H_0: \mu = 2.00$, $\alpha = 0.05$, $p$-value |
| 8 | 5 | 5.3 | `hyp-t-mean-right-tailed` | `hypothesis_test` | Sample $n = 20$, $\bar{x} = 53.1$, $s = 6.2$, right-tailed $t$-test $H_0: \mu = 50$, $df = 19$, `=T.DIST.RT` |
| 9 | 6 | 6.1 | `reg-fit-raw-points-price-units` | `linear_regression` | 6 raw $(x, y)$ coordinate pairs; compute $b_0, b_1$, regression equation, point prediction |
| 10 | 6 | 6.2 | `regout-r2-from-anova` | `regression_output` | ANOVA table: SSR and SST; compute $R^2 = \text{SSR}/\text{SST}$, percentage of variance explained |
| 11 | 6 | 6.3 | `regout-slope-t-pvalue` | `regression_output` | Read slope $b_1 = 2.4, SE = 0.8$; compute $t = 3.0, df = 23$, two-tailed $p$-value, statistical significance |
| 12 | 6 | 6.4 | `mreg-which-predictors-significant` | `regression_output` | Multiple regression table with 3 predictors ($p = 0.001, 0.04, 0.62$); select variables at $\alpha = 0.05$ |
| 13 | 6 | 6.5 | `regprob-multicollinearity` | `concept_check` | High $R^2$ and significant model $F$, but no individual $t$-tests significant; diagnose multicollinearity |
| 14 | 1 | 1.1 | `concept-descriptive-vs-inferential` | `concept_check` | Census summary (120 stores) vs sample estimation (40 sampled stores); descriptive vs inferential |
| 15 | 1 | 1.1 | `concept-parameter-vs-statistic` | `concept_check` | Survey 500 of 20,000 customers (62% satisfied); identify population, sample, parameter, statistic |
| 16 | 1 | 1.3 | `prob-multiplication-independent-suppliers` | `set_probability` | Independent suppliers on-time probabilities 0.90 and 0.85; multiplication rule $P(A \cap B)$ |
| 17 | 1 | 1.3 | `prob-at-least-one-server-failure` | `set_probability` | 3 independent servers fail with $P = 0.05$; complement rule $P(\text{at least one}) = 1 - P(\text{none})$ |
| 18 | 2 | 2.3 | `choose-binomial-or-poisson` | `concept_check` | Four business scenarios; distinguish fixed-trial count (Binomial) vs continuous-interval rate (Poisson) |
| 19 | 2 | 2.3 | `binomial-at-least-three-defects` | `binomial_pmf` | $n = 12, p = 0.15$; tail probability $P(X \ge 3) = 1 - P(X \le 2)$; `=1 - BINOM.DIST(2, 12, 0.15, TRUE)` |
| 20 | 2 | 2.3 | `poisson-rate-rescaling-quarter-hour` | `poisson_pmf` | Rate 6 calls/hour; rescale to 15-minute interval ($\lambda = 1.5$); $P(X \le 1)$; unscaled-$\lambda$ distractor |

Record each session's progress in [HANDOFF.md](HANDOFF.md).

## Engine prerequisites

Each one is a new or extended family. Each needs rule versioning, parameter validation, Go unit tests against hand- or table-verified values, and a review record before any question in it is activated.

| ID | Family | Scope | Open conventions |
|---|---|---|---|
| ENG-CONCEPT | `concept_check` (new) | Choice-only stages with reviewed option keys and misconception IDs; no numeric derivation. Validator enforces exactly one correct option per stage, a misconception and hint on every distractor, and non-empty explanations. Keys are reviewed content, so approval review carries the full burden. | None |
| ENG-DESC | `descriptive_stats` (new) | Mean, median, mode, range, population vs. sample variance/SD, coefficient of variation, z-score, empirical rule, weighted mean, quartiles/IQR, outlier fences. Excel: `AVERAGE`, `MEDIAN`, `STDEV.S`, `STDEV.P`, `VAR.S`, `QUARTILE.INC`/`.EXC`. | Quartile method **[confirm]**; whether Chebyshev is in scope |
| ENG-PMF | `discrete_rv` (new) | Validate a finite PMF table (non-negative, sums to 1), solve for a missing probability, CDF and interval probabilities, E[X], Var(X), SD. Excel: `SUMPRODUCT`. | None |
| ENG-AFFINE | extend `linear_combination` | `affine` operation for E[aX+b] and Var(aX+b) = a²Var(X); sum of n i.i.d. copies vs. nX; difference of independent variables; covariance/correlation case (engine has `CovarianceFromCorrelation`). | None |
| ENG-PROB-EXT | extend `set_probability` | Multiplication rule, independence test, contingency-table joint/marginal/conditional, "at least one" via complement, two-stage tree with Bayes. | Bayes depth **[confirm]** |
| ENG-PROP-SAMPLING | extend `sampling_distribution` | Sampling distribution of p̂ with SE = √(p(1−p)/n) and np, n(1−p) ≥ 10 check. Check first whether the family already accepts proportion parameters; `inference.go` has the SE helper. | Success-failure threshold 5 vs. 10 **[confirm]** |
| ENG-SAMPLESIZE | `sample_size` (new) | n = ⌈(z·σ/E)²⌉ for a mean; n = ⌈z²·p(1−p)/E²⌉ for a proportion, with conservative p = 0.5; SE as a function of n. Always round up. | z vs. t when planning (introductory default: z) |
| ENG-TWOSAMPLE | `two_sample_test` (new) | Paired t; independent two-mean t; two-proportion z with pooled p̂. Excel: `T.TEST` types 1/2/3, Data Analysis two-sample output. | Pooled-variance vs. Welch **[confirm]** |
| ENG-REGOUT | `regression_output` (new) | Work from an Excel-style output table rather than refitting: equation and prediction, R² = SSR/SST, adjusted R², s_e = √MSE, F = MSR/MSE and Significance F, coefficient t = b/SE with df = n−k−1, p-value, coefficient CI b ± t*·SE. Consistency checks: SST = SSR + SSE, t² = F when k = 1. Supports k = 1 to 4 predictors. | Whether Excel Data Analysis is the course software **[confirm]**; dummy variables in scope **[confirm]** |

## Question items

ID = proposed file slug in `curriculum/approved/`. Key parameters are proposals.

### Module 1 — Descriptive measures and probability

#### 1.1 Role of statistics in business (`concept_check`)

- [x] **[Wave 1]** `concept-descriptive-vs-inferential`: retailer summarizes last quarter's sales at all 120 stores vs. estimates the chain mean from 40 sampled stores; classify each.
- [x] **[Wave 1]** `concept-parameter-vs-statistic`: survey of 500 of 20,000 customers finds 62% satisfied; identify population, sample, parameter, statistic.
- [x] **[Wave 2]** `concept-sampling-bias-online-poll`: brand runs a voluntary poll on its loyalty app; which conclusion is unjustified and why (selection bias).
- [x] **[Wave 2]** `concept-data-types-business`: classify region, satisfaction rating (1–5), monthly revenue, and units returned as categorical/numerical and discrete/continuous.
- [x] **[Wave 3]** `concept-cross-sectional-vs-time-series`: dashboard of 2025 sales by store vs. monthly sales for one store; identify the data type and one valid question for each.

#### 1.2 Measures of center and variability (`descriptive_stats`)

- [x] **[Wave 4]** `desc-mean-median-outlier`: monthly sales of 7 stores, one flagship outlier; compute mean and median and decide which better describes a typical store.
- [x] **[Wave 4]** `desc-sample-sd-delivery-times`: 6 delivery times; sample SD (n−1) with a population-divisor distractor; `STDEV.S` vs. `STDEV.P`.
- [x] **[Wave 4]** `desc-cv-compare-stocks`: two stocks with different mean returns and SDs; compare relative risk with the coefficient of variation.
- [ ] **[Wave 5]** `desc-zscore-empirical-rule`: salesperson's quarterly sales vs. team mean/SD; z-score, then the share within ±2 SD by the empirical rule.
- [ ] **[Wave 5]** `desc-weighted-mean-price`: average selling price across a three-product mix weighted by units sold; unweighted-average distractor.
- [ ] **[Wave 5]** `desc-quartiles-iqr-fences` **[confirm]**: call-handling times; Q1, Q3, IQR, 1.5·IQR fences, and flag an outlier.

#### 1.3 Probability rules (`set_probability` + ENG-PROB-EXT)

- [x] **[Wave 1]** `prob-multiplication-independent-suppliers`: two independent suppliers on time with probability 0.90 and 0.85; P(both on time).
- [x] **[Wave 2]** `prob-independence-check`: P(A) = 0.4, P(B) = 0.5, P(A∩B) = 0.20; decide independence, then contrast with P(A∩B) = 0.25.
- [x] **[Wave 3]** `prob-contingency-age-purchase`: 2×3 table of age group × purchased; joint, marginal and conditional probabilities.
- [x] **[Wave 3]** `prob-disjoint-vs-independent`: misconception item; show that disjoint events with positive probabilities cannot be independent.
- [x] **[Wave 1]** `prob-at-least-one-server-failure`: 3 independent servers each fail with probability 0.05; P(at least one fails) via the complement.
- [x] **[Wave 4]** `prob-two-plant-defect-tree`: plants supply 60%/40% with defect rates 2%/5%; P(defective), then P(plant B | defective).

### Module 2 — Discrete random variables and distributions

#### 2.1 Discrete RV and its probability distribution (`discrete_rv`)

- [x] **[Wave 4]** `drv-valid-pmf-missing-value`: cars sold per day, 0–4 with one missing probability; solve for it and check validity.
- [x] **[Wave 4]** `drv-expected-cars-sold`: same dealership PMF; E[X] with `SUMPRODUCT`.
- [x] **[Wave 4]** `drv-variance-cars-sold`: Var(X) and SD; distractor uses Σx²p without subtracting μ².
- [ ] **[Wave 5]** `drv-compare-projects-risk`: two investment payoff tables with equal expected profit but different SD; choose using risk.
- [ ] **[Wave 5]** `drv-insurance-break-even`: policy payout distribution; expected payout and the break-even premium.
- [ ] **[Wave 5]** `drv-cumulative-from-table`: P(X ≥ 2) and P(1 ≤ X ≤ 3) from a demand PMF.

#### 2.2 Expected value, variance, transformations (`linear_combination` + ENG-AFFINE)

- [x] **[Wave 4]** `affine-profit-from-units`: profit = 50X − 200 with E[X] = 30, SD 6; E and SD of profit (the −200 does not affect SD).
- [x] **[Wave 4]** `affine-currency-conversion`: euro-denominated costs converted at 1.10; mean and SD scale, variance by 1.21.
- [x] **[Wave 4]** `lincomb-total-two-product-lines`: independent daily revenues; mean and SD of the total (adding SDs is a distractor).
- [x] **[Wave 4]** `lincomb-difference-independent`: Var(X − Y) for independent demand at two stores; subtracting variances is a distractor.

- [ ] **[Wave 5]** `lincomb-weekly-demand-sum-vs-scale`: 5 i.i.d. days vs. 5·X; variance 5σ² vs. 25σ².
- [ ] **[Wave 5]** `lincomb-portfolio-correlated`: 60/40 portfolio with ρ = 0.3; portfolio variance with the covariance term.

#### 2.3 Choosing and using binomial vs. Poisson

- [x] **[Wave 1]** `choose-binomial-or-poisson`: (`concept_check`) four scenarios, e.g. 20 surveyed customers who renew vs. website errors per hour; choose a model and justify it.
- [x] **[Wave 1]** `binomial-at-least-three-defects`: n = 12, p = 0.15, P(X ≥ 3) via complement and `BINOM.DIST(2,12,0.15,TRUE)`.
- [x] **[Wave 2]** `binomial-mean-sd-shipments`: n = 200, p = 0.02; expected defects and SD.
- [x] **[Wave 1]** `poisson-rate-rescaling-quarter-hour`: 6 calls/hour; P(X ≤ 1) in 15 minutes (λ = 1.5); unscaled-λ distractor.
- [x] **[Wave 2]** `binomial-condition-violated-small-lot`: (`concept_check`) sampling 5 of 20 items without replacement; which binomial condition fails.

### Module 3 — Continuous distributions

#### 3.1 Discrete vs. continuous (`concept_check`, `uniform_dist`)

- [x] `contvdisc-classify-variables`: wait time, complaints per day, package weight, defects per batch.
- [x] `contvdisc-point-probability-zero`: P(X = 5) for a continuous wait time vs. a Poisson count.
- [x] `contvdisc-area-is-probability`: uniform delivery window 10–40 minutes; density height and P(15 < X < 25).
- [x] `contvdisc-strict-vs-nonstrict`: P(X < a) = P(X ≤ a) for continuous X but not for a binomial counterexample.
- [x] `contvdisc-match-model`: match each business variable to normal, exponential, binomial or Poisson.

#### 3.2 Normal distribution (`normal_dist`)

- [x] `normal-between-package-weight`: N(500, 8) grams; P(490 < X < 510).
- [x] `normal-inverse-bonus-cutoff`: sales N(80k, 12k); top-5% cutoff with `NORM.INV(0.95, …)`.
- [x] `normal-warranty-lower-percentile`: tire life N(60k, 8k) miles; warranty mileage so only 2% fail before it.
- [x] `normal-compare-standardized-scores`: two regions with different means and SDs; which manager performed relatively better.
- [x] `normal-inventory-service-level`: daily demand N(200, 30); stock level that meets demand on 90% of days.
- [x] `normal-empirical-vs-exact`: P(μ − 2σ < X < μ + 2σ) by empirical rule (≈ 0.95) vs. exact (0.9545).

#### 3.3 Exponential distribution (`exponential_dist`)

- [x] `exponential-arrival-within-two`: mean gap 4 minutes; P(next arrival within 2 minutes) with `EXPON.DIST(2, 0.25, TRUE)`.
- [x] `exponential-memoryless-wait`: already waited 5 minutes; P(more than 3 more minutes) equals the unconditional value.
- [x] `exponential-rate-vs-mean-param`: misconception item; λ = 0.10 per minute vs. mean 10 minutes in Excel.
- [x] `exponential-percentile-service`: time by which 90% of service calls are complete (quantile).
- [x] `exponential-poisson-link`: 12 arrivals/hour ⇒ mean gap 5 minutes; P(gap > 10 minutes).
- [x] `exponential-between-two-times`: component life mean 1,000 hours; P(500 < X < 1,500).

### Module 4 — Sampling distributions and confidence intervals

#### 4.1 Central Limit Theorem, mean and proportion (`sampling_distribution` + ENG-PROP-SAMPLING)

- [x] `clt-mean-order-value`: μ = 80, σ = 24, n = 36; P(x̄ > 86).
- [x] `clt-skewed-service-times`: exponential population, mean 6, n = 64; justify the normal approximation and find P(x̄ < 4.8).
- [x] `clt-individual-vs-mean`: same population; P(one bag > 45) vs. P(x̄ > 45), the classic confusion.
- [x] `clt-small-n-not-justified`: (`concept_check` stage) n = 9 from a strongly skewed population; why the CLT approximation is not warranted.
- [x] `clt-proportion-coupon-users`: p = 0.30, n = 200; P(p̂ > 0.35) after a success-failure check.
- [x] `clt-finite-population-correction`: n = 64 drawn from N = 400 accounts; apply the FPC since n/N > 0.05.

#### 4.2 Sample size and standard error (`sampling_distribution`, `sample_size`)

- [x] `se-quadruple-n-halves`: σ = 20, n from 25 to 100; SE from 4 to 2.
- [x] `se-n-for-target-se`: σ = 15; smallest n with SE ≤ 2.
- [x] `se-proportion-depends-on-p`: n = 400; SE at p = 0.5 vs. p = 0.1.
- [x] `se-what-n-changes`: (`concept_check`) larger n narrows the sampling distribution but does not change μ or σ.
- [x] `se-survey-design-compare`: n = 100 vs. n = 900; ratio of margins of error (3×).

#### 4.3 Confidence intervals for mean and proportion (`confidence_interval`)

- [x] **[Wave 1]** `ci-mean-z-known-sigma`: σ = 12, n = 49, x̄ = 150, 95%; `CONFIDENCE.NORM`.
- [x] **[Wave 2]** `ci-mean-t-small-sample-90`: n = 10, x̄ = 42.3, s = 5.1, 90%; df = 9.
- [x] **[Wave 1]** `ci-proportion-recommend`: 132 of 400 customers recommend; 95% interval for p.
- [x] **[Wave 2]** `ci-interpretation-statement`: choose the correct interpretation of a 95% interval; "95% probability μ is in this interval" is a distractor.
- [x] **[Wave 3]** `ci-level-vs-width`: same data at 90%, 95%, 99%; order the widths and explain the trade-off.
- [x] **[Wave 3]** `ci-plausible-claim`: interval (146.6, 153.4); is a claimed μ = 155 plausible?

#### 4.4 Required sample size (`sample_size`)

- [x] **[Wave 1]** `n-mean-margin-five`: σ = 25, E = 5, 95% ⇒ n = ⌈96.04⌉ = 97.
- [x] **[Wave 2]** `n-proportion-pilot`: pilot p̂ = 0.20, E = 0.03, 95%.
- [x] **[Wave 1]** `n-proportion-conservative`: no prior estimate, E = 0.04, 90%; use p = 0.5.
- [x] **[Wave 2]** `n-halve-margin-quadruple`: halving E multiplies n by 4.
- [x] **[Wave 3]** `n-round-up-misconception`: 96.04 must round up to 97, not to 96.

### Module 5 — Hypothesis testing

#### 5.1 Steps and formulating H₀/H₁ (`hypothesis_test`, `concept_check`)

- [x] **[Wave 1]** `hyp-formulate-underfill`: label claims 16 oz; consumer group suspects underfill; H₀: μ = 16 (or ≥ 16), H₁: μ < 16.
- [x] **[Wave 2]** `hyp-formulate-right-tailed-sales`: new display claimed to raise mean daily sales above 50.
- [x] **[Wave 3]** `hyp-formulate-two-tailed-calibration`: machine calibrated to 2.00 cm; detect drift in either direction.
- [x] **[Wave 3]** `hyp-formulate-proportion`: marketing claims more than 40% of customers prefer the new package.
- [x] **[Wave 2]** `hyp-steps-and-conclusion`: order the testing steps; decision rule p ≤ α; "fail to reject" vs. "accept H₀".

#### 5.2 Type I and Type II errors (`concept_check`)

- [x] **[Wave 1]** `errors-define-in-context`: quality control rejects a good lot (Type I) vs. accepts a bad lot (Type II).
- [x] **[Wave 2]** `errors-which-is-worse`: food-safety recall vs. ad-campaign test; which error is costlier and how that drives α.
- [x] **[Wave 2]** `errors-alpha-beta-tradeoff`: lowering α from 0.05 to 0.01 with n fixed increases β.
- [x] **[Wave 3]** `errors-larger-n-reduces-beta`: increasing n reduces β at fixed α.
- [x] **[Wave 3]** `errors-identify-after-the-fact`: H₀ rejected but later shown true; name the error.

#### 5.3 Test for a mean or proportion with p-value (`hypothesis_test`)

- [x] **[Wave 1]** `hyp-z-mean-two-tailed`: σ known = 0.05, n = 36, x̄ = 2.02, H₀: μ = 2.00, α = 0.05.
- [x] **[Wave 1]** `hyp-t-mean-right-tailed`: n = 20, x̄ = 53.1, s = 6.2, H₀: μ = 50; `T.DIST.RT`.
- [x] **[Wave 2]** `hyp-proportion-left-tailed`: claim 20% defective; 28 of 200 sampled; test p < 0.20.
- [x] **[Wave 3]** `hyp-pvalue-interpretation`: choose the correct meaning of p = 0.03; "probability H₀ is true" is a distractor.
- [x] **[Wave 3]** `hyp-alpha-changes-decision`: p = 0.032; reject at α = 0.05, not at α = 0.01.
- [x] **[Wave 4]** `hyp-excel-tail-function`: choose `T.DIST.RT`, `T.DIST`, or `T.DIST.2T` for the tail stated in H₁.

#### 5.4 One-sample vs. two-sample tests (`two_sample_test`, `concept_check`)

- [x] **[Wave 4]** `twosample-choose-design`: (`concept_check`) classify scenarios as one-sample, independent two-sample, or paired.

- [ ] **[Wave 5]** `twosample-paired-training`: before/after weekly sales for 8 reps; paired t on the differences.
- [ ] **[Wave 5]** `twosample-independent-means-layouts` **[confirm]**: average basket under two store layouts; independent two-sample t.
- [ ] **[Wave 5]** `twosample-two-proportion-ads`: conversion 48/600 vs. 72/600 for two ad versions; pooled two-proportion z.
- [ ] **[Wave 4]** `twosample-read-excel-output`: interpret an Excel t-Test two-sample output block (t Stat, P(T<=t) one-tail/two-tail, t Critical).

### Module 6 — Regression analysis

#### 6.1 Regression equation and prediction (`linear_regression`)

- [x] **[Wave 1]** `reg-fit-raw-points-price-units`: 6 (price, units) pairs; engine fits b₀, b₁, then predict.
- [x] **[Wave 2]** `reg-interpret-slope-intercept`: interpret b₁ in units and judge whether b₀ is meaningful.
- [x] **[Wave 2]** `reg-residual-observed-point`: compute ŷ and residual e = y − ŷ for one observed store.
- [x] **[Wave 3]** `reg-extrapolation-warning`: x range 2–10; prediction at x = 25 is extrapolation.
- [x] **[Wave 3]** `reg-predict-from-excel-coefficients`: read Intercept and slope from an output table and predict.

#### 6.2 Model significance and fit (`regression_output`)

- [x] **[Wave 1]** `regout-r2-from-anova`: SSR and SST from the ANOVA table; R² and its interpretation (no causation).
- [x] **[Wave 2]** `regout-se-from-sse`: SSE, n = 20; s_e = √(SSE/(n−2)) and its meaning in y units.
- [x] **[Wave 3]** `regout-significance-f`: interpret Significance F = 0.0004 at α = 0.05.
- [x] **[Wave 3]** `regout-r-from-r2-sign`: R² = 0.64 with negative slope ⇒ r = −0.8.
- [ ] **[Wave 4]** `regout-compare-models-fit`: two candidate models; compare with s_e and adjusted R².

#### 6.3 Slope confidence interval and predictor significance (`regression_output`)

- [x] **[Wave 1]** `regout-slope-t-pvalue`: b₁ = 2.4, SE = 0.8, n = 25; t = 3.0, df = 23, p-value.
- [x] **[Wave 2]** `regout-slope-ci-includes-zero`: Lower 95% and Upper 95% columns; decide significance from whether 0 is inside.
- [x] **[Wave 2]** `regout-slope-ci-interpretation`: interpret the slope CI in business units.
- [x] **[Wave 3]** `regout-t-squared-equals-f`: simple regression output; verify t² = F.
- [x] **[Wave 3]** `regout-compute-slope-ci`: b₁ ± t*·SE with `T.INV.2T(0.05, df)`.

#### 6.4 Multiple regression and predictor selection (`regression_output`)

- [x] **[Wave 2]** `mreg-write-equation-predict`: output with 3 predictors (ad spend, price, store size); write ŷ and predict.
- [x] **[Wave 3]** `mreg-interpret-coefficient-ceteris-paribus`: interpret the price coefficient holding the others constant.
- [x] **[Wave 1]** `mreg-which-predictors-significant`: p-values 0.001, 0.04, 0.62; which to keep and which to drop at α = 0.05.
- [ ] **[Wave 4]** `mreg-adjusted-r2-weak-predictor`: adding a weak predictor raises R² but lowers adjusted R².
- [ ] **[Wave 4]** `mreg-overall-f-vs-individual-t`: significant F with one non-significant t; interpret both.
- [ ] **[Wave 4]** `mreg-dummy-variable` **[confirm]**: a region dummy (1 = urban); interpret its coefficient as a shift in the intercept.

#### 6.5 Common regression problems (`concept_check`)

- [x] **[Wave 1]** `regprob-multicollinearity`: high R² and significant F but no significant t; correlated predictors; diagnose.
- [x] **[Wave 3]** `regprob-heteroscedasticity`: fan-shaped residual plot; which assumption is violated.
- [ ] **[Wave 4]** `regprob-nonlinearity`: curved residual pattern; the model form is wrong.
- [ ] **[Wave 4]** `regprob-autocorrelation`: monthly time-series residuals alternate in runs; independence violated.
- [ ] **[Wave 5]** `regprob-outlier-influence`: one store far from the others changes the slope sign; what to do.
- [ ] **[Wave 5]** `regprob-omitted-variable-causation`: ice-cream sales vs. sunscreen; omitted variable and association vs. causation.

## Open course conventions

Confirm these from Canvas, the syllabus, or the instructor's worked examples. Record the answer in [COURSE-MAP.md](COURSE-MAP.md) and update affected `assumptions`:

1. Quartile/percentile method (Excel `QUARTILE.INC` vs. `.EXC` vs. a textbook method).
2. Whether Bayes' theorem and probability trees are assessed in Module 1.
3. Success-failure threshold for proportion normal approximations (np ≥ 5 or ≥ 10).
4. z vs. t when σ is unknown but n is large.
5. Pooled-variance vs. Welch for independent two-sample t.
6. Whether regression is taught through Excel Data Analysis output (assumed for ENG-REGOUT).
7. Whether dummy variables are covered in Module 6.
8. Rounding conventions for answers (z to 2 decimals from tables vs. exact software values).

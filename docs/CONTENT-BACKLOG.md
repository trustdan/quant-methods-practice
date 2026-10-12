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
| 1.1 | Role of statistics in business | `concept_check` | ENG-CONCEPT | 0 | 0/5 |
| 1.2 | Measures of center and variability | `descriptive_stats` | ENG-DESC | 0 | 0/6 |
| 1.3 | Probability rules | `set_probability` (+ ext) | ready / ENG-PROB-EXT | 3 | 0/6 |
| 2.1 | Discrete RV and its probability distribution | `discrete_rv` | ENG-PMF | 0 | 0/6 |
| 2.2 | E[X], Var(X), and transformations | `linear_combination` (+ affine) | ENG-AFFINE | 1 | 0/6 |
| 2.3 | Choosing and using binomial vs. Poisson | `binomial_pmf`, `poisson_pmf`, `concept_check` | ready / ENG-CONCEPT | 6 | 0/5 |
| 3.1 | Discrete vs. continuous distributions | `concept_check`, `uniform_dist` | ready (ENG-CONCEPT) | 1 | 5/5 |
| 3.2 | Normal distribution | `normal_dist` | ready | 1 | 6/6 |
| 3.3 | Exponential distribution | `exponential_dist` | ready | 1 | 6/6 |
| 4.1 | Central Limit Theorem (mean and proportion) | `sampling_distribution` | ready / ENG-PROP-SAMPLING | 0 | 6/6 |
| 4.2 | Sample size and standard error | `sampling_distribution`, `sample_size` | ENG-SAMPLESIZE (ready) | 0 | 5/5 |
| 4.3 | Confidence interval for mean and proportion | `confidence_interval` | ready | 1 | 0/6 |
| 4.4 | Required sample size for a margin of error | `sample_size` | ENG-SAMPLESIZE | 0 | 0/5 |
| 5.1 | Hypothesis-testing steps; formulating H₀/H₁ | `hypothesis_test`, `concept_check` | ready / ENG-CONCEPT | 0 | 0/5 |
| 5.2 | Type I and Type II errors | `concept_check` | ENG-CONCEPT | 0 | 0/5 |
| 5.3 | Test for a mean or proportion with p-value | `hypothesis_test` | ready | 1 | 0/6 |
| 5.4 | One-sample vs. two-sample tests | `two_sample_test`, `concept_check` | ENG-TWOSAMPLE | 0 | 0/5 |
| 6.1 | Regression equation and prediction | `linear_regression` | ready | 1 | 0/5 |
| 6.2 | Model significance and fit (p-value, R², s_e) | `regression_output` | ENG-REGOUT | 0 | 0/5 |
| 6.3 | Slope confidence interval and predictor significance | `regression_output` | ENG-REGOUT | 0 | 0/5 |
| 6.4 | Multiple regression and predictor selection | `regression_output` | ENG-REGOUT | 0 | 0/6 |
| 6.5 | Common regression problems | `concept_check` | ENG-CONCEPT | 0 | 0/6 |

Total planned: **121 new questions**, plus the 16 existing ones.

## Batch order

Batches are structured to align directly with active coursework pace (Module 3 active as of October 11, 2026), progressing through subsequent modules in sync with class, leading up to the Module 6 quiz (Nov 1) and cumulative final exam (Nov 8). Rows stay whole within each batch.

| Batch | Questions | Rows | Engine status | Target / Alignment |
|---|---|---|---|---|
| 1 | 17 | 3.1 (5), 3.2 (6), 3.3 (6) | All ready (`concept_check`, `normal_dist`, `exponential_dist`, `uniform_dist`) | **Module 3 (Active now, Oct 11)**: Immediate practice on continuous distributions. |
| 2 | 22 | 4.1 (6), 4.2 (5), 4.3 (6), 4.4 (5) | Ready / ENG-PROP-SAMPLING, ENG-SAMPLESIZE | **Module 4**: Sampling distributions, CLT, confidence intervals, sample size. |
| 3 | 21 | 5.1 (5), 5.2 (5), 5.3 (6), 5.4 (5) | Ready / ENG-TWOSAMPLE | **Module 5**: Hypothesis testing formulation, Type I/II errors, p-values, 2-sample tests. |
| 4 | 28 | 6.1 (5), 6.2 (5), 6.3 (5), 6.4 (6), 6.5 (6) | Ready (`regression_output`, `linear_regression`, `concept_check`) | **Module 6 (Due Nov 1)**: Simple and multiple regression before the Module 6 quiz. |
| 5 | 33 | 1.1 (5), 1.2 (6), 1.3 (6), 2.1 (6), 2.2 (6), 2.3 (5) | ENG-DESC, ENG-PMF, ENG-AFFINE, ENG-PROB-EXT | **Modules 1–2 (Due Nov 8)**: Descriptive measures and discrete RV review before final. |

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

- [ ] `concept-descriptive-vs-inferential`: retailer summarizes last quarter's sales at all 120 stores vs. estimates the chain mean from 40 sampled stores; classify each.
- [ ] `concept-parameter-vs-statistic`: survey of 500 of 20,000 customers finds 62% satisfied; identify population, sample, parameter, statistic.
- [ ] `concept-sampling-bias-online-poll`: brand runs a voluntary poll on its loyalty app; which conclusion is unjustified and why (selection bias).
- [ ] `concept-data-types-business`: classify region, satisfaction rating (1–5), monthly revenue, and units returned as categorical/numerical and discrete/continuous.
- [ ] `concept-cross-sectional-vs-time-series`: dashboard of 2025 sales by store vs. monthly sales for one store; identify the data type and one valid question for each.

#### 1.2 Measures of center and variability (`descriptive_stats`)

- [ ] `desc-mean-median-outlier`: monthly sales of 7 stores, one flagship outlier; compute mean and median and decide which better describes a typical store.
- [ ] `desc-sample-sd-delivery-times`: 6 delivery times; sample SD (n−1) with a population-divisor distractor; `STDEV.S` vs. `STDEV.P`.
- [ ] `desc-cv-compare-stocks`: two stocks with different mean returns and SDs; compare relative risk with the coefficient of variation.
- [ ] `desc-zscore-empirical-rule`: salesperson's quarterly sales vs. team mean/SD; z-score, then the share within ±2 SD by the empirical rule.
- [ ] `desc-weighted-mean-price`: average selling price across a three-product mix weighted by units sold; unweighted-average distractor.
- [ ] `desc-quartiles-iqr-fences` **[confirm]**: call-handling times; Q1, Q3, IQR, 1.5·IQR fences, and flag an outlier.

#### 1.3 Probability rules (`set_probability` + ENG-PROB-EXT)

- [ ] `prob-multiplication-independent-suppliers`: two independent suppliers on time with probability 0.90 and 0.85; P(both on time).
- [ ] `prob-independence-check`: P(A) = 0.4, P(B) = 0.5, P(A∩B) = 0.20; decide independence, then contrast with P(A∩B) = 0.25.
- [ ] `prob-contingency-age-purchase`: 2×3 table of age group × purchased; joint, marginal and conditional probabilities.
- [ ] `prob-disjoint-vs-independent`: misconception item; show that disjoint events with positive probabilities cannot be independent.
- [ ] `prob-at-least-one-server-failure`: 3 independent servers each fail with probability 0.05; P(at least one fails) via the complement.
- [ ] `prob-two-plant-defect-tree` **[confirm]**: plants supply 60%/40% with defect rates 2%/5%; P(defective), then P(plant B | defective).

### Module 2 — Discrete random variables and distributions

#### 2.1 Discrete RV and its probability distribution (`discrete_rv`)

- [ ] `drv-valid-pmf-missing-value`: cars sold per day, 0–4 with one missing probability; solve for it and check validity.
- [ ] `drv-expected-cars-sold`: same dealership PMF; E[X] with `SUMPRODUCT`.
- [ ] `drv-variance-cars-sold`: Var(X) and SD; distractor uses Σx²p without subtracting μ².
- [ ] `drv-compare-projects-risk`: two investment payoff tables with equal expected profit but different SD; choose using risk.
- [ ] `drv-insurance-break-even`: policy payout distribution; expected payout and the break-even premium.
- [ ] `drv-cumulative-from-table`: P(X ≥ 2) and P(1 ≤ X ≤ 3) from a demand PMF.

#### 2.2 Expected value, variance, transformations (`linear_combination` + ENG-AFFINE)

- [ ] `affine-profit-from-units`: profit = 50X − 200 with E[X] = 30, SD 6; E and SD of profit (the −200 does not affect SD).
- [ ] `affine-currency-conversion`: euro-denominated costs converted at 1.10; mean and SD scale, variance by 1.21.
- [ ] `lincomb-total-two-product-lines`: independent daily revenues; mean and SD of the total (adding SDs is a distractor).
- [ ] `lincomb-difference-independent`: Var(X − Y) for independent demand at two stores; subtracting variances is a distractor.
- [ ] `lincomb-weekly-demand-sum-vs-scale`: 5 i.i.d. days vs. 5·X; variance 5σ² vs. 25σ².
- [ ] `lincomb-portfolio-correlated`: 60/40 portfolio with ρ = 0.3; portfolio variance with the covariance term.

#### 2.3 Choosing and using binomial vs. Poisson

- [ ] `choose-binomial-or-poisson`: (`concept_check`) four scenarios, e.g. 20 surveyed customers who renew vs. website errors per hour; choose a model and justify it.
- [ ] `binomial-at-least-three-defects`: n = 12, p = 0.15, P(X ≥ 3) via complement and `BINOM.DIST(2,12,0.15,TRUE)`.
- [ ] `binomial-mean-sd-shipments`: n = 200, p = 0.02; expected defects and SD.
- [ ] `poisson-rate-rescaling-quarter-hour`: 6 calls/hour; P(X ≤ 1) in 15 minutes (λ = 1.5); unscaled-λ distractor.
- [ ] `binomial-condition-violated-small-lot`: (`concept_check`) sampling 5 of 20 items without replacement; which binomial condition fails.

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

- [ ] `ci-mean-z-known-sigma`: σ = 12, n = 49, x̄ = 150, 95%; `CONFIDENCE.NORM`.
- [ ] `ci-mean-t-small-sample-90`: n = 10, x̄ = 42.3, s = 5.1, 90%; df = 9.
- [ ] `ci-proportion-recommend`: 132 of 400 customers recommend; 95% interval for p.
- [ ] `ci-interpretation-statement`: choose the correct interpretation of a 95% interval; "95% probability μ is in this interval" is a distractor.
- [ ] `ci-level-vs-width`: same data at 90%, 95%, 99%; order the widths and explain the trade-off.
- [ ] `ci-plausible-claim`: interval (146.6, 153.4); is a claimed μ = 155 plausible?

#### 4.4 Required sample size (`sample_size`)

- [ ] `n-mean-margin-five`: σ = 25, E = 5, 95% ⇒ n = ⌈96.04⌉ = 97.
- [ ] `n-proportion-pilot`: pilot p̂ = 0.20, E = 0.03, 95%.
- [ ] `n-proportion-conservative`: no prior estimate, E = 0.04, 90%; use p = 0.5.
- [ ] `n-halve-margin-quadruple`: halving E multiplies n by 4.
- [ ] `n-round-up-misconception`: 96.04 must round up to 97, not to 96.

### Module 5 — Hypothesis testing

#### 5.1 Steps and formulating H₀/H₁ (`hypothesis_test`, `concept_check`)

- [ ] `hyp-formulate-underfill`: label claims 16 oz; consumer group suspects underfill; H₀: μ = 16 (or ≥ 16), H₁: μ < 16.
- [ ] `hyp-formulate-right-tailed-sales`: new display claimed to raise mean daily sales above 50.
- [ ] `hyp-formulate-two-tailed-calibration`: machine calibrated to 2.00 cm; detect drift in either direction.
- [ ] `hyp-formulate-proportion`: marketing claims more than 40% of customers prefer the new package.
- [ ] `hyp-steps-and-conclusion`: order the testing steps; decision rule p ≤ α; "fail to reject" vs. "accept H₀".

#### 5.2 Type I and Type II errors (`concept_check`)

- [ ] `errors-define-in-context`: quality control rejects a good lot (Type I) vs. accepts a bad lot (Type II).
- [ ] `errors-which-is-worse`: food-safety recall vs. ad-campaign test; which error is costlier and how that drives α.
- [ ] `errors-alpha-beta-tradeoff`: lowering α from 0.05 to 0.01 with n fixed increases β.
- [ ] `errors-larger-n-reduces-beta`: increasing n reduces β at fixed α.
- [ ] `errors-identify-after-the-fact`: H₀ rejected but later shown true; name the error.

#### 5.3 Test for a mean or proportion with p-value (`hypothesis_test`)

- [ ] `hyp-z-mean-two-tailed`: σ known = 0.05, n = 36, x̄ = 2.02, H₀: μ = 2.00, α = 0.05.
- [ ] `hyp-t-mean-right-tailed`: n = 20, x̄ = 53.1, s = 6.2, H₀: μ = 50; `T.DIST.RT`.
- [ ] `hyp-proportion-left-tailed`: claim 20% defective; 28 of 200 sampled; test p < 0.20.
- [ ] `hyp-pvalue-interpretation`: choose the correct meaning of p = 0.03; "probability H₀ is true" is a distractor.
- [ ] `hyp-alpha-changes-decision`: p = 0.032; reject at α = 0.05, not at α = 0.01.
- [ ] `hyp-excel-tail-function`: choose `T.DIST.RT`, `T.DIST`, or `T.DIST.2T` for the tail stated in H₁.

#### 5.4 One-sample vs. two-sample tests (`two_sample_test`, `concept_check`)

- [ ] `twosample-choose-design`: (`concept_check`) classify scenarios as one-sample, independent two-sample, or paired.
- [ ] `twosample-paired-training`: before/after weekly sales for 8 reps; paired t on the differences.
- [ ] `twosample-independent-means-layouts` **[confirm]**: average basket under two store layouts; independent two-sample t.
- [ ] `twosample-two-proportion-ads`: conversion 48/600 vs. 72/600 for two ad versions; pooled two-proportion z.
- [ ] `twosample-read-excel-output`: interpret an Excel t-Test two-sample output block (t Stat, P(T<=t) one-tail/two-tail, t Critical).

### Module 6 — Regression analysis

#### 6.1 Regression equation and prediction (`linear_regression`)

- [ ] `reg-fit-raw-points-price-units`: 6 (price, units) pairs; engine fits b₀, b₁, then predict.
- [ ] `reg-interpret-slope-intercept`: interpret b₁ in units and judge whether b₀ is meaningful.
- [ ] `reg-residual-observed-point`: compute ŷ and residual e = y − ŷ for one observed store.
- [ ] `reg-extrapolation-warning`: x range 2–10; prediction at x = 25 is extrapolation.
- [ ] `reg-predict-from-excel-coefficients`: read Intercept and slope from an output table and predict.

#### 6.2 Model significance and fit (`regression_output`)

- [ ] `regout-r2-from-anova`: SSR and SST from the ANOVA table; R² and its interpretation (no causation).
- [ ] `regout-se-from-sse`: SSE, n = 20; s_e = √(SSE/(n−2)) and its meaning in y units.
- [ ] `regout-significance-f`: interpret Significance F = 0.0004 at α = 0.05.
- [ ] `regout-r-from-r2-sign`: R² = 0.64 with negative slope ⇒ r = −0.8.
- [ ] `regout-compare-models-fit`: two candidate models; compare with s_e and adjusted R².

#### 6.3 Slope confidence interval and predictor significance (`regression_output`)

- [ ] `regout-slope-t-pvalue`: b₁ = 2.4, SE = 0.8, n = 25; t = 3.0, df = 23, p-value.
- [ ] `regout-slope-ci-includes-zero`: Lower 95% and Upper 95% columns; decide significance from whether 0 is inside.
- [ ] `regout-slope-ci-interpretation`: interpret the slope CI in business units.
- [ ] `regout-t-squared-equals-f`: simple regression output; verify t² = F.
- [ ] `regout-compute-slope-ci`: b₁ ± t*·SE with `T.INV.2T(0.05, df)`.

#### 6.4 Multiple regression and predictor selection (`regression_output`)

- [ ] `mreg-write-equation-predict`: output with 3 predictors (ad spend, price, store size); write ŷ and predict.
- [ ] `mreg-interpret-coefficient-ceteris-paribus`: interpret the price coefficient holding the others constant.
- [ ] `mreg-which-predictors-significant`: p-values 0.001, 0.04, 0.62; which to keep and which to drop at α = 0.05.
- [ ] `mreg-adjusted-r2-weak-predictor`: adding a weak predictor raises R² but lowers adjusted R².
- [ ] `mreg-overall-f-vs-individual-t`: significant F with one non-significant t; interpret both.
- [ ] `mreg-dummy-variable` **[confirm]**: a region dummy (1 = urban); interpret its coefficient as a shift in the intercept.

#### 6.5 Common regression problems (`concept_check`)

- [ ] `regprob-multicollinearity`: high R² and significant F but no significant t; correlated predictors; diagnose.
- [ ] `regprob-heteroscedasticity`: fan-shaped residual plot; which assumption is violated.
- [ ] `regprob-nonlinearity`: curved residual pattern; the model form is wrong.
- [ ] `regprob-autocorrelation`: monthly time-series residuals alternate in runs; independence violated.
- [ ] `regprob-outlier-influence`: one store far from the others changes the slope sign; what to do.
- [ ] `regprob-omitted-variable-causation`: ice-cream sales vs. sunscreen; omitted variable and association vs. causation.

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

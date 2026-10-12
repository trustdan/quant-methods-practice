# Final Wave 4 and Wave 5 content review

These 20 fixed-reference questions were authored and semantically reviewed by the assistant. Numerical choices are checked against the Go engine at bank admission. Course-specific quartile, pooled-test and dummy-variable conventions remain explicitly unconfirmed. Semantic review does not itself authorize activation. Approval status is recorded in each JSON and review file.

## Quartiles and outlier fences using Excel's inclusive convention

ID: `desc_quartiles_iqr_fences`; family `descriptive_stats`, rule 2; status **active**.

Nine observed call-handling times are 2, 3, 4, 5, 6, 7, 8, 9, and 30 minutes. Use Excel QUARTILE.INC, defined by zero-based position (n-1)p with linear interpolation, and the 1.5 IQR outlier rule. The course's required quartile convention is still unconfirmed.

Assumptions:

- Use QUARTILE.INC rather than QUARTILE.EXC or a median-of-halves method.
- An outlier flag is a diagnostic, not permission to remove an observation.

### 1. inclusive_first_quartile

What is Q1 in minutes under the stated inclusive convention? Select the value rounded to four decimal places.

1. $4.0000$ **[correct]**
2. $3.5000$
   Hint: A median-of-halves convention can produce a different value; use (n-1)*0.25.
3. $6.0000$
   Hint: The median is Q2, not Q1.

Explanation: Position $(9-1)(0.25)=2$ is the third ordered observation, 4 minutes. Excel: `=QUARTILE.INC(A2:A10,1)`.

Engine quantity `q1`: 4. Four-decimal choices; exact Go value retained for audit.

### 2. inclusive_third_quartile

What is Q3 in minutes? Select the value rounded to four decimal places.

1. $8.0000$ **[correct]**
2. $8.5000$
   Hint: Use the stated inclusive position, not an exclusive convention.
3. $30.0000$
   Hint: Q3 is the 75th percentile, not the maximum.

Explanation: Position $(9-1)(0.75)=6$ is the seventh ordered observation, 8 minutes. Excel: `=QUARTILE.INC(A2:A10,3)`.

Engine quantity `q3`: 8. Four-decimal choices; exact Go value retained for audit.

### 3. iqr_upper_fence

What is the upper 1.5 IQR fence in minutes? Select the value rounded to four decimal places.

1. $14.0000$ **[correct]**
2. $12.0000$
   Hint: Add 1.5 times IQR to Q3, not just IQR.
3. $10.0000$
   Hint: Use Q3 as the upper-fence anchor rather than Q1.

Explanation: $IQR=8-4=4$; upper fence $=8+1.5(4)=14$. The lower fence is $4-6=-2$. Excel: `=Q3+1.5*(Q3-Q1)`.

Engine quantity `upper_fence`: 14. Four-decimal choices; exact Go value retained for audit.

### 4. iqr_outlier_followup

Which observation is flagged, and what should happen next?

1. 30 minutes is above 14; investigate its context and accuracy before deciding how to handle it. **[correct]**
2. Every observation above Q3 must be deleted.
   Hint: The upper quartile is not the upper outlier fence.
3. 30 minutes must be a recording error because it is flagged.
   Hint: Unusual observations can be valid; the fence does not establish their cause.

Explanation: Only 30 lies outside [-2,14]. Investigate data quality and operational circumstances, and document any handling decision. Quartile conventions can affect fences, so the chosen method must be stated.
## Computing a unit-weighted selling price

ID: `desc_weighted_mean_price`; family `descriptive_stats`, rule 2; status **active**.

A subscription retailer sells 100 basic packages at $10, 50 standard packages at $20, and 50 premium packages at $40. Find total revenue per package across all 200 packages, treating units sold as the weights.

Assumptions:

- Prices and quantities refer to the same period and exclude tax and returns.
- The target is revenue per package, so weights are unit counts rather than revenue shares.

### 1. unit_weighted_mean_price

What is the average selling price in dollars per package? Select the value rounded to four decimal places.

1. $20.0000$ **[correct]**
2. $23.3333$
   Hint: The three products have unequal unit counts; do not weight them equally.
3. $4000.0000$
   Hint: Revenue must be divided by the total number of packages.

Explanation: $\bar p_w=(10(100)+20(50)+40(50))/(100+50+50)=4000/200=20$. Excel: `=SUMPRODUCT(A2:A4,B2:B4)/SUM(B2:B4)` with prices in A and quantities in B.

Engine quantity `mean`: 20. Four-decimal choices; exact Go value retained for audit.

### 2. unit_weighted_mean_interpretation

Why does the unweighted average of the three listed prices answer a different question?

1. It gives each product one vote instead of giving each sold package one vote. **[correct]**
2. An unweighted mean is always mathematically undefined when values differ.
   Hint: The arithmetic mean exists; its weights describe the wrong target here.
3. The highest price must receive the largest weight.
   Hint: Unit volume, not price, determines the weights for revenue per package.

Explanation: The unweighted price mean is $70/3=23.3333$. It describes an equally weighted product list. The transaction mix has twice as many basic packages as either other package, yielding $20 per package.
## Standardizing sales and applying the empirical rule

ID: `desc_zscore_empirical_rule`; family `descriptive_stats`, rule 2; status **active**.

A large sales team's quarterly revenue distribution is approximately bell-shaped, with mean 100 thousand dollars and standard deviation 20 thousand dollars. A representative sells 140 thousand dollars. Use the empirical rule as an approximation, rather than assuming exact normal probabilities.

Assumptions:

- All revenue values use the same units: thousands of dollars.
- The empirical rule is appropriate as an approximation for the stated bell-shaped distribution.

### 1. sales_z_score

What is this representative's z-score? Select the value rounded to four decimal places.

1. $2.0000$ **[correct]**
2. $7.0000$
   Hint: Subtract the mean before dividing by SD.
3. $-2.0000$
   Hint: The observation is above the mean, so its z-score is positive.

Explanation: $z=(140-100)/20=2$. Excel: `=STANDARDIZE(140,100,20)`. This sales result is two SD above the team mean.

Engine quantity `z`: 2. Four-decimal choices; exact Go value retained for audit.

### 2. empirical_rule_two_sd

Approximately what share of sales results falls from 60 to 140 thousand dollars?

1. About 95%, since these endpoints are two SD below and above the mean. **[correct]**
2. About 68%, because every symmetric interval contains one SD.
   Hint: Count how many SD each endpoint is from the mean.
3. Exactly 100%, because no result can be farther than two SD from its mean.
   Hint: The empirical rule allows observations outside two SD.

Explanation: The interval is $100\pm2(20)$. The empirical rule assigns about 95% of a bell-shaped distribution within two SD. Under an exact normal model the value is about 95.45%, which is a different calculation.

### 3. empirical_rule_condition

Could the same 95% approximation be assumed for a strongly skewed sales distribution?

1. No; mean and SD alone do not justify the bell-shaped empirical-rule percentages. **[correct]**
2. Yes; every distribution has exactly 95% within two SD.
   Hint: The empirical rule depends on distribution shape.
3. Yes; positive sales values guarantee symmetry.
   Hint: Nonnegative support does not imply a bell-shaped distribution.

Explanation: A z-score still standardizes using mean and SD, but empirical-rule percentages require a suitable shape. Check the distribution before using them.
## Comparing payoff variability at equal expected profit

ID: `drv_compare_projects_risk`; family `discrete_rv`, rule 2; status **active**.

Two fictional projects have these net payoffs in thousands of dollars: A pays 0 or 200, each with probability 0.5; B pays 80 or 120, each with probability 0.5. A manager's stated criterion is to minimize payoff SD among alternatives with the same expected profit. This is a classroom risk comparison, with no claim that SD captures every investment concern.

Assumptions:

- The listed distributions exhaust each project's outcomes.
- The decision criterion is equal mean followed by lower payoff standard deviation.

### 1. project_a_expected_payoff

What is Project A's expected payoff in thousands of dollars? Select the value rounded to four decimal places.

1. $100.0000$ **[correct]**
2. $200.0000$
   Hint: Weight both possible payoffs by their probabilities.
3. $50.0000$
   Hint: Use the full probability-weighted sum, not an extra averaging step.

Explanation: $E[A]=0(0.5)+200(0.5)=100$. Also $E[B]=80(0.5)+120(0.5)=100$. Excel: `=SUMPRODUCT(payoffs,probabilities)`.

Engine quantity `mean`: 100. Four-decimal choices; exact Go value retained for audit.

### 2. project_b_payoff_sd

What is Project B's payoff SD in thousands of dollars? Select the value rounded to four decimal places.

1. $20.0000$ **[correct]**
2. $400.0000$
   Hint: 400 is its variance, in squared units; take the square root.
3. $40.0000$
   Hint: The full range is not the standard deviation.

Explanation: $\operatorname{Var}(B)=0.5(80-100)^2+0.5(120-100)^2=400$, so SD is 20. A has variance 10000 and SD 100.

Engine quantity `sd_b`: 20. Four-decimal choices; exact Go value retained for audit.

### 3. equal_mean_risk_comparison

Which project meets this manager's stated criterion?

1. B: both means are 100, while B's SD is 20 versus A's 100. **[correct]**
2. A: its largest possible payoff is higher, so it must have lower risk.
   Hint: Maximum payoff does not measure dispersion around the mean.
3. The projects are identical because their means match.
   Hint: Equal means do not imply equal variances or distributions.

Explanation: The criterion selects B's lower variability at equal expectation. This conclusion is conditional on the stated objective; other preferences can value upside or downside differently.
## Tail and inclusive interval probabilities from a demand table

ID: `drv_cumulative_from_table`; family `discrete_rv`, rule 2; status **active**.

Daily emergency-parts demand X takes values 0, 1, 2, 3, 4 with probabilities 0.10, 0.20, 0.30, 0.25, 0.15 respectively. A depot stocks one unit, and also wants to know how often demand lies from one through three inclusive.

Assumptions:

- The PMF sums to one and exhausts all possible daily demands.
- The interval includes both endpoints; at least two means X>=2.

### 1. discrete_demand_tail

What is P(X>=2), the probability one stocked unit is insufficient? Select the value rounded to four decimal places.

1. $0.7000$ **[correct]**
2. $0.4000$
   Hint: At least two includes demand exactly two.
3. $0.3000$
   Hint: P(X=2) is only one outcome in the tail.

Explanation: $P(X\ge2)=0.30+0.25+0.15=0.70=1-P(X\le1)$. Excel: `=SUMIF(A2:A6,">=2",B2:B6)`.

Engine quantity `geq`: 0.7. Four-decimal choices; exact Go value retained for audit.

### 2. discrete_demand_interval

What is P(1<=X<=3)? Select the value rounded to four decimal places.

1. $0.7500$ **[correct]**
2. $0.2500$
   Hint: Include the endpoints 1 and 3, not just the interior.
3. $0.9000$
   Hint: Exclude demand 4 as well as demand 0.

Explanation: $P(1\le X\le3)=0.20+0.30+0.25=0.75$. Excel: `=SUMIFS(B2:B6,A2:A6,">=1",A2:A6,"<=3")`.

Engine quantity `between`: 0.75. Four-decimal choices; exact Go value retained for audit.

### 3. discrete_tail_complement_boundary

Which complement correctly computes P(X>=2)?

1. 1 - P(X<=1). **[correct]**
2. 1 - P(X<=2).
   Hint: The complement must exclude X=2 from the subtracted event.
3. 1 - P(X<1).
   Hint: The complement of X>=2 includes both 0 and 1.

Explanation: For integer counts, 'at least 2' includes 2. Its complement is 'at most 1'. Subtracting P(X<=2) instead computes P(X>=3), losing the mass at 2.
## Expected insurance payout and an actuarial break-even premium

ID: `drv_insurance_break_even`; family `discrete_rv`, rule 2; status **active**.

For a one-year policy, insurer payout L is $0 with probability 0.970, $1000 with probability 0.025, or $10000 with probability 0.005. A fixed annual premium is collected once. Ignore administration, commissions, investment income, capital costs and profit loading. Find the premium giving zero expected underwriting profit under these simplified assumptions.

Assumptions:

- The payouts are mutually exclusive and exhaustive.
- No costs or revenue other than the stated payout and fixed premium are modeled.
- Break-even refers to expected profit, not a guarantee on an individual policy.

### 1. insurance_expected_payout

What is expected payout in dollars per policy? Select the value rounded to four decimal places.

1. $75.0000$ **[correct]**
2. $5500.0000$
   Hint: Do not average only the two positive payouts; include their probabilities.
3. $25.0000$
   Hint: Include the rare $10000 payout as well as the $1000 payout.

Explanation: $E[L]=0(0.970)+1000(0.025)+10000(0.005)=25+50=75$. Excel: `=SUMPRODUCT(A2:A4,B2:B4)`.

Engine quantity `mean`: 75. Four-decimal choices; exact Go value retained for audit.

### 2. insurance_expected_break_even

Which premium gives zero expected underwriting profit in this model?

1. $75 per policy, because expected profit is premium minus expected payout. **[correct]**
2. $0, because the most likely payout is zero.
   Hint: Break-even depends on expected payout across all outcomes, not the modal payout.
3. $10000, because the premium must equal the largest possible claim to break even in expectation.
   Hint: Expected break-even is different from covering the worst outcome on each policy.

Explanation: For fixed premium c, $E[c-L]=c-E[L]$. Set c=75. Individual policies can make or lose money; a real insurer must also account for costs, capital, risk and loading outside this simplified model.
## Portfolio variance with a covariance term

ID: `lincomb_portfolio_correlated`; family `linear_combination`, rule 2; status **active**.

A fictional portfolio assigns weights 0.60 and 0.40 to returns X and Y. Mean returns are 8 and 4 percentage points; SDs are 10 and 5 percentage points; correlation is 0.30. Compute moments of R=0.60X+0.40Y. Use percentage-point units consistently, not decimal returns.

Assumptions:

- Weights are fixed, sum to one, and returns have finite second moments.
- Correlation 0.30 is given; independence must not be assumed.
- Variance and covariance use squared percentage-point units.

### 1. portfolio_covariance_from_correlation

What is Cov(X,Y) in squared percentage-point units? Select the value rounded to four decimal places.

1. $15.0000$ **[correct]**
2. $0.3000$
   Hint: Correlation is dimensionless; multiply it by both SDs.
3. $750.0000$
   Hint: Use SDs 10 and 5, not variances 100 and 25.

Explanation: $\operatorname{Cov}(X,Y)=\rho\sigma_X\sigma_Y=0.30(10)(5)=15$. Excel: `=0.3*10*5`.

Engine quantity `covariance`: 15. Four-decimal choices; exact Go value retained for audit.

### 2. portfolio_correlated_variance

What is Var(R) in squared percentage-point units? Select the value rounded to four decimal places.

1. $47.2000$ **[correct]**
2. $40.0000$
   Hint: Include the positive covariance contribution.
3. $70.0000$
   Hint: Square each portfolio weight in its own variance term.

Explanation: $\operatorname{Var}(R)=0.6^2(100)+0.4^2(25)+2(0.6)(0.4)(15)=36+4+7.2=47.2$. SD is $\sqrt{47.2}=6.8702$ percentage points and mean is 6.4 percentage points. Excel: `=0.6^2*100+0.4^2*25+2*0.6*0.4*0.3*10*5`.

Engine quantity `variance`: 47.2. Four-decimal choices; exact Go value retained for audit.

### 3. portfolio_positive_correlation_effect

Compared with the same inputs but zero correlation, what changes?

1. Variance rises from 40 to 47.2; the mean stays 6.4 percentage points. **[correct]**
2. The mean rises because positive correlation increases expected returns.
   Hint: The expectation formula has no covariance term.
3. Variance falls because positive correlation always improves diversification.
   Hint: The cross term is positive for positive weights and positive covariance.

Explanation: Zero correlation removes the cross term but does not alter the weighted mean. Correlation and independence are different concepts; zero correlation is sufficient for this variance calculation without proving independence.
## Five independent days versus five copies of one random day

ID: `lincomb_weekly_demand_sum_vs_scale`; family `linear_combination`, rule 2; status **active**.

Daily demand has mean 20 units and variance 9 units squared. Define S=X1+...+X5 for five independent, identically distributed days, and T=5X for five times one random day's demand. These are different random quantities despite having the same expected value.

Assumptions:

- The five Xi in S are independent and have identical finite moments.
- T repeats the same random X five times, not independent draws.

### 1. iid_weekly_sum_variance

What is Var(S) in squared demand units? Select the value rounded to four decimal places.

1. $45.0000$ **[correct]**
2. $225.0000$
   Hint: Independent days have zero covariance; add five daily variances.
3. $15.0000$
   Hint: Variance is given as 9; do not replace it with SD=3 before summing variances.

Explanation: $\operatorname{Var}(S)=5(9)=45$. SD is $\sqrt{45}$, whereas the expected sum is 5(20)=100. Excel variance arithmetic: `=5*9`.

Engine quantity `var_sum`: 45. Four-decimal choices; exact Go value retained for audit.

### 2. scaled_day_variance

What is Var(T) in squared demand units? Select the value rounded to four decimal places.

1. $225.0000$ **[correct]**
2. $45.0000$
   Hint: Scaling one random variable squares the multiplier.
3. $15.0000$
   Hint: The variance scaling rule uses 5 squared, not 5 times the SD.

Explanation: $\operatorname{Var}(5X)=5^2\operatorname{Var}(X)=25(9)=225$. Excel: `=5^2*9`. Its SD is 15.

Engine quantity `var_scaled`: 225. Four-decimal choices; exact Go value retained for audit.

### 3. sum_vs_scale_dependence

Why do these variances differ even though E[S]=E[T]=100?

1. S combines independent draws; T repeats the identical draw, so all five components move together. **[correct]**
2. Linearity of expectation fails for T.
   Hint: Expectation is linear with or without independence.
3. Equal means mathematically require equal variances.
   Hint: Location and dispersion are different distribution features.

Explanation: Independence removes cross-covariances in S. In T, each copy is perfectly dependent on the others. The distinction is probabilistic structure, not merely algebraic appearance.
## A weak predictor can reduce adjusted R-squared

ID: `mreg_adjusted_r2_weak_predictor`; family `regression_output`, rule 2; status **active**.

Using the same 30 cafes and response (SST=1000), a model with advertising and price has k=2, SSR=700, SSE=300. Adding window count gives k=3, SSR=702, SSE=298. Both models have intercepts. Decide whether the small fit improvement offsets the extra predictor under adjusted R-squared.

Assumptions:

- These are nested least-squares models on the identical response rows.
- Adjusted R-squared penalizes added predictors through residual degrees of freedom.
- Predictor retention may also depend on a pre-specified purpose and subject knowledge.

### 1. weak_predictor_raw_r2

What is ordinary R-squared after adding window count? Select the value rounded to four decimal places.

1. $0.7020$ **[correct]**
2. $0.7000$
   Hint: Use the new SSR and SSE.
3. $0.2980$
   Hint: SSE/SST is the unexplained fraction.

Explanation: $R_B^2=702/1000=0.7020$. An added predictor in nested OLS cannot lower ordinary R-squared. Excel: `=702/1000`.

Engine quantity `r2_b`: 0.702. Four-decimal choices; exact Go value retained for audit.

### 2. weak_predictor_adjusted_r2

What is adjusted R-squared after adding window count? Select the value rounded to four decimal places.

1. $0.6676$ **[correct]**
2. $0.7020$
   Hint: Adjusted R-squared includes a predictor penalty.
3. $0.6778$
   Hint: That is the adjusted value before adding the predictor.

Explanation: $\bar R_B^2=1-(298/26)/(1000/29)=0.6676$. Before addition it was $1-(300/27)/(1000/29)=0.6778$.

Engine quantity `adjusted_r2_b`: 0.667615384615. Four-decimal choices; exact Go value retained for audit.

### 3. weak_predictor_retention_decision

If adjusted fit is the stated criterion, what does the comparison support?

1. Prefer the simpler model: the small SSE decrease does not offset the loss of one residual degree of freedom. **[correct]**
2. Keep every new predictor whenever ordinary R-squared increases.
   Hint: Ordinary R-squared rewards added flexibility even when improvement is negligible.
3. Window count has been proved to have exactly zero population effect.
   Hint: A weak sample contribution is not proof of a zero population coefficient.

Explanation: Adjusted R-squared falls from 0.6778 to 0.6676. This criterion favors the simpler model, while theoretical controls or prediction goals could justify a different choice. It is not an automatic deletion rule for every setting.
## Interpreting an urban indicator coefficient

ID: `mreg_dummy_variable`; family `regression_output`, rule 2; status **active**.

A generic retail example uses $\hat y=20+3x+8D$, where y is weekly revenue in thousands of dollars, x is advertising spending in thousands of dollars, and D=1 for urban stores and 0 for rural stores. The model has no interaction term. Predict an urban store's revenue when x=10 and interpret the indicator. Dummy-variable assessment coverage is not yet confirmed for this course.

Assumptions:

- Rural stores (D=0) form the reference category; urban stores have D=1.
- The model is additive with a common advertising slope and no interaction.
- Indicator-variable teaching is generic pending course confirmation; association alone is not causal.

### 1. dummy_variable_prediction

What is predicted weekly revenue for x=10 and D=1, in thousands of dollars? Select the value rounded to four decimal places.

1. $58.0000$ **[correct]**
2. $50.0000$
   Hint: An urban store has D=1, so include the indicator term.
3. $88.0000$
   Hint: Multiply advertising by 3 and the indicator by 8; do not swap them.

Explanation: $\hat y=20+3(10)+8(1)=58$. Excel: `=20+3*10+8*1`.

Engine quantity `prediction`: 58. Four-decimal choices; exact Go value retained for audit.

### 2. dummy_variable_shift

At the same advertising spending, what is the predicted urban-minus-rural difference, in thousands of dollars? Select the value rounded to four decimal places.

1. $8.0000$ **[correct]**
2. $3.0000$
   Hint: The advertising slope describes changes in x, not the category difference.
3. $28.0000$
   Hint: The urban intercept includes the baseline; the difference between intercepts is the dummy coefficient.

Explanation: The rural intercept is 20 and urban intercept is 28. Their difference is 8, independent of x in this additive model.

Engine quantity `b2`: 8. Four-decimal choices; exact Go value retained for audit.

### 3. dummy_variable_conditional_interpretation

Which interpretation is supported by this additive model?

1. Holding advertising fixed, urban stores have predicted weekly revenue 8 thousand dollars higher, with the same advertising slope. **[correct]**
2. Urban advertising has slope 11 instead of 3.
   Hint: Different slopes require an interaction term such as x times D.
3. Moving any rural store to an urban location is guaranteed to add 8 thousand dollars of revenue.
   Hint: A regression association does not establish an intervention effect.

Explanation: The indicator shifts the intercept, not the advertising slope. Both category lines have slope 3. The fitted contrast controls for included advertising, but unmeasured differences may remain.
## A significant model need not have every significant predictor

ID: `mreg_overall_f_vs_individual_t`; family `regression_output`, rule 2; status **active**.

A chain models monthly sales with three predictors across 40 independent stores. ANOVA gives SSR=900, SSE=300, regression df=3 and residual df=36. Coefficient / standard-error pairs are advertising 4/1, store size 2/0.5, and sign age 0.2/1. Use alpha=0.05 for the overall F test and two-sided individual t tests.

Assumptions:

- The model includes an intercept and three slopes.
- Errors are independent, normal and have constant variance for the stated inference.
- Each coefficient test holds the other included predictors constant.

### 1. overall_f_computation

What is the overall F statistic? Select the value rounded to four decimal places.

1. $36.0000$ **[correct]**
2. $3.0000$
   Hint: SSR/SSE is not F; first divide by their degrees of freedom.
3. $108.0000$
   Hint: Divide SSR by regression df as well as SSE by residual df.

Explanation: $F=(900/3)/(300/36)=36$. Excel: `=F.DIST.RT(36,3,36)` gives Significance F, below 0.05.

Engine quantity `f`: 36. Four-decimal choices; exact Go value retained for audit.

### 2. individual_sign_age_pvalue

What is the two-sided p-value for sign age? Select the value rounded to four decimal places.

1. $0.8426$ **[correct]**
2. $0.2000$
   Hint: The t statistic is not its p-value.
3. $0.0500$
   Hint: Alpha is chosen in advance; calculate the observed p-value.

Explanation: $t=0.2/1=0.2$ with df=36. Excel: `=T.DIST.2T(ABS(0.2),36)`. This p-value exceeds 0.05.

Engine quantity `p3`: 0.842606092305. Four-decimal choices; exact Go value retained for audit.

### 3. overall_vs_individual_hypotheses

How should the overall and sign-age tests be reported?

1. Reject the joint null that all three slopes are zero; do not reject the individual null for sign age conditional on the other predictors. **[correct]**
2. A significant F test proves all three slopes are nonzero.
   Hint: The overall alternative says at least one slope is nonzero.
3. The nonsignificant sign-age test invalidates the significant overall F test.
   Hint: The tests concern different null hypotheses.

Explanation: Overall F tests $H_0:\beta_1=\beta_2=\beta_3=0$ against at least one nonzero slope. The sign-age t test asks only about its partial coefficient. Sign age can be nonsignificant while advertising and store size make the model significant.
## Comparing residual error and adjusted R-squared

ID: `regout_compare_models_fit`; family `regression_output`, rule 2; status **active**.

A retailer fits two models to the same 30 stores, with weekly revenue measured in thousands of dollars. Both use an intercept and the same response values (SST=1400). Model A uses floor area (k=1, SSR=840, SSE=560). Model B adds staffing (k=2, SSR=980, SSE=420). Compare their in-sample fit after accounting for predictor count.

Assumptions:

- Both models use the same stores, response units and intercept convention.
- Residual standard error uses n-k-1 degrees of freedom.
- Fit summaries alone do not establish causation or future prediction quality.

### 1. compare_model_a_residual_error

What is Model A's residual standard error, in thousands of dollars? Select the value rounded to four decimal places.

1. $4.4721$ **[correct]**
2. $4.3205$
   Hint: Use residual df n-k-1, not n.
3. $20.0000$
   Hint: MSE is in squared response units; take its square root.

Explanation: $s_e=\sqrt{560/(30-1-1)}=\sqrt{20}$. Excel: `=SQRT(560/28)`. It measures residual dispersion in the response units.

Engine quantity `se`: 4.472135955. Four-decimal choices; exact Go value retained for audit.

### 2. compare_model_b_adjusted_fit

What is Model B's adjusted R-squared (a proportion)? Select the value rounded to four decimal places.

1. $0.6778$ **[correct]**
2. $0.7000$
   Hint: Ordinary R-squared does not penalize predictor count.
3. $0.3000$
   Hint: SSE/SST is unexplained variation before adjustment.

Explanation: $\bar R_B^2=1-(420/27)/(1400/29)$. Excel: `=1-(420/27)/(1400/29)`.

Engine quantity `adjusted_r2_b`: 0.677777777778. Four-decimal choices; exact Go value retained for audit.

### 3. compare_fit_with_prediction_limit

Which fit comparison is supported?

1. B has lower residual standard error and higher adjusted R-squared; validate future prediction separately. **[correct]**
2. A must fit better because it has fewer predictors.
   Hint: Simplicity matters through the penalty, but compare the actual adjusted metrics.
3. B is guaranteed to predict every future store more accurately.
   Hint: In-sample fit does not guarantee generalization.

Explanation: A has adjusted R-squared 0.5857 and residual SE 4.4721; B has adjusted R-squared 0.6778 and residual SE 3.9441. These favor B for these fit criteria. Held-out performance and subject knowledge still matter.
## Recognizing time dependence in regression residuals

ID: `regprob_autocorrelation`; family `concept_check`, rule 1; status **active**.

A utility models monthly demand using temperature. In chronological order, its residuals form long positive runs followed by long negative runs. Adjacent months tend to have similar residual signs, while residual spread is roughly stable. The data are a time series, not independent stores.

Assumptions:

- Residuals are plotted in chronological order.
- Runs suggest positive serial dependence, which should be checked rather than treated as proved.
- Standard OLS inferential formulas here assume independent errors.

### 1. serial_dependence_diagnosis

Which regression assumption is most directly in question?

1. Error independence: adjacent residuals may be positively autocorrelated. **[correct]**
2. Errors must always alternate signs every month.
   Hint: Independence does not require any deterministic sign pattern.
3. The coefficient of determination must exceed 1.
   Hint: Temporal dependence does not imply an impossible R-squared.

Explanation: Long runs in time order suggest positive serial correlation: an unexplained positive deviation tends to be followed by another. This is distinct from curvature against a predictor or a fan-shaped spread.

### 2. serial_dependence_inference

What is the key concern for conventional inference if errors are autocorrelated?

1. Standard OLS standard errors and tests can be unreliable when they ignore dependence. **[correct]**
2. OLS coefficients necessarily become exactly zero.
   Hint: Dependence affects uncertainty calculations without forcing coefficient values.
3. Residual standard error alone automatically corrects all time dependence.
   Hint: The usual scalar residual variance does not model correlations across months.

Explanation: With serially dependent errors, conventional independent-error standard errors, confidence intervals and p-values can misstate uncertainty. Coefficient bias depends on additional assumptions; autocorrelation alone does not imply it.

### 3. serial_dependence_followup

Which follow-up is appropriate?

1. Examine residuals over time and their autocorrelation; consider seasonal or time-series structure and suitable inference. **[correct]**
2. Shuffle the months until the runs disappear, then use the original p-values.
   Hint: Reordering a plot does not remove dependence in the observations.
3. Add unrelated predictors until ordinary R-squared reaches 1.
   Hint: Extra variables without a justified temporal mechanism do not resolve dependence.

Explanation: Check whether omitted seasonality, trends, or a time-series error process explain the runs. A plot is a diagnostic prompt, not a complete proof or an automatic choice of correction.
## Diagnosing a curved residual pattern

ID: `regprob_nonlinearity`; family `concept_check`, rule 1; status **active**.

An analyst fits a straight line to shipping cost versus shipment weight. Ordered by weight, residuals are mostly positive for light parcels, negative for medium parcels, and positive again for heavy parcels. Their vertical spread is roughly constant. This U-shaped pattern persists across many observations.

Assumptions:

- Residuals are observed cost minus fitted cost.
- The pattern is against the predictor, not chronological order.
- Diagnostics suggest model issues; they do not prove a unique replacement model.

### 1. curved_residual_model_form

Which model issue is most directly suggested?

1. The straight-line conditional mean is inadequate; the residual pattern suggests curvature. **[correct]**
2. The only issue is heteroscedasticity.
   Hint: Heteroscedasticity concerns changing spread; this pattern concerns systematic signs.
3. The residuals prove weight and cost are independent.
   Hint: A systematic residual pattern indicates structure left unexplained.

Explanation: A correctly specified conditional mean should leave residuals centered near zero across the predictor. The U shape indicates a missing nonlinear component.

### 2. curved_residual_direction

What does a positive residual for a heavy parcel mean?

1. Its observed shipping cost exceeds the line's fitted cost. **[correct]**
2. Its fitted cost exceeds observed cost.
   Hint: Use residual = observed minus fitted.
3. The parcel necessarily has high leverage.
   Hint: Residual magnitude and leverage describe different properties.

Explanation: $e=y-\hat y$. Positive e means underprediction of the observed cost. A residual sign alone does not establish leverage or influence.

### 3. curved_residual_followup

What is a reasonable next step?

1. Consider a justified nonlinear term or transformation, refit, and check residuals and held-out performance. **[correct]**
2. Delete all parcels with positive residuals.
   Hint: A systematic model issue should not be removed by selecting convenient observations.
3. Keep the straight line because a large R-squared guarantees the form is correct.
   Hint: R-squared cannot rule out structure remaining in residuals.

Explanation: A quadratic term or a suitable transformation may represent the curvature, depending on shipping economics. Check the revised diagnostics and prediction performance rather than assuming a particular fix is guaranteed.
## An omitted common cause in ice-cream and sunscreen sales

ID: `regprob_omitted_variable_causation`; family `concept_check`, rule 1; status **active**.

Across summer days, a beach retailer observes a positive regression slope between ice-cream sales and sunscreen sales. Temperature was not recorded. Hotter days plausibly increase demand for both products. The retailer did not randomly change ice-cream sales or assign treatments.

Assumptions:

- The data are observational and temperature is a plausible common cause.
- A plausible omitted variable is a hypothesis to investigate, not proof of the entire causal mechanism.

### 1. omitted_common_cause

Which explanation most directly challenges a causal interpretation of the slope?

1. Temperature may influence both sales series, creating association without an ice-cream-to-sunscreen causal effect. **[correct]**
2. A positive slope proves buying ice cream makes customers buy sunscreen.
   Hint: The observational relationship can arise through a common cause.
3. A significant p-value eliminates all omitted-variable explanations.
   Hint: Statistical significance is conditional on a model and does not establish causal identification.

Explanation: Temperature is a plausible confounder: it relates to the predictor and may independently affect the response. The slope can capture this omitted common influence.

### 2. omitted_variable_analysis_step

Which next step would help assess this concern?

1. Collect temperature and other justified variables, examine the adjusted association, and state the limits of observational control. **[correct]**
2. Change sales units from dollars to cents to remove confounding.
   Hint: Rescaling changes coefficients' units, not the causal structure.
3. Add random unrelated predictors until R-squared is large.
   Hint: Variables need a defensible causal or predictive role.

Explanation: Including temperature can test sensitivity to a plausible common cause, but measurement error and unmeasured factors can remain. Adjustment alone is not automatic proof of causation.

### 3. association_causal_report

Which business statement is defensible from the original regression?

1. Ice-cream and sunscreen sales are associated in these observations; the data alone do not establish that increasing ice-cream sales would raise sunscreen sales. **[correct]**
2. Doubling ice-cream sales will necessarily double sunscreen sales.
   Hint: The observational slope is not a guaranteed intervention effect or a fixed proportional law.
3. The association is impossible unless the relationship is causal.
   Hint: Common causes can generate association between two outcomes.

Explanation: A regression can support prediction under suitable validation while leaving causal attribution unresolved. A causal claim requires a design and assumptions that address confounding and the proposed intervention.
## Investigating a store that changes the fitted slope

ID: `regprob_outlier_influence`; family `concept_check`, rule 1; status **active**.

A store-level regression of sales on advertising has a positive slope for the main cluster. A distant high-advertising store has low sales; including it changes the fitted slope to negative. The store is much farther along the predictor axis than the other stores. The analyst has not yet checked the record.

Assumptions:

- The change in fitted slope is based on a sensitivity comparison with and without the store.
- Distance along X concerns leverage; influence concerns changes to the fitted model.
- An unusual observation may be valid and must not be removed solely to obtain a preferred result.

### 1. leverage_vs_influence

Which diagnosis is supported by the description?

1. The store has high leverage and is influential, since its inclusion materially changes the fitted slope. **[correct]**
2. Every point with high leverage must be influential in every dataset.
   Hint: Leverage creates potential influence; the observed slope change demonstrates influence here.
3. The store has low leverage because its sales are low.
   Hint: Leverage depends on predictor positions, not whether Y is low.

Explanation: The unusual X value creates high leverage. The observed change in slope is direct sensitivity evidence of influence. A vertical outlier, leverage, and influence are related but distinct.

### 2. influential_store_investigation

What should the analyst do first?

1. Check data accuracy and business context, inspect diagnostics, and report sensitivity to the store. **[correct]**
2. Delete the store immediately so the desired positive slope returns.
   Hint: A preferred result is not a valid data-exclusion criterion.
3. Hide the model including the store from the report.
   Hint: Material sensitivity belongs in an honest analysis.

Explanation: Verify units, data entry and whether the store belongs to the target population. If valid, it may reveal a distinct regime or an inadequate model. Any exclusion needs a defensible, documented reason.

### 3. influence_reporting_limit

What does the slope reversal imply for interpretation?

1. The association is sensitive to one observation; avoid presenting a stable or causal advertising effect from this fit alone. **[correct]**
2. Advertising has been proved to cause lower sales everywhere.
   Hint: A sensitive observational slope cannot establish causation or universal effects.
3. The main-cluster fit is automatically the true population model.
   Hint: The cluster also requires a justified population and modeling assumptions.

Explanation: Report both sensitivity and uncertainty. Consider model form, population scope, robust analysis and more data as appropriate. The direction of a fitted coefficient is not a causal guarantee.
## Comparing independent store-layout means with stipulated pooling

ID: `twosample_independent_means_layouts`; family `two_sample_test`, rule 2; status **active**.

A pilot compares basket totals from independent random samples under two store layouts. Layout A: n=20, sample mean $52, sample variance 36 dollars squared. Layout B: n=20, sample mean $48, sample variance 64 dollars squared. Assume normal populations with equal population variances; use the pooled t test of equal means versus a two-sided difference at alpha=0.05. This pooled convention is stipulated for this exercise; the course's pooled-versus-Welch convention remains unconfirmed.

Assumptions:

- The groups and all observations are independent.
- Normality and equal population variances are assumed, not deduced from the sample summaries.
- A minus B defines the statistic, with a two-sided alternative.

### 1. independent_layout_pooled_variance

What is the pooled variance in dollars squared? Select the value rounded to four decimal places.

1. $50.0000$ **[correct]**
2. $100.0000$
   Hint: Pool with degrees-of-freedom weights rather than adding variances directly.
3. $7.0711$
   Hint: That is the pooled SD, not variance.

Explanation: $s_p^2=(19(36)+19(64))/38=50$. Excel: `=((20-1)*36+(20-1)*64)/(20+20-2)`. Residual df is 38.

Engine quantity `pooled_variance`: 50. Four-decimal choices; exact Go value retained for audit.

### 2. independent_layout_two_tail_pvalue

What is the two-sided p-value? Select the value rounded to four decimal places.

1. $0.0816$ **[correct]**
2. $0.0408$
   Hint: Do not report just one tail for a two-sided alternative.
3. $0.0500$
   Hint: Calculate the p-value; alpha is the threshold.

Explanation: $SE=\sqrt{50(1/20+1/20)}=\sqrt5$ and $t=4/\sqrt5=1.7889$. Excel: `=T.DIST.2T(ABS(4/SQRT(5)),38)` or `=T.TEST(A2:A21,B2:B21,2,2)` on raw samples.

Engine quantity `p_two`: 0.0816124166095. Four-decimal choices; exact Go value retained for audit.

### 3. independent_layout_nonsignificance

At alpha=0.05, which conclusion is appropriate?

1. Fail to reject equal population means; these samples provide insufficient evidence of a difference at this threshold. **[correct]**
2. Accept that the two population means are exactly equal.
   Hint: Failure to reject does not prove equality.
3. Reject H0 because the observed sample means differ by $4.
   Hint: A sample difference must be assessed relative to its uncertainty.

Explanation: The p-value is about 0.0816, above 0.05. The observed difference is positive, but this test does not establish a nonzero population difference or demonstrate practical equivalence.
## Paired differences in sales before and after training

ID: `twosample_paired_training`; family `two_sample_test`, rule 2; status **active**.

Eight randomly sampled sales reps have before sales (units/week) 40, 42, 45, 39, 50, 48, 43, 46 and corresponding after sales 42, 46, 48, 44, 51, 52, 45, 49. Each position is the same rep. Define d=after-before. At alpha=0.05, test H0: mean difference=0 against H1: mean difference>0. Without an untreated randomized control group, this comparison alone does not establish training causation.

Assumptions:

- Differences from distinct reps are independent and approximately normally distributed for the small-sample t method.
- The alternative is right-sided for d=after-before, specified before seeing results.
- Pairing is by rep; the two measurement arrays are not independent samples.

### 1. paired_training_mean_difference

What is the mean after-minus-before difference in units per week? Select the value rounded to four decimal places.

1. $3.0000$ **[correct]**
2. $-3.0000$
   Hint: The stated difference is after minus before.
3. $24.0000$
   Hint: 24 is the sum of eight differences; divide by eight.

Explanation: Differences are 2,4,3,5,1,4,2,3. Their sum is 24, so mean difference is 3. Excel: subtract before from after in a helper column and use `=AVERAGE(C2:C9)`.

Engine quantity `mean_diff`: 3. Four-decimal choices; exact Go value retained for audit.

### 2. paired_training_t_statistic

What is the paired t statistic for the zero-difference null? Select the value rounded to four decimal places.

1. $6.4807$ **[correct]**
2. $2.2913$
   Hint: Divide SD by sqrt(8) to get the standard error of the mean difference.
3. $-6.4807$
   Hint: The positive mean difference gives a positive statistic.

Explanation: Squared deviations of differences from 3 sum to 12; $s_d^2=12/7$. Thus $t=3/\sqrt{(12/7)/8}=\sqrt{42}=6.4807$, df=7. Excel: `=AVERAGE(C2:C9)/(STDEV.S(C2:C9)/SQRT(COUNT(C2:C9)))`.

Engine quantity `t`: 6.48074069841. Four-decimal choices; exact Go value retained for audit.

### 3. paired_training_directional_pvalue

What is the p-value for the stated right-sided alternative? Select the value rounded to four decimal places.

1. $0.0002$ **[correct]**
2. $0.9998$
   Hint: Use the right tail because H1 specifies a positive mean difference.
3. $0.0500$
   Hint: Alpha is not the observed p-value.

Explanation: Use the signed t statistic in `=T.DIST.RT(t,7)`. Here the observed sign agrees with H1, so the one-tail probability equals half the two-sided p-value. This halving shortcut would be wrong for a right-sided alternative with a negative observed t.

Engine quantity `p_right`: 0.000170123608404. Four-decimal choices; exact Go value retained for audit.

### 4. paired_training_evidence_limit

What conclusion is supported?

1. Reject the zero-difference null; mean after sales are higher under the model, but the before/after design alone does not prove training caused the increase. **[correct]**
2. Treat the 16 measurements as independent observations with df=14.
   Hint: There are eight independent differences, giving df=7.
3. The small p-value proves all reps improve and the training is the sole cause.
   Hint: A population mean comparison neither guarantees every individual outcome nor eliminates time confounding.

Explanation: The right-sided p-value is below 0.05. Pairing removes between-person levels from the difference, while temporal changes and other causes can still confound attribution.
## Reading Excel's pooled two-sample t-test output

ID: `twosample_read_excel_output`; family `two_sample_test`, rule 2; status **active**.

A manufacturer compares independent production lines A and B. Each line has 15 randomly sampled cycle times (minutes). The sample means are 78.4 and 72.1; sample variances are 24.5 and 20.2. Assume normal populations with equal variances. Test $H_0:\mu_A-\mu_B=0$ against $H_1:\mu_A-\mu_B\ne0$ at $\alpha=0.05$.

Excel-style output (A minus B):

- Pooled Variance: 22.3500
- df: 28
- t Stat: 3.6495
- P(T<=t) one-tail: 0.0005
- P(T<=t) two-tail: 0.0011
- t Critical one-tail: 1.7011
- t Critical two-tail: 2.0484

Assumptions:

- The samples are independent; times within each sample are independent.
- Populations are normal with equal variances; the pooled method is stipulated, not confirmed as the course convention.
- The alternative is two-sided and alpha was chosen before inspecting the data.

### 1. excel_two_tail_selection

Which p-value should the two-sided test report? Select the value rounded to four decimal places.

1. $0.0011$ **[correct]**
2. $0.0005$
   Hint: Do not halve a two-sided p-value.
3. $0.0500$
   Hint: Alpha is the decision threshold, not the observed p-value.

Explanation: Use the two-tail row because the alternative permits either sign. Excel: `=T.DIST.2T(ABS(t),28)` or `=T.TEST(A2:A16,B2:B16,2,2)` on the original data.

Engine quantity `p_two`: 0.00106634999474. Four-decimal choices; exact Go value retained for audit.

### 2. excel_two_tail_critical

Which positive critical t-value applies to this two-sided test at alpha 0.05? Select the value rounded to four decimal places.

1. $2.0484$ **[correct]**
2. $1.7011$
   Hint: The one-tail cutoff allocates all alpha to one tail.
3. $1.9600$
   Hint: Unknown population variances require a t distribution with df=28.

Explanation: Split alpha between two tails. Excel: `=T.INV.2T(0.05,28)`. Compare the absolute t statistic with this positive cutoff.

Engine quantity `critical_two`: 2.0484071418. Four-decimal choices; exact Go value retained for audit.

### 3. excel_output_test_conclusion

What conclusion follows from the correct p-value and critical value?

1. Reject H0; there is evidence that the population mean cycle times differ. **[correct]**
2. Fail to reject H0 because t Stat exceeds 1.
   Hint: Compare p with alpha, or |t| with the two-tail cutoff.
3. The two population means differ by exactly 6.3 minutes.
   Hint: The observed sample difference estimates an uncertain population difference.

Explanation: The two-tail p-value is below 0.05 and |t| exceeds the two-tail critical value. Both methods reject the same null. The sample difference is 6.3 minutes; rejection does not establish an exact population difference. The one-tail row is not appropriate for this stated alternative.
## A pooled two-proportion test for ad conversions

ID: `twosample_two_proportion_ads`; family `two_sample_test`, rule 2; status **active**.

A randomized ad experiment independently assigns 600 visitors to A and another 600 to B. A converts 48 visitors; B converts 72. Test H0:pA=pB versus H1:pA differs from pB at alpha=0.05 using the pooled two-proportion z test. Each visitor contributes one binary outcome.

Assumptions:

- The two groups are independent; each visitor is assigned once and contributes one outcome.
- Under H0 use the common pooled proportion; each group's expected successes and failures exceed 10.
- The statistic uses A minus B and the alternative is two-sided.

### 1. ad_test_pooled_proportion

What is the pooled proportion under the equal-proportions null? Select the value rounded to four decimal places.

1. $0.1000$ **[correct]**
2. $0.2000$
   Hint: Divide combined successes by combined trials, not by one group's sample size.
3. $0.0800$
   Hint: Use both groups under the common-proportion null.

Explanation: $\hat p=(48+72)/(600+600)=0.10$. Expected counts in each group are 60 conversions and 540 nonconversions, satisfying the stated normal-approximation threshold.

Engine quantity `pooled_p`: 0.1. Four-decimal choices; exact Go value retained for audit.

### 2. ad_test_two_tail_pvalue

What is the two-sided p-value? Select the value rounded to four decimal places.

1. $0.0209$ **[correct]**
2. $0.0105$
   Hint: The alternative includes differences in either direction; use both tails.
3. $0.0500$
   Hint: Alpha is specified separately from the observed evidence.

Explanation: $SE=\sqrt{0.1(0.9)(1/600+1/600)}=\sqrt{0.0003}$; $z=(0.08-0.12)/SE=-2.3094$. Excel: `=2*NORM.S.DIST(-ABS((48/600-72/600)/SQRT(0.1*0.9*(1/600+1/600))),TRUE)`.

Engine quantity `p_two`: 0.0209213353378. Four-decimal choices; exact Go value retained for audit.

### 3. ad_test_conversion_conclusion

What should the analyst report at alpha=0.05?

1. Reject equal population conversion proportions; B's observed conversion rate is 12% versus A's 8%. **[correct]**
2. B's population conversion rate is certainly exactly 4 percentage points higher.
   Hint: The sample difference estimates an uncertain population difference.
3. Fail to reject because the z statistic is negative.
   Hint: A two-sided test uses the magnitude of either sign.

Explanation: The two-sided p-value is about 0.0209. This supports a difference under the randomized study conditions. Distinguish the 4-percentage-point observed difference from relative increase (50%) and from an exact population effect.

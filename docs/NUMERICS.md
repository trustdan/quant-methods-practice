# Mathematical and numerical contracts

## Representation and normalization

Represent counts/support bounds as integers. Use exact rational values for simple reviewed discrete examples when practical; use finite floating-point values for numerical distribution calculations. Reject NaN/infinity and invalid ranges. Persist original input plus normalized value, units and grade-policy version. Do not apply the accounting app's integer-money rule to all statistics values.

Initial numeric parser accepts decimal, explicit percent, and simple integer fraction with a nonzero denominator; scientific notation may be added with tests. Locale is explicit: default decimal dot; ambiguous comma separators are rejected with useful guidance. A bare 37.5 is not silently interpreted as 37.5%. Expressions and arbitrary LaTeX are not evaluated. Return input-validation errors separately from wrong answers.

Display the rounding policy before submission. Numeric correctness uses a reviewed per-stage policy such as `abs(x-y) <= max(absTol, relTol*abs(y))`, with appropriate near-zero behavior. Rounded-answer tasks also compare with the interval implied by stated rounding, using a documented tie convention. Do not set a global permissive tolerance that accepts a neighboring discrete event or wrong distribution.

For the draft binomial fixture, exact canonical probability is 3/8; absolute tolerance is 0.0000005, relative tolerance 0.000001, and no mandatory rounding is specified. This illustrates a policy; approve family-specific policies before production.

## Initial family rules

| Family | Preconditions / canonical checks |
|---|---|
| Set probability | All supplied probabilities in [0,1]; max(0,P(A)+P(B)-1) <= P(A intersection B) <= min(P(A),P(B)); union subtracts overlap. Disjointness does not imply independence. |
| Binomial | Integer n>=0, 0<=p<=1, integer k; iid Bernoulli trials for the model. PMF outside support is 0. Mean np; variance np(1-p); SD is square root of variance. |
| Poisson | Lambda>=0 for stated exposure, integer count. Rate times exposure must reconcile units. PMF outside support is 0; mean/variance lambda; handle lambda=0 explicitly. |
| Linear combination | E[aX+bY+c]=aE[X]+bE[Y]+c when moments exist, regardless of independence. Variance=a^2 Var(X)+b^2 Var(Y)+2ab Cov(X,Y); constant adds no variance. |

The sum of independent binomials is binomial only when their success probability is the same. Independent Poissons sum to a Poisson with summed rates. Independence is sufficient for zero covariance when the relevant moments exist; zero covariance alone does not generally establish independence. Do not encode false inverse rules.

## Stable computation

Avoid naive factorial evaluation and underflow-prone tiny tail subtraction. Choose reviewed numerical methods/library APIs and verify their parameterization from official documentation. Use recurrence/log-gamma or well-tested distribution functions; compute survival functions directly for extreme upper tails where available. Record error bounds and supported parameter domains; finite parameter limits are acceptable initially.

Tail translation is explicit: at least k is P(X>=k), more than k is P(X>=k+1), at most k is P(X<=k), fewer than k is P(X<=k-1). Test k=0, k at support edges and degenerate p/lambda cases.

Simulation is illustration, never the canonical grade. Seed simulations and show trial count/random uncertainty. Plot values and highlighted areas are derived from the canonical model; avoid displaying a simulation estimate as exact probability.

## Later methods

For continuous families record whether endpoints matter, density versus probability, CDF/quantile conventions, scale versus rate, and units. For CIs/tests record population/sample status, sample versus population SD, df, one/two-tail policy, paired design, null/alternative direction and critical-value method. Regression records intercept, residual sign, units, df, estimator and assumptions. Implement only course-grounded methods with reviewed fixtures.

## Required fixtures

Binomial n=4,p=.5,k=2 -> .375; probabilities sum to 1 over finite support; complement/tail identities agree. Poisson zero-rate boundaries and independent sums agree with reviewed analytic fixtures. Set union validates overlap bounds. Variance-sum cases distinguish equal means but different dependence. A numerically plausible result from the wrong event must fail structured grading.

Use independent hand-derived/exact fixtures for small cases and authoritative numerical references for harder tails. Do not have an LLM or the same implementation generate its own sole expected answers.

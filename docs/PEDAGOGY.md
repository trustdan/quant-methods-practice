# Progressive teaching and problem sets

## Seven decisions from one scenario

A scenario is one coherent problem with several stages, not seven unrelated questions. Keep context visible and ask one decision at a time. Stage names and counts are family-specific; do not invent trivial steps simply to reach seven. Early drills avoid free-form equation entry.

Example draft: a fair coin is tossed four times independently; find the probability of exactly two heads.

| Stage | Learner decision | Expected reasoning |
|---|---|---|
| 1 Target | What random quantity are we counting? | X is the number of heads in four tosses. |
| 2 Model | Which distribution fits? | Binomial, not Poisson or continuous normal. |
| 3 Conditions/parameters | Which conditions and parameter tuple apply? | Fixed n=4, independent trials, constant p=0.5, two outcomes per trial. |
| 4 Event | Translate exactly two. | X=2, not X<=2 or X>=2. |
| 5 Expression | Choose the probability expression. | C(4,2)(0.5)^2(0.5)^2; the combination factor counts possible orders. |
| 6 Calculation | Enter the probability. | 0.375, 37.5%, or 3/8 under the displayed input policy. |
| 7 Interpretation | Explain what the number refers to. | Probability of two heads in a four-toss experiment; not a guarantee or probability of one ordered sequence. |

The starting JSON fixture encodes choices and a numeric field for these decisions. Engine derivation, not this document, becomes authoritative during implementation.

Other seven-stage examples:

- Set probability: define events; translate language; select union/intersection; identify overlap; choose rule; compute; check bounds/interpret.
- Poisson: identify count and interval; verify rate-model conditions; convert rate/exposure; define event; choose PMF/tail; compute; interpret units.
- Sums: define each variable; establish coefficients; check relevant dependence; compute expectation; choose variance/covariance rule; calculate SD if asked; interpret.
- Confidence interval later: identify estimand; check design/assumptions; select method; find SE/critical value; compute interval; interpret; evaluate a tempting incorrect claim.
- Testing later: state hypotheses; check assumptions; choose statistic/tail; compute; assess p-value against alpha; conclude in context; distinguish statistical/practical significance.

## Feedback and misconceptions

First error: one short hint that directs reasoning without giving the numeric answer. One assisted retry. Second error: concise explanation and revealed answer; then proceed. Offer a reviewed contrast after eligible mistakes. Invalid syntax is validation feedback, not conceptual failure.

Prioritize disjoint versus independent, union double counting, conditional denominator, exactly versus at least/at most, missing binomial coefficient, wrong rate interval, variance versus SD, variance-of-sums without independence, parameter versus statistic, CI interpretation, failure-to-reject versus proving null, correlation versus causation. Distractors must reflect a plausible error and carry a reviewed causal hint; random filler choices add little value.

## Fading and navigation

Full: up to seven decisions. Intermediate: combine target/model/conditions, then event/expression, calculation, interpretation. Faded: structured complete answer and interpretation/assumptions check. Preserve concept attribution under different scaffolds; fewer stages do not mean weaker numerical validation.

Previous stages remain browsable. A learner cannot submit an already exposed answer as fresh independent evidence. Future-stage views that reveal formula or parameters mark assistance or remain unavailable until eligible. Reading old solutions or a full recap is review, not a new independent attempt. Navigation never silently reveals exam answers.

Reviewed contrasts change a reasoning condition, not only numbers or nouns. Examples: exactly two versus at least two; mutually exclusive events versus independent events; count over one hour versus three hours; sum under independence versus correlated variables. Metadata ties paired questions and transfer groups to reviewed versions. Guided comparisons are assistance for the entire problem and cannot chain indefinitely.

Explain notation by connecting it to the event first. Keep formulas readable, units visible, and rounded displays separate from full-precision grading. Never promise a clinical memory estimate or diagnose a learner from errors.

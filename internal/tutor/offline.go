package tutor

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// OfflineTutor provides curriculum-grounded, reviewed explanations, causal hints,
// and follow-up guidance entirely offline using mathematical engine invariants.
type OfflineTutor struct {
	ChunkDelay time.Duration
	Clock      func() time.Time
}

// NewOfflineTutor creates an OfflineTutor instance.
func NewOfflineTutor(chunkDelay time.Duration, clock func() time.Time) *OfflineTutor {
	if clock == nil {
		clock = time.Now
	}
	return &OfflineTutor{
		ChunkDelay: chunkDelay,
		Clock:      clock,
	}
}

func (t *OfflineTutor) ProviderID() string {
	return "offline"
}

func (t *OfflineTutor) Capabilities() ProviderCapabilities {
	return ProviderCapabilities{
		SupportsStreaming:    true,
		SupportsFollowUps:    true,
		SupportsCancellation: true,
		IsOffline:            true,
	}
}

// Stream generates structured markdown text and streams it as events over a channel.
func (t *OfflineTutor) Stream(ctx context.Context, req TutorRequest) (<-chan TutorEvent, error) {
	out := make(chan TutorEvent, 16)

	text := t.generateContent(req)

	go func() {
		defer close(out)

		now := t.Clock().UTC()
		// 1. Emit started event
		select {
		case <-ctx.Done():
			out <- TutorEvent{
				Type:       EventCancelled,
				RequestID:  req.RequestID,
				SessionID:  req.SessionID,
				InstanceID: req.InstanceID,
				StageID:    req.StageID,
				Timestamp:  t.Clock().UTC(),
			}
			return
		case out <- TutorEvent{
			Type:       EventStarted,
			RequestID:  req.RequestID,
			SessionID:  req.SessionID,
			InstanceID: req.InstanceID,
			StageID:    req.StageID,
			Timestamp:  now,
		}:
		}

		// 2. Stream text in chunks
		chunkSize := 30
		runes := []rune(text)
		for i := 0; i < len(runes); i += chunkSize {
			end := i + chunkSize
			if end > len(runes) {
				end = len(runes)
			}
			chunk := string(runes[i:end])

			select {
			case <-ctx.Done():
				out <- TutorEvent{
					Type:       EventCancelled,
					RequestID:  req.RequestID,
					SessionID:  req.SessionID,
					InstanceID: req.InstanceID,
					StageID:    req.StageID,
					Timestamp:  t.Clock().UTC(),
				}
				return
			case out <- TutorEvent{
				Type:       EventTextDelta,
				RequestID:  req.RequestID,
				SessionID:  req.SessionID,
				InstanceID: req.InstanceID,
				StageID:    req.StageID,
				Delta:      chunk,
				Timestamp:  t.Clock().UTC(),
			}:
			}

			if t.ChunkDelay > 0 {
				select {
				case <-ctx.Done():
					out <- TutorEvent{
						Type:       EventCancelled,
						RequestID:  req.RequestID,
						SessionID:  req.SessionID,
						InstanceID: req.InstanceID,
						StageID:    req.StageID,
						Timestamp:  t.Clock().UTC(),
					}
					return
				case <-time.After(t.ChunkDelay):
				}
			}
		}

		// 3. Emit complete event
		select {
		case <-ctx.Done():
			out <- TutorEvent{
				Type:       EventCancelled,
				RequestID:  req.RequestID,
				SessionID:  req.SessionID,
				InstanceID: req.InstanceID,
				StageID:    req.StageID,
				Timestamp:  t.Clock().UTC(),
			}
		case out <- TutorEvent{
			Type:       EventComplete,
			RequestID:  req.RequestID,
			SessionID:  req.SessionID,
			InstanceID: req.InstanceID,
			StageID:    req.StageID,
			Text:       text,
			Timestamp:  t.Clock().UTC(),
		}:
		}
	}()

	return out, nil
}

func (t *OfflineTutor) generateContent(req TutorRequest) string {
	switch req.Action {
	case ActionHint:
		return t.generateHint(req)
	case ActionExplain:
		return t.generateExplanation(req)
	case ActionFollowUp:
		return t.generateFollowUp(req)
	default:
		return t.generateExplanation(req)
	}
}

func (t *OfflineTutor) generateHint(req TutorRequest) string {
	templateID := ""
	if req.Instance != nil {
		templateID = req.Instance.TemplateID
	}
	if templateID == "" {
		templateID = req.InstanceID
	}
	stageID := req.StageID

	switch {
	case strings.Contains(templateID, "binomial"):
		if strings.Contains(stageID, "calculate") || strings.Contains(stageID, "prob") {
			return "### Causal Hint: Binomial Probability Structure\n\n" +
				"Remember that the Binomial PMF has three distinct multiplicative components:\n\n" +
				"1. **Combinatorial Factor** $\\binom{n}{k}$: counts the number of mutually exclusive arrangements where $k$ successes occur.\n" +
				"2. **Success Factor** $p^k$: the joint probability of the $k$ required successes.\n" +
				"3. **Failure Factor** $(1-p)^{n-k}$: the joint probability of the remaining $n-k$ failures.\n\n" +
				"*Check*: Did you include the combinations factor $\\binom{n}{k}$, and did you calculate $(1-p)$ for all unobserved trials?"
		}
		return "### Causal Hint: Event Modeling\n\n" +
			"Identify the two key conditions for a Binomial random variable:\n" +
			"- A fixed number of independent trials $n$.\n" +
			"- A constant probability of success $p$ on each trial.\n\n" +
			"Ask yourself: what quantity is physically varying across repetitions of this experiment?"

	case strings.Contains(templateID, "poisson"):
		return "### Causal Hint: Poisson Arrival Rate\n\n" +
			"In a Poisson distribution with parameter $\\lambda$:\n\n" +
			"- $\\lambda$ is the expected number of occurrences in the specified interval.\n" +
			"- The probability of observing exactly $k$ events is given by:\n" +
			"$$P(X = k) = \\frac{\\lambda^k e^{-\\lambda}}{k!}$$\n\n" +
			"*Check*: Note that for boundary case $k = 0$, $0! = 1$ and $\\lambda^0 = 1$, reducing the probability directly to $e^{-\\lambda}$."

	case strings.Contains(templateID, "conditional"):
		return "### Causal Hint: Conditional Probability\n\n" +
			"Conditional probability restricts the sample space to the conditioning event $B$:\n\n" +
			"$$P(A \\mid B) = \\frac{P(A \\cap B)}{P(B)}$$\n\n" +
			"*Check*: Ensure the denominator is the conditioning event $P(B)$, and the numerator is the joint occurrence $P(A \\cap B)$."

	case strings.Contains(templateID, "union"):
		return "### Causal Hint: Addition Rule\n\n" +
			"When combining two events $A$ and $B$, simply adding $P(A) + P(B)$ double-counts their overlap:\n\n" +
			"$$P(A \\cup B) = P(A) + P(B) - P(A \\cap B)$$\n\n" +
			"*Check*: Only if $A$ and $B$ are mutually exclusive (disjoint, $P(A \\cap B) = 0$) does the formula simplify to $P(A) + P(B)$."

	case strings.Contains(templateID, "linear") || strings.Contains(templateID, "variance"):
		return "### Causal Hint: Variance of Independent Sums\n\n" +
			"For independent random variables $X_1$ and $X_2$:\n\n" +
			"- Variances add directly: $\\text{Var}(X_1 + X_2) = \\text{Var}(X_1) + \\text{Var}(X_2)$.\n" +
			"- Standard deviations do **not** add directly: $\\sigma_{X_1+X_2} = \\sqrt{\\text{Var}(X_1) + \\text{Var}(X_2)} \\neq \\sigma_{X_1} + \\sigma_{X_2}$."

	default:
		return "### Causal Hint: Probabilistic Modeling\n\n" +
			"Break down the problem into its foundational components: identify the random variable $X$, its parameters, the underlying sample space $\\Omega$, and the governing probability rule $P(A)$ before substituting numeric values."
	}
}

func (t *OfflineTutor) generateExplanation(req TutorRequest) string {
	templateID := ""
	if req.Instance != nil {
		templateID = req.Instance.TemplateID
	}
	if templateID == "" {
		templateID = req.InstanceID
	}

	switch {
	case strings.Contains(templateID, "binomial_fair_coin_exactly_two"):
		return "### Step-by-Step Derivation: Binomial Probability\n\n" +
			"#### 1. Problem Identification\n" +
			"We are tossing a fair coin $n = 4$ times and seeking the probability of observing exactly $k = 2$ heads.\n" +
			"- Number of trials: $n = 4$\n" +
			"- Probability of heads on any trial: $p = 0.5$\n" +
			"- Probability of tails (failure): $q = 1 - p = 0.5$\n" +
			"- Target event: $X = 2$ heads\n\n" +
			"#### 2. Binomial PMF Formula\n" +
			"$$P(X = k) = \\binom{n}{k} p^k (1 - p)^{n - k}$$\n\n" +
			"#### 3. Step-by-Step Calculation\n" +
			"1. **Combinations**: Calculate the number of distinct sequences of 2 heads in 4 tosses:\n" +
			"$$\\binom{4}{2} = \\frac{4!}{2!(4-2)!} = \\frac{24}{2 \\times 2} = 6$$\n\n" +
			"2. **Success & Failure Powers**:\n" +
			"$$p^2 = (0.5)^2 = 0.25$$\n" +
			"$$(1-p)^{4-2} = (0.5)^2 = 0.25$$\n" +
			"$$p^2 (1-p)^2 = 0.25 \\times 0.25 = 0.0625 = \\frac{1}{16}$$\n\n" +
			"3. **Multiply Factors**:\n" +
			"$$P(X = 2) = 6 \\times 0.0625 = 0.375 = \\frac{6}{16} = \\frac{3}{8}$$\n\n" +
			"#### 4. Moments Verification\n" +
			"- Expected Value: $\\mu = np = 4 \\times 0.5 = 2$\n" +
			"- Variance: $\\sigma^2 = np(1-p) = 4 \\times 0.5 \\times 0.5 = 1$\n" +
			"- Standard Deviation: $\\sigma = \\sqrt{1} = 1$\n\n" +
			"The distribution is perfectly symmetric around $\\mu = 2$."

	case strings.Contains(templateID, "binomial_defective_at_most_one"):
		return "### Step-by-Step Derivation: Cumulative Binomial Probability\n\n" +
			"#### 1. Parameters\n" +
			"- Sample size: $n = 5$ items\n" +
			"- Defect rate: $p = 0.10$\n" +
			"- Non-defect rate: $1 - p = 0.90$\n" +
			"- Target event: \"At most one defective item\" $\\implies X \\le 1$\n\n" +
			"#### 2. Event Partition\n" +
			"Since $X$ is discrete, the event $\\{X \\le 1\\}$ consists of two disjoint outcomes:\n" +
			"$$P(X \\le 1) = P(X = 0) + P(X = 1)$$\n\n" +
			"#### 3. Calculations\n" +
			"- $P(X = 0) = \\binom{5}{0} (0.10)^0 (0.90)^5 = 1 \\times 1 \\times 0.59049 = 0.59049$\n" +
			"- $P(X = 1) = \\binom{5}{1} (0.10)^1 (0.90)^4 = 5 \\times 0.10 \\times 0.6561 = 0.32805$\n\n" +
			"#### 4. Sum\n" +
			"$$P(X \\le 1) = 0.59049 + 0.32805 = 0.91854$$\n\n" +
			"Thus, there is approximately a $91.85%$ probability of finding at most one defective unit."

	case strings.Contains(templateID, "poisson_call_center_arrivals"):
		return "### Step-by-Step Derivation: Poisson PMF\n\n" +
			"#### 1. Parameters\n" +
			"- Arrival rate: $\\lambda = 3.0$ calls per hour\n" +
			"- Target event: exactly $k = 2$ calls\n\n" +
			"#### 2. Poisson Formula\n" +
			"$$P(X = k) = \\frac{\\lambda^k e^{-\\lambda}}{k!}$$\n\n" +
			"#### 3. Calculation\n" +
			"$$P(X = 2) = \\frac{3^2 e^{-3}}{2!} = \\frac{9 \\times 0.049787}{2} = \\frac{0.44808}{2} \\approx 0.22404$$\n\n" +
			"#### 4. Theoretical Properties\n" +
			"For any Poisson distribution, $\\text{Mean} = \\text{Variance} = \\lambda = 3.0$."

	case strings.Contains(templateID, "set_probability_union_rule"):
		return "### Step-by-Step Derivation: Addition Rule for Unions\n\n" +
			"#### 1. Event Probabilities\n" +
			"- $P(A) = 0.40$\n" +
			"- $P(B) = 0.50$\n" +
			"- $P(A \\cap B) = 0.20$\n\n" +
			"#### 2. General Addition Rule\n" +
			"$$P(A \\cup B) = P(A) + P(B) - P(A \\cap B)$$\n\n" +
			"#### 3. Calculation\n" +
			"$$P(A \\cup B) = 0.40 + 0.50 - 0.20 = 0.70$$\n\n" +
			"#### 4. Why Subtract the Intersection?\n" +
			"Adding $P(A)$ and $P(B)$ counts the overlapping region $A \\cap B$ twice. Subtracting $P(A \\cap B)$ ensures every element in $A \\cup B$ is counted exactly once."

	case strings.Contains(templateID, "set_probability_conditional"):
		return "### Step-by-Step Derivation: Conditional Probability\n\n" +
			"#### 1. Given Probabilities\n" +
			"- $P(B) = 0.40$\n" +
			"- $P(A \\cap B) = 0.25$\n\n" +
			"#### 2. Definition of Conditional Probability\n" +
			"$$P(A \\mid B) = \\frac{P(A \\cap B)}{P(B)}$$\n\n" +
			"#### 3. Calculation\n" +
			"$$P(A \\mid B) = \\frac{0.25}{0.40} = \\frac{25}{40} = \\frac{5}{8} = 0.625$$\n\n" +
			"#### 4. Interpretation\n" +
			"Conditioning on $B$ reduces the relevant universe from the entire sample space $\\Omega$ down to $B$."

	case strings.Contains(templateID, "linear_comb_sum_variance_independent"):
		return "### Step-by-Step Derivation: Variance of Sums of Random Variables\n\n" +
			"#### 1. General Formula for Variance of Sums\n" +
			"$$\\text{Var}(aX_1 + bX_2) = a^2 \\text{Var}(X_1) + b^2 \\text{Var}(X_2) + 2ab \\text{Cov}(X_1, X_2)$$\n\n" +
			"#### 2. Independence Condition\n" +
			"When $X_1$ and $X_2$ are independent, $\\text{Cov}(X_1, X_2) = 0$. The formula reduces to:\n" +
			"$$\\text{Var}(X_1 + X_2) = \\text{Var}(X_1) + \\text{Var}(X_2)$$\n\n" +
			"#### 3. Evaluation\n" +
			"Given $\\text{Var}(X_1) = 0.21$ and $\\text{Var}(X_2) = 0.21$:\n" +
			"$$\\text{Var}(X_1 + X_2) = 0.21 + 0.21 = 0.42$$\n\n" +
			"Notice that while variances add directly, standard deviations do **not** add linearly: $\\sigma_{\\text{sum}} = \\sqrt{0.42} \\approx 0.648$, whereas $\\sigma_1 + \\sigma_2 = 2 \\sqrt{0.21} \\approx 0.917$."

	default:
		return "### Analytical Solution and Concept Review\n\n" +
			"#### Mathematical Structure\n" +
			"This problem evaluates key probability concepts under strict analytical axioms.\n\n" +
			"1. **Define Random Variable & Support**: State whether the distribution is discrete or continuous, and enumerate permissible values.\n" +
			"2. **State Required Conditions**: Check independence, constancy of rates/probabilities, and sample space partition.\n" +
			"3. **Execute Numerical Evaluation**: Apply the governing PMF/CDF formula, preserving exact arithmetic before rounding."
	}
}

func (t *OfflineTutor) generateFollowUp(req TutorRequest) string {
	switch req.FollowUpKind {
	case FollowUpExplainDifferently:
		return "### Intuitive & Visual Explanation\n\n" +
			"Imagine a decision tree where each trial branches into Success or Failure:\n\n" +
			"- Every complete path through the tree represents a sequence of trials.\n" +
			"- If trials are independent, the probability of any single branch sequence is the product of its individual probabilities ($p^k (1-p)^{n-k}$).\n" +
			"- Because multiple paths can produce the exact same total number of successes (for example, HHTT, HTHT, HTTH, THHT, THTH, TTHH all yield 2 heads), we multiply the path probability by the number of valid paths, which is given by $\\binom{n}{k}$.\n\n" +
			"So the formula is simply: **(Number of valid paths) $\\times$ (Probability of each path)**."

	case FollowUpWorkedExample:
		return "### Parallel Worked Example with Different Numbers\n\n" +
			"Consider $n = 3$ trials with success probability $p = 0.2$ (e.g. rolling a 1 on a 5-sided die), looking for $k = 1$ success:\n\n" +
			"1. **Combinations**: $\\binom{3}{1} = 3$ paths (SHH, HSH, HHS).\n" +
			"2. **Single Path Probability**:\n" +
			"$$p^1 \\times (1-p)^2 = 0.2 \\times (0.8)^2 = 0.2 \\times 0.64 = 0.128$$\n" +
			"3. **Total Probability**:\n" +
			"$$P(X = 1) = 3 \\times 0.128 = 0.384$$\n\n" +
			"Notice the exact same mechanics: combinations count arrangements, while powers compute individual path probabilities."

	case FollowUpWhyConditionMatters:
		return "### Why Do the Assumptions Matter?\n\n" +
			"The validity of these probability formulas hinges upon strict mathematical conditions:\n\n" +
			"1. **Independence**:\n" +
			"   - If trials were dependent (e.g., drawing cards without replacement), the probability on subsequent trials changes based on prior outcomes.\n" +
			"   - Violating independence makes simple multiplication $p^k$ invalid; you would need a Hypergeometric distribution instead of a Binomial distribution.\n\n" +
			"2. **Constant Success Probability ($p$)**:\n" +
			"   - If $p$ fluctuates between trials, the distribution becomes a Poisson-Binomial distribution, which lacks a simple single-parameter formula.\n\n" +
			"3. **Disjointness vs Independence**:\n" +
			"   - A common pitfall is confusing disjoint events ($A \\cap B = \\emptyset$) with independent events ($P(A \\cap B) = P(A)P(B)$).\n" +
			"   - In fact, if $P(A) > 0$ and $P(B) > 0$, **disjoint events can NEVER be independent**, because knowing $A$ occurred guarantees $B$ did not occur ($P(B \\mid A) = 0 \\neq P(B)$)!"

	case FollowUpCompareConcepts:
		return "### Comparative Concept Analysis: Binomial vs. Poisson\n\n" +
			"| Feature | Binomial Distribution | Poisson Distribution |\n" +
			"|---|---|---|\n" +
			"| **Trial Structure** | Fixed, discrete count $n$ | Continuous domain (time, space, volume) |\n" +
			"| **Support** | Bounded: $k \\in \\{0, 1, \\dots, n\\}$ | Unbounded: $k \\in \\{0, 1, 2, \\dots\\}$ |\n" +
			"| **Key Parameters** | $n$ (trials) and $p$ (probability) | $\\lambda$ (expected rate) |\n" +
			"| **Mean** | $\\mu = np$ | $\\mu = \\lambda$ |\n" +
			"| **Variance** | $\\sigma^2 = np(1-p) < \\mu$ | $\\sigma^2 = \\lambda = \\mu$ |\n" +
			"| **Relationship** | Exact for coin flips, component batches | Rare-event limit of Binomial ($n \\to \\infty, p \\to 0, np \\to \\lambda$) |\n\n" +
			"**Summary**: Use Binomial when you have a known total number of trials $n$. Use Poisson when counting events occurring at an average rate $\\lambda$ over a continuous interval without a fixed upper bound."

	case FollowUpCustom:
		if strings.TrimSpace(req.CustomPrompt) != "" {
			return fmt.Sprintf("### Inquiry: %q\n\n"+
				"In the context of this quantitative methods exercise:\n\n"+
				"- **Direct Answer**: Regarding your question, probability models require verifying that observed phenomena strictly satisfy distribution assumptions (such as independence, mutual exclusivity, and support bounds).\n"+
				"- **Core Principle**: Formulate the sample space $\\Omega$, express the desired outcome as a subset of elementary events, and apply the appropriate probability measure axioms.", req.CustomPrompt)
		}
		return "### Contextual Guidance\n\nFeel free to select one of the guided follow-up inquiries above or ask a specific question regarding formulas, assumptions, or calculations."

	default:
		return "### Additional Explanation\n\nReview the core axioms and mathematical rules of this distribution. Each formula is built directly from first principles of probability theory."
	}
}

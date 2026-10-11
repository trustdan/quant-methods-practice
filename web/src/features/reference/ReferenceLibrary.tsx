import React, { useState, useMemo } from 'react';
import { MathMarkdown } from '../../components/MathMarkdown';

export interface ReferenceEntry {
  id: string;
  title: string;
  category: 'foundations' | 'distributions' | 'properties';
  module: string;
  summary: string;
  formulaTex?: string;
  contentMarkdown: string;
  pitfalls?: string[];
}

export const REFERENCE_ENTRIES: ReferenceEntry[] = [
  {
    id: 'probability_axioms',
    title: 'Kolmogorov Probability Axioms',
    category: 'foundations',
    module: 'Module 1: Probability Foundations',
    summary: 'The three fundamental axioms grounding all mathematical probability measures.',
    formulaTex: 'P(\\Omega) = 1, \\quad P(A) \\ge 0, \\quad P\\left(\\bigcup_{i=1}^\\infty A_i\\right) = \\sum_{i=1}^\\infty P(A_i)',
    contentMarkdown: `
For any sample space $\\Omega$ and event $A \\subseteq \\Omega$:
1. **Non-negativity**: $P(A) \\ge 0$ for all events $A$.
2. **Total Probability**: $P(\\Omega) = 1$. The empty set has $P(\\emptyset) = 0$.
3. **Countable Additivity**: For mutually disjoint events $A_1, A_2, \\dots$ ($A_i \\cap A_j = \\emptyset$ for $i \\ne j$):
   $$P\\left(\\bigcup_{i=1}^\\infty A_i\\right) = \\sum_{i=1}^\\infty P(A_i)$$
    `,
    pitfalls: [
      'Confusing disjoint events (which cannot occur together) with independent events.',
      'Assigning probabilities strictly outside the closed interval $[0, 1]$.',
    ],
  },
  {
    id: 'complement_rule',
    title: 'Complement Rule',
    category: 'foundations',
    module: 'Module 1: Probability Foundations',
    summary: 'The probability that an event does not happen equals one minus its probability.',
    formulaTex: 'P(A^c) = 1 - P(A)',
    contentMarkdown: `
Since $A$ and $A^c$ partition the entire sample space $\\Omega$, $A \\cup A^c = \\Omega$ and $A \\cap A^c = \\emptyset$.
By additivity:
$$P(A) + P(A^c) = P(\\Omega) = 1 \\implies P(A^c) = 1 - P(A)$$
**Usage Tip**: Whenever a question asks for "at least one" success, computing $1 - P(\\text{none})$ is almost always vastly simpler than summing multiple terms.
    `,
    pitfalls: [
      'Forgetting that the complement of "at least one" is "exactly zero", not "all".',
    ],
  },
  {
    id: 'addition_union_rule',
    title: 'Addition Rule (Union of Events)',
    category: 'foundations',
    module: 'Module 1: Probability Foundations',
    summary: 'Probability of the union of any two events, correcting for double counting.',
    formulaTex: 'P(A \\cup B) = P(A) + P(B) - P(A \\cap B)',
    contentMarkdown: `
For any two events $A$ and $B$:
$$P(A \\cup B) = P(A) + P(B) - P(A \\cap B)$$
If $A$ and $B$ are **disjoint** (mutually exclusive, meaning $A \\cap B = \\emptyset$), then $P(A \\cap B) = 0$, simplifying to:
$$P(A \\cup B) = P(A) + P(B)$$
**Fréchet Bounds**: $P(A \\cup B)$ is bounded:
$$\\max(P(A), P(B)) \\le P(A \\cup B) \\le \\min(1, P(A) + P(B))$$
    `,
    pitfalls: [
      'Assuming $P(A \\cup B) = P(A) + P(B)$ without verifying whether the events are disjoint.',
      'Adding probabilities past 1 without subtracting the overlap $P(A \\cap B)$.',
    ],
  },
  {
    id: 'conditional_probability',
    title: 'Conditional Probability & Multiplication Rule',
    category: 'foundations',
    module: 'Module 1: Probability Foundations',
    summary: 'Probability of an event given that another event is known to have occurred.',
    formulaTex: 'P(A \\mid B) = \\frac{P(A \\cap B)}{P(B)}, \\quad P(B) > 0',
    contentMarkdown: `
The conditional probability of $A$ given condition $B$ is the proportion of the $B$ sample space occupied by $A$:
$$P(A \\mid B) = \\frac{P(A \\cap B)}{P(B)}$$
Rearranging yields the **General Multiplication Rule**:
$$P(A \\cap B) = P(A \\mid B) P(B) = P(B \\mid A) P(A)$$
    `,
    pitfalls: [
      'Reversing conditioning: $P(A \\mid B) \\ne P(B \\mid A)$ in general (Base Rate Fallacy / Confusion of the Inverse).',
    ],
  },
  {
    id: 'independence_vs_disjointness',
    title: 'Independence vs. Disjointness',
    category: 'foundations',
    module: 'Module 1: Probability Foundations',
    summary: 'Crucial distinction between statistical independence and mutual exclusivity.',
    formulaTex: 'P(A \\cap B) = P(A)P(B) \\quad \\text{vs.} \\quad A \\cap B = \\emptyset \\implies P(A \\cap B) = 0',
    contentMarkdown: `
- **Independent**: Occurrence of $B$ gives zero predictive information about $A$:
  $$P(A \\mid B) = P(A) \\iff P(A \\cap B) = P(A)P(B)$$
- **Disjoint (Mutually Exclusive)**: $A$ and $B$ cannot co-occur ($A \\cap B = \\emptyset$):
  $$P(A \\cap B) = 0$$
**Critical Axiom**: If $P(A) > 0$ and $P(B) > 0$, **disjoint events can NEVER be independent**! If $B$ occurs, $A$ is guaranteed impossible ($P(A \\mid B) = 0 \\ne P(A)$).
    `,
    pitfalls: [
      'Saying independent events "have nothing to do with each other" and therefore $P(A \\cap B) = 0$. That is disjointness, NOT independence!',
    ],
  },
  {
    id: 'binomial_distribution',
    title: 'Binomial Distribution',
    category: 'distributions',
    module: 'Module 2: Discrete Random Variables',
    summary: 'Counts successes in a fixed number of independent and identically distributed Bernoulli trials.',
    formulaTex: 'P(X = k) = \\binom{n}{k} p^k (1-p)^{n-k}, \\quad k \\in \\{0, 1, \\dots, n\\}',
    contentMarkdown: `
A random variable $X \\sim \\text{Binomial}(n, p)$ represents the number of successes in $n$ identical independent trials with constant success probability $p$.
- **PMF**:
  $$P(X = k) = \\frac{n!}{k!(n-k)!} p^k (1-p)^{n-k}$$
- **Moments**:
  $$E[X] = np, \\quad \\operatorname{Var}(X) = np(1-p), \\quad \\sigma_X = \\sqrt{np(1-p)}$$
- **Assumptions**: Fixed $n$, binary outcomes (success/failure), constant $p$, mutual independence across trials.
    `,
    pitfalls: [
      'Omitting the binomial coefficient $\\binom{n}{k}$ and computing only the probability of a single sequence ($p^k (1-p)^{n-k}$).',
      'Confusing $\\le k$ ("at most $k$") with $< k$ ("strictly fewer than $k$").',
    ],
  },
  {
    id: 'poisson_distribution',
    title: 'Poisson Distribution',
    category: 'distributions',
    module: 'Module 2: Discrete Random Variables',
    summary: 'Counts rare events occurring independently in a continuous interval of time or space at constant average rate.',
    formulaTex: 'P(X = k) = \\frac{\\lambda^k e^{-\\lambda}}{k!}, \\quad k \\in \\{0, 1, 2, \\dots\\}',
    contentMarkdown: `
A random variable $X \\sim \\text{Poisson}(\\lambda)$ models arrival counts where $\\lambda > 0$ is the mean rate of occurrence.
- **PMF**:
  $$P(X = k) = \\frac{\\lambda^k e^{-\\lambda}}{k!}$$
- **Moments**:
  $$E[X] = \\lambda, \\quad \\operatorname{Var}(X) = \\lambda, \\quad \\sigma_X = \\sqrt{\\lambda}$$
- **Key Characteristic**: Mean equals variance ($E[X] = \\operatorname{Var}(X) = \\lambda$).
- **Zero Events**: $P(X = 0) = e^{-\\lambda}$.
    `,
    pitfalls: [
      'Multiplying by an incorrect scaling factor when the time window changes: $\\lambda$ scales directly with duration $t$ ($\\lambda_t = \\lambda_1 \\cdot t$).',
      'Thinking $k$ has an upper bound: $k$ can be any non-negative integer $0, 1, 2, \\dots, \\infty$.',
    ],
  },
  {
    id: 'linear_combinations',
    title: 'Sums & Linear Combinations of Random Variables',
    category: 'properties',
    module: 'Module 2: Random Variables & Covariance',
    summary: 'Rules for expectation, variance, and covariance under linear transformation and addition.',
    formulaTex: 'E[aX + bY] = aE[X] + bE[Y], \\quad \\operatorname{Var}(aX + bY) = a^2\\operatorname{Var}(X) + b^2\\operatorname{Var}(Y) + 2ab\\operatorname{Cov}(X,Y)',
    contentMarkdown: `
Let $X$ and $Y$ be random variables, and $a, b, c \\in \\mathbb{R}$:
1. **Linearity of Expectation** (ALWAYS holds, independence NOT required):
   $$E[aX + bY + c] = aE[X] + bE[Y] + c$$
2. **Variance Scaling**:
   $$\\operatorname{Var}(aX + c) = a^2 \\operatorname{Var}(X)$$
3. **Variance of Sum**:
   $$\\operatorname{Var}(aX + bY) = a^2\\operatorname{Var}(X) + b^2\\operatorname{Var}(Y) + 2ab\\operatorname{Cov}(X, Y)$$
4. **Independent Variables**:
   If $X$ and $Y$ are independent, $\\operatorname{Cov}(X, Y) = 0$:
   $$\\operatorname{Var}(X + Y) = \\operatorname{Var}(X) + \\operatorname{Var}(Y)$$
   $$\\operatorname{Var}(X - Y) = \\operatorname{Var}(X) + \\operatorname{Var}(Y) \\quad (\\text{variances always ADD!})$$
    `,
    pitfalls: [
      'Subtracting variances: $\\operatorname{Var}(X - Y) \\ne \\operatorname{Var}(X) - \\operatorname{Var}(Y)$. Uncorrelated variances always add positive dispersion!',
      'Forgetting that variance scales quadratically: $\\operatorname{Var}(2X) = 4\\operatorname{Var}(X)$, whereas $\\operatorname{Var}(X_1 + X_2) = 2\\operatorname{Var}(X)$ for independent copies.',
    ],
  },
  {
    id: 'normal_distribution',
    title: 'Normal (Gaussian) Distribution & Z-Scores',
    category: 'distributions',
    module: 'Module 3: Continuous Probability Distributions',
    summary: 'The bell-shaped continuous distribution defined by mean mu and standard deviation sigma.',
    formulaTex: 'f(x) = \\frac{1}{\\sigma \\sqrt{2\\pi}} e^{-\\frac{(x-\\mu)^2}{2\\sigma^2}}, \\quad Z = \\frac{X - \\mu}{\\sigma}',
    contentMarkdown: `
A continuous random variable $X \\sim N(\\mu, \\sigma^2)$ with mean $\\mu$ and standard deviation $\\sigma > 0$.
- **Standardization ($Z$-Score)**:
  $$Z = \\frac{X - \\mu}{\\sigma} \\sim N(0, 1)$$
- **Cumulative Probability**:
  $$P(X \\le x) = \\Phi\\left(\\frac{x - \\mu}{\\sigma}\\right) = \\frac{1}{2}\\left[1 + \\operatorname{erf}\\left(\\frac{x - \\mu}{\\sigma \\sqrt{2}}\\right)\\right]$$
- **Upper Tail & Intervals**:
  $$P(X \\ge x) = 1 - \\Phi(z) = \\frac{1}{2}\\operatorname{erfc}\\left(\\frac{z}{\\sqrt{2}}\\right), \\quad P(a \\le X \\le b) = \\Phi(z_b) - \\Phi(z_a)$$
- **Empirical Rule (68-95-99.7)**:
  - $\\mu \\pm 1\\sigma$: $\\approx 68.27\\%$
  - $\\mu \\pm 2\\sigma$: $\\approx 95.45\\%$
  - $\\mu \\pm 3\\sigma$: $\\approx 99.73\\%$
- **Excel Formulas**:
  - CDF: \`=NORM.DIST(x, mu, sigma, TRUE)\` or \`=NORM.S.DIST(z, TRUE)\`
  - PDF: \`=NORM.DIST(x, mu, sigma, FALSE)\`
  - Inverse/Quantile: \`=NORM.INV(p, mu, sigma)\` or \`=NORM.S.INV(p)\`
    `,
    pitfalls: [
      'Confusing probability density f(x) with probability: for any continuous variable, P(X = c) = 0. Probability is the AREA under the curve, not height!',
      'Using unstandardized raw distances without dividing by sigma.',
    ],
  },
  {
    id: 'uniform_distribution',
    title: 'Continuous Uniform Distribution',
    category: 'distributions',
    module: 'Module 3: Continuous Probability Distributions',
    summary: 'Constant probability density over a finite continuous support interval [a, b].',
    formulaTex: 'f(x) = \\frac{1}{b - a}, \\quad a \\le x \\le b',
    contentMarkdown: `
A random variable $X \\sim U(a, b)$ with $a < b$:
- **PDF**:
  $$f(x) = \\frac{1}{b - a} \\quad \\text{for } a \\le x \\le b, \\quad 0 \\text{ otherwise}$$
- **CDF**:
  $$F(x) = \\frac{x - a}{b - a} \\quad \\text{for } a \\le x \\le b$$
- **Moments**:
  $$E[X] = \\frac{a + b}{2}, \\quad \\operatorname{Var}(X) = \\frac{(b - a)^2}{12}, \\quad \\sigma_X = \\frac{b - a}{\\sqrt{12}}$$
- **Interval Probability**:
  $$P(x_1 \\le X \\le x_2) = \\frac{x_2 - x_1}{b - a}$$
- **Endpoint Invariance**: $P(x_1 \\le X \\le x_2) = P(x_1 < X < x_2)$ because individual boundary points carry 0 probability.
    `,
    pitfalls: [
      'Setting density height to (b - a) instead of the reciprocal 1/(b - a). Area must equal 1.',
      'Treating continuous uniform endpoints like discrete integers.',
    ],
  },
  {
    id: 'exponential_distribution',
    title: 'Exponential Distribution & Memoryless Property',
    category: 'distributions',
    module: 'Module 3: Continuous Probability Distributions',
    summary: 'Continuous waiting times between Poisson events with constant hazard rate lambda.',
    formulaTex: 'f(x) = \\lambda e^{-\\lambda x}, \\quad P(X \\ge x) = e^{-\\lambda x}, \\quad x \\ge 0',
    contentMarkdown: `
A random variable $X \\sim \\text{Exp}(\\lambda)$ with rate $\\lambda > 0$:
- **PDF & CDF**:
  $$f(x) = \\lambda e^{-\\lambda x}, \\quad F(x) = 1 - e^{-\\lambda x}$$
- **Survival Function**:
  $$P(X \\ge x) = e^{-\\lambda x}$$
- **Moments**:
  $$E[X] = \\frac{1}{\\lambda}, \\quad \\operatorname{Var}(X) = \\frac{1}{\\lambda^2}, \\quad \\sigma_X = \\frac{1}{\\lambda}$$
- **Memoryless Property**:
  $$P(X > s + t \\mid X > s) = P(X > t) = e^{-\\lambda t}$$
  Given that an item has operated for $s$ hours, the probability of lasting an additional $t$ hours is identical to fresh survival.
- **Excel Formula**:
  - Cumulative: \`=EXPON.DIST(x, lambda, TRUE)\`
  - Density: \`=EXPON.DIST(x, lambda, FALSE)\`
    `,
    pitfalls: [
      'Confusing rate parameter lambda with mean duration: E[X] = 1/lambda, NOT lambda.',
      'Assuming mechanical wear-out: the exponential distribution models pure random failure without aging.',
    ],
  },
];

interface ReferenceLibraryProps {
  onClose?: () => void;
}

export const ReferenceLibrary: React.FC<ReferenceLibraryProps> = ({ onClose }) => {
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedCategory, setSelectedCategory] = useState<string>('all');
  const [expandedId, setExpandedId] = useState<string | null>(null);

  const filteredEntries = useMemo(() => {
    return REFERENCE_ENTRIES.filter((entry) => {
      const matchesCat = selectedCategory === 'all' || entry.category === selectedCategory;
      if (!matchesCat) return false;
      if (!searchQuery.trim()) return true;

      const q = searchQuery.toLowerCase();
      return (
        entry.title.toLowerCase().includes(q) ||
        entry.module.toLowerCase().includes(q) ||
        entry.summary.toLowerCase().includes(q) ||
        entry.contentMarkdown.toLowerCase().includes(q)
      );
    });
  }, [searchQuery, selectedCategory]);

  return (
    <div className="reference-library" style={{ display: 'flex', flexDirection: 'column', gap: '1.25rem' }}>
      {/* Header bar */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '0.75rem' }}>
        <div>
          <h2 style={{ fontSize: '1.4rem', fontWeight: 700, margin: 0 }}>Reference Library</h2>
          <p style={{ color: 'var(--text-muted)', fontSize: '0.9rem', margin: '0.2rem 0 0 0' }}>
            Authoritative course definitions, axioms, distributions, and moment properties. 100% offline.
          </p>
        </div>
        {onClose && (
          <button className="btn btn-outline" onClick={onClose}>
            &larr; Return to Practice
          </button>
        )}
      </div>

      {/* Search and Category Filters */}
      <div style={{ display: 'flex', gap: '0.75rem', flexWrap: 'wrap', alignItems: 'center' }}>
        <input
          type="search"
          placeholder="Search formulas, axioms, distributions, or rules..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          style={{
            flex: '1 1 280px',
            padding: '0.55rem 0.85rem',
            borderRadius: 'var(--radius-md)',
            border: '1px solid var(--border-subtle)',
            background: 'var(--bg-surface-elevated)',
            color: 'var(--text-main)',
            fontSize: '0.9rem',
          }}
        />
        <div style={{ display: 'flex', gap: '0.35rem', flexWrap: 'wrap' }}>
          {[
            { id: 'all', label: 'All Topics' },
            { id: 'foundations', label: 'Probability & Sets' },
            { id: 'distributions', label: 'Distributions' },
            { id: 'properties', label: 'Moments & Variance' },
          ].map((cat) => (
            <button
              key={cat.id}
              className={`btn ${selectedCategory === cat.id ? 'btn-primary' : 'btn-outline'}`}
              onClick={() => setSelectedCategory(cat.id)}
              style={{ fontSize: '0.8rem', padding: '0.35rem 0.65rem' }}
            >
              {cat.label}
            </button>
          ))}
        </div>
      </div>

      {/* Entries List */}
      <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
        {filteredEntries.length === 0 ? (
          <div className="card" style={{ textAlign: 'center', padding: '2rem', color: 'var(--text-muted)' }}>
            No reference topics match your search query "{searchQuery}".
          </div>
        ) : (
          filteredEntries.map((entry) => {
            const isExpanded = expandedId === entry.id;
            return (
              <div
                key={entry.id}
                className="card"
                style={{
                  background: 'var(--bg-surface)',
                  border: isExpanded ? '1px solid var(--accent-cyan)' : '1px solid var(--border-subtle)',
                  borderRadius: 'var(--radius-md)',
                  transition: 'border-color 0.15s ease',
                }}
              >
                {/* Entry Summary Header */}
                <div
                  onClick={() => setExpandedId(isExpanded ? null : entry.id)}
                  style={{
                    cursor: 'pointer',
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'flex-start',
                    gap: '1rem',
                  }}
                >
                  <div style={{ flex: 1 }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '0.3rem', flexWrap: 'wrap' }}>
                      <span className="badge badge-cyan" style={{ fontSize: '0.75rem' }}>{entry.module}</span>
                      <h3 style={{ fontSize: '1.15rem', fontWeight: 600, margin: 0 }}>{entry.title}</h3>
                    </div>
                    <p style={{ color: 'var(--text-muted)', fontSize: '0.9rem', margin: '0.2rem 0' }}>
                      {entry.summary}
                    </p>
                    {entry.formulaTex && (
                      <div
                        className="math-markdown"
                        style={{
                          margin: '0.5rem 0 0 0',
                          padding: '0.4rem 0.75rem',
                          background: 'var(--bg-surface-elevated)',
                          borderRadius: 'var(--radius-sm)',
                          display: 'inline-block',
                          maxWidth: '100%',
                          overflowX: 'auto',
                        }}
                      >
                        <MathMarkdown content={`$$${entry.formulaTex}$$`} />
                      </div>
                    )}
                  </div>
                  <button
                    className="btn btn-outline"
                    style={{ fontSize: '0.75rem', padding: '0.25rem 0.5rem', alignSelf: 'flex-start' }}
                  >
                    {isExpanded ? 'Collapse ▲' : 'Details ▼'}
                  </button>
                </div>

                {/* Expanded Details */}
                {isExpanded && (
                  <div
                    style={{
                      marginTop: '1rem',
                      paddingTop: '1rem',
                      borderTop: '1px solid var(--border-subtle)',
                      display: 'flex',
                      flexDirection: 'column',
                      gap: '0.85rem',
                    }}
                  >
                    <div style={{ fontSize: '0.95rem', lineHeight: 1.6 }}>
                      <MathMarkdown content={entry.contentMarkdown} />
                    </div>

                    {entry.pitfalls && entry.pitfalls.length > 0 && (
                      <div
                        style={{
                          background: 'rgba(239, 68, 68, 0.08)',
                          borderLeft: '4px solid var(--accent-red, #ef4444)',
                          borderRadius: 'var(--radius-sm)',
                          padding: '0.65rem 0.85rem',
                        }}
                      >
                        <div style={{ fontWeight: 600, color: 'var(--accent-red, #ef4444)', fontSize: '0.85rem', marginBottom: '0.3rem' }}>
                          Common Exam Pitfalls & Misconceptions:
                        </div>
                        <ul style={{ margin: 0, paddingLeft: '1.2rem', fontSize: '0.85rem', color: 'var(--text-main)' }}>
                          {entry.pitfalls.map((p, idx) => (
                            <li key={idx} style={{ margin: '0.2rem 0' }}>{p}</li>
                          ))}
                        </ul>
                      </div>
                    )}
                  </div>
                )}
              </div>
            );
          })
        )}
      </div>
    </div>
  );
};

import React, { useState, useEffect, useCallback, useRef } from 'react';
import { MathMarkdown } from '../components/MathMarkdown';
import { StageStrip, StageInfo } from '../components/StageStrip';
import { QuestionStrip } from '../components/QuestionStrip';
import { resolveKeyCommand, isEditableElement } from '../navigation/keymap';

const DEMO_STAGES: StageInfo[] = [
  { id: 'target', number: 1, label: 'Target' },
  { id: 'model', number: 2, label: 'Model' },
  { id: 'conditions', number: 3, label: 'Conditions' },
  { id: 'event', number: 4, label: 'Event' },
  { id: 'expression', number: 5, label: 'Expression' },
  { id: 'calculate', number: 6, label: 'Calculate' },
  { id: 'interpret', number: 7, label: 'Interpret' },
];

const STAGE_CONTENT: Record<string, { prompt: string; options?: string[]; mathNote: string }> = {
  target: {
    prompt: 'Identify the random variable and the explicit probability target for a scenario with $n = 3$ independent Bernoulli trials where $p = 0.5$.',
    options: [
      'Probability of exactly $k = 2$ successes: $P(X = 2)$',
      'Cumulative probability of at least 2 successes: $P(X \\ge 2)$',
      'Expected number of successes: $E[X] = np$',
      'Probability of zero successes: $P(X = 0)$',
    ],
    mathNote: 'The scenario asks for the exact count of successes. Let $X \\sim \\operatorname{Binomial}(n = 3, p = 0.5)$. Target is $P(X = 2)$.',
  },
  model: {
    prompt: 'Which parametric probability distribution models the discrete count of successes in a fixed number of trials?',
    options: [
      'Binomial: $X \\sim \\operatorname{Binomial}(n, p)$ with support $\{0, 1, \\dots, n\}$',
      'Poisson: $X \\sim \\operatorname{Poisson}(\\lambda)$ with countably infinite support $\{0, 1, 2, \\dots\}$',
      'Uniform: $X \\sim \\operatorname{DiscreteUniform}(a, b)$',
      'Normal approximation: $X \\sim \\mathcal{N}(\\mu, \\sigma^2)$',
    ],
    mathNote: 'Fixed number of trials $n = 3$ with binary outcomes yields a Binomial model with discrete finite support.',
  },
  conditions: {
    prompt: 'Verify the required mathematical conditions for the Binomial model:',
    options: [
      '1. Fixed $n$ trials; 2. Binary outcomes; 3. Constant $p$; 4. Mutual independence of trials',
      '1. Continuous support; 2. Constant hazard rate',
      '1. Sampling without replacement from a small finite population',
      '1. Trials must be strictly correlated',
    ],
    mathNote: 'Each Bernoulli trial must be independent with identical parameter $p$. Without independence, the trials do not sum to a Binomial distribution.',
  },
  event: {
    prompt: 'Construct the subset event corresponding to "at least one success" among $n$ trials:',
    options: [
      'Complement formulation: $E = \\{X \\ge 1\\}$, so $P(X \\ge 1) = 1 - P(X = 0)$',
      'Intersection formulation: $E = \\{X = 1\\} \\cap \\{X = n\\}$',
      'Exclusion formulation: $P(X \\ge 1) = P(X = 1) + P(X = 2)$ regardless of $n$',
      'Zero probability: $P(X \\ge 1) = 0$',
    ],
    mathNote: 'By the complement rule of probability: $P(A^c) = 1 - P(A)$. Here $P(X \\ge 1) = 1 - P(X = 0)$.',
  },
  expression: {
    prompt: 'Write the canonical PMF formula for evaluating $P(X = k)$ where $X \\sim \\operatorname{Binomial}(n, p)$:',
    options: [
      '$$P(X = k) = \\binom{n}{k} p^k (1 - p)^{n - k}$$',
      '$$P(X = k) = \\frac{\\lambda^k e^{-\\lambda}}{k!}$$',
      '$$P(X = k) = p (1 - p)^{k - 1}$$',
      '$$P(X = k) = \\frac{1}{n}$$',
    ],
    mathNote: 'The combinatorial factor $\\binom{n}{k} = \\frac{n!}{k!(n-k)!}$ counts the distinct sequences of $k$ successes and $n-k$ failures.',
  },
  calculate: {
    prompt: 'Calculate the exact canonical value for $n = 3, k = 2, p = 0.5$:',
    mathNote: 'Substitute parameters into formula:\n\n$$\\binom{3}{2} (0.5)^2 (1 - 0.5)^{3 - 2} = 3 \\times 0.25 \\times 0.5 = 0.375 = \\frac{3}{8}$$',
  },
  interpret: {
    prompt: 'Interpret the numerical result $P(X = 2) = 0.375$ in context:',
    options: [
      'In a long sequence of identical 3-trial experiments, exactly 2 successes will occur in approximately $37.5\\%$ of repetitions.',
      'In exactly 3 out of 8 future experiments, 2 successes are guaranteed.',
      'The average number of successes is $0.375$.',
      'The probability that $X$ exceeds 2 is $37.5\\%$.',
    ],
    mathNote: 'Frequentist probability represents long-run relative frequency across independent replications, never an absolute guarantee on finite runs.',
  },
};

export const App: React.FC = () => {
  const [currentStageIdx, setCurrentStageIdx] = useState(0);
  const [currentQuestionIdx, setCurrentQuestionIdx] = useState(0);
  const [selectedOption, setSelectedOption] = useState<number | null>(null);
  const [numericAnswer, setNumericAnswer] = useState('');
  const [showFeedback, setShowFeedback] = useState(false);
  const [showHelpModal, setShowHelpModal] = useState(false);
  const [serverStatus, setServerStatus] = useState<'checking' | 'connected' | 'offline'>('checking');
  const [serverVersion, setServerVersion] = useState<string>('0.1.0-dev');
  const [sessionAuthenticated, setSessionAuthenticated] = useState<boolean>(false);

  const numericInputRef = useRef<HTMLInputElement>(null);

  // Authenticate local session using bootstrap token in URL fragment if present
  useEffect(() => {
    const handleBootstrap = async () => {
      let token = '';
      if (window.location.hash.startsWith('#bootstrap=')) {
        token = window.location.hash.replace('#bootstrap=', '');
        // Clean up hash immediately so bootstrap token is not retained or leaked in history
        window.history.replaceState(null, '', window.location.pathname + window.location.search);
      }

      try {
        if (token) {
          const res = await fetch('/api/local-session', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ bootstrap_token: token }),
          });
          if (res.ok) {
            setSessionAuthenticated(true);
          }
        }

        const healthRes = await fetch('/api/health');
        if (healthRes.ok) {
          const healthData = await healthRes.json();
          setServerStatus('connected');
          if (healthData.version) {
            setServerVersion(healthData.version);
          }
        } else {
          setServerStatus('offline');
        }
      } catch {
        setServerStatus('offline');
      }
    };

    handleBootstrap();
  }, []);

  // Keyboard navigation handler complying with docs/NAVIGATION.md
  const handleKeyDown = useCallback((e: KeyboardEvent) => {
    const isFieldActive = isEditableElement(document.activeElement);
    const cmd = resolveKeyCommand(e, {
      isEditableActive: isFieldActive,
      activeMode: showHelpModal ? 'modal' : 'practice',
    });

    if (!cmd) return;

    switch (cmd) {
      case 'NAV_PREV_STAGE':
        e.preventDefault();
        setCurrentStageIdx((prev) => Math.max(0, prev - 1));
        setSelectedOption(null);
        setShowFeedback(false);
        break;
      case 'NAV_NEXT_STAGE':
        e.preventDefault();
        setCurrentStageIdx((prev) => Math.min(DEMO_STAGES.length - 1, prev + 1));
        setSelectedOption(null);
        setShowFeedback(false);
        break;
      case 'SELECT_CHOICE_1':
        e.preventDefault();
        setSelectedOption(0);
        break;
      case 'SELECT_CHOICE_2':
        e.preventDefault();
        setSelectedOption(1);
        break;
      case 'SELECT_CHOICE_3':
        e.preventDefault();
        setSelectedOption(2);
        break;
      case 'SELECT_CHOICE_4':
        e.preventDefault();
        setSelectedOption(3);
        break;
      case 'SUBMIT':
      case 'CONTINUE':
        e.preventDefault();
        setShowFeedback((prev) => !prev);
        break;
      case 'HINT':
        e.preventDefault();
        setShowFeedback(true);
        break;
      case 'HELP':
        e.preventDefault();
        setShowHelpModal((prev) => !prev);
        break;
      case 'ESCAPE':
        e.preventDefault();
        setShowHelpModal(false);
        break;
      case 'QUIT':
        if (!isFieldActive) {
          e.preventDefault();
          setShowHelpModal(true);
        }
        break;
    }
  }, [showHelpModal]);

  useEffect(() => {
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [handleKeyDown]);

  const currentStage = DEMO_STAGES[currentStageIdx];
  const stageData = STAGE_CONTENT[currentStage.id] || STAGE_CONTENT.target;

  return (
    <>
      {/* Top Header */}
      <header className="app-header">
        <div className="app-header-left">
          <div className="app-logo">
            <div className="app-logo-badge">Q</div>
            <span>Quant Methods Practice</span>
          </div>
          <span className="badge badge-cyan">Stage 01 Shell</span>
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem', fontSize: '0.85rem' }}>
            <span
              style={{
                width: 8,
                height: 8,
                borderRadius: '50%',
                backgroundColor: serverStatus === 'connected' ? '#34d399' : '#fbbf24',
                boxShadow: serverStatus === 'connected' ? '0 0 8px #34d399' : 'none',
              }}
            />
            <span style={{ color: 'var(--text-muted)' }}>
              {serverStatus === 'connected'
                ? `Loopback Server Active (v${serverVersion}${sessionAuthenticated ? ' / Auth' : ''})`
                : 'Offline Client'}
            </span>
          </div>
          <button
            className="btn btn-outline"
            onClick={() => setShowHelpModal(true)}
            title="Keyboard Shortcuts (F1)"
          >
            <span className="kbd">F1</span> Help
          </button>
        </div>
      </header>

      {/* Main Practice Workspace */}
      <main className="app-main">
        {/* Navigation Strips */}
        <div className="strip-container">
          <StageStrip
            stages={DEMO_STAGES}
            currentStageIndex={currentStageIdx}
            onSelectStage={(idx) => {
              setCurrentStageIdx(idx);
              setSelectedOption(null);
              setShowFeedback(false);
            }}
          />
          <QuestionStrip
            totalQuestions={5}
            currentQuestionIndex={currentQuestionIdx}
            onSelectQuestion={(idx) => {
              setCurrentQuestionIdx(idx);
              setCurrentStageIdx(0);
              setSelectedOption(null);
              setShowFeedback(false);
            }}
          />
        </div>

        {/* Static Demonstration Banner */}
        <div
          style={{
            padding: '0.75rem 1.25rem',
            background: 'rgba(2, 132, 199, 0.08)',
            border: '1px solid rgba(56, 189, 248, 0.25)',
            borderRadius: 'var(--radius-md)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            gap: '1rem',
          }}
        >
          <div>
            <div style={{ fontWeight: 600, color: 'var(--accent-cyan)', fontSize: '0.9rem' }}>
              Stage 01 Foundation: Mathematical Rendering &amp; Keyboard Shell Demonstration (Ungraded)
            </div>
            <div style={{ color: 'var(--text-muted)', fontSize: '0.8rem' }}>
              Static demonstration only — not an approved question or active curriculum drill. Verifies bundled MathJax typesetting and keyboard navigation.
            </div>
          </div>
          <span className="badge badge-emerald">Offline Bundle</span>
        </div>

        {/* Stage Content Card */}
        <div className="card">
          <div className="card-header">
            <div>
              <span className="badge badge-cyan" style={{ marginRight: '0.6rem' }}>
                Stage {currentStage.number}/7: {currentStage.label}
              </span>
              <span style={{ color: 'var(--text-muted)', fontSize: '0.85rem' }}>
                Question {currentQuestionIdx + 1} of 5
              </span>
            </div>
            <div style={{ display: 'flex', gap: '0.5rem' }}>
              <button
                className="btn btn-outline"
                disabled={currentStageIdx === 0}
                onClick={() => {
                  setCurrentStageIdx((prev) => Math.max(0, prev - 1));
                  setSelectedOption(null);
                  setShowFeedback(false);
                }}
              >
                <span className="kbd">h</span> Previous
              </button>
              <button
                className="btn btn-outline"
                disabled={currentStageIdx === DEMO_STAGES.length - 1}
                onClick={() => {
                  setCurrentStageIdx((prev) => Math.min(DEMO_STAGES.length - 1, prev + 1));
                  setSelectedOption(null);
                  setShowFeedback(false);
                }}
              >
                Next <span className="kbd">l</span>
              </button>
            </div>
          </div>

          {/* Render Prompt with LaTeX math */}
          <div style={{ fontSize: '1.05rem', marginBottom: '1.25rem' }}>
            <MathMarkdown content={stageData.prompt} />
          </div>

          {/* If stage has multiple choice options */}
          {stageData.options && (
            <div className="options-grid">
              {stageData.options.map((opt, idx) => (
                <button
                  key={idx}
                  className={`option-btn ${selectedOption === idx ? 'selected' : ''}`}
                  onClick={() => setSelectedOption(idx)}
                >
                  <span className="kbd">{idx + 1}</span>
                  <div style={{ flex: 1 }}>
                    <MathMarkdown content={opt} />
                  </div>
                </button>
              ))}
            </div>
          )}

          {/* If stage is numeric calculation (Stage 6) */}
          {currentStage.id === 'calculate' && (
            <div style={{ margin: '1.25rem 0', display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
              <label htmlFor="numeric-input" style={{ fontSize: '0.9rem', color: 'var(--text-muted)' }}>
                Enter exact probability value (decimal or fraction, e.g. <code>0.375</code> or <code>3/8</code>):
              </label>
              <div style={{ display: 'flex', gap: '0.75rem', maxWidth: '400px' }}>
                <input
                  id="numeric-input"
                  ref={numericInputRef}
                  type="text"
                  value={numericAnswer}
                  onChange={(e) => setNumericAnswer(e.target.value)}
                  placeholder="e.g. 0.375"
                  style={{
                    flex: 1,
                    padding: '0.6rem 0.9rem',
                    background: 'var(--bg-surface-elevated)',
                    border: '1px solid var(--border-strong)',
                    borderRadius: 'var(--radius-sm)',
                    color: 'var(--text-main)',
                    fontSize: '1rem',
                    fontFamily: 'var(--font-mono)',
                  }}
                />
                <button
                  className="btn btn-primary"
                  onClick={() => setShowFeedback(true)}
                >
                  Submit
                </button>
              </div>
              <div style={{ fontSize: '0.8rem', color: 'var(--text-subtle)' }}>
                Typing inside this input field preserves native number editing and does not trigger choice keys.
              </div>
            </div>
          )}

          {/* Action Row */}
          <div style={{ marginTop: '1.5rem', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <div style={{ display: 'flex', gap: '0.75rem' }}>
              <button
                className="btn btn-primary"
                onClick={() => setShowFeedback((prev) => !prev)}
              >
                <span className="kbd">Enter</span> {showFeedback ? 'Hide Feedback' : 'Check / Feedback'}
              </button>
              <button
                className="btn btn-secondary"
                onClick={() => setShowFeedback(true)}
              >
                <span className="kbd">?</span> Hint
              </button>
            </div>
            {selectedOption !== null && (
              <span style={{ fontSize: '0.85rem', color: 'var(--accent-cyan)' }}>
                Choice {selectedOption + 1} selected
              </span>
            )}
          </div>

          {/* Feedback & Mathematical Derivation Panel */}
          {showFeedback && (
            <div
              style={{
                marginTop: '1.5rem',
                padding: '1.25rem',
                background: 'rgba(15, 23, 42, 0.75)',
                border: '1px solid var(--border-strong)',
                borderRadius: 'var(--radius-md)',
              }}
            >
              <div style={{ fontWeight: 600, color: 'var(--accent-emerald)', marginBottom: '0.5rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                <span>Explanation &amp; Mathematical Derivation:</span>
              </div>
              <MathMarkdown content={stageData.mathNote} />
            </div>
          )}
        </div>

        {/* Static Mathematical Gallery Across Covered Topics */}
        <div className="card">
          <div className="card-header">
            <h3 className="card-title">Mathematical Formula Gallery (Local MathJax Validation)</h3>
            <span className="badge badge-cyan">Course Topics Preview</span>
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(320px, 1fr))', gap: '1rem' }}>
            <div style={{ padding: '1rem', background: 'var(--bg-surface-elevated)', borderRadius: 'var(--radius-sm)' }}>
              <div style={{ fontWeight: 600, color: 'var(--accent-cyan)', marginBottom: '0.4rem' }}>
                1. Set Operations &amp; Probability
              </div>
              <MathMarkdown
                content={
                  'Union: $P(A \\cup B) = P(A) + P(B) - P(A \\cap B)$\n\n' +
                  'Conditional: $P(A \\mid B) = \\frac{P(A \\cap B)}{P(B)}$'
                }
              />
            </div>
            <div style={{ padding: '1rem', background: 'var(--bg-surface-elevated)', borderRadius: 'var(--radius-sm)' }}>
              <div style={{ fontWeight: 600, color: 'var(--accent-cyan)', marginBottom: '0.4rem' }}>
                2. Discrete Distributions
              </div>
              <MathMarkdown
                content={
                  'Binomial: $$P(X = k) = \\binom{n}{k} p^k (1 - p)^{n - k}$$\n\n' +
                  'Poisson: $$P(X = k) = \\frac{\\lambda^k e^{-\\lambda}}{k!}$$'
                }
              />
            </div>
            <div style={{ padding: '1rem', background: 'var(--bg-surface-elevated)', borderRadius: 'var(--radius-sm)' }}>
              <div style={{ fontWeight: 600, color: 'var(--accent-cyan)', marginBottom: '0.4rem' }}>
                3. Sums &amp; Covariance
              </div>
              <MathMarkdown
                content={
                  'Linearity: $E[X + Y] = E[X] + E[Y]$\n\n' +
                  'Variance: $$\\operatorname{Var}(X + Y) = \\operatorname{Var}(X) + \\operatorname{Var}(Y) + 2\\operatorname{Cov}(X, Y)$$'
                }
              />
            </div>
            <div style={{ padding: '1rem', background: 'var(--bg-surface-elevated)', borderRadius: 'var(--radius-sm)' }}>
              <div style={{ fontWeight: 600, color: 'var(--accent-cyan)', marginBottom: '0.4rem' }}>
                4. Continuous Normal Distribution
              </div>
              <MathMarkdown
                content={
                  'Standard Normal density:\n\n' +
                  '$$\\phi(z) = \\frac{1}{\\sqrt{2\\pi}} e^{-\\frac{z^2}{2}}$$'
                }
              />
            </div>
          </div>
        </div>
      </main>

      {/* Footer / Status Bar */}
      <footer className="app-footer">
        <div className="key-hints">
          <div className="key-hint">
            <span className="kbd">h</span> / <span className="kbd">l</span>
            <span>Stage</span>
          </div>
          <div className="key-hint">
            <span className="kbd">1</span>-<span className="kbd">4</span>
            <span>Select</span>
          </div>
          <div className="key-hint">
            <span className="kbd">Enter</span>
            <span>Submit</span>
          </div>
          <div className="key-hint">
            <span className="kbd">?</span>
            <span>Hint</span>
          </div>
          <div className="key-hint">
            <span className="kbd">F1</span>
            <span>Help</span>
          </div>
        </div>
        <div>
          <span>Local loopback mode: 127.0.0.1</span>
        </div>
      </footer>

      {/* Keyboard Shortcut Help Modal */}
      {showHelpModal && (
        <div
          role="dialog"
          aria-modal="true"
          aria-label="Keyboard Shortcuts"
          style={{
            position: 'fixed',
            inset: 0,
            background: 'rgba(0, 0, 0, 0.75)',
            backdropFilter: 'blur(4px)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            zIndex: 100,
          }}
          onClick={() => setShowHelpModal(false)}
        >
          <div
            className="card"
            style={{ maxWidth: '540px', width: '90%', maxHeight: '85vh', overflowY: 'auto' }}
            onClick={(e) => e.stopPropagation()}
          >
            <div className="card-header">
              <h3 className="card-title">Keyboard Navigation Contract</h3>
              <button className="btn btn-outline" onClick={() => setShowHelpModal(false)}>
                <span className="kbd">Esc</span>
              </button>
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem', fontSize: '0.9rem' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', borderBottom: '1px solid var(--border-subtle)', paddingBottom: '0.4rem' }}>
                <span style={{ color: 'var(--text-muted)' }}>Previous / Next Stage:</span>
                <span><span className="kbd">h</span> / <span className="kbd">l</span> or <span className="kbd">&larr;</span> / <span className="kbd">&rarr;</span></span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', borderBottom: '1px solid var(--border-subtle)', paddingBottom: '0.4rem' }}>
                <span style={{ color: 'var(--text-muted)' }}>Choice Selection:</span>
                <span><span className="kbd">1</span> - <span className="kbd">4</span></span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', borderBottom: '1px solid var(--border-subtle)', paddingBottom: '0.4rem' }}>
                <span style={{ color: 'var(--text-muted)' }}>Submit / Continue Feedback:</span>
                <span><span className="kbd">Enter</span> or <span className="kbd">Space</span></span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', borderBottom: '1px solid var(--border-subtle)', paddingBottom: '0.4rem' }}>
                <span style={{ color: 'var(--text-muted)' }}>Request Hint / Explanation:</span>
                <span><span className="kbd">?</span> or <span className="kbd">e</span></span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', borderBottom: '1px solid var(--border-subtle)', paddingBottom: '0.4rem' }}>
                <span style={{ color: 'var(--text-muted)' }}>Toggle Help Modal:</span>
                <span><span className="kbd">F1</span></span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', borderBottom: '1px solid var(--border-subtle)', paddingBottom: '0.4rem' }}>
                <span style={{ color: 'var(--text-muted)' }}>Dismiss Modal / Return:</span>
                <span><span className="kbd">Esc</span></span>
              </div>
            </div>
            <div style={{ marginTop: '1.25rem', textAlign: 'right' }}>
              <button className="btn btn-primary" onClick={() => setShowHelpModal(false)}>
                Close Help
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  );
};

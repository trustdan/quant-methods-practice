import React from 'react';
import { DrillRecap } from '../../types/practice';
import { MathMarkdown } from '../../components/MathMarkdown';

interface DrillRecapViewProps {
  recap: DrillRecap;
  onRestart: () => void;
}

export const DrillRecapView: React.FC<DrillRecapViewProps> = ({ recap, onRestart }) => {
  const deriv = recap.canonical_derivation;

  return (
    <div className="drill-recap-container" style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
      {/* Performance Summary Banner */}
      <div
        className="card"
        style={{
          background: 'linear-gradient(135deg, rgba(16, 185, 129, 0.1) 0%, rgba(56, 189, 248, 0.08) 100%)',
          border: '1px solid rgba(52, 211, 153, 0.3)',
        }}
      >
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '1rem' }}>
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.6rem', marginBottom: '0.4rem' }}>
              <span className="badge badge-emerald">Drill Completed</span>
              <h2 style={{ fontSize: '1.3rem', fontWeight: 700, margin: 0 }}>{recap.title}</h2>
            </div>
            <div style={{ color: 'var(--text-muted)', fontSize: '0.9rem' }}>
              <MathMarkdown content={recap.scenario_markdown} />
            </div>
          </div>
          <button className="btn btn-primary" onClick={onRestart} style={{ alignSelf: 'flex-start' }}>
            <span className="kbd">Enter</span> Practice Again
          </button>
        </div>

        {/* Score Counters */}
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(140px, 1fr))', gap: '1rem', marginTop: '1.25rem' }}>
          <div style={{ background: 'var(--bg-surface-elevated)', padding: '0.75rem 1rem', borderRadius: 'var(--radius-sm)', border: '1px solid var(--border-subtle)' }}>
            <div style={{ color: 'var(--text-muted)', fontSize: '0.8rem' }}>Total Decisions</div>
            <div style={{ fontSize: '1.4rem', fontWeight: 700, color: 'var(--accent-cyan)' }}>{recap.total_stages} / 7</div>
          </div>
          <div style={{ background: 'var(--bg-surface-elevated)', padding: '0.75rem 1rem', borderRadius: 'var(--radius-sm)', border: '1px solid var(--border-subtle)' }}>
            <div style={{ color: 'var(--text-muted)', fontSize: '0.8rem' }}>First Try Correct</div>
            <div style={{ fontSize: '1.4rem', fontWeight: 700, color: 'var(--accent-emerald)' }}>{recap.first_try_count}</div>
          </div>
          <div style={{ background: 'var(--bg-surface-elevated)', padding: '0.75rem 1rem', borderRadius: 'var(--radius-sm)', border: '1px solid var(--border-subtle)' }}>
            <div style={{ color: 'var(--text-muted)', fontSize: '0.8rem' }}>Solved on Retry</div>
            <div style={{ fontSize: '1.4rem', fontWeight: 700, color: 'var(--accent-amber)' }}>{recap.retry_count}</div>
          </div>
          <div style={{ background: 'var(--bg-surface-elevated)', padding: '0.75rem 1rem', borderRadius: 'var(--radius-sm)', border: '1px solid var(--border-subtle)' }}>
            <div style={{ color: 'var(--text-muted)', fontSize: '0.8rem' }}>Solution Revealed</div>
            <div style={{ fontSize: '1.4rem', fontWeight: 700, color: '#c084fc' }}>{recap.revealed_count}</div>
          </div>
        </div>
      </div>

      {/* Canonical Mathematical Derivation Card */}
      {deriv && (
        <div className="card">
          <div className="card-header">
            <h3 className="card-title">Canonical Mathematical Derivation</h3>
            <span className="badge badge-cyan">Engine Derivation</span>
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))', gap: '1.25rem', marginBottom: '1.25rem' }}>
            <div style={{ background: 'var(--bg-surface-elevated)', padding: '1rem', borderRadius: 'var(--radius-sm)' }}>
              <div style={{ fontWeight: 600, color: 'var(--accent-cyan)', marginBottom: '0.4rem' }}>Target &amp; Parameters</div>
              <MathMarkdown
                content={`Model: $X \\sim \\operatorname{Binomial}(n=${deriv.n}, p=${deriv.p})$\n\nEvent: $${deriv.event_tex}$`}
              />
            </div>
            <div style={{ background: 'var(--bg-surface-elevated)', padding: '1rem', borderRadius: 'var(--radius-sm)' }}>
              <div style={{ fontWeight: 600, color: 'var(--accent-emerald)', marginBottom: '0.4rem' }}>Canonical Probability</div>
              <MathMarkdown
                content={`$$P(X=${deriv.k}) = ${deriv.canonical_probability} = ${deriv.canonical_rational || '3/8'} = 37.5\\%$$`}
              />
            </div>
            <div style={{ background: 'var(--bg-surface-elevated)', padding: '1rem', borderRadius: 'var(--radius-sm)' }}>
              <div style={{ fontWeight: 600, color: 'var(--accent-amber)', marginBottom: '0.4rem' }}>Theoretical Moments</div>
              <MathMarkdown
                content={`Mean $\\mu = np = ${deriv.mean}$\n\nVariance $\\sigma^2 = np(1-p) = ${deriv.variance}$\n\nStd Dev $\\sigma = ${deriv.std_dev}$`}
              />
            </div>
          </div>

          <div style={{ background: 'var(--bg-surface-elevated)', padding: '1rem', borderRadius: 'var(--radius-sm)' }}>
            <div style={{ fontWeight: 600, color: 'var(--text-main)', marginBottom: '0.5rem' }}>Analytical Formula &amp; Step-by-Step Evaluation:</div>
            <MathMarkdown
              content={`Formula: $${deriv.expression_tex}$\n\nCalculation: $${deriv.calculation_tex}$`}
            />
          </div>
        </div>
      )}

      {/* Seven Stages Detailed Review Cards */}
      <div className="card">
        <div className="card-header">
          <h3 className="card-title">Seven-Stage Decision Breakdown</h3>
          <span style={{ fontSize: '0.85rem', color: 'var(--text-muted)' }}>Progressive Pedagogical Sequence</span>
        </div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
          {recap.stage_summaries.map((stage) => {
            let outcomeBadge = <span className="badge badge-emerald">First Try ✓</span>;
            if (stage.outcome === 'retry') {
              outcomeBadge = <span className="badge badge-amber">Solved on Retry ⟳</span>;
            } else if (stage.outcome === 'revealed') {
              outcomeBadge = <span className="badge" style={{ background: 'rgba(192, 132, 252, 0.15)', color: '#c084fc', border: '1px solid rgba(192, 132, 252, 0.3)' }}>Revealed 👁</span>;
            }

            return (
              <div
                key={stage.stage_id}
                style={{
                  background: 'var(--bg-surface-elevated)',
                  border: '1px solid var(--border-subtle)',
                  borderRadius: 'var(--radius-sm)',
                  padding: '1rem',
                  display: 'flex',
                  flexDirection: 'column',
                  gap: '0.6rem',
                }}
              >
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                    <span className="badge badge-cyan">Stage {stage.stage_number}: {stage.label}</span>
                  </div>
                  {outcomeBadge}
                </div>

                <div style={{ fontSize: '0.95rem' }}>
                  <MathMarkdown content={stage.prompt_markdown} />
                </div>

                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))', gap: '0.75rem', fontSize: '0.9rem', marginTop: '0.25rem' }}>
                  <div style={{ background: 'rgba(0,0,0,0.2)', padding: '0.6rem 0.8rem', borderRadius: 'var(--radius-sm)' }}>
                    <div style={{ color: 'var(--text-muted)', fontSize: '0.75rem' }}>Your Answer</div>
                    <div style={{ fontWeight: 500 }}><MathMarkdown content={stage.learner_answer} /></div>
                  </div>
                  <div style={{ background: 'rgba(0,0,0,0.2)', padding: '0.6rem 0.8rem', borderRadius: 'var(--radius-sm)' }}>
                    <div style={{ color: 'var(--text-muted)', fontSize: '0.75rem' }}>Canonical Answer</div>
                    <div style={{ fontWeight: 500, color: 'var(--accent-emerald)' }}><MathMarkdown content={stage.canonical_answer} /></div>
                  </div>
                </div>

                {stage.misconception_triggered && (
                  <div style={{ background: 'rgba(251, 191, 36, 0.08)', border: '1px solid rgba(251, 191, 36, 0.25)', borderRadius: 'var(--radius-sm)', padding: '0.5rem 0.75rem', fontSize: '0.85rem', color: 'var(--accent-amber)' }}>
                    <strong>Targeted Misconception:</strong> <code>{stage.misconception_triggered}</code>
                  </div>
                )}

                <div style={{ color: 'var(--text-muted)', fontSize: '0.85rem', borderTop: '1px solid var(--border-subtle)', paddingTop: '0.5rem' }}>
                  <MathMarkdown content={stage.explanation_markdown} />
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
};

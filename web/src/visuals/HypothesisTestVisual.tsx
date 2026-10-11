import React, { useState, useEffect } from 'react';
import { HypothesisVisualResponse } from '../types/visuals';

export const HypothesisTestVisual: React.FC = () => {
  const [testType, setTestType] = useState<string>('mean_t');
  const [tail, setTail] = useState<string>('two_tailed');
  const [alpha, setAlpha] = useState<number>(0.05);
  const [sampleMean, setSampleMean] = useState<number>(53.0);
  const [data, setData] = useState<HypothesisVisualResponse | null>(null);

  useEffect(() => {
    let active = true;
    const fetchHypothesis = async () => {
      try {
        const params = new URLSearchParams({
          test_type: testType,
          tail,
          alpha: alpha.toString(),
          null_mu: '50.0',
          x_bar: sampleMean.toString(),
          s: '8.0',
          n: '25',
        });
        const res = await fetch(`/api/visuals/hypothesis_test?${params.toString()}`);
        if (res.ok && active) {
          const json = await res.json();
          setData(json);
        }
      } catch {
        // Fallback
      }
    };
    fetchHypothesis();
    return () => {
      active = false;
    };
  }, [testType, tail, alpha, sampleMean]);

  const svgWidth = 640;
  const svgHeight = 280;
  const padding = { top: 25, right: 30, bottom: 45, left: 45 };
  const plotWidth = svgWidth - padding.left - padding.right;
  const plotHeight = svgHeight - padding.top - padding.bottom;

  const minX = -4.0;
  const maxX = 4.0;
  const maxDensity = 0.42;

  const scaleX = (x: number) => padding.left + ((x - minX) / (maxX - minX)) * plotWidth;
  const scaleY = (d: number) => padding.top + plotHeight - (d / maxDensity) * plotHeight;

  let curvePath = '';
  if (data && data.points.length > 1) {
    curvePath = data.points.reduce((path, pt, i) => {
      const px = scaleX(pt.x);
      const py = scaleY(pt.pdf);
      return i === 0 ? `M ${px} ${py}` : `${path} L ${px} ${py}`;
    }, '');
  }

  // Build rejection shading polygon(s)
  const buildRejectionPath = () => {
    if (!data || data.points.length === 0) return '';
    let p = '';

    if (tail === 'right' || tail === 'two_tailed') {
      const crit = data.critical_value;
      const rightPts = data.points.filter((pt) => pt.x >= crit);
      if (rightPts.length > 0) {
        const startX = scaleX(rightPts[0].x);
        p += `M ${startX} ${scaleY(0)} `;
        rightPts.forEach((pt) => {
          p += `L ${scaleX(pt.x)} ${scaleY(pt.pdf)} `;
        });
        p += `L ${scaleX(rightPts[rightPts.length - 1].x)} ${scaleY(0)} Z `;
      }
    }

    if (tail === 'left' || tail === 'two_tailed') {
      const crit = data.critical_value_2 ?? -data.critical_value;
      const leftPts = data.points.filter((pt) => pt.x <= crit);
      if (leftPts.length > 0) {
        p += `M ${scaleX(leftPts[0].x)} ${scaleY(0)} `;
        leftPts.forEach((pt) => {
          p += `L ${scaleX(pt.x)} ${scaleY(pt.pdf)} `;
        });
        p += `L ${scaleX(leftPts[leftPts.length - 1].x)} ${scaleY(0)} Z `;
      }
    }

    return p;
  };

  const statX = data ? scaleX(Math.max(-3.95, Math.min(3.95, data.test_statistic))) : 0;

  return (
    <div className="visual-container" style={{ display: 'flex', flexDirection: 'column', gap: '1rem', width: '100%', maxWidth: '100%', boxSizing: 'border-box' }}>
      <div className="visual-header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '0.5rem' }}>
        <div>
          <h3 style={{ margin: 0, fontSize: '1.25rem', color: 'var(--color-text-primary, #f1f5f9)' }}>
            Hypothesis Testing & Rejection Region Lab
          </h3>
          <p style={{ margin: '0.25rem 0 0', fontSize: '0.875rem', color: 'var(--color-text-secondary, #94a3b8)' }}>
            Visualize critical rejection regions, test statistic position, and $p$-value decision rules.
          </p>
        </div>
      </div>

      {/* Control panel */}
      <div
        className="visual-controls"
        style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))',
          gap: '1rem',
          padding: '1rem',
          background: 'var(--color-surface, #1e293b)',
          borderRadius: '0.5rem',
          border: '1px solid var(--color-border, #334155)',
        }}
      >
        <div>
          <label style={{ display: 'block', fontSize: '0.75rem', fontWeight: 600, color: 'var(--color-text-secondary, #94a3b8)', marginBottom: '0.35rem' }}>
            Test Model
          </label>
          <select
            aria-label="Test model selector"
            value={testType}
            onChange={(e) => setTestType(e.target.value)}
            style={{ width: '100%', padding: '0.45rem', borderRadius: '0.25rem', background: 'var(--color-bg, #0f172a)', color: '#fff', border: '1px solid var(--color-border, #334155)' }}
          >
            <option value="mean_t">One-sample t-test (sigma unknown, s=8, n=25)</option>
            <option value="mean_z">One-sample Z-test (sigma known = 8, n=25)</option>
          </select>
        </div>

        <div>
          <label style={{ display: 'block', fontSize: '0.75rem', fontWeight: 600, color: 'var(--color-text-secondary, #94a3b8)', marginBottom: '0.35rem' }}>
            Alternative Tail Direction
          </label>
          <select
            aria-label="Tail direction selector"
            value={tail}
            onChange={(e) => setTail(e.target.value)}
            style={{ width: '100%', padding: '0.45rem', borderRadius: '0.25rem', background: 'var(--color-bg, #0f172a)', color: '#fff', border: '1px solid var(--color-border, #334155)' }}
          >
            <option value="two_tailed">Two-Tailed ($H_a: \mu \neq 50$)</option>
            <option value="right">Right-Tailed ($H_a: \mu &gt; 50$)</option>
            <option value="left">Left-Tailed ($H_a: \mu &lt; 50$)</option>
          </select>
        </div>

        <div>
          <label style={{ display: 'block', fontSize: '0.75rem', fontWeight: 600, color: 'var(--color-text-secondary, #94a3b8)', marginBottom: '0.35rem' }}>
            Significance Level $\alpha$
          </label>
          <div style={{ display: 'flex', gap: '0.5rem' }}>
            {[0.01, 0.05, 0.10].map((a) => (
              <button
                key={a}
                onClick={() => setAlpha(a)}
                style={{
                  flex: 1,
                  padding: '0.45rem',
                  borderRadius: '0.25rem',
                  border: alpha === a ? '2px solid #ef4444' : '1px solid var(--color-border, #334155)',
                  background: alpha === a ? 'rgba(239, 68, 68, 0.2)' : 'var(--color-bg, #0f172a)',
                  color: '#fff',
                  fontWeight: alpha === a ? 700 : 400,
                  cursor: 'pointer',
                }}
              >
                {a}
              </button>
            ))}
          </div>
        </div>

        <div>
          <label style={{ display: 'block', fontSize: '0.75rem', fontWeight: 600, color: 'var(--color-text-secondary, #94a3b8)', marginBottom: '0.35rem' }}>
            Sample Mean $\bar&#123;x&#125; = {sampleMean}$
          </label>
          <input
            aria-label="Sample mean slider"
            type="range"
            min="45"
            max="56"
            step="0.5"
            value={sampleMean}
            onChange={(e) => setSampleMean(parseFloat(e.target.value))}
            style={{ width: '100%' }}
          />
        </div>
      </div>

      {/* SVG Canvas */}
      <div
        className="visual-canvas-container"
        style={{
          background: 'var(--color-surface, #1e293b)',
          borderRadius: '0.5rem',
          border: '1px solid var(--color-border, #334155)',
          padding: '1rem',
          overflowX: 'auto',
        }}
      >
        <svg
          viewBox={`0 0 ${svgWidth} ${svgHeight}`}
          style={{ width: '100%', height: 'auto', display: 'block' }}
          role="img"
          aria-label={data?.summary || 'Hypothesis test visual'}
        >
          {/* Rejection region shading */}
          <path d={buildRejectionPath()} fill="rgba(239, 68, 68, 0.35)" stroke="none" />

          {/* Null distribution curve */}
          {curvePath && <path d={curvePath} fill="none" stroke="#38bdf8" strokeWidth="2.5" />}

          {/* Baseline axis */}
          <line
            x1={padding.left}
            y1={padding.top + plotHeight}
            x2={padding.left + plotWidth}
            y2={padding.top + plotHeight}
            stroke="var(--color-border, #475569)"
            strokeWidth="1.5"
          />

          {/* Null mean vertical center guide */}
          <line
            x1={scaleX(0)}
            y1={padding.top}
            x2={scaleX(0)}
            y2={padding.top + plotHeight}
            stroke="#64748b"
            strokeWidth="1"
            strokeDasharray="3 3"
          />

          {/* Test statistic marker */}
          {data && (
            <g>
              <line
                x1={statX}
                y1={padding.top}
                x2={statX}
                y2={padding.top + plotHeight}
                stroke="#a855f7"
                strokeWidth="2.5"
              />
              <circle cx={statX} cy={padding.top + 8} r="5" fill="#a855f7" />
              <text x={statX} y={padding.top - 8} textAnchor="middle" fill="#c084fc" fontSize="11" fontWeight="bold">
                Test Stat = {data.test_statistic}
              </text>
            </g>
          )}

          {/* Critical Value Cutoffs */}
          {data && (
            <>
              {(tail === 'right' || tail === 'two_tailed') && (
                <line
                  x1={scaleX(data.critical_value)}
                  y1={padding.top + 20}
                  x2={scaleX(data.critical_value)}
                  y2={padding.top + plotHeight}
                  stroke="#ef4444"
                  strokeWidth="2"
                  strokeDasharray="4 2"
                />
              )}
              {(tail === 'left' || tail === 'two_tailed') && (
                <line
                  x1={scaleX(data.critical_value_2 ?? -data.critical_value)}
                  y1={padding.top + 20}
                  x2={scaleX(data.critical_value_2 ?? -data.critical_value)}
                  y2={padding.top + plotHeight}
                  stroke="#ef4444"
                  strokeWidth="2"
                  strokeDasharray="4 2"
                />
              )}
            </>
          )}

          {/* X axis ticks */}
          {[-3, -2, -1, 0, 1, 2, 3].map((tick) => (
            <g key={tick}>
              <line
                x1={scaleX(tick)}
                y1={padding.top + plotHeight}
                x2={scaleX(tick)}
                y2={padding.top + plotHeight + 5}
                stroke="#64748b"
                strokeWidth="1"
              />
              <text x={scaleX(tick)} y={padding.top + plotHeight + 18} textAnchor="middle" fill="#94a3b8" fontSize="11">
                {tick}
              </text>
            </g>
          ))}
          <text x={padding.left + plotWidth / 2} y={svgHeight - 8} textAnchor="middle" fill="#94a3b8" fontSize="11">
            Standardized Test Statistic ($t$ or $Z$) under $H_0$
          </text>
        </svg>

        {/* Legend */}
        <div style={{ display: 'flex', gap: '1.25rem', justifyContent: 'center', marginTop: '0.75rem', fontSize: '0.8rem', color: '#cbd5e1', flexWrap: 'wrap' }}>
          <span style={{ display: 'flex', alignItems: 'center', gap: '0.35rem' }}>
            <span style={{ width: 14, height: 10, background: 'rgba(239, 68, 68, 0.4)', display: 'inline-block' }}></span>
            Rejection Region ($\alpha = {alpha}$)
          </span>
          <span style={{ display: 'flex', alignItems: 'center', gap: '0.35rem' }}>
            <span style={{ width: 16, height: 3, background: '#a855f7', display: 'inline-block' }}></span>
            Observed Test Statistic ({data?.test_statistic})
          </span>
          <span style={{ display: 'flex', alignItems: 'center', gap: '0.35rem' }}>
            <span style={{ width: 16, height: 2, borderTop: '2px dashed #ef4444', display: 'inline-block' }}></span>
            Critical Value Cutoff ({data?.critical_value})
          </span>
        </div>
      </div>

      {/* Decision Card */}
      {data && (
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))',
            gap: '0.75rem',
            alignItems: 'stretch',
          }}
        >
          <div
            id="ht-decision-badge"
            style={{
              background: data.reject_null ? 'rgba(239, 68, 68, 0.15)' : 'rgba(16, 185, 129, 0.15)',
              border: `2px solid ${data.reject_null ? '#ef4444' : '#10b981'}`,
              padding: '0.75rem 1rem',
              borderRadius: '0.375rem',
            }}
          >
            <div style={{ fontSize: '0.75rem', fontWeight: 600, color: data.reject_null ? '#fca5a5' : '#86efac' }}>
              STATISTICAL DECISION
            </div>
            <div style={{ fontSize: '1.35rem', fontWeight: 800, color: data.reject_null ? '#ef4444' : '#10b981', marginTop: '0.2rem' }}>
              {data.decision ? data.decision.toUpperCase() : ''}
            </div>
            <div style={{ fontSize: '0.8rem', color: '#cbd5e1', marginTop: '0.3rem' }}>
              $P$-value = {data.p_value} {data.reject_null ? '<' : '>'} $\alpha = {alpha}$
            </div>
          </div>

          <div style={{ background: 'var(--color-surface, #1e293b)', padding: '0.75rem', borderRadius: '0.375rem', border: '1px solid var(--color-border, #334155)' }}>
            <div style={{ fontSize: '0.75rem', color: '#94a3b8' }}>Rejection Rule</div>
            <div style={{ fontSize: '1rem', fontWeight: 600, color: '#f1f5f9', marginTop: '0.25rem' }}>
              Reject $H_0$ if {data.rejection_region}
            </div>
            <div style={{ fontSize: '0.75rem', color: '#64748b', marginTop: '0.2rem' }}>
              Standard Error $SE = {data.standard_error}$
            </div>
          </div>

          <div style={{ background: 'var(--color-surface, #1e293b)', padding: '0.75rem', borderRadius: '0.375rem', border: '1px solid var(--color-border, #334155)' }}>
            <div style={{ fontSize: '0.75rem', color: '#94a3b8' }}>Excel Formula</div>
            <code id="ht-excel-formula" style={{ fontSize: '0.8rem', color: '#fbbf24', display: 'block', marginTop: '0.25rem' }}>
              {data.excel_formula}
            </code>
          </div>
        </div>
      )}

      {/* Decision caution */}
      <div style={{ background: 'rgba(245, 158, 11, 0.08)', borderLeft: '4px solid #f59e0b', padding: '0.75rem 1rem', borderRadius: '0.25rem' }}>
        <p style={{ margin: 0, fontSize: '0.85rem', color: '#cbd5e1' }}>
          <strong>Critical Rule:</strong> If we fail to reject $H_0$, we never conclude that $H_0$ is proven true. We merely state that the sample data does not provide sufficient evidence to refute it at significance level $\alpha$.
        </p>
      </div>

      {/* ARIA summary */}
      <div role="status" aria-live="polite" className="sr-only" style={{ position: 'absolute', width: 1, height: 1, padding: 0, margin: -1, overflow: 'hidden', clip: 'rect(0, 0, 0, 0)', border: 0 }}>
        {data?.summary}
      </div>
    </div>
  );
};

import React, { useState, useEffect } from 'react';
import { ConfidenceIntervalVisualResponse } from '../types/visuals';

export const ConfidenceIntervalVisual: React.FC = () => {
  const [confLevel, setConfLevel] = useState<number>(0.95);
  const [sampleSize, setSampleSize] = useState<number>(16);
  const [numIntervals, setNumIntervals] = useState<number>(25);
  const [data, setData] = useState<ConfidenceIntervalVisualResponse | null>(null);

  useEffect(() => {
    let active = true;
    const fetchCI = async () => {
      try {
        const params = new URLSearchParams({
          conf_level: confLevel.toString(),
          mu: '50.0',
          sigma: '10.0',
          n: sampleSize.toString(),
          num_intervals: numIntervals.toString(),
          seed: '42',
        });
        const res = await fetch(`/api/visuals/confidence_interval?${params.toString()}`);
        if (res.ok && active) {
          const json = await res.json();
          setData(json);
        }
      } catch {
        // Fallback
      }
    };
    fetchCI();
    return () => {
      active = false;
    };
  }, [confLevel, sampleSize, numIntervals]);

  const svgWidth = 640;
  const rowHeight = 14;
  const svgHeight = Math.max(260, (data?.intervals.length || 25) * rowHeight + 60);
  const padding = { top: 30, right: 40, bottom: 35, left: 50 };
  const plotWidth = svgWidth - padding.left - padding.right;

  const minX = 40.0;
  const maxX = 60.0;
  const scaleX = (x: number) => padding.left + ((x - minX) / (maxX - minX)) * plotWidth;

  const trueMuX = scaleX(50.0);

  return (
    <div className="visual-container" style={{ display: 'flex', flexDirection: 'column', gap: '1rem', width: '100%', maxWidth: '100%', boxSizing: 'border-box' }}>
      <div className="visual-header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '0.5rem' }}>
        <div>
          <h3 style={{ margin: 0, fontSize: '1.25rem', color: 'var(--color-text-primary, #f1f5f9)' }}>
            Confidence Interval Coverage Explorer
          </h3>
          <p style={{ margin: '0.25rem 0 0', fontSize: '0.875rem', color: 'var(--color-text-secondary, #94a3b8)' }}>
            Observe how confidence level ($1 - \alpha$) governs the proportion of intervals that capture the true parameter $\mu = 50$.
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
            Confidence Level ($1 - \alpha$)
          </label>
          <div style={{ display: 'flex', gap: '0.5rem' }}>
            {[0.90, 0.95, 0.99].map((lvl) => (
              <button
                key={lvl}
                onClick={() => setConfLevel(lvl)}
                style={{
                  flex: 1,
                  padding: '0.45rem',
                  borderRadius: '0.25rem',
                  border: confLevel === lvl ? '2px solid #38bdf8' : '1px solid var(--color-border, #334155)',
                  background: confLevel === lvl ? 'rgba(56, 189, 248, 0.2)' : 'var(--color-bg, #0f172a)',
                  color: '#fff',
                  fontWeight: confLevel === lvl ? 700 : 400,
                  cursor: 'pointer',
                }}
              >
                {(lvl * 100).toFixed(0)}%
              </button>
            ))}
          </div>
        </div>

        <div>
          <label style={{ display: 'block', fontSize: '0.75rem', fontWeight: 600, color: 'var(--color-text-secondary, #94a3b8)', marginBottom: '0.35rem' }}>
            Sample Size: $n = {sampleSize}$ (df = {sampleSize - 1})
          </label>
          <input
            aria-label="Sample size slider"
            type="range"
            min="5"
            max="50"
            step="1"
            value={sampleSize}
            onChange={(e) => setSampleSize(parseInt(e.target.value, 10))}
            style={{ width: '100%' }}
          />
        </div>

        <div>
          <label style={{ display: 'block', fontSize: '0.75rem', fontWeight: 600, color: 'var(--color-text-secondary, #94a3b8)', marginBottom: '0.35rem' }}>
            Interval Count: {numIntervals}
          </label>
          <select
            aria-label="Interval count selector"
            value={numIntervals}
            onChange={(e) => setNumIntervals(parseInt(e.target.value, 10))}
            style={{ width: '100%', padding: '0.45rem', borderRadius: '0.25rem', background: 'var(--color-bg, #0f172a)', color: '#fff', border: '1px solid var(--color-border, #334155)' }}
          >
            <option value="15">15 intervals</option>
            <option value="25">25 intervals</option>
            <option value="40">40 intervals</option>
          </select>
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
          aria-label={data?.summary || 'Confidence interval simulation visual'}
        >
          {/* True parameter line mu = 50 */}
          <line
            x1={trueMuX}
            y1={padding.top - 5}
            x2={trueMuX}
            y2={svgHeight - padding.bottom + 5}
            stroke="#38bdf8"
            strokeWidth="2.5"
            strokeDasharray="4 3"
          />
          <text
            x={trueMuX}
            y={padding.top - 12}
            textAnchor="middle"
            fill="#38bdf8"
            fontSize="12"
            fontWeight="bold"
          >
            True $\mu = 50$
          </text>

          {/* Stacked intervals */}
          {data?.intervals.map((iv, idx) => {
            const y = padding.top + idx * rowHeight + 10;
            const x1 = Math.max(padding.left, scaleX(iv.lower_bound));
            const x2 = Math.min(padding.left + plotWidth, scaleX(iv.upper_bound));
            const xMean = scaleX(iv.sample_mean);
            const color = iv.captures_truth ? '#10b981' : '#ef4444';

            return (
              <g key={iv.index}>
                {/* Horizontal segment */}
                <line
                  x1={x1}
                  y1={y}
                  x2={x2}
                  y2={y}
                  stroke={color}
                  strokeWidth="2"
                  opacity={iv.captures_truth ? 0.85 : 1.0}
                />
                {/* Whiskers */}
                <line x1={x1} y1={y - 3} x2={x1} y2={y + 3} stroke={color} strokeWidth="2" />
                <line x2={x2} y1={y - 3} x1={x2} y2={y + 3} stroke={color} strokeWidth="2" />
                {/* Sample mean center point */}
                <circle cx={xMean} cy={y} r="2.5" fill={color} />
                {/* Index label */}
                <text x={padding.left - 8} y={y + 3} textAnchor="end" fill="#64748b" fontSize="9">
                  #{iv.index}
                </text>
              </g>
            );
          })}

          {/* X axis */}
          <line
            x1={padding.left}
            y1={svgHeight - padding.bottom}
            x2={padding.left + plotWidth}
            y2={svgHeight - padding.bottom}
            stroke="var(--color-border, #475569)"
            strokeWidth="1.5"
          />
          {[40, 45, 50, 55, 60].map((tick) => (
            <g key={tick}>
              <line
                x1={scaleX(tick)}
                y1={svgHeight - padding.bottom}
                x2={scaleX(tick)}
                y2={svgHeight - padding.bottom + 5}
                stroke="#64748b"
                strokeWidth="1"
              />
              <text
                x={scaleX(tick)}
                y={svgHeight - padding.bottom + 18}
                textAnchor="middle"
                fill="#94a3b8"
                fontSize="11"
              >
                {tick}
              </text>
            </g>
          ))}
        </svg>

        {/* Legend */}
        <div style={{ display: 'flex', gap: '1.5rem', justifyContent: 'center', marginTop: '0.75rem', fontSize: '0.8rem', color: '#cbd5e1', flexWrap: 'wrap' }}>
          <span style={{ display: 'flex', alignItems: 'center', gap: '0.35rem' }}>
            <span style={{ width: 14, height: 3, background: '#10b981', display: 'inline-block' }}></span>
            Captures True $\mu$ (Coverage success)
          </span>
          <span style={{ display: 'flex', alignItems: 'center', gap: '0.35rem' }}>
            <span style={{ width: 14, height: 3, background: '#ef4444', display: 'inline-block' }}></span>
            Misses True $\mu$ (Type I rate $\alpha = {(1 - confLevel).toFixed(2)}$)
          </span>
        </div>
      </div>

      {/* Coverage scorecard */}
      {data && (
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '0.75rem' }}>
          <div style={{ background: 'var(--color-surface, #1e293b)', padding: '0.75rem', borderRadius: '0.375rem', border: '1px solid var(--color-border, #334155)' }}>
            <div style={{ fontSize: '0.75rem', color: '#94a3b8' }}>Empirical Coverage</div>
            <div id="ci-coverage-rate" style={{ fontSize: '1.25rem', fontWeight: 700, color: data.empirical_coverage >= confLevel - 0.05 ? '#10b981' : '#ef4444' }}>
              {(data.empirical_coverage * 100).toFixed(1)}%
            </div>
            <div style={{ fontSize: '0.75rem', color: '#64748b', marginTop: '0.2rem' }}>
              {data.captured_count} of {data.num_intervals} captured (Nominal {(confLevel * 100).toFixed(0)}%)
            </div>
          </div>

          <div style={{ background: 'var(--color-surface, #1e293b)', padding: '0.75rem', borderRadius: '0.375rem', border: '1px solid var(--color-border, #334155)' }}>
            <div style={{ fontSize: '0.75rem', color: '#94a3b8' }}>Excel Margin of Error</div>
            <code id="ci-excel-formula" style={{ fontSize: '0.8rem', color: '#fbbf24', display: 'block', marginTop: '0.25rem' }}>
              {data.excel_formula}
            </code>
            <div style={{ fontSize: '0.75rem', color: '#64748b', marginTop: '0.2rem' }}>
              Or `=T.INV.2T(alpha, df) * (s / SQRT(n))`
            </div>
          </div>
        </div>
      )}

      {/* Interpretive note */}
      <div style={{ background: 'rgba(16, 185, 129, 0.08)', borderLeft: '4px solid #10b981', padding: '0.75rem 1rem', borderRadius: '0.25rem' }}>
        <p style={{ margin: 0, fontSize: '0.85rem', color: '#cbd5e1' }}>
          Frequentist Invariant: The parameter $\mu$ is fixed and stationary. The confidence level ${(confLevel * 100).toFixed(0)}\%$ refers to the long-run success rate of the <em>method</em> under repeated independent sampling.
        </p>
      </div>

      {/* ARIA summary */}
      <div role="status" aria-live="polite" className="sr-only" style={{ position: 'absolute', width: 1, height: 1, padding: 0, margin: -1, overflow: 'hidden', clip: 'rect(0, 0, 0, 0)', border: 0 }}>
        {data?.summary}
      </div>
    </div>
  );
};

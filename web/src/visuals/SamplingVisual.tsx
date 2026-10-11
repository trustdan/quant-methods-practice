import React, { useState, useEffect, useId } from 'react';
import { SamplingVisualResponse } from '../types/visuals';

export const SamplingVisual: React.FC = () => {
  const [popDist, setPopDist] = useState<string>('skewed');
  const [sampleSize, setSampleSize] = useState<number>(25);
  const [trials, setTrials] = useState<number>(400);
  const [data, setData] = useState<SamplingVisualResponse | null>(null);

  const patternId = useId();

  useEffect(() => {
    let active = true;
    const fetchSampling = async () => {
      try {
        const params = new URLSearchParams({
          pop_dist: popDist,
          n: sampleSize.toString(),
          trials: trials.toString(),
          seed: '12345',
        });
        const res = await fetch(`/api/visuals/sampling?${params.toString()}`);
        if (res.ok && active) {
          const json = await res.json();
          setData(json);
        }
      } catch {
        // Fallback placeholder
      }
    };
    fetchSampling();
    return () => {
      active = false;
    };
  }, [popDist, sampleSize, trials]);

  const svgWidth = 640;
  const svgHeight = 280;
  const padding = { top: 25, right: 30, bottom: 45, left: 45 };
  const plotWidth = svgWidth - padding.left - padding.right;
  const plotHeight = svgHeight - padding.top - padding.bottom;

  let minX = 0;
  let maxX = 10;
  let maxDensity = 1.0;

  if (data && data.bars.length > 0) {
    minX = data.bars[0].bin_mid - 0.5;
    maxX = data.bars[data.bars.length - 1].bin_mid + 0.5;
    const maxBarDens = Math.max(...data.bars.map((b) => b.density));
    const maxCurveDens = data.normal_curve.length > 0 ? Math.max(...data.normal_curve.map((p) => p.pdf)) : 0;
    maxDensity = Math.max(maxBarDens, maxCurveDens) * 1.15;
    if (maxDensity <= 0) maxDensity = 1.0;
  }

  const scaleX = (x: number) => padding.left + ((x - minX) / (maxX - minX || 1)) * plotWidth;
  const scaleY = (d: number) => padding.top + plotHeight - (d / maxDensity) * plotHeight;

  let curvePath = '';
  if (data && data.normal_curve.length > 1) {
    curvePath = data.normal_curve.reduce((path, pt, i) => {
      const px = scaleX(pt.x);
      const py = scaleY(pt.pdf);
      return i === 0 ? `M ${px} ${py}` : `${path} L ${px} ${py}`;
    }, '');
  }

  return (
    <div className="visual-container" style={{ display: 'flex', flexDirection: 'column', gap: '1rem', width: '100%', maxWidth: '100%', boxSizing: 'border-box' }}>
      <div className="visual-header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '0.5rem' }}>
        <div>
          <h3 style={{ margin: 0, fontSize: '1.25rem', color: 'var(--color-text-primary, #f1f5f9)' }}>
            Central Limit Theorem & Sampling Distribution Lab
          </h3>
          <p style={{ margin: '0.25rem 0 0', fontSize: '0.875rem', color: 'var(--color-text-secondary, #94a3b8)' }}>
            Observe how the distribution of sample means $\bar&#123;X&#125;$ converges to Normal as sample size $n$ increases.
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
            Population Distribution Shape
          </label>
          <select
            aria-label="Population distribution selector"
            value={popDist}
            onChange={(e) => setPopDist(e.target.value)}
            style={{ width: '100%', padding: '0.45rem', borderRadius: '0.25rem', background: 'var(--color-bg, #0f172a)', color: '#fff', border: '1px solid var(--color-border, #334155)' }}
          >
            <option value="skewed">Skewed (Chi-Square df=3)</option>
            <option value="uniform">Uniform U(0, 20)</option>
            <option value="exponential">Exponential (decaying)</option>
            <option value="normal">Normal (bell curve)</option>
          </select>
        </div>

        <div>
          <label style={{ display: 'block', fontSize: '0.75rem', fontWeight: 600, color: 'var(--color-text-secondary, #94a3b8)', marginBottom: '0.35rem' }}>
            Sample Size: $n = {sampleSize}$ {sampleSize >= 30 ? ' (CLT rule $n \\ge 30$)' : ' (Small sample)'}
          </label>
          <input
            aria-label="Sample size slider"
            type="range"
            min="2"
            max="60"
            step="1"
            value={sampleSize}
            onChange={(e) => setSampleSize(parseInt(e.target.value, 10))}
            style={{ width: '100%' }}
          />
        </div>

        <div>
          <label style={{ display: 'block', fontSize: '0.75rem', fontWeight: 600, color: 'var(--color-text-secondary, #94a3b8)', marginBottom: '0.35rem' }}>
            Simulation Repetitions: {trials} samples
          </label>
          <select
            aria-label="Trials selector"
            value={trials}
            onChange={(e) => setTrials(parseInt(e.target.value, 10))}
            style={{ width: '100%', padding: '0.45rem', borderRadius: '0.25rem', background: 'var(--color-bg, #0f172a)', color: '#fff', border: '1px solid var(--color-border, #334155)' }}
          >
            <option value="150">150 sample means</option>
            <option value="400">400 sample means</option>
            <option value="800">800 sample means</option>
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
          aria-label={data?.summary || 'Sampling distribution histogram'}
        >
          <defs>
            <pattern id={patternId} width="8" height="8" patternUnits="userSpaceOnUse" patternTransform="rotate(45)">
              <line x1="0" y1="0" x2="0" y2="8" stroke="rgba(56, 189, 248, 0.6)" strokeWidth="2" />
            </pattern>
          </defs>

          {/* Grid lines */}
          <line
            x1={padding.left}
            y1={padding.top + plotHeight}
            x2={padding.left + plotWidth}
            y2={padding.top + plotHeight}
            stroke="var(--color-border, #475569)"
            strokeWidth="1.5"
          />
          <line
            x1={padding.left}
            y1={padding.top}
            x2={padding.left}
            y2={padding.top + plotHeight}
            stroke="var(--color-border, #475569)"
            strokeWidth="1.5"
          />

          {/* Histogram bars */}
          {data?.bars.map((b, idx) => {
            const barW = Math.max(2, plotWidth / (data.bars.length * 1.15));
            const bx = scaleX(b.bin_mid) - barW / 2;
            const by = scaleY(b.density);
            const bh = Math.max(0, padding.top + plotHeight - by);
            return (
              <g key={idx}>
                <rect
                  x={bx}
                  y={by}
                  width={barW}
                  height={bh}
                  fill={`url(#${patternId})`}
                  stroke="var(--color-primary, #38bdf8)"
                  strokeWidth="1"
                />
              </g>
            );
          })}

          {/* Theoretical normal overlay curve */}
          {curvePath && (
            <path
              d={curvePath}
              fill="none"
              stroke="#f59e0b"
              strokeWidth="2.5"
            />
          )}

          {/* Population Mean guide line */}
          {data && (
            <line
              x1={scaleX(data.pop_mean)}
              y1={padding.top}
              x2={scaleX(data.pop_mean)}
              y2={padding.top + plotHeight}
              stroke="#ef4444"
              strokeWidth="2"
              strokeDasharray="4 3"
            />
          )}

          {/* Axes labels */}
          <text
            x={padding.left + plotWidth / 2}
            y={svgHeight - 10}
            textAnchor="middle"
            fill="var(--color-text-secondary, #94a3b8)"
            fontSize="12"
          >
            Sample Mean ($\bar&#123;X&#125;$)
          </text>
          <text
            x={15}
            y={padding.top + plotHeight / 2}
            textAnchor="middle"
            transform={`rotate(-90, 15, ${padding.top + plotHeight / 2})`}
            fill="var(--color-text-secondary, #94a3b8)"
            fontSize="11"
          >
            Density
          </text>
        </svg>

        {/* Legend */}
        <div style={{ display: 'flex', gap: '1.25rem', justifyContent: 'center', marginTop: '0.5rem', fontSize: '0.8rem', color: '#cbd5e1', flexWrap: 'wrap' }}>
          <span style={{ display: 'flex', alignItems: 'center', gap: '0.35rem' }}>
            <span style={{ width: 12, height: 12, background: 'rgba(56, 189, 248, 0.4)', border: '1px solid #38bdf8', display: 'inline-block' }}></span>
            Simulated Sample Means ($\bar&#123;X&#125;$)
          </span>
          <span style={{ display: 'flex', alignItems: 'center', gap: '0.35rem' }}>
            <span style={{ width: 16, height: 3, background: '#f59e0b', display: 'inline-block' }}></span>
            Theoretical CLT Curve $N(\mu, \sigma / \sqrt&#123;n&#125;)$
          </span>
          <span style={{ display: 'flex', alignItems: 'center', gap: '0.35rem' }}>
            <span style={{ width: 16, height: 2, borderTop: '2px dashed #ef4444', display: 'inline-block' }}></span>
            Population Mean $\mu = {data?.pop_mean}$
          </span>
        </div>
      </div>

      {/* Metrics and CLT notes */}
      {data && (
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '0.75rem' }}>
          <div style={{ background: 'var(--color-surface, #1e293b)', padding: '0.75rem', borderRadius: '0.375rem', border: '1px solid var(--color-border, #334155)' }}>
            <div style={{ fontSize: '0.75rem', color: '#94a3b8' }}>Theoretical Standard Error ($SE$)</div>
            <div id="sampling-stats-theoretical-se" style={{ fontSize: '1.2rem', fontWeight: 700, color: '#38bdf8' }}>
              $\sigma / \sqrt&#123;n&#125; = {data.standard_error}$
            </div>
            <div style={{ fontSize: '0.75rem', color: '#64748b', marginTop: '0.2rem' }}>
              Pop $\sigma = {data.pop_std_dev}$, $n = {data.sample_size}$
            </div>
          </div>

          <div style={{ background: 'var(--color-surface, #1e293b)', padding: '0.75rem', borderRadius: '0.375rem', border: '1px solid var(--color-border, #334155)' }}>
            <div style={{ fontSize: '0.75rem', color: '#94a3b8' }}>Empirical Sample SD ($s_&#123;\bar&#123;x&#125;&#125;$)</div>
            <div id="sampling-stats-empirical-se" style={{ fontSize: '1.2rem', fontWeight: 700, color: '#10b981' }}>
              {data.empirical_std_err}
            </div>
            <div style={{ fontSize: '0.75rem', color: '#64748b', marginTop: '0.2rem' }}>
              Simulated across {data.trials} trials
            </div>
          </div>

          <div style={{ background: 'var(--color-surface, #1e293b)', padding: '0.75rem', borderRadius: '0.375rem', border: '1px solid var(--color-border, #334155)' }}>
            <div style={{ fontSize: '0.75rem', color: '#94a3b8' }}>Excel Formula</div>
            <code id="sampling-excel-formula" style={{ fontSize: '0.8rem', color: '#fbbf24', display: 'block', marginTop: '0.25rem' }}>
              {data.excel_formula}
            </code>
          </div>
        </div>
      )}

      {/* CLT Explanation Callout */}
      {data && (
        <div style={{ background: 'rgba(56, 189, 248, 0.08)', borderLeft: '4px solid #38bdf8', padding: '0.75rem 1rem', borderRadius: '0.25rem' }}>
          <p style={{ margin: 0, fontSize: '0.85rem', color: '#cbd5e1' }}>
            {data.clt_note}
          </p>
        </div>
      )}

      {/* ARIA summary */}
      <div role="status" aria-live="polite" className="sr-only" style={{ position: 'absolute', width: 1, height: 1, padding: 0, margin: -1, overflow: 'hidden', clip: 'rect(0, 0, 0, 0)', border: 0 }}>
        {data?.summary}
      </div>
    </div>
  );
};

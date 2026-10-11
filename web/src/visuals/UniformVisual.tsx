import React, { useState, useEffect, useId } from 'react';
import { UniformVisualResponse } from '../types/visuals';

export const UniformVisual: React.FC = () => {
  const [a, setA] = useState<number>(0);
  const [b, setB] = useState<number>(20);
  const [tail, setTail] = useState<'lower' | 'upper' | 'interval'>('interval');
  const [cutoffX, setCutoffX] = useState<number>(10);
  const [cutoffX1, setCutoffX1] = useState<number>(5);
  const [cutoffX2, setCutoffX2] = useState<number>(12);
  const [viewMode, setViewMode] = useState<'pdf' | 'cdf'>('pdf');
  const [data, setData] = useState<UniformVisualResponse | null>(null);

  const patternId = useId();

  useEffect(() => {
    let active = true;
    const fetchCurve = async () => {
      try {
        const params = new URLSearchParams({
          a: a.toString(),
          b: b.toString(),
          tail,
        });
        if (tail === 'interval') {
          params.set('x1', cutoffX1.toString());
          params.set('x2', cutoffX2.toString());
        } else {
          params.set('x', cutoffX.toString());
        }
        const res = await fetch(`/api/visuals/uniform?${params.toString()}`);
        if (res.ok && active) {
          const json = await res.json();
          setData(json);
        }
      } catch {
        // local fallback
      }
    };
    fetchCurve();
    return () => {
      active = false;
    };
  }, [a, b, tail, cutoffX, cutoffX1, cutoffX2]);

  const width = 640;
  const height = 280;
  const margin = { top: 25, right: 30, bottom: 45, left: 55 };
  const innerWidth = width - margin.left - margin.right;
  const innerHeight = height - margin.top - margin.bottom;

  const pad = (b - a) * 0.2;
  const xMin = a - pad;
  const xMax = b + pad;
  const heightDensity = 1.0 / (b - a);
  const yMax = viewMode === 'pdf' ? heightDensity * 1.3 : 1.1;

  const scaleX = (x: number) => margin.left + ((x - xMin) / (xMax - xMin)) * innerWidth;
  const scaleY = (y: number) => margin.top + innerHeight - (y / yMax) * innerHeight;

  // Determine shaded bounds
  let shadeLeft = a;
  let shadeRight = b;
  if (tail === 'lower') {
    shadeLeft = Math.max(a, xMin);
    shadeRight = Math.min(b, Math.max(a, cutoffX));
  } else if (tail === 'upper') {
    shadeLeft = Math.max(a, Math.min(b, cutoffX));
    shadeRight = b;
  } else {
    shadeLeft = Math.max(a, Math.min(b, cutoffX1));
    shadeRight = Math.max(a, Math.min(b, cutoffX2));
  }

  return (
    <div className="visual-card" style={{ background: 'var(--bg-surface)', padding: '1.25rem', borderRadius: 'var(--radius-md)', border: '1px solid var(--border-subtle)' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem', flexWrap: 'wrap', gap: '0.75rem' }}>
        <div>
          <h3 style={{ margin: 0, fontSize: '1.2rem', color: 'var(--text-main)', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            Uniform Distribution Explorer
            <span className="badge badge-cyan" style={{ fontSize: '0.75rem' }}>Continuous U(a, b)</span>
          </h3>
          <p style={{ margin: '0.25rem 0 0', fontSize: '0.85rem', color: 'var(--text-muted)' }}>
            Constant density height $f(x) = \frac{1}{b - a}$ across support interval $[a, b]$. Shaded rectangular area equals probability.
          </p>
        </div>

        <div style={{ display: 'flex', gap: '0.5rem' }}>
          <button
            type="button"
            className={`btn btn-sm ${viewMode === 'pdf' ? 'btn-primary' : 'btn-outline'}`}
            onClick={() => setViewMode('pdf')}
            id="btn-toggle-uniform-pdf"
            aria-pressed={viewMode === 'pdf'}
          >
            PDF (Flat Density)
          </button>
          <button
            type="button"
            className={`btn btn-sm ${viewMode === 'cdf' ? 'btn-primary' : 'btn-outline'}`}
            onClick={() => setViewMode('cdf')}
            id="btn-toggle-uniform-cdf"
            aria-pressed={viewMode === 'cdf'}
          >
            CDF (Linear Ramp)
          </button>
        </div>
      </div>

      {/* Parameter Controls */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))', gap: '1rem', marginBottom: '1.25rem', background: 'var(--bg-surface-elevated)', padding: '1rem', borderRadius: 'var(--radius-sm)' }}>
        <div>
          <label htmlFor="uni-a-input" style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-main)', marginBottom: '0.25rem' }}>
            Lower Bound a: {a}
          </label>
          <input
            id="uni-a-input"
            type="range"
            min="-10"
            max={b - 1}
            step="1"
            value={a}
            onChange={(e) => setA(parseFloat(e.target.value))}
            style={{ width: '100%' }}
          />
        </div>

        <div>
          <label htmlFor="uni-b-input" style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-main)', marginBottom: '0.25rem' }}>
            Upper Bound b: {b}
          </label>
          <input
            id="uni-b-input"
            type="range"
            min={a + 1}
            max="50"
            step="1"
            value={b}
            onChange={(e) => setB(parseFloat(e.target.value))}
            style={{ width: '100%' }}
          />
        </div>

        <div>
          <label htmlFor="uni-tail-select" style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-main)', marginBottom: '0.25rem' }}>
            Event:
          </label>
          <select
            id="uni-tail-select"
            value={tail}
            onChange={(e) => setTail(e.target.value as 'interval' | 'lower' | 'upper')}
            className="input"
            style={{ width: '100%', padding: '0.35rem 0.6rem' }}
          >
            <option value="interval">Interval: P(x₁ ≤ X ≤ x₂)</option>
            <option value="lower">Lower: P(X ≤ x)</option>
            <option value="upper">Upper: P(X ≥ x)</option>
          </select>
        </div>

        {tail === 'interval' ? (
          <>
            <div>
              <label htmlFor="uni-x1-input" style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-main)', marginBottom: '0.25rem' }}>
                Cutoff x₁: {cutoffX1}
              </label>
              <input
                id="uni-x1-input"
                type="range"
                min={a}
                max={cutoffX2}
                step="0.5"
                value={cutoffX1}
                onChange={(e) => setCutoffX1(parseFloat(e.target.value))}
                style={{ width: '100%' }}
              />
            </div>
            <div>
              <label htmlFor="uni-x2-input" style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-main)', marginBottom: '0.25rem' }}>
                Cutoff x₂: {cutoffX2}
              </label>
              <input
                id="uni-x2-input"
                type="range"
                min={cutoffX1}
                max={b}
                step="0.5"
                value={cutoffX2}
                onChange={(e) => setCutoffX2(parseFloat(e.target.value))}
                style={{ width: '100%' }}
              />
            </div>
          </>
        ) : (
          <div>
            <label htmlFor="uni-cutoff-input" style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-main)', marginBottom: '0.25rem' }}>
              Cutoff x: {cutoffX}
            </label>
            <input
              id="uni-cutoff-input"
              type="range"
              min={a}
              max={b}
              step="0.5"
              value={cutoffX}
              onChange={(e) => setCutoffX(parseFloat(e.target.value))}
              style={{ width: '100%' }}
            />
          </div>
        )}
      </div>

      {/* SVG Canvas */}
      <div style={{ position: 'relative', width: '100%', overflowX: 'auto', background: 'var(--bg-canvas, #0d1117)', borderRadius: 'var(--radius-sm)', border: '1px solid var(--border-subtle)', marginBottom: '1rem' }}>
        <svg
          viewBox={`0 0 ${width} ${height}`}
          style={{ width: '100%', height: 'auto', display: 'block' }}
          role="img"
          aria-label={data?.summary || 'Uniform distribution chart'}
        >
          <defs>
            <pattern id={patternId} width="8" height="8" patternTransform="rotate(45 0 0)" patternUnits="userSpaceOnUse">
              <line x1="0" y1="0" x2="0" y2="8" stroke="var(--accent-cyan)" strokeWidth="2.5" strokeOpacity="0.4" />
            </pattern>
          </defs>

          {/* Baseline */}
          <line
            x1={margin.left}
            y1={scaleY(0)}
            x2={margin.left + innerWidth}
            y2={scaleY(0)}
            stroke="var(--border-subtle)"
            strokeWidth="1.5"
          />

          {viewMode === 'pdf' ? (
            <>
              {/* Shaded Rectangle */}
              {shadeRight > shadeLeft && (
                <>
                  <rect
                    x={scaleX(shadeLeft)}
                    y={scaleY(heightDensity)}
                    width={scaleX(shadeRight) - scaleX(shadeLeft)}
                    height={scaleY(0) - scaleY(heightDensity)}
                    fill="var(--accent-cyan)"
                    fillOpacity="0.25"
                  />
                  <rect
                    x={scaleX(shadeLeft)}
                    y={scaleY(heightDensity)}
                    width={scaleX(shadeRight) - scaleX(shadeLeft)}
                    height={scaleY(0) - scaleY(heightDensity)}
                    fill={`url(#${patternId})`}
                  />
                </>
              )}

              {/* PDF Outline: 0 outside, step up to 1/(b-a), step down to 0 */}
              <polyline
                points={`
                  ${scaleX(xMin)},${scaleY(0)}
                  ${scaleX(a)},${scaleY(0)}
                  ${scaleX(a)},${scaleY(heightDensity)}
                  ${scaleX(b)},${scaleY(heightDensity)}
                  ${scaleX(b)},${scaleY(0)}
                  ${scaleX(xMax)},${scaleY(0)}
                `}
                fill="none"
                stroke="var(--accent-cyan)"
                strokeWidth="2.5"
              />
            </>
          ) : (
            // CDF Linear Ramp: 0 before a, ramp to 1 at b, flat 1 after b
            <polyline
              points={`
                ${scaleX(xMin)},${scaleY(0)}
                ${scaleX(a)},${scaleY(0)}
                ${scaleX(b)},${scaleY(1)}
                ${scaleX(xMax)},${scaleY(1)}
              `}
              fill="none"
              stroke="var(--accent-emerald)"
              strokeWidth="2.5"
            />
          )}

          {/* Bound ticks */}
          <line x1={scaleX(a)} y1={scaleY(0)} x2={scaleX(a)} y2={scaleY(0) + 6} stroke="var(--text-muted)" />
          <text x={scaleX(a)} y={scaleY(0) + 18} textAnchor="middle" fill="var(--text-muted)" fontSize="11">
            a = {a}
          </text>
          <line x1={scaleX(b)} y1={scaleY(0)} x2={scaleX(b)} y2={scaleY(0) + 6} stroke="var(--text-muted)" />
          <text x={scaleX(b)} y={scaleY(0) + 18} textAnchor="middle" fill="var(--text-muted)" fontSize="11">
            b = {b}
          </text>
        </svg>
      </div>

      {/* KPI Cards */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))', gap: '0.75rem', marginBottom: '1rem' }}>
        <div style={{ background: 'var(--bg-surface-elevated)', padding: '0.75rem', borderRadius: 'var(--radius-sm)', borderLeft: '3px solid var(--accent-cyan)' }}>
          <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', textTransform: 'uppercase' }}>Probability (Area)</div>
          <div style={{ fontSize: '1.4rem', fontWeight: 700, color: 'var(--accent-cyan)' }} id="uniform-shaded-prob-value">
            {data ? (data.shaded_probability * 100).toFixed(2) + '%' : '...'}
          </div>
          <div style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>
            P = {data ? data.shaded_probability.toFixed(4) : '...'}
          </div>
        </div>

        <div style={{ background: 'var(--bg-surface-elevated)', padding: '0.75rem', borderRadius: 'var(--radius-sm)', borderLeft: '3px solid var(--accent-amber, #e3b341)' }}>
          <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', textTransform: 'uppercase' }}>Constant Density f(x)</div>
          <div style={{ fontSize: '1.4rem', fontWeight: 700, color: 'var(--text-main)' }} id="uniform-density-value">
            {heightDensity.toFixed(4)}
          </div>
          <div style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>
            Height = 1 / ({b} - {a})
          </div>
        </div>

        <div style={{ background: 'var(--bg-surface-elevated)', padding: '0.75rem', borderRadius: 'var(--radius-sm)', borderLeft: '3px solid var(--accent-emerald)' }}>
          <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', textTransform: 'uppercase' }}>Mean E[X]</div>
          <div style={{ fontSize: '1.4rem', fontWeight: 700, color: 'var(--accent-emerald)' }}>
            {((a + b) / 2).toFixed(2)}
          </div>
          <div style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>
            Var = {(((b - a) ** 2) / 12).toFixed(2)}
          </div>
        </div>
      </div>

      <div
        role="region"
        aria-live="polite"
        style={{ padding: '0.75rem 1rem', background: 'var(--bg-surface-elevated)', borderRadius: 'var(--radius-sm)', fontSize: '0.85rem', color: 'var(--text-main)' }}
        id="uniform-text-summary"
      >
        <strong>Summary: </strong>
        {data?.summary || 'Adjust bounds to evaluate continuous uniform intervals.'}
      </div>
    </div>
  );
};

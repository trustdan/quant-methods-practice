import React, { useState, useEffect, useId } from 'react';
import { ExponentialVisualResponse } from '../types/visuals';

export const ExponentialVisual: React.FC = () => {
  const [lambda, setLambda] = useState<number>(1.0);
  const [tail, setTail] = useState<'lower' | 'upper' | 'interval'>('upper');
  const [cutoffX, setCutoffX] = useState<number>(1.0);
  const [cutoffX1, setCutoffX1] = useState<number>(0.5);
  const [cutoffX2, setCutoffX2] = useState<number>(2.0);
  const [viewMode, setViewMode] = useState<'pdf' | 'cdf'>('pdf');
  const [data, setData] = useState<ExponentialVisualResponse | null>(null);

  // Memoryless demonstrator inputs
  const [sVal, setSVal] = useState<number>(2.0);
  const [tVal, setTVal] = useState<number>(1.5);

  const patternId = useId();

  useEffect(() => {
    let active = true;
    const fetchCurve = async () => {
      try {
        const params = new URLSearchParams({
          lambda: lambda.toString(),
          tail,
        });
        if (tail === 'interval') {
          params.set('x1', cutoffX1.toString());
          params.set('x2', cutoffX2.toString());
        } else {
          params.set('x', cutoffX.toString());
        }
        const res = await fetch(`/api/visuals/exponential?${params.toString()}`);
        if (res.ok && active) {
          const json = await res.json();
          setData(json);
        }
      } catch {
        // fallback
      }
    };
    fetchCurve();
    return () => {
      active = false;
    };
  }, [lambda, tail, cutoffX, cutoffX1, cutoffX2]);

  const width = 640;
  const height = 280;
  const margin = { top: 25, right: 30, bottom: 45, left: 55 };
  const innerWidth = width - margin.left - margin.right;
  const innerHeight = height - margin.top - margin.bottom;

  const points = data?.points || [];
  const xMax = 4.5 / lambda;
  const yMax = viewMode === 'pdf' ? lambda * 1.15 : 1.1;

  const scaleX = (x: number) => margin.left + (x / xMax) * innerWidth;
  const scaleY = (y: number) => margin.top + innerHeight - (y / yMax) * innerHeight;

  const curvePath = points.length > 0
    ? points.reduce((path, pt, i) => {
        const px = scaleX(pt.x);
        const py = scaleY(viewMode === 'pdf' ? pt.pdf : pt.cdf);
        return i === 0 ? `M ${px} ${py}` : `${path} L ${px} ${py}`;
      }, '')
    : '';

  let shadedPath = '';
  if (viewMode === 'pdf' && points.length > 0) {
    const isShaded = (x: number) => {
      if (tail === 'lower') return x <= cutoffX;
      if (tail === 'upper') return x >= cutoffX;
      return x >= cutoffX1 && x <= cutoffX2;
    };
    const shadedPoints = points.filter((pt) => isShaded(pt.x));
    if (shadedPoints.length > 0) {
      const first = shadedPoints[0];
      const last = shadedPoints[shadedPoints.length - 1];
      let p = `M ${scaleX(first.x)} ${scaleY(0)}`;
      shadedPoints.forEach((pt) => {
        p += ` L ${scaleX(pt.x)} ${scaleY(pt.pdf)}`;
      });
      p += ` L ${scaleX(last.x)} ${scaleY(0)} Z`;
      shadedPath = p;
    }
  }

  // Memoryless calculations
  const condProb = Math.exp(-lambda * (sVal + tVal)) / Math.exp(-lambda * sVal);
  const directProb = Math.exp(-lambda * tVal);

  return (
    <div className="visual-card" style={{ background: 'var(--bg-surface)', padding: '1.25rem', borderRadius: 'var(--radius-md)', border: '1px solid var(--border-subtle)' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem', flexWrap: 'wrap', gap: '0.75rem' }}>
        <div>
          <h3 style={{ margin: 0, fontSize: '1.2rem', color: 'var(--text-main)', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            Exponential Distribution Explorer
            <span className="badge badge-cyan" style={{ fontSize: '0.75rem' }}>Continuous Exp(λ)</span>
          </h3>
          <p style={{ margin: '0.25rem 0 0', fontSize: '0.85rem', color: 'var(--text-muted)' }}>
            Models waiting times and component service lifetimes with constant rate $\lambda$ and memoryless property.
          </p>
        </div>

        <div style={{ display: 'flex', gap: '0.5rem' }}>
          <button
            type="button"
            className={`btn btn-sm ${viewMode === 'pdf' ? 'btn-primary' : 'btn-outline'}`}
            onClick={() => setViewMode('pdf')}
            id="btn-toggle-exp-pdf"
            aria-pressed={viewMode === 'pdf'}
          >
            PDF (Decaying Density)
          </button>
          <button
            type="button"
            className={`btn btn-sm ${viewMode === 'cdf' ? 'btn-primary' : 'btn-outline'}`}
            onClick={() => setViewMode('cdf')}
            id="btn-toggle-exp-cdf"
            aria-pressed={viewMode === 'cdf'}
          >
            CDF (Cumulative 1 - e^(-λx))
          </button>
        </div>
      </div>

      {/* Parameter Controls */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))', gap: '1rem', marginBottom: '1.25rem', background: 'var(--bg-surface-elevated)', padding: '1rem', borderRadius: 'var(--radius-sm)' }}>
        <div>
          <label htmlFor="exp-lambda-input" style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-main)', marginBottom: '0.25rem' }}>
            Rate (λ): {lambda} (Mean = {(1 / lambda).toFixed(2)})
          </label>
          <input
            id="exp-lambda-input"
            type="range"
            min="0.1"
            max="4.0"
            step="0.1"
            value={lambda}
            onChange={(e) => setLambda(parseFloat(e.target.value))}
            style={{ width: '100%' }}
          />
        </div>

        <div>
          <label htmlFor="exp-tail-select" style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-main)', marginBottom: '0.25rem' }}>
            Event Region:
          </label>
          <select
            id="exp-tail-select"
            value={tail}
            onChange={(e) => setTail(e.target.value as 'upper' | 'lower' | 'interval')}
            className="input"
            style={{ width: '100%', padding: '0.35rem 0.6rem' }}
          >
            <option value="upper">Upper Survival: P(X ≥ x) = e^(-λx)</option>
            <option value="lower">Lower: P(X ≤ x) = 1 - e^(-λx)</option>
            <option value="interval">Interval: P(x₁ ≤ X ≤ x₂)</option>
          </select>
        </div>

        {tail === 'interval' ? (
          <>
            <div>
              <label htmlFor="exp-x1-input" style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-main)', marginBottom: '0.25rem' }}>
                Bound x₁: {cutoffX1.toFixed(2)}
              </label>
              <input
                id="exp-x1-input"
                type="range"
                min="0"
                max={cutoffX2}
                step="0.1"
                value={cutoffX1}
                onChange={(e) => setCutoffX1(parseFloat(e.target.value))}
                style={{ width: '100%' }}
              />
            </div>
            <div>
              <label htmlFor="exp-x2-input" style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-main)', marginBottom: '0.25rem' }}>
                Bound x₂: {cutoffX2.toFixed(2)}
              </label>
              <input
                id="exp-x2-input"
                type="range"
                min={cutoffX1}
                max={xMax}
                step="0.1"
                value={cutoffX2}
                onChange={(e) => setCutoffX2(parseFloat(e.target.value))}
                style={{ width: '100%' }}
              />
            </div>
          </>
        ) : (
          <div>
            <label htmlFor="exp-cutoff-input" style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-main)', marginBottom: '0.25rem' }}>
              Cutoff x: {cutoffX.toFixed(2)}
            </label>
            <input
              id="exp-cutoff-input"
              type="range"
              min="0"
              max={xMax}
              step="0.1"
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
          aria-label={data?.summary || 'Exponential distribution chart'}
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

          {/* Shaded Area for PDF */}
          {viewMode === 'pdf' && shadedPath && (
            <>
              <path d={shadedPath} fill="var(--accent-cyan)" fillOpacity="0.25" />
              <path d={shadedPath} fill={`url(#${patternId})`} />
            </>
          )}

          {/* Curve */}
          {curvePath && (
            <path
              d={curvePath}
              fill="none"
              stroke="var(--accent-cyan)"
              strokeWidth="2.5"
            />
          )}

          {/* Cutoff indicator */}
          {viewMode === 'pdf' && tail !== 'interval' && (
            <line
              x1={scaleX(cutoffX)}
              y1={margin.top}
              x2={scaleX(cutoffX)}
              y2={scaleY(0)}
              stroke="var(--accent-amber, #e3b341)"
              strokeWidth="2"
              strokeDasharray="4 2"
            />
          )}
        </svg>
      </div>

      {/* KPI Cards */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))', gap: '0.75rem', marginBottom: '1rem' }}>
        <div style={{ background: 'var(--bg-surface-elevated)', padding: '0.75rem', borderRadius: 'var(--radius-sm)', borderLeft: '3px solid var(--accent-cyan)' }}>
          <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', textTransform: 'uppercase' }}>Probability (Area)</div>
          <div style={{ fontSize: '1.4rem', fontWeight: 700, color: 'var(--accent-cyan)' }} id="exp-shaded-prob-value">
            {data ? (data.shaded_probability * 100).toFixed(2) + '%' : '...'}
          </div>
          <div style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>
            P = {data ? data.shaded_probability.toFixed(4) : '...'}
          </div>
        </div>

        <div style={{ background: 'var(--bg-surface-elevated)', padding: '0.75rem', borderRadius: 'var(--radius-sm)', borderLeft: '3px solid var(--accent-amber, #e3b341)' }}>
          <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', textTransform: 'uppercase' }}>Initial Density f(0)</div>
          <div style={{ fontSize: '1.4rem', fontWeight: 700, color: 'var(--text-main)' }}>
            {lambda.toFixed(2)}
          </div>
          <div style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>
            Peak height at x = 0 is λ
          </div>
        </div>

        <div style={{ background: 'var(--bg-surface-elevated)', padding: '0.75rem', borderRadius: 'var(--radius-sm)', borderLeft: '3px solid var(--accent-emerald)' }}>
          <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', textTransform: 'uppercase' }}>Excel Formula</div>
          <div style={{ fontSize: '0.95rem', fontWeight: 600, color: 'var(--text-main)', wordBreak: 'break-all', fontFamily: 'monospace' }} id="exp-excel-formula">
            {data?.excel_formula || '=EXPON.DIST(...)'}
          </div>
          <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
            Standard course syntax
          </div>
        </div>
      </div>

      {/* Memoryless Property Demonstrator */}
      <div style={{ background: 'var(--bg-surface-elevated)', padding: '0.9rem', borderRadius: 'var(--radius-sm)', border: '1px solid var(--border-subtle)', marginBottom: '1rem' }}>
        <div style={{ fontWeight: 600, fontSize: '0.9rem', color: 'var(--accent-cyan)', marginBottom: '0.4rem' }}>
          Interactive Memoryless Property Check: P(X &gt; s + t | X &gt; s) = P(X &gt; t)
        </div>
        <p style={{ fontSize: '0.82rem', color: 'var(--text-muted)', margin: '0 0 0.6rem 0' }}>
          Given that a component has already survived $s$ hours, the probability it survives $t$ additional hours is identical to fresh survival $P(X &gt; t)$.
        </p>
        <div style={{ display: 'flex', gap: '1rem', flexWrap: 'wrap', alignItems: 'center' }}>
          <div>
            <label htmlFor="memoryless-s-input" style={{ fontSize: '0.8rem', color: 'var(--text-muted)', marginRight: '0.4rem' }}>Elapsed s:</label>
            <input
              id="memoryless-s-input"
              type="number"
              min="0.5"
              max="10"
              step="0.5"
              value={sVal}
              onChange={(e) => setSVal(parseFloat(e.target.value) || 1)}
              style={{ width: '60px', padding: '0.2rem' }}
            />
          </div>
          <div>
            <label htmlFor="memoryless-t-input" style={{ fontSize: '0.8rem', color: 'var(--text-muted)', marginRight: '0.4rem' }}>Additional t:</label>
            <input
              id="memoryless-t-input"
              type="number"
              min="0.5"
              max="10"
              step="0.5"
              value={tVal}
              onChange={(e) => setTVal(parseFloat(e.target.value) || 1)}
              style={{ width: '60px', padding: '0.2rem' }}
            />
          </div>
          <div style={{ fontSize: '0.85rem', color: 'var(--text-main)', fontWeight: 600 }}>
            Conditional: P(X &gt; {sVal + tVal} | X &gt; {sVal}) = <span style={{ color: 'var(--accent-emerald)' }}>{condProb.toFixed(4)}</span> | Fresh P(X &gt; {tVal}) = <span style={{ color: 'var(--accent-emerald)' }}>{directProb.toFixed(4)}</span> (Exact Match!)
          </div>
        </div>
      </div>

      <div
        role="region"
        aria-live="polite"
        style={{ padding: '0.75rem 1rem', background: 'var(--bg-surface-elevated)', borderRadius: 'var(--radius-sm)', fontSize: '0.85rem', color: 'var(--text-main)' }}
        id="exp-text-summary"
      >
        <strong>Summary: </strong>
        {data?.summary || 'Adjust parameters to inspect exponential survival probabilities.'}
      </div>
    </div>
  );
};

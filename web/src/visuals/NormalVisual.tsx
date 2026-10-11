import React, { useState, useEffect, useId } from 'react';
import { NormalVisualResponse } from '../types/visuals';

interface NormalVisualProps {
  initialMu?: number;
  initialSigma?: number;
}

export const NormalVisual: React.FC<NormalVisualProps> = ({
  initialMu = 0,
  initialSigma = 1,
}) => {
  const [mu, setMu] = useState<number>(initialMu);
  const [sigma, setSigma] = useState<number>(initialSigma);
  const [tail, setTail] = useState<'lower' | 'upper' | 'interval' | 'two_tailed'>('lower');
  const [cutoffX, setCutoffX] = useState<number>(initialMu);
  const [cutoffX1, setCutoffX1] = useState<number>(initialMu - initialSigma);
  const [cutoffX2, setCutoffX2] = useState<number>(initialMu + initialSigma);
  const [viewMode, setViewMode] = useState<'pdf' | 'cdf'>('pdf');
  const [data, setData] = useState<NormalVisualResponse | null>(null);

  const patternId = useId();

  // Fetch or calculate normal curve data
  useEffect(() => {
    let active = true;
    const fetchCurve = async () => {
      try {
        const params = new URLSearchParams({
          mu: mu.toString(),
          sigma: sigma.toString(),
          tail,
          samples: '100',
        });
        if (tail === 'interval') {
          params.set('x1', cutoffX1.toString());
          params.set('x2', cutoffX2.toString());
        } else if (tail === 'two_tailed') {
          params.set('x', Math.abs(cutoffX - mu).toString());
        } else {
          params.set('x', cutoffX.toString());
        }

        const res = await fetch(`/api/visuals/normal?${params.toString()}`);
        if (res.ok && active) {
          const json = await res.json();
          setData(json);
        }
      } catch {
        // Local mathematical fallback if server unavailable
        if (active) {
          const z = (cutoffX - mu) / sigma;
          const prob = 0.5 * (1 + Math.sign(z) * Math.sqrt(1 - Math.exp(-2 * z * z / Math.PI)));
          setData({
            family: 'normal',
            mu,
            sigma,
            tail,
            cutoff_x: cutoffX,
            z_score: z,
            shaded_probability: prob,
            excel_formula: `=NORM.DIST(${cutoffX}, ${mu}, ${sigma}, TRUE)`,
            summary: `Normal N(${mu}, ${sigma}): P(X <= ${cutoffX}) ≈ ${prob.toFixed(4)}`,
            empirical_rule: { one_sigma: 0.6827, two_sigma: 0.9545, three_sigma: 0.9973 },
            points: [],
          });
        }
      }
    };

    fetchCurve();
    return () => {
      active = false;
    };
  }, [mu, sigma, tail, cutoffX, cutoffX1, cutoffX2]);

  // SVG Chart Dimensions
  const width = 640;
  const height = 300;
  const margin = { top: 25, right: 30, bottom: 45, left: 55 };
  const innerWidth = width - margin.left - margin.right;
  const innerHeight = height - margin.top - margin.bottom;

  const points = data?.points || [];
  const xMin = mu - 3.8 * sigma;
  const xMax = mu + 3.8 * sigma;
  const maxPDF = 1.0 / (sigma * Math.sqrt(2 * Math.PI));
  const yMax = viewMode === 'pdf' ? maxPDF * 1.15 : 1.1;

  const scaleX = (x: number) => margin.left + ((x - xMin) / (xMax - xMin)) * innerWidth;
  const scaleY = (y: number) => margin.top + innerHeight - (y / yMax) * innerHeight;

  // Build SVG path
  const curvePath = points.length > 0
    ? points.reduce((path, pt, i) => {
        const px = scaleX(pt.x);
        const py = scaleY(viewMode === 'pdf' ? pt.pdf : pt.cdf);
        return i === 0 ? `M ${px} ${py}` : `${path} L ${px} ${py}`;
      }, '')
    : '';

  // Shaded region path for PDF
  let shadedPath = '';
  if (viewMode === 'pdf' && points.length > 0) {
    const isShaded = (x: number) => {
      if (tail === 'lower') return x <= (data?.cutoff_x ?? cutoffX);
      if (tail === 'upper') return x >= (data?.cutoff_x ?? cutoffX);
      if (tail === 'interval') {
        const x1 = data?.cutoff_x1 ?? cutoffX1;
        const x2 = data?.cutoff_x2 ?? cutoffX2;
        return x >= x1 && x <= x2;
      }
      if (tail === 'two_tailed') {
        const x1 = data?.cutoff_x1 ?? (mu - Math.abs(cutoffX - mu));
        const x2 = data?.cutoff_x2 ?? (mu + Math.abs(cutoffX - mu));
        return x <= x1 || x >= x2;
      }
      return false;
    };

    const shadedPoints = points.filter((pt) => isShaded(pt.x));
    if (shadedPoints.length > 0) {
      if (tail === 'two_tailed') {
        // Two disjoint regions
        const leftRegion = points.filter((pt) => pt.x <= (data?.cutoff_x1 ?? (mu - Math.abs(cutoffX - mu))));
        const rightRegion = points.filter((pt) => pt.x >= (data?.cutoff_x2 ?? (mu + Math.abs(cutoffX - mu))));

        const makeSubPath = (sub: typeof points) => {
          if (sub.length === 0) return '';
          const first = sub[0];
          const last = sub[sub.length - 1];
          let p = `M ${scaleX(first.x)} ${scaleY(0)}`;
          sub.forEach((pt) => {
            p += ` L ${scaleX(pt.x)} ${scaleY(pt.pdf)}`;
          });
          p += ` L ${scaleX(last.x)} ${scaleY(0)} Z`;
          return p;
        };

        shadedPath = `${makeSubPath(leftRegion)} ${makeSubPath(rightRegion)}`;
      } else {
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
  }

  // Ticks on X-axis (z = -3, -2, -1, 0, 1, 2, 3)
  const zTicks = [-3, -2, -1, 0, 1, 2, 3];

  return (
    <div className="visual-card" style={{ background: 'var(--bg-surface)', padding: '1.25rem', borderRadius: 'var(--radius-md)', border: '1px solid var(--border-subtle)' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem', flexWrap: 'wrap', gap: '0.75rem' }}>
        <div>
          <h3 style={{ margin: 0, fontSize: '1.2rem', color: 'var(--text-main)', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            Normal Distribution Explorer
            <span className="badge badge-cyan" style={{ fontSize: '0.75rem' }}>Continuous N(μ, σ²)</span>
          </h3>
          <p style={{ margin: '0.25rem 0 0', fontSize: '0.85rem', color: 'var(--text-muted)' }}>
            Bell-shaped density curve where curve height is probability density $f(x)$ and shaded area is probability $P(X \in A)$.
          </p>
        </div>

        <div style={{ display: 'flex', gap: '0.5rem' }}>
          <button
            type="button"
            className={`btn btn-sm ${viewMode === 'pdf' ? 'btn-primary' : 'btn-outline'}`}
            onClick={() => setViewMode('pdf')}
            id="btn-toggle-normal-pdf"
            aria-pressed={viewMode === 'pdf'}
          >
            PDF (Density Curve)
          </button>
          <button
            type="button"
            className={`btn btn-sm ${viewMode === 'cdf' ? 'btn-primary' : 'btn-outline'}`}
            onClick={() => setViewMode('cdf')}
            id="btn-toggle-normal-cdf"
            aria-pressed={viewMode === 'cdf'}
          >
            CDF (Cumulative S-Curve)
          </button>
        </div>
      </div>

      {/* Parameter Controls */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '1rem', marginBottom: '1.25rem', background: 'var(--bg-surface-elevated)', padding: '1rem', borderRadius: 'var(--radius-sm)' }}>
        <div>
          <label htmlFor="norm-mean-input" style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-main)', marginBottom: '0.25rem' }}>
            Mean (μ): {mu}
          </label>
          <input
            id="norm-mean-input"
            type="range"
            min="-50"
            max="150"
            step="1"
            value={mu}
            onChange={(e) => setMu(parseFloat(e.target.value))}
            style={{ width: '100%' }}
            aria-label="Normal distribution mean parameter"
          />
        </div>

        <div>
          <label htmlFor="norm-sigma-input" style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-main)', marginBottom: '0.25rem' }}>
            Std Dev (σ): {sigma}
          </label>
          <input
            id="norm-sigma-input"
            type="range"
            min="0.5"
            max="30"
            step="0.5"
            value={sigma}
            onChange={(e) => setSigma(Math.max(0.1, parseFloat(e.target.value)))}
            style={{ width: '100%' }}
            aria-label="Normal distribution standard deviation parameter"
          />
        </div>

        <div>
          <label htmlFor="norm-tail-select" style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-main)', marginBottom: '0.25rem' }}>
            Shaded Event Region:
          </label>
          <select
            id="norm-tail-select"
            value={tail}
            onChange={(e) => setTail(e.target.value as 'lower' | 'upper' | 'interval' | 'two_tailed')}
            className="input"
            style={{ width: '100%', padding: '0.35rem 0.6rem' }}
          >
            <option value="lower">Lower Tail: P(X ≤ x)</option>
            <option value="upper">Upper Tail: P(X ≥ x)</option>
            <option value="interval">Interval: P(x₁ ≤ X ≤ x₂)</option>
            <option value="two_tailed">Two-Tailed: |Z| ≥ z</option>
          </select>
        </div>

        {tail === 'interval' ? (
          <>
            <div>
              <label htmlFor="norm-x1-input" style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-main)', marginBottom: '0.25rem' }}>
                Bound x₁: {cutoffX1.toFixed(2)}
              </label>
              <input
                id="norm-x1-input"
                type="range"
                min={xMin}
                max={cutoffX2}
                step={sigma / 20}
                value={cutoffX1}
                onChange={(e) => setCutoffX1(parseFloat(e.target.value))}
                style={{ width: '100%' }}
              />
            </div>
            <div>
              <label htmlFor="norm-x2-input" style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-main)', marginBottom: '0.25rem' }}>
                Bound x₂: {cutoffX2.toFixed(2)}
              </label>
              <input
                id="norm-x2-input"
                type="range"
                min={cutoffX1}
                max={xMax}
                step={sigma / 20}
                value={cutoffX2}
                onChange={(e) => setCutoffX2(parseFloat(e.target.value))}
                style={{ width: '100%' }}
              />
            </div>
          </>
        ) : (
          <div>
            <label htmlFor="norm-cutoff-input" style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-main)', marginBottom: '0.25rem' }}>
              Cutoff Cutoff x: {cutoffX.toFixed(2)} (Z = {((cutoffX - mu) / sigma).toFixed(2)})
            </label>
            <input
              id="norm-cutoff-input"
              type="range"
              min={xMin}
              max={xMax}
              step={sigma / 20}
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
          aria-label={data?.summary || 'Normal distribution chart'}
        >
          <defs>
            {/* Pattern fill for accessible non-color discrimination */}
            <pattern id={patternId} width="8" height="8" patternTransform="rotate(45 0 0)" patternUnits="userSpaceOnUse">
              <line x1="0" y1="0" x2="0" y2="8" stroke="var(--accent-cyan)" strokeWidth="2.5" strokeOpacity="0.4" />
            </pattern>
          </defs>

          {/* Grid lines */}
          {zTicks.map((z) => {
            const xVal = mu + z * sigma;
            const px = scaleX(xVal);
            return (
              <g key={`grid-${z}`}>
                <line
                  x1={px}
                  y1={margin.top}
                  x2={px}
                  y2={margin.top + innerHeight}
                  stroke="var(--border-subtle)"
                  strokeDasharray="3 3"
                  strokeWidth="1"
                />
                <text
                  x={px}
                  y={margin.top + innerHeight + 16}
                  textAnchor="middle"
                  fill="var(--text-muted)"
                  fontSize="10"
                  fontFamily="monospace"
                >
                  {xVal % 1 === 0 ? xVal.toFixed(0) : xVal.toFixed(1)}
                </text>
                <text
                  x={px}
                  y={margin.top + innerHeight + 28}
                  textAnchor="middle"
                  fill="var(--text-muted)"
                  fontSize="9"
                  opacity="0.8"
                >
                  {z === 0 ? 'μ' : `${z > 0 ? '+' : ''}${z}σ`}
                </text>
              </g>
            );
          })}

          {/* X Axis baseline */}
          <line
            x1={margin.left}
            y1={margin.top + innerHeight}
            x2={margin.left + innerWidth}
            y2={margin.top + innerHeight}
            stroke="var(--border-subtle)"
            strokeWidth="1.5"
          />

          {/* Shaded Area for PDF */}
          {viewMode === 'pdf' && shadedPath && (
            <>
              {/* Semi-transparent color fill */}
              <path d={shadedPath} fill="var(--accent-cyan)" fillOpacity="0.25" />
              {/* Pattern overlay for color-blind accessibility */}
              <path d={shadedPath} fill={`url(#${patternId})`} />
            </>
          )}

          {/* Main distribution curve */}
          {curvePath && (
            <path
              d={curvePath}
              fill="none"
              stroke="var(--accent-cyan)"
              strokeWidth="2.5"
            />
          )}

          {/* Cutoff vertical line(s) */}
          {viewMode === 'pdf' && tail !== 'interval' && (
            <line
              x1={scaleX(data?.cutoff_x ?? cutoffX)}
              y1={margin.top}
              x2={scaleX(data?.cutoff_x ?? cutoffX)}
              y2={margin.top + innerHeight}
              stroke="var(--accent-amber, #e3b341)"
              strokeWidth="2"
              strokeDasharray="4 2"
            />
          )}

          {/* Title label inside canvas */}
          <text
            x={margin.left + 10}
            y={margin.top + 15}
            fill="var(--text-muted)"
            fontSize="11"
            fontWeight="600"
          >
            {viewMode === 'pdf' ? 'Density f(x)' : 'Cumulative P(X ≤ x)'}
          </text>
        </svg>
      </div>

      {/* KPI Cards: Values & Formulas */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))', gap: '0.75rem', marginBottom: '1rem' }}>
        <div style={{ background: 'var(--bg-surface-elevated)', padding: '0.75rem', borderRadius: 'var(--radius-sm)', borderLeft: '3px solid var(--accent-cyan)' }}>
          <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', textTransform: 'uppercase' }}>Shaded Probability</div>
          <div style={{ fontSize: '1.4rem', fontWeight: 700, color: 'var(--accent-cyan)' }} id="normal-shaded-prob-value">
            {data ? (data.shaded_probability * 100).toFixed(2) + '%' : '...'}
          </div>
          <div style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>
            Area = {data ? data.shaded_probability.toFixed(4) : '...'}
          </div>
        </div>

        <div style={{ background: 'var(--bg-surface-elevated)', padding: '0.75rem', borderRadius: 'var(--radius-sm)', borderLeft: '3px solid var(--accent-emerald)' }}>
          <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', textTransform: 'uppercase' }}>Standard Z-Score</div>
          <div style={{ fontSize: '1.4rem', fontWeight: 700, color: 'var(--accent-emerald)' }} id="normal-z-score-value">
            {data && data.z_score !== undefined ? data.z_score.toFixed(2) : ((cutoffX - mu) / sigma).toFixed(2)}
          </div>
          <div style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>
            Z = (X - μ) / σ
          </div>
        </div>

        <div style={{ background: 'var(--bg-surface-elevated)', padding: '0.75rem', borderRadius: 'var(--radius-sm)', borderLeft: '3px solid var(--accent-amber, #e3b341)' }}>
          <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', textTransform: 'uppercase' }}>Density at Cutoff f(x)</div>
          <div style={{ fontSize: '1.4rem', fontWeight: 700, color: 'var(--text-main)' }} id="normal-density-value">
            {data && data.density_at_cutoff !== undefined ? data.density_at_cutoff.toFixed(4) : '...'}
          </div>
          <div style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>
            Curve height ≠ probability
          </div>
        </div>

        <div style={{ background: 'var(--bg-surface-elevated)', padding: '0.75rem', borderRadius: 'var(--radius-sm)', borderLeft: '3px solid var(--accent-purple, #a371f7)' }}>
          <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', textTransform: 'uppercase' }}>Excel Formula</div>
          <div style={{ fontSize: '0.95rem', fontWeight: 600, color: 'var(--text-main)', wordBreak: 'break-all', fontFamily: 'monospace' }} id="normal-excel-formula">
            {data?.excel_formula || '=NORM.DIST(...)'}
          </div>
          <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
            Matches course spreadsheets
          </div>
        </div>
      </div>

      {/* Accessible Text Summary */}
      <div
        role="region"
        aria-live="polite"
        style={{ padding: '0.75rem 1rem', background: 'var(--bg-surface-elevated)', borderRadius: 'var(--radius-sm)', fontSize: '0.85rem', color: 'var(--text-main)' }}
        id="normal-text-summary"
      >
        <strong>Summary: </strong>
        {data?.summary || 'Adjust sliders above to evaluate normal probabilities and shaded regions.'}
      </div>
    </div>
  );
};

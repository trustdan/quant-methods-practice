import React, { useState, useEffect, useId } from 'react';
import { DiscreteVisualResponse } from '../types/visuals';

export const DiscreteVisual: React.FC = () => {
  const [family, setFamily] = useState<'binomial' | 'poisson'>('binomial');
  const [n, setN] = useState<number>(10);
  const [p, setP] = useState<number>(0.5);
  const [lambda, setLambda] = useState<number>(4.0);
  const [event, setEvent] = useState<'equal' | 'at_most' | 'at_least' | 'between'>('equal');
  const [k, setK] = useState<number>(5);
  const [k1, setK1] = useState<number>(2);
  const [k2, setK2] = useState<number>(6);
  const [viewMode, setViewMode] = useState<'pmf' | 'cdf'>('pmf');
  const [data, setData] = useState<DiscreteVisualResponse | null>(null);

  const patternId = useId();

  useEffect(() => {
    let active = true;
    const fetchBars = async () => {
      try {
        const params = new URLSearchParams({
          family,
          event,
          k: k.toString(),
          k1: k1.toString(),
          k2: k2.toString(),
        });
        if (family === 'binomial') {
          params.set('n', n.toString());
          params.set('p', p.toString());
        } else {
          params.set('lambda', lambda.toString());
        }

        const res = await fetch(`/api/visuals/discrete?${params.toString()}`);
        if (res.ok && active) {
          const json = await res.json();
          setData(json);
        }
      } catch {
        // fallback
      }
    };
    fetchBars();
    return () => {
      active = false;
    };
  }, [family, n, p, lambda, event, k, k1, k2]);

  const width = 640;
  const height = 280;
  const margin = { top: 25, right: 30, bottom: 45, left: 55 };
  const innerWidth = width - margin.left - margin.right;
  const innerHeight = height - margin.top - margin.bottom;

  const bars = data?.bars || [];
  const maxPMF = bars.length > 0 ? Math.max(...bars.map((b) => b.pmf)) : 0.4;
  const yMax = viewMode === 'pmf' ? maxPMF * 1.2 : 1.1;

  const barCount = bars.length;
  const barWidth = barCount > 0 ? (innerWidth / barCount) * 0.7 : 20;
  const barGap = barCount > 0 ? innerWidth / barCount : 25;

  const scaleY = (y: number) => margin.top + innerHeight - (y / yMax) * innerHeight;

  return (
    <div className="visual-card" style={{ background: 'var(--bg-surface)', padding: '1.25rem', borderRadius: 'var(--radius-md)', border: '1px solid var(--border-subtle)' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem', flexWrap: 'wrap', gap: '0.75rem' }}>
        <div>
          <h3 style={{ margin: 0, fontSize: '1.2rem', color: 'var(--text-main)', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            Discrete Distribution Explorer
            <span className="badge badge-cyan" style={{ fontSize: '0.75rem' }}>PMF &amp; CDF</span>
          </h3>
          <p style={{ margin: '0.25rem 0 0', fontSize: '0.85rem', color: 'var(--text-muted)' }}>
            Probability mass function (bars) where bar height is exact probability $P(X = k)$, plus cumulative step function.
          </p>
        </div>

        <div style={{ display: 'flex', gap: '0.5rem' }}>
          <button
            type="button"
            className={`btn btn-sm ${viewMode === 'pmf' ? 'btn-primary' : 'btn-outline'}`}
            onClick={() => setViewMode('pmf')}
            id="btn-toggle-discrete-pmf"
            aria-pressed={viewMode === 'pmf'}
          >
            PMF (Bars)
          </button>
          <button
            type="button"
            className={`btn btn-sm ${viewMode === 'cdf' ? 'btn-primary' : 'btn-outline'}`}
            onClick={() => setViewMode('cdf')}
            id="btn-toggle-discrete-cdf"
            aria-pressed={viewMode === 'cdf'}
          >
            CDF (Staircase)
          </button>
        </div>
      </div>

      {/* Parameter Controls */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))', gap: '1rem', marginBottom: '1.25rem', background: 'var(--bg-surface-elevated)', padding: '1rem', borderRadius: 'var(--radius-sm)' }}>
        <div>
          <label htmlFor="disc-family-select" style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-main)', marginBottom: '0.25rem' }}>
            Distribution:
          </label>
          <select
            id="disc-family-select"
            value={family}
            onChange={(e) => setFamily(e.target.value as 'binomial' | 'poisson')}
            className="input"
            style={{ width: '100%', padding: '0.35rem 0.6rem' }}
          >
            <option value="binomial">Binomial (n trials, prob p)</option>
            <option value="poisson">Poisson (rate λ)</option>
          </select>
        </div>

        {family === 'binomial' ? (
          <>
            <div>
              <label htmlFor="disc-n-input" style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-main)', marginBottom: '0.25rem' }}>
                Trials (n): {n}
              </label>
              <input
                id="disc-n-input"
                type="range"
                min="1"
                max="25"
                step="1"
                value={n}
                onChange={(e) => {
                  const val = parseInt(e.target.value, 10);
                  setN(val);
                  setK(Math.min(k, val));
                }}
                style={{ width: '100%' }}
              />
            </div>
            <div>
              <label htmlFor="disc-p-input" style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-main)', marginBottom: '0.25rem' }}>
                Success Prob (p): {p}
              </label>
              <input
                id="disc-p-input"
                type="range"
                min="0.05"
                max="0.95"
                step="0.05"
                value={p}
                onChange={(e) => setP(parseFloat(e.target.value))}
                style={{ width: '100%' }}
              />
            </div>
          </>
        ) : (
          <div>
            <label htmlFor="disc-lambda-input" style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-main)', marginBottom: '0.25rem' }}>
              Rate (λ): {lambda}
            </label>
            <input
              id="disc-lambda-input"
              type="range"
              min="0.5"
              max="15"
              step="0.5"
              value={lambda}
              onChange={(e) => setLambda(parseFloat(e.target.value))}
              style={{ width: '100%' }}
            />
          </div>
        )}

        <div>
          <label htmlFor="disc-event-select" style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-main)', marginBottom: '0.25rem' }}>
            Event:
          </label>
          <select
            id="disc-event-select"
            value={event}
            onChange={(e) => setEvent(e.target.value as 'equal' | 'at_most' | 'at_least' | 'between')}
            className="input"
            style={{ width: '100%', padding: '0.35rem 0.6rem' }}
          >
            <option value="equal">Exactly: P(X = k)</option>
            <option value="at_most">At Most: P(X ≤ k)</option>
            <option value="at_least">At Least: P(X ≥ k)</option>
            <option value="between">Between: P(k₁ ≤ X ≤ k₂)</option>
          </select>
        </div>

        {event === 'between' ? (
          <>
            <div>
              <label htmlFor="disc-k1-input" style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-main)', marginBottom: '0.25rem' }}>
                Lower k₁: {k1}
              </label>
              <input
                id="disc-k1-input"
                type="range"
                min="0"
                max={k2}
                step="1"
                value={k1}
                onChange={(e) => setK1(parseInt(e.target.value, 10))}
                style={{ width: '100%' }}
              />
            </div>
            <div>
              <label htmlFor="disc-k2-input" style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-main)', marginBottom: '0.25rem' }}>
                Upper k₂: {k2}
              </label>
              <input
                id="disc-k2-input"
                type="range"
                min={k1}
                max={family === 'binomial' ? n : 20}
                step="1"
                value={k2}
                onChange={(e) => setK2(parseInt(e.target.value, 10))}
                style={{ width: '100%' }}
              />
            </div>
          </>
        ) : (
          <div>
            <label htmlFor="disc-k-input" style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-main)', marginBottom: '0.25rem' }}>
              Target Value k: {k}
            </label>
            <input
              id="disc-k-input"
              type="range"
              min="0"
              max={family === 'binomial' ? n : 20}
              step="1"
              value={k}
              onChange={(e) => setK(parseInt(e.target.value, 10))}
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
          aria-label={data?.summary || 'Discrete distribution chart'}
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

          {viewMode === 'pmf' ? (
            // Bars
            bars.map((bar, i) => {
              const xPos = margin.left + i * barGap + (barGap - barWidth) / 2;
              const yPos = scaleY(bar.pmf);
              const barH = scaleY(0) - yPos;

              return (
                <g key={`bar-${bar.k}`}>
                  <rect
                    x={xPos}
                    y={yPos}
                    width={barWidth}
                    height={barH}
                    fill={bar.is_highlighted ? 'var(--accent-cyan)' : 'var(--border-subtle)'}
                    fillOpacity={bar.is_highlighted ? 0.9 : 0.4}
                    stroke={bar.is_highlighted ? 'var(--accent-cyan)' : 'transparent'}
                    strokeWidth="1"
                    rx="2"
                  />
                  {bar.is_highlighted && (
                    <rect
                      x={xPos}
                      y={yPos}
                      width={barWidth}
                      height={barH}
                      fill={`url(#${patternId})`}
                      rx="2"
                    />
                  )}
                  {barCount <= 16 && (
                    <text
                      x={xPos + barWidth / 2}
                      y={scaleY(0) + 16}
                      textAnchor="middle"
                      fill={bar.is_highlighted ? 'var(--accent-cyan)' : 'var(--text-muted)'}
                      fontSize="10"
                      fontWeight={bar.is_highlighted ? 700 : 400}
                    >
                      {bar.k}
                    </text>
                  )}
                </g>
              );
            })
          ) : (
            // CDF Staircase step line
            bars.map((bar, i) => {
              const xStart = margin.left + i * barGap;
              const xEnd = margin.left + (i + 1) * barGap;
              const yVal = scaleY(bar.cdf);
              const nextY = i < bars.length - 1 ? scaleY(bars[i + 1].cdf) : yVal;

              return (
                <g key={`step-${bar.k}`}>
                  {/* Step horizontal */}
                  <line
                    x1={xStart}
                    y1={yVal}
                    x2={xEnd}
                    y2={yVal}
                    stroke="var(--accent-emerald)"
                    strokeWidth="2.5"
                  />
                  {/* Step vertical rise */}
                  {i < bars.length - 1 && (
                    <line
                      x1={xEnd}
                      y1={yVal}
                      x2={xEnd}
                      y2={nextY}
                      stroke="var(--accent-emerald)"
                      strokeWidth="1.5"
                      strokeDasharray="2 2"
                    />
                  )}
                  {/* Point at jump */}
                  <circle
                    cx={xStart}
                    cy={yVal}
                    r="3.5"
                    fill="var(--accent-emerald)"
                  />
                </g>
              );
            })
          )}
        </svg>
      </div>

      {/* KPI Cards */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))', gap: '0.75rem', marginBottom: '1rem' }}>
        <div style={{ background: 'var(--bg-surface-elevated)', padding: '0.75rem', borderRadius: 'var(--radius-sm)', borderLeft: '3px solid var(--accent-cyan)' }}>
          <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', textTransform: 'uppercase' }}>Event Probability</div>
          <div style={{ fontSize: '1.4rem', fontWeight: 700, color: 'var(--accent-cyan)' }} id="discrete-total-prob-value">
            {data ? (data.total_probability * 100).toFixed(2) + '%' : '...'}
          </div>
          <div style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>
            P = {data ? data.total_probability.toFixed(4) : '...'}
          </div>
        </div>

        <div style={{ background: 'var(--bg-surface-elevated)', padding: '0.75rem', borderRadius: 'var(--radius-sm)', borderLeft: '3px solid var(--accent-emerald)' }}>
          <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', textTransform: 'uppercase' }}>Moments</div>
          <div style={{ fontSize: '1.4rem', fontWeight: 700, color: 'var(--text-main)' }}>
            μ = {data ? data.mean.toFixed(2) : '...'}
          </div>
          <div style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>
            σ² = {data ? data.variance.toFixed(2) : '...'}
          </div>
        </div>

        <div style={{ background: 'var(--bg-surface-elevated)', padding: '0.75rem', borderRadius: 'var(--radius-sm)', borderLeft: '3px solid var(--accent-purple, #a371f7)' }}>
          <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', textTransform: 'uppercase' }}>Excel Formula</div>
          <div style={{ fontSize: '0.95rem', fontWeight: 600, color: 'var(--text-main)', wordBreak: 'break-all', fontFamily: 'monospace' }} id="discrete-excel-formula">
            {data?.excel_formula || '=BINOM.DIST(...)'}
          </div>
          <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
            Spreadsheet equivalent
          </div>
        </div>
      </div>

      <div
        role="region"
        aria-live="polite"
        style={{ padding: '0.75rem 1rem', background: 'var(--bg-surface-elevated)', borderRadius: 'var(--radius-sm)', fontSize: '0.85rem', color: 'var(--text-main)' }}
        id="discrete-text-summary"
      >
        <strong>Summary: </strong>
        {data?.summary || 'Adjust discrete parameters to view probabilities.'}
      </div>
    </div>
  );
};

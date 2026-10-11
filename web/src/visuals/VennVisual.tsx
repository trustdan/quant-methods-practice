import React, { useState, useEffect, useId } from 'react';
import { VennVisualResponse } from '../types/visuals';

export const VennVisual: React.FC = () => {
  const [pA, setPA] = useState<number>(0.6);
  const [pB, setPB] = useState<number>(0.5);
  const [pInter, setPInter] = useState<number>(0.3);
  const [region, setRegion] = useState<string>('intersection');
  const [data, setData] = useState<VennVisualResponse | null>(null);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  const patternId = useId();

  // Validate Fréchet bounds
  const minInter = Math.max(0, pA + pB - 1);
  const maxInter = Math.min(pA, pB);
  const isInterValid = pInter >= minInter - 1e-6 && pInter <= maxInter + 1e-6;

  useEffect(() => {
    let active = true;
    const fetchVenn = async () => {
      if (!isInterValid) {
        setErrorMsg(`Fréchet bound violation: For P(A)=${pA.toFixed(2)} and P(B)=${pB.toFixed(2)}, P(A ∩ B) must be between ${minInter.toFixed(2)} and ${maxInter.toFixed(2)}.`);
        return;
      }
      setErrorMsg(null);

      try {
        const params = new URLSearchParams({
          p_a: pA.toString(),
          p_b: pB.toString(),
          p_intersection: pInter.toString(),
          region,
        });
        const res = await fetch(`/api/visuals/venn?${params.toString()}`);
        if (res.ok && active) {
          const json = await res.json();
          setData(json);
        }
      } catch {
        // local fallback
      }
    };
    fetchVenn();
    return () => {
      active = false;
    };
  }, [pA, pB, pInter, region, isInterValid, minInter, maxInter]);

  const width = 500;
  const height = 300;

  // Circle centers and radius
  const r = 85;
  const cAx = 195;
  const cAy = 150;
  const cBx = 305;
  const cBy = 150;

  // Region active states
  const isAOnlyActive = region === 'a_only' || region === 'union' || region === 'complement_b';
  const isBOnlyActive = region === 'b_only' || region === 'union' || region === 'complement_a';
  const isInterActive = region === 'intersection' || region === 'union';
  const isNeitherActive = region === 'neither' || region === 'complement_a' || region === 'complement_b';

  return (
    <div className="visual-card" style={{ background: 'var(--bg-surface)', padding: '1.25rem', borderRadius: 'var(--radius-md)', border: '1px solid var(--border-subtle)' }}>
      <div style={{ marginBottom: '1rem' }}>
        <h3 style={{ margin: 0, fontSize: '1.2rem', color: 'var(--text-main)', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
          Venn Diagram Explorer
          <span className="badge badge-cyan" style={{ fontSize: '0.75rem' }}>Set Operations</span>
        </h3>
        <p style={{ margin: '0.25rem 0 0', fontSize: '0.85rem', color: 'var(--text-muted)' }}>
          Visualize union, intersection, conditional subsets, and complement events with Fréchet inequality constraints.
        </p>
      </div>

      {/* Controls */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))', gap: '1rem', marginBottom: '1.25rem', background: 'var(--bg-surface-elevated)', padding: '1rem', borderRadius: 'var(--radius-sm)' }}>
        <div>
          <label htmlFor="venn-pa-input" style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-main)', marginBottom: '0.25rem' }}>
            P(A): {pA.toFixed(2)}
          </label>
          <input
            id="venn-pa-input"
            type="range"
            min="0.1"
            max="0.9"
            step="0.05"
            value={pA}
            onChange={(e) => {
              const val = parseFloat(e.target.value);
              setPA(val);
              setPInter(Math.max(0, Math.min(val, pB, Math.max(0, val + pB - 1))));
            }}
            style={{ width: '100%' }}
          />
        </div>

        <div>
          <label htmlFor="venn-pb-input" style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-main)', marginBottom: '0.25rem' }}>
            P(B): {pB.toFixed(2)}
          </label>
          <input
            id="venn-pb-input"
            type="range"
            min="0.1"
            max="0.9"
            step="0.05"
            value={pB}
            onChange={(e) => {
              const val = parseFloat(e.target.value);
              setPB(val);
              setPInter(Math.max(0, Math.min(pA, val, Math.max(0, pA + val - 1))));
            }}
            style={{ width: '100%' }}
          />
        </div>

        <div>
          <label htmlFor="venn-pinter-input" style={{ display: 'block', fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-main)', marginBottom: '0.25rem' }}>
            P(A ∩ B): {pInter.toFixed(2)}
          </label>
          <input
            id="venn-pinter-input"
            type="range"
            min={minInter}
            max={maxInter}
            step="0.05"
            value={pInter}
            onChange={(e) => setPInter(parseFloat(e.target.value))}
            style={{ width: '100%' }}
          />
        </div>
      </div>

      {errorMsg && (
        <div style={{ color: 'var(--danger-color, #f85149)', background: 'rgba(248, 81, 73, 0.1)', padding: '0.5rem 0.75rem', borderRadius: 'var(--radius-sm)', marginBottom: '1rem', fontSize: '0.85rem' }}>
          {errorMsg}
        </div>
      )}

      {/* Region Selector Buttons */}
      <div style={{ display: 'flex', gap: '0.4rem', flexWrap: 'wrap', marginBottom: '1rem' }}>
        {[
          { id: 'intersection', label: 'A ∩ B (Overlap)' },
          { id: 'union', label: 'A ∪ B (Union)' },
          { id: 'a_only', label: 'A \\ B (A Only)' },
          { id: 'b_only', label: 'B \\ A (B Only)' },
          { id: 'neither', label: '(A ∪ B)ᶜ (Neither)' },
          { id: 'complement_a', label: 'Aᶜ (Not A)' },
          { id: 'complement_b', label: 'Bᶜ (Not B)' },
        ].map((btn) => (
          <button
            key={btn.id}
            type="button"
            className={`btn btn-sm ${region === btn.id ? 'btn-primary' : 'btn-outline'}`}
            onClick={() => setRegion(btn.id)}
            id={`btn-venn-${btn.id}`}
            aria-pressed={region === btn.id}
          >
            {btn.label}
          </button>
        ))}
      </div>

      {/* SVG Canvas */}
      <div style={{ position: 'relative', width: '100%', overflowX: 'auto', background: 'var(--bg-canvas, #0d1117)', borderRadius: 'var(--radius-sm)', border: '1px solid var(--border-subtle)', marginBottom: '1rem' }}>
        <svg
          viewBox={`0 0 ${width} ${height}`}
          style={{ width: '100%', height: 'auto', display: 'block' }}
          role="img"
          aria-label={data?.summary || 'Venn diagram of events A and B'}
        >
          <defs>
            <pattern id={patternId} width="8" height="8" patternTransform="rotate(45 0 0)" patternUnits="userSpaceOnUse">
              <line x1="0" y1="0" x2="0" y2="8" stroke="var(--accent-cyan)" strokeWidth="2.5" strokeOpacity="0.4" />
            </pattern>
            {/* Clip paths for boolean set rendering */}
            <clipPath id="clip-circle-a">
              <circle cx={cAx} cy={cAy} r={r} />
            </clipPath>
            <clipPath id="clip-circle-b">
              <circle cx={cBx} cy={cBy} r={r} />
            </clipPath>
          </defs>

          {/* Sample space rectangle S */}
          <rect
            x="20"
            y="20"
            width={width - 40}
            height={height - 40}
            fill={isNeitherActive ? 'var(--accent-cyan)' : 'transparent'}
            fillOpacity={isNeitherActive ? 0.2 : 0}
            stroke="var(--border-subtle)"
            strokeWidth="2"
            rx="6"
          />
          {isNeitherActive && (
            <rect
              x="20"
              y="20"
              width={width - 40}
              height={height - 40}
              fill={`url(#${patternId})`}
              rx="6"
            />
          )}
          <text x="35" y="45" fill="var(--text-muted)" fontSize="14" fontWeight="700">
            S (Universal Sample Space)
          </text>

          {/* Circle A background */}
          <circle
            cx={cAx}
            cy={cAy}
            r={r}
            fill={isAOnlyActive ? 'var(--accent-cyan)' : 'var(--bg-surface)'}
            fillOpacity={isAOnlyActive ? 0.35 : 0.8}
            stroke="var(--accent-cyan)"
            strokeWidth="2.5"
          />
          {isAOnlyActive && (
            <circle
              cx={cAx}
              cy={cAy}
              r={r}
              fill={`url(#${patternId})`}
            />
          )}

          {/* Circle B background */}
          <circle
            cx={cBx}
            cy={cBy}
            r={r}
            fill={isBOnlyActive ? 'var(--accent-emerald)' : 'var(--bg-surface)'}
            fillOpacity={isBOnlyActive ? 0.35 : 0.8}
            stroke="var(--accent-emerald)"
            strokeWidth="2.5"
          />
          {isBOnlyActive && (
            <circle
              cx={cBx}
              cy={cBy}
              r={r}
              fill={`url(#${patternId})`}
            />
          )}

          {/* Intersection region: Circle B clipped by Circle A */}
          <g clipPath="url(#clip-circle-a)">
            <circle
              cx={cBx}
              cy={cBy}
              r={r}
              fill={isInterActive ? 'var(--accent-amber, #e3b341)' : 'var(--bg-surface-elevated)'}
              fillOpacity={isInterActive ? 0.5 : 0.9}
              stroke="none"
            />
            {isInterActive && (
              <circle
                cx={cBx}
                cy={cBy}
                r={r}
                fill={`url(#${patternId})`}
              />
            )}
          </g>

          {/* Redraw circle boundaries */}
          <circle cx={cAx} cy={cAy} r={r} fill="none" stroke="var(--accent-cyan)" strokeWidth="2.5" />
          <circle cx={cBx} cy={cBy} r={r} fill="none" stroke="var(--accent-emerald)" strokeWidth="2.5" />

          {/* Labels & probabilities inside circles */}
          <text x={cAx - 35} y={cAy - 10} fill="var(--accent-cyan)" fontSize="15" fontWeight="700" textAnchor="middle">
            Event A
          </text>
          <text x={cAx - 35} y={cAy + 12} fill="var(--text-main)" fontSize="12" textAnchor="middle">
            {(pA - pInter).toFixed(2)}
          </text>

          <text x={cBx + 35} y={cBy - 10} fill="var(--accent-emerald)" fontSize="15" fontWeight="700" textAnchor="middle">
            Event B
          </text>
          <text x={cBx + 35} y={cBy + 12} fill="var(--text-main)" fontSize="12" textAnchor="middle">
            {(pB - pInter).toFixed(2)}
          </text>

          <text x={(cAx + cBx) / 2} y={cAy - 10} fill="var(--accent-amber, #e3b341)" fontSize="13" fontWeight="700" textAnchor="middle">
            A ∩ B
          </text>
          <text x={(cAx + cBx) / 2} y={cAy + 12} fill="var(--text-main)" fontSize="12" textAnchor="middle">
            {pInter.toFixed(2)}
          </text>

          {/* Neither label */}
          <text x={width - 55} y={height - 35} fill="var(--text-muted)" fontSize="11" textAnchor="end">
            Neither: {(1 - (pA + pB - pInter)).toFixed(2)}
          </text>
        </svg>
      </div>

      {/* KPI Cards */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))', gap: '0.75rem', marginBottom: '1rem' }}>
        <div style={{ background: 'var(--bg-surface-elevated)', padding: '0.75rem', borderRadius: 'var(--radius-sm)', borderLeft: '3px solid var(--accent-cyan)' }}>
          <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', textTransform: 'uppercase' }}>Selected Region Probability</div>
          <div style={{ fontSize: '1.4rem', fontWeight: 700, color: 'var(--accent-cyan)' }} id="venn-shaded-prob-value">
            {data ? (data.shaded_probability * 100).toFixed(1) + '%' : '...'}
          </div>
          <div style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>
            P = {data ? data.shaded_probability.toFixed(4) : '...'}
          </div>
        </div>

        <div style={{ background: 'var(--bg-surface-elevated)', padding: '0.75rem', borderRadius: 'var(--radius-sm)', borderLeft: '3px solid var(--accent-emerald)' }}>
          <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', textTransform: 'uppercase' }}>Formula</div>
          <div style={{ fontSize: '0.95rem', fontWeight: 600, color: 'var(--text-main)', fontFamily: 'monospace' }} id="venn-formula">
            {data?.formula || 'P(A ∩ B)'}
          </div>
          <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
            Analytical identity
          </div>
        </div>

        <div style={{ background: 'var(--bg-surface-elevated)', padding: '0.75rem', borderRadius: 'var(--radius-sm)', borderLeft: '3px solid var(--accent-amber, #e3b341)' }}>
          <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', textTransform: 'uppercase' }}>Independence Check</div>
          <div style={{ fontSize: '1.1rem', fontWeight: 700, color: 'var(--text-main)' }}>
            {Math.abs(pInter - pA * pB) < 1e-6 ? 'Independent' : 'Dependent'}
          </div>
          <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
            P(A)×P(B) = {(pA * pB).toFixed(2)} vs P(A∩B) = {pInter.toFixed(2)}
          </div>
        </div>
      </div>

      <div
        role="region"
        aria-live="polite"
        style={{ padding: '0.75rem 1rem', background: 'var(--bg-surface-elevated)', borderRadius: 'var(--radius-sm)', fontSize: '0.85rem', color: 'var(--text-main)' }}
        id="venn-text-summary"
      >
        <strong>Summary: </strong>
        {data?.summary || 'Select regions and adjust probabilities.'}
      </div>
    </div>
  );
};

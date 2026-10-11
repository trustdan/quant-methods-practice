import React, { useState, useEffect } from 'react';
import { RegressionVisualResponse } from '../types/visuals';

export const RegressionVisual: React.FC = () => {
  const [dataset, setDataset] = useState<string>('advertising_sales');
  const [showResiduals, setShowResiduals] = useState<boolean>(true);
  const [xPred, setXPred] = useState<number>(5.0);
  const [data, setData] = useState<RegressionVisualResponse | null>(null);

  useEffect(() => {
    let active = true;
    const fetchRegression = async () => {
      try {
        const params = new URLSearchParams({
          dataset,
          x_pred: xPred.toString(),
        });
        const res = await fetch(`/api/visuals/regression?${params.toString()}`);
        if (res.ok && active) {
          const json = await res.json();
          setData(json);
        }
      } catch {
        // Fallback
      }
    };
    fetchRegression();
    return () => {
      active = false;
    };
  }, [dataset, xPred]);

  const svgWidth = 640;
  const svgHeight = 320;
  const padding = { top: 25, right: 35, bottom: 45, left: 50 };
  const plotWidth = svgWidth - padding.left - padding.right;
  const plotHeight = svgHeight - padding.top - padding.bottom;

  let minX = 0;
  let maxX = 10;
  let minY = 0;
  let maxY = 45;

  if (data && data.points.length > 0) {
    const xs = data.points.map((p) => p.x);
    const ys = data.points.map((p) => p.y);
    minX = Math.min(...xs) - 1.0;
    maxX = Math.max(...xs, xPred) + 1.0;
    minY = Math.min(0, Math.min(...ys) - 5.0);
    maxY = Math.max(...ys, data.predicted_y ?? 0) + 5.0;
  }

  const scaleX = (x: number) => padding.left + ((x - minX) / (maxX - minX || 1)) * plotWidth;
  const scaleY = (y: number) => padding.top + plotHeight - ((y - minY) / (maxY - minY || 1)) * plotHeight;

  return (
    <div className="visual-container" style={{ display: 'flex', flexDirection: 'column', gap: '1rem', width: '100%', maxWidth: '100%', boxSizing: 'border-box' }}>
      <div className="visual-header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '0.5rem' }}>
        <div>
          <h3 style={{ margin: 0, fontSize: '1.25rem', color: 'var(--color-text-primary, #f1f5f9)' }}>
            Simple Linear Regression & Residuals Lab
          </h3>
          <p style={{ margin: '0.25rem 0 0', fontSize: '0.875rem', color: 'var(--color-text-secondary, #94a3b8)' }}>
            Ordinary Least Squares (OLS) fit, scatter plot, vertical residuals $e_i = y_i - \hat&#123;y&#125;_i$, and $R^2$ variance decomposition.
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
            Dataset
          </label>
          <select
            aria-label="Regression dataset selector"
            value={dataset}
            onChange={(e) => {
              setDataset(e.target.value);
              setXPred(e.target.value === 'study_exam' ? 6.0 : 5.0);
            }}
            style={{ width: '100%', padding: '0.45rem', borderRadius: '0.25rem', background: 'var(--color-bg, #0f172a)', color: '#fff', border: '1px solid var(--color-border, #334155)' }}
          >
            <option value="advertising_sales">Advertising ($k) vs Sales ($k)</option>
            <option value="study_exam">Study Hours vs Exam Score</option>
          </select>
        </div>

        <div>
          <label style={{ display: 'block', fontSize: '0.75rem', fontWeight: 600, color: 'var(--color-text-secondary, #94a3b8)', marginBottom: '0.35rem' }}>
            Predict $\hat&#123;y&#125;$ at $X^* = {xPred}$
          </label>
          <input
            aria-label="Prediction X input"
            type="range"
            min={dataset === 'study_exam' ? '1' : '1'}
            max={dataset === 'study_exam' ? '14' : '10'}
            step="0.5"
            value={xPred}
            onChange={(e) => setXPred(parseFloat(e.target.value))}
            style={{ width: '100%' }}
          />
        </div>

        <div style={{ display: 'flex', alignItems: 'center', marginTop: '1rem' }}>
          <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', cursor: 'pointer', fontSize: '0.85rem', color: '#cbd5e1' }}>
            <input
              type="checkbox"
              checked={showResiduals}
              onChange={(e) => setShowResiduals(e.target.checked)}
            />
            Show Vertical Residuals ($e_i$)
          </label>
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
          aria-label={data?.summary || 'Linear regression visual'}
        >
          {/* Axes */}
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

          {/* OLS Trendline */}
          {data && (
            <line
              x1={scaleX(data.line_start_x)}
              y1={scaleY(data.line_start_y)}
              x2={scaleX(data.line_end_x)}
              y2={scaleY(data.line_end_y)}
              stroke="#38bdf8"
              strokeWidth="2.5"
            />
          )}

          {/* Residual lines */}
          {showResiduals &&
            data?.residuals.map((r, i) => (
              <line
                key={`res-${i}`}
                x1={scaleX(r.x)}
                y1={scaleY(r.y)}
                x2={scaleX(r.x)}
                y2={scaleY(r.y_hat)}
                stroke="#ef4444"
                strokeWidth="1.75"
                strokeDasharray="3 2"
              />
            ))}

          {/* Data Points */}
          {data?.points.map((p, i) => (
            <circle
              key={`pt-${i}`}
              cx={scaleX(p.x)}
              cy={scaleY(p.y)}
              r="5"
              fill="#f59e0b"
              stroke="#1e293b"
              strokeWidth="1.5"
            />
          ))}

          {/* Prediction Marker */}
          {data && data.predicted_y !== undefined && (
            <g>
              <line
                x1={scaleX(xPred)}
                y1={scaleY(minY)}
                x2={scaleX(xPred)}
                y2={scaleY(data.predicted_y)}
                stroke="#10b981"
                strokeWidth="1.5"
                strokeDasharray="2 2"
              />
              <circle
                cx={scaleX(xPred)}
                cy={scaleY(data.predicted_y)}
                r="6"
                fill="#10b981"
                stroke="#fff"
                strokeWidth="2"
              />
              <text
                x={scaleX(xPred)}
                y={scaleY(data.predicted_y) - 10}
                textAnchor="middle"
                fill="#34d399"
                fontSize="11"
                fontWeight="bold"
              >
                $\hat&#123;y&#125; = {data.predicted_y}$
              </text>
            </g>
          )}

          {/* Axes labels */}
          <text
            x={padding.left + plotWidth / 2}
            y={svgHeight - 10}
            textAnchor="middle"
            fill="var(--color-text-secondary, #94a3b8)"
            fontSize="12"
          >
            {dataset === 'study_exam' ? 'Study Hours (X)' : 'Advertising Expenditure in $k (X)'}
          </text>
          <text
            x={15}
            y={padding.top + plotHeight / 2}
            textAnchor="middle"
            transform={`rotate(-90, 15, ${padding.top + plotHeight / 2})`}
            fill="var(--color-text-secondary, #94a3b8)"
            fontSize="11"
          >
            {dataset === 'study_exam' ? 'Exam Score (Y)' : 'Sales Revenue in $k (Y)'}
          </text>
        </svg>

        {/* Legend */}
        <div style={{ display: 'flex', gap: '1.25rem', justifyContent: 'center', marginTop: '0.75rem', fontSize: '0.8rem', color: '#cbd5e1', flexWrap: 'wrap' }}>
          <span style={{ display: 'flex', alignItems: 'center', gap: '0.35rem' }}>
            <span style={{ width: 10, height: 10, borderRadius: '50%', background: '#f59e0b', display: 'inline-block' }}></span>
            Observed Data $(x_i, y_i)$
          </span>
          <span style={{ display: 'flex', alignItems: 'center', gap: '0.35rem' }}>
            <span style={{ width: 16, height: 3, background: '#38bdf8', display: 'inline-block' }}></span>
            Fitted Line $\hat&#123;y&#125; = b_0 + b_1 x$
          </span>
          <span style={{ display: 'flex', alignItems: 'center', gap: '0.35rem' }}>
            <span style={{ width: 16, height: 2, borderTop: '2px dashed #ef4444', display: 'inline-block' }}></span>
            Residuals $e_i = y_i - \hat&#123;y&#125;_i$
          </span>
        </div>
      </div>

      {/* Metrics Row */}
      {data && (
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))', gap: '0.75rem' }}>
          <div style={{ background: 'var(--color-surface, #1e293b)', padding: '0.75rem', borderRadius: '0.375rem', border: '1px solid var(--color-border, #334155)' }}>
            <div style={{ fontSize: '0.75rem', color: '#94a3b8' }}>Fitted Equation</div>
            <div id="reg-slope-value" style={{ fontSize: '1.05rem', fontWeight: 700, color: '#38bdf8', marginTop: '0.2rem' }}>
              $\hat&#123;y&#125; = {data.slope}x + {data.intercept}$
            </div>
            <div style={{ fontSize: '0.75rem', color: '#64748b', marginTop: '0.2rem' }}>
              Slope $b_1 = {data.slope}$, Intercept $b_0 = {data.intercept}$
            </div>
          </div>

          <div style={{ background: 'var(--color-surface, #1e293b)', padding: '0.75rem', borderRadius: '0.375rem', border: '1px solid var(--color-border, #334155)' }}>
            <div style={{ fontSize: '0.75rem', color: '#94a3b8' }}>Fit & Correlation</div>
            <div id="reg-r2-value" style={{ fontSize: '1.05rem', fontWeight: 700, color: '#10b981', marginTop: '0.2rem' }}>
              $R^2 = {data.r_squared}$ ($r = {data.correlation}$)
            </div>
            <div style={{ fontSize: '0.75rem', color: '#64748b', marginTop: '0.2rem' }}>
              {(data.r_squared * 100).toFixed(1)}% of variance in $Y$ explained
            </div>
          </div>

          <div style={{ background: 'var(--color-surface, #1e293b)', padding: '0.75rem', borderRadius: '0.375rem', border: '1px solid var(--color-border, #334155)' }}>
            <div style={{ fontSize: '0.75rem', color: '#94a3b8' }}>Excel Formulas</div>
            <code id="reg-excel-formula" style={{ fontSize: '0.8rem', color: '#fbbf24', display: 'block', marginTop: '0.25rem' }}>
              {data.excel_formula || '=SLOPE(y_range, x_range)'}
            </code>
            <div style={{ fontSize: '0.75rem', color: '#64748b', marginTop: '0.2rem' }}>
              `=INTERCEPT(...)` &amp; `=RSQ(...)`
            </div>
          </div>

          <div style={{ background: 'var(--color-surface, #1e293b)', padding: '0.75rem', borderRadius: '0.375rem', border: '1px solid var(--color-border, #334155)' }}>
            <div style={{ fontSize: '0.75rem', color: '#94a3b8' }}>Standard Error of Estimate</div>
            <div style={{ fontSize: '1.05rem', fontWeight: 700, color: '#f59e0b', marginTop: '0.2rem' }}>
              $s_e = {data.std_err_est}$
            </div>
            <div style={{ fontSize: '0.75rem', color: '#64748b', marginTop: '0.2rem' }}>
              Residual standard deviation
            </div>
          </div>
        </div>
      )}

      {/* Association is not Causation Callout */}
      <div id="reg-association-warning" style={{ background: 'rgba(239, 68, 68, 0.08)', borderLeft: '4px solid #ef4444', padding: '0.75rem 1rem', borderRadius: '0.25rem' }}>
        <p style={{ margin: 0, fontSize: '0.85rem', color: '#cbd5e1' }}>
          <strong>Core Statistical Invariant:</strong> Association is not causation. A high $r$ or $R^2$ indicates strong linear predictive association, but cannot establish causal influence without experimental control (lurking variables like store size or firm reputation may drive both variables).
        </p>
      </div>

      {/* ARIA summary */}
      <div role="status" aria-live="polite" className="sr-only" style={{ position: 'absolute', width: 1, height: 1, padding: 0, margin: -1, overflow: 'hidden', clip: 'rect(0, 0, 0, 0)', border: 0 }}>
        {data?.summary}
      </div>
    </div>
  );
};

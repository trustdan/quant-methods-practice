import React, { useState, useMemo } from 'react';
import { SOFTWARE_RECIPES } from './softwareRecipes';

export interface SoftwareCheatsheetCatalogProps {
  initialModule?: string;
}

export const MODULE_NAMES: Record<string, string> = {
  module_1: 'Module 1: Descriptive & Set Probability',
  module_2: 'Module 2: Discrete RVs (Binomial, Poisson)',
  module_3: 'Module 3: Continuous (Normal, Uniform, Exponential)',
  module_4: 'Module 4: Sampling, CLT & Estimation',
  module_5: 'Module 5: Hypothesis Testing',
  module_6: 'Module 6: Regression Analysis & Diagnostics',
};

export const SoftwareCheatsheetCatalog: React.FC<SoftwareCheatsheetCatalogProps> = ({
  initialModule = 'all',
}) => {
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedModule, setSelectedModule] = useState<string>(initialModule);
  const [displayMode, setDisplayMode] = useState<'side-by-side' | 'excel' | 'r'>('side-by-side');
  const [expandedId, setExpandedId] = useState<string | null>(null);
  const [copiedKey, setCopiedKey] = useState<string | null>(null);

  const filteredRecipes = useMemo(() => {
    return SOFTWARE_RECIPES.filter((recipe) => {
      if (selectedModule !== 'all' && recipe.moduleId !== selectedModule) {
        return false;
      }
      if (!searchQuery.trim()) return true;

      const q = searchQuery.toLowerCase();
      return (
        recipe.title.toLowerCase().includes(q) ||
        recipe.statisticalTask.toLowerCase().includes(q) ||
        recipe.excel.functionName.toLowerCase().includes(q) ||
        recipe.excel.genericSyntax.toLowerCase().includes(q) ||
        recipe.r.functionName.toLowerCase().includes(q) ||
        recipe.r.genericSyntax.toLowerCase().includes(q) ||
        recipe.applicableKeywords?.some((kw) => kw.toLowerCase().includes(q))
      );
    });
  }, [searchQuery, selectedModule]);

  const handleCopy = async (key: string, textToCopy: string) => {
    try {
      await navigator.clipboard.writeText(textToCopy);
      setCopiedKey(key);
      setTimeout(() => setCopiedKey(null), 2000);
    } catch {
      const ta = document.createElement('textarea');
      ta.value = textToCopy;
      document.body.appendChild(ta);
      ta.select();
      document.execCommand('copy');
      document.body.removeChild(ta);
      setCopiedKey(key);
      setTimeout(() => setCopiedKey(null), 2000);
    }
  };

  const handlePrint = () => {
    window.print();
  };

  return (
    <div className="software-cheatsheet-catalog" data-testid="software-cheatsheet-catalog">
      {/* Toolbar & Controls */}
      <div
        className="card cheatsheet-toolbar no-print"
        style={{
          padding: '1rem',
          display: 'flex',
          flexDirection: 'column',
          gap: '0.85rem',
          marginBottom: '1.25rem',
        }}
      >
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '0.75rem' }}>
          <div>
            <h3 style={{ margin: 0, fontSize: '1.15rem', fontWeight: 600 }}>
              Statistical Software Cheatsheet (Excel &amp; R)
            </h3>
            <p style={{ margin: '0.2rem 0 0 0', fontSize: '0.82rem', color: 'var(--text-muted)' }}>
              Complete reference catalog of all 50 canonical statistical software recipes mapped across the curriculum.
            </p>
          </div>

          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            {/* Display Mode Switcher */}
            <div role="group" aria-label="Software Display Mode" style={{ display: 'flex', gap: '0.25rem' }}>
              <button
                type="button"
                className={`btn ${displayMode === 'side-by-side' ? 'btn-primary' : 'btn-outline'}`}
                style={{ padding: '0.3rem 0.6rem', fontSize: '0.78rem' }}
                onClick={() => setDisplayMode('side-by-side')}
              >
                📊 Side-by-Side
              </button>
              <button
                type="button"
                className={`btn ${displayMode === 'excel' ? 'btn-primary' : 'btn-outline'}`}
                style={{ padding: '0.3rem 0.6rem', fontSize: '0.78rem' }}
                onClick={() => setDisplayMode('excel')}
              >
                📊 Excel Only
              </button>
              <button
                type="button"
                className={`btn ${displayMode === 'r' ? 'btn-primary' : 'btn-outline'}`}
                style={{ padding: '0.3rem 0.6rem', fontSize: '0.78rem' }}
                onClick={() => setDisplayMode('r')}
              >
                🔷 R Only
              </button>
            </div>

            {/* Print / Export Button */}
            <button
              type="button"
              className="btn btn-outline"
              style={{ padding: '0.3rem 0.6rem', fontSize: '0.78rem' }}
              onClick={handlePrint}
              title="Print or save as PDF for offline study"
            >
              🖨 Print / PDF
            </button>
          </div>
        </div>

        {/* Search & Module Pills */}
        <div style={{ display: 'flex', gap: '0.75rem', flexWrap: 'wrap', alignItems: 'center' }}>
          <input
            type="search"
            placeholder="Search Excel & R functions, topics, or keywords (e.g. BINOM.DIST, qnorm, slope, p-value)..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            style={{
              flex: '1 1 280px',
              padding: '0.5rem 0.8rem',
              borderRadius: 'var(--radius-sm)',
              border: '1px solid var(--border-subtle)',
              background: 'var(--bg-surface-elevated)',
              color: 'var(--text-main)',
              fontSize: '0.88rem',
            }}
          />

          <div style={{ display: 'flex', gap: '0.3rem', flexWrap: 'wrap' }}>
            {[
              { id: 'all', label: 'All Modules (50)' },
              { id: 'module_1', label: 'M1 Descriptive & Sets' },
              { id: 'module_2', label: 'M2 Discrete RVs' },
              { id: 'module_3', label: 'M3 Continuous' },
              { id: 'module_4', label: 'M4 CLT & Estimation' },
              { id: 'module_5', label: 'M5 Testing' },
              { id: 'module_6', label: 'M6 Regression' },
            ].map((m) => (
              <button
                key={m.id}
                type="button"
                className={`btn ${selectedModule === m.id ? 'btn-primary' : 'btn-outline'}`}
                style={{ padding: '0.25rem 0.55rem', fontSize: '0.76rem' }}
                onClick={() => setSelectedModule(m.id)}
              >
                {m.label}
              </button>
            ))}
          </div>
        </div>
      </div>

      {/* Recipes List */}
      <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
        {filteredRecipes.length === 0 ? (
          <div className="card" style={{ textAlign: 'center', padding: '2rem', color: 'var(--text-muted)' }}>
            No software recipes match "{searchQuery}" in {selectedModule === 'all' ? 'any module' : selectedModule}.
          </div>
        ) : (
          filteredRecipes.map((recipe) => {
            const isExpanded = expandedId === recipe.id;
            return (
              <div
                key={recipe.id}
                className="card cheatsheet-recipe-card"
                data-testid={`recipe-card-${recipe.id}`}
                style={{
                  background: 'var(--bg-surface)',
                  border: isExpanded ? '1px solid var(--accent-cyan)' : '1px solid var(--border-subtle)',
                  borderRadius: 'var(--radius-md)',
                  padding: '1rem 1.25rem',
                }}
              >
                {/* Recipe Header */}
                <div
                  onClick={() => setExpandedId(isExpanded ? null : recipe.id)}
                  style={{
                    cursor: 'pointer',
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'flex-start',
                    gap: '1rem',
                  }}
                >
                  <div style={{ flex: 1 }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '0.35rem', flexWrap: 'wrap' }}>
                      <span className="badge badge-cyan" style={{ fontSize: '0.72rem' }}>
                        {MODULE_NAMES[recipe.moduleId] || recipe.moduleId}
                      </span>
                      <h4 style={{ fontSize: '1.1rem', fontWeight: 600, margin: 0, color: 'var(--text-main)' }}>
                        {recipe.title}
                      </h4>
                    </div>
                    <p style={{ color: 'var(--text-muted)', fontSize: '0.88rem', margin: '0.15rem 0' }}>
                      {recipe.statisticalTask}
                    </p>
                  </div>

                  <button
                    type="button"
                    className="btn btn-outline no-print"
                    style={{ fontSize: '0.75rem', padding: '0.25rem 0.5rem', flexShrink: 0 }}
                  >
                    {isExpanded ? 'Collapse ▲' : 'Details ▼'}
                  </button>
                </div>

                {/* Software Functions Comparison Grid */}
                <div
                  style={{
                    display: 'grid',
                    gridTemplateColumns: displayMode === 'side-by-side' ? 'repeat(auto-fit, minmax(320px, 1fr))' : '1fr',
                    gap: '1rem',
                    marginTop: '0.85rem',
                  }}
                >
                  {/* Excel Box */}
                  {(displayMode === 'side-by-side' || displayMode === 'excel') && (
                    <div
                      style={{
                        padding: '0.85rem',
                        background: 'rgba(16, 185, 129, 0.04)',
                        border: '1px solid rgba(16, 185, 129, 0.25)',
                        borderRadius: 'var(--radius-sm)',
                        display: 'flex',
                        flexDirection: 'column',
                        gap: '0.5rem',
                      }}
                    >
                      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem' }}>
                          <span style={{ fontSize: '0.8rem', fontWeight: 600, color: '#34d399' }}>
                            📊 Excel: <code>{recipe.excel.functionName}</code>
                          </span>
                        </div>
                        <button
                          type="button"
                          className="btn btn-outline no-print"
                          style={{ padding: '0.15rem 0.45rem', fontSize: '0.72rem' }}
                          onClick={() => handleCopy(`excel-${recipe.id}`, recipe.excel.genericSyntax)}
                        >
                          {copiedKey === `excel-${recipe.id}` ? '✓ Copied' : 'Copy'}
                        </button>
                      </div>

                      <div
                        style={{
                          fontFamily: 'var(--font-mono, monospace)',
                          fontSize: '0.85rem',
                          color: '#34d399',
                          background: 'rgba(0, 0, 0, 0.35)',
                          padding: '0.45rem 0.65rem',
                          borderRadius: '4px',
                          wordBreak: 'break-all',
                        }}
                      >
                        {recipe.excel.genericSyntax}
                      </div>

                      {recipe.excel.exampleFormula && (
                        <div style={{ fontSize: '0.78rem', color: 'var(--text-muted)' }}>
                          <strong>Worked Example:</strong> <code>{recipe.excel.exampleFormula}</code>
                        </div>
                      )}

                      {/* Expandable arguments and pitfalls */}
                      {isExpanded && (
                        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem', marginTop: '0.4rem', borderTop: '1px solid rgba(16, 185, 129, 0.2)', paddingTop: '0.5rem' }}>
                          <div style={{ fontSize: '0.78rem', fontWeight: 600, color: '#34d399' }}>
                            Parameters &amp; Arguments:
                          </div>
                          <ul style={{ margin: 0, paddingLeft: '1.1rem', fontSize: '0.78rem', color: 'var(--text-main)', lineHeight: 1.4 }}>
                            {recipe.excel.argumentDescriptions.map((arg, idx) => (
                              <li key={idx}>
                                <code>{arg.name}</code>: {arg.description}
                              </li>
                            ))}
                          </ul>

                          {recipe.excel.commonPitfalls && (
                            <div
                              style={{
                                background: 'rgba(239, 68, 68, 0.08)',
                                border: '1px solid rgba(239, 68, 68, 0.25)',
                                borderRadius: '4px',
                                padding: '0.4rem 0.6rem',
                                fontSize: '0.76rem',
                                color: '#f87171',
                              }}
                            >
                              <strong>⚠️ Common Trap:</strong> {recipe.excel.commonPitfalls}
                            </div>
                          )}

                          {recipe.excel.notes && (
                            <div style={{ fontSize: '0.76rem', color: 'var(--text-muted)' }}>
                              💡 {recipe.excel.notes}
                            </div>
                          )}
                        </div>
                      )}
                    </div>
                  )}

                  {/* R Box */}
                  {(displayMode === 'side-by-side' || displayMode === 'r') && (
                    <div
                      style={{
                        padding: '0.85rem',
                        background: 'rgba(56, 189, 248, 0.04)',
                        border: '1px solid rgba(56, 189, 248, 0.25)',
                        borderRadius: 'var(--radius-sm)',
                        display: 'flex',
                        flexDirection: 'column',
                        gap: '0.5rem',
                      }}
                    >
                      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem' }}>
                          <span style={{ fontSize: '0.8rem', fontWeight: 600, color: '#38bdf8' }}>
                            🔷 R: <code>{recipe.r.functionName}</code>
                          </span>
                        </div>
                        <button
                          type="button"
                          className="btn btn-outline no-print"
                          style={{ padding: '0.15rem 0.45rem', fontSize: '0.72rem' }}
                          onClick={() => handleCopy(`r-${recipe.id}`, recipe.r.genericSyntax)}
                        >
                          {copiedKey === `r-${recipe.id}` ? '✓ Copied' : 'Copy'}
                        </button>
                      </div>

                      <div
                        style={{
                          fontFamily: 'var(--font-mono, monospace)',
                          fontSize: '0.85rem',
                          color: '#38bdf8',
                          background: 'rgba(0, 0, 0, 0.35)',
                          padding: '0.45rem 0.65rem',
                          borderRadius: '4px',
                          wordBreak: 'break-all',
                        }}
                      >
                        {recipe.r.genericSyntax}
                      </div>

                      {recipe.r.exampleFormula && (
                        <div style={{ fontSize: '0.78rem', color: 'var(--text-muted)' }}>
                          <strong>Worked Example:</strong> <code>{recipe.r.exampleFormula}</code>
                        </div>
                      )}

                      {/* Expandable arguments and pitfalls */}
                      {isExpanded && (
                        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem', marginTop: '0.4rem', borderTop: '1px solid rgba(56, 189, 248, 0.2)', paddingTop: '0.5rem' }}>
                          <div style={{ fontSize: '0.78rem', fontWeight: 600, color: '#38bdf8' }}>
                            Parameters &amp; Arguments:
                          </div>
                          <ul style={{ margin: 0, paddingLeft: '1.1rem', fontSize: '0.78rem', color: 'var(--text-main)', lineHeight: 1.4 }}>
                            {recipe.r.argumentDescriptions.map((arg, idx) => (
                              <li key={idx}>
                                <code>{arg.name}</code>: {arg.description}
                              </li>
                            ))}
                          </ul>

                          {recipe.r.commonPitfalls && (
                            <div
                              style={{
                                background: 'rgba(239, 68, 68, 0.08)',
                                border: '1px solid rgba(239, 68, 68, 0.25)',
                                borderRadius: '4px',
                                padding: '0.4rem 0.6rem',
                                fontSize: '0.76rem',
                                color: '#f87171',
                              }}
                            >
                              <strong>⚠️ Common Trap:</strong> {recipe.r.commonPitfalls}
                            </div>
                          )}

                          {recipe.r.notes && (
                            <div style={{ fontSize: '0.76rem', color: 'var(--text-muted)' }}>
                              💡 {recipe.r.notes}
                            </div>
                          )}
                        </div>
                      )}
                    </div>
                  )}
                </div>
              </div>
            );
          })
        )}
      </div>
    </div>
  );
};

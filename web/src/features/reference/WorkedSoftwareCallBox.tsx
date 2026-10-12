import React, { useState, useEffect } from 'react';
import { generateWorkedSoftwareCalls, WorkedSoftwareOutput, InterpolationInput } from './softwareWorkedCalls';

interface WorkedSoftwareCallBoxProps extends InterpolationInput {
  compact?: boolean;
  allowSideBySide?: boolean;
}

export const WorkedSoftwareCallBox: React.FC<WorkedSoftwareCallBoxProps> = ({
  recipe,
  lookup,
  parameters,
  expectedAnswer,
  stageLabel,
  stageNumber,
  compact = false,
  allowSideBySide = true,
}) => {
  const [activeTab, setActiveTab] = useState<'excel' | 'r'>('excel');
  const [showSideBySide, setShowSideBySide] = useState<boolean>(false);
  const [copiedSoftware, setCopiedSoftware] = useState<'excel' | 'r' | null>(null);

  // Read preferred tool from localStorage
  useEffect(() => {
    try {
      const saved = localStorage.getItem('quant_preferred_software_tool');
      if (saved === 'r' || saved === 'excel') {
        setActiveTab(saved);
      }
    } catch {
      // Ignore localStorage access errors
    }
  }, []);

  const handleSelectTab = (tool: 'excel' | 'r') => {
    setActiveTab(tool);
    try {
      localStorage.setItem('quant_preferred_software_tool', tool);
    } catch {
      // Ignore localStorage access errors
    }
  };

  const output: WorkedSoftwareOutput | null = generateWorkedSoftwareCalls({
    recipe,
    lookup,
    parameters,
    expectedAnswer,
    stageLabel,
    stageNumber,
  });

  if (!output) return null;

  const handleCopy = async (software: 'excel' | 'r', textToCopy: string) => {
    try {
      await navigator.clipboard.writeText(textToCopy);
      setCopiedSoftware(software);
      setTimeout(() => setCopiedSoftware(null), 2000);
    } catch {
      // Fallback
      const ta = document.createElement('textarea');
      ta.value = textToCopy;
      document.body.appendChild(ta);
      ta.select();
      document.execCommand('copy');
      document.body.removeChild(ta);
      setCopiedSoftware(software);
      setTimeout(() => setCopiedSoftware(null), 2000);
    }
  };

  return (
    <div
      className="worked-software-call-box"
      data-testid="worked-software-call-box"
      style={{
        marginTop: '0.85rem',
        padding: compact ? '0.75rem' : '1rem',
        background: 'var(--bg-surface-elevated, #1e293b)',
        border: '1px solid var(--border-subtle, rgba(255, 255, 255, 0.1))',
        borderRadius: 'var(--radius-md, 8px)',
      }}
    >
      {/* Header bar */}
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          flexWrap: 'wrap',
          gap: '0.5rem',
          marginBottom: '0.65rem',
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
          <span style={{ fontSize: '0.95rem' }}>💻</span>
          <span style={{ fontWeight: 600, fontSize: '0.88rem', color: 'var(--text-main, #f8fafc)' }}>
            Worked Software Call
          </span>
          <span
            className="badge badge-cyan"
            style={{ fontSize: '0.72rem', padding: '0.15rem 0.45rem' }}
          >
            {output.recipeTitle}
          </span>
        </div>

        {/* View switcher buttons */}
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.35rem' }}>
          {allowSideBySide && (
            <button
              type="button"
              className={`btn ${showSideBySide ? 'btn-primary' : 'btn-outline'}`}
              style={{ padding: '0.2rem 0.5rem', fontSize: '0.75rem' }}
              onClick={() => setShowSideBySide(!showSideBySide)}
              title="Toggle side-by-side comparison"
            >
              Side-by-Side
            </button>
          )}

          {!showSideBySide && (
            <div role="tablist" aria-label="Worked Software Tool" style={{ display: 'flex', gap: '0.25rem' }}>
              <button
                type="button"
                role="tab"
                id="worked-tab-excel"
                aria-selected={activeTab === 'excel'}
                aria-controls="worked-panel-excel"
                className={`software-tab ${activeTab === 'excel' ? 'active-excel' : ''}`}
                style={{ padding: '0.2rem 0.6rem', fontSize: '0.78rem' }}
                onClick={() => handleSelectTab('excel')}
              >
                📊 Excel
              </button>
              <button
                type="button"
                role="tab"
                id="worked-tab-r"
                aria-selected={activeTab === 'r'}
                aria-controls="worked-panel-r"
                className={`software-tab ${activeTab === 'r' ? 'active-r' : ''}`}
                style={{ padding: '0.2rem 0.6rem', fontSize: '0.78rem' }}
                onClick={() => handleSelectTab('r')}
              >
                🔷 R
              </button>
            </div>
          )}
        </div>
      </div>

      {/* Side-by-side or Single tab render */}
      {showSideBySide ? (
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fit, minmax(260px, 1fr))',
            gap: '0.75rem',
          }}
        >
          {/* Excel card */}
          <div
            style={{
              padding: '0.75rem',
              background: 'rgba(16, 185, 129, 0.05)',
              border: '1px solid rgba(16, 185, 129, 0.25)',
              borderRadius: 'var(--radius-sm, 6px)',
            }}
          >
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.35rem' }}>
              <span style={{ fontSize: '0.75rem', fontWeight: 600, color: '#34d399' }}>📊 Excel Call</span>
              <button
                type="button"
                className="btn btn-outline"
                style={{ padding: '0.1rem 0.4rem', fontSize: '0.7rem' }}
                onClick={() => handleCopy('excel', output.excel.call)}
              >
                {copiedSoftware === 'excel' ? '✓ Copied' : 'Copy'}
              </button>
            </div>
            <div
              style={{
                fontFamily: 'var(--font-mono, monospace)',
                fontSize: '0.85rem',
                color: '#34d399',
                wordBreak: 'break-all',
                background: 'rgba(0, 0, 0, 0.3)',
                padding: '0.4rem 0.6rem',
                borderRadius: '4px',
              }}
            >
              {output.excel.call}
              {output.excel.result && (
                <span style={{ color: 'var(--text-muted, #94a3b8)', marginLeft: '0.5rem' }}>
                  &rarr; {output.excel.result}
                </span>
              )}
            </div>
            <div style={{ fontSize: '0.75rem', color: 'var(--text-muted, #94a3b8)', marginTop: '0.35rem' }}>
              {output.excel.description}
            </div>
          </div>

          {/* R card */}
          <div
            style={{
              padding: '0.75rem',
              background: 'rgba(56, 189, 248, 0.05)',
              border: '1px solid rgba(56, 189, 248, 0.25)',
              borderRadius: 'var(--radius-sm, 6px)',
            }}
          >
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.35rem' }}>
              <span style={{ fontSize: '0.75rem', fontWeight: 600, color: '#38bdf8' }}>🔷 R Call</span>
              <button
                type="button"
                className="btn btn-outline"
                style={{ padding: '0.1rem 0.4rem', fontSize: '0.7rem' }}
                onClick={() => handleCopy('r', output.r.call)}
              >
                {copiedSoftware === 'r' ? '✓ Copied' : 'Copy'}
              </button>
            </div>
            <div
              style={{
                fontFamily: 'var(--font-mono, monospace)',
                fontSize: '0.85rem',
                color: '#38bdf8',
                wordBreak: 'break-all',
                background: 'rgba(0, 0, 0, 0.3)',
                padding: '0.4rem 0.6rem',
                borderRadius: '4px',
              }}
            >
              {output.r.call}
              {output.r.result && (
                <span style={{ color: 'var(--text-muted, #94a3b8)', marginLeft: '0.5rem' }}>
                  &rarr; {output.r.result}
                </span>
              )}
            </div>
            <div style={{ fontSize: '0.75rem', color: 'var(--text-muted, #94a3b8)', marginTop: '0.35rem' }}>
              {output.r.description}
            </div>
          </div>
        </div>
      ) : activeTab === 'excel' ? (
        <div
          role="tabpanel"
          id="worked-panel-excel"
          aria-labelledby="worked-tab-excel"
          style={{
            padding: '0.75rem',
            background: 'rgba(16, 185, 129, 0.05)',
            border: '1px solid rgba(16, 185, 129, 0.25)',
            borderRadius: 'var(--radius-sm, 6px)',
          }}
        >
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.35rem' }}>
            <span style={{ fontSize: '0.78rem', fontWeight: 600, color: '#34d399' }}>
              Microsoft Excel Evaluated Formula
            </span>
            <button
              type="button"
              className="btn btn-outline"
              style={{ padding: '0.15rem 0.5rem', fontSize: '0.75rem' }}
              onClick={() => handleCopy('excel', output.excel.call)}
            >
              {copiedSoftware === 'excel' ? '✓ Copied to Clipboard' : '📋 Copy Formula'}
            </button>
          </div>
          <div
            style={{
              fontFamily: 'var(--font-mono, monospace)',
              fontSize: '0.9rem',
              color: '#34d399',
              wordBreak: 'break-all',
              background: 'rgba(0, 0, 0, 0.35)',
              padding: '0.5rem 0.75rem',
              borderRadius: '4px',
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
              flexWrap: 'wrap',
              gap: '0.5rem',
            }}
          >
            <span>{output.excel.call}</span>
            {output.excel.result && (
              <span
                style={{
                  background: 'rgba(16, 185, 129, 0.2)',
                  border: '1px solid rgba(16, 185, 129, 0.4)',
                  padding: '0.15rem 0.5rem',
                  borderRadius: '4px',
                  color: '#6ee7b7',
                  fontSize: '0.85rem',
                  fontWeight: 600,
                }}
              >
                = {output.excel.result}
              </span>
            )}
          </div>
          <div style={{ fontSize: '0.78rem', color: 'var(--text-muted, #94a3b8)', marginTop: '0.4rem' }}>
            {output.excel.description}
          </div>
        </div>
      ) : (
        <div
          role="tabpanel"
          id="worked-panel-r"
          aria-labelledby="worked-tab-r"
          style={{
            padding: '0.75rem',
            background: 'rgba(56, 189, 248, 0.05)',
            border: '1px solid rgba(56, 189, 248, 0.25)',
            borderRadius: 'var(--radius-sm, 6px)',
          }}
        >
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.35rem' }}>
            <span style={{ fontSize: '0.78rem', fontWeight: 600, color: '#38bdf8' }}>
              R / RStudio Evaluated Command
            </span>
            <button
              type="button"
              className="btn btn-outline"
              style={{ padding: '0.15rem 0.5rem', fontSize: '0.75rem' }}
              onClick={() => handleCopy('r', output.r.call)}
            >
              {copiedSoftware === 'r' ? '✓ Copied to Clipboard' : '📋 Copy Code'}
            </button>
          </div>
          <div
            style={{
              fontFamily: 'var(--font-mono, monospace)',
              fontSize: '0.9rem',
              color: '#38bdf8',
              wordBreak: 'break-all',
              background: 'rgba(0, 0, 0, 0.35)',
              padding: '0.5rem 0.75rem',
              borderRadius: '4px',
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
              flexWrap: 'wrap',
              gap: '0.5rem',
            }}
          >
            <span>{output.r.call}</span>
            {output.r.result && (
              <span
                style={{
                  background: 'rgba(56, 189, 248, 0.2)',
                  border: '1px solid rgba(56, 189, 248, 0.4)',
                  padding: '0.15rem 0.5rem',
                  borderRadius: '4px',
                  color: '#7dd3fc',
                  fontSize: '0.85rem',
                  fontWeight: 600,
                }}
              >
                [1] {output.r.result}
              </span>
            )}
          </div>
          <div style={{ fontSize: '0.78rem', color: 'var(--text-muted, #94a3b8)', marginTop: '0.4rem' }}>
            {output.r.description}
          </div>
        </div>
      )}
    </div>
  );
};

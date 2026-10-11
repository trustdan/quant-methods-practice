import React, { useState } from 'react';
import { NormalVisual } from './NormalVisual';
import { UniformVisual } from './UniformVisual';
import { ExponentialVisual } from './ExponentialVisual';
import { DiscreteVisual } from './DiscreteVisual';
import { VennVisual } from './VennVisual';

export type VisualTab = 'normal' | 'uniform' | 'exponential' | 'discrete' | 'venn';

interface VisualsExplorerProps {
  initialTab?: VisualTab;
  onClose?: () => void;
}

export const VisualsExplorer: React.FC<VisualsExplorerProps> = ({
  initialTab = 'normal',
  onClose,
}) => {
  const [activeTab, setActiveTab] = useState<VisualTab>(initialTab);

  const tabs: { id: VisualTab; label: string; badge: string }[] = [
    { id: 'normal', label: 'Normal Distribution', badge: 'Continuous N(μ, σ²)' },
    { id: 'uniform', label: 'Uniform Distribution', badge: 'Continuous U(a, b)' },
    { id: 'exponential', label: 'Exponential Distribution', badge: 'Continuous Exp(λ)' },
    { id: 'discrete', label: 'Discrete PMF / CDF', badge: 'Binomial & Poisson' },
    { id: 'venn', label: 'Venn Diagram', badge: 'Set Probability' },
  ];

  return (
    <div className="card visuals-card" style={{ maxWidth: '1000px', width: '100%', margin: '0 auto', minWidth: 0, boxSizing: 'border-box' }}>
      <div className="card-header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '0.5rem' }}>
        <div>
          <h2 className="card-title" style={{ display: 'flex', alignItems: 'center', gap: '0.6rem', flexWrap: 'wrap' }}>
            Mathematical Visuals &amp; Distribution Lab
            <span className="badge badge-purple">Module 1 - 3 Figures</span>
          </h2>
          <p style={{ margin: '0.2rem 0 0', fontSize: '0.85rem', color: 'var(--text-muted)' }}>
            Interactive, keyboard-operable SVG statistical visualizations reconciling exact Go math engine results with spreadsheet formulas.
          </p>
        </div>

        {onClose && (
          <button
            type="button"
            className="btn btn-outline btn-sm"
            onClick={onClose}
            id="btn-close-visuals"
          >
            Back to Practice (Esc)
          </button>
        )}
      </div>

      {/* Visual Sub-tabs */}
      <div
        role="tablist"
        aria-label="Distribution Visual Categories"
        style={{ display: 'flex', gap: '0.5rem', overflowX: 'auto', maxWidth: '100%', minWidth: 0, paddingBottom: '0.5rem', marginBottom: '1.25rem', borderBottom: '1px solid var(--border-subtle)' }}
      >
        {tabs.map((tab) => (
          <button
            key={tab.id}
            role="tab"
            aria-selected={activeTab === tab.id}
            className={`btn btn-sm ${activeTab === tab.id ? 'btn-primary' : 'btn-outline'}`}
            onClick={() => setActiveTab(tab.id)}
            id={`tab-visual-${tab.id}`}
            style={{ display: 'flex', alignItems: 'center', gap: '0.4rem', whiteSpace: 'nowrap' }}
          >
            <span>{tab.label}</span>
          </button>
        ))}
      </div>

      {/* Active Visual Sub-view */}
      <div>
        {activeTab === 'normal' && <NormalVisual />}
        {activeTab === 'uniform' && <UniformVisual />}
        {activeTab === 'exponential' && <ExponentialVisual />}
        {activeTab === 'discrete' && <DiscreteVisual />}
        {activeTab === 'venn' && <VennVisual />}
      </div>
    </div>
  );
};

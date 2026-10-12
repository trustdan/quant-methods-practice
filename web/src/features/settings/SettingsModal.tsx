import React, { useState, useEffect } from 'react';
import { ProviderSettings } from '../providers/ProviderSettings';
import { setTourDismissed } from '../tour/ProductTourModal';

export interface SessionConfig {
  question_count: number;
  module_ids: string[];
  intensity: string;
  seed?: number;
  tour_dismissed?: boolean;
}

interface SettingsModalProps {
  isOpen: boolean;
  onClose: () => void;
  onStartSession: (config: SessionConfig) => Promise<void>;
  currentSettings?: SessionConfig;
  bankTotalQuestions?: number;
}

export const SettingsModal: React.FC<SettingsModalProps> = ({
  isOpen,
  onClose,
  onStartSession,
  currentSettings,
  bankTotalQuestions,
}) => {
  const [activeTab, setActiveTab] = useState<'session' | 'providers'>('session');
  const [questionCount, setQuestionCount] = useState<number>(currentSettings?.question_count || 10);
  const [selectedModules, setSelectedModules] = useState<string[]>(currentSettings?.module_ids || []);
  const [intensity, setIntensity] = useState<string>(currentSettings?.intensity || 'standard');
  const [seedInput, setSeedInput] = useState<string>(currentSettings?.seed ? String(currentSettings.seed) : '');
  const [tourDismissed, setTourDismissedState] = useState<boolean>(currentSettings?.tour_dismissed ?? false);
  const [isSaving, setIsSaving] = useState(false);

  useEffect(() => {
    if (currentSettings) {
      setQuestionCount(currentSettings.question_count || 10);
      setSelectedModules(currentSettings.module_ids || []);
      setIntensity(currentSettings.intensity || 'standard');
      setSeedInput(currentSettings.seed ? String(currentSettings.seed) : '');
      setTourDismissedState(currentSettings.tour_dismissed ?? false);
    }
  }, [currentSettings, isOpen]);

  useEffect(() => {
    if (!isOpen) return;
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        e.preventDefault();
        onClose();
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, onClose]);

  if (!isOpen) return null;

  const toggleModule = (modId: string) => {
    setSelectedModules((prev) =>
      prev.includes(modId) ? prev.filter((m) => m !== modId) : [...prev, modId]
    );
  };

  const handleStart = async () => {
    setIsSaving(true);
    try {
      const cfg: SessionConfig = {
        question_count: questionCount,
        module_ids: selectedModules,
        intensity,
        seed: seedInput.trim() ? parseInt(seedInput.trim(), 10) : undefined,
      };
      await onStartSession(cfg);
      onClose();
    } finally {
      setIsSaving(false);
    }
  };

  const handleToggleTourReset = async () => {
    const nextVal = !tourDismissed;
    setTourDismissedState(nextVal);
    setTourDismissed(nextVal);
    try {
      await fetch('/api/settings', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          question_count: questionCount,
          module_ids: selectedModules,
          intensity,
          tour_dismissed: nextVal,
        }),
      });
    } catch {
      // offline fallback
    }
  };

  const handleSaveDefaults = async () => {
    setIsSaving(true);
    try {
      await fetch('/api/settings', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          question_count: questionCount,
          module_ids: selectedModules,
          intensity,
          tour_dismissed: tourDismissed,
        }),
      });
    } catch {
      // offline fallback
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <div
      className="modal-backdrop"
      style={{
        position: 'fixed',
        top: 0,
        left: 0,
        right: 0,
        bottom: 0,
        backgroundColor: 'rgba(0, 0, 0, 0.65)',
        backdropFilter: 'blur(4px)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        zIndex: 1000,
        padding: '1rem',
      }}
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
    >
      <div
        className="card"
        role="dialog"
        aria-modal="true"
        aria-labelledby="settings-title"
        style={{
          width: '100%',
          maxWidth: '620px',
          maxHeight: '90vh',
          overflowY: 'auto',
          background: 'var(--bg-surface)',
          border: '1px solid var(--border-subtle)',
          borderRadius: 'var(--radius-lg)',
          boxShadow: 'var(--shadow-xl)',
          padding: '1.5rem',
          display: 'flex',
          flexDirection: 'column',
          gap: '1.25rem',
        }}
      >
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <h2 id="settings-title" style={{ fontSize: '1.25rem', fontWeight: 700, margin: 0 }}>
            Session Settings & Module Picker
          </h2>
          <button className="btn btn-outline" onClick={onClose} aria-label="Close settings">
            ✕
          </button>
        </div>

        {/* Tab Navigation */}
        <div style={{ display: 'flex', gap: '0.5rem', borderBottom: '1px solid var(--border-subtle)', paddingBottom: '0.5rem' }}>
          <button
            type="button"
            id="tab-session-settings"
            className={`btn ${activeTab === 'session' ? 'btn-primary' : 'btn-ghost'}`}
            onClick={() => setActiveTab('session')}
            style={{ fontSize: '0.85rem', padding: '0.4rem 0.75rem' }}
          >
            📚 Session & Curriculum
          </button>
          <button
            type="button"
            id="tab-provider-settings"
            className={`btn ${activeTab === 'providers' ? 'btn-primary' : 'btn-ghost'}`}
            onClick={() => setActiveTab('providers')}
            style={{ fontSize: '0.85rem', padding: '0.4rem 0.75rem' }}
          >
            🔑 AI Providers & Credentials
          </button>
        </div>

        {activeTab === 'providers' ? (
          <ProviderSettings />
        ) : (
          <>

        {/* Question Count Selector */}
        <div>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.4rem', flexWrap: 'wrap', gap: '0.5rem' }}>
            <label style={{ fontWeight: 600, fontSize: '0.9rem', margin: 0 }}>
              Question Count:
            </label>
            {bankTotalQuestions && bankTotalQuestions > 0 ? (
              <span style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>
                🎲 Randomly sampled from <strong>{bankTotalQuestions}</strong> approved questions
              </span>
            ) : null}
          </div>
          <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap' }}>
            {[5, 10, 15, 20].map((cnt) => (
              <button
                key={cnt}
                type="button"
                className={`btn ${questionCount === cnt ? 'btn-primary' : 'btn-outline'}`}
                onClick={() => setQuestionCount(cnt)}
                style={{ flex: 1, minWidth: '70px', padding: '0.4rem' }}
              >
                {cnt} Questions
              </button>
            ))}
            {bankTotalQuestions && bankTotalQuestions > 20 && (
              <button
                type="button"
                className={`btn ${questionCount === bankTotalQuestions ? 'btn-primary' : 'btn-outline'}`}
                onClick={() => setQuestionCount(bankTotalQuestions)}
                style={{ flex: 1, minWidth: '90px', padding: '0.4rem' }}
                title="Practice all approved questions in the bank"
              >
                All ({bankTotalQuestions})
              </button>
            )}
          </div>
        </div>

        {/* Module Filters */}
        <div>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.4rem', flexWrap: 'wrap', gap: '0.5rem' }}>
            <label style={{ fontWeight: 600, fontSize: '0.9rem', margin: 0 }}>
              Curriculum Modules:
            </label>
            <div style={{ display: 'flex', gap: '0.4rem' }}>
              <button
                type="button"
                className="btn btn-ghost"
                onClick={() => setSelectedModules(['module_1', 'module_2', 'module_3', 'module_4', 'module_5', 'module_6'])}
                style={{ fontSize: '0.75rem', padding: '0.15rem 0.5rem' }}
              >
                Select All
              </button>
              <button
                type="button"
                className="btn btn-ghost"
                onClick={() => setSelectedModules([])}
                style={{ fontSize: '0.75rem', padding: '0.15rem 0.5rem' }}
              >
                Clear All
              </button>
            </div>
          </div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem', maxHeight: '280px', overflowY: 'auto' }}>
            {[
              { id: 'module_1', num: 1, title: 'Probability Foundations & Sets', desc: 'Axioms, Complement, Union, Intersection, Conditional, Independence' },
              { id: 'module_2', num: 2, title: 'Discrete Random Variables & Distributions', desc: 'PMF tables, E[X], Var(X), Binomial, Poisson, Linear combinations' },
              { id: 'module_3', num: 3, title: 'Continuous Distributions', desc: 'Continuous vs Discrete, Uniform, Normal curve, Empirical rule, Z-scores, Exponential' },
              { id: 'module_4', num: 4, title: 'Sampling Distributions & Estimation', desc: 'Central Limit Theorem (CLT), Standard error, Sample size planning, Confidence intervals' },
              { id: 'module_5', num: 5, title: 'Hypothesis Testing & Decisions', desc: 'Null/Alternative formulation, Type I & II errors, One/Two-sample tests, p-values' },
              { id: 'module_6', num: 6, title: 'Regression Analysis & Diagnostics', desc: 'OLS equation, Prediction, ANOVA table, R², Adjusted R², Significance F, Slope CI' },
            ].map((mod) => (
              <label
                key={mod.id}
                style={{
                  display: 'flex',
                  alignItems: 'flex-start',
                  gap: '0.5rem',
                  padding: '0.5rem 0.75rem',
                  borderRadius: 'var(--radius-sm)',
                  background: 'var(--bg-surface-elevated)',
                  cursor: 'pointer',
                }}
              >
                <input
                  type="checkbox"
                  aria-label={`Module ${mod.num}: ${mod.title}`}
                  checked={selectedModules.includes(mod.id)}
                  onChange={() => toggleModule(mod.id)}
                  style={{ marginTop: '0.2rem' }}
                />
                <div style={{ display: 'flex', flexDirection: 'column' }}>
                  <span style={{ fontSize: '0.9rem', fontWeight: 600 }}>
                    Module {mod.num}: {mod.title}
                  </span>
                  <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
                    {mod.desc}
                  </span>
                </div>
              </label>
            ))}
          </div>
          <p style={{ fontSize: '0.75rem', color: 'var(--text-muted)', margin: '0.35rem 0 0 0' }}>
            {selectedModules.length === 0
              ? 'All modules active: sampling randomly across the full approved curriculum.'
              : `Filtering active: sampling questions from ${selectedModules.length} selected module${selectedModules.length > 1 ? 's' : ''}.`}
          </p>
        </div>

        {/* Intensity Selector */}
        <div>
          <label style={{ display: 'block', fontWeight: 600, fontSize: '0.9rem', marginBottom: '0.4rem' }}>
            Drill Intensity:
          </label>
          <div style={{ display: 'flex', gap: '0.5rem' }}>
            {[
              { id: 'gentle', label: 'Gentle (Generous hints)' },
              { id: 'standard', label: 'Standard (Exam-like)' },
              { id: 'intensive', label: 'Intensive (Speed focus)' },
            ].map((item) => (
              <button
                key={item.id}
                type="button"
                className={`btn ${intensity === item.id ? 'btn-primary' : 'btn-outline'}`}
                onClick={() => setIntensity(item.id)}
                style={{ flex: 1, padding: '0.4rem', fontSize: '0.8rem' }}
              >
                {item.label}
              </button>
            ))}
          </div>
        </div>

        {/* Seed Input */}
        <div>
          <label style={{ display: 'block', fontWeight: 600, fontSize: '0.9rem', marginBottom: '0.4rem' }}>
            Random Seed (Optional):
          </label>
          <input
            type="number"
            placeholder="0 for randomized session"
            value={seedInput}
            onChange={(e) => setSeedInput(e.target.value)}
            style={{
              width: '100%',
              padding: '0.45rem 0.75rem',
              borderRadius: 'var(--radius-sm)',
              border: '1px solid var(--border-subtle)',
              background: 'var(--bg-surface-elevated)',
              color: 'var(--text-main)',
              fontSize: '0.9rem',
            }}
          />
        </div>

        {/* Product Tour Preferences */}
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            padding: '0.65rem 0.85rem',
            background: 'var(--bg-surface-elevated)',
            borderRadius: 'var(--radius-sm)',
            border: '1px solid var(--border-subtle)',
          }}
        >
          <div>
            <div style={{ fontWeight: 600, fontSize: '0.85rem' }}>Interactive Product Tour</div>
            <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
              {tourDismissed ? 'Currently suppressed on app startup' : 'Will show automatically on app startup'}
            </div>
          </div>
          <button
            type="button"
            className="btn btn-outline"
            id="btn-settings-toggle-tour"
            style={{ fontSize: '0.8rem', padding: '0.25rem 0.65rem' }}
            onClick={handleToggleTourReset}
          >
            {tourDismissed ? 'Re-enable on Startup' : 'Reset Startup Tour'}
          </button>
        </div>

        {/* Actions */}
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: '0.5rem' }}>
          <button
            type="button"
            className="btn btn-secondary"
            onClick={handleSaveDefaults}
            disabled={isSaving}
            style={{ fontSize: '0.85rem' }}
          >
            Save Defaults
          </button>
          <div style={{ display: 'flex', gap: '0.5rem' }}>
            <button type="button" className="btn btn-outline" onClick={onClose} disabled={isSaving}>
              Cancel
            </button>
            <button
              type="button"
              className="btn btn-primary"
              onClick={handleStart}
              disabled={isSaving}
            >
              {isSaving ? 'Starting...' : 'Start New Session'}
            </button>
          </div>
        </div>
        </>
        )}
      </div>
    </div>
  );
};

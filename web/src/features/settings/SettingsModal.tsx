import React, { useState, useEffect } from 'react';

export interface SessionConfig {
  question_count: number;
  module_ids: string[];
  intensity: string;
  seed?: number;
}

interface SettingsModalProps {
  isOpen: boolean;
  onClose: () => void;
  onStartSession: (config: SessionConfig) => Promise<void>;
  currentSettings?: SessionConfig;
}

export const SettingsModal: React.FC<SettingsModalProps> = ({
  isOpen,
  onClose,
  onStartSession,
  currentSettings,
}) => {
  const [questionCount, setQuestionCount] = useState<number>(currentSettings?.question_count || 10);
  const [selectedModules, setSelectedModules] = useState<string[]>(currentSettings?.module_ids || []);
  const [intensity, setIntensity] = useState<string>(currentSettings?.intensity || 'standard');
  const [seedInput, setSeedInput] = useState<string>(currentSettings?.seed ? String(currentSettings.seed) : '');
  const [isSaving, setIsSaving] = useState(false);

  useEffect(() => {
    if (currentSettings) {
      setQuestionCount(currentSettings.question_count || 10);
      setSelectedModules(currentSettings.module_ids || []);
      setIntensity(currentSettings.intensity || 'standard');
      setSeedInput(currentSettings.seed ? String(currentSettings.seed) : '');
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
          maxWidth: '540px',
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

        {/* Question Count Selector */}
        <div>
          <label style={{ display: 'block', fontWeight: 600, fontSize: '0.9rem', marginBottom: '0.4rem' }}>
            Question Count:
          </label>
          <div style={{ display: 'flex', gap: '0.5rem' }}>
            {[5, 10, 15, 20].map((cnt) => (
              <button
                key={cnt}
                type="button"
                className={`btn ${questionCount === cnt ? 'btn-primary' : 'btn-outline'}`}
                onClick={() => setQuestionCount(cnt)}
                style={{ flex: 1, padding: '0.4rem' }}
              >
                {cnt} Questions
              </button>
            ))}
          </div>
        </div>

        {/* Module Filters */}
        <div>
          <label style={{ display: 'block', fontWeight: 600, fontSize: '0.9rem', marginBottom: '0.4rem' }}>
            Curriculum Modules:
          </label>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
            <label
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: '0.5rem',
                padding: '0.5rem 0.75rem',
                borderRadius: 'var(--radius-sm)',
                background: 'var(--bg-surface-elevated)',
                cursor: 'pointer',
              }}
            >
              <input
                type="checkbox"
                checked={selectedModules.includes('module_1')}
                onChange={() => toggleModule('module_1')}
              />
              <span style={{ fontSize: '0.9rem' }}>
                <strong>Module 1</strong>: Probability Foundations & Sets (Axioms, Complement, Union, Conditional, Independence)
              </span>
            </label>
            <label
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: '0.5rem',
                padding: '0.5rem 0.75rem',
                borderRadius: 'var(--radius-sm)',
                background: 'var(--bg-surface-elevated)',
                cursor: 'pointer',
              }}
            >
              <input
                type="checkbox"
                checked={selectedModules.includes('module_2')}
                onChange={() => toggleModule('module_2')}
              />
              <span style={{ fontSize: '0.9rem' }}>
                <strong>Module 2</strong>: Discrete Random Variables & Combinations (Binomial, Poisson, Sums & Variances)
              </span>
            </label>
          </div>
          <p style={{ fontSize: '0.75rem', color: 'var(--text-muted)', margin: '0.3rem 0 0 0' }}>
            Leave both checked (or unchecked) to sample across the entire 10-question curriculum.
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
      </div>
    </div>
  );
};

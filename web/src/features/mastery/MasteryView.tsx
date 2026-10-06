import React, { useState, useEffect } from 'react';
import { MasterySummary, ConceptMastery } from '../../types/practice';

interface MasteryViewProps {
  onClose: () => void;
}

export const MasteryView: React.FC<MasteryViewProps> = ({ onClose }) => {
  const [summary, setSummary] = useState<MasterySummary | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [searchTerm, setSearchTerm] = useState<string>('');
  const [statusFilter, setStatusFilter] = useState<string>('all');

  useEffect(() => {
    let isMounted = true;
    const fetchMastery = async () => {
      try {
        setLoading(true);
        const res = await fetch('/api/mastery');
        if (!res.ok) {
          throw new Error(`Server returned ${res.status}`);
        }
        const data: MasterySummary = await res.json();
        if (isMounted) {
          setSummary(data);
          setError(null);
        }
      } catch (err) {
        if (isMounted) {
          setError((err instanceof Error && err.message) || 'Failed to load concept mastery');
        }
      } finally {
        if (isMounted) {
          setLoading(false);
        }
      }
    };

    fetchMastery();
    return () => {
      isMounted = false;
    };
  }, []);

  const formatConceptName = (id: string) => {
    return id
      .split('_')
      .map(word => word.charAt(0).toUpperCase() + word.slice(1))
      .join(' ');
  };

  const filteredConcepts = (summary?.concepts || []).filter((c: ConceptMastery) => {
    const matchesSearch = c.concept_id.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesFilter = statusFilter === 'all' || c.status === statusFilter;
    return matchesSearch && matchesFilter;
  });

  return (
    <div className="mastery-view-container" id="mastery-view">
      {/* Header */}
      <header className="mastery-header">
        <div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
            <h2 className="mastery-title">Concept Evidence &amp; Transfer</h2>
            <span className="badge badge-accent">Policy v{summary?.policy_version || 1}</span>
          </div>
          <p className="mastery-subtitle">
            Heuristic scheduling evidence: decay, reasoning transfer across problem groups, and scaffold degradation.
          </p>
        </div>
        <button
          className="btn btn-secondary"
          onClick={onClose}
          id="btn-close-mastery"
          aria-label="Return to practice"
        >
          &larr; Return to Practice <span className="kbd">Esc</span>
        </button>
      </header>

      {/* Summary Stats Grid */}
      {summary && (
        <div className="mastery-stats-grid" id="mastery-stats">
          <div className="stat-card">
            <div className="stat-label">Overall Retention Score</div>
            <div className="stat-value" style={{ color: 'var(--accent-cyan)' }}>
              {(summary.overall_score * 100).toFixed(1)}%
            </div>
            <div className="stat-meta">3-day half-life decay heuristic</div>
          </div>
          <div className="stat-card">
            <div className="stat-label">Mastered (Faded)</div>
            <div className="stat-value" style={{ color: 'var(--accent-emerald)' }}>
              {summary.total_mastered}
            </div>
            <div className="stat-meta">&ge;3 successes, &ge;80% score, &ge;2 groups</div>
          </div>
          <div className="stat-card">
            <div className="stat-label">Transferring (Intermediate)</div>
            <div className="stat-value" style={{ color: 'var(--accent-amber)' }}>
              {summary.total_transferring}
            </div>
            <div className="stat-meta">&ge;2 successes, &ge;60% score, 4-stage scaffold</div>
          </div>
          <div className="stat-card">
            <div className="stat-label">Learning (Full Guidance)</div>
            <div className="stat-value" style={{ color: 'var(--accent-indigo)' }}>
              {summary.total_learning}
            </div>
            <div className="stat-meta">Active practice with causal hints</div>
          </div>
          <div className="stat-card">
            <div className="stat-label">Unseen Concepts</div>
            <div className="stat-value" style={{ color: 'var(--text-muted)' }}>
              {summary.total_new}
            </div>
            <div className="stat-meta">Prioritized for upcoming practice</div>
          </div>
        </div>
      )}

      {/* Filter and Search Bar */}
      <div className="mastery-controls">
        <div className="mastery-filter-pills">
          {['all', 'mastered', 'transferring', 'learning', 'new'].map(filterKey => (
            <button
              key={filterKey}
              className={`pill-btn ${statusFilter === filterKey ? 'active' : ''}`}
              onClick={() => setStatusFilter(filterKey)}
              id={`filter-${filterKey}`}
            >
              {filterKey.charAt(0).toUpperCase() + filterKey.slice(1)}
            </button>
          ))}
        </div>
        <div className="mastery-search">
          <input
            type="text"
            className="input-field"
            placeholder="Search concepts..."
            value={searchTerm}
            onChange={e => setSearchTerm(e.target.value)}
            id="mastery-search-input"
            aria-label="Search concepts"
          />
        </div>
      </div>

      {/* Content Area */}
      {loading ? (
        <div className="loading-state">Computing read-time decay and transfer records...</div>
      ) : error ? (
        <div className="error-notice">
          <strong>Notice:</strong> {error}
        </div>
      ) : filteredConcepts.length === 0 ? (
        <div className="empty-state">
          No concepts match your filter. Practice drills to accumulate independent evidence.
        </div>
      ) : (
        <div className="concept-cards-grid" id="concept-cards-container">
          {filteredConcepts.map(c => {
            const statusColor =
              c.status === 'mastered'
                ? 'var(--accent-emerald)'
                : c.status === 'transferring'
                ? 'var(--accent-amber)'
                : c.status === 'learning'
                ? 'var(--accent-cyan)'
                : 'var(--text-subtle)';

            const scaffoldLabel =
              c.scaffold_level === 'faded'
                ? 'Faded (2 stages)'
                : c.scaffold_level === 'intermediate'
                ? 'Intermediate (4 stages)'
                : 'Full (7 stages)';

            return (
              <div key={c.concept_id} className="concept-card" id={`concept-${c.concept_id}`}>
                <div className="concept-card-header">
                  <div>
                    <h3 className="concept-name">{formatConceptName(c.concept_id)}</h3>
                    <code className="concept-id-code">{c.concept_id}</code>
                  </div>
                  <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
                    <span
                      className="status-pill"
                      style={{
                        borderColor: statusColor,
                        color: statusColor,
                      }}
                    >
                      {c.status.toUpperCase()}
                    </span>
                    <span className="scaffold-badge">{scaffoldLabel}</span>
                  </div>
                </div>

                {/* Score Progress */}
                <div className="concept-score-section">
                  <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.25rem' }}>
                    <span style={{ fontSize: '0.85rem', color: 'var(--text-muted)' }}>
                      Decayed Score: <strong>{(c.decayed_score * 100).toFixed(1)}%</strong>
                    </span>
                    <span style={{ fontSize: '0.85rem', color: 'var(--text-subtle)' }}>
                      Base: {(c.base_score * 100).toFixed(1)}% | Retention: {(c.retention_factor * 100).toFixed(0)}%
                    </span>
                  </div>
                  <div className="progress-bar-bg">
                    <div
                      className="progress-bar-fill"
                      style={{
                        width: `${Math.min(100, Math.max(0, c.decayed_score * 100))}%`,
                        backgroundColor: statusColor,
                      }}
                    />
                  </div>
                </div>

                {/* Evidence Metrics */}
                <div className="concept-metrics-row">
                  <div className="metric-item">
                    <div className="metric-label">Independent Success / Err</div>
                    <div className="metric-value">
                      <span style={{ color: 'var(--accent-emerald)' }}>{c.independent_successes}</span> /{' '}
                      <span style={{ color: c.independent_errors > 0 ? 'var(--accent-rose)' : 'var(--text-muted)' }}>
                        {c.independent_errors}
                      </span>
                    </div>
                  </div>
                  <div className="metric-item">
                    <div className="metric-label">Assisted Count</div>
                    <div className="metric-value" style={{ color: 'var(--text-muted)' }}>
                      {c.assisted_count}
                    </div>
                  </div>
                  <div className="metric-item">
                    <div className="metric-label">Setting Groups Seen</div>
                    <div className="metric-value">
                      {c.setting_groups_seen.length} / 2 required
                    </div>
                  </div>
                  <div className="metric-item">
                    <div className="metric-label">Delayed Transfer</div>
                    <div
                      className="metric-value"
                      style={{
                        color: c.delayed_transfer_achieved ? 'var(--accent-emerald)' : 'var(--text-subtle)',
                      }}
                    >
                      {c.delayed_transfer_achieved ? 'Achieved' : 'Pending'}
                    </div>
                  </div>
                </div>

                {/* Groups and Transfer Info */}
                {c.setting_groups_seen.length > 0 && (
                  <div className="setting-groups-tags">
                    <span style={{ fontSize: '0.75rem', color: 'var(--text-subtle)' }}>Groups:</span>
                    {c.setting_groups_seen.map(g => (
                      <span key={g} className="group-tag">
                        {g}
                      </span>
                    ))}
                  </div>
                )}
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
};

import React from 'react';

export type QuestionStatus = 'pending' | 'in_progress' | 'completed' | 'skipped';

export interface QuestionInfo {
  index: number;
  title: string;
  status: QuestionStatus;
}

export interface QuestionStripProps {
  totalQuestions: number;
  currentQuestionIndex: number;
  questions?: QuestionInfo[];
  onSelectQuestion: (index: number) => void;
  onPrevProblem?: () => void;
  onNextProblem?: () => void;
}

export const QuestionStrip: React.FC<QuestionStripProps> = ({
  totalQuestions,
  currentQuestionIndex,
  questions,
  onSelectQuestion,
  onPrevProblem,
  onNextProblem,
}) => {
  const items: QuestionInfo[] =
    questions && questions.length === totalQuestions
      ? questions
      : Array.from({ length: totalQuestions }).map((_, idx) => ({
          index: idx,
          title: `Problem ${idx + 1}`,
          status: idx === currentQuestionIndex ? 'in_progress' : idx < currentQuestionIndex ? 'completed' : 'pending',
        }));

  return (
    <div
      className="question-strip-bar"
      role="navigation"
      aria-label="Problem Jump List"
      style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        flexWrap: 'wrap',
        gap: '0.75rem',
        padding: '0.6rem 1rem',
        background: 'var(--bg-surface)',
        border: '1px solid var(--border-subtle)',
        borderRadius: 'var(--radius-md)',
        marginBottom: '1rem',
      }}
    >
      {/* Left: Problem counter & Jump list */}
      <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', flexWrap: 'wrap' }}>
        <span style={{ fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-muted)' }}>
          Problem {currentQuestionIndex + 1} of {totalQuestions}:
        </span>
        <div className="question-strip" style={{ display: 'flex', gap: '0.35rem', flexWrap: 'wrap' }}>
          {items.map((q) => {
            const isActive = q.index === currentQuestionIndex;
            let statusBadge = '';
            if (q.status === 'completed') statusBadge = ' ✓';
            else if (q.status === 'skipped') statusBadge = ' ↷';

            return (
              <button
                key={q.index}
                className={`question-pill ${isActive ? 'active' : ''} ${q.status}`}
                onClick={() => onSelectQuestion(q.index)}
                aria-label={`Jump to ${q.title} (${q.status})`}
                aria-current={isActive ? 'step' : undefined}
                title={`${q.title} - Status: ${q.status}`}
                style={{
                  padding: '0.25rem 0.6rem',
                  fontSize: '0.85rem',
                  borderRadius: 'var(--radius-sm)',
                  border: isActive ? '1px solid var(--accent-cyan)' : '1px solid var(--border-subtle)',
                  background: isActive ? 'var(--accent-cyan-subtle)' : 'var(--bg-surface-elevated)',
                  color: isActive ? 'var(--accent-cyan)' : 'var(--text-main)',
                  cursor: 'pointer',
                  fontWeight: isActive ? 600 : 400,
                }}
              >
                {q.index + 1}
                {statusBadge}
              </button>
            );
          })}
        </div>
      </div>

      {/* Right: Big problem navigation buttons with keyboard hints */}
      {(onPrevProblem || onNextProblem) && (
        <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap' }}>
          <button
            className="btn btn-outline"
            disabled={currentQuestionIndex === 0 || !onPrevProblem}
            onClick={onPrevProblem}
            title="Previous Big Problem (Ctrl+Left or Shift+H)"
            style={{ fontSize: '0.8rem', padding: '0.25rem 0.6rem' }}
          >
            &larr; <span className="kbd">Ctrl+&larr;</span> Prev Problem
          </button>
          <button
            className="btn btn-outline"
            disabled={currentQuestionIndex === totalQuestions - 1 || !onNextProblem}
            onClick={onNextProblem}
            title="Next Big Problem (Ctrl+Right or Shift+L)"
            style={{ fontSize: '0.8rem', padding: '0.25rem 0.6rem' }}
          >
            Next Problem <span className="kbd">Ctrl+&rarr;</span> &rarr;
          </button>
        </div>
      )}
    </div>
  );
};

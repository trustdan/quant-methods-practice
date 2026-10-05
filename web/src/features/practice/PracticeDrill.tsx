import React, { useState, useEffect, useRef, useCallback } from 'react';
import {
  PublicSessionView,
  SessionCommandDTO,
  CommandResultDTO,
  PublicStageView,
} from '../../types/practice';
import { MathMarkdown } from '../../components/MathMarkdown';
import { StageStrip, StageInfo } from '../../components/StageStrip';
import { QuestionStrip } from '../../components/QuestionStrip';
import { DrillRecapView } from './DrillRecapView';

interface PracticeDrillProps {
  session: PublicSessionView;
  onSendCommand: (cmd: SessionCommandDTO) => Promise<CommandResultDTO | null>;
  onNavigateStage: (index: number) => void;
  onNavigateQuestion?: (index: number) => void;
  onResetDrill: () => void;
  onDraftChange?: (hasDraft: boolean) => void;
}

interface DraftRecord {
  kind: 'choice' | 'numeric';
  optionId?: string;
  numericRaw?: string;
}

export const PracticeDrill: React.FC<PracticeDrillProps> = ({
  session,
  onSendCommand,
  onNavigateStage,
  onNavigateQuestion,
  onResetDrill,
  onDraftChange,
}) => {
  const currentStage: PublicStageView | undefined = session.stages[session.current_stage_index];

  const [selectedOptionId, setSelectedOptionId] = useState<string | null>(null);
  const [numericInput, setNumericInput] = useState<string>('');
  const [isSubmitting, setIsSubmitting] = useState<boolean>(false);
  const [showRecapTab, setShowRecapTab] = useState<boolean>(session.completed);
  const [lockedStageNotice, setLockedStageNotice] = useState<string | null>(null);

  // In-memory draft store per stage ID
  const draftsRef = useRef<Record<string, DraftRecord>>({});
  const numericInputRef = useRef<HTMLInputElement>(null);
  const prevStageIndexRef = useRef<number>(session.current_stage_index);
  const saveTimeoutRef = useRef<any>(null);

  // Determine reachable stages: learner cannot jump ahead of current highest uncompleted stage
  let maxReachableIndex = 0;
  for (let i = 0; i < session.stages.length; i++) {
    if (session.stages[i].status !== 'unvisited') {
      maxReachableIndex = i;
    }
  }

  // Populate draftsRef from server session draft_answer on initial load
  useEffect(() => {
    session.stages.forEach((st) => {
      if (st.draft_answer && !draftsRef.current[st.id]) {
        if (st.draft_answer.kind === 'choice' && st.draft_answer.option_id) {
          draftsRef.current[st.id] = { kind: 'choice', optionId: st.draft_answer.option_id };
        } else if (st.draft_answer.kind === 'numeric' && st.draft_answer.numeric_raw) {
          draftsRef.current[st.id] = { kind: 'numeric', numericRaw: st.draft_answer.numeric_raw };
        }
      }
    });
  }, [session.stages]);

  // Check if any draft is dirty
  const checkHasDraft = useCallback(() => {
    for (const st of session.stages) {
      if (st.status !== 'completed') {
        const d = draftsRef.current[st.id];
        if (d) {
          if (d.kind === 'choice' && d.optionId) return true;
          if (d.kind === 'numeric' && d.numericRaw && d.numericRaw.trim().length > 0) return true;
        }
      }
    }
    return false;
  }, [session.stages]);

  // Synchronize input fields when switching stages
  useEffect(() => {
    setLockedStageNotice(null);

    if (session.completed) {
      setShowRecapTab(true);
    } else {
      setShowRecapTab(false);
    }

    if (!currentStage) return;

    if (currentStage.status === 'completed') {
      // Completed stage: render historical submitted answer
      if (currentStage.kind === 'choice') {
        const lastAtt = currentStage.attempts[currentStage.attempts.length - 1];
        if (lastAtt && lastAtt.submitted_answer?.option_id) {
          setSelectedOptionId(lastAtt.submitted_answer.option_id);
        } else {
          setSelectedOptionId(null);
        }
        setNumericInput('');
      } else if (currentStage.kind === 'numeric') {
        const lastAtt = currentStage.attempts[currentStage.attempts.length - 1];
        if (lastAtt && lastAtt.submitted_answer?.numeric_raw) {
          setNumericInput(lastAtt.submitted_answer.numeric_raw);
        } else {
          setNumericInput('');
        }
        setSelectedOptionId(null);
      }
    } else {
      // In-progress stage: restore draft if available
      const existingDraft = draftsRef.current[currentStage.id];
      if (existingDraft) {
        if (existingDraft.kind === 'choice') {
          setSelectedOptionId(existingDraft.optionId || null);
          setNumericInput('');
        } else if (existingDraft.kind === 'numeric') {
          setNumericInput(existingDraft.numericRaw || '');
          setSelectedOptionId(null);
        }
      } else if (currentStage.draft_answer) {
        if (currentStage.draft_answer.kind === 'choice') {
          setSelectedOptionId(currentStage.draft_answer.option_id || null);
          setNumericInput('');
        } else if (currentStage.draft_answer.kind === 'numeric') {
          setNumericInput(currentStage.draft_answer.numeric_raw || '');
          setSelectedOptionId(null);
        }
      } else {
        setSelectedOptionId(null);
        setNumericInput('');
      }

      // Auto-focus numeric input if in numeric stage
      if (currentStage.kind === 'numeric') {
        setTimeout(() => {
          numericInputRef.current?.focus();
        }, 50);
      }
    }

    prevStageIndexRef.current = session.current_stage_index;
    if (onDraftChange) {
      onDraftChange(checkHasDraft());
    }
  }, [session.current_stage_index, session.completed, currentStage, onDraftChange, checkHasDraft]);

  // Debounced server draft sync
  const syncDraftToServer = (stageId: string, answerPayload: any) => {
    if (saveTimeoutRef.current) {
      clearTimeout(saveTimeoutRef.current);
    }
    saveTimeoutRef.current = setTimeout(() => {
      onSendCommand({
        command_id: `cmd_draft_${Date.now()}`,
        expected_revision: session.revision,
        type: 'save_draft',
        stage_id: stageId,
        answer: answerPayload,
      });
    }, 600);
  };

  const handleSelectOption = (optId: string) => {
    if (!currentStage || currentStage.status === 'completed') return;
    setSelectedOptionId(optId);
    draftsRef.current[currentStage.id] = { kind: 'choice', optionId: optId };
    if (onDraftChange) onDraftChange(checkHasDraft());
    syncDraftToServer(currentStage.id, { kind: 'choice', option_id: optId });
  };

  const handleNumericChange = (value: string) => {
    if (!currentStage || currentStage.status === 'completed') return;
    setNumericInput(value);
    draftsRef.current[currentStage.id] = { kind: 'numeric', numericRaw: value };
    if (onDraftChange) onDraftChange(checkHasDraft());
    syncDraftToServer(currentStage.id, { kind: 'numeric', numeric_raw: value });
  };

  const stripStages: StageInfo[] = session.stages.map((st, idx) => ({
    id: st.id,
    number: st.number,
    label: st.label,
    completed: st.status === 'completed',
    revealed: st.revealed,
    locked: idx > maxReachableIndex,
    lockReason: idx > maxReachableIndex ? `Stage ${idx + 1} (${st.label}) is locked: Complete previous stages first.` : undefined,
  }));

  const handleSubmit = async () => {
    if (!currentStage || isSubmitting) return;

    let answerPayload: any = null;
    if (currentStage.kind === 'choice') {
      if (!selectedOptionId) return;
      answerPayload = {
        kind: 'choice',
        option_id: selectedOptionId,
      };
    } else if (currentStage.kind === 'numeric') {
      if (!numericInput.trim()) return;
      answerPayload = {
        kind: 'numeric',
        numeric_raw: numericInput.trim(),
      };
    }

    if (!answerPayload) return;

    setIsSubmitting(true);
    const cmdID = `cmd_${Date.now()}_${Math.random().toString(36).substring(2, 7)}`;
    try {
      const res = await onSendCommand({
        command_id: cmdID,
        expected_revision: session.revision,
        type: 'submit_answer',
        stage_id: currentStage.id,
        answer: answerPayload,
      });

      if (res && res.success) {
        delete draftsRef.current[currentStage.id];
        if (onDraftChange) onDraftChange(checkHasDraft());
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleRequestHint = async () => {
    if (!currentStage || isSubmitting) return;

    setIsSubmitting(true);
    const cmdID = `cmd_hint_${Date.now()}`;
    try {
      await onSendCommand({
        command_id: cmdID,
        expected_revision: session.revision,
        type: 'request_hint',
        stage_id: currentStage.id,
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleNextStage = () => {
    if (session.current_stage_index < session.stages.length - 1) {
      const nextIdx = session.current_stage_index + 1;
      if (nextIdx <= maxReachableIndex) {
        onNavigateStage(nextIdx);
      } else {
        setLockedStageNotice(`Stage ${nextIdx + 1} (${session.stages[nextIdx].label}) is locked: You must solve Stage ${session.current_stage_index + 1} first.`);
      }
    } else if (session.completed) {
      setShowRecapTab(true);
    }
  };

  const handlePrevStage = () => {
    if (session.current_stage_index > 0) {
      onNavigateStage(session.current_stage_index - 1);
    }
  };

  const handleSelectStage = (idx: number) => {
    if (idx <= maxReachableIndex) {
      onNavigateStage(idx);
    } else {
      setLockedStageNotice(`Stage ${idx + 1} (${session.stages[idx].label}) is locked: Complete previous stages before jumping ahead.`);
    }
  };

  const handleNavigateQuestion = async (targetIdx: number) => {
    if (targetIdx === (session.current_question_index || 0)) return;
    if (onNavigateQuestion) {
      onNavigateQuestion(targetIdx);
      return;
    }
    const cmdID = `cmd_nav_q_${Date.now()}`;
    await onSendCommand({
      command_id: cmdID,
      expected_revision: session.revision,
      type: 'navigate_question',
      target_question_index: targetIdx,
    });
  };

  const handlePrevQuestion = () => {
    const cur = session.current_question_index || 0;
    if (cur > 0) {
      handleNavigateQuestion(cur - 1);
    }
  };

  const handleNextQuestion = () => {
    const cur = session.current_question_index || 0;
    const total = session.total_questions || (session.questions ? session.questions.length : 1);
    if (cur < total - 1) {
      handleNavigateQuestion(cur + 1);
    }
  };

  // Keyboard navigation complying with docs/NAVIGATION.md
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Pass through IME composing
      if (e.isComposing) return;

      const target = e.target as HTMLElement;
      const isInput = target && (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable);

      // Big problem navigation shortcuts (Ctrl+Left, Ctrl+Right, [, ])
      if ((e.ctrlKey && e.key === 'ArrowLeft') || e.key === '[') {
        e.preventDefault();
        handlePrevQuestion();
        return;
      }
      if ((e.ctrlKey && e.key === 'ArrowRight') || e.key === ']') {
        e.preventDefault();
        handleNextQuestion();
        return;
      }

      // Pass through standard browser combinations
      if (e.ctrlKey || e.metaKey || e.altKey) return;

      if (isInput) {
        if (e.key === 'Enter') {
          e.preventDefault();
          handleSubmit();
        }
        return;
      }

      if (e.key === 'Enter' || e.key === ' ') {
        e.preventDefault();
        if (currentStage?.status === 'completed') {
          handleNextStage();
        } else {
          handleSubmit();
        }
      } else if (e.key === 'h' || e.key === 'ArrowLeft') {
        e.preventDefault();
        handlePrevStage();
      } else if (e.key === 'l' || e.key === 'ArrowRight') {
        e.preventDefault();
        handleNextStage();
      } else if (e.key === '?' || e.key === 'e') {
        e.preventDefault();
        handleRequestHint();
      } else if (currentStage?.kind === 'choice' && currentStage.status !== 'completed') {
        // Choice selection 1-4 or a-d
        const letterMap: Record<string, number> = { a: 0, b: 1, c: 2, d: 3 };
        let choiceIdx = -1;
        if (e.key >= '1' && e.key <= '4') {
          choiceIdx = parseInt(e.key, 10) - 1;
        } else if (letterMap[e.key.toLowerCase()] !== undefined) {
          choiceIdx = letterMap[e.key.toLowerCase()];
        }

        if (choiceIdx >= 0 && choiceIdx < currentStage.options.length) {
          e.preventDefault();
          handleSelectOption(currentStage.options[choiceIdx].id);
        } else if (e.key === 'j' || e.key === 'ArrowDown') {
          // Move down choice list
          e.preventDefault();
          const curIdx = currentStage.options.findIndex((o) => o.id === selectedOptionId);
          const nextIdx = curIdx < currentStage.options.length - 1 ? curIdx + 1 : 0;
          handleSelectOption(currentStage.options[nextIdx].id);
        } else if (e.key === 'k' || e.key === 'ArrowUp') {
          // Move up choice list
          e.preventDefault();
          const curIdx = currentStage.options.findIndex((o) => o.id === selectedOptionId);
          const prevIdx = curIdx > 0 ? curIdx - 1 : currentStage.options.length - 1;
          handleSelectOption(currentStage.options[prevIdx].id);
        }
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [currentStage, selectedOptionId, numericInput, isSubmitting, session.completed, session.current_stage_index, session.stages, maxReachableIndex]);

  if (showRecapTab && session.recap) {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
        <QuestionStrip
          totalQuestions={session.total_questions || (session.questions ? session.questions.length : 1)}
          currentQuestionIndex={session.current_question_index || 0}
          questions={session.questions?.map((q) => ({
            index: q.index,
            title: q.title,
            status: q.status,
          }))}
          onSelectQuestion={handleNavigateQuestion}
          onPrevProblem={handlePrevQuestion}
          onNextProblem={handleNextQuestion}
        />
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <button className="btn btn-outline" onClick={() => setShowRecapTab(false)}>
            &larr; Back to Problem Stages
          </button>
          <button className="btn btn-secondary" onClick={onResetDrill}>
            Restart Drill
          </button>
        </div>
        <DrillRecapView recap={session.recap} onRestart={onResetDrill} />
      </div>
    );
  }

  if (!currentStage) {
    return <div>Loading stage...</div>;
  }

  const isCompleted = currentStage.status === 'completed';
  const isRetry = currentStage.status === 'retry';

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '1.25rem' }}>
      {/* Question Strip / Problem Jump List Header */}
      <QuestionStrip
        totalQuestions={session.total_questions || (session.questions ? session.questions.length : 1)}
        currentQuestionIndex={session.current_question_index || 0}
        questions={session.questions?.map((q) => ({
          index: q.index,
          title: q.title,
          status: q.status,
        }))}
        onSelectQuestion={handleNavigateQuestion}
        onPrevProblem={handlePrevQuestion}
        onNextProblem={handleNextQuestion}
      />

      {/* Scenario & Context Header Card */}
      <div
        className="card"
        style={{
          borderLeft: '4px solid var(--accent-cyan)',
          background: 'var(--bg-surface)',
        }}
      >
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', flexWrap: 'wrap', gap: '1rem' }}>
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.6rem', marginBottom: '0.4rem' }}>
              <span className="badge badge-cyan">Approved Scenario</span>
              <h2 style={{ fontSize: '1.2rem', fontWeight: 600, margin: 0 }}>{session.title}</h2>
            </div>
            <div style={{ fontSize: '1.05rem', color: 'var(--text-main)', marginTop: '0.4rem' }}>
              <MathMarkdown content={session.scenario_markdown} />
            </div>
          </div>
          {session.completed && (
            <button className="btn btn-primary" onClick={() => setShowRecapTab(true)}>
              View Full Recap &rarr;
            </button>
          )}
        </div>

        {/* Assumptions / Conditions Tags */}
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.5rem', marginTop: '1rem' }}>
          {session.assumptions.map((assump, idx) => (
            <span
              key={idx}
              style={{
                fontSize: '0.75rem',
                color: 'var(--text-muted)',
                background: 'var(--bg-surface-elevated)',
                border: '1px solid var(--border-subtle)',
                borderRadius: 'var(--radius-sm)',
                padding: '0.2rem 0.5rem',
              }}
            >
              • {assump}
            </span>
          ))}
        </div>
      </div>

      {/* Stage Navigation Strip */}
      <div className="strip-container">
        <StageStrip
          stages={stripStages}
          currentStageIndex={session.current_stage_index}
          onSelectStage={handleSelectStage}
        />
        <div style={{ display: 'flex', gap: '0.5rem' }}>
          <button
            className="btn btn-outline"
            disabled={session.current_stage_index === 0}
            onClick={handlePrevStage}
            title="Previous Stage (h)"
          >
            <span className="kbd">h</span> Prev
          </button>
          <button
            className="btn btn-outline"
            disabled={session.current_stage_index === session.stages.length - 1 && !session.completed}
            onClick={handleNextStage}
            title="Next Stage (l)"
          >
            Next <span className="kbd">l</span>
          </button>
        </div>
      </div>

      {/* Locked Stage Warning Notice */}
      {lockedStageNotice && (
        <div
          role="alert"
          style={{
            padding: '0.75rem 1rem',
            background: 'rgba(239, 68, 68, 0.1)',
            border: '1px solid rgba(239, 68, 68, 0.4)',
            borderRadius: 'var(--radius-sm)',
            color: 'var(--accent-red, #f87171)',
            fontSize: '0.9rem',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
          }}
        >
          <span>
            <strong>Locked Stage:</strong> {lockedStageNotice}
          </span>
          <button
            className="btn btn-outline"
            style={{ padding: '0.15rem 0.5rem', fontSize: '0.75rem' }}
            onClick={() => setLockedStageNotice(null)}
          >
            Dismiss
          </button>
        </div>
      )}

      {/* Active Stage Card */}
      <div className="card">
        {/* Stage Header */}
        <div className="card-header">
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.6rem' }}>
            <span className="badge badge-cyan">
              Stage {currentStage.number}/7: {currentStage.label}
            </span>
            {isCompleted && (
              <span className="badge badge-emerald">
                {currentStage.solved_on_retry ? 'Solved on Retry' : currentStage.revealed ? 'Solution Revealed' : 'Completed'}
              </span>
            )}
            {isRetry && <span className="badge badge-amber">1 Retry Remaining</span>}
          </div>

          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', fontSize: '0.85rem', color: 'var(--text-muted)' }}>
            <span>Attempt {Math.min(currentStage.attempt_count + (isCompleted ? 0 : 1), 2)} of 2</span>
          </div>
        </div>

        {/* Prompt Markdown */}
        <div style={{ fontSize: '1.1rem', margin: '1rem 0 1.25rem 0' }}>
          <MathMarkdown content={currentStage.prompt_markdown} />
        </div>

        {/* Invalid Input Validation Notice (Ungraded) */}
        {currentStage.invalid_input_notice && (
          <div
            role="alert"
            style={{
              padding: '0.75rem 1rem',
              background: 'rgba(251, 191, 36, 0.1)',
              border: '1px solid rgba(251, 191, 36, 0.4)',
              borderRadius: 'var(--radius-sm)',
              color: 'var(--accent-amber)',
              fontSize: '0.9rem',
              marginBottom: '1.25rem',
            }}
          >
            <strong>Input Format Notice:</strong> {currentStage.invalid_input_notice}
          </div>
        )}

        {/* Choice Stage Options */}
        {currentStage.kind === 'choice' && (
          <div className="options-grid" style={{ display: 'grid', gap: '0.75rem' }}>
            {currentStage.options.map((opt, idx) => {
              const isSelected = selectedOptionId === opt.id;
              return (
                <button
                  key={opt.id}
                  className={`option-btn ${isSelected ? 'selected' : ''}`}
                  disabled={isCompleted}
                  onClick={() => handleSelectOption(opt.id)}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '0.75rem',
                    textAlign: 'left',
                    opacity: isCompleted && !isSelected ? 0.6 : 1,
                  }}
                >
                  <span className="kbd">{idx + 1}</span>
                  <div style={{ flex: 1, fontSize: '0.95rem' }}>
                    <MathMarkdown content={opt.text_markdown} />
                  </div>
                </button>
              );
            })}
          </div>
        )}

        {/* Numeric Stage Input */}
        {currentStage.kind === 'numeric' && (
          <div style={{ margin: '1.25rem 0', display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
            <label htmlFor="drill-numeric-input" style={{ fontSize: '0.9rem', color: 'var(--text-muted)' }}>
              Enter exact probability value (decimal, explicit percent, or fraction):
            </label>
            <div style={{ display: 'flex', gap: '0.75rem', maxWidth: '420px' }}>
              <input
                id="drill-numeric-input"
                ref={numericInputRef}
                type="text"
                disabled={isCompleted}
                value={numericInput}
                onChange={(e) => handleNumericChange(e.target.value)}
                placeholder="e.g. 0.375, 37.5%, or 3/8"
                style={{
                  flex: 1,
                  padding: '0.65rem 0.9rem',
                  background: 'var(--bg-surface-elevated)',
                  border: '1px solid var(--border-strong)',
                  borderRadius: 'var(--radius-sm)',
                  color: 'var(--text-main)',
                  fontSize: '1rem',
                  fontFamily: 'var(--font-mono)',
                }}
              />
              {!isCompleted && (
                <button
                  className="btn btn-primary"
                  disabled={isSubmitting || !numericInput.trim()}
                  onClick={handleSubmit}
                >
                  Submit
                </button>
              )}
            </div>
            <div style={{ fontSize: '0.8rem', color: 'var(--text-subtle)' }}>
              Accepted forms: <code>0.375</code>, <code>37.5%</code>, or <code>3/8</code>. Commas are rejected to avoid localization ambiguity.
            </div>
          </div>
        )}

        {/* Causal Hint Banner (on Attempt 1 error or Hint request) */}
        {currentStage.active_hint && !isCompleted && (
          <div
            style={{
              marginTop: '1.25rem',
              padding: '1rem',
              background: 'rgba(251, 191, 36, 0.08)',
              border: '1px solid rgba(251, 191, 36, 0.3)',
              borderRadius: 'var(--radius-md)',
            }}
          >
            <div style={{ fontWeight: 600, color: 'var(--accent-amber)', marginBottom: '0.35rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              <span>Causal Hint &amp; Guidance:</span>
            </div>
            <div style={{ fontSize: '0.95rem' }}>
              <MathMarkdown content={currentStage.active_hint} />
            </div>
          </div>
        )}

        {/* Completed Explanation / Feedback Banner */}
        {isCompleted && currentStage.last_feedback && (
          <div
            style={{
              marginTop: '1.5rem',
              padding: '1.25rem',
              background: currentStage.revealed ? 'rgba(192, 132, 252, 0.08)' : 'rgba(16, 185, 129, 0.08)',
              border: `1px solid ${currentStage.revealed ? 'rgba(192, 132, 252, 0.3)' : 'rgba(52, 211, 153, 0.3)'}`,
              borderRadius: 'var(--radius-md)',
            }}
          >
            <div
              style={{
                fontWeight: 600,
                color: currentStage.revealed ? '#c084fc' : 'var(--accent-emerald)',
                marginBottom: '0.5rem',
              }}
            >
              {currentStage.revealed ? 'Solution Revealed:' : 'Mathematical Explanation:'}
            </div>
            <MathMarkdown content={currentStage.last_feedback} />
          </div>
        )}

        {/* Action Row */}
        <div style={{ marginTop: '1.5rem', display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: '0.75rem' }}>
          <div style={{ display: 'flex', gap: '0.75rem' }}>
            {!isCompleted ? (
              <>
                <button
                  className="btn btn-primary"
                  disabled={isSubmitting || (currentStage.kind === 'choice' && !selectedOptionId) || (currentStage.kind === 'numeric' && !numericInput.trim())}
                  onClick={handleSubmit}
                >
                  <span className="kbd">Enter</span> Submit Answer
                </button>
                <button
                  className="btn btn-secondary"
                  disabled={isSubmitting}
                  onClick={handleRequestHint}
                >
                  <span className="kbd">?</span> Offline Hint
                </button>
              </>
            ) : (
              <button className="btn btn-primary" onClick={handleNextStage}>
                <span className="kbd">Enter</span> {session.current_stage_index === session.stages.length - 1 ? 'View Complete Recap' : 'Next Stage'} &rarr;
              </button>
            )}
          </div>

          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            {selectedOptionId && !isCompleted && (
              <span style={{ fontSize: '0.85rem', color: 'var(--accent-cyan)' }}>
                Option selected
              </span>
            )}
            <button className="btn btn-outline" onClick={onResetDrill} title="Reset Drill">
              Reset Drill
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

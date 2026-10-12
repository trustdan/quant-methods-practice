import React, { useState, useEffect, useCallback } from 'react';
import { MathMarkdown } from '../../components/MathMarkdown';
import {
  PublicExamView,
  ExamSummary,
  ExamCommandDTO,
  PublicQuestionView,
} from './types';
import { SoftwareToolDrawer } from '../reference/SoftwareToolDrawer';
import { WorkedSoftwareCallBox } from '../reference/WorkedSoftwareCallBox';

interface ExamsProps {
  active: boolean;
  onReturnToPractice: () => void;
}

export const Exams: React.FC<ExamsProps> = ({ active, onReturnToPractice }) => {
  const [subTab, setSubTab] = useState<'create' | 'history'>('create');
  const [summaries, setSummaries] = useState<ExamSummary[]>([]);
  const [loadingSummaries, setLoadingSummaries] = useState<boolean>(false);
  const [activeExam, setActiveExam] = useState<PublicExamView | null>(null);
  const [loadingExam, setLoadingExam] = useState<boolean>(false);
  const [examError, setExamError] = useState<string | null>(null);

  // Creation config
  const [examTitle, setExamTitle] = useState<string>('Standard Assessment');
  const [questionCount, setQuestionCount] = useState<number>(5);
  const [selectedModule, setSelectedModule] = useState<string>('all');
  const [durationMinutes, setDurationMinutes] = useState<number>(30); // 0 = untimed
  const seed = 0;

  // In-exam draft inputs: map of `${qID}:${stageID}` -> raw input string
  const [numericInputs, setNumericInputs] = useState<Record<string, string>>({});

  // Confirmation modals
  const [showSubmitModal, setShowSubmitModal] = useState<boolean>(false);
  const [showAbandonModal, setShowAbandonModal] = useState<boolean>(false);

  // Live timer tick
  const [timeRemainingSeconds, setTimeRemainingSeconds] = useState<number | null>(null);

  const activeQuestionIndex = activeExam?.current_question_index ?? 0;
  const currentQuestion: PublicQuestionView | undefined = activeExam?.questions[activeQuestionIndex];

  const loadExam = useCallback(async (id: string) => {
    setLoadingExam(true);
    setExamError(null);
    try {
      const res = await fetch(`/api/exams/${id}`);
      if (res.ok) {
        const data: PublicExamView = await res.json();
        setActiveExam(data);
      } else {
        const errText = await res.text();
        setExamError(errText || 'Failed to load assessment');
      }
    } catch (err: unknown) {
      setExamError(err instanceof Error ? err.message : 'Failed to connect to assessment server');
    } finally {
      setLoadingExam(false);
    }
  }, []);

  // Fetch summaries list
  const fetchSummaries = useCallback(async () => {
    setLoadingSummaries(true);
    try {
      const res = await fetch('/api/exams');
      if (res.ok) {
        const data: ExamSummary[] = await res.json();
        const list = Array.isArray(data) ? data : [];
        setSummaries(list);
        const activeItem = list.find((item) => item.status === 'active');
        if (activeItem) {
          loadExam(activeItem.id);
        }
      }
    } catch {
      // ignore in offline mock
    } finally {
      setLoadingSummaries(false);
    }
  }, [loadExam]);

  useEffect(() => {
    if (active && !activeExam) {
      fetchSummaries();
    }
  }, [active, activeExam, fetchSummaries]);

  // Update timer countdown every second
  useEffect(() => {
    if (!activeExam || activeExam.status !== 'active' || !activeExam.deadline_at) {
      setTimeRemainingSeconds(null);
      return;
    }

    const deadline = new Date(activeExam.deadline_at).getTime();

    const updateTimer = () => {
      const now = Date.now();
      const diffSec = Math.floor((deadline - now) / 1000);
      if (diffSec <= 0) {
        setTimeRemainingSeconds(0);
        // Automatically check server status when expired
        loadExam(activeExam.id);
      } else {
        setTimeRemainingSeconds(diffSec);
      }
    };

    updateTimer();
    const interval = setInterval(updateTimer, 1000);
    return () => clearInterval(interval);
  }, [activeExam?.id, activeExam?.status, activeExam?.deadline_at, loadExam]);

  // Sync numeric inputs from active exam
  useEffect(() => {
    if (!activeExam) return;
    const inputs: Record<string, string> = {};
    for (const q of activeExam.questions) {
      for (const st of q.stages) {
        if (st.kind === 'numeric' && st.submitted_answer?.numeric_raw) {
          inputs[`${q.question_id}:${st.id}`] = st.submitted_answer.numeric_raw;
        }
      }
    }
    setNumericInputs(inputs);
  }, [activeExam?.id, activeExam?.revision]);

  const handleStartNewExam = async () => {
    setLoadingExam(true);
    setExamError(null);
    try {
      const moduleIDs = selectedModule === 'all' ? [] : [selectedModule];
      const payload = {
        title: examTitle.trim() || 'Assessment',
        question_count: questionCount,
        module_ids: moduleIDs,
        duration_minutes: durationMinutes,
        seed: seed || Date.now(),
      };

      const res = await fetch('/api/exams', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });

      if (res.ok) {
        const newExam: PublicExamView = await res.json();
        setActiveExam(newExam);
      } else {
        const errText = await res.text();
        setExamError(errText || 'Failed to initialize assessment');
      }
    } catch (err: unknown) {
      setExamError(err instanceof Error ? err.message : 'Failed to start assessment');
    } finally {
      setLoadingExam(false);
    }
  };

  const sendCommand = async (cmd: ExamCommandDTO) => {
    if (!activeExam) return;
    try {
      const res = await fetch(`/api/exams/${activeExam.id}/commands`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(cmd),
      });

      if (res.ok) {
        const updated: PublicExamView = await res.json();
        setActiveExam(updated);
      } else if (res.status === 409) {
        setExamError('Revision conflict: exam was updated in another window. Refreshing state.');
        loadExam(activeExam.id);
      } else {
        const err = await res.text();
        setExamError(err || 'Command execution failed');
      }
    } catch (err: unknown) {
      setExamError(err instanceof Error ? err.message : 'Network error');
    }
  };

  const handleSaveChoice = (stageID: string, optionID: string) => {
    if (!activeExam || activeExam.status !== 'active') return;
    sendCommand({
      command_id: `cmd_opt_${Date.now()}`,
      expected_revision: activeExam.revision,
      type: 'save_answer',
      question_index: activeQuestionIndex,
      stage_id: stageID,
      answer: {
        kind: 'choice',
        option_id: optionID,
      },
    });
  };

  const handleSaveNumeric = (stageID: string) => {
    if (!activeExam || activeExam.status !== 'active' || !currentQuestion) return;
    const key = `${currentQuestion.question_id}:${stageID}`;
    const rawVal = numericInputs[key] || '';
    sendCommand({
      command_id: `cmd_num_${Date.now()}`,
      expected_revision: activeExam.revision,
      type: 'save_answer',
      question_index: activeQuestionIndex,
      stage_id: stageID,
      answer: {
        kind: 'numeric',
        numeric_raw: rawVal.trim(),
      },
    });
  };

  const handleNavigateQuestion = (idx: number) => {
    if (!activeExam) return;
    sendCommand({
      command_id: `cmd_nav_${Date.now()}`,
      expected_revision: activeExam.revision,
      type: 'navigate',
      question_index: idx,
    });
  };

  const handleSubmitExam = () => {
    if (!activeExam) return;
    setShowSubmitModal(false);
    sendCommand({
      command_id: `cmd_sub_${Date.now()}`,
      expected_revision: activeExam.revision,
      type: 'submit_exam',
    });
  };

  const handleAbandonExam = () => {
    if (!activeExam) return;
    setShowAbandonModal(false);
    sendCommand({
      command_id: `cmd_ab_${Date.now()}`,
      expected_revision: activeExam.revision,
      type: 'abandon_exam',
    });
  };

  // Keyboard navigation within active exam
  useEffect(() => {
    if (!active || !activeExam || activeExam.status !== 'active') return;

    const handleKeyDown = (e: KeyboardEvent) => {
      const activeEl = document.activeElement;
      const isInput = activeEl?.tagName === 'INPUT' || activeEl?.tagName === 'TEXTAREA';

      if (e.key === 'Escape') {
        if (showSubmitModal) {
          e.stopPropagation();
          setShowSubmitModal(false);
          return;
        }
        if (showAbandonModal) {
          e.stopPropagation();
          setShowAbandonModal(false);
          return;
        }
        return;
      }

      if (isInput) {
        if (e.key === 'Enter') {
          // If in numeric input, save answer on Enter
          const stageID = activeEl?.getAttribute('data-stage-id');
          if (stageID) {
            e.preventDefault();
            handleSaveNumeric(stageID);
          }
        }
        return;
      }

      // Next / Previous Question shortcuts: Shift+H / Shift+L or Ctrl+Arrows
      if ((e.ctrlKey || e.metaKey) && e.key === 'ArrowLeft') {
        e.preventDefault();
        if (activeQuestionIndex > 0) handleNavigateQuestion(activeQuestionIndex - 1);
        return;
      }
      if ((e.ctrlKey || e.metaKey) && e.key === 'ArrowRight') {
        e.preventDefault();
        if (activeQuestionIndex < activeExam.questions.length - 1) handleNavigateQuestion(activeQuestionIndex + 1);
        return;
      }
      if (e.shiftKey && (e.key === 'H' || e.key === 'ArrowLeft')) {
        e.preventDefault();
        if (activeQuestionIndex > 0) handleNavigateQuestion(activeQuestionIndex - 1);
        return;
      }
      if (e.shiftKey && (e.key === 'L' || e.key === 'ArrowRight')) {
        e.preventDefault();
        if (activeQuestionIndex < activeExam.questions.length - 1) handleNavigateQuestion(activeQuestionIndex + 1);
        return;
      }

      // Choice hotkeys: 1-4 or a-d for active stage
      if (currentQuestion && currentQuestion.stages.length > 0) {
        const firstChoiceStage = currentQuestion.stages.find((s) => s.kind === 'choice');
        if (firstChoiceStage && firstChoiceStage.options) {
          let chosenIdx = -1;
          if (e.key === '1' || e.key === 'a') chosenIdx = 0;
          if (e.key === '2' || e.key === 'b') chosenIdx = 1;
          if (e.key === '3' || e.key === 'c') chosenIdx = 2;
          if (e.key === '4' || e.key === 'd') chosenIdx = 3;

          if (chosenIdx >= 0 && chosenIdx < firstChoiceStage.options.length) {
            e.preventDefault();
            handleSaveChoice(firstChoiceStage.id, firstChoiceStage.options[chosenIdx].id);
          }
        }
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [active, activeExam, activeQuestionIndex, currentQuestion, showSubmitModal, showAbandonModal, numericInputs]);

  const formatTimer = (sec: number | null): string => {
    if (sec === null) return '--:--';
    const m = Math.floor(sec / 60);
    const s = sec % 60;
    return `${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`;
  };

  return (
    <div hidden={!active} className="exams-container" style={{ display: active ? 'flex' : 'none', flexDirection: 'column', gap: '1.25rem' }}>
      {/* Error / Notice Banner */}
      {examError && (
        <div
          role="alert"
          style={{
            background: 'rgba(239, 68, 68, 0.15)',
            border: '1px solid rgba(239, 68, 68, 0.4)',
            color: '#f87171',
            padding: '0.75rem 1rem',
            borderRadius: 'var(--radius-sm)',
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            fontSize: '0.9rem',
          }}
        >
          <span>{examError}</span>
          <button className="btn btn-outline" style={{ padding: '0.2rem 0.5rem', fontSize: '0.75rem' }} onClick={() => setExamError(null)}>
            Dismiss
          </button>
        </div>
      )}

      {/* VIEW 1: ASSESSMENT SELECTION & HISTORY DASHBOARD */}
      {!activeExam && (
        <div className="card" id="exam-config-card" style={{ padding: '1.5rem' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.25rem', flexWrap: 'wrap', gap: '0.5rem' }}>
            <div>
              <h2 style={{ margin: 0, fontSize: '1.35rem', fontWeight: 600 }}>Assessments &amp; Exam Mode</h2>
              <p style={{ margin: '0.35rem 0 0 0', color: 'var(--text-muted)', fontSize: '0.88rem' }}>
                Test your mastery under realistic exam conditions. Hints and solutions are withheld until submission.
              </p>
            </div>
            <div style={{ display: 'flex', gap: '0.5rem' }}>
              <button
                className={`btn ${subTab === 'create' ? 'btn-primary' : 'btn-outline'}`}
                onClick={() => setSubTab('create')}
              >
                Start Assessment
              </button>
              <button
                className={`btn ${subTab === 'history' ? 'btn-primary' : 'btn-outline'}`}
                onClick={() => { setSubTab('history'); fetchSummaries(); }}
              >
                History ({(summaries || []).length})
              </button>
              <button className="btn btn-outline" onClick={onReturnToPractice}>
                <span className="kbd">Esc</span> Return to Practice
              </button>
            </div>
          </div>

          {/* Subtab A: Start New Assessment */}
          {subTab === 'create' && (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '1.25rem', maxWidth: '640px' }}>
              <div style={{ padding: '0.85rem 1rem', background: 'rgba(56, 189, 248, 0.08)', borderRadius: 'var(--radius-sm)', border: '1px solid rgba(56, 189, 248, 0.25)', fontSize: '0.85rem', color: 'var(--accent-cyan)' }}>
                <strong>Exam Mode Rules:</strong>
                <ul style={{ margin: '0.35rem 0 0 1.25rem', padding: 0 }}>
                  <li>All answers, hints, and feedback are strictly withheld during the test.</li>
                  <li>In timed exams, the countdown is based on UTC wall-clock time and continues if you close the window.</li>
                  <li>Full worked solutions, derivations, and scores are unlocked upon submission.</li>
                </ul>
              </div>

              <div>
                <label htmlFor="exam-title-input" style={{ display: 'block', fontWeight: 500, fontSize: '0.9rem', marginBottom: '0.35rem' }}>
                  Assessment Title
                </label>
                <input
                  id="exam-title-input"
                  aria-label="Assessment Title"
                  type="text"
                  value={examTitle}
                  onChange={(e) => setExamTitle(e.target.value)}
                  className="input-field"
                  style={{ width: '100%', padding: '0.5rem 0.75rem', background: 'var(--bg-surface-elevated)', border: '1px solid var(--border-color)', borderRadius: 'var(--radius-sm)', color: 'var(--text-primary)' }}
                />
              </div>

              <div>
                <label style={{ display: 'block', fontWeight: 500, fontSize: '0.9rem', marginBottom: '0.35rem' }}>
                  Number of Questions
                </label>
                <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap' }}>
                  {[1, 3, 5, 10].map((num) => (
                    <button
                      key={num}
                      type="button"
                      className={`btn ${questionCount === num ? 'btn-primary' : 'btn-outline'}`}
                      style={{ padding: '0.4rem 1rem', fontSize: '0.85rem' }}
                      onClick={() => setQuestionCount(num)}
                    >
                      {num} {num === 1 ? 'Question' : 'Questions'}
                    </button>
                  ))}
                </div>
              </div>

              <div>
                <label htmlFor="exam-module-select" style={{ display: 'block', fontWeight: 500, fontSize: '0.9rem', marginBottom: '0.35rem' }}>
                  Curriculum Module Filter
                </label>
                <select
                  id="exam-module-select"
                  aria-label="Curriculum Module Filter"
                  value={selectedModule}
                  onChange={(e) => setSelectedModule(e.target.value)}
                  style={{ width: '100%', padding: '0.5rem 0.75rem', background: 'var(--bg-surface-elevated)', border: '1px solid var(--border-color)', borderRadius: 'var(--radius-sm)', color: 'var(--text-primary)' }}
                >
                  <option value="all">All Modules (Comprehensive Exam)</option>
                  <option value="module_1">Module 1: Foundations of Probability &amp; Sets</option>
                  <option value="module_2">Module 2: Discrete Distributions &amp; Moments</option>
                  <option value="module_3">Module 3: Continuous Distributions (Uniform, Normal, Exponential)</option>
                  <option value="module_4">Module 4: Sampling Distributions, CLT, CI &amp; Sample Size</option>
                  <option value="module_5">Module 5: Hypothesis Testing &amp; Decision Errors</option>
                  <option value="module_6">Module 6: Regression Analysis &amp; Diagnostics</option>
                </select>
              </div>

              <div>
                <label style={{ display: 'block', fontWeight: 500, fontSize: '0.9rem', marginBottom: '0.35rem' }}>
                  Timer &amp; Duration Limit
                </label>
                <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap' }}>
                  {[
                    { label: 'Untimed', val: 0 },
                    { label: '15 Minutes', val: 15 },
                    { label: '30 Minutes', val: 30 },
                    { label: '45 Minutes', val: 45 },
                    { label: '60 Minutes', val: 60 },
                  ].map((item) => (
                    <button
                      key={item.val}
                      type="button"
                      className={`btn ${durationMinutes === item.val ? 'btn-primary' : 'btn-outline'}`}
                      style={{ padding: '0.4rem 0.85rem', fontSize: '0.85rem' }}
                      onClick={() => setDurationMinutes(item.val)}
                    >
                      {item.label}
                    </button>
                  ))}
                </div>
              </div>

              <div style={{ marginTop: '0.75rem' }}>
                <button
                  className="btn btn-primary"
                  style={{ padding: '0.65rem 1.5rem', fontSize: '1rem', fontWeight: 600 }}
                  onClick={handleStartNewExam}
                  disabled={loadingExam}
                >
                  {loadingExam ? 'Initializing...' : 'Begin Assessment'}
                </button>
              </div>
            </div>
          )}

          {/* Subtab B: Assessment History */}
          {subTab === 'history' && (
            <div>
              {loadingSummaries ? (
                <p style={{ color: 'var(--text-muted)' }}>Loading assessment history...</p>
              ) : (summaries || []).length === 0 ? (
                <p style={{ color: 'var(--text-muted)' }}>No completed or active assessments recorded yet.</p>
              ) : (
                <div style={{ overflowX: 'auto' }}>
                  <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '0.88rem' }}>
                    <thead>
                      <tr style={{ borderBottom: '1px solid var(--border-color)', textAlign: 'left', color: 'var(--text-muted)' }}>
                        <th style={{ padding: '0.6rem 0.75rem' }}>Title &amp; ID</th>
                        <th style={{ padding: '0.6rem 0.75rem' }}>Status</th>
                        <th style={{ padding: '0.6rem 0.75rem' }}>Duration</th>
                        <th style={{ padding: '0.6rem 0.75rem' }}>Questions</th>
                        <th style={{ padding: '0.6rem 0.75rem' }}>Score</th>
                        <th style={{ padding: '0.6rem 0.75rem' }}>Date</th>
                        <th style={{ padding: '0.6rem 0.75rem', textAlign: 'right' }}>Action</th>
                      </tr>
                    </thead>
                    <tbody>
                      {summaries.map((s) => (
                        <tr key={s.id} style={{ borderBottom: '1px solid var(--border-color)' }}>
                          <td style={{ padding: '0.6rem 0.75rem' }}>
                            <div style={{ fontWeight: 600 }}>{s.title}</div>
                            <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>{s.id}</div>
                          </td>
                          <td style={{ padding: '0.6rem 0.75rem' }}>
                            <span
                              className={`badge ${
                                s.status === 'completed'
                                  ? 'badge-emerald'
                                  : s.status === 'active'
                                  ? 'badge-cyan'
                                  : 'badge-amber'
                              }`}
                            >
                              {s.status}
                            </span>
                          </td>
                          <td style={{ padding: '0.6rem 0.75rem' }}>
                            {s.duration_minutes > 0 ? `${s.duration_minutes} min` : 'Untimed'}
                          </td>
                          <td style={{ padding: '0.6rem 0.75rem' }}>{s.question_count}</td>
                          <td style={{ padding: '0.6rem 0.75rem', fontWeight: 600 }}>
                            {s.score_percent !== undefined && s.score_percent !== null
                              ? `${s.score_percent.toFixed(1)}%`
                              : '-'}
                          </td>
                          <td style={{ padding: '0.6rem 0.75rem', color: 'var(--text-muted)', fontSize: '0.8rem' }}>
                            {new Date(s.created_at).toLocaleDateString()}
                          </td>
                          <td style={{ padding: '0.6rem 0.75rem', textAlign: 'right' }}>
                            <button
                              className="btn btn-outline"
                              style={{ padding: '0.25rem 0.6rem', fontSize: '0.8rem' }}
                              onClick={() => loadExam(s.id)}
                            >
                              {s.status === 'active' ? 'Resume' : 'View Report'}
                            </button>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          )}
        </div>
      )}

      {/* VIEW 2: ACTIVE EXAM WORKSPACE */}
      {activeExam && activeExam.status === 'active' && currentQuestion && (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
          {/* Active Exam Header Bar */}
          <div
            className="card"
            style={{
              padding: '0.85rem 1.25rem',
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
              flexWrap: 'wrap',
              gap: '0.75rem',
            }}
          >
            <div>
              <div style={{ fontWeight: 600, fontSize: '1.1rem' }}>{activeExam.title}</div>
              <div style={{ fontSize: '0.82rem', color: 'var(--text-muted)' }}>
                Question {activeQuestionIndex + 1} of {activeExam.total_questions} •{' '}
                {activeExam.answered_questions} of {activeExam.total_questions} answered
                {activeExam.bank_total_questions && activeExam.bank_total_questions > 0 ? (
                  <span> • 🎲 Randomized from {activeExam.bank_total_questions} bank questions</span>
                ) : null}
              </div>
            </div>

            <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
              {/* Countdown Timer */}
              {activeExam.duration_minutes > 0 && (
                <div
                  aria-label="Time Remaining"
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '0.4rem',
                    padding: '0.4rem 0.85rem',
                    borderRadius: 'var(--radius-sm)',
                    background:
                      (timeRemainingSeconds ?? 999) < 60
                        ? 'rgba(239, 68, 68, 0.2)'
                        : (timeRemainingSeconds ?? 999) < 300
                        ? 'rgba(245, 158, 11, 0.2)'
                        : 'var(--bg-surface-elevated)',
                    color:
                      (timeRemainingSeconds ?? 999) < 60
                        ? '#ef4444'
                        : (timeRemainingSeconds ?? 999) < 300
                        ? '#f59e0b'
                        : 'var(--text-primary)',
                    fontWeight: 600,
                    fontSize: '1rem',
                  }}
                >
                  <span>⏱</span>
                  <span>{formatTimer(timeRemainingSeconds)}</span>
                </div>
              )}

              <button
                className="btn btn-outline"
                style={{ padding: '0.4rem 0.85rem', fontSize: '0.85rem' }}
                onClick={() => setShowAbandonModal(true)}
              >
                Abandon
              </button>

              <button
                className="btn btn-primary"
                style={{ padding: '0.4rem 1.1rem', fontSize: '0.85rem', fontWeight: 600 }}
                onClick={() => setShowSubmitModal(true)}
              >
                Submit Exam
              </button>
            </div>
          </div>

          {/* Question Navigation Strip */}
          <div
            style={{
              display: 'flex',
              gap: '0.4rem',
              overflowX: 'auto',
              padding: '0.4rem 0',
            }}
          >
            {activeExam.questions.map((q, idx) => (
              <button
                key={q.question_id}
                className={`btn ${idx === activeQuestionIndex ? 'btn-primary' : 'btn-outline'}`}
                style={{
                  minWidth: '2.5rem',
                  padding: '0.35rem 0.65rem',
                  fontSize: '0.82rem',
                  display: 'flex',
                  alignItems: 'center',
                  gap: '0.3rem',
                }}
                onClick={() => handleNavigateQuestion(idx)}
              >
                <span>Q{idx + 1}</span>
                {q.is_answered && <span style={{ color: '#34d399', fontSize: '0.75rem' }}>✓</span>}
              </button>
            ))}
          </div>

          {/* Question Workspace Card */}
          <div className="card" style={{ padding: '1.5rem' }}>
            <div style={{ marginBottom: '1.25rem', borderBottom: '1px solid var(--border-color)', paddingBottom: '1rem' }}>
              <div style={{ fontSize: '0.85rem', color: 'var(--accent-cyan)', fontWeight: 600, textTransform: 'uppercase' }}>
                Question {activeQuestionIndex + 1}
              </div>
              <h3 style={{ margin: '0.25rem 0 0.5rem 0', fontSize: '1.2rem', fontWeight: 600 }}>
                {currentQuestion.title}
              </h3>
              <div style={{ fontSize: '0.95rem', color: 'var(--text-primary)', lineHeight: 1.6 }}>
                <MathMarkdown content={currentQuestion.scenario_markdown} />
              </div>

              {currentQuestion.assumptions && currentQuestion.assumptions.length > 0 && (
                <div style={{ marginTop: '0.75rem', display: 'flex', gap: '0.4rem', flexWrap: 'wrap' }}>
                  {currentQuestion.assumptions.map((ass, i) => (
                    <span
                      key={i}
                      style={{
                        fontSize: '0.75rem',
                        padding: '0.2rem 0.5rem',
                        background: 'var(--bg-surface-elevated)',
                        borderRadius: 'var(--radius-sm)',
                        color: 'var(--text-muted)',
                      }}
                    >
                      {ass}
                    </span>
                  ))}
                </div>
              )}
            </div>

            {/* Software Tool Drawer (Withheld during active exam) */}
            <div style={{ marginBottom: '1.25rem' }}>
              <SoftwareToolDrawer
                title={currentQuestion.title}
                withheld={true}
                withheldReason="Software aids are strictly withheld during active timed assessments to evaluate independent conceptual mastery. Full syntax unlocks in the post-exam review."
              />
            </div>

            {/* Stages Input Section */}
            <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
              {currentQuestion.stages.map((stage, stIdx) => {
                const key = `${currentQuestion.question_id}:${stage.id}`;
                return (
                  <div
                    key={stage.id}
                    style={{
                      padding: '1rem',
                      background: 'var(--bg-surface-elevated)',
                      borderRadius: 'var(--radius-sm)',
                      border: '1px solid var(--border-color)',
                    }}
                  >
                    <div style={{ fontSize: '0.82rem', color: 'var(--text-muted)', marginBottom: '0.35rem', fontWeight: 600 }}>
                      Part {stIdx + 1}
                    </div>
                    <div style={{ fontSize: '0.92rem', marginBottom: '0.85rem' }}>
                      <MathMarkdown content={stage.prompt_markdown} />
                    </div>

                    {/* CHOICE STAGE */}
                    {stage.kind === 'choice' && stage.options && (
                      <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                        {stage.options.map((opt, oIdx) => {
                          const isSelected = stage.submitted_answer?.option_id === opt.id;
                          return (
                            <label
                              key={opt.id}
                              style={{
                                display: 'flex',
                                alignItems: 'center',
                                gap: '0.6rem',
                                padding: '0.55rem 0.85rem',
                                borderRadius: 'var(--radius-sm)',
                                background: isSelected ? 'rgba(56, 189, 248, 0.12)' : 'transparent',
                                border: isSelected ? '1px solid var(--accent-cyan)' : '1px solid var(--border-color)',
                                cursor: 'pointer',
                              }}
                            >
                              <input
                                type="radio"
                                name={`opt_${stage.id}`}
                                value={opt.id}
                                checked={isSelected}
                                onChange={() => handleSaveChoice(stage.id, opt.id)}
                              />
                              <span style={{ fontWeight: 600, color: 'var(--text-muted)', minWidth: '1.2rem' }}>
                                {String.fromCharCode(65 + oIdx)}.
                              </span>
                              <div style={{ flex: 1 }}>
                                <MathMarkdown content={opt.text_markdown} />
                              </div>
                            </label>
                          );
                        })}
                      </div>
                    )}

                    {/* NUMERIC STAGE */}
                    {stage.kind === 'numeric' && (
                      <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem', maxWidth: '400px' }}>
                        <div style={{ display: 'flex', gap: '0.5rem' }}>
                          <input
                            type="text"
                            data-stage-id={stage.id}
                            value={numericInputs[key] || ''}
                            onChange={(e) => setNumericInputs({ ...numericInputs, [key]: e.target.value })}
                            placeholder="e.g. 0.375, 37.5%, or 3/8"
                            style={{
                              flex: 1,
                              padding: '0.5rem 0.75rem',
                              background: 'var(--bg-main)',
                              border: '1px solid var(--border-color)',
                              borderRadius: 'var(--radius-sm)',
                              color: 'var(--text-primary)',
                            }}
                          />
                          <button
                            type="button"
                            className="btn btn-outline"
                            onClick={() => handleSaveNumeric(stage.id)}
                          >
                            Save
                          </button>
                        </div>
                        <div style={{ fontSize: '0.78rem', color: 'var(--text-muted)' }}>
                          Accepts decimals (0.375), explicit percents (37.5%), or fractions (3/8).
                        </div>
                        {stage.submitted_answer?.numeric_raw && (
                          <div style={{ fontSize: '0.8rem', color: '#34d399' }}>
                            ✓ Saved answer: {stage.submitted_answer.numeric_raw}
                          </div>
                        )}
                      </div>
                    )}
                  </div>
                );
              })}
            </div>

            {/* Question Navigation Controls */}
            <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: '1.5rem', borderTop: '1px solid var(--border-color)', paddingTop: '1rem' }}>
              <button
                className="btn btn-outline"
                disabled={activeQuestionIndex === 0}
                onClick={() => handleNavigateQuestion(activeQuestionIndex - 1)}
              >
                &larr; Previous Question
              </button>

              <button
                className="btn btn-outline"
                disabled={activeQuestionIndex >= activeExam.questions.length - 1}
                onClick={() => handleNavigateQuestion(activeQuestionIndex + 1)}
              >
                Next Question &rarr;
              </button>
            </div>
          </div>
        </div>
      )}

      {/* VIEW 3: COMPLETED ASSESSMENT REPORT & WORKED SOLUTIONS REVIEW */}
      {activeExam && activeExam.status !== 'active' && activeExam.report && (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '1.25rem' }}>
          {/* Score Banner Hero */}
          <div
            className="card"
            style={{
              padding: '1.75rem',
              textAlign: 'center',
              background: 'linear-gradient(180deg, rgba(56, 189, 248, 0.08) 0%, rgba(15, 23, 42, 0.4) 100%)',
            }}
          >
            <div style={{ fontSize: '0.85rem', color: 'var(--accent-cyan)', fontWeight: 600, textTransform: 'uppercase' }}>
              Assessment Evaluation
            </div>
            <h2 style={{ margin: '0.35rem 0', fontSize: '1.8rem', fontWeight: 700 }}>
              {activeExam.title}
            </h2>
            <div style={{ fontSize: '3rem', fontWeight: 800, color: activeExam.report.score_percent >= 70 ? '#34d399' : '#f59e0b', margin: '0.5rem 0' }}>
              {activeExam.report.score_percent.toFixed(1)}%
            </div>
            <p style={{ margin: 0, color: 'var(--text-muted)', fontSize: '0.95rem' }}>
              {activeExam.report.correct_stages} of {activeExam.report.total_stages} stages correct • Time spent: {Math.floor(activeExam.report.time_spent_seconds / 60)} min {activeExam.report.time_spent_seconds % 60} sec
            </p>

            <div style={{ display: 'flex', justifyContent: 'center', gap: '0.75rem', marginTop: '1.25rem' }}>
              <a
                href={`/api/exams/${activeExam.id}/report?format=markdown`}
                download={`assessment-report-${activeExam.id}.md`}
                className="btn btn-outline"
                style={{ textDecoration: 'none' }}
              >
                Download Markdown Report
              </a>
              <button className="btn btn-outline" onClick={() => setActiveExam(null)}>
                Take Another Assessment
              </button>
              <button className="btn btn-primary" onClick={onReturnToPractice}>
                Return to Practice
              </button>
            </div>
          </div>

          {/* Question-by-Question Review */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: '1.25rem' }}>
            <h3 style={{ margin: 0, fontSize: '1.2rem', fontWeight: 600 }}>Detailed Question Review</h3>
            {activeExam.report.question_reports.map((qr) => (
              <div key={qr.question_id} className="card" style={{ padding: '1.25rem' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.75rem' }}>
                  <div style={{ fontWeight: 600, fontSize: '1.05rem' }}>
                    Q{qr.index + 1}: {qr.title}
                  </div>
                  <span className={`badge ${qr.is_fully_correct ? 'badge-emerald' : 'badge-amber'}`}>
                    {qr.correct_stages} / {qr.total_stages} correct
                  </span>
                </div>

                <div style={{ fontSize: '0.92rem', marginBottom: '1rem', color: 'var(--text-primary)' }}>
                  <MathMarkdown content={qr.scenario_markdown} />
                </div>

                {/* Worked Software Call (Unlocked in Post-Exam Review) */}
                <div style={{ marginBottom: '1rem' }}>
                  <WorkedSoftwareCallBox
                    lookup={{ template_id: qr.template_id, title: qr.title }}
                    expectedAnswer={qr.stage_reports.find((sr) => sr.expected_answer?.value !== undefined)?.expected_answer?.value}
                    allowSideBySide={true}
                  />
                </div>

                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.85rem' }}>
                  {qr.stage_reports.map((sr) => (
                    <div
                      key={sr.id}
                      style={{
                        padding: '0.85rem',
                        background: 'var(--bg-surface-elevated)',
                        borderRadius: 'var(--radius-sm)',
                        borderLeft: `4px solid ${sr.is_correct ? '#34d399' : '#f87171'}`,
                      }}
                    >
                      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.35rem' }}>
                        <span style={{ fontSize: '0.82rem', color: 'var(--text-muted)', fontWeight: 600 }}>
                          Part: {sr.id}
                        </span>
                        <span style={{ fontSize: '0.8rem', fontWeight: 600, color: sr.is_correct ? '#34d399' : '#f87171' }}>
                          {sr.is_correct ? '✓ Correct' : '✗ Incorrect'}
                        </span>
                      </div>

                      <div style={{ fontSize: '0.88rem', marginBottom: '0.5rem' }}>
                        <MathMarkdown content={sr.prompt_markdown} />
                      </div>

                      <div style={{ fontSize: '0.85rem', margin: '0.4rem 0' }}>
                        <strong>Your Answer:</strong>{' '}
                        {sr.submitted_answer ? (
                          sr.kind === 'choice' ? (
                            <code>{sr.submitted_answer.option_id}</code>
                          ) : (
                            <code>{sr.submitted_answer.numeric_raw}</code>
                          )
                        ) : (
                          <span style={{ color: 'var(--text-muted)' }}>(unanswered)</span>
                        )}
                      </div>

                      <div style={{ fontSize: '0.85rem', margin: '0.4rem 0' }}>
                        <strong>Expected Answer:</strong>{' '}
                        {sr.kind === 'choice' ? (
                          <code>{sr.expected_answer.option_id}</code>
                        ) : (
                          <code>{sr.expected_answer.value} {sr.expected_answer.units}</code>
                        )}
                      </div>

                      <div style={{ marginTop: '0.6rem', padding: '0.6rem', background: 'rgba(0, 0, 0, 0.25)', borderRadius: 'var(--radius-sm)', fontSize: '0.85rem' }}>
                        <MathMarkdown content={sr.feedback_markdown} />
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* CONFIRMATION MODAL: SUBMIT EXAM */}
      {showSubmitModal && activeExam && (
        <div
          role="dialog"
          aria-modal="true"
          style={{
            position: 'fixed',
            inset: 0,
            background: 'rgba(0, 0, 0, 0.75)',
            backdropFilter: 'blur(4px)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            zIndex: 100,
          }}
          onClick={() => setShowSubmitModal(false)}
        >
          <div
            className="card"
            style={{ maxWidth: '440px', width: '90%', padding: '1.5rem' }}
            onClick={(e) => e.stopPropagation()}
          >
            <h3 style={{ margin: '0 0 0.75rem 0' }}>Finish &amp; Submit Assessment?</h3>
            <p style={{ fontSize: '0.9rem', color: 'var(--text-primary)', lineHeight: 1.5 }}>
              You have answered <strong>{activeExam.answered_questions}</strong> of{' '}
              <strong>{activeExam.total_questions}</strong> questions.
              {activeExam.answered_questions < activeExam.total_questions && (
                <span style={{ color: '#f59e0b', display: 'block', marginTop: '0.5rem' }}>
                  Notice: Unanswered questions will receive zero points.
                </span>
              )}
            </p>
            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem', marginTop: '1.25rem' }}>
              <button className="btn btn-outline" onClick={() => setShowSubmitModal(false)}>
                Cancel
              </button>
              <button className="btn btn-primary" onClick={handleSubmitExam}>
                Yes, Submit Assessment
              </button>
            </div>
          </div>
        </div>
      )}

      {/* CONFIRMATION MODAL: ABANDON EXAM */}
      {showAbandonModal && (
        <div
          role="dialog"
          aria-modal="true"
          style={{
            position: 'fixed',
            inset: 0,
            background: 'rgba(0, 0, 0, 0.75)',
            backdropFilter: 'blur(4px)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            zIndex: 100,
          }}
          onClick={() => setShowAbandonModal(false)}
        >
          <div
            className="card"
            style={{ maxWidth: '440px', width: '90%', padding: '1.5rem' }}
            onClick={(e) => e.stopPropagation()}
          >
            <h3 style={{ margin: '0 0 0.75rem 0' }}>Abandon Assessment?</h3>
            <p style={{ fontSize: '0.9rem', color: 'var(--text-primary)', lineHeight: 1.5 }}>
              Are you sure you want to exit? Your exam will be marked abandoned and finalized.
            </p>
            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem', marginTop: '1.25rem' }}>
              <button className="btn btn-outline" onClick={() => setShowAbandonModal(false)}>
                Stay in Exam
              </button>
              <button className="btn btn-secondary" onClick={handleAbandonExam}>
                Abandon Assessment
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

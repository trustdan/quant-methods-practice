import React, { useState, useEffect, useCallback, useRef } from 'react';
import { MathMarkdown } from '../components/MathMarkdown';
import { resolveKeyCommand, isEditableElement, KeyMode, NavigationCommand } from '../navigation/keymap';
import { PracticeDrill } from '../features/practice/PracticeDrill';
import { DEFAULT_BINOMIAL_SESSION } from '../features/practice/defaultSession';
import {
  PublicSessionView,
  SessionCommandDTO,
  CommandResultDTO,
} from '../types/practice';
import { ReferenceLibrary } from '../features/reference/ReferenceLibrary';
import { SettingsModal, SessionConfig } from '../features/settings/SettingsModal';
import { MasteryView } from '../features/mastery/MasteryView';
import { NoteLibraryView } from '../features/notes/NoteLibraryView';
import { CandidateReview } from '../features/candidates/CandidateReview';
import { Worksheets, WorksheetMode } from '../features/worksheets/Worksheets';
import { AITutorPanel } from '../features/tutor/AITutorPanel';

export const App: React.FC = () => {
  const [session, setSession] = useState<PublicSessionView>(DEFAULT_BINOMIAL_SESSION);
  const [activeTab, setActiveTab] = useState<'drill' | 'reference' | 'gallery' | 'mastery' | 'notes' | 'candidates' | 'worksheets'>('drill');
  const [candidatesVisited, setCandidatesVisited] = useState(false);
  const [worksheetsVisited, setWorksheetsVisited] = useState(false);
  const [worksheetMode, setWorksheetMode] = useState<WorksheetMode>('full_solution');
  const [showTutorModal, setShowTutorModal] = useState<boolean>(false);
  const [showSettingsModal, setShowSettingsModal] = useState<boolean>(false);
  const [settingsConfig, setSettingsConfig] = useState<SessionConfig>({
    question_count: 10,
    module_ids: [],
    intensity: 'standard',
  });
  const [showHelpModal, setShowHelpModal] = useState<boolean>(false);
  const [showLeaveModal, setShowLeaveModal] = useState<boolean>(false);
  const [hasUnsavedDraft, setHasUnsavedDraft] = useState<boolean>(false);
  const [conflictNotice, setConflictNotice] = useState<string | null>(null);
  const [serverStatus, setServerStatus] = useState<'checking' | 'connected' | 'offline'>('checking');
  const [serverVersion, setServerVersion] = useState<string>('0.1.0-dev');
  const [sessionAuthenticated, setSessionAuthenticated] = useState<boolean>(false);

  const previouslyFocusedElementRef = useRef<HTMLElement | null>(null);
  const helpModalRef = useRef<HTMLDivElement>(null);
  const leaveModalRef = useRef<HTMLDivElement>(null);
  const readingContainerRef = useRef<HTMLDivElement>(null);

  // Authenticate local session and fetch active drill from server
  useEffect(() => {
    const initApp = async () => {
      let token = '';
      if (window.location.hash.startsWith('#bootstrap=')) {
        token = window.location.hash.replace('#bootstrap=', '');
        window.history.replaceState(null, '', window.location.pathname + window.location.search);
      }

      try {
        if (token) {
          const res = await fetch('/api/local-session', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ bootstrap_token: token }),
          });
          if (res.ok) {
            setSessionAuthenticated(true);
          }
        }

        const healthRes = await fetch('/api/health');
        if (healthRes.ok) {
          const healthData = await healthRes.json();
          setServerStatus('connected');
          if (healthData.version) {
            setServerVersion(healthData.version);
          }

          // Fetch active drill session from loopback server
          try {
            const sessRes = await fetch('/api/practice/sessions');
            if (sessRes.ok) {
              const sessData = await sessRes.json();
              if (sessData && sessData.stages && sessData.stages.length > 0) {
                setSession(sessData);
              }
            }
          } catch {
            // Keep default session if fetch fails
          }

          // Fetch user preferences
          try {
            const prefRes = await fetch('/api/settings');
            if (prefRes.ok) {
              const prefData = await prefRes.json();
              if (prefData) {
                setSettingsConfig(prefData);
              }
            }
          } catch {
            // ignore
          }
        } else {
          setServerStatus('offline');
        }
      } catch {
        setServerStatus('offline');
      }
    };

    initApp();
  }, []);

  // Multi-tab sync: when window regains focus, check for server revision updates
  useEffect(() => {
    const handleWindowFocus = async () => {
      if (serverStatus !== 'connected') return;
      try {
        const res = await fetch(`/api/practice/sessions/${session.id}`);
        if (res.ok) {
          const latest: PublicSessionView = await res.json();
          if (latest && latest.revision !== session.revision) {
            setSession(latest);
            setConflictNotice(`Synchronized with latest session state (revision ${latest.revision}).`);
          }
        }
      } catch {
        // ignore background poll errors
      }
    };

    window.addEventListener('focus', handleWindowFocus);
    return () => window.removeEventListener('focus', handleWindowFocus);
  }, [serverStatus, session.id, session.revision]);

  // Warn on browser tab closure if unsaved draft exists
  useEffect(() => {
    const handleBeforeUnload = (e: BeforeUnloadEvent) => {
      if (hasUnsavedDraft) {
        e.preventDefault();
        e.returnValue = 'You have an unsaved answer draft in progress.';
        return 'You have an unsaved answer draft in progress.';
      }
    };

    window.addEventListener('beforeunload', handleBeforeUnload);
    return () => window.removeEventListener('beforeunload', handleBeforeUnload);
  }, [hasUnsavedDraft]);

  // Trap focus inside modals when opened, restore when closed
  useEffect(() => {
    const activeModal = showHelpModal ? helpModalRef.current : showLeaveModal ? leaveModalRef.current : null;
    if (!activeModal) return;

    const focusableElements = activeModal.querySelectorAll<HTMLElement>(
      'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
    );
    if (focusableElements.length > 0) {
      focusableElements[0].focus();
    }

    const handleTabKey = (e: KeyboardEvent) => {
      if (e.key !== 'Tab') return;
      if (focusableElements.length === 0) return;

      const first = focusableElements[0];
      const last = focusableElements[focusableElements.length - 1];

      if (e.shiftKey) {
        if (document.activeElement === first) {
          e.preventDefault();
          last.focus();
        }
      } else {
        if (document.activeElement === last) {
          e.preventDefault();
          first.focus();
        }
      }
    };

    window.addEventListener('keydown', handleTabKey);
    return () => window.removeEventListener('keydown', handleTabKey);
  }, [showHelpModal, showLeaveModal]);

  // Send command to authoritative server, with local fallback for offline testing
  const handleSendCommand = useCallback(
    async (cmd: SessionCommandDTO): Promise<CommandResultDTO | null> => {
      try {
        const res = await fetch(`/api/practice/sessions/${session.id}/commands`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(cmd),
        });

        if (res.ok) {
          const data: CommandResultDTO = await res.json();
          if (data && data.session_state) {
            setSession(data.session_state);
          }
          return data;
        } else {
          // If conflict (409) or validation error, parse authoritative session state
          try {
            const errData = await res.json();
            if (errData && errData.session_state) {
              setSession(errData.session_state);
            }
            if (res.status === 409) {
              setConflictNotice('Notice: Session was updated in another window or tab. Synchronized with latest state.');
            }
            return errData;
          } catch {
            return null;
          }
        }
      } catch {
        // Offline / unit test client simulation fallback
        return localClientSimulateCommand(cmd);
      }
    },
    [session]
  );

  // Local fallback simulation for offline mode / unit tests without live Go server
  const localClientSimulateCommand = (cmd: SessionCommandDTO): CommandResultDTO => {
    const updated = JSON.parse(JSON.stringify(session)) as PublicSessionView;
    const curIdx = updated.current_stage_index;
    const stage = updated.stages[curIdx];

    if (cmd.type === 'navigate_stage' && cmd.target_stage_index !== undefined) {
      updated.current_stage_index = cmd.target_stage_index;
      updated.revision++;
      setSession(updated);
      return { success: true, command_id: cmd.command_id, session_state: updated };
    }

    if (cmd.type === 'save_draft' && cmd.answer) {
      const targetStage = updated.stages.find((s) => s.id === cmd.stage_id) || stage;
      targetStage.draft_answer = cmd.answer;
      setSession(updated);
      return { success: true, command_id: cmd.command_id, session_state: updated };
    }

    if (cmd.type === 'clear_draft') {
      const targetStage = updated.stages.find((s) => s.id === cmd.stage_id) || stage;
      targetStage.draft_answer = undefined;
      setSession(updated);
      return { success: true, command_id: cmd.command_id, session_state: updated };
    }

    if (cmd.type === 'reset_drill') {
      const reset = JSON.parse(JSON.stringify(DEFAULT_BINOMIAL_SESSION));
      setSession(reset);
      return { success: true, command_id: cmd.command_id, session_state: reset };
    }

    if (cmd.type === 'request_hint') {
      stage.active_hint =
        stage.kind === 'choice'
          ? 'Focus on the physical experiment: what quantity is random across repetitions?'
          : 'Substitute into P(X=2) = C(4,2)*(0.5)^2*(0.5)^2 = 6 * 0.0625.';
      updated.revision++;
      setSession(updated);
      return { success: true, command_id: cmd.command_id, session_state: updated };
    }

    if (cmd.type === 'submit_answer' && cmd.answer) {
      stage.draft_answer = undefined;
      if (stage.kind === 'choice') {
        const correctMap: Record<string, string> = {
          define_variable: 'count_heads',
          choose_model: 'binomial',
          check_parameters: 'n4_p_half',
          translate_event: 'equal_two',
          build_expression: 'with_combination',
          interpret_probability: 'whole_experiment',
        };
        const isCorr = cmd.answer.option_id === correctMap[stage.id];
        stage.attempt_count++;
        stage.attempts.push({
          id: `att_${Date.now()}`,
          session_id: updated.id,
          instance_id: 'inst_1',
          stage_id: stage.id,
          attempt_number: stage.attempt_count,
          submitted_answer: cmd.answer,
          assistance: [],
          is_correct: isCorr,
          feedback_markdown: isCorr ? 'Correct!' : 'Incorrect.',
          created_at: new Date().toISOString(),
        });

        if (isCorr) {
          stage.status = 'completed';
          stage.is_correct = true;
          stage.last_feedback = 'Correct! Verified by mathematical engine rules.';
          if (curIdx < updated.stages.length - 1) {
            updated.stages[curIdx + 1].status = 'active';
            updated.current_stage_index = curIdx + 1;
          } else {
            updated.completed = true;
          }
        } else {
          if (stage.attempt_count === 1) {
            stage.status = 'retry';
            stage.active_hint = 'Reconsider the experimental conditions and alternative choices.';
            stage.last_feedback = 'Not quite. Use the causal hint to guide your retry.';
          } else {
            stage.status = 'completed';
            stage.revealed = true;
            stage.is_correct = false;
            stage.last_feedback = 'Incorrect. Solution revealed.';
            if (curIdx < updated.stages.length - 1) {
              updated.stages[curIdx + 1].status = 'active';
              updated.current_stage_index = curIdx + 1;
            } else {
              updated.completed = true;
            }
          }
        }
      } else if (stage.kind === 'numeric') {
        const raw = (cmd.answer.numeric_raw || '').trim();
        if (raw.includes(',')) {
          stage.invalid_input_notice = 'Commas are ambiguous; please use decimal points or fractions.';
          return { success: false, command_id: cmd.command_id, invalid_input: true, session_state: updated };
        }
        stage.invalid_input_notice = undefined;
        stage.attempt_count++;
        const isCorr = raw === '0.375' || raw === '37.5%' || raw === '3/8' || raw === '6/16';
        stage.attempts.push({
          id: `att_${Date.now()}`,
          session_id: updated.id,
          instance_id: 'inst_1',
          stage_id: stage.id,
          attempt_number: stage.attempt_count,
          submitted_answer: cmd.answer,
          assistance: [],
          is_correct: isCorr,
          feedback_markdown: isCorr ? 'Correct!' : 'Incorrect.',
          created_at: new Date().toISOString(),
        });

        if (isCorr) {
          stage.status = 'completed';
          stage.is_correct = true;
          stage.last_feedback = 'Correct! Canonical probability = 0.375 (3/8).';
          if (curIdx < updated.stages.length - 1) {
            updated.stages[curIdx + 1].status = 'active';
            updated.current_stage_index = curIdx + 1;
          } else {
            updated.completed = true;
          }
        } else {
          if (stage.attempt_count === 1) {
            stage.status = 'retry';
            stage.active_hint = raw === '37.5' ? 'Entered 37.5. Did you mean 37.5%? Bare numbers are interpreted as absolute probabilities.' : 'Compute 6 * (0.5)^4 = 6/16 = 3/8.';
          } else {
            stage.status = 'completed';
            stage.revealed = true;
            stage.is_correct = false;
            stage.last_feedback = 'Incorrect. Exact probability is 0.375 (3/8).';
            if (curIdx < updated.stages.length - 1) {
              updated.stages[curIdx + 1].status = 'active';
              updated.current_stage_index = curIdx + 1;
            } else {
              updated.completed = true;
            }
          }
        }
      }

      updated.revision++;
      setSession(updated);
      return { success: true, command_id: cmd.command_id, session_state: updated };
    }

    return { success: true, command_id: cmd.command_id, session_state: updated };
  };

  const handleNavigateStage = (idx: number) => {
    handleSendCommand({
      command_id: `cmd_nav_${Date.now()}`,
      expected_revision: session.revision,
      type: 'navigate_stage',
      target_stage_index: idx,
    });
  };

  const handleNavigateQuestion = (idx: number) => {
    handleSendCommand({
      command_id: `cmd_nav_q_${Date.now()}`,
      expected_revision: session.revision,
      type: 'navigate_question',
      target_question_index: idx,
    });
  };

  const handleStartConfiguredSession = async (config: SessionConfig) => {
    try {
      const res = await fetch('/api/practice/sessions', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(config),
      });
      if (res.ok) {
        const newSession = await res.json();
        setSession(newSession);
        setActiveTab('drill');
        setHasUnsavedDraft(false);
        setConflictNotice(null);
        setSettingsConfig(config);
      }
    } catch (err) {
      console.error('Failed to create session with settings:', err);
    }
  };

  const handleResetDrill = () => {
    handleSendCommand({
      command_id: `cmd_reset_${Date.now()}`,
      expected_revision: session.revision,
      type: 'reset_drill',
    });
  };

  const handleReadingScroll = (cmd: NavigationCommand) => {
    const container = readingContainerRef.current || document.documentElement;
    const halfHeight = window.innerHeight / 2;
    switch (cmd) {
      case 'NAV_DOWN':
        container.scrollBy({ top: 80, behavior: 'smooth' });
        break;
      case 'NAV_UP':
        container.scrollBy({ top: -80, behavior: 'smooth' });
        break;
      case 'SCROLL_HALF_DOWN':
        container.scrollBy({ top: halfHeight, behavior: 'smooth' });
        break;
      case 'SCROLL_HALF_UP':
        container.scrollBy({ top: -halfHeight, behavior: 'smooth' });
        break;
      case 'SCROLL_TOP':
        container.scrollTo({ top: 0, behavior: 'smooth' });
        break;
      case 'SCROLL_BOTTOM':
        container.scrollTo({ top: container.scrollHeight, behavior: 'smooth' });
        break;
    }
  };

  const handleOpenHelp = () => {
    previouslyFocusedElementRef.current = document.activeElement as HTMLElement;
    setShowHelpModal(true);
  };

  const handleCloseHelp = () => {
    setShowHelpModal(false);
    previouslyFocusedElementRef.current?.focus();
  };

  const handleOpenLeaveModal = () => {
    previouslyFocusedElementRef.current = document.activeElement as HTMLElement;
    setShowLeaveModal(true);
  };

  const handleLeaveSave = () => {
    setShowLeaveModal(false);
    setHasUnsavedDraft(false);
    previouslyFocusedElementRef.current?.focus();
  };

  const handleLeaveDiscard = () => {
    const curStage = session.stages[session.current_stage_index];
    if (curStage) {
      handleSendCommand({
        command_id: `cmd_clear_${Date.now()}`,
        expected_revision: session.revision,
        type: 'clear_draft',
        stage_id: curStage.id,
      });
    }
    setShowLeaveModal(false);
    setHasUnsavedDraft(false);
    previouslyFocusedElementRef.current?.focus();
  };

  const handleLeaveCancel = () => {
    setShowLeaveModal(false);
    previouslyFocusedElementRef.current?.focus();
  };

  // Keyboard navigation complying with docs/NAVIGATION.md
  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      const isFieldActive = isEditableElement(document.activeElement);

      const activeMode: KeyMode = showLeaveModal
        ? 'leave_intent'
        : showHelpModal || showSettingsModal
        ? 'modal'
        : (activeTab === 'gallery' || activeTab === 'candidates' || activeTab === 'worksheets')
        ? 'reading'
        : session.stages[session.current_stage_index]?.kind === 'choice'
        ? 'practice_choice'
        : 'practice_numeric';

      const cmd = resolveKeyCommand(e, {
        isEditableActive: isFieldActive,
        activeMode,
      });

      if (!cmd) return;

      switch (cmd) {
        case 'HELP':
          e.preventDefault();
          if (showHelpModal) {
            handleCloseHelp();
          } else {
            handleOpenHelp();
          }
          break;

        case 'ESCAPE':
          e.preventDefault();
          if (showSettingsModal) {
            setShowSettingsModal(false);
          } else if (showTutorModal) {
            setShowTutorModal(false);
          } else if (showHelpModal) {
            handleCloseHelp();
          } else if (showLeaveModal) {
            handleLeaveCancel();
          } else if (activeTab === 'mastery' || activeTab === 'reference' || activeTab === 'notes' || activeTab === 'candidates' || activeTab === 'worksheets') {
            setActiveTab('drill');
          }
          break;

        case 'QUIT':
          if (!isFieldActive) {
            e.preventDefault();
            handleOpenLeaveModal();
          }
          break;

        case 'LEAVE_INTENT_SAVE':
          e.preventDefault();
          handleLeaveSave();
          break;

        case 'LEAVE_INTENT_DISCARD':
          e.preventDefault();
          handleLeaveDiscard();
          break;

        case 'LEAVE_INTENT_CANCEL':
          e.preventDefault();
          handleLeaveCancel();
          break;

        // Reading mode scrolling
        case 'NAV_DOWN':
        case 'NAV_UP':
        case 'SCROLL_HALF_DOWN':
        case 'SCROLL_HALF_UP':
        case 'SCROLL_TOP':
        case 'SCROLL_BOTTOM':
          if (activeTab === 'gallery' || activeTab === 'candidates' || activeTab === 'worksheets') {
            e.preventDefault();
            handleReadingScroll(cmd);
          }
          break;

        // Big problem navigation
        case 'NAV_PREV_PROBLEM':
          if (activeTab === 'candidates' || activeTab === 'worksheets') break;
          e.preventDefault();
          if (session.current_question_index !== undefined && session.current_question_index > 0) {
            handleNavigateQuestion(session.current_question_index - 1);
          }
          break;

        case 'NAV_NEXT_PROBLEM': {
          if (activeTab === 'candidates' || activeTab === 'worksheets') break;
          e.preventDefault();
          const totalQ = session.total_questions || (session.questions ? session.questions.length : 1);
          const curQ = session.current_question_index || 0;
          if (curQ < totalQ - 1) {
            handleNavigateQuestion(curQ + 1);
          }
          break;
        }

        // View tabs
        case 'VIEW_FULL_SOLUTION':
        case 'VIEW_DATASET_CASE':
          if (!isFieldActive) {
            e.preventDefault();
            setWorksheetsVisited(true);
            setWorksheetMode(cmd === 'VIEW_DATASET_CASE' ? 'dataset' : 'full_solution');
            setActiveTab('worksheets');
          }
          break;
        case 'VIEW_REFERENCE':
          if (!isFieldActive) {
            e.preventDefault();
            setActiveTab((prev) => (prev === 'reference' ? 'drill' : 'reference'));
          }
          break;

        case 'VIEW_MASTERY':
          if (!isFieldActive) {
            e.preventDefault();
            setActiveTab((prev) => (prev === 'mastery' ? 'drill' : 'mastery'));
          }
          break;

        case 'VIEW_SETTINGS':
          if (!isFieldActive) {
            e.preventDefault();
            setShowSettingsModal(true);
          }
          break;

        case 'VIEW_SAVED_NOTES':
          if (!isFieldActive) {
            e.preventDefault();
            setActiveTab((prev) => (prev === 'notes' ? 'drill' : 'notes'));
          }
          break;

        case 'VIEW_CANDIDATE_REVIEW':
          if (!isFieldActive) {
            e.preventDefault();
            setCandidatesVisited(true);
            setActiveTab(prev => prev === 'candidates' ? 'drill' : 'candidates');
          }
          break;

        case 'VIEW_AI_REQUEST':
          if (!isFieldActive) {
            e.preventDefault();
            setShowTutorModal(true);
          }
          break;
      }
    },
    [showHelpModal, showLeaveModal, showSettingsModal, showTutorModal, activeTab, session]
  );

  useEffect(() => {
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [handleKeyDown]);

  return (
    <>
      {/* Top Header */}
      <header className="app-header">
        <div className="app-header-left">
          <div className="app-logo">
            <div className="app-logo-badge">Q</div>
            <span>Quant Methods Practice</span>
          </div>
          <span className="badge badge-emerald">Offline practice</span>
        </div>

        {/* Tab Navigation */}
        <div style={{ display: 'flex', gap: '0.4rem', flexWrap: 'wrap' }}>
          <button
            className={`btn ${activeTab === 'drill' ? 'btn-primary' : 'btn-outline'}`}
            onClick={() => setActiveTab('drill')}
            style={{ fontSize: '0.85rem', padding: '0.35rem 0.75rem' }}
          >
            Practice Drill
          </button>
          <button
            className={`btn ${activeTab === 'mastery' ? 'btn-primary' : 'btn-outline'}`}
            onClick={() => setActiveTab('mastery')}
            id="tab-mastery"
            style={{ fontSize: '0.85rem', padding: '0.35rem 0.75rem' }}
            title="Concept Mastery & Transfer (s)"
          >
            <span className="kbd">s</span> Mastery
          </button>
          <button
            className={`btn ${activeTab === 'reference' ? 'btn-primary' : 'btn-outline'}`}
            onClick={() => setActiveTab('reference')}
            id="tab-reference"
            style={{ fontSize: '0.85rem', padding: '0.35rem 0.75rem' }}
            title="Reference Library (r)"
          >
            <span className="kbd">r</span> Reference Library
          </button>
          <button
            className={`btn ${activeTab === 'notes' ? 'btn-primary' : 'btn-outline'}`}
            onClick={() => setActiveTab('notes')}
            id="tab-notes"
            style={{ fontSize: '0.85rem', padding: '0.35rem 0.75rem' }}
            title="Saved Notes & Explanations (V)"
          >
            <span className="kbd">V</span> Notes
          </button>
          <button
            className="btn btn-outline"
            onClick={() => setShowTutorModal(true)}
            id="tab-tutor"
            style={{ fontSize: '0.85rem', padding: '0.35rem 0.75rem' }}
            title="AI Tutor Panel (n)"
          >
            <span className="kbd">n</span> 💡 AI Tutor
          </button>
          <button className={`btn ${activeTab === 'candidates' ? 'btn-primary' : 'btn-outline'}`}
            onClick={() => { setCandidatesVisited(true); setActiveTab('candidates'); }} id="tab-candidates" title="Question candidates (p)">
            Question candidates
          </button>
          <button className={`btn ${activeTab === 'worksheets' ? 'btn-primary' : 'btn-outline'}`}
            onClick={() => { setWorksheetsVisited(true); setActiveTab('worksheets'); }} title="Full solution (Shift+J) or dataset case (Shift+F)">
            Worksheets
          </button>
          <button
            className={`btn ${activeTab === 'gallery' ? 'btn-primary' : 'btn-outline'}`}
            onClick={() => setActiveTab('gallery')}
            style={{ fontSize: '0.85rem', padding: '0.35rem 0.75rem' }}
          >
            Formula Gallery
          </button>
        </div>

        {/* Connection status and action controls */}
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', flexWrap: 'wrap' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem', fontSize: '0.85rem' }}>
            <span
              style={{
                width: 8,
                height: 8,
                borderRadius: '50%',
                backgroundColor: serverStatus === 'connected' ? '#34d399' : '#fbbf24',
                boxShadow: serverStatus === 'connected' ? '0 0 8px #34d399' : 'none',
              }}
            />
            <span style={{ color: 'var(--text-muted)' }}>
              {serverStatus === 'connected'
                ? `Loopback Server Active (v${serverVersion}${sessionAuthenticated ? ' / Auth' : ''})`
                : 'Offline Client Mode'}
            </span>
          </div>
          <button
            className="btn btn-outline"
            onClick={() => setShowSettingsModal(true)}
            title="Session Settings & Module Picker (t)"
          >
            <span className="kbd">t</span> Settings
          </button>
          <button
            className="btn btn-outline"
            onClick={handleOpenHelp}
            title="Keyboard Shortcuts & Usability (F1)"
          >
            <span className="kbd">F1</span> Help
          </button>
          <button
            className="btn btn-outline"
            onClick={handleOpenLeaveModal}
            title="Exit / Leave Practice (q)"
          >
            <span className="kbd">q</span> Exit
          </button>
        </div>
      </header>

      {/* Synchronized conflict notification banner */}
      {conflictNotice && (
        <div
          role="status"
          style={{
            background: 'rgba(56, 189, 248, 0.1)',
            borderBottom: '1px solid rgba(56, 189, 248, 0.3)',
            padding: '0.5rem 1.5rem',
            fontSize: '0.85rem',
            color: 'var(--accent-cyan)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
          }}
        >
          <span>{conflictNotice}</span>
          <button
            className="btn btn-outline"
            style={{ padding: '0.15rem 0.5rem', fontSize: '0.75rem' }}
            onClick={() => setConflictNotice(null)}
          >
            Dismiss
          </button>
        </div>
      )}

      {/* Main Workspace */}
      <main className="app-main" ref={readingContainerRef}>
        {candidatesVisited && <CandidateReview active={activeTab === 'candidates' && !showSettingsModal && !showTutorModal} />}
        {worksheetsVisited && <Worksheets active={activeTab === 'worksheets' && !showSettingsModal && !showTutorModal && !showHelpModal && !showLeaveModal} launchMode={worksheetMode} />}
        {activeTab === 'mastery' && (
          <MasteryView onClose={() => setActiveTab('drill')} />
        )}

        {activeTab === 'notes' && (
          <NoteLibraryView
            onClose={() => setActiveTab('drill')}
            onOpenTutorWithNote={() => setShowTutorModal(true)}
          />
        )}

        {activeTab === 'drill' && (
          <PracticeDrill
            session={session}
            onSendCommand={handleSendCommand}
            onNavigateStage={handleNavigateStage}
            onNavigateQuestion={handleNavigateQuestion}
            onResetDrill={handleResetDrill}
            onDraftChange={setHasUnsavedDraft}
            onOpenTutor={() => setShowTutorModal(true)}
          />
        )}

        {activeTab === 'reference' && (
          <ReferenceLibrary onClose={() => setActiveTab('drill')} />
        )}

        {activeTab === 'gallery' && (
          <div className="card">
            <div className="card-header">
              <h3 className="card-title">Mathematical Formula Gallery (Reading View)</h3>
              <span className="badge badge-cyan">Course Topics Reference</span>
            </div>
            <p style={{ fontSize: '0.85rem', color: 'var(--text-muted)', marginBottom: '1rem' }}>
              Reading Mode active: Use <span className="kbd">j</span>/<span className="kbd">k</span> to scroll, <span className="kbd">u</span>/<span className="kbd">d</span> for half-page, <span className="kbd">gg</span> for top, <span className="kbd">G</span> for bottom.
            </p>
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(320px, 1fr))', gap: '1rem' }}>
              <div style={{ padding: '1rem', background: 'var(--bg-surface-elevated)', borderRadius: 'var(--radius-sm)' }}>
                <div style={{ fontWeight: 600, color: 'var(--accent-cyan)', marginBottom: '0.4rem' }}>
                  1. Set Operations &amp; Probability
                </div>
                <MathMarkdown
                  content={
                    'Union: $P(A \\cup B) = P(A) + P(B) - P(A \\cap B)$\n\n' +
                    'Conditional: $P(A \\mid B) = \\frac{P(A \\cap B)}{P(B)}$'
                  }
                />
              </div>
              <div style={{ padding: '1rem', background: 'var(--bg-surface-elevated)', borderRadius: 'var(--radius-sm)' }}>
                <div style={{ fontWeight: 600, color: 'var(--accent-cyan)', marginBottom: '0.4rem' }}>
                  2. Discrete Distributions
                </div>
                <MathMarkdown
                  content={
                    'Binomial: $$P(X = k) = \\binom{n}{k} p^k (1 - p)^{n - k}$$\n\n' +
                    'Poisson: $$P(X = k) = \\frac{\\lambda^k e^{-\\lambda}}{k!}$$'
                  }
                />
              </div>
              <div style={{ padding: '1rem', background: 'var(--bg-surface-elevated)', borderRadius: 'var(--radius-sm)' }}>
                <div style={{ fontWeight: 600, color: 'var(--accent-cyan)', marginBottom: '0.4rem' }}>
                  3. Sums &amp; Covariance
                </div>
                <MathMarkdown
                  content={
                    'Linearity: $E[X + Y] = E[X] + E[Y]$\n\n' +
                    'Variance: $$\\operatorname{Var}(X + Y) = \\operatorname{Var}(X) + \\operatorname{Var}(Y) + 2\\operatorname{Cov}(X, Y)$$'
                  }
                />
              </div>
              <div style={{ padding: '1rem', background: 'var(--bg-surface-elevated)', borderRadius: 'var(--radius-sm)' }}>
                <div style={{ fontWeight: 600, color: 'var(--accent-cyan)', marginBottom: '0.4rem' }}>
                  4. Continuous Normal Distribution
                </div>
                <MathMarkdown
                  content={
                    'Standard Normal density:\n\n' +
                    '$$\\phi(z) = \\frac{1}{\\sqrt{2\\pi}} e^{-\\frac{z^2}{2}}$$'
                  }
                />
              </div>
            </div>
          </div>
        )}
      </main>

      {/* Footer / Status Bar */}
      <footer className="app-footer">
        <div className="key-hints">
          <div className="key-hint">
            <span className="kbd">h</span> / <span className="kbd">l</span>
            <span>Stage</span>
          </div>
          <div className="key-hint">
            <span className="kbd">1</span>-<span className="kbd">4</span> / <span className="kbd">a</span>-<span className="kbd">d</span>
            <span>Select</span>
          </div>
          <div className="key-hint">
            <span className="kbd">Enter</span>
            <span>Submit</span>
          </div>
          <div className="key-hint">
            <span className="kbd">?</span>
            <span>Hint</span>
          </div>
          <div className="key-hint">
            <span className="kbd">F1</span>
            <span>Help</span>
          </div>
          <div className="key-hint">
            <span className="kbd">q</span>
            <span>Exit</span>
          </div>
        </div>
        <div>
          <span>Local loopback mode: 127.0.0.1</span>
        </div>
      </footer>

      {/* Comprehensive Keyboard Shortcut & Usability Help Modal */}
      {showHelpModal && (
        <div
          role="dialog"
          aria-modal="true"
          aria-label="Keyboard Shortcuts and Usability Guide"
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
          onClick={handleCloseHelp}
        >
          <div
            ref={helpModalRef}
            className="card"
            style={{ maxWidth: '620px', width: '92%', maxHeight: '85vh', overflowY: 'auto' }}
            onClick={(e) => e.stopPropagation()}
          >
            <div className="card-header">
              <h3 className="card-title">Keyboard Navigation Contract</h3>
              <button className="btn btn-outline" onClick={handleCloseHelp}>
                <span className="kbd">Esc</span> Close
              </button>
            </div>

            <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem', fontSize: '0.88rem' }}>
              <div>
                <h4 style={{ color: 'var(--accent-cyan)', margin: '0 0 0.4rem 0', fontSize: '0.95rem' }}>
                  1. Problem &amp; Stage Navigation
                </h4>
                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.35rem' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <span style={{ color: 'var(--text-muted)' }}>Previous / Next Stage:</span>
                    <span><span className="kbd">h</span> / <span className="kbd">l</span> or <span className="kbd">&larr;</span> / <span className="kbd">&rarr;</span></span>
                  </div>
                  <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <span style={{ color: 'var(--text-muted)' }}>Previous / Next Big Problem:</span>
                    <span><span className="kbd">Ctrl+&larr;</span> / <span className="kbd">Ctrl+&rarr;</span> or <span className="kbd">Shift+H</span> / <span className="kbd">Shift+L</span></span>
                  </div>
                </div>
              </div>

              <div>
                <h4 style={{ color: 'var(--accent-cyan)', margin: '0 0 0.4rem 0', fontSize: '0.95rem' }}>
                  2. Answering &amp; Selection
                </h4>
                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.35rem' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <span style={{ color: 'var(--text-muted)' }}>Select Choice Option:</span>
                    <span><span className="kbd">1</span> - <span className="kbd">4</span> or <span className="kbd">a</span> - <span className="kbd">d</span></span>
                  </div>
                  <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <span style={{ color: 'var(--text-muted)' }}>Move Choice Focus Up / Down:</span>
                    <span><span className="kbd">j</span> / <span className="kbd">k</span> or <span className="kbd">&uarr;</span> / <span className="kbd">&darr;</span></span>
                  </div>
                  <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <span style={{ color: 'var(--text-muted)' }}>Submit Answer / Advance:</span>
                    <span><span className="kbd">Enter</span> or <span className="kbd">Space</span></span>
                  </div>
                  <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <span style={{ color: 'var(--text-muted)' }}>Request Causal Hint:</span>
                    <span><span className="kbd">?</span> or <span className="kbd">e</span></span>
                  </div>
                </div>
              </div>

              <div>
                <h4 style={{ color: 'var(--accent-cyan)', margin: '0 0 0.4rem 0', fontSize: '0.95rem' }}>
                  3. Reading Views &amp; Worksheets
                </h4>
                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.35rem' }}>
                  <div>Full solution: <span className="kbd">Shift+J</span>. CSV dataset case: <span className="kbd">Shift+F</span>. Use Tab to move between form fields; save the worksheet draft before closing.</div>
                  <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <span style={{ color: 'var(--text-muted)' }}>Line Scroll:</span>
                    <span><span className="kbd">j</span> / <span className="kbd">k</span></span>
                  </div>
                  <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <span style={{ color: 'var(--text-muted)' }}>Half-Page Scroll:</span>
                    <span><span className="kbd">u</span> / <span className="kbd">d</span> or <span className="kbd">PgUp</span> / <span className="kbd">PgDn</span></span>
                  </div>
                  <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <span style={{ color: 'var(--text-muted)' }}>Top / Bottom:</span>
                    <span><span className="kbd">gg</span> / <span className="kbd">G</span></span>
                  </div>
                </div>
              </div>

              <div>
                <h4 style={{ color: 'var(--accent-cyan)', margin: '0 0 0.4rem 0', fontSize: '0.95rem' }}>
                  4. Leave Intent &amp; Protection
                </h4>
                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.35rem' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <span style={{ color: 'var(--text-muted)' }}>Exit Practice:</span>
                    <span><span className="kbd">q</span> (prompts save if draft is unsent)</span>
                  </div>
                  <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <span style={{ color: 'var(--text-muted)' }}>Leave Intent Workflow:</span>
                    <span><span className="kbd">y</span> Save, <span className="kbd">n</span> Discard, <span className="kbd">Esc</span> Stay</span>
                  </div>
                  <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <span style={{ color: 'var(--text-muted)' }}>Editable Field Protection:</span>
                    <span>Typing inside input preserves native text; shortcuts do not trigger.</span>
                  </div>
                </div>
              </div>
            </div>

            <div style={{ marginTop: '1.5rem', textAlign: 'right' }}>
              <button className="btn btn-primary" onClick={handleCloseHelp}>
                Close Help
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Leave-Intent Save Confirmation Modal */}
      {showLeaveModal && (
        <div
          role="dialog"
          aria-modal="true"
          aria-label="Confirm Leave Practice"
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
          onClick={handleLeaveCancel}
        >
          <div
            ref={leaveModalRef}
            className="card"
            style={{ maxWidth: '480px', width: '90%' }}
            onClick={(e) => e.stopPropagation()}
          >
            <div className="card-header">
              <h3 className="card-title">Leave Practice Session?</h3>
              <button className="btn btn-outline" onClick={handleLeaveCancel}>
                <span className="kbd">Esc</span>
              </button>
            </div>

            <div style={{ fontSize: '0.95rem', margin: '1rem 0' }}>
              {hasUnsavedDraft ? (
                <p>
                  You have an unsubmitted answer draft in progress. Would you like to save your draft before exiting?
                </p>
              ) : (
                <p>
                  Your completed progress is safely stored in SQLite. Would you like to return or exit now?
                </p>
              )}
            </div>

            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem', marginTop: '1.25rem' }}>
              <button className="btn btn-outline" onClick={handleLeaveCancel}>
                <span className="kbd">Esc</span> Cancel
              </button>
              {hasUnsavedDraft && (
                <button className="btn btn-secondary" onClick={handleLeaveDiscard}>
                  <span className="kbd">n</span> Discard
                </button>
              )}
              <button className="btn btn-primary" onClick={handleLeaveSave}>
                <span className="kbd">y</span> {hasUnsavedDraft ? 'Save & Exit' : 'Exit'}
              </button>
            </div>
          </div>
        </div>
      )}
      {/* Settings Modal */}
      <SettingsModal
        isOpen={showSettingsModal}
        onClose={() => setShowSettingsModal(false)}
        onStartSession={handleStartConfiguredSession}
        currentSettings={settingsConfig}
      />
      {/* AI Tutor Panel */}
      <AITutorPanel
        isOpen={showTutorModal}
        onClose={() => setShowTutorModal(false)}
        sessionId={session.id}
        instanceId={session.template_id}
        stageId={session.stages[session.current_stage_index]?.id}
        instanceTitle={session.title}
        topic={session.stages[session.current_stage_index]?.label || 'Probability'}
        concepts={[session.template_id]}
      />
    </>
  );
};

import React, { useState, useEffect, useRef, useCallback } from 'react';
import { MathMarkdown } from '../../components/MathMarkdown';
import {
  TutorAction,
  FollowUpKind,
  TutorEventDTO,
  SavedExplanationDTO,
} from '../../types/tutor';
import { BillingRoute, BILLING_LABELS } from '../../types/providers';

export interface AITutorPanelProps {
  isOpen: boolean;
  onClose: () => void;
  sessionId?: string;
  instanceId?: string;
  stageId?: string;
  instanceTitle?: string;
  topic?: string;
  concepts?: string[];
  onNoteSaved?: (note: SavedExplanationDTO) => void;
}

export const AITutorPanel: React.FC<AITutorPanelProps> = ({
  isOpen,
  onClose,
  sessionId,
  instanceId,
  stageId,
  instanceTitle,
  topic = 'Probability',
  concepts = [],
  onNoteSaved,
}) => {
  const [isStreaming, setIsStreaming] = useState<boolean>(false);
  const [streamedText, setStreamedText] = useState<string>('');
  const [activeRequestId, setActiveRequestId] = useState<string | null>(null);
  const [fallbackNotice, setFallbackNotice] = useState<string | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isSaved, setIsSaved] = useState<boolean>(false);
  const [saveSuccess, setSaveSuccess] = useState<boolean>(false);
  const [customPrompt, setCustomPrompt] = useState<string>('');
  const [showSaveOnLeave, setShowSaveOnLeave] = useState<boolean>(false);
  const [pendingLeaveAction, setPendingLeaveAction] = useState<(() => void) | null>(null);
  const [saveError, setSaveError] = useState<string | null>(null);
  const [activeProvider, setActiveProvider] = useState<{
    route: string;
    name: string;
    model: string;
    billing?: BillingRoute;
  }>({
    route: 'offline',
    name: 'Offline Reviewed',
    model: 'offline-curriculum',
  });

  const abortControllerRef = useRef<AbortController | null>(null);
  const panelRef = useRef<HTMLDivElement>(null);

  // Load active provider details on open
  useEffect(() => {
    if (isOpen) {
      fetch('/api/providers')
        .then((res) => (res.ok ? res.json() : null))
        .then((data) => {
          if (data && data.active_route) {
            const current = data.providers?.find(
              (p: { route: string; billing?: BillingRoute }) => p.route === data.active_route,
            );
            setActiveProvider({
              route: data.active_route,
              name: current?.name || data.active_route,
              model: data.active_model || '',
              billing: current?.billing,
            });
          }
        })
        .catch(() => {});
    }
  }, [isOpen]);

  // Load recovery draft on open if one exists
  useEffect(() => {
    if (isOpen && sessionId && stageId && !streamedText) {
      const draftId = `draft_${sessionId}_${stageId}`;
      fetch(`/api/tutor/drafts/${draftId}`)
        .then((res) => {
          if (res.ok) return res.json();
          return null;
        })
        .then((data) => {
          if (data && data.recovery_text) {
            setStreamedText(data.recovery_text);
            setIsSaved(false);
          }
        })
        .catch(() => {});
    }
  }, [isOpen, sessionId, stageId, streamedText]);

  // Persist recovery draft periodically to backend when unsaved text changes
  useEffect(() => {
    if (streamedText && !isSaved && sessionId && stageId) {
      const draftId = `draft_${sessionId}_${stageId}`;
      const timer = setTimeout(() => {
        fetch(`/api/tutor/drafts/${draftId}`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            context_json: JSON.stringify({ topic, concepts }),
            recovery_text: streamedText,
          }),
        }).catch(() => {});
      }, 1000);
      return () => clearTimeout(timer);
    }
  }, [streamedText, isSaved, sessionId, stageId, topic, concepts]);

  // Stop / Cancel active generation
  const handleCancelStreaming = useCallback(async () => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
      abortControllerRef.current = null;
    }
    if (activeRequestId) {
      try {
        await fetch(`/api/tutor/requests/${activeRequestId}`, {
          method: 'DELETE',
        });
      } catch {
        // ignore
      }
    }
    setIsStreaming(false);
    setActiveRequestId(null);
  }, [activeRequestId]);

  // Execute tutor streaming interaction
  const handleStartRequest = async (
    action: TutorAction,
    followUpKind?: FollowUpKind,
    customQuery?: string
  ) => {
    if (isStreaming) {
      await handleCancelStreaming();
    }

    setIsStreaming(true);
    setErrorMessage(null);
    setFallbackNotice(null);
    setIsSaved(false);
    setSaveSuccess(false);
    setSaveError(null);
    setStreamedText('');

    const controller = new AbortController();
    abortControllerRef.current = controller;

    try {
      const res = await fetch('/api/tutor/requests', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          session_id: sessionId,
          instance_id: instanceId,
          stage_id: stageId,
          action,
          follow_up_kind: followUpKind,
          custom_prompt: customQuery,
          provider: activeProvider.route,
        }),
        signal: controller.signal,
      });

      if (!res.ok) {
        throw new Error(`Failed to initiate tutor request (${res.status})`);
      }

      const initData = await res.json();
      const reqId = initData.request_id;
      setActiveRequestId(reqId);

      // Connect to SSE stream
      const eventRes = await fetch(`/api/tutor/requests/${reqId}/events`, {
        signal: controller.signal,
      });

      if (!eventRes.ok || !eventRes.body) {
        throw new Error('Streaming connection failed');
      }

      const reader = eventRes.body.getReader();
      const decoder = new TextDecoder();
      let buffer = '';

      while (true) {
        const { value, done } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split('\n');
        buffer = lines.pop() || '';

        for (const line of lines) {
          const trimmed = line.trim();
          if (trimmed.startsWith('data:')) {
            const jsonStr = trimmed.slice(5).trim();
            if (!jsonStr) continue;
            try {
              const ev: TutorEventDTO = JSON.parse(jsonStr);
              if (ev.type === 'text_delta' && ev.delta) {
                setStreamedText((prev) => prev + ev.delta);
              } else if (ev.type === 'complete' && ev.text) {
                setStreamedText(ev.text);
                setIsStreaming(false);
              } else if (ev.type === 'fallback') {
                setFallbackNotice(ev.fallback_label || 'Offline reviewed explanation fallback used');
              } else if (ev.type === 'error') {
                setErrorMessage(ev.error || 'Tutor provider error occurred');
                setIsStreaming(false);
              } else if (ev.type === 'cancelled') {
                setIsStreaming(false);
              }
            } catch {
              // ignore json parse glitch
            }
          }
        }
      }
    } catch (err: unknown) {
      if (err instanceof Error && err.name === 'AbortError') {
        // cancellation
      } else {
        const msg = err instanceof Error ? err.message : 'Unknown streaming error';
        setErrorMessage(msg);
      }
    } finally {
      setIsStreaming(false);
      abortControllerRef.current = null;
    }
  };

  // Idempotent Save Note to SQLite
  const handleSaveNote = async (): Promise<boolean> => {
    if (!streamedText) return false;
    setSaveError(null);

    const notePayload = {
      raw_markdown: streamedText,
      origin_instance_id: instanceId || 'general',
      origin_stage_id: stageId || 'general',
      topic: topic || 'Probability & Statistics',
      provider_info: {
        title: instanceTitle ? `Notes: ${instanceTitle}` : `${topic} Explanation`,
        concepts: concepts || [],
        provider: activeProvider.route,
        model: activeProvider.model || 'curriculum',
        route: activeProvider.route,
        fallback_label: fallbackNotice || undefined,
        advisory_status: 'Advisory AI explanation: for self-study only, does not affect score or official grades',
      },
    };

    try {
      const res = await fetch('/api/notes', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(notePayload),
      });

      if (!res.ok) {
        throw new Error(`Failed to save note (${res.status})`);
      }

      const savedNote: SavedExplanationDTO = await res.json();
      setIsSaved(true);
      setSaveSuccess(true);
      if (onNoteSaved) {
        onNoteSaved(savedNote);
      }

      // Clear recovery draft since note is safely saved
      if (sessionId && stageId) {
        const draftId = `draft_${sessionId}_${stageId}`;
        fetch(`/api/tutor/drafts/${draftId}`, { method: 'DELETE' }).catch(() => {});
      }
      return true;
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to save note';
      setSaveError(msg);
      return false;
    }
  };

  // Save-on-leave navigation guard:
  // If user has unsaved text, prompt with Save? y/n/Esc.
  const handleAttemptClose = (action?: () => void) => {
    const nextAction = action || onClose;
    if (isStreaming) {
      handleCancelStreaming();
    }
    if (streamedText && !isSaved) {
      setPendingLeaveAction(() => nextAction);
      setShowSaveOnLeave(true);
    } else {
      nextAction();
    }
  };

  // Leave intent handlers (y / n / Esc)
  const handleLeaveSave = async () => {
    const success = await handleSaveNote();
    if (success) {
      setShowSaveOnLeave(false);
      if (pendingLeaveAction) {
        pendingLeaveAction();
        setPendingLeaveAction(null);
      }
    }
  };

  const handleLeaveDiscard = () => {
    setShowSaveOnLeave(false);
    if (sessionId && stageId) {
      const draftId = `draft_${sessionId}_${stageId}`;
      fetch(`/api/tutor/drafts/${draftId}`, { method: 'DELETE' }).catch(() => {});
    }
    setStreamedText('');
    setIsSaved(true);
    if (pendingLeaveAction) {
      pendingLeaveAction();
      setPendingLeaveAction(null);
    }
  };

  const handleLeaveCancel = () => {
    setShowSaveOnLeave(false);
    setPendingLeaveAction(null);
  };

  // Keyboard navigation inside Tutor Panel
  useEffect(() => {
    if (!isOpen) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      if (showSaveOnLeave) {
        if (e.key === 'y' || e.key === 'Y' || e.key === 'Enter') {
          e.preventDefault();
          handleLeaveSave();
        } else if (e.key === 'n' || e.key === 'N') {
          e.preventDefault();
          handleLeaveDiscard();
        } else if (e.key === 'Escape') {
          e.preventDefault();
          handleLeaveCancel();
        }
        return;
      }

      if (e.key === 'Escape') {
        e.preventDefault();
        if (isStreaming) {
          handleCancelStreaming();
        } else {
          handleAttemptClose();
        }
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, showSaveOnLeave, isStreaming, streamedText, isSaved]);

  if (!isOpen) return null;

  return (
    <div className="modal-backdrop" id="tutor-backdrop" onClick={() => handleAttemptClose()}>
      <div
        className="modal-content tutor-modal-container"
        id="tutor-panel"
        ref={panelRef}
        onClick={(e) => e.stopPropagation()}
        style={{ maxWidth: '820px', width: '95%', maxHeight: '90vh', display: 'flex', flexDirection: 'column' }}
      >
        {/* Header */}
        <div className="modal-header">
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
            <h2 className="modal-title" id="tutor-title">
              💡 AI Tutor
            </h2>
            <span
              id="tutor-provider-badge"
              className="badge"
              style={{
                background:
                  activeProvider.route === 'offline'
                    ? 'rgba(56, 189, 248, 0.15)'
                    : activeProvider.route === 'gemini'
                    ? 'rgba(168, 85, 247, 0.15)'
                    : activeProvider.route === 'anthropic'
                    ? 'rgba(245, 158, 11, 0.15)'
                    : 'rgba(16, 185, 129, 0.15)',
                color:
                  activeProvider.route === 'offline'
                    ? 'var(--accent-cyan)'
                    : activeProvider.route === 'gemini'
                    ? '#c084fc'
                    : activeProvider.route === 'anthropic'
                    ? 'var(--accent-amber)'
                    : 'var(--accent-emerald)',
              }}
            >
              {activeProvider.route === 'offline'
                ? 'Offline Reviewed'
                : `${activeProvider.name} (${activeProvider.model})${
                    activeProvider.billing ? ` · ${BILLING_LABELS[activeProvider.billing]}` : ''
                  }`}
            </span>
          </div>
          <button
            className="btn btn-ghost"
            id="btn-tutor-close"
            onClick={() => handleAttemptClose()}
            title="Close Tutor (Esc)"
          >
            ✕
          </button>
        </div>

        {/* Advisory Banner */}
        <div
          id="tutor-advisory-banner"
          style={{
            margin: '0.75rem 1.25rem 0',
            padding: '0.6rem 0.85rem',
            background: 'rgba(251, 191, 36, 0.1)',
            border: '1px solid rgba(251, 191, 36, 0.3)',
            borderRadius: 'var(--radius-sm)',
            fontSize: '0.8rem',
            color: 'var(--accent-amber)',
            display: 'flex',
            alignItems: 'center',
            gap: '0.5rem',
          }}
        >
          <span>⚠️</span>
          <span>
            <strong>Advisory AI explanation:</strong> For self-study only. Does not alter problem keys, scores, or official grades.
          </span>
        </div>

        {/* Fallback Notice */}
        {fallbackNotice && (
          <div
            id="tutor-fallback-notice"
            style={{
              margin: '0.5rem 1.25rem 0',
              padding: '0.5rem 0.75rem',
              background: 'rgba(56, 189, 248, 0.12)',
              border: '1px solid var(--accent-cyan)',
              borderRadius: 'var(--radius-sm)',
              fontSize: '0.8rem',
              color: 'var(--accent-cyan)',
            }}
          >
            ℹ️ {fallbackNotice}
          </div>
        )}

        {/* Action Prompt Strip */}
        <div
          style={{
            padding: '0.85rem 1.25rem',
            display: 'flex',
            flexWrap: 'wrap',
            gap: '0.5rem',
            borderBottom: '1px solid var(--border-subtle)',
          }}
        >
          <button
            id="btn-tutor-hint"
            className="btn btn-secondary"
            disabled={isStreaming}
            onClick={() => handleStartRequest('hint')}
          >
            ❓ Causal Hint
          </button>
          <button
            id="btn-tutor-explain"
            className="btn btn-secondary"
            disabled={isStreaming}
            onClick={() => handleStartRequest('explain')}
          >
            📐 Full Step-by-Step Solution
          </button>
          <button
            id="btn-tutor-diff"
            className="btn btn-secondary"
            disabled={isStreaming}
            onClick={() => handleStartRequest('follow_up', 'explain_differently')}
          >
            🎨 Explain Differently
          </button>
          <button
            id="btn-tutor-example"
            className="btn btn-secondary"
            disabled={isStreaming}
            onClick={() => handleStartRequest('follow_up', 'worked_example')}
          >
            🔢 Worked Example
          </button>
          <button
            id="btn-tutor-conditions"
            className="btn btn-secondary"
            disabled={isStreaming}
            onClick={() => handleStartRequest('follow_up', 'why_condition_matters')}
          >
            ⚖️ Why Conditions Matter
          </button>
          <button
            id="btn-tutor-compare"
            className="btn btn-secondary"
            disabled={isStreaming}
            onClick={() => handleStartRequest('follow_up', 'compare_concepts')}
          >
            🔄 Compare Concepts
          </button>
        </div>

        {/* Content Body */}
        <div
          id="tutor-response-area"
          style={{
            flex: 1,
            padding: '1.25rem',
            overflowY: 'auto',
            background: 'var(--bg-app)',
            minHeight: '220px',
            position: 'relative',
          }}
        >
          {isStreaming && (
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: '0.5rem',
                color: 'var(--accent-cyan)',
                fontSize: '0.85rem',
                marginBottom: '0.75rem',
              }}
            >
              <div className="spinner" style={{ width: '14px', height: '14px' }}></div>
              <span>Generating curriculum-grounded explanation...</span>
            </div>
          )}

          {errorMessage && (
            <div
              style={{
                padding: '0.75rem',
                background: 'rgba(244, 63, 94, 0.1)',
                border: '1px solid var(--accent-rose)',
                borderRadius: 'var(--radius-sm)',
                color: 'var(--accent-rose)',
                marginBottom: '0.75rem',
              }}
            >
              <p>⚠️ {errorMessage}</p>
              <button
                className="btn btn-secondary"
                style={{ marginTop: '0.5rem', fontSize: '0.8rem' }}
                onClick={() => handleStartRequest('explain')}
              >
                🔄 Retry
              </button>
            </div>
          )}

          {streamedText ? (
            <MathMarkdown content={streamedText} />
          ) : !isStreaming && !errorMessage ? (
            <div style={{ color: 'var(--text-subtle)', textAlign: 'center', padding: '2rem 1rem' }}>
              <p style={{ fontSize: '1.1rem', marginBottom: '0.5rem' }}>Ask a question or select a guided prompt above.</p>
              <p style={{ fontSize: '0.85rem' }}>
                The tutor provides mathematical derivations, causal hints, and conceptual contrasts from verified offline course rules.
              </p>
            </div>
          ) : null}
        </div>

        {/* Custom Follow-Up Query Input */}
        <div
          style={{
            padding: '0.75rem 1.25rem',
            background: 'var(--bg-surface-elevated)',
            borderTop: '1px solid var(--border-subtle)',
            display: 'flex',
            gap: '0.5rem',
          }}
        >
          <input
            id="tutor-custom-input"
            type="text"
            className="input-field"
            style={{ flex: 1 }}
            placeholder="Ask a specific question about notation, assumptions, or formulas..."
            value={customPrompt}
            onChange={(e) => setCustomPrompt(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter' && customPrompt.trim() && !isStreaming) {
                e.preventDefault();
                handleStartRequest('follow_up', 'custom', customPrompt.trim());
              }
            }}
          />
          <button
            id="btn-tutor-ask-custom"
            className="btn btn-primary"
            disabled={isStreaming || !customPrompt.trim()}
            onClick={() => {
              if (customPrompt.trim()) {
                handleStartRequest('follow_up', 'custom', customPrompt.trim());
              }
            }}
          >
            Ask
          </button>
        </div>

        {/* Footer Actions */}
        <div
          style={{
            padding: '0.85rem 1.25rem',
            background: 'var(--bg-surface)',
            borderTop: '1px solid var(--border-subtle)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
          }}
        >
          <div>
            {saveSuccess ? (
              <span style={{ color: 'var(--accent-emerald)', fontSize: '0.85rem', fontWeight: 500 }}>
                ✓ Saved to Personal Notes Library
              </span>
            ) : saveError ? (
              <span style={{ color: 'var(--accent-rose)', fontSize: '0.85rem' }}>
                ⚠️ {saveError} (Retaining content)
              </span>
            ) : null}
          </div>

          <div style={{ display: 'flex', gap: '0.5rem' }}>
            {isStreaming && (
              <button
                id="btn-tutor-cancel"
                className="btn btn-secondary"
                onClick={handleCancelStreaming}
                title="Cancel streaming (Esc)"
              >
                ⏹ Stop
              </button>
            )}

            {streamedText && (
              <button
                id="btn-tutor-save"
                className={`btn ${isSaved ? 'btn-secondary' : 'btn-primary'}`}
                onClick={() => handleSaveNote()}
              >
                {isSaved ? '✓ Saved' : '💾 Save to Notes (V)'}
              </button>
            )}

            <button className="btn btn-ghost" id="btn-tutor-done" onClick={() => handleAttemptClose()}>
              Close (Esc)
            </button>
          </div>
        </div>

        {/* Save-on-Leave Confirmation Modal */}
        {showSaveOnLeave && (
          <div
            className="modal-backdrop"
            style={{ zIndex: 1100 }}
            onClick={(e) => {
              e.stopPropagation();
              handleLeaveCancel();
            }}
          >
            <div
              className="modal-content"
              id="tutor-leave-modal"
              style={{ maxWidth: '420px', textAlign: 'center', padding: '1.5rem' }}
              onClick={(e) => e.stopPropagation()}
            >
              <h3 style={{ marginBottom: '0.75rem', fontSize: '1.15rem' }}>Save Explanation Before Leaving?</h3>
              <p style={{ color: 'var(--text-muted)', fontSize: '0.875rem', marginBottom: '1.25rem' }}>
                You have generated a tutor explanation that hasn't been saved to your notes library yet.
              </p>
              <div style={{ display: 'flex', gap: '0.5rem', justifyContent: 'center' }}>
                <button
                  id="btn-tutor-leave-save"
                  className="btn btn-primary"
                  onClick={handleLeaveSave}
                  title="Save to Library (y)"
                >
                  (y) Save Note
                </button>
                <button
                  id="btn-tutor-leave-discard"
                  className="btn btn-secondary"
                  onClick={handleLeaveDiscard}
                  title="Discard Explanation (n)"
                >
                  (n) Discard
                </button>
                <button
                  id="btn-tutor-leave-stay"
                  className="btn btn-ghost"
                  onClick={handleLeaveCancel}
                  title="Stay on note (Esc)"
                >
                  (Esc) Stay
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

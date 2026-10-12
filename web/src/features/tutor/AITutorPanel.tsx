import React, { useState, useEffect, useRef, useCallback } from 'react';
import { MathMarkdown } from '../../components/MathMarkdown';
import {
  TutorAction,
  FollowUpKind,
  TutorEventDTO,
  SavedExplanationDTO,
  TutorThreadDTO,
  TutorChatMessageDTO,
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
  stageResolved?: boolean;
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
  stageResolved = false,
  onNoteSaved,
}) => {
  // Mode switcher: 'quick' (one-shot guidance/hints) vs 'chat' (multi-turn interactive thread)
  const [panelMode, setPanelMode] = useState<'quick' | 'chat'>('quick');

  // One-shot state
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

  // Multi-turn chat state
  const [threads, setThreads] = useState<TutorThreadDTO[]>([]);
  const [activeThread, setActiveThread] = useState<TutorThreadDTO | null>(null);
  const [chatMessages, setChatMessages] = useState<TutorChatMessageDTO[]>([]);
  const [chatInput, setChatInput] = useState<string>('');
  const [isChatStreaming, setIsChatStreaming] = useState<boolean>(false);
  const [chatStreamingText, setChatStreamingText] = useState<string>('');
  const [chatActiveReqId, setChatActiveReqId] = useState<string | null>(null);
  const [chatError, setChatError] = useState<string | null>(null);
  const [chatNoteSavedMsg, setChatNoteSavedMsg] = useState<string | null>(null);

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
  const chatAbortControllerRef = useRef<AbortController | null>(null);
  const panelRef = useRef<HTMLDivElement>(null);
  const chatEndRef = useRef<HTMLDivElement>(null);

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

  // Load threads for current problem instance and stage
  const loadThreads = useCallback(async () => {
    if (!instanceId || !stageId) return;
    try {
      const res = await fetch(`/api/tutor/threads?instance_id=${encodeURIComponent(instanceId)}&stage_id=${encodeURIComponent(stageId)}`);
      if (res.ok) {
        const data = await res.json();
        const threadList: TutorThreadDTO[] = Array.isArray(data) ? data : (data?.threads || []);
        setThreads(threadList);
        if (threadList.length > 0) {
          // Select most recently updated thread
          const threadRes = await fetch(`/api/tutor/threads/${threadList[0].id}`);
          if (threadRes.ok) {
            const fullThread: TutorThreadDTO = await threadRes.json();
            setActiveThread(fullThread);
            setChatMessages(fullThread.messages || []);
          }
        } else {
          setActiveThread(null);
          setChatMessages([]);
        }
      }
    } catch {
      // offline fallback
    }
  }, [instanceId, stageId]);

  useEffect(() => {
    if (isOpen && panelMode === 'chat') {
      loadThreads();
    }
  }, [isOpen, panelMode, loadThreads]);

  // Scroll to bottom of chat on new message or delta
  useEffect(() => {
    if (panelMode === 'chat' && chatEndRef.current && typeof chatEndRef.current.scrollIntoView === 'function') {
      chatEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [chatMessages, chatStreamingText, panelMode]);

  // Load recovery draft on open if one exists (for quick mode)
  useEffect(() => {
    if (isOpen && sessionId && stageId && !streamedText && panelMode === 'quick') {
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
  }, [isOpen, sessionId, stageId, streamedText, panelMode]);

  // Persist recovery draft periodically to backend when unsaved text changes
  useEffect(() => {
    if (streamedText && !isSaved && sessionId && stageId && panelMode === 'quick') {
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
  }, [streamedText, isSaved, sessionId, stageId, topic, concepts, panelMode]);

  // Stop / Cancel active one-shot generation
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

  // Stop / Cancel active chat generation
  const handleCancelChatStreaming = useCallback(async () => {
    if (chatAbortControllerRef.current) {
      chatAbortControllerRef.current.abort();
      chatAbortControllerRef.current = null;
    }
    if (chatActiveReqId) {
      try {
        await fetch(`/api/tutor/requests/${chatActiveReqId}`, {
          method: 'DELETE',
        });
      } catch {
        // ignore
      }
    }
    setIsChatStreaming(false);
    setChatActiveReqId(null);
  }, [chatActiveReqId]);

  // Execute quick tutor interaction
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

  // Multi-turn chat message submission
  const handleSendChatMessage = async (textToSend?: string) => {
    const messageContent = (textToSend !== undefined ? textToSend : chatInput).trim();
    if (!messageContent) return;

    if (isChatStreaming) {
      await handleCancelChatStreaming();
    }

    setChatInput('');
    setChatError(null);
    setChatNoteSavedMsg(null);

    let thread = activeThread;

    // If no active thread exists for this problem/stage, create one first
    if (!thread) {
      try {
        const createRes = await fetch('/api/tutor/threads', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            session_id: sessionId,
            instance_id: instanceId || 'general',
            stage_id: stageId || 'general',
            provider: activeProvider.route,
            model: activeProvider.model,
            title: messageContent.length > 40 ? messageContent.slice(0, 40) + '...' : messageContent,
          }),
        });

        if (!createRes.ok) {
          throw new Error('Failed to initialize conversation thread');
        }

        thread = await createRes.json();
        setActiveThread(thread);
        setThreads((prev) => [thread!, ...prev]);
      } catch (err: unknown) {
        setChatError(err instanceof Error ? err.message : 'Thread creation failed');
        return;
      }
    }

    if (!thread) return;

    const controller = new AbortController();
    chatAbortControllerRef.current = controller;
    setIsChatStreaming(true);
    setChatStreamingText('');

    try {
      const postRes = await fetch(`/api/tutor/threads/${thread.id}/messages`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          role: 'user',
          content: messageContent,
        }),
        signal: controller.signal,
      });

      if (!postRes.ok) {
        throw new Error(`Failed to send message (${postRes.status})`);
      }

      const postData = await postRes.json();
      const asstReqId = postData.assistant_message_id || postData.request_id;
      setChatActiveReqId(asstReqId);

      // Optimistically append user message to local state
      const newUserMsg: TutorChatMessageDTO = {
        id: postData.user_message_id || `usr_${Date.now()}`,
        thread_id: thread.id,
        role: 'user',
        content: messageContent,
        created_at: new Date().toISOString(),
      };
      setChatMessages((prev) => [...prev, newUserMsg]);

      // Connect to SSE stream
      const eventRes = await fetch(`/api/tutor/threads/${thread.id}/messages/${asstReqId}/events`, {
        signal: controller.signal,
      });

      if (!eventRes.ok || !eventRes.body) {
        throw new Error('Chat streaming connection failed');
      }

      const reader = eventRes.body.getReader();
      const decoder = new TextDecoder();
      let buffer = '';
      let accumulatedText = '';
      let fallbackLabel = '';

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
                accumulatedText += ev.delta;
                setChatStreamingText(accumulatedText);
              } else if (ev.type === 'fallback') {
                fallbackLabel = ev.fallback_label || 'Offline guidance used';
              } else if (ev.type === 'complete') {
                const finalContent = ev.text || accumulatedText;
                setChatMessages((prev) => [
                  ...prev,
                  {
                    id: asstReqId,
                    thread_id: thread!.id,
                    role: 'assistant',
                    content: finalContent,
                    fallback_label: fallbackLabel || undefined,
                    created_at: new Date().toISOString(),
                  },
                ]);
                setChatStreamingText('');
                setIsChatStreaming(false);
              } else if (ev.type === 'error') {
                setChatError(ev.error || 'Chat assistant error');
                setIsChatStreaming(false);
              }
            } catch {
              // ignore parse errors
            }
          }
        }
      }
    } catch (err: unknown) {
      if (err instanceof Error && err.name === 'AbortError') {
        // cancellation
      } else {
        setChatError(err instanceof Error ? err.message : 'Error in chat conversation');
      }
    } finally {
      setIsChatStreaming(false);
      chatAbortControllerRef.current = null;
    }
  };

  // Save single assistant message to Note Library
  const handleSaveChatMessageAsNote = async (msg: TutorChatMessageDTO) => {
    try {
      const notePayload = {
        raw_markdown: msg.content,
        origin_instance_id: instanceId || 'general',
        origin_stage_id: stageId || 'general',
        topic: topic || 'Probability & Statistics',
        provider_info: {
          title: instanceTitle ? `Chat Note: ${instanceTitle}` : `${topic} Discussion Note`,
          concepts: concepts || [],
          provider: activeProvider.route,
          model: activeProvider.model || 'curriculum',
          route: activeProvider.route,
          thread_id: activeThread?.id,
          fallback_label: msg.fallback_label,
          advisory_status: 'Advisory AI explanation: for self-study only, does not affect score or official grades',
        },
      };

      const res = await fetch('/api/notes', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(notePayload),
      });

      if (!res.ok) {
        throw new Error('Failed to save message to notes');
      }

      const savedNote: SavedExplanationDTO = await res.json();
      setChatNoteSavedMsg('✓ Message saved to Note Library');
      setTimeout(() => setChatNoteSavedMsg(null), 3000);
      if (onNoteSaved) {
        onNoteSaved(savedNote);
      }
    } catch (err: unknown) {
      setChatError(err instanceof Error ? err.message : 'Could not save message note');
    }
  };

  // Save complete multi-turn conversation transcript to Note Library
  const handleSaveWholeConversationAsNote = async () => {
    if (chatMessages.length === 0) return;

    try {
      const lines: string[] = [];
      lines.push(`# AI Tutor Discussion: ${activeThread?.title || instanceTitle || topic}\n`);
      if (instanceTitle) lines.push(`**Problem**: ${instanceTitle}`);
      lines.push(`**Stage**: ${topic}`);
      lines.push(`**Provider**: ${activeProvider.name} (${activeProvider.model})`);
      lines.push(`**Date**: ${new Date().toLocaleDateString()}\n`);
      lines.push('---\n');

      for (const m of chatMessages) {
        if (m.role === 'user') {
          lines.push(`### 👤 Learner\n${m.content}\n`);
        } else {
          lines.push(`### 🤖 AI Tutor\n${m.content}\n`);
        }
      }

      const notePayload = {
        raw_markdown: lines.join('\n'),
        origin_instance_id: instanceId || 'general',
        origin_stage_id: stageId || 'general',
        topic: topic || 'Probability & Statistics',
        provider_info: {
          title: activeThread?.title ? `Transcript: ${activeThread.title}` : `Transcript: ${instanceTitle || topic}`,
          concepts: concepts || [],
          provider: activeProvider.route,
          model: activeProvider.model || 'curriculum',
          route: activeProvider.route,
          thread_id: activeThread?.id,
          advisory_status: 'Advisory AI explanation: for self-study only, does not affect score or official grades',
        },
      };

      const res = await fetch('/api/notes', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(notePayload),
      });

      if (!res.ok) {
        throw new Error('Failed to save transcript to notes');
      }

      const savedNote: SavedExplanationDTO = await res.json();
      setChatNoteSavedMsg('✓ Full conversation transcript saved to Note Library');
      setTimeout(() => setChatNoteSavedMsg(null), 3500);
      if (onNoteSaved) {
        onNoteSaved(savedNote);
      }
    } catch (err: unknown) {
      setChatError(err instanceof Error ? err.message : 'Could not save conversation note');
    }
  };

  // Idempotent Save Note to SQLite (Quick mode)
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

  // Save-on-leave navigation guard
  const handleAttemptClose = (action?: () => void) => {
    const nextAction = action || onClose;
    if (isStreaming) {
      handleCancelStreaming();
    }
    if (isChatStreaming) {
      handleCancelChatStreaming();
    }

    if (panelMode === 'quick' && streamedText && !isSaved) {
      setPendingLeaveAction(() => nextAction);
      setShowSaveOnLeave(true);
    } else {
      nextAction();
    }
  };

  const handleConfirmSaveOnLeave = async () => {
    const ok = await handleSaveNote();
    if (ok) {
      setShowSaveOnLeave(false);
      if (pendingLeaveAction) pendingLeaveAction();
    }
  };

  const handleDiscardOnLeave = () => {
    if (sessionId && stageId) {
      const draftId = `draft_${sessionId}_${stageId}`;
      fetch(`/api/tutor/drafts/${draftId}`, { method: 'DELETE' }).catch(() => {});
    }
    setShowSaveOnLeave(false);
    if (pendingLeaveAction) pendingLeaveAction();
  };

  // Keyboard navigation & accessibility within tutor panel
  useEffect(() => {
    if (!isOpen) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        e.preventDefault();
        e.stopPropagation();
        if (showSaveOnLeave) {
          setShowSaveOnLeave(false);
          setPendingLeaveAction(null);
        } else {
          handleAttemptClose();
        }
      }
      if (showSaveOnLeave) {
        if (e.key === 'y' || e.key === 'Y') {
          e.preventDefault();
          handleConfirmSaveOnLeave();
        } else if (e.key === 'n' || e.key === 'N') {
          e.preventDefault();
          handleDiscardOnLeave();
        }
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, showSaveOnLeave, streamedText, isSaved]);

  if (!isOpen) return null;

  return (
    <div
      className="modal-overlay"
      style={{
        position: 'fixed',
        top: 0,
        left: 0,
        right: 0,
        bottom: 0,
        background: 'rgba(5, 8, 20, 0.75)',
        backdropFilter: 'blur(4px)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        zIndex: 1000,
        padding: '1rem',
      }}
      onClick={(e) => {
        if (e.target === e.currentTarget) {
          handleAttemptClose();
        }
      }}
    >
      <div
        ref={panelRef}
        id="tutor-panel"
        className="modal-card"
        style={{
          width: '100%',
          maxWidth: '820px',
          maxHeight: '90vh',
          display: 'flex',
          flexDirection: 'column',
          background: 'var(--bg-surface)',
          border: '1px solid var(--border-subtle)',
          borderRadius: 'var(--radius-lg)',
          boxShadow: '0 20px 40px -15px rgba(0, 0, 0, 0.5)',
          overflow: 'hidden',
        }}
      >
        {/* Header Bar */}
        <div
          style={{
            padding: '1rem 1.25rem',
            borderBottom: '1px solid var(--border-subtle)',
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            background: 'var(--bg-surface-elevated)',
            flexWrap: 'wrap',
            gap: '0.75rem',
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', flexWrap: 'wrap' }}>
            <h2 style={{ margin: 0, fontSize: '1.2rem', fontWeight: 700, color: 'var(--text-normal)' }}>
              🤖 AI Tutor
            </h2>

            {/* Mode Switcher Tabs */}
            <div
              style={{
                display: 'inline-flex',
                background: 'var(--bg-app)',
                padding: '2px',
                borderRadius: 'var(--radius-sm)',
                border: '1px solid var(--border-subtle)',
              }}
              role="tablist"
            >
              <button
                id="tab-tutor-quick"
                type="button"
                className={`btn ${panelMode === 'quick' ? 'btn-primary' : 'btn-ghost'}`}
                style={{ fontSize: '0.78rem', padding: '0.25rem 0.6rem' }}
                onClick={() => setPanelMode('quick')}
              >
                ⚡ Quick Guidance
              </button>
              <button
                id="tab-tutor-chat"
                type="button"
                className={`btn ${panelMode === 'chat' ? 'btn-primary' : 'btn-ghost'}`}
                style={{ fontSize: '0.78rem', padding: '0.25rem 0.6rem' }}
                onClick={() => setPanelMode('chat')}
              >
                💬 Interactive Chat
              </button>
            </div>

            {/* Provider Badge */}
            <span
              id="tutor-provider-badge"
              style={{
                fontSize: '0.75rem',
                fontWeight: 600,
                padding: '0.2rem 0.5rem',
                borderRadius: 'var(--radius-sm)',
                background:
                  activeProvider.route === 'offline'
                    ? 'rgba(56, 189, 248, 0.15)'
                    : activeProvider.route === 'local'
                    ? 'rgba(16, 185, 129, 0.15)'
                    : activeProvider.route === 'gemini'
                    ? 'rgba(168, 85, 247, 0.15)'
                    : activeProvider.route === 'anthropic'
                    ? 'rgba(245, 158, 11, 0.15)'
                    : 'rgba(59, 130, 246, 0.15)',
                color:
                  activeProvider.route === 'offline'
                    ? 'var(--accent-cyan)'
                    : activeProvider.route === 'local'
                    ? 'var(--accent-emerald)'
                    : activeProvider.route === 'gemini'
                    ? '#c084fc'
                    : activeProvider.route === 'anthropic'
                    ? 'var(--accent-amber)'
                    : '#60a5fa',
              }}
            >
              {activeProvider.route === 'offline'
                ? 'Offline Reviewed'
                : activeProvider.route === 'local'
                ? `Local LLM (${activeProvider.model})`
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
            aria-label="Close (Esc)"
          >
            ✕
          </button>
        </div>

        {/* Advisory / Socratic Status Banner */}
        <div
          id="tutor-advisory-banner"
          style={{
            margin: '0.65rem 1.25rem 0',
            padding: '0.5rem 0.85rem',
            background: stageResolved
              ? 'rgba(16, 185, 129, 0.1)'
              : 'rgba(251, 191, 36, 0.1)',
            border: `1px solid ${stageResolved ? 'rgba(16, 185, 129, 0.3)' : 'rgba(251, 191, 36, 0.3)'}`,
            borderRadius: 'var(--radius-sm)',
            fontSize: '0.8rem',
            color: stageResolved ? 'var(--accent-emerald)' : 'var(--accent-amber)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            flexWrap: 'wrap',
            gap: '0.5rem',
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            <span>{stageResolved ? '✓' : '🛡️'}</span>
            <span>
              <strong>Advisory AI explanation:</strong>{' '}
              {stageResolved ? (
                <span>Stage Completed. Full derivations and Excel/R formula breakdowns are available.</span>
              ) : (
                <span>Socratic Guardrail Active. Working on active stage. The tutor provides guiding questions & intuition without revealing final answers.</span>
              )}
            </span>
          </div>
          <span style={{ fontSize: '0.72rem', opacity: 0.85 }}>
            Self-study aid · Does not alter grades
          </span>
        </div>

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

        {/* MODE 1: QUICK GUIDANCE (Single-turn hint/explain/follow-up) */}
        {panelMode === 'quick' && (
          <>
            {/* Action Prompt Strip */}
            <div
              style={{
                padding: '0.75rem 1.25rem',
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

            {/* Quick Content Body */}
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
                className="input"
                placeholder="Ask a specific question about this step... (e.g., Why do variances add for independent variables?)"
                value={customPrompt}
                onChange={(e) => setCustomPrompt(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter' && customPrompt.trim() && !isStreaming) {
                    handleStartRequest('follow_up', 'custom', customPrompt.trim());
                    setCustomPrompt('');
                  }
                }}
                disabled={isStreaming}
                style={{ flex: 1 }}
              />
              <button
                id="btn-tutor-send-custom"
                className="btn btn-primary"
                disabled={isStreaming || !customPrompt.trim()}
                onClick={() => {
                  handleStartRequest('follow_up', 'custom', customPrompt.trim());
                  setCustomPrompt('');
                }}
              >
                Send
              </button>
            </div>
          </>
        )}

        {/* MODE 2: INTERACTIVE MULTI-TURN CHAT */}
        {panelMode === 'chat' && (
          <>
            {/* Thread selector & Header */}
            <div
              style={{
                padding: '0.6rem 1.25rem',
                background: 'var(--bg-surface)',
                borderBottom: '1px solid var(--border-subtle)',
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                flexWrap: 'wrap',
                gap: '0.5rem',
              }}
            >
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', flexWrap: 'wrap' }}>
                <span style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>Discussion Thread:</span>
                {threads.length > 0 ? (
                  <select
                    className="select"
                    id="tutor-thread-select"
                    value={activeThread?.id || ''}
                    onChange={async (e) => {
                      const thId = e.target.value;
                      const res = await fetch(`/api/tutor/threads/${thId}`);
                      if (res.ok) {
                        const fullTh: TutorThreadDTO = await res.json();
                        setActiveThread(fullTh);
                        setChatMessages(fullTh.messages || []);
                      }
                    }}
                    style={{ fontSize: '0.8rem', padding: '0.25rem 0.5rem', maxWidth: '280px' }}
                  >
                    {threads.map((t) => (
                      <option key={t.id} value={t.id}>
                        {t.title}
                      </option>
                    ))}
                  </select>
                ) : (
                  <span style={{ fontSize: '0.8rem', fontStyle: 'italic', color: 'var(--text-muted)' }}>
                    New conversation
                  </span>
                )}
              </div>

              <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
                <button
                  type="button"
                  id="btn-new-chat-thread"
                  className="btn btn-outline"
                  onClick={() => {
                    setActiveThread(null);
                    setChatMessages([]);
                    setChatInput('');
                    setChatStreamingText('');
                  }}
                  style={{ fontSize: '0.75rem', padding: '0.25rem 0.55rem' }}
                >
                  + New Chat
                </button>

                {chatMessages.length > 0 && (
                  <button
                    type="button"
                    id="btn-save-chat-transcript"
                    className="btn btn-secondary"
                    onClick={handleSaveWholeConversationAsNote}
                    style={{ fontSize: '0.75rem', padding: '0.25rem 0.55rem' }}
                  >
                    💾 Save Full Chat
                  </button>
                )}
              </div>
            </div>

            {/* Notification alert */}
            {chatNoteSavedMsg && (
              <div
                style={{
                  margin: '0.5rem 1.25rem 0',
                  padding: '0.4rem 0.75rem',
                  background: 'rgba(16, 185, 129, 0.15)',
                  border: '1px solid var(--accent-emerald)',
                  borderRadius: 'var(--radius-sm)',
                  color: 'var(--accent-emerald)',
                  fontSize: '0.8rem',
                }}
              >
                {chatNoteSavedMsg}
              </div>
            )}

            {/* Chat Messages Feed */}
            <div
              id="tutor-chat-feed"
              style={{
                flex: 1,
                padding: '1.25rem',
                overflowY: 'auto',
                background: 'var(--bg-app)',
                minHeight: '280px',
                maxHeight: '420px',
                display: 'flex',
                flexDirection: 'column',
                gap: '1rem',
              }}
            >
              {chatMessages.length === 0 && !isChatStreaming && (
                <div style={{ textAlign: 'center', color: 'var(--text-muted)', margin: 'auto', padding: '1.5rem' }}>
                  <p style={{ fontSize: '1.05rem', fontWeight: 600, color: 'var(--text-normal)', marginBottom: '0.4rem' }}>
                    Interactive Statistical Dialogue
                  </p>
                  <p style={{ fontSize: '0.85rem', maxWidth: '440px', margin: '0 auto 1rem' }}>
                    Ask questions back and forth. All conversation turns are stored persistently in your local offline SQLite database.
                  </p>
                  <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.4rem', justifyContent: 'center' }}>
                    {[
                      'What concept applies here?',
                      'Can you give an analogous example?',
                      'Why does this assumption matter?',
                      'What formula should I use?',
                    ].map((chip) => (
                      <button
                        key={chip}
                        type="button"
                        className="btn btn-outline"
                        style={{ fontSize: '0.75rem', padding: '0.25rem 0.55rem' }}
                        onClick={() => handleSendChatMessage(chip)}
                      >
                        {chip}
                      </button>
                    ))}
                  </div>
                </div>
              )}

              {/* Message bubbles */}
              {chatMessages.map((msg, idx) => (
                <div
                  key={msg.id || idx}
                  style={{
                    display: 'flex',
                    flexDirection: 'column',
                    alignItems: msg.role === 'user' ? 'flex-end' : 'flex-start',
                    maxWidth: '100%',
                  }}
                >
                  <div
                    style={{
                      fontSize: '0.72rem',
                      color: 'var(--text-muted)',
                      marginBottom: '0.25rem',
                      display: 'flex',
                      alignItems: 'center',
                      gap: '0.4rem',
                    }}
                  >
                    <span>{msg.role === 'user' ? '👤 Learner' : '🤖 AI Tutor'}</span>
                    {msg.fallback_label && (
                      <span style={{ color: 'var(--accent-cyan)' }}>({msg.fallback_label})</span>
                    )}
                  </div>

                  <div
                    style={{
                      maxWidth: '85%',
                      padding: '0.75rem 1rem',
                      borderRadius:
                        msg.role === 'user'
                          ? 'var(--radius-md) var(--radius-md) 0 var(--radius-md)'
                          : 'var(--radius-md) var(--radius-md) var(--radius-md) 0',
                      background:
                        msg.role === 'user'
                          ? 'var(--accent-blue-subtle, #1e3a8a)'
                          : 'var(--bg-surface-elevated)',
                      border: '1px solid var(--border-subtle)',
                      color: 'var(--text-normal)',
                      fontSize: '0.9rem',
                    }}
                  >
                    <MathMarkdown content={msg.content} />

                    {msg.role === 'assistant' && (
                      <div style={{ marginTop: '0.5rem', display: 'flex', justifyContent: 'flex-end' }}>
                        <button
                          type="button"
                          className="btn btn-ghost"
                          style={{ fontSize: '0.7rem', padding: '0.15rem 0.4rem', color: 'var(--text-muted)' }}
                          onClick={() => handleSaveChatMessageAsNote(msg)}
                          title="Save this specific response as a personal note"
                        >
                          📌 Save Note
                        </button>
                      </div>
                    )}
                  </div>
                </div>
              ))}

              {/* Live Streaming Delta Bubble */}
              {isChatStreaming && (
                <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-start', maxWidth: '100%' }}>
                  <div style={{ fontSize: '0.72rem', color: 'var(--accent-cyan)', marginBottom: '0.25rem' }}>
                    🤖 AI Tutor (typing...)
                  </div>
                  <div
                    style={{
                      maxWidth: '85%',
                      padding: '0.75rem 1rem',
                      borderRadius: 'var(--radius-md) var(--radius-md) var(--radius-md) 0',
                      background: 'var(--bg-surface-elevated)',
                      border: '1px solid var(--accent-cyan)',
                      color: 'var(--text-normal)',
                      fontSize: '0.9rem',
                    }}
                  >
                    <MathMarkdown content={chatStreamingText || 'Thinking...'} />
                  </div>
                </div>
              )}

              {chatError && (
                <div
                  style={{
                    padding: '0.5rem 0.75rem',
                    background: 'rgba(244, 63, 94, 0.1)',
                    border: '1px solid var(--accent-rose)',
                    borderRadius: 'var(--radius-sm)',
                    color: 'var(--accent-rose)',
                    fontSize: '0.8rem',
                  }}
                >
                  ⚠️ {chatError}
                </div>
              )}

              <div ref={chatEndRef} />
            </div>

            {/* Chat Input & Sockets */}
            <div
              style={{
                padding: '0.75rem 1.25rem',
                background: 'var(--bg-surface-elevated)',
                borderTop: '1px solid var(--border-subtle)',
                display: 'flex',
                gap: '0.5rem',
              }}
            >
              <textarea
                id="tutor-chat-input"
                className="input"
                placeholder={
                  stageResolved
                    ? 'Ask a question or request a complete formula breakdown... (Enter to send)'
                    : 'Ask about intuition, conditions, or test your reasoning... (Enter to send)'
                }
                value={chatInput}
                onChange={(e) => setChatInput(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter' && !e.shiftKey) {
                    e.preventDefault();
                    if (chatInput.trim() && !isChatStreaming) {
                      handleSendChatMessage();
                    }
                  }
                }}
                disabled={isChatStreaming}
                rows={2}
                style={{ flex: 1, resize: 'none' }}
              />
              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.35rem' }}>
                {isChatStreaming ? (
                  <button
                    id="btn-tutor-stop-chat"
                    type="button"
                    className="btn btn-outline"
                    onClick={handleCancelChatStreaming}
                    style={{ flex: 1, color: 'var(--accent-rose)', borderColor: 'var(--accent-rose)' }}
                  >
                    ⏹ Stop
                  </button>
                ) : (
                  <button
                    id="btn-tutor-send-chat"
                    type="button"
                    className="btn btn-primary"
                    disabled={!chatInput.trim()}
                    onClick={() => handleSendChatMessage()}
                    style={{ flex: 1 }}
                  >
                    Send ↵
                  </button>
                )}
              </div>
            </div>
          </>
        )}

        {/* Footer Actions (Quick mode save button & leave intent) */}
        {panelMode === 'quick' && (
          <div
            style={{
              padding: '0.75rem 1.25rem',
              background: 'var(--bg-surface)',
              borderTop: '1px solid var(--border-subtle)',
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
              flexWrap: 'wrap',
              gap: '0.5rem',
            }}
          >
            <div>
              {saveSuccess && (
                <span id="tutor-save-success-msg" style={{ color: 'var(--accent-emerald)', fontSize: '0.85rem' }}>
                  ✓ Saved to Personal Notes Library (V to view)
                </span>
              )}
              {saveError && (
                <span style={{ color: 'var(--accent-rose)', fontSize: '0.85rem' }}>
                  ⚠️ {saveError}
                </span>
              )}
            </div>

            <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
              {isStreaming ? (
                <button
                  id="btn-tutor-cancel"
                  className="btn btn-outline"
                  onClick={handleCancelStreaming}
                  style={{ color: 'var(--accent-rose)', borderColor: 'var(--accent-rose)' }}
                >
                  ⏹ Stop Generation
                </button>
              ) : (
                <button
                  id="btn-tutor-save"
                  className="btn btn-secondary"
                  disabled={!streamedText || isSaved}
                  onClick={handleSaveNote}
                >
                  {isSaved ? '✓ Note Saved' : '💾 Save to Notes'}
                </button>
              )}
              <button className="btn btn-outline" onClick={() => handleAttemptClose()}>
                Close
              </button>
            </div>
          </div>
        )}

        {panelMode === 'chat' && (
          <div
            style={{
              padding: '0.5rem 1.25rem',
              background: 'var(--bg-surface)',
              borderTop: '1px solid var(--border-subtle)',
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
              fontSize: '0.78rem',
              color: 'var(--text-muted)',
            }}
          >
            <span>💬 Persistent SQLite thread · MathJax formulas enabled</span>
            <button className="btn btn-outline" style={{ fontSize: '0.75rem', padding: '0.2rem 0.5rem' }} onClick={() => handleAttemptClose()}>
              Close
            </button>
          </div>
        )}

        {/* Save on Leave Confirmation Modal */}
        {showSaveOnLeave && (
          <div
            id="tutor-leave-modal"
            style={{
              position: 'absolute',
              top: 0,
              left: 0,
              right: 0,
              bottom: 0,
              background: 'rgba(5, 8, 20, 0.85)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              padding: '1rem',
              zIndex: 10,
            }}
          >
            <div
              style={{
                background: 'var(--bg-surface)',
                border: '1px solid var(--border-subtle)',
                borderRadius: 'var(--radius-md)',
                padding: '1.25rem',
                maxWidth: '420px',
                width: '100%',
                display: 'flex',
                flexDirection: 'column',
                gap: '1rem',
                boxShadow: '0 10px 25px rgba(0, 0, 0, 0.5)',
              }}
            >
              <h3 style={{ margin: 0, fontSize: '1.1rem', color: 'var(--text-normal)' }}>
                Save Explanation Before Leaving?
              </h3>
              <p style={{ margin: 0, fontSize: '0.85rem', color: 'var(--text-muted)' }}>
                You have an unsaved explanation. Would you like to save it to your local Note Library?
              </p>
              <div style={{ display: 'flex', gap: '0.5rem', justifyContent: 'flex-end', flexWrap: 'wrap' }}>
                <button
                  id="btn-tutor-leave-save"
                  className="btn btn-primary"
                  onClick={handleConfirmSaveOnLeave}
                >
                  <kbd className="kbd">y</kbd> Save Note
                </button>
                <button
                  id="btn-tutor-leave-discard"
                  className="btn btn-secondary"
                  onClick={handleDiscardOnLeave}
                >
                  <kbd className="kbd">n</kbd> Discard
                </button>
                <button
                  id="btn-tutor-leave-stay"
                  className="btn btn-outline"
                  aria-label="(Esc) Stay"
                  onClick={() => {
                    setShowSaveOnLeave(false);
                    setPendingLeaveAction(null);
                  }}
                >
                  <kbd className="kbd">Esc</kbd> Stay
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

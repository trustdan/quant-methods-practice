import React, { useState, useEffect, useRef, useCallback } from 'react';
import { TOUR_STEPS, TourStep, AppTab } from './tourData';

export const STORAGE_KEY_TOUR_DISMISSED = 'quant_tour_dismissed_v1';

export function isTourDismissed(): boolean {
  try {
    if (typeof window !== 'undefined') {
      if (window.location.search.includes('enable_tour=1')) {
        return false;
      }
      if (window.location.search.includes('disable_tour=1') || window.navigator?.webdriver) {
        return true;
      }
    }
    return localStorage.getItem(STORAGE_KEY_TOUR_DISMISSED) === 'true';
  } catch {
    return false;
  }
}

export function setTourDismissed(dismissed: boolean): void {
  try {
    if (dismissed) {
      localStorage.setItem(STORAGE_KEY_TOUR_DISMISSED, 'true');
    } else {
      localStorage.removeItem(STORAGE_KEY_TOUR_DISMISSED);
    }
  } catch {
    // Ignore storage quota/permission exceptions
  }
}

export async function persistTourDismissed(dismissed: boolean): Promise<void> {
  setTourDismissed(dismissed);
  try {
    const res = await fetch('/api/settings');
    let current: Record<string, unknown> = {};
    if (res.ok) {
      current = await res.json();
    }
    await fetch('/api/settings', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        ...current,
        tour_dismissed: dismissed,
      }),
    });
  } catch {
    // Ignore fetch errors in offline or headless mode
  }
}

export interface ProductTourModalProps {
  isOpen: boolean;
  onClose: () => void;
  onNavigateTab?: (tab: AppTab) => void;
  onOpenTutor?: () => void;
}

interface BubbleLayout {
  top: number;
  left: number;
  arrowPlacement: 'top' | 'bottom' | 'left' | 'right' | 'center';
  arrowOffsetPercent: number;
}

export const ProductTourModal: React.FC<ProductTourModalProps> = ({
  isOpen,
  onClose,
  onNavigateTab,
  onOpenTutor,
}) => {
  const [currentStepIndex, setCurrentStepIndex] = useState(0);
  const [dontShowAgain, setDontShowAgain] = useState(true);
  const [targetRect, setTargetRect] = useState<DOMRect | null>(null);
  const [bubbleLayout, setBubbleLayout] = useState<BubbleLayout | null>(null);

  const bubbleRef = useRef<HTMLDivElement>(null);
  const previouslyFocusedRef = useRef<HTMLElement | null>(null);

  const step: TourStep = TOUR_STEPS[currentStepIndex] || TOUR_STEPS[0];
  const isFirstStep = currentStepIndex === 0;
  const isLastStep = currentStepIndex === TOUR_STEPS.length - 1;
  const progressPercent = ((currentStepIndex + 1) / TOUR_STEPS.length) * 100;

  // Reposition bubble relative to target element
  const updatePosition = useCallback(() => {
    if (!isOpen || typeof document === 'undefined') return;

    const el = document.querySelector(step.targetSelector) as HTMLElement | null;
    if (!el) {
      setTargetRect(null);
      setBubbleLayout({
        top: Math.max(20, Math.floor(window.innerHeight / 2 - 200)),
        left: Math.max(16, Math.floor(window.innerWidth / 2 - 220)),
        arrowPlacement: 'center',
        arrowOffsetPercent: 50,
      });
      return;
    }

    const rect = el.getBoundingClientRect();
    setTargetRect(rect);

    const bubbleWidth = Math.min(440, window.innerWidth - 32);
    const bubbleHeight = 320;
    const padding = 16;
    const arrowMargin = 12;

    const preferred = step.placement || 'bottom';
    let top = 0;
    let left = 0;
    let placement: 'top' | 'bottom' | 'left' | 'right' = 'bottom';

    if (preferred === 'bottom') {
      if (rect.bottom + bubbleHeight + padding <= window.innerHeight) {
        top = rect.bottom + arrowMargin;
        placement = 'bottom';
      } else if (rect.top - bubbleHeight - arrowMargin >= padding) {
        top = rect.top - bubbleHeight - arrowMargin;
        placement = 'top';
      } else {
        top = Math.max(padding, window.innerHeight - bubbleHeight - padding);
        placement = 'bottom';
      }
      left = rect.left + rect.width / 2 - bubbleWidth / 2;
    } else if (preferred === 'top') {
      if (rect.top - bubbleHeight - arrowMargin >= padding) {
        top = rect.top - bubbleHeight - arrowMargin;
        placement = 'top';
      } else {
        top = rect.bottom + arrowMargin;
        placement = 'bottom';
      }
      left = rect.left + rect.width / 2 - bubbleWidth / 2;
    } else if (preferred === 'right') {
      if (rect.right + bubbleWidth + padding <= window.innerWidth) {
        left = rect.right + arrowMargin;
        placement = 'right';
      } else {
        left = rect.left - bubbleWidth - arrowMargin;
        placement = 'left';
      }
      top = rect.top + rect.height / 2 - bubbleHeight / 2;
    } else {
      left = rect.left + rect.width / 2 - bubbleWidth / 2;
      top = rect.bottom + arrowMargin;
      placement = 'bottom';
    }

    // Clamp horizontally to avoid off-screen overflow
    const clampedLeft = Math.max(padding, Math.min(window.innerWidth - bubbleWidth - padding, left));
    const clampedTop = Math.max(padding, Math.min(window.innerHeight - bubbleHeight - padding, top));

    // Calculate arrow offset relative to the bubble
    const targetCenterX = rect.left + rect.width / 2;
    const relativeX = targetCenterX - clampedLeft;
    const arrowOffsetPercent = Math.max(10, Math.min(90, (relativeX / bubbleWidth) * 100));

    setBubbleLayout({
      top: clampedTop,
      left: clampedLeft,
      arrowPlacement: placement,
      arrowOffsetPercent,
    });
  }, [isOpen, step]);

  // Navigate to step tab and position bubble
  useEffect(() => {
    if (!isOpen) return;

    if (step.tabTarget && step.tabTarget !== 'tutor') {
      onNavigateTab?.(step.tabTarget);
    }

    // Scroll into view if needed
    const timer1 = setTimeout(() => {
      const el = document.querySelector(step.targetSelector) as HTMLElement | null;
      if (el) {
        el.scrollIntoView({ behavior: 'smooth', block: 'nearest', inline: 'nearest' });
      }
      updatePosition();
    }, 40);

    const timer2 = setTimeout(() => {
      updatePosition();
    }, 180);

    return () => {
      clearTimeout(timer1);
      clearTimeout(timer2);
    };
  }, [isOpen, step, onNavigateTab, updatePosition]);

  // Manage focus trapping and window listeners
  useEffect(() => {
    if (isOpen) {
      previouslyFocusedRef.current = document.activeElement as HTMLElement | null;
      bubbleRef.current?.focus();
      window.addEventListener('resize', updatePosition);
      window.addEventListener('scroll', updatePosition, true);
    } else {
      if (previouslyFocusedRef.current) {
        previouslyFocusedRef.current.focus();
      }
    }
    return () => {
      window.removeEventListener('resize', updatePosition);
      window.removeEventListener('scroll', updatePosition, true);
    };
  }, [isOpen, updatePosition]);

  // Reset step on new open
  useEffect(() => {
    if (isOpen) {
      setCurrentStepIndex(0);
      setDontShowAgain(true);
    }
  }, [isOpen]);

  const handleNext = useCallback(() => {
    if (isLastStep) {
      if (dontShowAgain) {
        void persistTourDismissed(true);
      }
      onClose();
    } else {
      setCurrentStepIndex((prev) => Math.min(prev + 1, TOUR_STEPS.length - 1));
    }
  }, [isLastStep, dontShowAgain, onClose]);

  const handlePrev = useCallback(() => {
    setCurrentStepIndex((prev) => Math.max(prev - 1, 0));
  }, []);

  const handleSkip = useCallback(() => {
    if (dontShowAgain) {
      void persistTourDismissed(true);
    }
    onClose();
  }, [dontShowAgain, onClose]);

  const handleDismissForever = useCallback(() => {
    void persistTourDismissed(true);
    onClose();
  }, [onClose]);

  const handleJumpFeature = useCallback(() => {
    if (step.id === 'tutor' && onOpenTutor) {
      onOpenTutor();
    } else if (step.tabTarget && step.tabTarget !== 'tutor' && onNavigateTab) {
      onNavigateTab(step.tabTarget);
    }
    void persistTourDismissed(true);
    onClose();
  }, [step, onNavigateTab, onOpenTutor, onClose]);

  // Keyboard navigation
  useEffect(() => {
    if (!isOpen) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'ArrowRight' || e.key === 'Enter') {
        e.preventDefault();
        handleNext();
      } else if (e.key === 'ArrowLeft') {
        e.preventDefault();
        handlePrev();
      } else if (e.key === 'Escape') {
        e.preventDefault();
        if (dontShowAgain) {
          void persistTourDismissed(true);
        }
        onClose();
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, handleNext, handlePrev, dontShowAgain, onClose]);

  if (!isOpen) return null;

  return (
    <div
      role="dialog"
      aria-modal="true"
      aria-label="Quant Methods Product Tour"
      className="tour-modal-backdrop"
      style={{
        position: 'fixed',
        top: 0,
        left: 0,
        right: 0,
        bottom: 0,
        zIndex: 1040,
        pointerEvents: 'auto',
      }}
      onClick={(e) => {
        if (e.target === e.currentTarget) {
          handleSkip();
        }
      }}
    >
      {/* Target Element Spotlight Highlight Ring */}
      {targetRect && targetRect.width > 0 && targetRect.height > 0 ? (
        <div
          className="tour-spotlight-highlight"
          style={{
            top: Math.max(0, targetRect.top - 6),
            left: Math.max(0, targetRect.left - 6),
            width: targetRect.width + 12,
            height: targetRect.height + 12,
          }}
        />
      ) : (
        <div className="tour-spotlight-fallback" />
      )}

      {/* Anchored Callout Speech Bubble */}
      <div
        ref={bubbleRef}
        tabIndex={-1}
        className="tour-callout-bubble"
        style={{
          top: bubbleLayout ? `${bubbleLayout.top}px` : '20%',
          left: bubbleLayout ? `${bubbleLayout.left}px` : '50%',
          transform: bubbleLayout ? 'none' : 'translateX(-50%)',
        }}
        onClick={(e) => e.stopPropagation()}
      >
        {/* Callout Arrow Pointer */}
        {bubbleLayout && bubbleLayout.arrowPlacement !== 'center' && (
          <div
            className={`tour-bubble-arrow arrow-${bubbleLayout.arrowPlacement === 'bottom' ? 'top' : bubbleLayout.arrowPlacement === 'top' ? 'bottom' : bubbleLayout.arrowPlacement === 'right' ? 'left' : 'right'}`}
            style={{
              left:
                bubbleLayout.arrowPlacement === 'top' || bubbleLayout.arrowPlacement === 'bottom'
                  ? `${bubbleLayout.arrowOffsetPercent}%`
                  : undefined,
              top:
                bubbleLayout.arrowPlacement === 'left' || bubbleLayout.arrowPlacement === 'right'
                  ? '50%'
                  : undefined,
            }}
          />
        )}

        {/* Progress Bar */}
        <div
          style={{
            position: 'absolute',
            top: 0,
            left: 0,
            right: 0,
            height: '4px',
            background: 'rgba(255, 255, 255, 0.08)',
            borderTopLeftRadius: 'var(--radius-lg)',
            borderTopRightRadius: 'var(--radius-lg)',
            overflow: 'hidden',
          }}
        >
          <div
            style={{
              height: '100%',
              width: `${progressPercent}%`,
              background: 'linear-gradient(90deg, #38bdf8 0%, #818cf8 50%, #34d399 100%)',
              transition: 'width 0.25s ease-out',
            }}
          />
        </div>

        {/* Header */}
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            gap: '0.5rem',
            marginBottom: '0.75rem',
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', flexWrap: 'wrap' }}>
            <span
              className="badge"
              style={{
                background: step.badgeColor || 'var(--accent-cyan)',
                color: '#0f172a',
                fontWeight: 700,
                fontSize: '0.72rem',
                padding: '0.15rem 0.5rem',
              }}
            >
              {step.badge}
            </span>
            <span
              style={{
                fontSize: '0.75rem',
                color: 'var(--text-muted)',
                fontWeight: 600,
              }}
            >
              Step {currentStepIndex + 1} of {TOUR_STEPS.length}
            </span>
          </div>

          <button
            type="button"
            className="btn btn-outline"
            onClick={handleSkip}
            title="Dismiss Tour (Esc)"
            aria-label="Close Tour"
            style={{
              padding: '0.2rem 0.5rem',
              fontSize: '0.8rem',
              lineHeight: 1,
              border: 'none',
              color: 'var(--text-muted)',
            }}
          >
            ✕
          </button>
        </div>

        {/* Title & Icon */}
        <div style={{ display: 'flex', alignItems: 'flex-start', gap: '0.6rem', marginBottom: '0.4rem' }}>
          <span style={{ fontSize: '1.4rem', lineHeight: 1 }}>{step.icon}</span>
          <div>
            <h3
              style={{
                fontSize: '1.15rem',
                fontWeight: 700,
                color: 'var(--text-main)',
                margin: 0,
                letterSpacing: '-0.01em',
              }}
            >
              {step.title}
            </h3>
            <p
              style={{
                fontSize: '0.85rem',
                color: 'var(--text-muted)',
                margin: '0.25rem 0 0 0',
                lineHeight: 1.35,
              }}
            >
              {step.subtitle}
            </p>
          </div>
        </div>

        {/* Feature Highlights Grid */}
        <div
          style={{
            display: 'flex',
            flexDirection: 'column',
            gap: '0.45rem',
            margin: '0.75rem 0',
            maxHeight: '180px',
            overflowY: 'auto',
          }}
        >
          {step.highlights.map((h, idx) => (
            <div
              key={idx}
              style={{
                display: 'flex',
                alignItems: 'flex-start',
                gap: '0.5rem',
                background: 'rgba(255, 255, 255, 0.03)',
                border: '1px solid rgba(255, 255, 255, 0.06)',
                borderRadius: 'var(--radius-sm)',
                padding: '0.4rem 0.6rem',
              }}
            >
              <span style={{ fontSize: '1rem', marginTop: '0.1rem' }}>{h.icon}</span>
              <div style={{ flex: 1, fontSize: '0.8rem', lineHeight: 1.3 }}>
                <div style={{ fontWeight: 600, color: 'var(--text-main)' }}>
                  {h.title}
                </div>
                <div style={{ color: 'var(--text-muted)', marginTop: '0.1rem' }}>
                  {h.description}
                </div>
              </div>
            </div>
          ))}
        </div>

        {/* Keyboard Shortcut Hint */}
        {step.shortcutHint && (
          <div
            style={{
              fontSize: '0.75rem',
              color: 'var(--accent-cyan)',
              background: 'rgba(56, 189, 248, 0.08)',
              padding: '0.3rem 0.6rem',
              borderRadius: 'var(--radius-sm)',
              marginBottom: '0.75rem',
              display: 'flex',
              alignItems: 'center',
              gap: '0.4rem',
            }}
          >
            <span>💡</span>
            <span>{step.shortcutHint}</span>
          </div>
        )}

        {/* Feature Jump Button if available */}
        {step.jumpLabel && (
          <div style={{ marginBottom: '0.75rem' }}>
            <button
              type="button"
              className="btn btn-outline"
              onClick={handleJumpFeature}
              style={{
                width: '100%',
                fontSize: '0.8rem',
                padding: '0.35rem 0.75rem',
                borderColor: 'rgba(56, 189, 248, 0.35)',
                color: 'var(--accent-cyan)',
              }}
            >
              {step.jumpLabel} ↗
            </button>
          </div>
        )}

        {/* Step Pips */}
        <div
          style={{
            display: 'flex',
            justifyContent: 'center',
            alignItems: 'center',
            gap: '0.35rem',
            marginBottom: '0.75rem',
          }}
        >
          {TOUR_STEPS.map((s, idx) => (
            <button
              key={s.id}
              type="button"
              aria-label={`Jump to tour step ${idx + 1}: ${s.title}`}
              onClick={() => setCurrentStepIndex(idx)}
              style={{
                width: idx === currentStepIndex ? '20px' : '6px',
                height: '6px',
                borderRadius: '3px',
                background:
                  idx === currentStepIndex
                    ? 'var(--accent-cyan)'
                    : 'rgba(255, 255, 255, 0.2)',
                border: 'none',
                cursor: 'pointer',
                padding: 0,
                transition: 'all 0.2s ease',
              }}
            />
          ))}
        </div>

        {/* Footer Actions */}
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            borderTop: '1px solid rgba(255, 255, 255, 0.08)',
            paddingTop: '0.65rem',
            flexWrap: 'wrap',
            gap: '0.5rem',
          }}
        >
          {/* Checkbox preference */}
          <label
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '0.4rem',
              fontSize: '0.75rem',
              color: 'var(--text-muted)',
              cursor: 'pointer',
              userSelect: 'none',
            }}
          >
            <input
              type="checkbox"
              checked={dontShowAgain}
              onChange={(e) => setDontShowAgain(e.target.checked)}
              aria-label="Don't show again on startup"
            />
            Don't show again on startup
          </label>

          {/* Nav Controls */}
          <div style={{ display: 'flex', gap: '0.4rem', alignItems: 'center' }}>
            <button
              type="button"
              className="btn btn-outline"
              onClick={handleDismissForever}
              style={{ fontSize: '0.75rem', padding: '0.3rem 0.5rem' }}
            >
              Don't show again
            </button>
            <button
              type="button"
              className="btn btn-outline"
              onClick={handleSkip}
              style={{ fontSize: '0.75rem', padding: '0.3rem 0.5rem' }}
            >
              Skip Tour
            </button>
            <button
              type="button"
              className="btn btn-outline"
              disabled={isFirstStep}
              onClick={handlePrev}
              style={{ fontSize: '0.75rem', padding: '0.3rem 0.5rem' }}
            >
              &larr; Back
            </button>
            <button
              type="button"
              className="btn btn-primary"
              onClick={handleNext}
              style={{ fontSize: '0.75rem', padding: '0.3rem 0.65rem' }}
            >
              {isLastStep ? 'Get Started 🚀' : 'Next →'}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

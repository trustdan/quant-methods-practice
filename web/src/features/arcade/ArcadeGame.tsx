import React, { useEffect, useRef, useState, useCallback } from 'react';
import {
  ArcadeInput,
  ArcadeState,
  FIXED_DT,
  LOGICAL_HEIGHT,
  LOGICAL_WIDTH,
} from './types';
import { createInitialState, createRng, step, takeoverManual } from './simulation';
import { renderArcade } from './renderer';

interface ArcadeGameProps {
  onExit: () => void;
  seed?: number;
  initialMode?: 'attract' | 'manual';
}

export const ArcadeGame: React.FC<ArcadeGameProps> = ({
  onExit,
  seed = Date.now(),
  initialMode = 'attract',
}) => {
  const canvasRef = useRef<HTMLCanvasElement | null>(null);

  // Check system prefers-reduced-motion
  const [reducedMotion, setReducedMotion] = useState<boolean>(() => {
    if (typeof window !== 'undefined' && window.matchMedia) {
      return window.matchMedia('(prefers-reduced-motion: reduce)').matches;
    }
    return false;
  });

  const [gameState, setGameState] = useState<ArcadeState>(() =>
    createInitialState(seed, initialMode)
  );

  const inputRef = useRef<ArcadeInput>({
    up: false,
    down: false,
    left: false,
    right: false,
    fire: false,
  });

  const stateRef = useRef<ArcadeState>(gameState);
  stateRef.current = gameState;

  const rngRef = useRef<() => number>(createRng(seed));

  const clearHeldKeys = useCallback(() => {
    inputRef.current = {
      up: false,
      down: false,
      left: false,
      right: false,
      fire: false,
    };
  }, []);

  const handleTogglePause = useCallback(() => {
    clearHeldKeys();
    setGameState((prev) => ({
      ...prev,
      paused: !prev.paused,
    }));
  }, [clearHeldKeys]);

  // Window input listeners
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Exit commands immediately return to practice
      if (e.key === 'Escape' || e.key === 'Enter') {
        e.preventDefault();
        onExit();
        return;
      }

      // Pause toggle
      if (e.key === 'p' || e.key === 'P') {
        e.preventDefault();
        handleTogglePause();
        return;
      }

      // Movement & Fire
      let handled = false;
      const inp = inputRef.current;

      switch (e.code) {
        case 'ArrowUp':
        case 'KeyW':
        case 'KeyK':
          inp.up = true;
          handled = true;
          break;
        case 'ArrowDown':
        case 'KeyS':
        case 'KeyJ':
          inp.down = true;
          handled = true;
          break;
        case 'ArrowLeft':
        case 'KeyA':
          inp.left = true;
          handled = true;
          break;
        case 'ArrowRight':
        case 'KeyD':
          inp.right = true;
          handled = true;
          break;
        case 'Space':
        case 'KeyF':
          inp.fire = true;
          handled = true;
          break;
      }

      if (handled) {
        if (stateRef.current.mode === 'attract') {
          const next = takeoverManual(stateRef.current);
          stateRef.current = next;
          setGameState(next);
        }
        e.preventDefault();
      }
    };

    const handleKeyUp = (e: KeyboardEvent) => {
      const inp = inputRef.current;
      switch (e.code) {
        case 'ArrowUp':
        case 'KeyW':
        case 'KeyK':
          inp.up = false;
          break;
        case 'ArrowDown':
        case 'KeyS':
        case 'KeyJ':
          inp.down = false;
          break;
        case 'ArrowLeft':
        case 'KeyA':
          inp.left = false;
          break;
        case 'ArrowRight':
        case 'KeyD':
          inp.right = false;
          break;
        case 'Space':
        case 'KeyF':
          inp.fire = false;
          break;
      }
    };

    const handleBlur = () => {
      clearHeldKeys();
    };

    const handleVisibilityChange = () => {
      if (document.hidden) {
        clearHeldKeys();
        // Automatically pause manual game when backgrounded
        setGameState((prev) => (prev.mode === 'manual' ? { ...prev, paused: true } : prev));
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    window.addEventListener('keyup', handleKeyUp);
    window.addEventListener('blur', handleBlur);
    document.addEventListener('visibilitychange', handleVisibilityChange);

    return () => {
      window.removeEventListener('keydown', handleKeyDown);
      window.removeEventListener('keyup', handleKeyUp);
      window.removeEventListener('blur', handleBlur);
      document.removeEventListener('visibilitychange', handleVisibilityChange);
      clearHeldKeys();
    };
  }, [clearHeldKeys, handleTogglePause, onExit]);

  // Main fixed-step simulation & render loop
  useEffect(() => {
    let animId: number;
    let lastTime = performance.now();
    let accumulator = 0;

    const canvas = canvasRef.current;
    if (!canvas) return;

    let ctx: CanvasRenderingContext2D | null = null;
    try {
      ctx = canvas.getContext('2d');
    } catch {
      // jsdom fallback in test environments
    }
    if (!ctx) return;

    // Handle high-DPI crisp rendering
    const dpr = window.devicePixelRatio || 1;
    canvas.width = LOGICAL_WIDTH * dpr;
    canvas.height = LOGICAL_HEIGHT * dpr;
    ctx.scale(dpr, dpr);

    const frame = (now: number) => {
      const dtSec = Math.min(0.1, (now - lastTime) / 1000);
      lastTime = now;

      accumulator += dtSec;

      let s = stateRef.current;
      const rng = rngRef.current;

      while (accumulator >= FIXED_DT) {
        s = step(s, inputRef.current, FIXED_DT, rng, reducedMotion);
        accumulator -= FIXED_DT;
      }

      stateRef.current = s;
      setGameState(s);

      renderArcade(ctx, s, reducedMotion);

      animId = requestAnimationFrame(frame);
    };

    animId = requestAnimationFrame(frame);

    return () => {
      cancelAnimationFrame(animId);
    };
  }, [reducedMotion]);

  return (
    <div
      className="card"
      style={{
        maxWidth: 880,
        margin: '0 auto',
        padding: '1.25rem',
        display: 'flex',
        flexDirection: 'column',
        gap: '1rem',
      }}
    >
      {/* Top Header & Actions */}
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          flexWrap: 'wrap',
          gap: '0.75rem',
        }}
      >
        <div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.6rem' }}>
            <h2 style={{ margin: 0, fontSize: '1.4rem', fontWeight: 700, color: '#f8fafc' }}>
              🚀 Quant Flight
            </h2>
            <span
              className={`badge ${
                gameState.mode === 'attract' ? 'badge-amber' : 'badge-emerald'
              }`}
            >
              {gameState.mode === 'attract' ? '⚡ Attract Mode' : '🎮 Manual Flight'}
            </span>
          </div>
          <p style={{ margin: '0.25rem 0 0', fontSize: '0.85rem', color: '#94a3b8' }}>
            Original space arcade navigation drill. Scores remain strictly separate from practice mastery.
          </p>
        </div>

        <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
          <button
            id="btn-pause-arcade"
            className="btn btn-outline"
            onClick={handleTogglePause}
            style={{ fontSize: '0.85rem', padding: '0.35rem 0.75rem' }}
            title="Toggle pause (P)"
          >
            {gameState.paused ? '▶ Resume (P)' : '⏸ Pause (P)'}
          </button>

          <button
            id="btn-skip-arcade"
            className="btn btn-primary"
            onClick={onExit}
            style={{ fontSize: '0.85rem', padding: '0.35rem 0.85rem' }}
            title="Return to practice drill immediately (Esc / Enter)"
          >
            ⏭ Skip to Practice (Esc)
          </button>
        </div>
      </div>

      {/* Screen Reader accessible announcement region */}
      <div
        className="sr-only"
        aria-live="polite"
        id="arcade-status-announcement"
        style={{
          position: 'absolute',
          width: 1,
          height: 1,
          padding: 0,
          margin: -1,
          overflow: 'hidden',
          clip: 'rect(0, 0, 0, 0)',
          whiteSpace: 'nowrap',
          border: 0,
        }}
      >
        Quant Flight: Mode {gameState.mode}, Score {gameState.score}, Targets Cleared{' '}
        {gameState.targetsDestroyed}, Shields {gameState.ship.shields}%.
      </div>

      {/* Main Arcade Canvas Container */}
      <div
        style={{
          position: 'relative',
          width: '100%',
          aspectRatio: `${LOGICAL_WIDTH} / ${LOGICAL_HEIGHT}`,
          background: '#050814',
          borderRadius: 8,
          overflow: 'hidden',
          boxShadow: '0 8px 30px rgba(0, 0, 0, 0.45)',
          border: '1px solid #1e293b',
          cursor: gameState.mode === 'attract' ? 'pointer' : 'default',
        }}
        onClick={() => {
          if (stateRef.current.mode === 'attract') {
            const next = takeoverManual(stateRef.current);
            stateRef.current = next;
            setGameState(next);
          }
        }}
      >
        <canvas
          ref={canvasRef}
          id="arcade-canvas"
          tabIndex={0}
          aria-label="Quant Flight game canvas"
          style={{
            width: '100%',
            height: '100%',
            display: 'block',
            outline: 'none',
          }}
        />
      </div>

      {/* Footer Controls & Accessibility Options */}
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          flexWrap: 'wrap',
          gap: '0.75rem',
          fontSize: '0.85rem',
          color: '#94a3b8',
        }}
      >
        <div style={{ display: 'flex', gap: '1rem', flexWrap: 'wrap' }}>
          <span>
            <kbd className="kbd">Arrows</kbd> / <kbd className="kbd">WASD</kbd> /{' '}
            <kbd className="kbd">j</kbd>
            <kbd className="kbd">k</kbd> Steer
          </span>
          <span>
            <kbd className="kbd">Space</kbd> / <kbd className="kbd">F</kbd> Fire
          </span>
          <span>
            <kbd className="kbd">P</kbd> Pause
          </span>
          <span>
            <kbd className="kbd">Esc</kbd> / <kbd className="kbd">Enter</kbd> Skip
          </span>
        </div>

        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
          <label style={{ display: 'flex', alignItems: 'center', gap: '0.4rem', cursor: 'pointer' }}>
            <input
              type="checkbox"
              id="arcade-reduced-motion-toggle"
              checked={reducedMotion}
              onChange={(e) => setReducedMotion(e.target.checked)}
            />
            <span>Reduced Motion</span>
          </label>
        </div>
      </div>
    </div>
  );
};

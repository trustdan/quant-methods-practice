import React, { useEffect, useRef, useState, useCallback } from 'react';
import {
  ArcadeInput,
  ArcadeScoreEntry,
  ArcadeState,
  FIXED_DT,
  LOGICAL_HEIGHT,
  LOGICAL_WIDTH,
  SECTORS,
} from './types';
import { createInitialState, createRng, step, takeoverManual } from './simulation';
import { renderArcade } from './renderer';

interface ArcadeGameProps {
  onExit: () => void;
  seed?: number;
  initialMode?: 'attract' | 'manual';
  initialState?: ArcadeState;
}

const LOCAL_STORAGE_SCORES_KEY = 'quant_arcade_leaderboard';
const LOCAL_STORAGE_INITIALS_KEY = 'quant_pilot_initials';

export const ArcadeGame: React.FC<ArcadeGameProps> = ({
  onExit,
  seed = Date.now(),
  initialMode = 'attract',
  initialState,
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
    initialState ? { ...initialState } : createInitialState(seed, initialMode)
  );

  const [leaderboard, setLeaderboard] = useState<ArcadeScoreEntry[]>(() => {
    if (typeof window !== 'undefined') {
      try {
        const saved = localStorage.getItem(LOCAL_STORAGE_SCORES_KEY);
        if (saved) return JSON.parse(saved);
      } catch {
        // Fallback
      }
    }
    return [];
  });

  const [showLeaderboard, setShowLeaderboard] = useState(false);
  const [initials, setInitials] = useState<string>(() => {
    if (typeof window !== 'undefined') {
      try {
        return localStorage.getItem(LOCAL_STORAGE_INITIALS_KEY) || 'PIL';
      } catch {
        return 'PIL';
      }
    }
    return 'PIL';
  });
  const [activeSlot, setActiveSlot] = useState<number>(0);
  const [scoreSubmitted, setScoreSubmitted] = useState<boolean>(false);
  const [submittingScore, setSubmittingScore] = useState<boolean>(false);

  const inputRef = useRef<ArcadeInput>({
    up: false,
    down: false,
    left: false,
    right: false,
    fire: false,
    bomb: false,
  });

  const stateRef = useRef<ArcadeState>(gameState);
  stateRef.current = gameState;

  const rngRef = useRef<() => number>(createRng(seed));

  // Load scores from backend on mount
  useEffect(() => {
    let cancelled = false;
    const fetchScores = async () => {
      try {
        const res = await fetch('/api/arcade/scores');
        if (res.ok) {
          const data = await res.json();
          if (!cancelled && Array.isArray(data.scores)) {
            setLeaderboard((prev) => {
              const merged = [...data.scores, ...prev];
              // De-duplicate by score and timestamp/initials
              const unique: ArcadeScoreEntry[] = [];
              const seen = new Set<string>();
              for (const s of merged) {
                const key = `${s.score}-${s.initials}-${s.duration_seconds}`;
                if (!seen.has(key)) {
                  seen.add(key);
                  unique.push(s);
                }
              }
              unique.sort((a, b) => b.score - a.score);
              const top10 = unique.slice(0, 10);
              try {
                localStorage.setItem(LOCAL_STORAGE_SCORES_KEY, JSON.stringify(top10));
              } catch {
                // Ignore storage limits
              }
              return top10;
            });
          }
        }
      } catch {
        // Network unavailable, rely on localStorage
      }
    };
    fetchScores();
    return () => {
      cancelled = true;
    };
  }, []);

  const clearHeldKeys = useCallback(() => {
    inputRef.current = {
      up: false,
      down: false,
      left: false,
      right: false,
      fire: false,
      bomb: false,
    };
  }, []);

  const handleTogglePause = useCallback(() => {
    clearHeldKeys();
    setGameState((prev) => ({
      ...prev,
      paused: !prev.paused,
    }));
  }, [clearHeldKeys]);

  const handleRestart = useCallback(() => {
    clearHeldKeys();
    setScoreSubmitted(false);
    setShowLeaderboard(false);
    const next = takeoverManual({ ...createInitialState(Date.now(), 'attract'), mode: 'attract' });
    stateRef.current = next;
    setGameState(next);
  }, [clearHeldKeys]);

  // Window input listeners
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Exit commands immediately return to practice
      if (e.key === 'Escape') {
        e.preventDefault();
        onExit();
        return;
      }

      // Enter skips to practice if not entering initials on game over
      if (e.key === 'Enter' && (!stateRef.current.gameOver || scoreSubmitted || showLeaderboard)) {
        e.preventDefault();
        onExit();
        return;
      }

      // If Game Over and initials input active, let the modal handle input
      if (stateRef.current.gameOver && !scoreSubmitted && !showLeaderboard) {
        return;
      }

      // Pause toggle
      if (e.key === 'p' || e.key === 'P') {
        e.preventDefault();
        handleTogglePause();
        return;
      }

      // Movement, Fire, and Bombs
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
        case 'KeyH':
          inp.left = true;
          handled = true;
          break;
        case 'ArrowRight':
        case 'KeyD':
        case 'KeyL':
          inp.right = true;
          handled = true;
          break;
        case 'Space':
        case 'KeyF':
          inp.fire = true;
          handled = true;
          break;
        case 'KeyB':
        case 'KeyX':
          inp.bomb = true;
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
        case 'KeyH':
          inp.left = false;
          break;
        case 'ArrowRight':
        case 'KeyD':
        case 'KeyL':
          inp.right = false;
          break;
        case 'Space':
        case 'KeyF':
          inp.fire = false;
          break;
        case 'KeyB':
        case 'KeyX':
          inp.bomb = false;
          break;
      }
    };

    const handleBlur = () => {
      clearHeldKeys();
    };

    const handleVisibilityChange = () => {
      if (document.hidden) {
        clearHeldKeys();
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
  }, [clearHeldKeys, handleTogglePause, onExit, scoreSubmitted, showLeaderboard]);

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

  // Initials manipulation handlers for arcade score entry
  const changeInitialLetter = (slotIndex: number, delta: number) => {
    const chars = initials.padEnd(3, 'A').split('');
    const currentCode = chars[slotIndex].charCodeAt(0);
    // Cycle A (65) to Z (90)
    let nextCode = currentCode + delta;
    if (nextCode < 65) nextCode = 90;
    if (nextCode > 90) nextCode = 65;
    chars[slotIndex] = String.fromCharCode(nextCode);
    setInitials(chars.join('').slice(0, 3));
  };

  const handleInitialsKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'ArrowUp') {
      e.preventDefault();
      changeInitialLetter(activeSlot, 1);
    } else if (e.key === 'ArrowDown') {
      e.preventDefault();
      changeInitialLetter(activeSlot, -1);
    } else if (e.key === 'ArrowLeft') {
      e.preventDefault();
      setActiveSlot((prev) => Math.max(0, prev - 1));
    } else if (e.key === 'ArrowRight') {
      e.preventDefault();
      setActiveSlot((prev) => Math.min(2, prev + 1));
    } else if (/^[a-zA-Z]$/.test(e.key)) {
      e.preventDefault();
      const chars = initials.padEnd(3, 'A').split('');
      chars[activeSlot] = e.key.toUpperCase();
      setInitials(chars.join('').slice(0, 3));
      setActiveSlot((prev) => Math.min(2, prev + 1));
    } else if (e.key === 'Backspace') {
      e.preventDefault();
      setActiveSlot((prev) => Math.max(0, prev - 1));
    } else if (e.key === 'Enter') {
      e.preventDefault();
      handleSubmitScore();
    }
  };

  const handleSubmitScore = async () => {
    if (scoreSubmitted || submittingScore) return;
    setSubmittingScore(true);

    const safeInitials = (initials || 'PIL').trim().toUpperCase().slice(0, 3);
    try {
      localStorage.setItem(LOCAL_STORAGE_INITIALS_KEY, safeInitials);
    } catch {
      // Ignore
    }

    const newRecord: ArcadeScoreEntry = {
      game_version: 'v1.0.0',
      seed,
      is_demo: false, // Strict invariant: demo never submits
      duration_seconds: Math.round(gameState.elapsedTime * 10) / 10,
      sectors_cleared: Math.max(0, gameState.sector - 1),
      score: gameState.score,
      initials: safeInitials,
      created_at: new Date().toISOString(),
    };

    // Save locally
    setLeaderboard((prev) => {
      const updated = [...prev, newRecord].sort((a, b) => b.score - a.score).slice(0, 10);
      try {
        localStorage.setItem(LOCAL_STORAGE_SCORES_KEY, JSON.stringify(updated));
      } catch {
        // Ignore
      }
      return updated;
    });

    // Post to backend
    try {
      await fetch('/api/arcade/scores', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(newRecord),
      });
    } catch {
      // Offline fallback
    }

    setScoreSubmitted(true);
    setSubmittingScore(false);
    setShowLeaderboard(true);
  };

  const topHighScore = leaderboard.length > 0 ? leaderboard[0].score : 0;
  const topPilot = leaderboard.length > 0 ? leaderboard[0].initials : 'NONE';

  const sectorName = SECTORS[gameState.sector - 1]?.name || 'Univariate Void';

  return (
    <div
      className="card"
      style={{
        maxWidth: 900,
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
            <span
              className="badge"
              style={{ background: '#1e293b', color: '#38bdf8', border: '1px solid #0284c7' }}
            >
              🏆 Record: {topHighScore} ({topPilot})
            </span>
          </div>
          <p style={{ margin: '0.25rem 0 0', fontSize: '0.85rem', color: '#94a3b8' }}>
            Original space arcade navigation drill. Scores remain strictly separate from practice mastery.
          </p>
        </div>

        <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
          <button
            id="btn-arcade-leaderboard"
            className="btn btn-outline"
            onClick={() => setShowLeaderboard((prev) => !prev)}
            style={{ fontSize: '0.85rem', padding: '0.35rem 0.75rem' }}
            title="Toggle High Scores Leaderboard"
          >
            {showLeaderboard ? '🚀 Back to Flight' : '🏆 Hall of Fame'}
          </button>

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
        Quant Flight: Mode {gameState.mode}, Score {gameState.score}, Sector {gameState.sector}{' '}
        ({sectorName}), Shields {Math.round(gameState.ship.shields)}%, Hull{' '}
        {Math.round(gameState.ship.hull)}%, Bombs {gameState.ship.bombs}.
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

        {/* Modal: Hall of Fame Leaderboard Overlay */}
        {showLeaderboard && (
          <div
            style={{
              position: 'absolute',
              inset: 0,
              background: 'rgba(15, 23, 42, 0.95)',
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              justifyContent: 'center',
              padding: '1.5rem',
              backdropFilter: 'blur(4px)',
              zIndex: 20,
            }}
          >
            <h3
              style={{
                margin: '0 0 1rem',
                fontSize: '1.5rem',
                fontWeight: 700,
                color: '#38bdf8',
                letterSpacing: '0.05em',
              }}
            >
              🏆 TOP PILOT HALL OF FAME
            </h3>
            <div
              style={{
                width: '100%',
                maxWidth: 480,
                background: '#0a0f1d',
                borderRadius: 6,
                border: '1px solid #1e293b',
                padding: '0.75rem 1rem',
                maxHeight: 280,
                overflowY: 'auto',
              }}
            >
              {leaderboard.length === 0 ? (
                <p style={{ textAlign: 'center', color: '#64748b', margin: '1rem 0' }}>
                  No flight records logged yet. Take flight and set the record!
                </p>
              ) : (
                <table style={{ width: '100%', fontSize: '0.85rem', borderCollapse: 'collapse' }}>
                  <thead>
                    <tr style={{ color: '#94a3b8', borderBottom: '1px solid #1e293b' }}>
                      <th style={{ textAlign: 'left', padding: '0.4rem' }}>RANK</th>
                      <th style={{ textAlign: 'left', padding: '0.4rem' }}>PILOT</th>
                      <th style={{ textAlign: 'right', padding: '0.4rem' }}>SCORE</th>
                      <th style={{ textAlign: 'right', padding: '0.4rem' }}>SECTOR</th>
                    </tr>
                  </thead>
                  <tbody>
                    {leaderboard.map((rec, i) => (
                      <tr
                        key={i}
                        style={{
                          borderBottom: '1px solid #1e293b',
                          color: i === 0 ? '#fbbf24' : '#f8fafc',
                          fontFamily: 'ui-monospace, monospace',
                        }}
                      >
                        <td style={{ padding: '0.4rem' }}>#{i + 1}</td>
                        <td style={{ padding: '0.4rem', fontWeight: 700 }}>{rec.initials}</td>
                        <td style={{ textAlign: 'right', padding: '0.4rem' }}>
                          {rec.score.toLocaleString()}
                        </td>
                        <td style={{ textAlign: 'right', padding: '0.4rem', color: '#38bdf8' }}>
                          S{rec.sectors_cleared + 1}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>
            <div style={{ display: 'flex', gap: '0.75rem', marginTop: '1.25rem' }}>
              <button
                id="btn-play-again-leaderboard"
                className="btn btn-primary"
                onClick={handleRestart}
                style={{ fontSize: '0.85rem', padding: '0.4rem 1rem' }}
              >
                🎮 Launch New Flight
              </button>
              <button
                className="btn btn-outline"
                onClick={() => setShowLeaderboard(false)}
                style={{ fontSize: '0.85rem', padding: '0.4rem 1rem' }}
              >
                ✕ Close
              </button>
            </div>
          </div>
        )}

        {/* Modal: Game Over 3-Letter Initials Entry */}
        {gameState.gameOver && gameState.mode === 'manual' && !scoreSubmitted && !showLeaderboard && (
          <div
            id="arcade-initials-modal"
            tabIndex={0}
            autoFocus
            ref={(el) => el?.focus()}
            onKeyDown={handleInitialsKeyDown}
            style={{
              position: 'absolute',
              inset: 0,
              background: 'rgba(15, 23, 42, 0.92)',
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              justifyContent: 'center',
              padding: '1.5rem',
              backdropFilter: 'blur(4px)',
              zIndex: 15,
            }}
          >
            <div
              style={{
                background: '#090d16',
                borderRadius: 8,
                border: '1px solid #ef4444',
                padding: '1.5rem 2rem',
                maxWidth: 440,
                width: '100%',
                textAlign: 'center',
                boxShadow: '0 10px 25px rgba(239, 68, 68, 0.25)',
              }}
            >
              <h3 style={{ margin: '0 0 0.5rem', fontSize: '1.3rem', color: '#ef4444', fontWeight: 700 }}>
                💥 CRITICAL FAILURE — ENTER INITIALS
              </h3>
              <p style={{ margin: '0 0 1rem', fontSize: '0.85rem', color: '#94a3b8' }}>
                Score: <strong style={{ color: '#f8fafc' }}>{gameState.score}</strong> | Sector:{' '}
                <strong style={{ color: '#38bdf8' }}>{gameState.sector}</strong> | Hazards Cleared:{' '}
                <strong style={{ color: '#f59e0b' }}>{gameState.targetsDestroyed}</strong>
              </p>

              {/* 3-Letter Arcade Initials Picker */}
              <div
                style={{
                  display: 'flex',
                  justifyContent: 'center',
                  gap: '0.75rem',
                  margin: '1rem 0 1.25rem',
                }}
              >
                {[0, 1, 2].map((idx) => {
                  const letter = initials[idx] || 'A';
                  const isCurrent = activeSlot === idx;
                  return (
                    <div
                      key={idx}
                      onClick={() => setActiveSlot(idx)}
                      style={{
                        display: 'flex',
                        flexDirection: 'column',
                        alignItems: 'center',
                        cursor: 'pointer',
                      }}
                    >
                      <button
                        type="button"
                        onClick={(e) => {
                          e.stopPropagation();
                          changeInitialLetter(idx, 1);
                        }}
                        style={{
                          background: 'none',
                          border: 'none',
                          color: '#64748b',
                          cursor: 'pointer',
                          padding: '2px',
                          fontSize: '10px',
                        }}
                        aria-label={`Increment letter ${idx + 1}`}
                      >
                        ▲
                      </button>
                      <div
                        data-testid={`initial-slot-${idx}`}
                        style={{
                          width: 48,
                          height: 56,
                          background: isCurrent ? '#1e293b' : '#0f172a',
                          border: `2px solid ${isCurrent ? '#38bdf8' : '#334155'}`,
                          borderRadius: 6,
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          fontSize: '1.8rem',
                          fontWeight: 700,
                          color: isCurrent ? '#38bdf8' : '#f8fafc',
                          fontFamily: 'ui-monospace, monospace',
                        }}
                      >
                        {letter}
                      </div>
                      <button
                        type="button"
                        onClick={(e) => {
                          e.stopPropagation();
                          changeInitialLetter(idx, -1);
                        }}
                        style={{
                          background: 'none',
                          border: 'none',
                          color: '#64748b',
                          cursor: 'pointer',
                          padding: '2px',
                          fontSize: '10px',
                        }}
                        aria-label={`Decrement letter ${idx + 1}`}
                      >
                        ▼
                      </button>
                    </div>
                  );
                })}
              </div>

              <p style={{ margin: '0 0 1.25rem', fontSize: '0.75rem', color: '#64748b' }}>
                Use <kbd className="kbd">Arrows</kbd> to cycle/move, or type letters directly. Press <kbd className="kbd">Enter</kbd> to record.
              </p>

              <div style={{ display: 'flex', gap: '0.75rem', justifyContent: 'center' }}>
                <button
                  id="btn-submit-arcade-score"
                  className="btn btn-primary"
                  onClick={handleSubmitScore}
                  disabled={submittingScore}
                  style={{ fontSize: '0.85rem', padding: '0.45rem 1.2rem' }}
                >
                  {submittingScore ? 'Saving...' : '💾 Submit Score (Enter)'}
                </button>
                <button
                  id="btn-restart-arcade-gameover"
                  className="btn btn-outline"
                  onClick={handleRestart}
                  style={{ fontSize: '0.85rem', padding: '0.45rem 1rem' }}
                >
                  🔄 Retry Flight
                </button>
              </div>
            </div>
          </div>
        )}
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
            <kbd className="kbd">Arrows</kbd> / <kbd className="kbd">WASD</kbd> Steer
          </span>
          <span>
            <kbd className="kbd">Space</kbd> / <kbd className="kbd">F</kbd> Autofire
          </span>
          <span>
            <kbd className="kbd">B</kbd> / <kbd className="kbd">X</kbd> Smart Bomb
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

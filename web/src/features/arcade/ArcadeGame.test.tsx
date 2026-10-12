import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { ArcadeGame } from './ArcadeGame';
import { createInitialState } from './simulation';

describe('ArcadeGame Component', () => {
  it('renders canvas, status badge, and skip button', () => {
    const handleExit = vi.fn();
    render(<ArcadeGame onExit={handleExit} initialMode="attract" />);

    expect(screen.getByRole('heading', { name: /Quant Flight/i })).toBeInTheDocument();
    expect(screen.getByText(/⚡ Attract Mode/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Skip to Practice/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Pause/i })).toBeInTheDocument();
    expect(screen.getByLabelText(/Quant Flight game canvas/i)).toBeInTheDocument();
  });

  it('calls onExit when clicking Skip to Practice button', () => {
    const handleExit = vi.fn();
    render(<ArcadeGame onExit={handleExit} />);

    const skipBtn = screen.getByRole('button', { name: /Skip to Practice/i });
    fireEvent.click(skipBtn);

    expect(handleExit).toHaveBeenCalledTimes(1);
  });

  it('calls onExit when pressing Escape or Enter', () => {
    const handleExit = vi.fn();
    render(<ArcadeGame onExit={handleExit} />);

    fireEvent.keyDown(window, { key: 'Escape' });
    expect(handleExit).toHaveBeenCalledTimes(1);

    fireEvent.keyDown(window, { key: 'Enter' });
    expect(handleExit).toHaveBeenCalledTimes(2);
  });

  it('toggles pause on button click or P key', () => {
    const handleExit = vi.fn();
    render(<ArcadeGame onExit={handleExit} />);

    const pauseBtn = screen.getByRole('button', { name: /Pause/i });
    fireEvent.click(pauseBtn);

    expect(screen.getByRole('button', { name: /Resume/i })).toBeInTheDocument();

    // Press P to resume
    fireEvent.keyDown(window, { key: 'p' });
    expect(screen.getByRole('button', { name: /Pause/i })).toBeInTheDocument();
  });

  it('toggles reduced motion option', () => {
    render(<ArcadeGame onExit={vi.fn()} />);

    const toggle = screen.getByLabelText(/Reduced Motion/i) as HTMLInputElement;
    expect(toggle).toBeInTheDocument();
    fireEvent.click(toggle);
    expect(toggle.checked).toBe(true);
  });

  it('opens and closes Hall of Fame leaderboard overlay', () => {
    render(<ArcadeGame onExit={vi.fn()} />);

    const hallOfFameBtn = screen.getByRole('button', { name: /Hall of Fame/i });
    expect(hallOfFameBtn).toBeInTheDocument();

    fireEvent.click(hallOfFameBtn);
    expect(screen.getByText(/TOP PILOT HALL OF FAME/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Launch New Flight/i })).toBeInTheDocument();

    // Close overlay
    const closeBtn = screen.getByRole('button', { name: /Close/i });
    fireEvent.click(closeBtn);
    expect(screen.queryByText(/TOP PILOT HALL OF FAME/i)).not.toBeInTheDocument();
  });

  it('resets and launches new flight from leaderboard modal', () => {
    render(<ArcadeGame onExit={vi.fn()} initialMode="attract" />);

    // Open Hall of Fame
    fireEvent.click(screen.getByRole('button', { name: /Hall of Fame/i }));
    expect(screen.getByText(/TOP PILOT HALL OF FAME/i)).toBeInTheDocument();

    // Click Launch New Flight
    fireEvent.click(screen.getByRole('button', { name: /Launch New Flight/i }));
    expect(screen.queryByText(/TOP PILOT HALL OF FAME/i)).not.toBeInTheDocument();
    expect(screen.getByText(/🎮 Manual Flight/i)).toBeInTheDocument();
  });

  it('allows entering initials containing W, A, S, and D without moving slots or cycling letters', () => {
    const gameOverState = {
      ...createInitialState(42, 'manual'),
      gameOver: true,
      mode: 'manual' as const,
      score: 5000,
    };

    render(<ArcadeGame onExit={vi.fn()} initialState={gameOverState} />);

    expect(screen.getByText(/CRITICAL FAILURE — ENTER INITIALS/i)).toBeInTheDocument();

    const modal = screen.getByText(/CRITICAL FAILURE — ENTER INITIALS/i).closest('#arcade-initials-modal')!;

    // Type 'D' into slot 0 -> should set 'D' and advance to slot 1
    fireEvent.keyDown(modal, { key: 'd' });
    expect(screen.getByTestId('initial-slot-0')).toHaveTextContent('D');

    // Type 'A' into slot 1 -> should set 'A' and advance to slot 2
    fireEvent.keyDown(modal, { key: 'a' });
    expect(screen.getByTestId('initial-slot-1')).toHaveTextContent('A');

    // Type 'N' into slot 2 -> should set 'N'
    fireEvent.keyDown(modal, { key: 'n' });
    expect(screen.getByTestId('initial-slot-2')).toHaveTextContent('N');
  });

  it('allows entering W and S as initials and using arrow keys to cycle', () => {
    const gameOverState = {
      ...createInitialState(42, 'manual'),
      gameOver: true,
      mode: 'manual' as const,
      score: 5000,
    };

    render(<ArcadeGame onExit={vi.fn()} initialState={gameOverState} />);
    const modal = screen.getByText(/CRITICAL FAILURE — ENTER INITIALS/i).closest('#arcade-initials-modal')!;

    // Type 'W' into slot 0
    fireEvent.keyDown(modal, { key: 'w' });
    expect(screen.getByTestId('initial-slot-0')).toHaveTextContent('W');

    // Type 'S' into slot 1
    fireEvent.keyDown(modal, { key: 's' });
    expect(screen.getByTestId('initial-slot-1')).toHaveTextContent('S');

    // Arrow keys cycle slot 2
    fireEvent.keyDown(modal, { key: 'ArrowUp' });
    // Slot 2 should have cycled from its default letter
    expect(screen.getByTestId('initial-slot-2')).not.toHaveTextContent('');

    // Arrow Left moves active slot back to slot 1
    fireEvent.keyDown(modal, { key: 'ArrowLeft' });
    // Overwrite slot 1 with 'X'
    fireEvent.keyDown(modal, { key: 'x' });
    expect(screen.getByTestId('initial-slot-1')).toHaveTextContent('X');
  });
});


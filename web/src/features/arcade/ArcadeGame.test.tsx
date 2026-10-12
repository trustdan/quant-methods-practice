import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { ArcadeGame } from './ArcadeGame';

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
});


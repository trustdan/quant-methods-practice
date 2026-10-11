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
});

import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ProductTourModal, STORAGE_KEY_TOUR_DISMISSED, isTourDismissed, setTourDismissed } from './ProductTourModal';

describe('ProductTourModal Component', () => {
  beforeEach(() => {
    localStorage.clear();
    vi.clearAllMocks();
  });

  it('does not render when isOpen is false', () => {
    render(<ProductTourModal isOpen={false} onClose={vi.fn()} />);
    expect(screen.queryByRole('dialog', { name: /Quant Methods Product Tour/i })).toBeNull();
  });

  it('renders interactive first step when isOpen is true', () => {
    render(<ProductTourModal isOpen={true} onClose={vi.fn()} />);
    
    expect(screen.getByRole('dialog', { name: /Quant Methods Product Tour/i })).toBeInTheDocument();
    expect(screen.getByText('Multi-Stage Practice Drills')).toBeInTheDocument();
    expect(screen.getByText(/Systematically decompose complex quantitative problems/i)).toBeInTheDocument();
    expect(screen.getByText(/Step 1 of 8/i)).toBeInTheDocument();
    expect(screen.getByText(/7-Stage Scaffolding/i)).toBeInTheDocument();
    expect(screen.getByText(/Cognitive Gate Enforcement/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Next →/i })).toBeInTheDocument();
  });

  it('advances through steps when clicking Next and goes back with Back button', () => {
    render(<ProductTourModal isOpen={true} onClose={vi.fn()} />);

    // Step 1 -> Next
    fireEvent.click(screen.getByRole('button', { name: /Next →/i }));
    expect(screen.getByText('Calculations & Non-Punitive Hints')).toBeInTheDocument();
    expect(screen.getByText(/Step 2 of 8/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /← Back/i })).toBeInTheDocument();

    // Step 2 -> Back
    fireEvent.click(screen.getByRole('button', { name: /← Back/i }));
    expect(screen.getByText('Multi-Stage Practice Drills')).toBeInTheDocument();
    expect(screen.getByText(/Step 1 of 8/i)).toBeInTheDocument();
  });

  it('allows clicking step pip dots to jump directly to any step', () => {
    render(<ProductTourModal isOpen={true} onClose={vi.fn()} />);

    // Jump to step 4 (Mastery)
    const pip4 = screen.getByLabelText(/Jump to tour step 4/i);
    fireEvent.click(pip4);

    expect(screen.getByText('Evidence-Based Concept Mastery')).toBeInTheDocument();
    expect(screen.getByText(/Step 4 of 8/i)).toBeInTheDocument();
    expect(screen.getByText(/Independent vs\. Assisted Evidence/i)).toBeInTheDocument();

    // Jump to step 5 (Reference Library)
    const pip5 = screen.getByLabelText(/Jump to tour step 5/i);
    fireEvent.click(pip5);

    expect(screen.getByText('Reference Library & Excel / R Cheatsheets')).toBeInTheDocument();
    expect(screen.getByText(/Step 5 of 8/i)).toBeInTheDocument();
    expect(screen.getByText(/Excel & R Cheatsheet Catalog/i)).toBeInTheDocument();
  });

  it('persists dismissed status in localStorage when clicking "Don\'t show again" button', () => {
    const handleClose = vi.fn();
    render(<ProductTourModal isOpen={true} onClose={handleClose} />);

    expect(isTourDismissed()).toBe(false);

    const dontShowBtn = screen.getByRole('button', { name: /Don't show again/i });
    fireEvent.click(dontShowBtn);

    expect(isTourDismissed()).toBe(true);
    expect(localStorage.getItem(STORAGE_KEY_TOUR_DISMISSED)).toBe('true');
    expect(handleClose).toHaveBeenCalledTimes(1);
  });

  it('persists dismissed status when finishing tour with checkbox checked', () => {
    const handleClose = vi.fn();
    render(<ProductTourModal isOpen={true} onClose={handleClose} />);

    // Jump to final step (Step 8: Exams)
    const pip8 = screen.getByLabelText(/Jump to tour step 8/i);
    fireEvent.click(pip8);

    expect(screen.getByText('Timed Exams & Assessment Replays')).toBeInTheDocument();
    const finishBtn = screen.getByRole('button', { name: /Get Started 🚀/i });
    fireEvent.click(finishBtn);

    expect(isTourDismissed()).toBe(true);
    expect(handleClose).toHaveBeenCalledTimes(1);
  });

  it('does not persist dismissed status if checkbox is unchecked when skipping tour', () => {
    const handleClose = vi.fn();
    render(<ProductTourModal isOpen={true} onClose={handleClose} />);

    const checkbox = screen.getByLabelText(/Don't show again on startup/i) as HTMLInputElement;
    expect(checkbox.checked).toBe(true);
    fireEvent.click(checkbox);
    expect(checkbox.checked).toBe(false);

    const skipBtn = screen.getByRole('button', { name: /Skip Tour/i });
    fireEvent.click(skipBtn);

    expect(isTourDismissed()).toBe(false);
    expect(handleClose).toHaveBeenCalledTimes(1);
  });

  it('navigates to specific tab when clicking feature jump button', () => {
    const handleNavigateTab = vi.fn();
    const handleClose = vi.fn();
    render(
      <ProductTourModal
        isOpen={true}
        onClose={handleClose}
        onNavigateTab={handleNavigateTab}
      />
    );

    // Jump to step 4 (Mastery)
    const pip4 = screen.getByLabelText(/Jump to tour step 4/i);
    fireEvent.click(pip4);

    const jumpBtn = screen.getByRole('button', { name: /View Mastery Dashboard ↗/i });
    fireEvent.click(jumpBtn);

    expect(handleNavigateTab).toHaveBeenCalledWith('mastery');
    expect(handleClose).toHaveBeenCalledTimes(1);
    expect(isTourDismissed()).toBe(true);
  });

  it('supports keyboard navigation with ArrowRight, ArrowLeft, and Escape', () => {
    const handleClose = vi.fn();
    render(<ProductTourModal isOpen={true} onClose={handleClose} />);

    // ArrowRight advances to step 2
    fireEvent.keyDown(window, { key: 'ArrowRight' });
    expect(screen.getByText('Calculations & Non-Punitive Hints')).toBeInTheDocument();

    // ArrowLeft goes back to step 1
    fireEvent.keyDown(window, { key: 'ArrowLeft' });
    expect(screen.getByText('Multi-Stage Practice Drills')).toBeInTheDocument();

    // Escape closes modal
    fireEvent.keyDown(window, { key: 'Escape' });
    expect(handleClose).toHaveBeenCalledTimes(1);
  });

  it('setTourDismissed utility cleanly sets and clears localStorage', () => {
    expect(isTourDismissed()).toBe(false);
    setTourDismissed(true);
    expect(isTourDismissed()).toBe(true);
    setTourDismissed(false);
    expect(isTourDismissed()).toBe(false);
  });
});

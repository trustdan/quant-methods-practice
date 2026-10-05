import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { App } from './App';

describe('App Shell', () => {
  beforeEach(() => {
    // Mock fetch for health and session endpoints
    vi.stubGlobal('fetch', vi.fn().mockImplementation((url: string) => {
      if (url === '/api/health') {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ status: 'ok', version: '0.1.0-dev' }),
        });
      }
      if (url === '/api/local-session') {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ status: 'ok', session_token: 'test-session' }),
        });
      }
      if (url === '/api/practice/sessions') {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve(null), // triggers default fallback session
        });
      }
      return Promise.reject(new Error('Unknown url'));
    }));
  });

  it('renders application header, title, and approved drill scenario', async () => {
    render(<App />);

    expect(screen.getByText('Quant Methods Practice')).toBeInTheDocument();
    expect(screen.getByText(/Stage 04 Drill/)).toBeInTheDocument();
    expect(screen.getByText(/Exactly two heads in four tosses/)).toBeInTheDocument();
    expect(screen.getByText(/What should \$X\$ represent\?/)).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByText(/Loopback Server Active/)).toBeInTheDocument();
    });
  });

  it('allows selecting choice, submitting answer, and navigating visited stages', async () => {
    render(<App />);

    // Initially at Stage 1/7 (Target)
    expect(screen.getByText(/Stage 1\/7: Target/)).toBeInTheDocument();

    // Select choice 1 ("The number of heads in four tosses")
    const option1 = screen.getByText(/The number of heads in four tosses/);
    fireEvent.click(option1);

    // Submit answer
    const submitBtn = screen.getByRole('button', { name: /Submit Answer/ });
    fireEvent.click(submitBtn);

    // After correct submission, advances to Stage 2: Model
    await waitFor(() => {
      expect(screen.getByText(/Stage 2\/7: Model/)).toBeInTheDocument();
    });

    // Navigate back to visited Stage 1 with 'h' key
    fireEvent.keyDown(window, { key: 'h' });
    await waitFor(() => {
      expect(screen.getByText(/Stage 1\/7: Target/)).toBeInTheDocument();
    });

    // Navigate forward to Stage 2 with 'l' key
    fireEvent.keyDown(window, { key: 'l' });
    await waitFor(() => {
      expect(screen.getByText(/Stage 2\/7: Model/)).toBeInTheDocument();
    });
  });

  it('toggles keyboard shortcut help modal with F1 and Esc', async () => {
    render(<App />);

    // Press F1
    fireEvent.keyDown(window, { key: 'F1' });
    expect(screen.getByRole('dialog', { name: /Keyboard Shortcuts/i })).toBeInTheDocument();

    // Press Esc to dismiss
    fireEvent.keyDown(window, { key: 'Escape' });
    expect(screen.queryByRole('dialog', { name: /Keyboard Shortcuts/i })).toBeNull();

    await waitFor(() => {
      expect(screen.getByText(/Loopback Server Active/)).toBeInTheDocument();
    });
  });

  it('allows toggling between Practice Drill and Formula Gallery', async () => {
    render(<App />);

    const galleryBtn = screen.getByRole('button', { name: /Formula Gallery/ });
    fireEvent.click(galleryBtn);

    expect(screen.getByText(/Mathematical Formula Gallery/)).toBeInTheDocument();
    expect(screen.getByText(/1\. Set Operations & Probability/)).toBeInTheDocument();

    const drillBtn = screen.getByRole('button', { name: /Practice Drill/ });
    fireEvent.click(drillBtn);

    expect(screen.getByText(/Stage 1\/7: Target/)).toBeInTheDocument();
  });

  it('opens leave-intent modal with q and allows dismissing with Esc or Cancel', async () => {
    render(<App />);

    // Press 'q' outside input
    fireEvent.keyDown(window, { key: 'q' });
    expect(screen.getByRole('dialog', { name: /Confirm Leave Practice/i })).toBeInTheDocument();
    expect(screen.getByText(/Leave Practice Session\?/)).toBeInTheDocument();

    // Press Escape to dismiss
    fireEvent.keyDown(window, { key: 'Escape' });
    expect(screen.queryByRole('dialog', { name: /Confirm Leave Practice/i })).toBeNull();

    // Press 'q' again
    fireEvent.keyDown(window, { key: 'q' });
    expect(screen.getByRole('dialog', { name: /Confirm Leave Practice/i })).toBeInTheDocument();

    // Click 'Cancel' button
    const cancelBtn = screen.getByRole('button', { name: /Cancel/ });
    fireEvent.click(cancelBtn);
    expect(screen.queryByRole('dialog', { name: /Confirm Leave Practice/i })).toBeNull();
  });
});

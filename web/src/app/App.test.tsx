import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { App } from './App';
import { STORAGE_KEY_TOUR_DISMISSED } from '../features/tour/ProductTourModal';

describe('App Shell', () => {
  let mockServerSettings: {
    question_count: number;
    module_ids: string[];
    intensity: string;
    tour_dismissed?: boolean;
  };

  beforeEach(() => {
    localStorage.clear();
    mockServerSettings = {
      question_count: 10,
      module_ids: [],
      intensity: 'standard',
      tour_dismissed: false,
    };

    // Mock fetch for health, session, and settings endpoints
    vi.stubGlobal('fetch', vi.fn().mockImplementation((url: string, init?: RequestInit) => {
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
      if (url === '/api/settings') {
        if (init?.method === 'POST' && init.body) {
          const body = JSON.parse(init.body as string);
          mockServerSettings = { ...mockServerSettings, ...body };
          return Promise.resolve({
            ok: true,
            json: () => Promise.resolve(mockServerSettings),
          });
        }
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve(mockServerSettings),
        });
      }
      return Promise.reject(new Error('Unknown url'));
    }));
  });

  it('renders application header, title, and approved drill scenario', async () => {
    localStorage.setItem(STORAGE_KEY_TOUR_DISMISSED, 'true');
    mockServerSettings.tour_dismissed = true;
    render(<App />);

    expect(screen.getByText('Quant Methods Practice')).toBeInTheDocument();
    expect(screen.getByText(/Exactly two heads in four tosses/)).toBeInTheDocument();
    expect(screen.getByText(/What should \$X\$ represent\?/)).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByText(/Loopback Server Active/)).toBeInTheDocument();
    });
  });

  it('shows website-like product tour on startup when not dismissed, and dismisses with "Don\'t show again"', async () => {
    render(<App />);

    // Tour dialog should be presented when settings check completes
    await waitFor(() => {
      expect(screen.getByRole('dialog', { name: /Quant Methods Product Tour/i })).toBeInTheDocument();
    });
    expect(screen.getByText('Multi-Stage Practice Drills')).toBeInTheDocument();
    expect(screen.getByText(/Step 1 of 8/i)).toBeInTheDocument();

    // Click "Don't show again" button
    const dontShowBtn = screen.getByRole('button', { name: /Don't show again/i });
    fireEvent.click(dontShowBtn);

    // Dialog closes
    await waitFor(() => {
      expect(screen.queryByRole('dialog', { name: /Quant Methods Product Tour/i })).toBeNull();
    });
    // Preference is saved both to localStorage and to server settings
    expect(localStorage.getItem(STORAGE_KEY_TOUR_DISMISSED)).toBe('true');
    expect(mockServerSettings.tour_dismissed).toBe(true);
  });

  it('does not show product tour if server SQLite settings indicate tour_dismissed, even with empty localStorage', async () => {
    mockServerSettings.tour_dismissed = true;
    render(<App />);

    await waitFor(() => {
      expect(localStorage.getItem(STORAGE_KEY_TOUR_DISMISSED)).toBe('true');
    });

    // Tour should NOT appear because server settings has tour_dismissed: true
    expect(screen.queryByRole('dialog', { name: /Quant Methods Product Tour/i })).toBeNull();
  });

  it('allows reopening product tour from header button and from Help modal', async () => {
    localStorage.setItem(STORAGE_KEY_TOUR_DISMISSED, 'true');
    mockServerSettings.tour_dismissed = true;
    render(<App />);

    // Initially tour is not shown because it was dismissed
    expect(screen.queryByRole('dialog', { name: /Quant Methods Product Tour/i })).toBeNull();

    // Click header "🚀 Tour" button
    const headerTourBtn = screen.getByRole('button', { name: /🚀 Tour/i });
    fireEvent.click(headerTourBtn);

    expect(screen.getByRole('dialog', { name: /Quant Methods Product Tour/i })).toBeInTheDocument();

    // Close tour via Skip Tour
    fireEvent.click(screen.getByRole('button', { name: /Skip Tour/i }));
    expect(screen.queryByRole('dialog', { name: /Quant Methods Product Tour/i })).toBeNull();

    // Open Help modal via F1
    fireEvent.keyDown(window, { key: 'F1' });
    expect(screen.getByRole('dialog', { name: /Keyboard Shortcuts/i })).toBeInTheDocument();

    // Click "Launch Product Tour" inside Help modal
    const helpTourBtn = screen.getByRole('button', { name: /🚀 Launch Product Tour/i });
    fireEvent.click(helpTourBtn);

    // Help modal closes and tour opens
    expect(screen.queryByRole('dialog', { name: /Keyboard Shortcuts/i })).toBeNull();
    expect(screen.getByRole('dialog', { name: /Quant Methods Product Tour/i })).toBeInTheDocument();
  });

  it('allows selecting choice, submitting answer, and navigating visited stages', async () => {
    localStorage.setItem(STORAGE_KEY_TOUR_DISMISSED, 'true');
    mockServerSettings.tour_dismissed = true;
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
    localStorage.setItem(STORAGE_KEY_TOUR_DISMISSED, 'true');
    mockServerSettings.tour_dismissed = true;
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

  it('allows toggling between Practice Drill and Reference Library', async () => {
    localStorage.setItem(STORAGE_KEY_TOUR_DISMISSED, 'true');
    mockServerSettings.tour_dismissed = true;
    render(<App />);

    const refBtn = screen.getByRole('button', { name: /Reference Library/ });
    fireEvent.click(refBtn);

    expect(screen.getByRole('heading', { name: /Reference Library/ })).toBeInTheDocument();
    expect(screen.getByText(/Authoritative course definitions/)).toBeInTheDocument();

    const drillBtn = screen.getByRole('button', { name: /Practice Drill/ });
    fireEvent.click(drillBtn);

    expect(screen.getByText(/Stage 1\/7: Target/)).toBeInTheDocument();
  });

  it('opens leave-intent modal with q and allows dismissing with Esc or Cancel', async () => {
    localStorage.setItem(STORAGE_KEY_TOUR_DISMISSED, 'true');
    mockServerSettings.tour_dismissed = true;
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

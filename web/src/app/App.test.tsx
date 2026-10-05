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
      return Promise.reject(new Error('Unknown url'));
    }));
  });

  it('renders application header, title, and demonstration notice', async () => {
    render(<App />);

    expect(screen.getByText('Quant Methods Practice')).toBeInTheDocument();
    expect(screen.getByText(/Stage 01 Foundation: Mathematical Rendering/)).toBeInTheDocument();
    expect(screen.getByText(/Static demonstration only/)).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByText(/Loopback Server Active/)).toBeInTheDocument();
    });
  });

  it('allows navigating stages using pointer buttons and keyboard', async () => {
    render(<App />);

    // Initially at Stage 1/7 (Target)
    expect(screen.getByText(/Stage 1\/7: Target/)).toBeInTheDocument();

    // Click Next button
    const nextBtn = screen.getByRole('button', { name: /Next/ });
    fireEvent.click(nextBtn);

    // Now at Stage 2/7 (Model)
    expect(screen.getByText(/Stage 2\/7: Model/)).toBeInTheDocument();

    // Navigate back with 'h' key
    fireEvent.keyDown(window, { key: 'h' });
    expect(screen.getByText(/Stage 1\/7: Target/)).toBeInTheDocument();

    // Navigate forward with 'l' key
    fireEvent.keyDown(window, { key: 'l' });
    expect(screen.getByText(/Stage 2\/7: Model/)).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByText(/Loopback Server Active/)).toBeInTheDocument();
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
});

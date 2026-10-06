import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { AITutorPanel } from './AITutorPanel';

describe('AITutorPanel Component', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('renders nothing when isOpen is false', () => {
    const { container } = render(
      <AITutorPanel isOpen={false} onClose={vi.fn()} />
    );
    expect(container.firstChild).toBeNull();
  });

  it('renders tutor dialog with advisory banner and action buttons when open', () => {
    render(
      <AITutorPanel
        isOpen={true}
        onClose={vi.fn()}
        sessionId="sess_test"
        instanceId="inst_test"
        stageId="stage_test"
        instanceTitle="Binomial Coin Toss"
      />
    );

    expect(screen.getByText(/AI Tutor/i)).toBeInTheDocument();
    expect(screen.getByText(/Advisory AI explanation:/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Causal Hint/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Full Step-by-Step Solution/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Explain Differently/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Why Conditions Matter/i })).toBeInTheDocument();
  });

  it('initiates streaming request and renders completed explanation', async () => {
    // Mock /api/tutor/requests and /api/tutor/requests/:id/events
    const mockEvents = [
      'data: {"type":"started","request_id":"req_1"}\n\n',
      'data: {"type":"text_delta","request_id":"req_1","delta":"Binomial probability derivation: "}\n\n',
      'data: {"type":"text_delta","request_id":"req_1","delta":"$P(X=2) = 0.375$."}\n\n',
      'data: {"type":"complete","request_id":"req_1","text":"Binomial probability derivation: $P(X=2) = 0.375$."}\n\n',
    ];

    const stream = new ReadableStream({
      start(controller) {
        mockEvents.forEach((chunk) => {
          controller.enqueue(new TextEncoder().encode(chunk));
        });
        controller.close();
      },
    });

    global.fetch = vi.fn().mockImplementation((url: string, _opts?: any) => {
      if (url === '/api/tutor/requests') {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ request_id: 'req_1', status: 'started' }),
        });
      }
      if (url === '/api/tutor/requests/req_1/events') {
        return Promise.resolve({
          ok: true,
          body: stream,
        });
      }
      if (url.startsWith('/api/tutor/drafts/')) {
        return Promise.resolve({ ok: true, json: () => Promise.resolve(null) });
      }
      return Promise.resolve({ ok: true, json: () => Promise.resolve({}) });
    });

    render(
      <AITutorPanel
        isOpen={true}
        onClose={vi.fn()}
        sessionId="sess_test"
        instanceId="inst_test"
        stageId="stage_test"
      />
    );

    const explainBtn = screen.getByRole('button', { name: /Full Step-by-Step Solution/i });
    fireEvent.click(explainBtn);

    await waitFor(() => {
      expect(screen.getByText(/Binomial probability derivation:/i)).toBeInTheDocument();
    });

    // Save note button should now be visible
    const saveBtn = screen.getByRole('button', { name: /Save to Notes/i });
    expect(saveBtn).toBeInTheDocument();
  });

  it('triggers save note and clears draft', async () => {
    global.fetch = vi.fn().mockImplementation((url: string, opts?: any) => {
      if (url.startsWith('/api/tutor/drafts/')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ recovery_text: 'Recovered draft text' }),
        });
      }
      if (url === '/api/notes' && opts?.method === 'POST') {
        return Promise.resolve({
          ok: true,
          json: () =>
            Promise.resolve({
              id: 'note_123',
              raw_markdown: 'Recovered draft text',
              topic: 'Probability',
              provider_info: { title: 'Saved Note' },
            }),
        });
      }
      return Promise.resolve({ ok: true, json: () => Promise.resolve({}) });
    });

    const onNoteSaved = vi.fn();
    render(
      <AITutorPanel
        isOpen={true}
        onClose={vi.fn()}
        sessionId="sess_1"
        stageId="stage_1"
        onNoteSaved={onNoteSaved}
      />
    );

    await waitFor(() => {
      expect(screen.getByText(/Recovered draft text/i)).toBeInTheDocument();
    });

    const saveBtn = screen.getByRole('button', { name: /Save to Notes/i });
    fireEvent.click(saveBtn);

    await waitFor(() => {
      expect(screen.getByText(/Saved to Personal Notes Library/i)).toBeInTheDocument();
      expect(onNoteSaved).toHaveBeenCalled();
    });
  });

  it('protects unsaved explanation on close with leave-intent dialog', async () => {
    global.fetch = vi.fn().mockImplementation((url: string) => {
      if (url.startsWith('/api/tutor/drafts/')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ recovery_text: 'Unsaved note content' }),
        });
      }
      return Promise.resolve({ ok: true, json: () => Promise.resolve({}) });
    });

    const onClose = vi.fn();
    render(
      <AITutorPanel
        isOpen={true}
        onClose={onClose}
        sessionId="sess_1"
        stageId="stage_1"
      />
    );

    await waitFor(() => {
      expect(screen.getByText(/Unsaved note content/i)).toBeInTheDocument();
    });

    // Clicking close or pressing Escape should prompt Save explanation?
    const closeBtn = screen.getByRole('button', { name: /Close \(Esc\)/i });
    fireEvent.click(closeBtn);

    expect(screen.getByText(/Save Explanation Before Leaving\?/i)).toBeInTheDocument();
    expect(onClose).not.toHaveBeenCalled();

    // Clicking (Esc) Stay retains content and dismisses leave dialog
    const stayBtn = screen.getByRole('button', { name: /\(Esc\) Stay/i });
    fireEvent.click(stayBtn);

    expect(screen.queryByText(/Save Explanation Before Leaving\?/i)).toBeNull();
    expect(onClose).not.toHaveBeenCalled();
    expect(screen.getByText(/Unsaved note content/i)).toBeInTheDocument();
  });
});

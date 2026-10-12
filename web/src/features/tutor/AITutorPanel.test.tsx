import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { AITutorPanel } from './AITutorPanel';

describe('AITutorPanel Component', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    global.fetch = vi.fn().mockImplementation(() =>
      Promise.resolve({
        ok: true,
        json: () => Promise.resolve({}),
      })
    );
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

    global.fetch = vi.fn().mockImplementation((url: string, _opts?: RequestInit) => {
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
    global.fetch = vi.fn().mockImplementation((url: string, opts?: RequestInit) => {
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

  it('displays Stage Completed banner when stageResolved is true', () => {
    render(
      <AITutorPanel
        isOpen={true}
        onClose={vi.fn()}
        sessionId="sess_test"
        instanceId="inst_test"
        stageId="stage_test"
        stageResolved={true}
      />
    );

    expect(screen.getByText('Advisory AI explanation:')).toBeInTheDocument();
    expect(screen.getByText(/Stage Completed/i)).toBeInTheDocument();
    expect(screen.getByText(/Full derivations and Excel\/R formula breakdowns are available/i)).toBeInTheDocument();
  });

  it('switches to Interactive Chat tab and displays conversation threads and messages', async () => {
    const mockThreads = [
      {
        id: 'th_123',
        session_id: 'sess_test',
        instance_id: 'inst_test',
        stage_id: 'stage_test',
        title: 'Hypergeometric vs Binomial Discussion',
        messages_count: 2,
        messages: [
          {
            id: 'msg_1',
            thread_id: 'th_123',
            role: 'user',
            content: 'Is this sampling with or without replacement?',
            created_at: new Date().toISOString(),
          },
          {
            id: 'msg_2',
            thread_id: 'th_123',
            role: 'assistant',
            content: 'Consider whether each draw changes the remaining pool size.',
            created_at: new Date().toISOString(),
          },
        ],
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      },
    ];

    global.fetch = vi.fn().mockImplementation((url: string) => {
      if (url.startsWith('/api/tutor/threads/th_123')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve(mockThreads[0]),
        });
      }
      if (url.startsWith('/api/tutor/threads')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve(mockThreads),
        });
      }
      if (url.startsWith('/api/tutor/drafts/')) {
        return Promise.resolve({ ok: true, json: () => Promise.resolve(null) });
      }
      if (url === '/api/providers') {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ active_route: 'local', active_model: 'llama-3.2-3b' }),
        });
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

    // Switch to Interactive Chat tab
    const chatTab = screen.getByRole('button', { name: /Interactive Chat/i });
    fireEvent.click(chatTab);

    await waitFor(() => {
      expect(screen.getByText('Hypergeometric vs Binomial Discussion')).toBeInTheDocument();
      expect(screen.getByText(/Is this sampling with or without replacement\?/i)).toBeInTheDocument();
      expect(screen.getByText(/Consider whether each draw changes the remaining pool size\./i)).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /Save Full Chat/i })).toBeInTheDocument();
    });
  });

  it('allows saving entire chat conversation transcript to notes', async () => {
    const mockThreads = [
      {
        id: 'th_456',
        session_id: 'sess_test',
        instance_id: 'inst_test',
        stage_id: 'stage_test',
        title: 'Poisson Distribution Limit',
        messages_count: 2,
        messages: [
          {
            id: 'm1',
            thread_id: 'th_456',
            role: 'user',
            content: 'When does binomial approximate Poisson?',
            created_at: new Date().toISOString(),
          },
          {
            id: 'm2',
            thread_id: 'th_456',
            role: 'assistant',
            content: 'When n is large and p is small, with lambda = n*p constant.',
            created_at: new Date().toISOString(),
          },
        ],
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      },
    ];

    const onNoteSaved = vi.fn();

    global.fetch = vi.fn().mockImplementation((url: string, opts?: RequestInit) => {
      if (url.startsWith('/api/tutor/threads/th_456')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve(mockThreads[0]),
        });
      }
      if (url.startsWith('/api/tutor/threads')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve(mockThreads),
        });
      }
      if (url === '/api/notes' && opts?.method === 'POST') {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ id: 'note_conversation_1' }),
        });
      }
      if (url.startsWith('/api/tutor/drafts/')) {
        return Promise.resolve({ ok: true, json: () => Promise.resolve(null) });
      }
      if (url === '/api/providers') {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ active_route: 'local', active_model: 'llama-3.2-3b' }),
        });
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
        onNoteSaved={onNoteSaved}
      />
    );

    // Switch to Chat tab
    fireEvent.click(screen.getByRole('button', { name: /Interactive Chat/i }));

    await waitFor(() => {
      expect(screen.getByText(/When does binomial approximate Poisson\?/i)).toBeInTheDocument();
    });

    // Click Save Full Chat
    const saveConvoBtn = screen.getByRole('button', { name: /Save Full Chat/i });
    fireEvent.click(saveConvoBtn);

    await waitFor(() => {
      expect(screen.getByText(/✓ Full conversation transcript saved to Note Library/i)).toBeInTheDocument();
      expect(onNoteSaved).toHaveBeenCalled();
    });
  });
});

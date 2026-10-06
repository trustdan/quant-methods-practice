import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { ChatGPTPlanPanel } from './ChatGPTPlanPanel';
import { ProviderSettings } from './ProviderSettings';

const account = {
  key: 'abc123',
  label: 'learner@example.com',
  email: 'learner@example.com',
  plan_granted: true,
  needs_reauth: false,
  selected: true,
  expires_at: '2026-10-05T12:00:00Z',
  updated_at: '2026-10-05T11:00:00Z',
};

const ok = (body: unknown) => Promise.resolve({ ok: true, json: () => Promise.resolve(body), text: () => Promise.resolve('') });

function mockBackend(finalStatus: Record<string, unknown>) {
  let accounts: unknown[] = [];
  const calls: string[] = [];
  global.fetch = vi.fn().mockImplementation((url: string, opts?: RequestInit) => {
    const method = opts?.method ?? 'GET';
    calls.push(`${method} ${url}`);
    if (url === '/api/providers/chatgpt/accounts') return ok(accounts);
    if (url === '/api/providers/chatgpt/signin' && method === 'POST') {
      return ok({
        authorize_url: 'https://auth.openai.com/api/accounts/authorize?client_id=dynamic_agent_client',
        status: { attempt_id: 'a1', state: 'pending' },
      });
    }
    if (url === '/api/providers/chatgpt/signin') {
      if (finalStatus.state === 'complete') accounts = [account];
      return ok(finalStatus);
    }
    if (url === `/api/providers/chatgpt/accounts/${account.key}/signout`) {
      accounts = [];
      return ok({ revoked: true, accounts: [] });
    }
    return ok({});
  }) as unknown as typeof fetch;
  return calls;
}

describe('ChatGPTPlanPanel', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('opens the OpenAI sign-in page and reports completion from backend polling', async () => {
    mockBackend({ attempt_id: 'a1', state: 'complete', account });
    const openSpy = vi.spyOn(window, 'open').mockImplementation(() => null);
    const onStatus = vi.fn();
    const onAccountsChanged = vi.fn();
    render(<ChatGPTPlanPanel onStatus={onStatus} onError={vi.fn()} onAccountsChanged={onAccountsChanged} />);

    fireEvent.click(await screen.findByText('Sign in with ChatGPT'));

    await waitFor(() => expect(openSpy).toHaveBeenCalled());
    const [url, target, features] = openSpy.mock.calls[0];
    expect(String(url)).toMatch(/^https:\/\/auth\.openai\.com\//);
    expect(target).toBe('_blank');
    expect(features).toContain('noopener');
    expect(screen.getByText(/Waiting for you to finish signing in/)).toBeInTheDocument();

    await waitFor(() => expect(onStatus).toHaveBeenCalledWith(expect.stringContaining('learner@example.com')), {
      timeout: 4000,
    });
    expect(onAccountsChanged).toHaveBeenCalled();
    expect(await screen.findByText('Plan usage granted')).toBeInTheDocument();
  });

  it('surfaces a failed sign-in such as a missing plan grant', async () => {
    mockBackend({ attempt_id: 'a1', state: 'failed', error: 'this ChatGPT account did not grant plan usage' });
    vi.spyOn(window, 'open').mockImplementation(() => null);
    const onError = vi.fn();
    render(<ChatGPTPlanPanel onStatus={vi.fn()} onError={onError} onAccountsChanged={vi.fn()} />);

    fireEvent.click(await screen.findByText('Sign in with ChatGPT'));
    await waitFor(() => expect(onError).toHaveBeenCalledWith(expect.stringContaining('did not grant plan usage')), {
      timeout: 4000,
    });
    expect(screen.getByText('Sign in with ChatGPT')).toBeInTheDocument();
  });

  it('signs out an account and reports revocation', async () => {
    const calls = mockBackend({ state: 'idle' });
    global.fetch = vi.fn().mockImplementation((url: string, opts?: RequestInit) => {
      calls.push(`${opts?.method ?? 'GET'} ${url}`);
      if (url === '/api/providers/chatgpt/accounts') return ok([account]);
      if (url.endsWith('/signout')) return ok({ revoked: true, accounts: [] });
      return ok({});
    }) as unknown as typeof fetch;
    const onStatus = vi.fn();
    render(<ChatGPTPlanPanel onStatus={onStatus} onError={vi.fn()} onAccountsChanged={vi.fn()} />);

    fireEvent.click(await screen.findByText('Sign out'));
    await waitFor(() => expect(onStatus).toHaveBeenCalledWith(expect.stringContaining('tokens revoked')));
    expect(calls).toContain(`POST /api/providers/chatgpt/accounts/${account.key}/signout`);
  });
});

describe('ProviderSettings ChatGPT plan tab', () => {
  it('shows plan billing and offers sign-in instead of a key field', async () => {
    global.fetch = vi.fn().mockImplementation((url: string) => {
      if (url === '/api/providers') {
        return ok({
          providers: [
            { route: 'offline', name: 'Offline Reviewed (Default)', configured: true, active: true, active_model: 'offline-curriculum', source: 'none', models_count: 1, requires_key: false, billing: 'none' },
            { route: 'chatgpt', name: 'ChatGPT Plan', configured: false, active: false, active_model: '', source: 'none', models_count: 0, requires_key: false, auth_kind: 'oauth', billing: 'chatgpt_plan' },
          ],
          active_route: 'offline',
          active_model: 'offline-curriculum',
          budget: { max_requests_per_session: 20, current_requests: 0, remaining_requests: 20, estimated_tokens: 0, cap_reached: false },
        });
      }
      if (url === '/api/providers/chatgpt/accounts') return ok([]);
      if (url.endsWith('/models')) return ok({ route: 'chatgpt', models: [], is_stale: true });
      return ok({});
    }) as unknown as typeof fetch;

    render(<ProviderSettings />);
    fireEvent.click(await screen.findByText('ChatGPT Plan', { selector: 'span' }));

    expect(await screen.findByText('Sign in with ChatGPT')).toBeInTheDocument();
    expect(screen.getByText('Billing: ChatGPT plan usage')).toBeInTheDocument();
    expect(screen.getByText('Not signed in.')).toBeInTheDocument();
    expect(document.querySelector('input[type="password"]')).toBeNull();
    expect(screen.getByText('Set as Active')).toBeDisabled();
  });
});

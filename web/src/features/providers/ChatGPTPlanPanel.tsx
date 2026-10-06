import React, { useCallback, useEffect, useRef, useState } from 'react';
import {
  ChatGPTAccountDTO,
  ChatGPTSignInStartDTO,
  ChatGPTSignInStatusDTO,
  ChatGPTSignOutDTO,
} from '../../types/providers';

const POLL_MS = 1500;

export interface ChatGPTPlanPanelProps {
  /** Called after sign-in, account switch or sign-out so the parent reloads provider summaries. */
  onAccountsChanged: () => void;
  onStatus: (message: string) => void;
  onError: (message: string) => void;
}

const panelStyle: React.CSSProperties = {
  background: 'var(--bg-app)',
  border: '1px solid var(--border-subtle)',
  borderRadius: 'var(--radius-sm)',
  padding: '0.75rem',
  display: 'flex',
  flexDirection: 'column',
  gap: '0.6rem',
};

const smallBtn: React.CSSProperties = { fontSize: '0.75rem', padding: '0.3rem 0.6rem' };

export const ChatGPTPlanPanel: React.FC<ChatGPTPlanPanelProps> = ({ onAccountsChanged, onStatus, onError }) => {
  const [accounts, setAccounts] = useState<ChatGPTAccountDTO[]>([]);
  const [signIn, setSignIn] = useState<ChatGPTSignInStatusDTO>({ state: 'idle' });
  const [authorizeUrl, setAuthorizeUrl] = useState<string | null>(null);
  const pollRef = useRef<number | null>(null);

  const stopPolling = () => {
    if (pollRef.current !== null) {
      window.clearInterval(pollRef.current);
      pollRef.current = null;
    }
  };

  const loadAccounts = useCallback(async () => {
    try {
      const res = await fetch('/api/providers/chatgpt/accounts');
      if (res.ok) setAccounts((await res.json()) || []);
    } catch {
      // offline fallback
    }
  }, []);

  useEffect(() => {
    loadAccounts();
    return stopPolling;
  }, [loadAccounts]);

  const finishAttempt = useCallback(
    async (st: ChatGPTSignInStatusDTO) => {
      stopPolling();
      setAuthorizeUrl(null);
      if (st.state === 'complete') {
        onStatus(`✓ Signed in to ChatGPT as ${st.account?.label ?? 'your account'}. Refresh the model list to choose a plan model.`);
      } else if (st.state === 'failed' || st.state === 'expired') {
        onError(st.error || 'ChatGPT sign-in did not complete.');
      }
      await loadAccounts();
      onAccountsChanged();
    },
    [loadAccounts, onAccountsChanged, onError, onStatus],
  );

  const poll = useCallback(async () => {
    try {
      const res = await fetch('/api/providers/chatgpt/signin');
      if (!res.ok) return;
      const st: ChatGPTSignInStatusDTO = await res.json();
      setSignIn(st);
      if (st.state !== 'pending') await finishAttempt(st);
    } catch {
      // transient; keep polling until the backend attempt times out
    }
  }, [finishAttempt]);

  const startSignIn = async (accountKey?: string) => {
    stopPolling();
    try {
      const res = await fetch('/api/providers/chatgpt/signin', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(accountKey ? { account_key: accountKey } : {}),
      });
      if (!res.ok) throw new Error((await res.text()) || 'Could not start ChatGPT sign-in');
      const data: ChatGPTSignInStartDTO = await res.json();
      setSignIn(data.status);
      setAuthorizeUrl(data.authorize_url);
      window.open(data.authorize_url, '_blank', 'noopener,noreferrer');
      pollRef.current = window.setInterval(poll, POLL_MS);
    } catch (err: unknown) {
      onError(err instanceof Error ? err.message : 'Could not start ChatGPT sign-in');
    }
  };

  const cancelSignIn = async () => {
    stopPolling();
    setAuthorizeUrl(null);
    try {
      const res = await fetch('/api/providers/chatgpt/signin/cancel', { method: 'POST' });
      if (res.ok) setSignIn(await res.json());
    } catch {
      setSignIn({ state: 'cancelled' });
    }
  };

  const selectAccount = async (key: string) => {
    try {
      const res = await fetch(`/api/providers/chatgpt/accounts/${encodeURIComponent(key)}/select`, { method: 'POST' });
      if (!res.ok) throw new Error(await res.text());
      setAccounts(await res.json());
      onAccountsChanged();
    } catch (err: unknown) {
      onError(err instanceof Error ? err.message : 'Could not switch ChatGPT account');
    }
  };

  const signOut = async (acct: ChatGPTAccountDTO) => {
    try {
      const res = await fetch(`/api/providers/chatgpt/accounts/${encodeURIComponent(acct.key)}/signout`, {
        method: 'POST',
      });
      if (!res.ok) throw new Error(await res.text());
      const data: ChatGPTSignOutDTO = await res.json();
      setAccounts(data.accounts || []);
      onStatus(
        data.revoked
          ? `✓ Signed out ${acct.label}; tokens revoked and removed from the vault.`
          : `Signed out ${acct.label} locally. OpenAI did not confirm revocation; you can also remove access from your ChatGPT settings.`,
      );
      onAccountsChanged();
    } catch (err: unknown) {
      onError(err instanceof Error ? err.message : 'Could not sign out');
    }
  };

  const pending = signIn.state === 'pending';

  return (
    <div style={panelStyle} id="chatgpt-plan-panel">
      <p style={{ margin: 0, fontSize: '0.8rem', color: 'var(--text-muted)' }}>
        <strong>Billing:</strong> requests use your ChatGPT plan&apos;s usage limits. This is separate from the
        OpenAI API route, which bills an API key. The tutor never switches between them on its own.
      </p>

      {accounts.length > 0 && (
        <ul style={{ listStyle: 'none', margin: 0, padding: 0, display: 'flex', flexDirection: 'column', gap: '0.4rem' }}>
          {accounts.map((a) => (
            <li
              key={a.key}
              style={{
                display: 'flex',
                flexWrap: 'wrap',
                alignItems: 'center',
                gap: '0.5rem',
                padding: '0.4rem 0.5rem',
                border: `1px solid ${a.selected ? 'var(--accent-emerald)' : 'var(--border-subtle)'}`,
                borderRadius: 'var(--radius-sm)',
              }}
            >
              <span style={{ flex: '1 1 12rem', fontSize: '0.85rem', overflowWrap: 'anywhere' }}>
                {a.selected && <span aria-label="selected account">● </span>}
                {a.label}
              </span>
              <span style={{ fontSize: '0.7rem', color: a.needs_reauth || !a.plan_granted ? 'var(--accent-amber)' : 'var(--text-muted)' }}>
                {a.needs_reauth ? 'Sign-in expired' : a.plan_granted ? 'Plan usage granted' : 'Plan usage not granted'}
              </span>
              {!a.selected && (
                <button type="button" className="btn btn-outline" style={smallBtn} onClick={() => selectAccount(a.key)}>
                  Use this account
                </button>
              )}
              {a.needs_reauth && (
                <button type="button" className="btn btn-secondary" style={smallBtn} disabled={pending} onClick={() => startSignIn(a.key)}>
                  Sign in again
                </button>
              )}
              <button type="button" className="btn btn-ghost" style={smallBtn} onClick={() => signOut(a)}>
                Sign out
              </button>
            </li>
          ))}
        </ul>
      )}

      {pending ? (
        <div role="status" style={{ display: 'flex', flexWrap: 'wrap', alignItems: 'center', gap: '0.5rem', fontSize: '0.8rem' }}>
          <span>Waiting for you to finish signing in on the ChatGPT page…</span>
          {authorizeUrl && (
            <a href={authorizeUrl} target="_blank" rel="noopener noreferrer">
              Open sign-in page
            </a>
          )}
          <button type="button" className="btn btn-outline" style={smallBtn} onClick={cancelSignIn}>
            Cancel
          </button>
        </div>
      ) : (
        <div>
          <button type="button" className="btn btn-primary" style={{ fontSize: '0.8rem' }} onClick={() => startSignIn()}>
            {accounts.length > 0 ? 'Add another ChatGPT account' : 'Sign in with ChatGPT'}
          </button>
        </div>
      )}

      <p style={{ margin: 0, fontSize: '0.75rem', color: 'var(--text-muted)' }}>
        🔒 Sign-in opens OpenAI&apos;s own page. Tokens go straight to the local backend vault and are never stored in the
        browser or logs. Plan usage must be granted on that page; a ChatGPT login alone is not enough.
      </p>
    </div>
  );
};

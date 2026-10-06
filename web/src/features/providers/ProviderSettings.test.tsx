import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { ProviderSettings } from './ProviderSettings';

describe('ProviderSettings Component', () => {
  const mockProvidersResponse = {
    providers: [
      {
        route: 'offline',
        name: 'Offline Reviewed (Default)',
        configured: true,
        active: true,
        active_model: 'offline-curriculum',
        source: 'none',
        models_count: 1,
        requires_key: false,
      },
      {
        route: 'anthropic',
        name: 'Anthropic Claude',
        configured: false,
        active: false,
        active_model: 'claude-3-5-sonnet-20241022',
        source: 'none',
        models_count: 3,
        requires_key: true,
      },
      {
        route: 'gemini',
        name: 'Google Gemini',
        configured: true,
        active: false,
        active_model: 'gemini-2.0-flash',
        source: 'vault',
        masked_key: 'AIzaSy...1234',
        models_count: 3,
        requires_key: true,
      },
      {
        route: 'openai',
        name: 'OpenAI',
        configured: false,
        active: false,
        active_model: 'gpt-4o',
        source: 'none',
        models_count: 3,
        requires_key: true,
      },
    ],
    active_route: 'offline',
    active_model: 'offline-curriculum',
    budget: {
      max_requests_per_session: 20,
      current_requests: 3,
      remaining_requests: 17,
      estimated_tokens: 1250,
      cap_reached: false,
    },
  };

  const mockGeminiModelsResponse = {
    route: 'gemini',
    models: [
      {
        id: 'gemini-2.0-flash',
        name: 'Gemini 2.0 Flash',
        provider: 'gemini',
        supports_streaming: true,
        context_window: 1048576,
        description: 'Fast multimodal model',
        is_default: true,
      },
      {
        id: 'gemini-1.5-pro',
        name: 'Gemini 1.5 Pro',
        provider: 'gemini',
        supports_streaming: true,
        context_window: 2097152,
        description: 'Deep reasoning model',
      },
    ],
    is_stale: false,
  };

  beforeEach(() => {
    vi.restoreAllMocks();
    global.fetch = vi.fn().mockImplementation((url: string, opts?: RequestInit) => {
      if (url === '/api/providers' && (!opts || opts.method === 'GET')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve(mockProvidersResponse),
        });
      }
      if (url.includes('/api/providers/offline/models')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({
            route: 'offline',
            models: [
              {
                id: 'offline-curriculum',
                name: 'Offline Verified Curriculum',
                provider: 'offline',
                supports_streaming: true,
                context_window: 8192,
                description: 'Offline curriculum',
                is_default: true,
              },
            ],
            is_stale: false,
          }),
        });
      }
      if (url.includes('/api/providers/gemini/models')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve(mockGeminiModelsResponse),
        });
      }
      if (url === '/api/providers/active') {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ status: 'ok', active_route: 'gemini', active_model: 'gemini-2.0-flash' }),
        });
      }
      if (url.includes('/credentials')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ route: 'gemini', configured: true, source: 'vault', masked_key: 'AIzaSy...5678' }),
        });
      }
      if (url.includes('/disconnect')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ route: 'gemini', configured: false, source: 'none' }),
        });
      }
      if (url === '/api/providers/budget/reset') {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({
            max_requests_per_session: 20,
            current_requests: 0,
            remaining_requests: 20,
            estimated_tokens: 0,
            cap_reached: false,
          }),
        });
      }
      return Promise.resolve({ ok: true, json: () => Promise.resolve({}) });
    }) as unknown as typeof fetch;
  });

  it('renders provider routes with offline active by default', async () => {
    render(<ProviderSettings />);

    await waitFor(() => {
      expect(screen.getByText('Offline Reviewed')).toBeInTheDocument();
      expect(screen.getByText('Google Gemini')).toBeInTheDocument();
      expect(screen.getByText('Anthropic Claude')).toBeInTheDocument();
      expect(screen.getByText('OpenAI API')).toBeInTheDocument();
      expect(screen.getByText('ChatGPT Plan')).toBeInTheDocument();
    });

    expect(screen.getByText('ACTIVE')).toBeInTheDocument();
    expect(screen.getByText(/3 \/ 20 requests used/)).toBeInTheDocument();
  });

  it('switches route tab and allows saving an API key', async () => {
    render(<ProviderSettings />);

    await waitFor(() => {
      expect(screen.getByText('Google Gemini')).toBeInTheDocument();
    });

    // Click Google Gemini tab
    fireEvent.click(screen.getByText('Google Gemini'));

    await waitFor(() => {
      expect(screen.getByText(/Google Gemini API Key/)).toBeInTheDocument();
    });

    // Enter API key
    const input = screen.getByPlaceholderText(/Enter new key to update|Paste API key/);
    fireEvent.change(input, { target: { value: 'AIzaSySecretTestKey5678' } });

    // Click Save Key
    const saveBtn = screen.getByText('Save Key');
    fireEvent.click(saveBtn);

    await waitFor(() => {
      expect(screen.getByText(/API key stored securely in backend vault/)).toBeInTheDocument();
    });

    // Key input should be cleared for security
    expect((input as HTMLInputElement).value).toBe('');
  });

  it('allows setting a provider as active', async () => {
    const onProviderChanged = vi.fn();
    render(<ProviderSettings onProviderChanged={onProviderChanged} />);

    await waitFor(() => {
      expect(screen.getByText('Google Gemini')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Google Gemini'));

    await waitFor(() => {
      expect(screen.getByText('Set as Active')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Set as Active'));

    await waitFor(() => {
      expect(onProviderChanged).toHaveBeenCalledWith('gemini', expect.any(String));
      expect(screen.getByText(/Active provider set to gemini/)).toBeInTheDocument();
    });
  });

  it('allows resetting session AI request budget', async () => {
    render(<ProviderSettings />);

    await waitFor(() => {
      expect(screen.getByText('Reset Budget')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Reset Budget'));

    await waitFor(() => {
      expect(screen.getByText(/Session AI request budget reset/)).toBeInTheDocument();
      expect(screen.getByText(/0 \/ 20 requests used/)).toBeInTheDocument();
    });
  });
});

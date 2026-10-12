import React, { useState, useEffect, useCallback } from 'react';
import {
  ProviderRoute,
  ProviderSummaryDTO,
  ModelInfoDTO,
  BudgetStatusDTO,
  ProvidersListResponseDTO,
  ModelsResponseDTO,
  BILLING_LABELS,
} from '../../types/providers';
import { ChatGPTPlanPanel } from './ChatGPTPlanPanel';

export interface ProviderSettingsProps {
  onProviderChanged?: (route: ProviderRoute, model: string) => void;
}

export const ProviderSettings: React.FC<ProviderSettingsProps> = ({ onProviderChanged }) => {
  const [providersList, setProvidersList] = useState<ProviderSummaryDTO[]>([]);
  const [activeRoute, setActiveRoute] = useState<ProviderRoute>('offline');
  const [activeModel, setActiveModel] = useState<string>('offline-curriculum');
  const [budget, setBudget] = useState<BudgetStatusDTO | null>(null);

  const [selectedRoute, setSelectedRoute] = useState<ProviderRoute>('offline');
  const [models, setModels] = useState<ModelInfoDTO[]>([]);
  const [selectedModel, setSelectedModel] = useState<string>('');
  const [isModelsStale, setIsModelsStale] = useState<boolean>(false);

  const [keyInput, setKeyInput] = useState<string>('');
  const [customModelInput, setCustomModelInput] = useState<string>('');
  const [statusMessage, setStatusMessage] = useState<string | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [isRefreshingModels, setIsRefreshingModels] = useState<boolean>(false);

  const [localEndpoint, setLocalEndpoint] = useState<string>('http://localhost:1234');
  const [isSavingEndpoint, setIsSavingEndpoint] = useState<boolean>(false);
  const [activeGuideTab, setActiveGuideTab] = useState<'lmstudio' | 'ollama' | 'faq'>('lmstudio');

  // Load providers and budget status
  const loadProviders = useCallback(async () => {
    setIsLoading(true);
    try {
      const res = await fetch('/api/providers');
      if (res.ok) {
        const data: ProvidersListResponseDTO = await res.json();
        setProvidersList(data.providers);
        setActiveRoute(data.active_route);
        setActiveModel(data.active_model);
        setBudget(data.budget);

        const loc = data.providers.find((p) => p.route === 'local');
        if (loc?.endpoint) {
          setLocalEndpoint(loc.endpoint);
        }
      }
    } catch {
      // offline fallback
    } finally {
      setIsLoading(false);
    }
  }, []);

  // Load models for selected route
  const loadModels = useCallback(async (route: ProviderRoute) => {
    try {
      const res = await fetch(`/api/providers/${route}/models`);
      if (res.ok) {
        const data: ModelsResponseDTO = await res.json();
        const loadedModels = data.models || [];
        setModels(loadedModels);
        setIsModelsStale(data.is_stale || false);
        if (loadedModels.length > 0) {
          const current = providersList.find((p) => p.route === route);
          setSelectedModel(current?.active_model || loadedModels[0].id);
        } else {
          // Never carry another route's model over to this one.
          setSelectedModel('');
        }
      }
    } catch {
      // offline fallback
    }
  }, [providersList]);

  useEffect(() => {
    loadProviders();
  }, [loadProviders]);

  useEffect(() => {
    loadModels(selectedRoute);
  }, [selectedRoute, loadModels]);

  const handleSetActive = async () => {
    setStatusMessage(null);
    setErrorMessage(null);
    try {
      const res = await fetch('/api/providers/active', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          route: selectedRoute,
          model: selectedModel,
        }),
      });

      if (!res.ok) {
        const errText = await res.text();
        throw new Error(errText || 'Failed to set active provider');
      }

      setActiveRoute(selectedRoute);
      setActiveModel(selectedModel);
      setStatusMessage(`✓ Active provider set to ${selectedRoute} (${selectedModel})`);
      if (onProviderChanged) {
        onProviderChanged(selectedRoute, selectedModel);
      }
      await loadProviders();
    } catch (err: unknown) {
      setErrorMessage(err instanceof Error ? err.message : 'Error setting provider');
    }
  };

  const handleSaveKey = async () => {
    if (!keyInput.trim()) return;
    setStatusMessage(null);
    setErrorMessage(null);

    const secretToSave = keyInput.trim();
    // Clear input field in state immediately for security
    setKeyInput('');

    try {
      const res = await fetch(`/api/providers/${selectedRoute}/credentials`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ api_key: secretToSave }),
      });

      if (!res.ok) {
        const errText = await res.text();
        throw new Error(errText || 'Failed to save API key');
      }

      setStatusMessage('✓ API key stored securely in backend vault.');
      await loadProviders();
      await loadModels(selectedRoute);
    } catch (err: unknown) {
      setErrorMessage(err instanceof Error ? err.message : 'Error saving key');
    }
  };

  const handleDisconnectKey = async () => {
    setStatusMessage(null);
    setErrorMessage(null);
    try {
      const res = await fetch(`/api/providers/${selectedRoute}/disconnect`, {
        method: 'POST',
      });

      if (!res.ok) {
        throw new Error('Failed to disconnect key');
      }

      setStatusMessage('✓ API key removed from backend vault.');
      await loadProviders();
    } catch (err: unknown) {
      setErrorMessage(err instanceof Error ? err.message : 'Error disconnecting key');
    }
  };

  const handleRefreshModels = async () => {
    setIsRefreshingModels(true);
    setStatusMessage(null);
    setErrorMessage(null);
    try {
      const res = await fetch(`/api/providers/${selectedRoute}/models/refresh`, {
        method: 'POST',
      });

      if (!res.ok) {
        const errText = await res.text();
        throw new Error(errText || 'Model discovery failed');
      }

      const data: ModelsResponseDTO = await res.json();
      setModels(data.models);
      setIsModelsStale(false);
      setStatusMessage(`✓ Discovered ${data.models.length} model(s) from official API catalog.`);
    } catch (err: unknown) {
      setErrorMessage(err instanceof Error ? err.message : 'Model discovery error');
    } finally {
      setIsRefreshingModels(false);
    }
  };

  const handleSaveEndpoint = async (overrideEndpoint?: string) => {
    const target = (overrideEndpoint || localEndpoint).trim();
    if (!target) return;
    setIsSavingEndpoint(true);
    setStatusMessage(null);
    setErrorMessage(null);
    try {
      const res = await fetch('/api/providers/local/endpoint', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ endpoint: target }),
      });
      if (!res.ok) {
        const errText = await res.text();
        throw new Error(errText || 'Failed to update local endpoint');
      }
      const data = await res.json();
      setLocalEndpoint(data.endpoint || target);
      setStatusMessage(`✓ Local LLM loopback endpoint updated to ${data.endpoint || target}`);
      await loadProviders();
    } catch (err: unknown) {
      setErrorMessage(err instanceof Error ? err.message : 'Error updating endpoint');
    } finally {
      setIsSavingEndpoint(false);
    }
  };

  const handleApplyPreset = (presetEndpoint: string) => {
    setLocalEndpoint(presetEndpoint);
    handleSaveEndpoint(presetEndpoint);
  };

  const handleAddCustomModel = async () => {
    if (!customModelInput.trim()) return;
    try {
      const res = await fetch(`/api/providers/${selectedRoute}/models/custom`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ model_id: customModelInput.trim() }),
      });

      if (res.ok) {
        const custom: ModelInfoDTO = await res.json();
        setModels((prev) => [custom, ...prev.filter((m) => m.id !== custom.id)]);
        setSelectedModel(custom.id);
        setCustomModelInput('');
        setStatusMessage(`✓ Added custom model ${custom.id}`);
      }
    } catch {
      // offline fallback
    }
  };

  const handleResetBudget = async () => {
    try {
      const res = await fetch('/api/providers/budget/reset', { method: 'POST' });
      if (res.ok) {
        const b: BudgetStatusDTO = await res.json();
        setBudget(b);
        setStatusMessage('✓ Session AI request budget reset.');
      }
    } catch {
      // offline fallback
    }
  };

  const currentSummary = providersList.find((p) => p.route === selectedRoute);

  return (
    <div className="provider-settings-container" style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
      {/* Route Selector Tabs */}
      <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap', alignItems: 'center' }}>
        {[
          { route: 'offline', label: 'Offline Reviewed' },
          { route: 'local', label: 'Local Offline LLM' },
          { route: 'anthropic', label: 'Anthropic Claude' },
          { route: 'gemini', label: 'Google Gemini' },
          { route: 'openai', label: 'OpenAI API' },
          { route: 'chatgpt', label: 'ChatGPT Plan' },
        ].map((tab) => {
          const isSelected = selectedRoute === tab.route;
          const isCurrActive = activeRoute === tab.route;
          return (
            <button
              key={tab.route}
              type="button"
              className={`btn ${isSelected ? 'btn-primary' : 'btn-outline'}`}
              onClick={() => setSelectedRoute(tab.route as ProviderRoute)}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: '0.4rem',
                padding: '0.45rem 0.75rem',
                fontSize: '0.85rem',
              }}
            >
              <span>{tab.label}</span>
              {isCurrActive && (
                <span
                  style={{
                    fontSize: '0.65rem',
                    background: 'var(--accent-emerald)',
                    color: '#fff',
                    padding: '0.1rem 0.35rem',
                    borderRadius: 'var(--radius-sm)',
                    fontWeight: 700,
                  }}
                  title={`Active model: ${activeModel}`}
                >
                  ACTIVE
                </span>
              )}
            </button>
          );
        })}
        {isLoading && <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>Updating...</span>}
      </div>

      {/* Selected Provider Card */}
      <div
        style={{
          background: 'var(--bg-surface-elevated)',
          border: '1px solid var(--border-subtle)',
          borderRadius: 'var(--radius-md)',
          padding: '1rem',
          display: 'flex',
          flexDirection: 'column',
          gap: '0.75rem',
        }}
      >
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <div>
            <h3 style={{ margin: 0, fontSize: '1.05rem', fontWeight: 600 }}>
              {currentSummary?.name || selectedRoute}
            </h3>
            <span style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>
              {selectedRoute === 'offline'
                ? 'Zero external network calls. Mathematical engine derivations & causal hints.'
                : selectedRoute === 'local'
                ? 'Loopback OpenAI-compatible server (LM Studio, Ollama). Zero external calls, completely private.'
                : selectedRoute === 'chatgpt'
                ? currentSummary?.account_label
                  ? `Signed in as ${currentSummary.account_label}${currentSummary.configured ? '' : ' (sign-in needs attention)'}`
                  : 'Not signed in.'
                : currentSummary?.configured
                ? `Key configured via ${currentSummary.source} (${currentSummary.masked_key})`
                : 'API key not configured.'}
            </span>
            {currentSummary?.billing && (
              <span
                id="provider-billing-label"
                style={{ display: 'block', fontSize: '0.75rem', color: 'var(--text-muted)' }}
              >
                Billing: {BILLING_LABELS[currentSummary.billing]}
              </span>
            )}
          </div>

          <div>
            {activeRoute === selectedRoute ? (
              <span style={{ color: 'var(--accent-emerald)', fontSize: '0.85rem', fontWeight: 600 }}>
                ✓ Currently Active
              </span>
            ) : (
              <button
                type="button"
                className="btn btn-secondary"
                onClick={handleSetActive}
                disabled={!selectedModel}
                title={selectedModel ? undefined : 'Choose a model first'}
                style={{ fontSize: '0.8rem', padding: '0.35rem 0.65rem' }}
              >
                Set as Active
              </button>
            )}
          </div>
        </div>

        {selectedRoute === 'chatgpt' && (
          <ChatGPTPlanPanel
            onAccountsChanged={loadProviders}
            onStatus={(m) => {
              setErrorMessage(null);
              setStatusMessage(m);
            }}
            onError={(m) => {
              setStatusMessage(null);
              setErrorMessage(m);
            }}
          />
        )}

        {/* Local LLM Configuration Section */}
        {selectedRoute === 'local' && (
          <div
            style={{
              background: 'var(--bg-app)',
              border: '1px solid var(--border-subtle)',
              borderRadius: 'var(--radius-sm)',
              padding: '0.85rem',
              display: 'flex',
              flexDirection: 'column',
              gap: '0.85rem',
            }}
          >
            {/* Header and Endpoint Switcher */}
            <div>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '0.5rem', marginBottom: '0.4rem' }}>
                <div>
                  <span style={{ fontSize: '0.85rem', fontWeight: 600 }}>Local Server Endpoint & Loopback Port</span>
                  <p style={{ margin: '0.15rem 0 0', fontSize: '0.75rem', color: 'var(--text-muted)' }}>
                    Antigravity connects via loopback HTTP to your local OpenAI-compatible server.
                  </p>
                </div>
                <div style={{ display: 'flex', gap: '0.35rem' }}>
                  <button
                    type="button"
                    className={`btn ${localEndpoint === 'http://localhost:1234' ? 'btn-primary' : 'btn-outline'}`}
                    onClick={() => handleApplyPreset('http://localhost:1234')}
                    style={{ fontSize: '0.75rem', padding: '0.25rem 0.55rem' }}
                    title="Switch to default LM Studio port (1234)"
                  >
                    LM Studio (:1234)
                  </button>
                  <button
                    type="button"
                    className={`btn ${localEndpoint === 'http://localhost:11434' ? 'btn-primary' : 'btn-outline'}`}
                    onClick={() => handleApplyPreset('http://localhost:11434')}
                    style={{ fontSize: '0.75rem', padding: '0.25rem 0.55rem' }}
                    title="Switch to default Ollama port (11434)"
                  >
                    Ollama (:11434)
                  </button>
                </div>
              </div>

              <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center', flexWrap: 'wrap' }}>
                <input
                  type="text"
                  value={localEndpoint}
                  onChange={(e) => setLocalEndpoint(e.target.value)}
                  placeholder="http://localhost:1234"
                  style={{
                    flex: '1 1 240px',
                    padding: '0.45rem 0.65rem',
                    fontSize: '0.85rem',
                    borderRadius: 'var(--radius-sm)',
                    border: '1px solid var(--border-subtle)',
                    background: 'var(--bg-surface)',
                    color: 'var(--text-main)',
                    fontFamily: 'monospace',
                  }}
                  aria-label="Local server endpoint URL"
                />
                <button
                  type="button"
                  className="btn btn-secondary"
                  onClick={() => handleSaveEndpoint()}
                  disabled={isSavingEndpoint}
                  style={{ fontSize: '0.8rem', padding: '0.45rem 0.75rem', whiteSpace: 'nowrap' }}
                >
                  {isSavingEndpoint ? 'Saving...' : 'Save Endpoint'}
                </button>
                <button
                  type="button"
                  className="btn btn-primary"
                  onClick={handleRefreshModels}
                  disabled={isRefreshingModels}
                  style={{ fontSize: '0.8rem', padding: '0.45rem 0.75rem', whiteSpace: 'nowrap' }}
                >
                  {isRefreshingModels ? 'Scanning...' : '🔄 Probe / Refresh Local Models'}
                </button>
              </div>
            </div>

            {/* Offline Privacy Guarantee Badge */}
            <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', background: 'var(--bg-surface)', padding: '0.5rem 0.75rem', borderRadius: 'var(--radius-sm)', borderLeft: '3px solid var(--accent-emerald)' }}>
              🔒 <strong>Offline Privacy Guarantee:</strong> Antigravity only connects to loopback addresses (<code>localhost</code>, <code>127.0.0.1</code>, <code>::1</code>). Zero external calls, zero API billing, complete offline privacy.
            </div>

            {/* Step-by-Step Setup Walkthrough Accordion / Tabs */}
            <div
              style={{
                background: 'var(--bg-surface)',
                border: '1px solid var(--border-subtle)',
                borderRadius: 'var(--radius-sm)',
                overflow: 'hidden',
              }}
            >
              <div
                style={{
                  display: 'flex',
                  borderBottom: '1px solid var(--border-subtle)',
                  background: 'var(--bg-surface-elevated)',
                }}
              >
                <button
                  type="button"
                  onClick={() => setActiveGuideTab('lmstudio')}
                  style={{
                    flex: 1,
                    padding: '0.5rem',
                    fontSize: '0.8rem',
                    fontWeight: activeGuideTab === 'lmstudio' ? 600 : 400,
                    border: 'none',
                    borderBottom: activeGuideTab === 'lmstudio' ? '2px solid var(--accent-primary)' : '2px solid transparent',
                    background: activeGuideTab === 'lmstudio' ? 'var(--bg-surface)' : 'transparent',
                    color: activeGuideTab === 'lmstudio' ? 'var(--accent-primary)' : 'var(--text-muted)',
                    cursor: 'pointer',
                  }}
                >
                  🖥️ LM Studio Setup Guide
                </button>
                <button
                  type="button"
                  onClick={() => setActiveGuideTab('ollama')}
                  style={{
                    flex: 1,
                    padding: '0.5rem',
                    fontSize: '0.8rem',
                    fontWeight: activeGuideTab === 'ollama' ? 600 : 400,
                    border: 'none',
                    borderBottom: activeGuideTab === 'ollama' ? '2px solid var(--accent-primary)' : '2px solid transparent',
                    background: activeGuideTab === 'ollama' ? 'var(--bg-surface)' : 'transparent',
                    color: activeGuideTab === 'ollama' ? 'var(--accent-primary)' : 'var(--text-muted)',
                    cursor: 'pointer',
                  }}
                >
                  🦙 Ollama Setup Guide
                </button>
                <button
                  type="button"
                  onClick={() => setActiveGuideTab('faq')}
                  style={{
                    flex: 1,
                    padding: '0.5rem',
                    fontSize: '0.8rem',
                    fontWeight: activeGuideTab === 'faq' ? 600 : 400,
                    border: 'none',
                    borderBottom: activeGuideTab === 'faq' ? '2px solid var(--accent-primary)' : '2px solid transparent',
                    background: activeGuideTab === 'faq' ? 'var(--bg-surface)' : 'transparent',
                    color: activeGuideTab === 'faq' ? 'var(--accent-primary)' : 'var(--text-muted)',
                    cursor: 'pointer',
                  }}
                >
                  ❓ Troubleshooting & FAQ
                </button>
              </div>

              <div style={{ padding: '0.85rem', fontSize: '0.82rem', lineHeight: 1.55 }}>
                {activeGuideTab === 'lmstudio' && (
                  <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                    <div style={{ fontWeight: 600, color: 'var(--text-main)' }}>
                      How to set up and run with LM Studio (Recommended for GUI users):
                    </div>
                    <ol style={{ margin: 0, paddingLeft: '1.25rem', display: 'flex', flexDirection: 'column', gap: '0.4rem' }}>
                      <li>
                        <strong>Download & Install:</strong> Download LM Studio from{' '}
                        <a href="https://lmstudio.ai" target="_blank" rel="noopener noreferrer" style={{ color: 'var(--accent-primary)', textDecoration: 'underline' }}>
                          lmstudio.ai
                        </a>{' '}
                        (Windows, macOS, or Linux).
                      </li>
                      <li>
                        <strong>Download a Model:</strong> In LM Studio, click the search icon (🔍) on the left sidebar. Recommended models for quantitative reasoning:
                        <ul style={{ margin: '0.2rem 0', paddingLeft: '1.1rem' }}>
                          <li><code>Llama-3.1-8B-Instruct-GGUF</code> (or 3.2-3B for lighter hardware)</li>
                          <li><code>Qwen2.5-7B-Instruct-GGUF</code> (excellent at mathematics)</li>
                          <li><code>Mistral-7B-Instruct-v0.3-GGUF</code></li>
                        </ul>
                      </li>
                      <li>
                        <strong>Open Developer / Server:</strong> Click the Developer tab on the left sidebar (icon with <code>&lt;-&gt;</code>).
                      </li>
                      <li>
                        <strong>Load Model into Memory:</strong> Click the top dropdown (<em>"Select a model to load"</em>) and choose your downloaded model.
                      </li>
                      <li>
                        <strong>Start Server:</strong> Ensure the port is <code>1234</code>, then click <strong>Start Server</strong>.
                      </li>
                      <li>
                        <strong>Connect Antigravity:</strong> Click the <strong>LM Studio (:1234)</strong> preset button above and click <strong>Probe / Refresh Local Models</strong>.
                      </li>
                    </ol>
                  </div>
                )}

                {activeGuideTab === 'ollama' && (
                  <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                    <div style={{ fontWeight: 600, color: 'var(--text-main)' }}>
                      How to set up and run with Ollama (Recommended for Command Line users):
                    </div>
                    <ol style={{ margin: 0, paddingLeft: '1.25rem', display: 'flex', flexDirection: 'column', gap: '0.4rem' }}>
                      <li>
                        <strong>Download & Install:</strong> Download Ollama from{' '}
                        <a href="https://ollama.com" target="_blank" rel="noopener noreferrer" style={{ color: 'var(--accent-primary)', textDecoration: 'underline' }}>
                          ollama.com
                        </a>.
                      </li>
                      <li>
                        <strong>Pull & Run a Model:</strong> Open your terminal, PowerShell, or Command Prompt and run:
                        <div style={{ background: 'var(--bg-app)', padding: '0.4rem 0.6rem', borderRadius: 'var(--radius-sm)', margin: '0.3rem 0', fontFamily: 'monospace', fontSize: '0.8rem', border: '1px solid var(--border-subtle)' }}>
                          ollama run llama3.1
                        </div>
                        <em>(Or try: <code>ollama run qwen2.5:7b</code> or <code>ollama run mistral</code>)</em>
                      </li>
                      <li>
                        <strong>Loopback API:</strong> Ollama automatically serves OpenAI-compatible endpoints on port <code>11434</code>.
                      </li>
                      <li>
                        <strong>Connect Antigravity:</strong> Click the <strong>Ollama (:11434)</strong> preset button above (or enter <code>http://localhost:11434</code>), click <strong>Save Endpoint</strong>, and then click <strong>Probe / Refresh Local Models</strong>.
                      </li>
                    </ol>
                  </div>
                )}

                {activeGuideTab === 'faq' && (
                  <div style={{ display: 'flex', flexDirection: 'column', gap: '0.7rem' }}>
                    <div>
                      <strong style={{ color: 'var(--text-main)' }}>Q: Why does probing say "cannot connect to local LLM server"?</strong>
                      <p style={{ margin: '0.2rem 0 0', color: 'var(--text-muted)' }}>
                        Ensure your local model runner is actively executing. In LM Studio, the green "Start Server" toggle must be ON. In Ollama, verify the background service is running by typing <code>ollama list</code> in your terminal.
                      </p>
                    </div>
                    <div>
                      <strong style={{ color: 'var(--text-main)' }}>Q: Why does probing say "no loaded models found"?</strong>
                      <p style={{ margin: '0.2rem 0 0', color: 'var(--text-muted)' }}>
                        In LM Studio, simply having a model downloaded is not enough; you must select it in the top dropdown of the Local Server tab to load it into memory. In Ollama, ensure you ran <code>ollama pull &lt;model&gt;</code>.
                      </p>
                    </div>
                    <div>
                      <strong style={{ color: 'var(--text-main)' }}>Q: Can I use vLLM, llama.cpp, or a custom port?</strong>
                      <p style={{ margin: '0.2rem 0 0', color: 'var(--text-muted)' }}>
                        Yes! Enter any OpenAI-compatible loopback URL (such as <code>http://127.0.0.1:8000</code> or <code>http://localhost:8080</code>) in the endpoint field and click <strong>Save Endpoint</strong>. Antigravity strictly enforces loopback (<code>localhost</code>, <code>127.0.0.1</code>, <code>::1</code>) to prevent data egress.
                      </p>
                    </div>
                  </div>
                )}
              </div>
            </div>
          </div>
        )}

        {/* API Key Section (for API-key routes) */}
        {selectedRoute !== 'offline' && selectedRoute !== 'local' && selectedRoute !== 'chatgpt' && (
          <div
            style={{
              background: 'var(--bg-app)',
              border: '1px solid var(--border-subtle)',
              borderRadius: 'var(--radius-sm)',
              padding: '0.75rem',
              display: 'flex',
              flexDirection: 'column',
              gap: '0.5rem',
            }}
          >
            <label style={{ fontSize: '0.85rem', fontWeight: 600 }}>
              {selectedRoute === 'anthropic' && 'Anthropic API Key (ANTHROPIC_API_KEY)'}
              {selectedRoute === 'gemini' && 'Google Gemini API Key (GEMINI_API_KEY)'}
              {selectedRoute === 'openai' && 'OpenAI API Key (OPENAI_API_KEY)'}
            </label>

            <div style={{ display: 'flex', gap: '0.5rem' }}>
              <input
                type="password"
                placeholder={currentSummary?.configured ? 'Enter new key to update...' : 'Paste API key...'}
                value={keyInput}
                onChange={(e) => setKeyInput(e.target.value)}
                style={{
                  flex: 1,
                  padding: '0.4rem 0.65rem',
                  borderRadius: 'var(--radius-sm)',
                  border: '1px solid var(--border-subtle)',
                  background: 'var(--bg-surface)',
                  color: 'var(--text-main)',
                  fontSize: '0.85rem',
                }}
              />
              <button
                type="button"
                className="btn btn-primary"
                onClick={handleSaveKey}
                disabled={!keyInput.trim()}
                style={{ fontSize: '0.8rem', padding: '0.4rem 0.75rem' }}
              >
                Save Key
              </button>
              {currentSummary?.configured && (
                <button
                  type="button"
                  className="btn btn-outline"
                  onClick={handleDisconnectKey}
                  style={{ fontSize: '0.8rem', padding: '0.4rem 0.75rem' }}
                >
                  Disconnect
                </button>
              )}
            </div>

            <p style={{ margin: 0, fontSize: '0.75rem', color: 'var(--text-muted)' }}>
              🔒 <strong>Protected Backend Vault:</strong> Keys are saved directly to your local backend vault and never stored in browser memory or logs.
            </p>
          </div>
        )}

        {/* Model Selection Dropdown & Discovery */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.4rem' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <label style={{ fontSize: '0.85rem', fontWeight: 600 }}>Model Selection:</label>
            {selectedRoute !== 'offline' && (
              <button
                type="button"
                className="btn btn-ghost"
                onClick={handleRefreshModels}
                disabled={isRefreshingModels || !currentSummary?.configured}
                style={{ fontSize: '0.75rem', padding: '0.2rem 0.5rem' }}
              >
                {isRefreshingModels ? 'Discovering...' : '🔄 Refresh Official Catalog'}
              </button>
            )}
          </div>

          <select
            value={selectedModel}
            onChange={(e) => setSelectedModel(e.target.value)}
            style={{
              padding: '0.45rem 0.65rem',
              borderRadius: 'var(--radius-sm)',
              border: '1px solid var(--border-subtle)',
              background: 'var(--bg-surface)',
              color: 'var(--text-main)',
              fontSize: '0.85rem',
            }}
          >
            {(models || []).map((m) => (
              <option key={m.id} value={m.id}>
                {m.name} ({m.id}) {m.is_custom ? '[Custom]' : ''}
              </option>
            ))}
          </select>

          {models.length === 0 && selectedRoute !== 'offline' && (
            <span style={{ fontSize: '0.7rem', color: 'var(--text-muted)' }}>
              No models discovered yet.{' '}
              {currentSummary?.configured
                ? 'Click "Refresh Official Catalog" to list models available to this account.'
                : selectedRoute === 'chatgpt'
                ? 'Sign in first, then refresh the catalog.'
                : 'Save an API key first, then refresh the catalog.'}
            </span>
          )}

          {isModelsStale && models.length > 0 && selectedRoute !== 'offline' && (
            <span style={{ fontSize: '0.7rem', color: 'var(--accent-amber)' }}>
              ℹ️ Catalog cache older than 24h. Click "Refresh Official Catalog" to update.
            </span>
          )}
        </div>

        {/* Custom Model ID Entry */}
        {selectedRoute !== 'offline' && (
          <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
            <input
              type="text"
              placeholder="Or enter custom model ID (e.g. gemini-exp-1206)..."
              value={customModelInput}
              onChange={(e) => setCustomModelInput(e.target.value)}
              style={{
                flex: 1,
                padding: '0.35rem 0.65rem',
                borderRadius: 'var(--radius-sm)',
                border: '1px solid var(--border-subtle)',
                background: 'var(--bg-surface)',
                color: 'var(--text-main)',
                fontSize: '0.8rem',
              }}
            />
            <button
              type="button"
              className="btn btn-outline"
              onClick={handleAddCustomModel}
              disabled={!customModelInput.trim()}
              style={{ fontSize: '0.75rem', padding: '0.35rem 0.65rem' }}
            >
              Add Custom
            </button>
          </div>
        )}
      </div>

      {/* Session Request & Token Budget Section */}
      {budget && (
        <div
          style={{
            background: 'var(--bg-surface)',
            border: '1px solid var(--border-subtle)',
            borderRadius: 'var(--radius-md)',
            padding: '0.75rem 1rem',
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
          }}
        >
          <div>
            <span style={{ fontSize: '0.85rem', fontWeight: 600, display: 'block' }}>
              Session AI Request Budget:
            </span>
            <span style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>
              {budget.current_requests} / {budget.max_requests_per_session} requests used (
              {budget.remaining_requests} remaining) · ~{budget.estimated_tokens} tokens estimated
            </span>
          </div>

          <button
            type="button"
            className="btn btn-ghost"
            onClick={handleResetBudget}
            style={{ fontSize: '0.75rem' }}
          >
            Reset Budget
          </button>
        </div>
      )}

      {/* Feedback Notifications */}
      {statusMessage && (
        <div
          style={{
            padding: '0.5rem 0.75rem',
            borderRadius: 'var(--radius-sm)',
            background: 'rgba(16, 185, 129, 0.1)',
            border: '1px solid var(--accent-emerald)',
            color: 'var(--accent-emerald)',
            fontSize: '0.85rem',
          }}
        >
          {statusMessage}
        </div>
      )}

      {errorMessage && (
        <div
          style={{
            padding: '0.5rem 0.75rem',
            borderRadius: 'var(--radius-sm)',
            background: 'rgba(244, 63, 94, 0.1)',
            border: '1px solid var(--accent-rose)',
            color: 'var(--accent-rose)',
            fontSize: '0.85rem',
          }}
        >
          ⚠️ {errorMessage}
        </div>
      )}
    </div>
  );
};

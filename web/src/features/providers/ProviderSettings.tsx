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

        {/* API Key Section (for API-key routes) */}
        {selectedRoute !== 'offline' && selectedRoute !== 'chatgpt' && (
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

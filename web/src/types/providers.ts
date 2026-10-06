export type ProviderRoute = 'offline' | 'anthropic' | 'gemini' | 'openai' | 'chatgpt';

/** Which account pays: none (offline), API usage billing, or a ChatGPT plan subscription. */
export type BillingRoute = 'none' | 'api_usage' | 'chatgpt_plan';

export interface ProviderSummaryDTO {
  route: ProviderRoute;
  name: string;
  configured: boolean;
  active: boolean;
  active_model: string;
  source: string;
  masked_key?: string;
  models_count: number;
  requires_key: boolean;
  auth_kind?: 'none' | 'api_key' | 'oauth';
  billing?: BillingRoute;
  account_label?: string;
}

export interface ModelInfoDTO {
  id: string;
  name: string;
  provider: string;
  supports_streaming: boolean;
  context_window: number;
  description: string;
  is_custom?: boolean;
  is_default?: boolean;
}

export interface BudgetStatusDTO {
  max_requests_per_session: number;
  current_requests: number;
  remaining_requests: number;
  estimated_tokens: number;
  cap_reached: boolean;
}

export interface ProvidersListResponseDTO {
  providers: ProviderSummaryDTO[];
  active_route: ProviderRoute;
  active_model: string;
  budget: BudgetStatusDTO;
}

export interface ModelsResponseDTO {
  route: ProviderRoute;
  models: ModelInfoDTO[];
  is_stale: boolean;
}

export interface ChatGPTAccountDTO {
  key: string;
  label: string;
  email?: string;
  plan_granted: boolean;
  needs_reauth: boolean;
  selected: boolean;
  expires_at: string;
  updated_at: string;
}

export type ChatGPTSignInState = 'idle' | 'pending' | 'complete' | 'failed' | 'cancelled' | 'expired';

export interface ChatGPTSignInStatusDTO {
  attempt_id?: string;
  state: ChatGPTSignInState;
  error?: string;
  account?: ChatGPTAccountDTO;
}

export interface ChatGPTSignInStartDTO {
  authorize_url: string;
  status: ChatGPTSignInStatusDTO;
}

export interface ChatGPTSignOutDTO {
  revoked: boolean;
  accounts: ChatGPTAccountDTO[];
}

export const BILLING_LABELS: Record<BillingRoute, string> = {
  none: 'No billing (offline)',
  api_usage: 'API usage billing',
  chatgpt_plan: 'ChatGPT plan usage',
};

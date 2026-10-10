# Provider connections

Verified from official documentation on **October 4, 2026**; ChatGPT plan protocol reverified **October 5, 2026** (see below). Reverify before implementing each adapter: availability, authentication, model catalogs, limits and request fields change. No credentials are committed. User-run ChatGPT plan inference succeeded October 5; macOS sign-in/sign-out succeeded October 9, 2026. The inference model was not recorded; see HANDOFF for evidence limits. See [SOURCES.md](SOURCES.md).

## Connection matrix

| Route | Planned support | Meaning |
|---|---|---|
| Offline | Required, default | Reviewed local hints/explanations; no login/network |
| Local LLM (LM Studio / Ollama) | Yes, planned | Loopback OpenAI-compatible endpoint (e.g. http://localhost:1234/v1); zero internet, zero cloud billing |
| ChatGPT plan | Yes, documented local/open-source flow | Eligible account must grant plan-use permission; identity alone is insufficient |
| OpenAI API key | Yes | User's API billing; separate adapter from ChatGPT plan |
| Anthropic API key | Yes | User's Claude API billing through Messages API |
| Gemini API key | Yes | User's Gemini API project/key, associated API quota/billing |
| Claude.ai consumer login | Not offered | Current official guidance does not permit third-party apps to offer this login/route consumer-plan credentials |
| Google OAuth for Gemini API | Optional later project-based connection | Requires configured Google Cloud OAuth/API project and applicable scopes; not demonstrated consumer Gemini plan access |
| Gemini consumer subscription login | Deferred/unverified | No reviewed official route establishes subscription inference for this tutor; do not reuse CLI credentials |

The UI may show a concise explanation of unavailable routes, but never a working-looking button whose action is unsupported. No prompt to paste subscription session tokens.

## Local LLM (LM Studio / Ollama) implementation

LM Studio and Ollama expose standard OpenAI-compatible REST APIs on loopback (`http://localhost:1234/v1` for LM Studio, `http://localhost:11434/v1` for Ollama). This adapter leverages the OpenAI adapter streaming SSE parser and request structures, targeting a configurable loopback base URL.

Key properties:
- **Zero external network calls:** All HTTP communication stays on loopback (`127.0.0.1`). Student responses, scenario data, and prompts never touch an external server.
- **No credentials required:** Bypasses API key validation or uses a static local placeholder.
- **Dynamic model discovery:** Queries `GET /v1/models` on the local port to auto-detect whatever model is currently loaded in memory.
- **True offline generative tutoring:** Enables generative hints, alternative explanations, and multi-turn chat completely offline.

## ChatGPT plan implementation

Use the local/open-source integration, not the separate limited-partner website identity flow. Persist this installation's host identifier. Initial authorization uses the documented dynamic-registration entry point; save the account/workspace's issued client ID, not the entry-point identifier. Use the app's own name and registration. Generate fresh PKCE/state/nonce, receive the callback on loopback, validate identity and granted plan permission, and store backend credentials. Keep account registrations separate, refresh/revoke per official metadata, and show the selected account/route. Recheck current eligible-account availability at implementation.

Implemented values (protocol reverified October 5, 2026; user-run live gate completed October 9):

- Discovery: `https://auth.openai.com/.well-known/openid-configuration`; issuer `https://auth.openai.com`; authorize `/api/accounts/authorize`; token `/api/accounts/oauth/token`; revoke `/api/accounts/oauth/revoke`; JWKS `/.well-known/jwks.json` (RS256); public client (token endpoint auth method `none`).
- Authorize: `client_id` (`dynamic_agent_client` for new registration, otherwise the saved issued `oaiapp_...` ID), `agent_name_hint` (new registration only), `ext_agent_host_id` (persisted `urn:uuid:`), optional `id_token_hint`, `response_type=code`, `redirect_uri=http://127.0.0.1:<port>/auth/callback` (fixed path, port may vary, not `localhost`), `scope=openid profile email offline_access resource.invoke chatgpt.tokens.use.direct`, `resource=https://api.openai.com/v1`, `state`, `nonce`, `code_challenge_method=S256`.
- The callback returns `code`, `state`, the issued `client_id` (new registration) and optional `scope`. Exchange and refresh POST form data with the issued `client_id` and `resource`; no client secret. Access tokens last 1 hour; refresh tokens last 30 days and rotate.
- Plan usage requires the granted `chatgpt.tokens.use.direct` scope. Inference: `POST https://api.openai.com/v1/responses` with the bearer token. Models: `GET /v1/models` returning `models[]` with `slug`, `display_name`, `visibility`.
- Errors handled per [errors and recovery](https://developers.openai.com/siwc/token-sharing-open-source/errors-and-recovery): `subscription_sharing_usage_limit_exceeded` (429: pause, no retry), `..._user_not_eligible` (403: no OAuth loop), `..._invalid_user` (401: sign in again), `..._usage_unavailable` / `..._user_unavailable` (503: bounded backoff), `..._unsupported_capability` (400), refresh `invalid_grant` / `refresh_token_*` (clear tokens, reauthorize with the saved client ID).

Sources: [registration/sign-in](https://developers.openai.com/siwc/token-sharing-open-source/sign-in), [local/open-source overview](https://developers.openai.com/siwc/token-sharing-open-source), [accounts and sessions](https://developers.openai.com/siwc/token-sharing-open-source/profiles-and-sessions).

The current plan HTTP route uses Responses with store=false, stream=true and array input; required history is sent explicitly. It disallows several usual parameters including max_output_tokens, temperature, top_p and previous_response_id over HTTP. Validate the complete current restrictions instead of copying API-key payloads. Source: [preview limitations](https://developers.openai.com/siwc/token-sharing-open-source/preview-limitations).

## API-key adapters

Backend-only environment variables: ANTHROPIC_API_KEY, GEMINI_API_KEY (optionally documented GOOGLE_API_KEY precedence), and OPENAI_API_KEY. Interactive fields send keys once to the local backend vault, clear frontend state afterwards, and receive only redacted status. No frontend VITE_* secrets or browser storage. Prefer OS vault; a session-only mode works when persistent storage is unavailable.

Use provider-specific request/stream/error/capability handling, not a universal endpoint assumption. Sources: [Claude API overview](https://platform.claude.com/docs/en/api/overview), [Gemini API keys](https://ai.google.dev/gemini-api/docs/api-key), [OpenAI text generation with Responses](https://developers.openai.com/api/docs/guides/text), [OpenAI API quickstart](https://developers.openai.com/api/docs/quickstart).

## Claude and Gemini login distinctions

Anthropic's current authentication guidance directs third-party product developers to API keys/supported cloud credentials and prohibits offering Claude.ai login in their own apps. Running an unmodified Claude Code product has a different contract and is not a substitute tutor adapter. Use the key route. Source: [authentication and credential use](https://code.claude.com/docs/en/legal-and-compliance#authentication-and-credential-use).

Google documents OAuth for Gemini API access through a configured Cloud project, consent screen and client. This is an API authorization route; the reviewed document does not establish consumer subscription entitlement. Implement it only as an explicitly configured optional adapter. Source: [Gemini OAuth guide](https://ai.google.dev/gemini-api/docs/oauth).

## Models, limits and diagnostics

Discover models through the selected route's official catalog; filter by actual text/stream/inference capabilities. Persist a local nonsecret catalog cache with fetched time/route/account identity where relevant; label stale results. Refresh is explicit. Allow a manually entered model ID with validation and clear unverified status. Model availability is account-specific; never guarantee that the same model works on every billing route.

Expose selected provider/account/model and plan/API route in settings and tutor panels. Rate-limit or usage-limit errors offer route-appropriate next actions. Never retry using another billing route silently. Diagnostics distinguish configured credentials, authenticated connection, catalog discovery and successful inference; a mock or diagnostic pass cannot establish live access.

## Acceptance evidence

Mock tests cover invalid/expired keys, missing scopes, OAuth state/nonce/PKCE mismatch, token renewal, cancellation, model filtering, API errors and usage limits. Live tests require configured explicit opt-in and record route, date and redacted outcome separately. Offline drills run before any provider initialization. No unattended model catalog fetch at startup when offline mode is selected.

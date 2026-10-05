# Provider connections

Verified from official documentation on **October 4, 2026**. Reverify before implementing each adapter: availability, authentication, model catalogs, limits and request fields change. This scaffold contains no credentials and makes no live-inference claim. See [SOURCES.md](SOURCES.md).

## Connection matrix

| Route | Planned support | Meaning |
|---|---|---|
| Offline | Required, default | Reviewed local hints/explanations; no login/network |
| ChatGPT plan | Yes, documented local/open-source flow | Eligible account must grant plan-use permission; identity alone is insufficient |
| OpenAI API key | Yes | User's API billing; separate adapter from ChatGPT plan |
| Anthropic API key | Yes | User's Claude API billing through Messages API |
| Gemini API key | Yes | User's Gemini API project/key, associated API quota/billing |
| Claude.ai consumer login | Not offered | Current official guidance does not permit third-party apps to offer this login/route consumer-plan credentials |
| Google OAuth for Gemini API | Optional later project-based connection | Requires configured Google Cloud OAuth/API project and applicable scopes; not demonstrated consumer Gemini plan access |
| Gemini consumer subscription login | Deferred/unverified | No reviewed official route establishes subscription inference for this tutor; do not reuse CLI credentials |

The UI may show a concise explanation of unavailable routes, but never a working-looking button whose action is unsupported. No prompt to paste subscription session tokens.

## ChatGPT plan implementation

Use the local/open-source integration, not the separate limited-partner website identity flow. Persist this installation's host identifier. Initial authorization uses the documented dynamic-registration entry point; save the account/workspace's issued client ID, not the entry-point identifier. Use the app's own name and registration. Generate fresh PKCE/state/nonce, receive the callback on loopback, validate identity and granted plan permission, and store backend credentials. Keep account registrations separate, refresh/revoke per official metadata, and show the selected account/route. Recheck current eligible-account availability at implementation.

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

# Local application and content security

These are concrete boundaries needed for a local browser app with provider credentials. Localhost is not automatically trusted: other sites/processes can attempt requests.

## Loopback service

Bind only loopback, use a random port by default, validate Host and expected Origin, and reject DNS rebinding/cross-origin mutation. Establish an unpredictable per-launch local application session using a short-lived one-use bootstrap code. If opening via URL, prefer a fragment consumed/removed by the UI, not a query logged or sent as a referrer. Exchange it for an HttpOnly SameSite application cookie and CSRF protection appropriate to the chosen local HTTP design. Do not place provider tokens in any URL.

Strictly separate OAuth callback acceptance from ordinary API origin/session checks. Validate OAuth state and registered loopback callback before exchanging a code. Serve only embedded/public UI assets, never arbitrary filesystem paths. Non-GET mutations require session/CSRF/origin validation and bounded body sizes. GET routes do not change state or trigger billed inference.

Development Vite proxy follows the same origin/session model; do not solve integration failures with permissive wildcard CORS. A --port override is not permission to bind publicly. Hosted deployment requires a new threat/authentication design.

## Credentials

Backend credential-vault interface prefers Windows Credential Manager/macOS Keychain/Linux secret service, selected and tested at implementation. If unavailable, support session-only credentials; persistent file fallback must be opt-in, use restrictive permissions, and be accurately described as unencrypted if it is. Learning DB and export formats never contain secrets.

Browser key inputs transmit a secret once over the authenticated local connection; clear input/state afterwards and never echo it. No localStorage/sessionStorage/IndexedDB, frontend environment variables, logs, error reports or Git for keys/refresh tokens. Environment variables are loaded only in the backend. Redact errors and diagnostics. Account selection stores only nonsecret display metadata in UI state.

Use the app's own official OAuth registration; no scraping browser cookies or another CLI's credential files. Verify identity/scopes before replacing an active connection. Refresh tokens and revocation behavior are provider-specific and must be sourced.

## Untrusted content

Treat AI replies, imported Markdown/CSV, bank proposals, filenames and syllabus text as data. Sanitize Markdown HTML and math-related markup; restrict link schemes, disable remote embeds by default and enforce a practical CSP after verifying MathJax requirements. Provider content cannot become shell/SQL/JS instructions. Never eval generated formula strings or accept model-provided grades.

Use safe export filenames/paths and parameterized SQL. Reject path traversal and archive-slip if imports later accept archives. Quote or neutralize spreadsheet-formula injection in CSV export and document import handling. File paths from generated notes are not export authorization.

## Data scope

No telemetry/cloud sync by default. Send only user-requested tutoring context. Do not include entire databases, unrelated notes or full course libraries in prompts. Keep consent and selected billing route visible. Backup/export/delete actions are explicit and recoverable where practical. Releases contain approved public curriculum only, not personal/course-private data.

Test rejected origins/hosts, missing session/CSRF, malicious Markdown/link/path inputs, duplicate/stale commands, secret redaction and unsupported login routes before distributing provider-enabled builds.

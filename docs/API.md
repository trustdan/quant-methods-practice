# Proposed local API and command contract

These routes are planned, not implemented. Freeze DTOs with strict parsers/tests in Stage 02. Use one authenticated local session and server revision for mutations. Do not expose canonical unresolved answers in question DTOs.

| Method / route | Purpose |
|---|---|
| GET /api/health | Minimal nonsecret runtime status/version |
| POST /api/local-session | Exchange one-use launcher bootstrap code for local application session |
| GET/PATCH /api/settings | Public preferences, no credential values |
| POST /api/practice/sessions | Create seeded session from approved content |
| GET /api/practice/sessions/{id} | Permitted current state, original order/status and drafts |
| POST /api/practice/sessions/{id}/commands | Submit, navigate, skip, reveal, reference, finish with command_id + expected_revision |
| GET /api/mastery | Rebuilt/current read-only concept summary |
| GET /api/history | Paginated immutable practice/exam records |
| POST /api/tutor/requests | Explicit hint/explanation/follow-up; derive context on server |
| GET /api/tutor/requests/{id}/events | Authenticated streamed events, not secrets or provider raw errors |
| DELETE /api/tutor/requests/{id} | Cancel upstream and mark final state |
| GET/POST /api/notes | Search/list and idempotent save |
| GET/DELETE /api/notes/{id} | Reopen or explicitly delete note |
| POST /api/exports | Export selected notes/history/bank/report via bounded export contract |
| GET /api/providers | Redacted configured routes/accounts/capabilities |
| POST /api/providers/{route}/credentials | Store supplied key in backend vault; never echo it |
| POST /api/providers/{route}/connect | Begin supported OAuth transaction |
| POST /api/providers/{route}/disconnect | Sign out/remove local credential after explicit action |
| GET /api/providers/{route}/models | Cached/provider model list with stale/capability metadata |
| POST /api/providers/{route}/models/refresh | Explicit authenticated discovery request |
| GET/POST /api/candidates | List/create untrusted proposals |
| POST /api/candidates/{id}/review | Explicit approve/reject/retire with reviewer provenance |
| POST/GET /api/exams | Create/list assessment sessions |
| POST /api/exams/{id}/commands | Submit/resume/abandon/complete under exam policy |
| GET /api/exams/{id}/report | Original content and results only after allowed completion |
| GET/POST /api/arcade/runs | Start/finalize human run and bounded score metadata |
| GET /api/arcade/high-scores | Local leaderboard |

## Practice mutations

Answer command specifies instance/stage ID, stable option ID or structured numeric response, unique command ID and expected session revision. Server re-derives eligibility and grade; browser must not send a trusted correct flag or mastery delta. Duplicate ID returns prior response; stale revision returns conflict/current state without duplicate mutation. One active writable tab per session or revision-based conflict handling is explicit in the UI.

Navigation command records relevant exposure before returning new content. Grade eligibility survives client manipulation, reload and retry. Browser state does not grant access to future hints/answers. In exams the same engine grades later, while the transport omits feedback/key material.

## Tutor events

Stream typed events: started, text_delta, complete, fallback, cancelled and error. Include request/session/instance/stage IDs to reject stale UI completions. Client streaming uses an authenticated fetch-compatible design; never put bearer credentials in stream URLs. An explicit cancel aborts upstream work; closing the panel also cancels unless a documented save/recovery transition is pending.

Errors use stable codes, concise messages and recoverable actions. Rate/usage limits cannot trigger another billing route. Cap request/body/response sizes and rate-limit local commands that initiate external calls. Security, save-on-leave and snapshot rules apply to every route, including CLI equivalents.

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

## Candidate endpoints (Stage 12 implemented)

These endpoints require the launcher's `quant_session` cookie; mutations additionally require the exact app Origin (or an explicitly configured development origin). Bodies are capped at 16 KB and reject unknown fields and trailing JSON documents.

- `GET /api/candidates`: persisted proposals, complete previews and review history (administrative answer material).
- `POST /api/candidates`: `{mode:"local"|"manual"|"ai", seed:integer, proposal?:object}`. Only manual mode accepts a proposal; AI uses the selected provider and rewrites fixed parameters' wording. Returns a pending candidate, never active content.
- `POST /api/candidates/{id}/review`: `{expected_revision, action:"approve"|"reject"|"retire", reviewer, notes, semantic_confirmed}`. Approval requires semantic confirmation. Stale/duplicate transitions return 409. Save failures preserve client review input.
- `GET /api/candidates/export`: active approved bank JSON download including bundled and local approved templates. No drafts, secrets or learner history.

Candidate review changes only the local content bank and append-only review audit. Practice DTOs continue withholding unresolved keys. Candidate content revisions are separate immutable drafts with new IDs; no edit/delete endpoint is exposed.

## Worksheet endpoints (Stage 13)

Require the launcher's `quant_session` cookie; POST also requires the exact app Origin (or configured development Origin). JSON bodies are capped at 128 KiB and reject unknown fields and trailing documents. CSV has its separate 64 KiB/500-row bound. No endpoint accepts paths or trusted canonical keys.

- `GET /api/worksheets`: saved public views, newest update first.
- `POST /api/worksheets`: `{mode:"full_solution", template_id, seed}` copies approved content. The CSV case takes `{mode:"dataset", csv, seed, reviewer, source_note, reviewed:true}` with full-case semantic confirmation; the theoretical reference is fixed by the server. Returns 201 and a key-withholding public worksheet.
- `POST /api/worksheets/dataset-preview`: `{csv}` returns validated rows and proposed case snapshots with keys/hints for explicit author review. No persistence or activation. Later practice records reference exposure.
- `GET /api/worksheets/{id}`: saved public view. A field's key/explanation appears only when that field completes.
- `POST /api/worksheets/{id}/commands`: `{command_id, expected_revision, type:"save_draft"|"submit", answers:{"instance_id:stage_id":{kind:"choice",option_id}|{kind:"numeric",numeric_raw}}}`. Submit requires all unfinished fields; invalid input rejects the whole form without attempts. Client normalization is discarded. Identical command replay returns its saved result; changed payload/stale revision returns 409. Completed fields/worksheets cannot be edited.
- `GET /api/worksheets/{id}/export`: safe Markdown attachment from saved public state, without unresolved keys.

See [worksheet scope and review policy](WORKSHEETS.md). Worksheet evidence is separate from independent mastery and bank activation.

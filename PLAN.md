# Staged implementation plan

Stage 00 through Stage 13 are **complete within their recorded scope** (Stage 12: binomial exactly-k candidates; Stage 13: approved-bank structured full solutions and four-toss CSV cases). Stage 11's live gate combines user-run sign-in/inference on October 5 and macOS sign-in/sign-out on October 9, 2026; the inference model was not recorded. Each stage needs implementation, relevant tests and a HANDOFF entry. Never mark a gate passed because source files exist. Stages are ordered for a working offline tutor before advanced features.

| Stage | Work | Required exit evidence |
|---|---|---|
| 00 Orientation and course inventory [COMPLETE] | Move to new repo, inspect existing files, read contract, record actual course sources and available tools. Preserve user work. | Source inventory and unknowns recorded; toolchains and module path established; no invented syllabus or inherited accounting bank. |
| 01 Runtime and dependency bootstrap [COMPLETE] | Choose current supported Go/Node toolchains and pin Go/React/Vite/TypeScript/MathJax/SQLite/test dependencies. Scaffold server, web shell, local assets, dev scripts and CI. | Compiled UI served by loopback Go server; production smoke with network disabled, math/font assets loaded locally; build/type/lint/unit commands established. |
| 02 Domain and bank contracts [COMPLETE] | Stable IDs, parameters/constraints, answer/assistance types, versions, schemas, family registry, strict parsing and draft validation. | Unknown fields/invalid parameters/unsupported families rejected; example agrees with runtime schema; active bank only admits approved records. |
| 03 Mathematical engine v0 [COMPLETE] | Binomial first, then set operations, Poisson and sums; stable numerics, event construction, canonical answers, reviewed distractors, rounding. | Reference fixtures, boundaries/properties, complement and dependence cases pass; engine has no UI/provider/storage imports. |
| 04 Seven-stage drill vertical slice [COMPLETE] | One approved binomial scenario, full scaffold, options/numeric answer, feedback/retry, recap, offline hints. Review content before activation. | Complete seven-stage drill offline; targeted misconception; 0.375, 37.5% and 3/8 grade under documented policy; invalid inputs are ungraded. |
| 05 Durable storage and replay [COMPLETE] | Migrations, sessions, original instances/options, attempts, assistance, idempotency, restart and backup. | Close/reopen restores drill; duplicate command cannot inflate evidence; changed bank cannot change historical content/grades. |
| 06 Navigation and browser usability [COMPLETE] | Complete keymap, stage strip/question list, focus, scroll, leave/save intent, multi-tab/reload semantics, help. | Browser tests cover shortcuts in views and editable fields; prior/future stage visits don't inflate evidence; navigation preserves unsent answers. |
| 07 Offline useful release [COMPLETE] | Reviewed small bank across learner's covered topics, module picker/settings, CLI essentials, bundled references, launch scripts. | Native offline ten-question session, quit/restart and replay; responsive math works on narrow and normal screens; tests/build pass. |
| 08 Scheduling and transfer [COMPLETE] | Concept evidence, decay, reviewed reasoning groups/pairs, weighted selection, full/four/two scaffolds, contrast queue. | One contribution per concept/instance; aid/retry/review distinct; delayed group-changing retrieval required; contrast bounded and cannot chain. |
| 09 Read-only tutor and note library [COMPLETE] | OfflineTutor/fake async provider, streams/cancellation, follow-ups, dirty explanation navigation, SQLite saves, search/export. | Slow/error/stale tutor can't alter state; save y/n/Esc and failure retry work; Markdown math round-trip; no model writes grades. |
| 10 API-key providers and discovery [COMPLETE] | Anthropic/Gemini/OpenAI adapters, key UI/vault, backend requests, dynamic model lists/cache, request budgets. | Mock auth/rate/timeout tests; provider-specific contract checks; optional live request per configured route recorded separately. |
| 11 ChatGPT plan sign-in [COMPLETE] | Reverify official docs; host identity, dynamic registration, PKCE/state/nonce, token validation/refresh/revoke, account/workspace selection, route indicators. | Mock protocol tests and actual eligible-account sign-in/inference/sign-out before claiming live support; no API-key fallback without selection. |
| 12 Creative candidate workflow [COMPLETE — binomial scope] | AI/local generation and manual proposals for four tested binomial exactly-k variations; strict parsing, Go-derived stages/keys, full preview, human approval provenance, activation/retirement/export, UI and CLI. Other candidate families fail closed. | Semantically wrong but numerically valid candidates cannot auto-publish; old snapshots survive edits; approval changes bank only. |
| 13 Full solutions and cases [COMPLETE — initial scope] | Structured full-form input over approved questions; generic four-toss CSV batches with empirical/model comparison, method/assumption/value/interpretation fields, explicit case review, drafts and Markdown worksheets. See [scope](docs/WORKSHEETS.md). | Shared grader agrees over all approved templates; bounded CSV and human-reviewed cases; atomic/idempotent submission, native restart/review, offline browser and full checks passed. Worksheet evidence stays separate from independent mastery. |
| 14 Exam mode | Timed/untimed assessment, aid policies, immutable original order, deadlines, resume, report/history/review, CLI flags. | Feedback/help withheld; expired timer persists across restart; changed bank resume retains original problems/options/answers. |
| 15 Continuous distributions and math visuals | Course-grounded next families, integrals/PDF/CDF, accessible Venn/PMF/CDF/PDF/shaded-area controls. | Each family's reviewed numerical/teaching tests; graphical area/values reconcile with engine; labels/keyboard/text alternatives checked. |
| 16 Sampling, CI, testing and regression | Course-grounded later modules, simulations, confidence coverage, hypothesis choices/tails, regression/residuals. Implement one reviewed family at a time. | Tests preserve population/sample and dependence distinctions; known datasets/fixtures; no claims of causal inference from association; exact syllabus methods recorded. |
| 17 Arcade first playable | Canvas fixed-step engine, original starfield/ship, attract/manual/skip, held-key steering+fire, pause, reduced motion. | Immediate skip at every state; keyup/blur safe; game cleanup leaves drills responsive; seeded simulation and browser smoke. |
| 18 Arcade complete parity | Autofire/bombs, acceleration, dual health, heavy hazards/fragments, four sectors/continuation, high scores/initials and leaderboard/title transition. | Scores persist; demo ineligible; simultaneous input works; deterministic collisions/bombs/transitions; small screen/reduced-motion behavior. |
| 19 Release hardening | Windows/macOS/Linux binaries with embedded UI/bank/fonts, launchers, backup/export, checksums, license notices, diagnostics/docs. | Extracted packages run offline natively, no Go/Node required; no personal data/secrets shipped; full release test matrix recorded. |
| 20 Content quality and ongoing breadth | Review transfer groups, distractions, amount/event bounds, candidate quality, course alignment and bank breadth in small batches. | Coverage/repetition review before count increases; every new family has reviewed rules/tests; empirical learning claims remain separate. |

## Priorities and dependencies

Stages 00-07 establish useful offline practice. Stages 08-14 carry over learning evidence, AI, saved notes, creative review, complete solutions and exams. Stages 15-16 need actual course details for specific methods. Their source-dependent portions may remain open while generic verified features and Stage 17-18 game work proceed; document the reason and unmet gates.

Arcade work depends on a stable browser shell, navigation and score storage. It must not delay the first offline release. If the user explicitly reprioritizes the game, a thin skippable attract screen may be added after Stage 07, with later arcade gates still open. Do not claim game mechanics teach probability unless an actual learning design is reviewed.

No fixed deadline is promised. Each stage may use several reviewable changes. End a session with the next concrete task, actual checks and known limitations in HANDOFF.

## Backlog items

Four explicit enhancements are prioritized across the roadmap:

1. **Relevant Excel formulas in answer explanations**:
   - Provide standard spreadsheet equivalents in engine derivations, drill recaps, stage explanations, reference library cards, and AI tutor prompts (e.g. `=BINOM.DIST(k, n, p, FALSE/TRUE)`, `=POISSON.DIST(...)`, `=COMBIN(n, k)`, `=NORM.DIST(...)`).
   - Connects analytical math derivations with standard course spreadsheet tools; integrates with Stage 03/04 math engine expansions and Stage 13 casework.
2. **Interactive multi-turn AI tutor chat**:
   - Conversational thread persistence in SQLite (`tutor_threads` and `tutor_messages` tables), streaming SSE deltas across ongoing threads, sliding-window token management, and budget tracking.
   - Strict pedagogical system prompt guardrails preventing answer leakage on unresolved active stages.
   - Dual save capability: save complete chat transcripts or individual selected AI responses to the personal note library.
3. **ChatGPT plan sign-in (Stage 11 [COMPLETE])**:
   - Official OpenAI local/open-source token-sharing OAuth integration with PKCE (`S256`), dynamic client registration, state/nonce validation, loopback redirect (`http://localhost:<port>/auth/callback`), token refresh/revocation, account/workspace selection, and isolated Responses route.
4. **Local offline LLM provider (LM Studio / Ollama)**:
   - Loopback OpenAI-compatible adapter (`http://localhost:1234/v1` or configurable host/port).
   - Auto-discovery of loaded local models via `GET /v1/models`.
   - Zero internet calls, zero API billing, complete privacy, enabling true offline generative tutoring and chat.


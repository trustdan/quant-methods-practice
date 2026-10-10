# Requirements

Status: implemented through Stage 13 within the scope recorded in PLAN/HANDOFF; later stages remain planned. IDs are stable traceability anchors; completion requires the linked stage's recorded evidence, not a checkbox in this file. See [feature parity](docs/FEATURE-PARITY.md) for the accounting reference.

## Local runtime and mathematics

| ID | Requirement and acceptance |
|---|---|
| R01 | Go local server + React/TypeScript browser UI + SQLite. Packaged offline app starts without Go, Node, provider credentials, or network; existing browser required. Loopback only by default. |
| R02 | Markdown/LaTeX with local MathJax and fonts. Greek letters, unions/intersections, fractions, sums, integrals, cases, and matrices render on startup, dynamically, in notes and exam reviews without a CDN. Math remains selectable/accessible where supported. |
| R03 | Engine computes canonical results and explanations with versioned rules. Support set operations, conditional probability as course evidence allows, binomial, Poisson, expectation/variance of sums first; later topics obey COURSE-MAP. |
| R04 | Numeric input accepts explicit supported forms such as decimal, percent, and rational fraction. Units, exactness, rounding, and tolerances are visible and tested. Symbolic input is a later restricted feature, never a promise to grade arbitrary LaTeX. |
| R05 | Seeded generation uses finite reviewed parameter sets/constraints, stable option IDs, reviewed distractors and provenance. Same inputs recreate the same instance. Invalid/unsupported candidates fail closed. |

## Practice and navigation

| ID | Requirement and acceptance |
|---|---|
| R06 | One scenario supports a coherent seven-stage sequence, reduced four/two-stage scaffolds, one short decision per stage, recap, and worked solution. Stages are family-specific, not filler questions to reach seven. |
| R07 | First error leads to causal hint, one assisted retry, then explanation. Help/reference use is recorded. Reviewed guided contrasts can follow eligible errors without recursive remediation or unbounded session extension. |
| R08 | Vim-like j/k selection and scrolling, h/l or arrows between stages, direct question navigation, Enter submission, numbered choices, gg/G on reading views, ?/e help, Esc cancellation. All shortcuts respect text entry, modal focus, browser shortcuts and saved-note prompts. |
| R09 | Previous/next problems, stage navigation and jump list preserve work; revisiting a graded/revealed stage cannot become fresh independent evidence. Leaving an unfinished problem records skipped/pending status without inventing a grade. |
| R10 | Configurable session size, module/concept focus and standard/spaced/intensive/transfer intensity; seeded mixed review and anti-repeat selection. Settings persist; no fixed bank-size target replaces quality review. |
| R11 | Mastery view separates independent, assisted, new and delayed evidence. Time decay is a scheduling heuristic, reconstructable from events. Cosmetic variants do not graduate concepts. |
| R12 | Durable sessions and immutable attempts survive reload, backend restart, and bank update. Multi-tab conflicts cannot double-grade. Replay includes original questions, options, assistance and policy versions. |

## AI and saved content

| ID | Requirement and acceptance |
|---|---|
| R13 | OfflineTutor first; optional model-agnostic adapters. Explicit hint/explain/follow-up requests show streaming progress, allow cancellation, enforce time/size/request limits, and fall back visibly to reviewed offline help. |
| R14 | Continue with ChatGPT uses the documented local/open-source OAuth flow and separately verifies plan permission. Anthropic, Google Gemini and OpenAI API keys can be entered locally or loaded from backend environment variables. Billing route/model/account are visible. See PROVIDERS for unsupported consumer logins. |
| R15 | Live model discovery, provider-specific capability filters, manual model selection, cache and refresh; stale lists are labeled. Do not pin a permanent model snapshot or show unavailable models as confirmed. |
| R16 | Save AI explanation when leaving it, using y/n/Esc or buttons; scrolling does not prompt. Failed saving preserves the note and intended navigation. Local library supports search/topic filters, reopen, Markdown export, explicit deletion and follow-up context. |
| R17 | Save raw Markdown/LaTeX, originating immutable question/stage, provider/model/route, time and assistance context. Exports open in Typora/Obsidian, include no secrets, and label AI notes as advisory. Chat-only learning does not write mastery. |
| R18 | Generate AI question candidate, local variation, preview, validate, approve/reject/retire and export bank. Content approval is separate from attempts. Unsupported families require code/rule review before activation; generated solutions are not answer keys. |
| R30 | Multi-turn AI tutor chat: Conversational threads with SQLite thread persistence, streaming SSE, sliding context budgets, strict answer-withholding on unresolved stages, and selective note exports. |
| R31 | Local offline LLM adapter: OpenAI-compatible loopback adapter for LM Studio/Ollama on local port with model discovery, no API key required, and zero network calls beyond loopback. |

## Assessment, visuals and tools

| ID | Requirement and acceptance |
|---|---|
| R19 | Timed/untimed exams, allowed-aid policy, withheld hints/results, interruption/resume, abandonment, history/report and post-completion question review. Original snapshots and wall-clock deadline survive restart. |
| R20 | Statistics equivalents of full journal/case practice: structured full solutions and multi-part datasets/cases, checking method, parameters, calculation and interpretation. Progressive and full-solution grading share engine rules. Initial Stage 13 scope: approved-bank structured choices/numbers and reviewed four-toss empirical/model CSV cases, with durable drafts/submission/review and Markdown worksheets; see [scope and limits](docs/WORKSHEETS.md). |
| R21 | Venn regions, discrete PMF/CDF, continuous PDF/CDF/areas, sampling simulations, CI coverage, test-tail diagrams and regression/residual plots. Figures have units, assumptions, textual summaries and keyboard-accessible controls. |
| R22 | Reference library/glossary, reviewed offline explanations, course notation/parameterization and module objectives. Reference opening marks assistance on the active unresolved problem. Exam aids follow an explicit policy. |
| R23 | CLI equivalents for bank validation/reconciliation, session defaults, data-dir, seed, mastery/history/export, exams, candidates, provider/model diagnostics, high scores and skip-intro. Mutating content commands require explicit IDs/actions; diagnostics redact credentials. |
| R24 | Backup/export/import plan, migration backup, failure recovery, per-profile data path, question bank versioning and native packaging for Windows/macOS/Linux. No automatic cloud sync or telemetry. |
| R29 | Excel formula equivalents: Mathematical engine derivations, drill recaps, stage explanations, reference library and tutor prompts provide standard Excel function formulas (e.g. BINOM.DIST, POISSON.DIST, COMBIN, NORM.DIST) matching canonical results. |

## Arcade, accessibility and reliability

| ID | Requirement and acceptance |
|---|---|
| R25 | Startup space flight with autoplay attract mode, manual takeover, simultaneous steering/fire, autofire, bombs, starfield, escalating difficulty, heavy/fragmenting hazards, dual health, four-sector completion and continue/retry. Original assets and quant-themed labels. |
| R26 | Scores and three-initial record entry persist locally, human play only; leaderboard from startup/practice/CLI. Enter/Esc or visible Skip enters drills immediately; no arcade result changes mastery. Skip preference and reduced-motion behavior persist. |
| R27 | Responsive layout, keyboard/pointer equivalence, sensible focus, visible shortcuts, color-independent feedback, readable zoom, reduced motion and screen-reader summaries. No accidental shortcut interception inside numeric, note, key, or initials fields. |
| R28 | Offline end-to-end smoke, math boundary/property tests, SQLite reopen/migration tests, browser focus/navigation tests, provider cancellation/failure tests and arcade simulation tests. Cross-platform release/live-provider claims require actual evidence. |

## Scope and deferred decisions

The complete roadmap includes all rows above. First release prioritizes R01-R12, R22 and relevant reliability checks; later stages deliver providers, notes, full solutions, exams, figures and arcade. Standalone hosted SaaS, remote authentication, cloud sync, telemetry, arbitrary computer algebra, native mobile clients and a desktop wrapper are outside the initial design.

Full syllabus and exam conventions, exact advanced distribution/test list, license, repository remote, and toolchain versions remain open. Resolve ordinary technical choices without repeatedly asking permission; request course-specific facts only when they actually block alignment.

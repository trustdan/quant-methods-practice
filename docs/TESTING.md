# Verification plan

Record actual commands/results in HANDOFF. The scaffold validator checks documents only. Application test tools are selected/pinned at Stage 01; no dependency or application test is claimed to have run yet.

## Test layers

| Layer | Meaningful coverage |
|---|---|
| Go domain/mathengine | Exact fixtures, probability bounds, normalization, moments, covariance, tail translation, degenerate inputs, extreme numeric stability |
| Bank/drill | Strict schema, unknown fields/families, approved-only loading, deterministic generation/options, wrong-event/distractor handling, finite parameter coverage |
| Mastery | One concept contribution per instance, hint/reference/retry/review separation, clock rollback/decay, distinct settings, delayed retrieval, bounded contrasts |
| SQLite | Fresh/migrated/reopened DB, backup failure, idempotent commands, immutable snapshots, exam resume and note recovery |
| HTTP | Unresolved keys withheld, session/revision handling, CSRF/origin/host protection, safe export/import, redacted errors |
| Tutor/provider | Fake streaming/slow/error/cancel/stale completion, bounded output/history/budget, credential lifecycle, model catalogs, route-specific payloads |
| Frontend units | Central command resolver, field/IME exclusions, leave-intent/save state machine, math lifecycle, plot-event bounds |
| Browser end-to-end | Real local server + temporary DB + fake providers: full drills, keyboard/focus, reload/restart, notes, exam withholding, game skip |
| Arcade simulation | Seeded physics/input/collisions/bombs/caps, state transitions, demo exclusion and scoring |

## First vertical slice checklist

Complete seven-stage binomial problem with choices and numeric answer. Trigger missing-combination or wrong-event hint, retry and reveal. Test equivalent numeric input, invalid syntax and explicit percentage semantics. Navigate previous stages/questions without losing drafts or obtaining new independent credit. Quit/reload/restart and verify original attempts/options. With network blocked, math and reviewed help still render and no provider discovery occurs.

## Provider and note checklist

Explicitly request an explanation; display streaming math; cancel it and navigate before completion. Verify stale content does not attach to a new stage. Save y resumes intended action; n leaves without library save; Esc remains; scrolling never prompts. Simulate DB failure and preserve note; recover unsaved draft after restart. Export .md and open it in a math-aware editor. Confirm no credentials in DB, export, logs, bundle or diff.

Live tests are separately opt-in. Record ChatGPT plan versus OpenAI API, Anthropic API and Gemini API independently. Mock success is not an OAuth/inference claim; a successful model list is not a successful tutor request. Never include live secrets in CI.

## Exams and release

Start a timed exam, leave/restart, alter bank and resume; original questions/order/answers remain and time does not reset. Withhold help/results during exam. Full solution/report uses correct original grading policy. Test incomplete snapshot failure without regeneration.

Build all target binaries and verify archive checksums/contents. Native Windows/macOS/Linux extracted-package checks are separate from cross-compilation. Smoke with Go/Node absent and network disabled, local math/fonts, browser launch fallback, data path and clean shutdown. Check readability at 200% zoom, narrow screen, keyboard-only and reduced-motion modes. Keep learning-effectiveness evaluation separate from correctness and usability.

Once established, normal checks include gofmt, go test ./..., go vet ./..., frontend typecheck/lint/unit/build, focused browser tests and scaffold validation. Repeat broader checks only when changes or failures justify it.

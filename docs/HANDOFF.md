# Handoff

## Current state - October 4, 2026

Move-ready documentation scaffold prepared inside the learner-created quant-methods-practice folder. No application implementation, manifests/lockfiles, initialized nested Git repository, active content, provider credentials, learner history or release artifacts. Source directories contain .gitkeep placeholders.

Created root requirements/overview/plan and shared Codex/Claude instruction entry points; technical contracts for architecture, numerics, course evidence, bank/review, pedagogy, mastery, navigation, local API/security/storage, AI notes/provider connections, figures, arcade, testing and release. Added strict starting schemas, an explicitly unapproved seven-stage binomial reference fixture and a stdlib-only scaffold validator.

The accounting app was inspected as a feature reference. No accounting code/bank/data or personal files were copied. User's supplied module headings and covered topics are the only quant course evidence; syllabus/slides/datasets remain absent.

## Checks

- `python scripts/validate_scaffold.py` passed: 31 Markdown files, 49 local links, three JSON files and 35 planned source directories; seven-stage draft shape/evidence and exact small binomial calculation checked.
- All three JSON documents parse; all 11 internal JSON Schema references resolve. Python validator source parses with ast.parse. A full JSON Schema implementation was not installed or run; starting schemas require runtime reconciliation/validation at Stage 02.
- An initial check caught missing placeholder directories and incorrect LaTeX JSON escaping; both were corrected before the passing run. No active content was created.
- Local links resolve inside the scaffold and no machine-specific absolute paths were found. Parent Git status showed only the new folder before its handoff update. No nested Git initialization, dependency installation or remote publication.
- Final validator run from the parent working directory also passed. All 72 scaffold files passed UTF-8/final-newline/trailing-whitespace checks; decoded LaTeX contains no unexpected control characters. Parent git diff --check passed with the existing LF/CRLF advisory; final changes are this folder plus the parent handoff record.
- No Go/frontend/application/provider/browser/native-release check applies until implementation exists. Draft fixture is not approved by creating or validating it.

## Next action

Move the whole folder to the new repo location. Start Stage 00: inspect toolchains and course sources, establish repository/module/license decisions, then Stage 01 builds the local Go + browser shell with fully local math assets. Read README/contract/requirements/plan before edits. Do not attempt the whole roadmap in the first session.

## Unresolved external gates

Actual course notation/objectives/exam rules; dependency versions; actual ChatGPT-plan account access and native vault/runtime tests; optional Google project OAuth; advanced test/regression methods; license/use rights and empirical learning effectiveness. Current official docs do not support offering a third-party Claude.ai consumer login; Gemini consumer subscription inference is unverified.

# Quant Methods Practice

A planned local, keyboard-driven probability and statistics tutor with readable mathematics, progressive problems, persistent practice history, optional AI help, saved explanations, and a skippable space arcade.

**Status: documentation scaffold, October 4, 2026. No runnable application or approved question bank yet.** Move this entire folder into its own repository before implementation. All project links and instructions are self-contained; the accounting repository is reference material, not a runtime dependency.

## Start building

1. Read [OVERVIEW.md](OVERVIEW.md), [REQUIREMENTS.md](REQUIREMENTS.md), and [docs/AGENT-CONTRACT.md](docs/AGENT-CONTRACT.md).
2. Resume from [docs/HANDOFF.md](docs/HANDOFF.md), then perform the first unfinished stage in [PLAN.md](PLAN.md).
3. Inventory actual course materials using [docs/COURSE-MAP.md](docs/COURSE-MAP.md). The module headings supplied by the learner are evidence of broad scope; they are not a full syllabus.
4. Initialize this folder as a separate Git repository. Choose the license and remote there; neither is assumed here.
5. Use [docs/BOOTSTRAP.md](docs/BOOTSTRAP.md) as the implementation checklist. Select and lock dependencies during Stage 01. Empty source directories are intentional.

Codex reads [AGENTS.md](AGENTS.md); Claude Code reads [CLAUDE.md](CLAUDE.md). Both point to the same contract. Suggested opening prompt:

> Read AGENTS.md, README.md, OVERVIEW.md, REQUIREMENTS.md, PLAN.md, and docs/HANDOFF.md. Follow docs/AGENT-CONTRACT.md. Complete the earliest unfinished stage whose dependencies are satisfied, preserve existing work, run its checks, and update the handoff. Start with the offline application; do not implement the whole roadmap at once.

## Selected design

- Go owns mathematical rules, grading, scheduling, SQLite persistence, provider requests, and local files.
- TypeScript + React + Vite provides a browser interface with vim-like navigation.
- Markdown with LaTeX math is rendered by a bundled local MathJax installation.
- SVG supplies accessible statistics diagrams; Canvas 2D supplies the arcade.
- A local Go process serves the compiled UI on loopback and opens the browser. No remote hosting, login, Node runtime, or internet is required for packaged offline drills.
- Provider tokens stay in the backend. ChatGPT sign-in and paid API-key connections are separate routes. Unsupported subscription connections are not offered as working features.

These are chosen architectural defaults, not installed or tested dependencies. A desktop wrapper can be evaluated after the browser release works.

## Documentation map

| Document | Purpose |
|---|---|
| [REQUIREMENTS.md](REQUIREMENTS.md) | Full feature inventory and acceptance criteria |
| [PLAN.md](PLAN.md) | Ordered implementation stages and evidence gates |
| [docs/FEATURE-PARITY.md](docs/FEATURE-PARITY.md) | Accounting feature to statistics feature mapping |
| [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) | Packages, ownership, and runtime boundaries |
| [docs/DIRECTORY-STRUCTURE.md](docs/DIRECTORY-STRUCTURE.md) | Actual scaffold and eventual file placement |
| [docs/COURSE-MAP.md](docs/COURSE-MAP.md) | Course evidence, confirmed topics, and unknowns |
| [docs/PEDAGOGY.md](docs/PEDAGOGY.md) | Seven-stage examples, feedback, scaffolding |
| [docs/QUESTION-BANK.md](docs/QUESTION-BANK.md) | Deterministic generation, review, provenance |
| [docs/NUMERICS.md](docs/NUMERICS.md) | Mathematical rules and grading tolerances |
| [docs/NAVIGATION.md](docs/NAVIGATION.md) | Keyboard, focus, stage and question navigation |
| [docs/MASTERY.md](docs/MASTERY.md) | Evidence and delayed transfer scheduling |
| [docs/STORAGE.md](docs/STORAGE.md) | History, migrations, saved notes and exports |
| [docs/API.md](docs/API.md) | Proposed local API and state transitions |
| [docs/AI-TUTOR.md](docs/AI-TUTOR.md) | Hints, explanations, follow-ups and saving |
| [docs/PROVIDERS.md](docs/PROVIDERS.md) | Verified sign-in options and live-check gates |
| [docs/MATH-AND-VISUALS.md](docs/MATH-AND-VISUALS.md) | Math rendering and interactive figures |
| [docs/ARCADE.md](docs/ARCADE.md) | Startup flight game design and staged work |
| [docs/SECURITY.md](docs/SECURITY.md) | Local HTTP, credentials and untrusted content |
| [docs/TESTING.md](docs/TESTING.md) | Required automated and manual verification |
| [docs/RELEASE.md](docs/RELEASE.md) | Cross-platform build and packaging plan |
| [docs/DECISIONS.md](docs/DECISIONS.md) | Decisions, assumptions and open questions |
| [docs/SOURCES.md](docs/SOURCES.md) | Official source links and verification date |

The [example question](curriculum/examples/binomial-seven-stage.draft.json) illustrates the contract; it is a **draft**, not approved course content. [schemas/question-template.schema.json](schemas/question-template.schema.json) and [schemas/saved-explanation.schema.json](schemas/saved-explanation.schema.json) are starting schemas that must be reconciled with runtime contracts during Stage 02.

## Scaffold check

With Python 3.9 or newer installed, run `python scripts/validate_scaffold.py` from this folder. This checks local Markdown links, JSON syntax, required documents/directories, and draft fixture basics. It does not replace JSON Schema validation, math tests, or application tests. Python is only a documentation-check convenience, not a planned application runtime dependency.

No build or launch command is presented as working until Stage 01 creates the application. See [CONTRIBUTING.md](CONTRIBUTING.md) for the development workflow.

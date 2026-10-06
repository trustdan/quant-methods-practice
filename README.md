# Quant Methods Practice

A planned local, keyboard-driven probability and statistics tutor with readable mathematics, progressive problems, persistent practice history, optional AI help, saved explanations, and a skippable space arcade.

**Status: Stages 00 through 10 complete; Stage 11 (ChatGPT plan sign-in) implemented with mock protocol tests, live account verification pending, October 5, 2026.** Standalone repository established at `https://github.com/trustdan/quant-methods-practice.git`. Standalone Go executable and React web interface are fully working offline with local MathJax math rendering, 10-question reviewed curriculum bank, deliberate practice scheduling & transfer engine, concept mastery dashboard, offline formula reference library, personal AI notes library with Markdown export, protected credential vault & AI provider adapters (Anthropic/Gemini/OpenAI), responsive layout, and complete keyboard navigation.

## Launching the application

To start the local application and open it in your browser immediately:

- **PowerShell**: `.\scripts\start.ps1`
- **Windows Command Prompt**: `.\scripts\start.bat`
- **Linux / macOS**: `./scripts/start.sh`
- **Or launch the binary directly**: `.\bin\quant-practice.exe`

CLI inspection tools:
- `.\bin\quant-practice.exe -list-bank` — View all approved active curriculum questions.
- `.\bin\quant-practice.exe -mastery` — View current concept retention and transfer status.
- `.\bin\quant-practice.exe -notes` — View personal saved notes library.
- `.\bin\quant-practice.exe -providers` — View configured AI provider and vault credential status.

## Start building

1. Read [OVERVIEW.md](OVERVIEW.md), [REQUIREMENTS.md](REQUIREMENTS.md), and [docs/AGENT-CONTRACT.md](docs/AGENT-CONTRACT.md).
2. Resume from [docs/HANDOFF.md](docs/HANDOFF.md), then perform the first unfinished stage in [PLAN.md](PLAN.md).
3. Inventory actual course materials using [docs/COURSE-MAP.md](docs/COURSE-MAP.md).
4. Run tests with `.\scripts\test.ps1` or `./scripts/test.sh`.
5. Complete the Stage 11 live sign-in check described in [docs/HANDOFF.md](docs/HANDOFF.md), then continue with [PLAN.md](PLAN.md).

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

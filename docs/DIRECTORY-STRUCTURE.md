# Directory structure

This is the standalone repository root after the folder is moved. The source directories below exist as empty .gitkeep placeholders; their packages and modules are planned. Stage 01 creates manifests/configuration only after dependency selection.

```text
quant-methods-practice/
  AGENTS.md                 # Codex entry -> shared contract
  CLAUDE.md                 # Claude entry -> shared contract
  README.md
  OVERVIEW.md
  REQUIREMENTS.md
  PLAN.md
  CONTRIBUTING.md
  .editorconfig
  .gitignore
  docs/                     # Architecture, pedagogy, providers, handoff, etc.
  schemas/                  # Starting strict JSON schemas
  curriculum/
    examples/               # Draft reference fixture; never auto-loaded as active
    approved/               # Reviewed templates and concepts, not populated yet
    reviews/                # Versioned review provenance
  course-materials/
    README.md
    private/                # Ignored original syllabus/slides/homework/datasets
  cmd/quant-practice/        # Local app CLI and launcher
  internal/
    app/ auth/ assets/ bank/ domain/ drill/ exam/ httpapi/
    mastery/ mathengine/ providers/ storage/ tutor/
  web/
    src/
      app/                  # Composition, routes and API client
      components/           # MathMarkdown and shared accessible components
      navigation/           # Keymap, focus, command/leave-intent handling
      features/
        practice/ exams/ mastery/ tutor/ notes/ providers/
        candidates/ reference/ cases/
      visuals/              # Accessible SVG statistics figures
      arcade/               # Simulation, Canvas renderer, input and HUD
    public/                 # Public original assets; no secrets or answer keys
  scripts/
    validate_scaffold.py    # Working doc/JSON checker
  tests/
    fixtures/               # Numerical/reference fixtures and safe test datasets
    e2e/                    # Browser flows with fake providers
  .github/workflows/        # CI added in Stage 01
```

Later root files: go.mod/go.sum and a chosen LICENSE. Later web files: package.json/lockfile, TypeScript/Vite/test configs, HTML entry. Later internal/storage/migrations: numbered SQL with backups/migration tests. Later scripts: build/development/release launchers for PowerShell/POSIX. Generated web/dist, binary outputs and personal application data are ignored.

Default personal data belongs outside the source tree. Safe standalone move means no absolute link back to the accounting folder, no copied learner data, no .git directory, no API keys and no nested dependency on the parent repository.

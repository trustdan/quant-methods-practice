# Implementation bootstrap checklist

Use this after moving the scaffold into its own repo. Do not run a bulk generator that overwrites these documents.

## Stage 00 [COMPLETE]

- Inspect existing files, instructions, Git status and HANDOFF: Complete. Clean git status on `main` branch.
- Inventory Go, Node/package manager, browser and platform versions:
  - Installed Go: `go1.27.1` (windows/amd64).
  - Installed Node: `v22.21.1`.
  - Installed Package Manager: `npm 10.9.4`.
  - Installed Python: `Python 3.13.15` (scaffold validator).
  - Platform / OS: Windows 11 amd64; Microsoft Edge installed.
- Inventory actual course files under course-materials/private: Confirmed empty (`.gitkeep` only). Baseline learner-provided topics recorded in [docs/COURSE-MAP.md](COURSE-MAP.md).
- Git remote and module path:
  - Git remote: `https://github.com/trustdan/quant-methods-practice.git` (origin).
  - Go module path: `github.com/trustdan/quant-methods-practice`.

## Stage 01

- Create root go.mod with the actual module path and verified supported Go version; pin SQLite driver and any numerical/auth dependencies selected for the first slice.
- Create web/package.json and one lockfile, TypeScript strict configuration, Vite config, React app entry, unit setup and browser test configuration. Use one package manager throughout.
- Establish scripts: dev, typecheck, lint, test, test:e2e and build. Implement reproducible Go and frontend build scripts for PowerShell and POSIX shells.
- Create a minimal Go server and embedded production asset boundary; development proxy must preserve the local-session/origin model in SECURITY.
- Pin/bundle MathJax JS, extensions, fonts and sanitizer dependencies with licenses. Verify the dynamic math lifecycle rather than loading a CDN script.
- Add one static ungraded math demonstration and accessible focus/navigation shell. This demonstration must not masquerade as an approved drill.
- Establish CI for formatting, math/domain tests when present, Go vet, frontend checks/build, local HTTP tests and documentation check. No live keys in CI.
- Build the production package and load it with external network disabled. Confirm no missing dynamic fonts/extensions or failed remote requests.

## Stage 02 onward

Reconcile starting JSON schemas with actual Go types and strict validators. Do not commit an active seed bank until mathematical and semantic content review exists. Implement binomial rule and one seven-stage drill before broadening the curriculum. Persist immutable history before adaptive mastery.

## Planned commands

The final executable should support `quant-practice --skip-intro`, `--data-dir`, `--db`, `--seed`, `--size`, `--intensity`, `--version`, `--help`, and local administrative commands. These are target contracts, **not working commands yet**. Startup defaults to opening the browser; add `--no-browser` and `--port` for troubleshooting.

Expected development commands after Stage 01 are Go build/test/vet and the npm scripts chosen above. Until then only the scaffold validator runs. Do not create fake successful launch scripts to satisfy the directory tree.

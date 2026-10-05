# Handoff

## Current state - October 4, 2026

Stage 00 is complete and Stage 01 is implemented and verified. The standalone repository is active on branch `main` at `https://github.com/trustdan/quant-methods-practice.git`.

### Stage 00 Record
- Toolchain versions: Go `go1.27.1` (windows/amd64), Node `v22.21.1`, npm `10.9.4`, Python `3.13.15`, Windows 11 amd64 (Microsoft Edge available).
- Module path: `github.com/trustdan/quant-methods-practice`.
- Course material inventory: Confirmed `course-materials/private` is empty (`.gitkeep` only). Baseline learner-reported topics remain recorded in [docs/COURSE-MAP.md](COURSE-MAP.md).
- Stale move/setup instructions updated across `README.md`, `PLAN.md`, `docs/BOOTSTRAP.md`, `docs/COURSE-MAP.md`, and `docs/DECISIONS.md`.

### Stage 01 Implementation
- Go backend: Loopback HTTP server in [internal/httpapi](file:///internal/httpapi/server.go) and CLI entry point in [cmd/quant-practice](file:///cmd/quant-practice/main.go) with embedded production frontend assets in [internal/assets](file:///internal/assets/assets.go).
- Security boundaries: Loopback-only binding, Host header validation (DNS rebinding prevention), Origin validation on mutating requests, HttpOnly SameSite session cookies, single-use bootstrap tokens via URL fragment `#bootstrap=<token>`, strict Content-Security-Policy headers, and SPA route fallback.
- Web frontend: React 18, TypeScript strict, Vite, Vanilla CSS design system, [MathMarkdown](file:///web/src/components/MathMarkdown.tsx) component with DOMPurify sanitization and MathJax typeset lifecycle, [StageStrip](file:///web/src/components/StageStrip.tsx) (7 stages), [QuestionStrip](file:///web/src/components/QuestionStrip.tsx), central keyboard command resolver [keymap.ts](file:///web/src/navigation/keymap.ts), and static ungraded math demonstration shell.
- MathJax bundling: MathJax v3 TeX-SVG bundled completely locally in `web/public/vendor/mathjax/` via `web/scripts/copy-mathjax.js` with deterministic prebuild hook. Zero external CDN or network requests required.
- Build & test pipeline: `scripts/build.ps1`, `scripts/build.sh`, `scripts/test.ps1`, `scripts/test.sh`, and GitHub Actions CI in `.github/workflows/ci.yml`.

## Checks

- `python scripts/validate_scaffold.py`: Passed (31 Markdown files, 50 local links, 20 JSON files, 35 source directories).
- `go test -v ./internal/... ./cmd/...`: Passed (5 unit tests covering health status, single-use bootstrap token exchange, host header rejection, cross-origin mutation rejection, and SPA fallback).
- `go vet ./internal/... ./cmd/...`: Passed (0 warnings).
- `gofmt -s -l cmd internal`: Passed (clean formatting).
- `npm run typecheck` (in `web/`): Passed (strict TypeScript, 0 errors).
- `npm run test` (in `web/`): Passed (17 unit tests covering keymap resolution, editable field exclusions, MathMarkdown sanitization, StageStrip, QuestionStrip, and App shell navigation).
- `npm run build` (in `web/`): Passed (outputting compiled HTML/JS/CSS to `internal/assets/dist`).
- `go build -o bin/quant-practice.exe ./cmd/quant-practice`: Passed (standalone binary built).
- Local loopback server smoke: Launched `.\bin\quant-practice.exe --port 8976 --no-browser`. Verified `GET /api/health` returned HTTP 200 with status ok and version `0.1.0-dev`. Verified `GET /vendor/mathjax/tex-svg.js` returned HTTP 200 with 2,108,580 bytes of local JavaScript.
- Automated browser subagent: Playwright browser environment reported driver download failure (404 from external Playwright CDN); loopback HTTP verified.

## Next action

Proceed to Stage 02: Domain and bank contracts (stable IDs, parameters/constraints, answer/assistance types, versions, schemas, family registry, strict parsing, and reconciling the draft binomial fixture with runtime Go types).

## Unresolved external gates

Actual course syllabus/slides/notation; ChatGPT-plan sign-in account access; native vault integration tests; optional Google project OAuth; continuous distribution/test/regression course details.

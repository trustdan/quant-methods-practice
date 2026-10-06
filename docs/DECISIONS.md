# Decisions and assumptions

Recorded October 4, 2026; revisit with explicit reasons rather than quietly changing architecture.

| Decision | Reason / review trigger |
|---|---|
| Local Go backend with browser UI | Preserve Go/SQLite and support equations/figures; hosted service is a separate design |
| React + TypeScript + Vite | Structured view state, typed commands and keyboard/focus testing; pin current supported versions at bootstrap |
| Bundled MathJax | Familiar math-note format and local rendering; verify font/extension packaging |
| Markdown + math for prose, JSON for content contracts | Human-editable explanations with strict machine metadata/parameters |
| SVG statistics diagrams, Canvas 2D arcade | Accessible math visuals and efficient original game rendering without a 3D dependency |
| Binomial first slice, then reported covered topics | Enough for concrete seven-stage drill; generic assumptions labeled pending course evidence |
| Stable numerical grading in Go | AI/DOM cannot be mathematical truth; arbitrary symbolic grading deferred |
| Backend vault preferred, session-only fallback | Credentials must not live in browser/learning DB; platform support needs native testing |
| ChatGPT plan and API routes distinct | Different eligibility, payload constraints and billing; no automatic switch |
| Anthropic key only; Google project OAuth optional | Reviewed official guidance does not establish equivalent consumer-login routes |
| No manifests/source application in scaffold | User asked for a move-ready plan before starting implementation in new repo |
| Full parity staged after offline drill | All desired workflows tracked without implementing everything simultaneously |
| Module path `github.com/trustdan/quant-methods-practice` | Established at Stage 00; remote `https://github.com/trustdan/quant-methods-practice.git` on branch `main` |
| Toolchain baseline | Go `go1.27.1` (windows/amd64), Node `v22.21.1`, npm `10.9.4`, Python `3.13.15` |
| Excel formula equivalents | Connect analytical derivations with standard business stats spreadsheet functions (=BINOM.DIST, =POISSON.DIST, =NORM.DIST, =COMBIN) |
| Multi-turn AI tutor chat | Stateful conversational threads with SQLite persistence and strict answer-withholding guardrails on active stages |
| ChatGPT plan callback on its own one-shot loopback listener | Keeps OAuth callback acceptance separate from the authenticated app API (SECURITY); the listener exists only during an attempt |
| ChatGPT plan tokens in the existing encrypted vault | Reuses the DPAPI/AES vault rather than adding a second token file; the UI only sees redacted account summaries |
| Local LLM (LM Studio / Ollama) | OpenAI-compatible loopback adapter for zero-network, cost-free, private offline generative tutoring |

## Open facts

Actual syllabus/slides, professor notation and exam rules; complete later distribution/test/regression list; datasets and use rights; project license; target browser support; OS vault library and native build implications; last-tab shutdown policy; game balance/playtesting; optional symbolic grammar; optional desktop wrapper. No open fact prevents building the local application shell or standard offline binomial slice.

No empirical effectiveness/retention claims are established. No provider auth/inference or native package smoke has run for this new project.

# Build and distribution plan

No release exists. Select toolchain/dependency versions during Stage 01 and license before redistribution. Packaged app is a Go binary with embedded compiled frontend, local MathJax/fonts, approved bank/reference text and original assets. Existing browser required; Go/Node/SQLite installer not required for end users.

## Pipeline

1. Clean reproducible frontend dependency install from lockfile; type/lint/unit checks and production build.
2. Copy/build assets into the explicit Go embed input boundary; fail if absent or stale. Record generated asset hash/build version.
3. Format/test/vet Go and run integration/browser tests appropriate to changes.
4. Build Windows amd64, macOS arm64/amd64 and Linux amd64/arm64 where dependencies allow; document SQLite/credential-vault native dependencies honestly. A pure-Go SQLite driver is preferred for portability.
5. Package binary, platform launcher where useful, quickstart, licenses/notices and checksums. No private course sources, keys, history or notes.
6. Verify every archive entry against sources/checksums and smoke the extracted package natively on the claimed platform.

The executable launches a loopback service, opens the default browser and displays address/status in its terminal. Provide --no-browser for manual opening and clear shutdown behavior. In-app quit can end the local process only after saving/finishing pending work; closing a browser tab must not silently terminate another active tab/session. Record the eventual last-tab/shutdown policy in DECISIONS.

## CLI target surface

--help/--version; --skip-intro; --port/--no-browser; --data-dir/--db; --seed/--size/--intensity/--module; --mastery/--export-json; --validate-bank/--reconcile-all/--active-bank/--export-bank; candidate review commands; --exam/--exam-time/--resume-exam/--exam-history/--exam-report; --high-scores; --list-models/--fetch-models/--test-llm; explicit --tutor/--tutor-model. These are planned flags, not presently supported commands.

Diagnostics are redacted and distinguish configuration, auth, model discovery and actual inference. Bank validation/reconciliation checks all allowed combinations or a justified deterministic coverage scheme, not one conveniently chosen fixture.

## Release claims

Cross-compiled binaries do not establish native runtime, browser interaction, vault availability or live provider compatibility. Document unsigned/notarization status accurately; do not automatically bypass operating-system security. Desktop wrapper, signing and hosted publication need their own future decisions.

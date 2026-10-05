# Contributing

Follow the shared agent contract and current handoff. Work on one coherent stage increment at a time. Keep runtime implementation separate from imported course sources and personal data. Use strict, versioned content and review records.

Before a change, inspect Git status and relevant documents. Before finishing, run the checks appropriate to that stage, inspect the diff, and update HANDOFF. Once Stage 01 establishes toolchains, use `gofmt`, `go test ./...`, `go vet ./...`, frontend type/lint/unit/build scripts, and focused browser tests; these commands are not currently runnable against an app.

For content changes, record sources, conditions, expected canonical outputs, all stage wording/options/hints, parameter bounds and semantic review. A count increase is not a quality gate. Do not commit personal databases, note exports, credentials or course files whose sharing is not authorized.

PR descriptions should state the concrete change, learner-visible behavior and checks actually run. Explicitly distinguish mocks, manual browser work, native package checks and live provider calls. Choose a project license before redistribution; no license is selected in this scaffold.

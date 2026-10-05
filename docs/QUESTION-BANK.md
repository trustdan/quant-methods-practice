# Question bank and review contract

Versioned approved content is separate from candidate and draft content. The sample file under curriculum/examples is never auto-loaded into active practice. Schemas here are starting contracts, not implemented validators. Runtime strict JSON parsing and semantic checks are required at Stage 02.

## Template versus instance

A template contains stable ID/version, supported family/rule version, status, module/concepts, source/provenance, reviewed setting group, parameter choices/constraints, scenario wording, stage metadata, diagnostic distractors/hints and grading policy. A family owns executable mathematical derivation and semantic parameter constraints. An instance stores the selected parameters/seed, original stage prompts/options/order and canonical derived values.

The initial example is a fixed-parameter reference fixture to demonstrate seven-stage content. Generalized templates must generate choices/keys from family rules rather than retaining a fixed .375 expected answer while changing n/p/k. Extend the schema with bounded parameters/expressions only after a reviewed runtime design.

## Admission pipeline

AI/local proposal -> strict parsing -> supported family/parameter validation -> canonical computation -> numerical bounds/invariants -> rendered preview of every stage/amount/variant -> explicit semantic review -> approval event -> active versioned content.

Validation cannot prove unrestricted natural language describes the claimed mathematical model. A story can yield a valid binomial calculation yet fail independence or constant-p conditions. Human wording review remains required unless explicitly delegated by the user and recorded honestly. Model self-review is supporting evidence, not automatic admission.

New families need sources or labeled generic scope, parameterization, rules, tests, grade policies and reviewed teaching. Candidate-provided formula or numeric answer is a proposal/fixture, never an independent grading authority. Reject unknown fields, nonfinite/invalid probabilities, unsupported distributions, missing assumptions, impossible event bounds, duplicate IDs/options and ambiguous units.

## Review record

Record candidate/template version, complete scenario and stage wording, correct results and how derived, distractors/hints, source locations, allowed parameter combinations, contrast/group bindings, math/rendering/restart checks, reviewer/timestamp, delegation basis if any and limitations. An assistant review must not be described as the user's personal reading.

Approved version is immutable. Editing wording creates a new version and fresh approval. Retirement excludes future selection but preserves snapshots. Corrections to historical evidence use explicit correction events/policy, never silent answer-key replacement. Export approved content separately from candidates and personal history.

## CLI/UI parity

Plan commands for --validate-bank, --reconcile-all, --active-bank, --export-bank, --generate-candidate, --candidates, --preview-candidate, --approve-candidate, --reject-candidate and --retire-question. Review UI offers the same actions with visible provenance. No bulk approve by learner performance or model confidence.

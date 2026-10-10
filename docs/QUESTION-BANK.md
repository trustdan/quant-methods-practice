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

## Stage 12 implementation (binomial candidates)

Open **Question candidates** (keyboard `p`). Generate a local variation, request AI wording using the explicitly selected provider/model, or import a small proposal JSON object. The first candidate builder supports `binomial_pmf`, rule version 1, exactly-k events, and four tested `(n,p,k)` combinations: `(4,0.5,2)`, `(8,0.25,2)`, `(6,0.5,3)`, `(10,0.2,2)`. Other families and parameter combinations fail closed until their candidate builders and teaching rules are reviewed. Existing reviewed questions in other families remain available for practice.

[Proposal schema](../schemas/candidate-proposal.schema.json) describes the seven accepted fields. Proposals cannot supply answer keys, options, formulas, grading policies, approval records, or setting groups. Go builds all seven stages, keys, hints, tolerances and explanations. AI only rewrites the title/scenario/success label for fixed server-selected parameters. Provider output is bounded to 12 KB, requires a completed stream, times out after 60 seconds, and cancels on navigation. Failure never activates a fallback candidate. No AI request occurs merely by opening the review screen.

Candidates live in `candidate_questions`, outside the active bank. The preview shows the complete scenario, assumptions, parameters, all stages, correct options/numeric result, distractors, hints, explanations and provenance. Structural/numeric validity does not prove the story is valid. Approval requires an explicit semantic-review confirmation, reviewer name and review findings. Rejection and retirement require a reviewer and reason. A preview revision prevents stale/duplicate reviews; each successful transition appends `content_approval_events` in the same SQLite transaction.

Approved candidates join future practice selection immediately and reload at startup. Retirement removes locally approved candidates from future selection and keeps their audit history and existing immutable session snapshots. Review does not write attempts or mastery. All four variations retain `binomial_fixed_independent_exact_count`, so new wording/numbers do not establish transfer. No generated content was activated in the developer's personal database during automated verification; approvals in tests use isolated temporary databases and fixture reviewer names.

Candidate content is immutable after creation. To change wording, copy its proposal to the editor and save a separate draft with a fresh content ID/version 1 and fresh approval. In-place editing, reactivating rejected/retired records, arbitrary full-template imports and new-family generation are not offered. The first retirement UI operates on locally approved candidates; bundled curriculum remains source-controlled. Export downloads the entire active approved bank, excluding pending/rejected/retired proposals and learner history.

CLI equivalents (use `bin/quant-practice` on macOS/Linux, `.exe` on Windows):

```sh
bin/quant-practice -generate-candidate -seed 1
bin/quant-practice -candidates
bin/quant-practice -preview-candidate CANDIDATE_ID
bin/quant-practice -import-candidate proposal.json
bin/quant-practice -approve-candidate CANDIDATE_ID -candidate-revision 1 -candidate-reviewer "Your name" -candidate-notes "Review findings" -confirm-semantic-review
bin/quant-practice -reject-candidate CANDIDATE_ID -candidate-revision 1 -candidate-reviewer "Your name" -candidate-notes "Rejection reason"
bin/quant-practice -retire-question CANDIDATE_ID -candidate-revision 2 -candidate-reviewer "Your name" -candidate-notes "Retirement reason"
bin/quant-practice -export-bank approved-bank.json
```

CLI generation is offline; AI generation is an explicit UI action. CLI commands use the same database and review rules; restart an already running app after CLI bank changes to reload its in-memory selection bank. Exports create a new file and refuse to overwrite an existing path. `-data-dir` isolates test or alternate libraries. No live provider request is part of automated verification.

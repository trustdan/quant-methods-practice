# Full solutions and dataset cases

Stage 13's initial scope is structured full-form practice over the approved bank and one generic CSV case. No course files are available; this does not claim course-specific alignment or arbitrary statistical-method coverage.

## Full solutions

Open **Worksheets**, or press **Shift+J** outside an editable field. Choose an approved question and create a worksheet. Its existing method/model, assumptions/parameters, expression, calculation and interpretation stages appear together. These are structured choices and numeric entries, not automatically graded prose.

Submit every unfinished field together. Go uses the same `drill.GradeSubmission` and numeric policy as progressive practice. A malformed field rejects the whole submission without consuming attempts. Correct fields lock; each incorrect field gets one causal hint and one retry, then a solution explanation. Decimals, percentages and fractions follow the existing parser and per-question tolerances.

**Save worksheet draft** persists raw ungraded answers. App navigation retains unsaved input; browser closure warns about it. Creating/selecting another worksheet is disabled until edits are saved or discarded. **Discard unsaved edits** and **Reload saved version** deliberately replace local input. Reloading the app selects the most recently updated worksheet when this tab opens. Completed worksheets are read-only.

**Download worksheet** exports prompts, choices, assumptions, template versions/seed, numeric policies and earned feedback/attempts. It uses saved state. Hidden unresolved keys are omitted and HTML is escaped.

## CSV case: four-toss experiment counts

Press **Shift+F**, or select **CSV dataset case**. Paste data or choose a UTF-8 file with exactly these columns:

```csv
experiment_id,successes
run_1,2
run_2,1
run_3,4
run_4,2
run_5,0
```

Each row represents one complete four-toss experiment; `successes` counts heads, from 0 to 4. IDs must be unique, 1–80 ASCII letters/digits/underscores/hyphens, starting with a letter/digit/underscore. Limits: 64 KiB, 500 experiments, two columns. Empty datasets, duplicate IDs, extra columns, out-of-support counts, formula-like values, paths, control characters and invalid UTF-8 fail closed. Cells are never evaluated; CSV content and file names are never opened as paths.

**Preview CSV** presents validated rows and the complete proposed two-part case, including every field, option, hint, answer and numeric policy for author review. It creates no worksheet or bank entry. Enter your name and source/row note, review the entire case and explicitly confirm before creating it. This is human semantic approval, not automatic validation. Editing CSV or review metadata clears confirmation. Approval provenance and immutable data are saved with the case.

The first part checks the method, assumptions, observed fraction and interpretation. Go counts experiments with exactly two heads and divides by all recorded experiments. The example gives `2/5 = 0.4`. The second part uses the approved fair-independent-coin reference: `3/8 = 0.375`. The observational unit is an experiment, not a toss. Recorded frequency can differ from the hypothetical model probability; this alone cannot establish fairness or independence. No hypothesis test, confidence interval or fitted model is implied.

Empirical rule v1 is restricted to four-toss counts and target two. Numeric policy v1 accepts decimal/percent/fraction with absolute and relative tolerances `1e-6`, without mandatory rounding. Independent hand-derived fixtures cover zero/all matches, `1/3`, `2/5`, invalid support and empty data. The frequency definition follows [OpenStax, frequency and relative frequency](https://openstax.org/books/introductory-statistics-2e/pages/1-3-frequency-frequency-tables-and-levels-of-measurement). The model follows [NIST's binomial distribution](https://www.itl.nist.gov/div898/handbook/eda/section3/eda366i.htm), with independence described in [NIST's binomial definition](https://csrc.nist.gov/glossary/term/Binomial_Distribution). These references do not establish syllabus alignment.

## Evidence and persistence

Attempts record `full_solution_form`; retries also record `hint` and `retry`. Dataset cases also record `reference`, since author review exposes keys. Reveals are recorded on stage state. Worksheet results remain separate from independent mastery and transfer. This workflow makes no provider calls.

Snapshots preserve the approved source template, rule/approval provenance, parameters, seed, option order, stage prompts/keys and numeric policies. Dataset rows and case approval are immutable. Retirement or later bank edits cannot replace saved content. Commands require exact revisions and IDs; state, attempts and command result commit in one transaction. Identical replay returns the original durable outcome after restart; changed content under the same ID conflicts. After a network failure the UI retains the exact request for explicit retry.

Migration `002_worksheets.sql` adds separate worksheet and command tables. Existing databases get a consistent pre-migration backup through the launcher. No prior migration is edited and no drill/mastery table is written by worksheet commands.

Arbitrary CSV schemas, free-prose grading, course-specific cases, worksheet-to-mastery policy, CSV export and worksheet CLI commands are outside this initial scope. More cases require reviewed family rules, teaching content and fixtures.

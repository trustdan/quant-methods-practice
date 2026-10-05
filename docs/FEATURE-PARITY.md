# Accounting-app feature parity

Reference: the inspected AccountTutor README, architecture, pedagogy, mastery and question-bank documents, working-tree version 0.24.0. This is a feature map, not a claim that accounting code is verified for statistics or that any new functionality is implemented.

| Accounting workflow | Quant Methods equivalent | Requirements / stages |
|---|---|---|
| Local offline Go executable, SQLite | Local Go server with embedded browser UI, SQLite | R01/R24; 01/05/07/19 |
| Seven account-to-entry decisions | Seven model-to-interpretation decisions per scenario | R06; 04 |
| Four/two-step faded drills | Family-specific reduced scaffolds | R06/R11; 08 |
| Hint, one retry, causal explanation | Misconception-specific mathematical feedback | R07; 04/08 |
| Reviewed matched contrasts | At least/exactly; independence/disjointness; SD/variance | R07/R11; 08 |
| Vim/arrow keys, scrolling and question shortcuts | Focus-aware browser keymap and persistent stage/problem strips | R08/R09; 06 |
| Session length/intensity, seeded generation | Module/concept filters, standard/spaced/intensive/transfer | R05/R10; 07/08 |
| Durable immutable attempt history | Original numerical/semantic snapshots, stable order/policies | R12; 05 |
| Mastery, decay and delayed cross-setting success | Versioned concept evidence and reviewed transfer groups | R11; 08 |
| Offline tutor, cancellable model requests | Same read-only tutor contract with math Markdown | R13; 09 |
| ChatGPT account, Anthropic/Gemini/OpenAI keys | Verified ChatGPT-plan flow and separate API adapters | R14; 10/11 |
| Live model picker/refresh/cache | Provider/capability-aware dynamic discovery | R15; 10 |
| Save-on-leave explanations and V library | Saved math notes, search, follow-ups, Markdown disk export | R16/R17; 09 |
| AI/new local question generation | Supported-family candidate generation/variation | R18; 12 |
| Holding area, preview/approve/reject/retire | Semantic and numerical review before activation | R18; 12 |
| Full journal entry input | Complete structured probability/statistical solution | R20; 13 |
| T-accounts/equation reconciliation | Event tables, PMF/CDF checks, expectation/variance decomposition | R03/R21; 03/15 |
| Financial statements / accounting-cycle case | Multi-part dataset/case worksheets and statistical reports | R20/R21; 13/16 |
| Timed/untimed exams, resume and history | Snapshot-based quant exams, withheld feedback and reports | R19; 14 |
| Help/reference grid and causal offline explanations | Notation glossary, reviewed formula/conditions reference | R22; 04/07 |
| CLI diagnostics, reconciliation, exports | Bank/engine validation, session/exam/provider/admin commands | R23; 07/10/12/14 |
| Backups, migrations and history export | SQLite backup/recovery and safe note/bank/history exports | R24; 05/09/19 |
| Native binaries/launchers/ZIP/checksums | Go binaries plus embedded UI; cross-platform browser launch | R24; 19 |
| Autopilot startup/steering/fire/bombs | Quant-themed Canvas flight with immediate skip | R25; 17/18 |
| Acceleration, dual health and heavy hazards | Escalating sectors, shield/integrity meters, fragments | R25; 18 |
| Four-quarter win, continuation and retry | Four-sector completion, endless continuation and retry | R25; 18 |
| Three initials, persistent scores, leaderboard | Separate human-only local arcade records | R26; 18 |
| Title transition / return to practice | Original Quant Methods title reveal and direct skip | R26/R27; 17/18 |

## Intentional changes

Accounting account names, debit/credit, money-only integer representation and journal schemas are not portable mathematical rules. Replace them with family-specific parameterizations, event bounds, numeric grading and units. Improve credential storage to an explicit backend vault; do not inherit a plaintext-storage claim as encrypted storage. A browser UI needs focus-aware shortcuts, sanitized generated Markdown and local HTTP protection that a terminal app did not require.

Do not inflate the feature list by duplicating screens with accounting labels. Every full-solution/worksheet feature must have a useful statistics task and shared canonical math.

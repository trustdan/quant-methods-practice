# Product overview

Build the deliberate-practice experience of AccountTutor for probability and statistics, with equations and figures suited to the subject. The learner wants local progress, keyboard fluency, manageable steps, causal hints, AI connections, saved explanations that work in Typora/Obsidian, and an enjoyable startup game.

The learner reports covering union, intersection, binomial, Poisson, and sums of random variables. Supplied module headings cover descriptive measures/probability; discrete random variables; continuous distributions; sampling distributions/confidence intervals; hypothesis testing; regression. Actual syllabus, slides, homework conventions, datasets, assessment rules, and calculator policy are unavailable. Do not infer exact coverage from headings.

## Experience

Launch the local executable, optionally play or skip the arcade, and enter practice. Pick a course module and short session. One scenario produces up to seven related decisions: identify the target, define the model, check conditions, choose the event, construct the expression, calculate, interpret. A visible problem/stage strip permits revisiting prior work and changing questions without losing context or history. Hints and retry evidence stay distinct from independent answers.

Start with full support, fade it when independent transfer supports that decision, and restore it on error. Immediate feedback, question recap, worked solution, reference library, mixed practice, exams, progress, and learner-saved AI notes are separate views. The math should resemble course notes. No learner must type LaTeX to answer an early drill.

## Technical defaults

Keep Go and SQLite, replace the terminal interface with a local browser interface, and separate mathematical domain rules from display, storage, providers, and arcade code. React/TypeScript manages view state and focus; Go is authoritative for sessions and grades. Bundle Markdown/math/font assets and the curriculum. Browser development uses Node; end users launch one packaged Go executable and their installed browser.

The first useful release is a restart-safe offline binomial drill with readable formulas and working keyboard navigation. Expand to the learner's other covered topics next. Provider connections and the game cannot be prerequisites for useful practice, but remain explicit deliverables in the full roadmap.

## Carry-over principles

AI creates candidate content and explains reviewed problems. Rules derive the answer. Review controls the bank. Immutable history drives scheduling. Rendering never grades; an AI conversation never awards mastery; arcade scores never affect learning evidence.

Avoid copying accounting-specific account semantics, monetary representation, migrations, or question banks into this repo. Reuse ideas and carefully reviewed generic code only. Full parity means analogous learning workflows, not journal entries hidden inside a statistics app.

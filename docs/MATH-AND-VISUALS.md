# Math rendering and statistics figures

## Content format

Store prose as UTF-8 Markdown and formulas as LaTeX-style math. Configure $...$ inline and $$...$$ display delimiters consistently in the app; MathJax defaults differ, so configuration is part of the contract. Restrict to a shared supported notation subset that renders in the app, Typora and Obsidian. Export raw text rather than screenshots/HTML as the primary note format.

Choose locally hosted MathJax and bundle all required runtime components, extensions and fonts. MathJax documents local hosting and browser rendering; local assets are required for the offline guarantee. Sources: [getting started](https://docs.mathjax.org/en/latest/web/start.html), [hosting locally](https://docs.mathjax.org/en/latest/web/hosting.html).

## Component lifecycle

MathMarkdown parses Markdown, sanitizes permitted markup and then typesets math in its owned container. Never execute provider HTML, scripts, links with unsafe schemes, arbitrary macros or remote embeds. Serialize/cancel stale typesetting and clear registered math when React removes a container. Debounce streaming text; an unfinished math delimiter must not crash the panel. Final completion typesets the complete text again safely.

Test union/intersection/complement, Greek symbols, subscripts/superscripts, stacked fractions, binomial coefficients, sums, integrals, cases, aligned equations and matrices. Check source Markdown export, dynamic replacement, fonts, zoom, narrow screens, large equations, accessibility and missing/malformed math. A rendering error is not a wrong learner answer.

Early answers use choices, numeric inputs and structured fields. Later symbolic entry is a separate restricted parser/grammar feature with equivalence tests; do not trust typeset LaTeX as executable math. Never evaluate generated JavaScript or arbitrary expressions.

## Visual progression

| Feature | Behavior / canonical relationship |
|---|---|
| Venn diagram | Select/shade A, B, union, intersection and complement; text names match event semantics |
| Discrete PMF/CDF | Bars and cumulative values; exact/at-most/at-least event highlights share engine bounds |
| Continuous curve | PDF versus CDF clearly labeled; highlighted area is probability, curve height is density |
| Sampling simulation | Seed, sample size and replicate count visible; theoretical target separate from observed estimate |
| Confidence coverage | Repeated intervals and fixed parameter; explain frequentist coverage without changing interpretation |
| Hypothesis diagram | Null distribution, tail selection, statistic, alpha and p-value are distinct layers |
| Regression | Scatter/fit/residuals, units, model assumptions and appropriate caveats about causal conclusions |

Use SVG for core diagrams and accessible controls; add a plotting dependency only when needed and after official documentation/license review. Every diagram has labels, textual/table equivalent and keyboard-operable controls. Color is not the only event distinction. Manipulating a visualization/reference during a live unresolved problem marks appropriate assistance. Exam permissions are explicit.

Math rules live in Go; UI transformations share returned plot data/parameters and never create a competing grader. For teaching simulations, expose seed/assumptions and test expected aggregate behavior without asserting each sample matches theory.

# AI tutor and saved explanations

## Read-only service

Conceptual Go interface: Hint(ctx, request), Explain(ctx, request), FollowUp(ctx, request), Chat(ctx, chatRequest), GenerateCandidate(ctx, generationRequest), plus provider/model capability methods. Stream prose as events where supported. OfflineTutor returns reviewed teaching and causal hints from immutable snapshots. When an offline local LLM (LM Studio or Ollama on loopback) is active, generative tutoring and chat run entirely offline without external network calls.

A request contains the original problem/stage, reviewed rule and relevant explanation in explanation mode, submitted response, units/assumptions, assistance state and only required prior conversation turns. Hints must not be given the complete worked explanation prematurely. Backend constructs this context; it does not trust browser-supplied keys. No personal history is sent by default. The provider has no execution, bank approval, grade or mastery API.

System/developer instructions ask for concise Markdown, LaTeX delimited by $/$$, standard Excel formula equivalents (e.g. `=BINOM.DIST(...)`, `=POISSON.DIST(...)`) when deriving calculations, defined Greek symbols, explicit assumptions/units and a short causal explanation. MathJax support is a subset of full LaTeX; do not request document classes or arbitrary macros/packages. A tutor can be wrong even when rendering succeeds. Show advisory status and provide report/correction workflow.

## Requests and fallback

Only explicit user actions call providers. Defaults: 60-second request deadline, 20 external requests per practice session, bounded prompt/output sizes and sliding-window conversation history; actual byte/token caps are chosen/tested at implementation. Streaming progress stays responsive. Esc/navigation cancels; old completions cannot attach to a new stage. Failure offers reviewed offline help and records the source accurately.

Provider-specific parameters differ; avoid applying unsupported token/temperature fields to the ChatGPT-plan route. Response-size caps can stop local consumption/cancel the request without implying the provider supports every parameter. No automatic billed-route fallback or hidden model substitution.

Follow-ups can ask explain differently, show a worked example, why a condition matters, compare concepts, or engage in multi-turn back-and-forth dialogue. In conversational mode, ongoing threads persist in SQLite (`tutor_threads`/`tutor_messages`). Strict pedagogical guardrails prevent the AI from revealing answer keys to unresolved active stages during multi-turn chats. Asking about the current unresolved problem marks assistance; chatting about a concept does not create graded evidence. Viewing a saved solution relevant to an active problem records exposure before any later answer.

## Save workflow

An AI explanation starts as an unsaved recovery draft. Leaving through in-app navigation prompts y save / n discard / Esc stay. Scrolling never prompts. After successful idempotent save, resume the exact intended action. On failure keep readable Markdown and show retry. Browser crashes/close use recovery drafts because custom leave dialogs are not guaranteed.

In multi-turn chat sessions, learners can save the complete formatted transcript or select individual high-value responses to pin as personal notes. Save raw Markdown/LaTeX, title/topic/concepts, question-instance snapshot reference, originating stage and response context, provider/model/billing route, created/saved timestamps and optional parent note/thread ID. Save the explanation actually shown, including a clear fallback label where relevant. Offline hints need not trigger the AI-save prompt, but an explicit Save action may save reviewed explanations too with source=offline.

## Library and disk output

V opens the local library. Search by title/text/topic/concept; use j/k to select and n/p or buttons to browse notes. Reopen offline, render math, navigate original question and ask a follow-up with explicit provider selection. Export individual/batch notes as UTF-8 .md files compatible with Typora/Obsidian. Include advisory status and safe context; no keys, tokens or learner history unrelated to the note.

Note editing/version history may be added after the base save/search/export feature. Deletion is explicit and does not remove practice evidence. Saved notes never enter the bank automatically. AI-generated questions have their own review workflow and provenance, not the note-save dialog.

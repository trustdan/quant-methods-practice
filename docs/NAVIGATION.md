# Navigation and focus contract

Keyboard and pointer controls dispatch the same commands through a central resolver. Do not distribute global document key handlers among screens. Modes: practice choice, practice numeric/text editing, feedback, reading, tutor, candidate review, exam, modal and arcade. Every screen shows available commands.

## Default bindings

| Key | Context / command |
|---|---|
| j/k or Up/Down | Move choice/focused list item; scroll reading views |
| h/l or Left/Right | Previous/next eligible stage when no editable field is active; at problem boundaries show/perform documented adjacent-problem action |
| Ctrl+Left/Right | Previous/next big problem outside text fields, matching accounting behavior; visible buttons are always available |
| Shift+H / Shift+L | Alternative previous/next big problem outside fields, avoiding reserved browser shortcuts |
| 1-4 | Select a visible choice; never insert an answer for a numeric field |
| a/b/c/d | Optional choice aliases only in choice mode; d is scroll only in reading mode |
| Enter | Submit current answer, activate selected item, or continue feedback |
| Space | Continue feedback; normal space in editable fields |
| u/d, PageUp/PageDown | Half-page scroll in reading views; native text behavior in fields |
| gg / G | Top/bottom in reading views; finite sequence timeout and no use in text fields |
| ? / e | Hint / explanation for an eligible active stage |
| h / F1 | Help outside h/l stage mode; use visible Help and F1 in practice to avoid h collision |
| s | Mastery dashboard |
| V | Saved explanations library |
| t | Tutor/provider settings |
| p / n | Candidate review / explicit AI candidate request in practice; p means previous note inside note library only |
| J / F | Complete solution / dataset case workspace (statistics counterparts to journal/statements) |
| E | Exam selection/history |
| [ / ] | Reduce/increase future session size when allowed; never change a running exam's size |
| i | Cycle future practice intensity |
| A / L | Arcade / leaderboard outside text entry |
| Esc | Cancel streaming first; otherwise close top modal/panel or request departure through leave-intent handling |
| q | Exit/leave application through save workflow; browser close is a different event |
| Ctrl/Cmd +/- | Native browser zoom, never intercepted |

Resolve the h conflict explicitly: h/l are stage navigation in practice; F1 and a button open help. Do not claim identical key behavior across every screen. Do not intercept Ctrl+C copy, Ctrl+L address bar, Ctrl+W close tab, or browser navigation shortcuts. Users may remap nonreserved commands in a later settings pass.

## Editable fields and accessibility

If focus is in input, textarea, select, contenteditable, math input, API-key entry or initials entry, printable keys and editing combinations retain their normal meaning. Numeric Enter may submit; Esc may leave edit mode subject to unsaved state. Ignore composition events and handle IME safely. Native Tab/Shift+Tab and selection shortcuts work. Modal focus is trapped appropriately and restored on exit; shortcuts never require a mouse.

Visible stage strip shows 1/7 through 7/7, current/problem status and completed/revealed markers. A question strip/jump list shows current index, pending/skipped/completed status and permitted review. Buttons duplicate stage and big-problem commands. Browser resizing changes layout, not problem identity or input eligibility.

## State transitions

Navigation cannot submit implicitly or alter grades. Persist draft answers by session/instance/stage; resume them without treating a draft as an attempt. Command ID/revision prevents repeated Enter or double click from generating multiple attempts. An answered stage opens its historical feedback, not another first-answer form. Explicit fresh practice generates a fresh instance and uses exposure rules.

Unavailable future stages explain why they are unavailable. If a permitted future/solution/reference view exposes helpful information, record assistance server-side before continuing. Skipping does not mark a wrong answer unless a defined exam policy says so. Return from tutor keeps the same original question even if a request completes after navigation.

## Save-on-leave

For an unsaved AI explanation, hold the intended action and display Save explanation? y saves then resumes; n discards then resumes; Esc stays. Scrolling, selection and harmless resizing do not prompt. If saving fails, retain content and intended action, show Retry, and remain on the note. Save is idempotent. Opening another problem, settings or arcade uses the same transition resolver.

Browsers cannot guarantee a custom y/n/Esc dialog during tab/window closure or crashes. Periodically keep an unsaved recovery draft locally through the backend, distinguish it from a saved library note, and restore it on restart. Use native beforeunload only for a meaningful unsaved state; do not claim browser close is identical to in-app quit.

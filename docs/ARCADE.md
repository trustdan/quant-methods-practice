# Startup space arcade plan

Working title: **Quant Flight**. Build an original arcade inspired by the existing side-flight/starfield/asteroid experience. Use original geometry, names, sounds and assets; do not copy Star Fox characters, branding, music or ship art. Canvas 2D first; no 3D/game-framework dependency is required for a first playable.

## Modes and flow

Launch -> attract/demo flight -> manual takeover when movement/fire is pressed -> sector progression -> win/continue or game over/retry -> optional new-record initials -> brief title reveal -> practice.

Enter/Esc and a visible Skip to practice work from every arcade state, including record entry/title reveal. A pending record may be recovered/saved later, never block practice. A stored skip-intro preference and --skip-intro bypass it entirely. From practice A opens arcade and L opens leaderboard; closing the leaderboard restores the original practice state. No game action changes a grade or mastery count.

Autopilot demonstrates steering/fire and hazards; demo scores are ineligible. Manual takeover starts a fresh scored run so pre-takeover demo points/survival cannot enter records. Four sectors replace accounting quarters. Finishing the fourth awards a survival/integrity bonus and offers continuation at increased difficulty or return to practice. Retry starts a new game seed/run, not a duplicate score submission.

## Controls and mechanics

| Control | Action |
|---|---|
| Arrow keys or WASD; j/k vertical alternatives | Steering, including simultaneous held directions |
| F or Space held | Continuous fire with controlled cooldown |
| g | Toggle autofire |
| b | Smart bomb; initial finite stock such as three, balanced during playtesting |
| p | Pause/resume |
| L | Leaderboard overlay, game paused |
| r | Retry after game over |
| y/n | Continue after sector-four win / return |
| Enter/Esc or Skip button | Immediate return to drills |

Use keydown/keyup state for simultaneous steering/fire; movement is not tied to OS repeat rate. Clear held keys on blur, pause, exit and visibility change. Audio stays muted until user gesture/opt-in; reduced motion suppresses flashes, shake and nonessential transitions.

Starfield/parallax, ship banking and perspective cues evoke flight while hazards move toward/along the play area. Light targets need one hit. Heavy hazards need multiple hits and fragment into a bounded number of smaller targets. Bomb consumes stock, creates a bounded shockwave and damages targets exactly once; cooldown prevents repeat from one held key.

Dual health: ship shields fall on collision; mission integrity falls when threats escape. Either reaching zero ends the run. Difficulty increases bounded speed/density over elapsed simulation time/sectors and score; retain spawn and entity caps. Suggested original hazard names: Selection Bias, Confounding, Bad Denominator, Missing Covariance and Spurious Correlation. Labels are atmosphere, not graded conceptual instruction.

## Simulation and renderer

Pure TypeScript step(state,input,dt,rng) drives entities, collisions, shots, particles, sectors and score. Use fixed simulation steps with a capped accumulation window, seeded RNG and bounded arrays. requestAnimationFrame renders interpolated state; pause on hidden tabs rather than simulating a giant catch-up. Reset timestamps on resume. Resizing updates logical coordinates without corrupting collisions or score.

Mount/unmount installs/removes one animation loop and one input owner. Exiting cancels animation and restores browser/practice focus. Game event handlers never run behind note fields or exam screens. Keep Canvas size/display scale consistent with devicePixelRatio. Provide text HUD and non-game alternatives for keyboard/screen-reader users.

## Persistence

Store game version/seed, mode, sectors, survival time, score, targets, final health, created time and bounded initials separately from practice. Only manual runs qualify. Prompt for three initials only for a new all-time record to match the accounting experience; offer explicit save and skip. Backend validates record shape/idempotency and accepted ranges. The local leaderboard is recreational; do not claim competitive anti-cheat security for a browser client. Demo results never receive a record prompt.

Show local record on HUD. --high-scores lists it. Settings allow resetting scores explicitly without deleting learner history. Title reveal can animate about one second then hold up to two seconds, but Skip/Enter/Esc bypass it immediately and reduced-motion mode opens practice directly.

## Staged delivery and acceptance

Stage 17: renderer + pure simulation + starfield + steering/fire + attract/takeover + skip/pause. Stage 18: bombs/autofire, accelerating hazards, fragmentation, dual-health failure, four-sector win/continue, persistent scores, initials and title reveal.

Tests cover simultaneous held input, one-time collision/kill scoring, bombs, fragmentation caps, speed/spawn bounds, demo exclusion, gameover/win/continue/reset, seeded determinism, pause/blur/resize and record idempotency. Browser smoke verifies skip in every state, no background game input, performance on an ordinary laptop, reduced motion and return to the same drill. Fun and balance require actual playtesting; unit tests do not establish them.

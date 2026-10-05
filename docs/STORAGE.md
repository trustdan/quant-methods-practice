# Storage and replay

SQLite stores personal runtime state in the platform user-data directory, separate from Git. Suggested application name: quant-methods-practice. Windows: APPDATA/quant-methods-practice; macOS: Library/Application Support/quant-methods-practice; Linux: XDG_DATA_HOME/quant-methods-practice or ~/.local/share/quant-methods-practice. Final path choice is recorded at implementation. Flags --data-dir and --db override it; temporary databases are used for tests.

## Tables and records

| Record | Required contents |
|---|---|
| schema_migrations | Version, checksum, applied time; transactional migrations |
| settings | Versioned local preferences, session size/intensity, skip-intro, reduced motion; no credentials |
| sessions | Type/mode, selected module/concepts, order, seed, active position, revision, start/end, exam deadline/policy |
| question_instances | Immutable family/template/rule/pedagogy versions, parameters/units, source metadata, stage prompts/options/order/hints, canonical results, grade policy and reviewed setting/contrast metadata |
| attempts | Immutable answer/normalized form, result, stage and concept attribution, command ID, UTC time, assistance and grading version |
| assistance_events | Reference/tutor/reveal/guided exposure, scope, UTC time; append-only |
| session_drafts | Ungraded inputs and active position with revision; recoverable, not evidence |
| mastery_projections | Rebuildable cache with evidence-policy version, never model-written |
| candidate_questions | Untrusted proposed content, source/provider, validation status; outside active bank |
| content_approval_events | Candidate/version, reviewer, delegated-review basis if any, source and notes, timestamp, publish/retire action |
| saved_explanations | Raw Markdown/LaTeX, originating snapshot/stage, topic/concepts, provider/model/route, creation/saving timestamps, thread references |
| tutor_drafts | Bounded unsaved recovery text and context, not saved notes or graded evidence |
| exam_responses | Original order/instance, submitted complete response, timestamp and status; grades withheld by policy |
| arcade_runs/high_scores | Game version/seed, manual/demo, duration, sectors, score, initials and timestamp; separate from mastery |

Question snapshots include private canonical answers but unresolved browser DTOs do not. Save an instance and its first attempt atomically where appropriate. Duplicate command IDs return the existing outcome. Do not allow an instance snapshot to be overwritten by later bank content. Retiring a question stops future selection without destroying history.

## Exams and restart

Create exam order and full snapshots before beginning. Persist started_at/deadline_at in UTC; a restart does not reset the clock. Pause behavior must be explicit and course-compatible. Resume from original stored instances and responses. Missing/corrupt snapshots produce a recoverable failure, never regenerated substitute questions. Time changes are handled conservatively and documented.

## Notes and exports

The saved library is authoritative in SQLite; Markdown export writes a human-readable copy with advisory label, question context, math and safe metadata. Raw text and context remain available without a provider. Explain save failures and leave the source readable. Explicit note deletion is a separate user command and does not delete originating attempts.

Export modes: notes Markdown; history JSON; approved bank JSON; exam report Markdown/JSON; complete database backup. Describe their different scopes. No export includes credentials or hidden unresolved answer keys inadvertently. Imported bank/content follows review; imported history cannot silently merge into mastery without validated policy/version handling.

Use safe local filenames and atomic replacement for generated exports; user-selected export locations are explicit. A browser download is a copy; it does not prove the database save succeeded. Directory export needs backend approval of the chosen path, not arbitrary path text from generated content.

## Migration and recovery

Create consistent SQLite backups before migrations using a supported backup mechanism. Never blindly copy only the main file while WAL writes are active. Tests cover fresh DB, each supported upgrade, failed migration preserving prior data, reopen/replay and duplicate writes. Show corruption/version errors with recovery guidance; never auto-reset or delete personal data.

-- 001_initial_schema.sql
-- Initial transactional schema for quant-methods-practice durable storage and replay.

CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value_json TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,
    mode TEXT NOT NULL,
    template_id TEXT NOT NULL,
    template_version INTEGER NOT NULL,
    seed INTEGER NOT NULL,
    revision INTEGER NOT NULL DEFAULT 1,
    current_question_index INTEGER NOT NULL DEFAULT 0,
    current_stage_index INTEGER NOT NULL DEFAULT 0,
    completed INTEGER NOT NULL DEFAULT 0,
    settings_json TEXT NOT NULL DEFAULT '{}',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    exam_deadline_at TEXT
);

CREATE INDEX IF NOT EXISTS idx_sessions_completed_updated ON sessions(completed, updated_at DESC);

CREATE TABLE IF NOT EXISTS question_instances (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    template_id TEXT NOT NULL,
    template_version INTEGER NOT NULL,
    seed INTEGER NOT NULL,
    parameters_json TEXT NOT NULL,
    title TEXT NOT NULL,
    scenario_markdown TEXT NOT NULL,
    stages_json TEXT NOT NULL,
    created_at TEXT NOT NULL,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_question_instances_session ON question_instances(session_id);

CREATE TABLE IF NOT EXISTS drill_stage_states (
    session_id TEXT NOT NULL,
    stage_index INTEGER NOT NULL,
    stage_id TEXT NOT NULL,
    status TEXT NOT NULL,
    first_try_correct INTEGER NOT NULL DEFAULT 0,
    solved_on_retry INTEGER NOT NULL DEFAULT 0,
    revealed INTEGER NOT NULL DEFAULT 0,
    active_hint TEXT NOT NULL DEFAULT '',
    last_feedback TEXT NOT NULL DEFAULT '',
    misconception_id TEXT,
    invalid_input_notice TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (session_id, stage_id),
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_drill_stage_states_session_idx ON drill_stage_states(session_id, stage_index ASC);

CREATE TABLE IF NOT EXISTS attempts (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    instance_id TEXT NOT NULL,
    stage_id TEXT NOT NULL,
    attempt_number INTEGER NOT NULL,
    command_id TEXT,
    submitted_answer_json TEXT NOT NULL,
    assistance_json TEXT NOT NULL,
    is_correct INTEGER NOT NULL,
    feedback_markdown TEXT NOT NULL,
    created_at TEXT NOT NULL,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE,
    FOREIGN KEY (instance_id) REFERENCES question_instances(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_attempts_session_stage ON attempts(session_id, stage_id, attempt_number ASC);

CREATE TABLE IF NOT EXISTS assistance_events (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    instance_id TEXT NOT NULL,
    stage_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    scope TEXT NOT NULL,
    created_at TEXT NOT NULL,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_assistance_events_session ON assistance_events(session_id, created_at ASC);

CREATE TABLE IF NOT EXISTS session_drafts (
    session_id TEXT NOT NULL,
    stage_id TEXT NOT NULL,
    draft_answer_json TEXT NOT NULL,
    active_position INTEGER NOT NULL DEFAULT 0,
    revision INTEGER NOT NULL,
    updated_at TEXT NOT NULL,
    PRIMARY KEY (session_id, stage_id),
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS command_idempotency (
    command_id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    command_type TEXT NOT NULL,
    result_json TEXT NOT NULL,
    created_at TEXT NOT NULL,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_command_idempotency_session ON command_idempotency(session_id);

CREATE TABLE IF NOT EXISTS mastery_projections (
    concept_id TEXT PRIMARY KEY,
    policy_version INTEGER NOT NULL,
    evidence_count INTEGER NOT NULL DEFAULT 0,
    independent_count INTEGER NOT NULL DEFAULT 0,
    assisted_count INTEGER NOT NULL DEFAULT 0,
    last_tested_at TEXT,
    projection_json TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS candidate_questions (
    id TEXT PRIMARY KEY,
    source TEXT NOT NULL,
    status TEXT NOT NULL,
    data_json TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS content_approval_events (
    id TEXT PRIMARY KEY,
    candidate_id TEXT NOT NULL,
    version INTEGER NOT NULL,
    reviewer TEXT NOT NULL,
    action TEXT NOT NULL,
    notes TEXT NOT NULL,
    timestamp TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS saved_explanations (
    id TEXT PRIMARY KEY,
    raw_markdown TEXT NOT NULL,
    origin_instance_id TEXT NOT NULL,
    origin_stage_id TEXT NOT NULL,
    topic TEXT NOT NULL,
    provider_info_json TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS tutor_drafts (
    id TEXT PRIMARY KEY,
    context_json TEXT NOT NULL,
    recovery_text TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS exam_responses (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    instance_id TEXT NOT NULL,
    stage_id TEXT NOT NULL,
    order_index INTEGER NOT NULL,
    response_json TEXT NOT NULL,
    status TEXT NOT NULL,
    submitted_at TEXT NOT NULL,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS arcade_scores (
    id TEXT PRIMARY KEY,
    game_version TEXT NOT NULL,
    seed INTEGER NOT NULL,
    is_demo INTEGER NOT NULL DEFAULT 0,
    duration_seconds REAL NOT NULL,
    sectors_cleared INTEGER NOT NULL,
    score INTEGER NOT NULL,
    initials TEXT NOT NULL,
    created_at TEXT NOT NULL
);

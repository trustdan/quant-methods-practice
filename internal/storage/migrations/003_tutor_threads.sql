-- Multi-turn conversational tutor threads and message persistence
CREATE TABLE tutor_threads (
    id TEXT PRIMARY KEY,
    session_id TEXT,
    instance_id TEXT NOT NULL,
    stage_id TEXT NOT NULL,
    provider TEXT NOT NULL,
    model TEXT NOT NULL,
    title TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE INDEX idx_tutor_threads_lookup ON tutor_threads(instance_id, stage_id);

CREATE TABLE tutor_messages (
    id TEXT PRIMARY KEY,
    thread_id TEXT NOT NULL REFERENCES tutor_threads(id) ON DELETE CASCADE,
    role TEXT NOT NULL,
    content TEXT NOT NULL,
    fallback_label TEXT,
    created_at TEXT NOT NULL
);

CREATE INDEX idx_tutor_messages_thread ON tutor_messages(thread_id, created_at);

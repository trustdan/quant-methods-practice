-- Full-form practice is explicitly separate from independent mastery projections.
CREATE TABLE worksheets (
    id TEXT PRIMARY KEY,
    revision INTEGER NOT NULL,
    data_json TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
CREATE TABLE worksheet_commands (
    worksheet_id TEXT NOT NULL REFERENCES worksheets(id),
    command_id TEXT NOT NULL,
    request_hash TEXT NOT NULL,
    result_json TEXT NOT NULL,
    PRIMARY KEY (worksheet_id, command_id)
);

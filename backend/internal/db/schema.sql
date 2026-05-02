-- nihongo-sensei SQLite schema

PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS grammar_points (
    id          INTEGER  PRIMARY KEY AUTOINCREMENT,
    jlpt_level  TEXT     NOT NULL CHECK(jlpt_level IN ('N5','N4','N3','N2','N1','')),
    pattern     TEXT     NOT NULL,
    meaning     TEXT     NOT NULL,
    example_jp  TEXT     NOT NULL DEFAULT '',
    example_en  TEXT     NOT NULL DEFAULT '',
    notes       TEXT     NOT NULL DEFAULT '',
    source      TEXT     NOT NULL DEFAULT 'manual',
    external_id TEXT     NOT NULL DEFAULT '',
    tags        TEXT     NOT NULL DEFAULT '',
    audio_url   TEXT     NOT NULL DEFAULT '',
    created_at  DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_grammar_jlpt ON grammar_points(jlpt_level);

-- Unique per (source, external_id) only when external_id is set.
CREATE UNIQUE INDEX IF NOT EXISTS idx_grammar_source_ext
    ON grammar_points(source, external_id) WHERE external_id != '';

CREATE TABLE IF NOT EXISTS review_cards (
    id               INTEGER  PRIMARY KEY AUTOINCREMENT,
    grammar_point_id INTEGER  NOT NULL REFERENCES grammar_points(id) ON DELETE CASCADE,
    interval         INTEGER  NOT NULL DEFAULT 1,
    repetitions      INTEGER  NOT NULL DEFAULT 0,
    e_factor         REAL     NOT NULL DEFAULT 2.5,
    due_date         DATETIME NOT NULL DEFAULT (datetime('now')),
    last_reviewed    DATETIME
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_card_grammar ON review_cards(grammar_point_id);
CREATE INDEX        IF NOT EXISTS idx_card_due     ON review_cards(due_date);

CREATE TABLE IF NOT EXISTS review_history (
    id           INTEGER  PRIMARY KEY AUTOINCREMENT,
    card_id      INTEGER  NOT NULL REFERENCES review_cards(id) ON DELETE CASCADE,
    grade        INTEGER  NOT NULL CHECK(grade BETWEEN 0 AND 5),
    reviewed_at  DATETIME NOT NULL DEFAULT (datetime('now')),
    llm_feedback TEXT     NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_history_card ON review_history(card_id);

CREATE TABLE IF NOT EXISTS import_log (
    id          INTEGER  PRIMARY KEY AUTOINCREMENT,
    source      TEXT     NOT NULL,
    started_at  DATETIME NOT NULL DEFAULT (datetime('now')),
    finished_at DATETIME,
    imported    INTEGER  NOT NULL DEFAULT 0,
    skipped     INTEGER  NOT NULL DEFAULT 0,
    errors      INTEGER  NOT NULL DEFAULT 0,
    status      TEXT     NOT NULL DEFAULT 'running',
    message     TEXT     NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_import_log_source ON import_log(source);

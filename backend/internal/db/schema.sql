-- shinkaku SQLite schema

PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS grammar_points (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    jlpt_level  TEXT    NOT NULL CHECK(jlpt_level IN ('N5','N4','N3','N2','N1','')),
    pattern     TEXT    NOT NULL,
    meaning     TEXT    NOT NULL,
    example_jp  TEXT    NOT NULL DEFAULT '',
    example_en  TEXT    NOT NULL DEFAULT '',
    notes       TEXT    NOT NULL DEFAULT '',
    source      TEXT    NOT NULL DEFAULT 'manual',
    created_at  DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_grammar_jlpt ON grammar_points(jlpt_level);

CREATE TABLE IF NOT EXISTS review_cards (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    grammar_point_id INTEGER NOT NULL REFERENCES grammar_points(id) ON DELETE CASCADE,
    interval         INTEGER NOT NULL DEFAULT 1,
    repetitions      INTEGER NOT NULL DEFAULT 0,
    e_factor         REAL    NOT NULL DEFAULT 2.5,
    due_date         DATETIME NOT NULL DEFAULT (datetime('now')),
    last_reviewed    DATETIME
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_card_grammar ON review_cards(grammar_point_id);
CREATE INDEX IF NOT EXISTS idx_card_due ON review_cards(due_date);

CREATE TABLE IF NOT EXISTS review_history (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    card_id      INTEGER NOT NULL REFERENCES review_cards(id) ON DELETE CASCADE,
    grade        INTEGER NOT NULL CHECK(grade BETWEEN 0 AND 5),
    reviewed_at  DATETIME NOT NULL DEFAULT (datetime('now')),
    llm_feedback TEXT    NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_history_card ON review_history(card_id);

-- Vocabulary -----------------------------------------------------------------

CREATE TABLE IF NOT EXISTS vocab_words (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    jlpt_level     TEXT    NOT NULL CHECK(jlpt_level IN ('N5','N4','N3','N2','N1','')),
    word           TEXT    NOT NULL,   -- kanji / primary kana form
    reading        TEXT    NOT NULL DEFAULT '',  -- hiragana reading
    meaning        TEXT    NOT NULL,   -- English gloss(es)
    part_of_speech TEXT    NOT NULL DEFAULT '',  -- e.g. "noun", "verb (godan)", etc.
    example_jp     TEXT    NOT NULL DEFAULT '',
    example_en     TEXT    NOT NULL DEFAULT '',
    notes          TEXT    NOT NULL DEFAULT '',
    source         TEXT    NOT NULL DEFAULT 'manual',
    created_at     DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_vocab_jlpt ON vocab_words(jlpt_level);

CREATE TABLE IF NOT EXISTS vocab_review_cards (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    vocab_word_id INTEGER NOT NULL REFERENCES vocab_words(id) ON DELETE CASCADE,
    interval      INTEGER NOT NULL DEFAULT 1,
    repetitions   INTEGER NOT NULL DEFAULT 0,
    e_factor      REAL    NOT NULL DEFAULT 2.5,
    due_date      DATETIME NOT NULL DEFAULT (datetime('now')),
    last_reviewed DATETIME
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_vocab_card_word ON vocab_review_cards(vocab_word_id);
CREATE INDEX IF NOT EXISTS idx_vocab_card_due ON vocab_review_cards(due_date);

CREATE TABLE IF NOT EXISTS vocab_review_history (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    card_id     INTEGER NOT NULL REFERENCES vocab_review_cards(id) ON DELETE CASCADE,
    grade       INTEGER NOT NULL CHECK(grade BETWEEN 0 AND 5),
    reviewed_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_vocab_history_card ON vocab_review_history(card_id);

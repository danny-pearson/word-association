-- +goose Up

CREATE TABLE puzzles (
  id                INTEGER PRIMARY KEY,
  answer            TEXT NOT NULL,
  answer_normalized TEXT NOT NULL UNIQUE,
  category          TEXT,
  difficulty        INTEGER,
  source_model      TEXT,
  created_at        TEXT NOT NULL
);

CREATE TABLE clues (
  id              INTEGER PRIMARY KEY,
  text            TEXT NOT NULL,
  text_normalized TEXT NOT NULL UNIQUE
);

CREATE TABLE puzzle_clues (
  puzzle_id INTEGER NOT NULL REFERENCES puzzles(id) ON DELETE CASCADE,
  clue_id   INTEGER NOT NULL REFERENCES clues(id),
  position  INTEGER NOT NULL CHECK (position BETWEEN 1 AND 5),
  PRIMARY KEY (puzzle_id, clue_id)
);
CREATE INDEX idx_puzzle_clues_clue ON puzzle_clues(clue_id);
CREATE UNIQUE INDEX idx_puzzle_clues_position ON puzzle_clues(puzzle_id, position);

CREATE TABLE answer_aliases (
  puzzle_id        INTEGER NOT NULL REFERENCES puzzles(id) ON DELETE CASCADE,
  alias_normalized TEXT NOT NULL,
  PRIMARY KEY (puzzle_id, alias_normalized)
);

CREATE TABLE games (
  id           TEXT PRIMARY KEY,
  player_id    TEXT NOT NULL,
  puzzle_id    INTEGER NOT NULL REFERENCES puzzles(id) ON DELETE RESTRICT,
  clues_shown  INTEGER NOT NULL DEFAULT 1 CHECK (clues_shown BETWEEN 1 AND 5),
  completed    BOOLEAN NOT NULL DEFAULT 0,
  won          BOOLEAN,
  started_at   TEXT NOT NULL,
  completed_at TEXT
);
CREATE INDEX idx_games_player ON games(player_id, puzzle_id);
CREATE INDEX idx_games_puzzle ON games(puzzle_id);

CREATE TABLE prompt_templates (
  id         INTEGER PRIMARY KEY,
  hash       TEXT NOT NULL UNIQUE,      -- sha256 of template text
  template   TEXT NOT NULL,
  label      TEXT NOT NULL,
  created_at TEXT NOT NULL
);

CREATE TABLE generation_runs (
  id          INTEGER PRIMARY KEY,
  prompt_id   INTEGER NOT NULL REFERENCES prompt_templates(id) ON DELETE RESTRICT,
  model       TEXT NOT NULL,
  requested   INTEGER NOT NULL,
  returned    INTEGER NOT NULL,
  accepted    INTEGER NOT NULL,
  started_at  TEXT NOT NULL,
  duration_ms INTEGER,
  error       TEXT
);

CREATE TABLE run_rejections (
  run_id INTEGER NOT NULL REFERENCES generation_runs(id) ON DELETE CASCADE,
  reason TEXT NOT NULL,
  count  INTEGER NOT NULL,
  PRIMARY KEY (run_id, reason)
);

-- +goose Down

DROP TABLE run_rejections;
DROP TABLE generation_runs;
DROP TABLE prompt_templates;
DROP TABLE games;
DROP TABLE answer_aliases;
DROP TABLE puzzle_clues;
DROP TABLE clues;
DROP TABLE puzzles;

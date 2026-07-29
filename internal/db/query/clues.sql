-- name: UpsertClue :one
INSERT INTO clues (text, text_normalized)
  VALUES (?, ?)
  ON CONFLICT (text_normalized) DO UPDATE SET text = excluded.text
  RETURNING id;

-- name: LinkClue :exec
INSERT INTO puzzle_clues (puzzle_id, clue_id, position)
  VALUES (?, ?, ?);

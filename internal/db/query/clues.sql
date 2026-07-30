-- name: UpsertClue :one
INSERT INTO clues (text, text_normalized)
  VALUES (?, ?)
  ON CONFLICT (text_normalized) DO UPDATE SET text = excluded.text
  RETURNING id;

-- name: LinkClue :exec
INSERT INTO puzzle_clues (puzzle_id, clue_id, position)
  VALUES (?, ?, ?);

-- name: GetPuzzleClue :one
SELECT c.text FROM clues c
  INNER JOIN puzzle_clues pc
    ON pc.clue_id = c.id

  WHERE pc.puzzle_id = ? AND pc.position = ?;

-- name: GetPuzzleCluesUpTo :many
SELECT c.text FROM clues c
  INNER JOIN puzzle_clues pc
    ON pc.clue_id = c.id

  WHERE pc.puzzle_id = ? AND pc.position <= ?
    
ORDER BY pc.position;

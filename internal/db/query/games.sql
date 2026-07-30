-- name: CreateGame :exec
INSERT INTO games (id, player_id, puzzle_id, started_at)
  VALUES (?, ?, ?, ?);

-- name: GetGame :one
SELECT id, puzzle_id, clues_shown, completed, won FROM games
  WHERE id = ? AND player_id = ?;

-- name: IncrementCluesShown :execrows
UPDATE games SET clues_shown = clues_shown + 1
  WHERE id = ?
    AND clues_shown < 5;

-- name: CompleteGame :execrows
UPDATE games SET completed = 1, completed_at = ?, won = ?
  WHERE id = ?
    AND completed = 0;

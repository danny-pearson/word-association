-- name: CountPuzzles :one
SELECT COUNT(*) FROM puzzles;

-- name: CreatePuzzle :one
INSERT INTO puzzles (answer, answer_normalized, source_model, created_at)
  VALUES (?, ?, ?, ?)
  RETURNING id;

-- name: FindExistingAnswers :many
SELECT answer_normalized FROM puzzles
  WHERE answer_normalized IN (sqlc.slice('answers'));

-- name: RecentAnswers :many
SELECT answer FROM puzzles
  ORDER BY created_at DESC
  LIMIT ?;
  
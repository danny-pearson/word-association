-- name: CountPuzzles :one
SELECT COUNT(*) FROM puzzles;

-- name: CreatePuzzle :one
INSERT INTO puzzles (answer, answer_normalized, source_model, created_at)
  VALUES (?, ?, ?, ?)
  RETURNING id;

-- name: GetPuzzle :one
SELECT * FROM puzzles
  WHERE id = ?;

-- name: GetRandomUnplayedPuzzle :one
SELECT p.id FROM puzzles p
  WHERE NOT EXISTS (
    SELECT 1 FROM games g
      WHERE g.puzzle_id = p.id
        AND g.player_id = ? AND g.completed = 1
  )
    
ORDER BY RANDOM()
LIMIT 1;

-- name: FindExistingAnswers :many
SELECT answer_normalized FROM puzzles
  WHERE answer_normalized IN (sqlc.slice('answers'));

-- name: RecentAnswers :many
SELECT answer FROM puzzles
ORDER BY created_at DESC
LIMIT ?;

-- name: PuzzleSupply :one
SELECT
  (SELECT COUNT(*) FROM puzzles) -
  (SELECT COUNT(DISTINCT puzzle_id) FROM games WHERE completed = 1) AS unplayed,
  (SELECT COUNT(DISTINCT player_id) FROM games) AS players;

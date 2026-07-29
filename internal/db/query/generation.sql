-- name: UpsertPromptTemplate :one
INSERT INTO prompt_templates (hash, template, label)
  VALUES (?, ?, ?)
  ON CONFLICT (hash) DO UPDATE SET
    template = excluded.template,
    label = excluded.label
  RETURNING id;

-- name: CreateGenerationRun :one
INSERT INTO generation_runs
(
  prompt_id,
  model,
  requested,
  returned,
  accepted,
  started_at,
  duration_ms,
  error
)
  VALUES (?, ?, ?, ?, ?, ?, ?, ?)
  RETURNING id;

-- name: RecordRejection :exec
INSERT INTO run_rejections (run_id, reason, count)
  VALUES (?, ?, ?);

-- name: CreateTask :one
INSERT INTO tasks (
  type, value, state
) VALUES (
  $1, $2, $3
)
RETURNING id, type, value, state, created_at, updated_at;

-- name: UpdateTaskState :exec
UPDATE tasks
SET state = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: GetTasksCount :one
SELECT COUNT(*) FROM tasks;

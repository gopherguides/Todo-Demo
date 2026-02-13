-- name: ListTasksByStatus :many
SELECT * FROM tasks
WHERE user_id = ? AND status = ?
ORDER BY position ASC;

-- name: ListTasksByUser :many
SELECT * FROM tasks
WHERE user_id = ?
ORDER BY status, position ASC;

-- name: GetTask :one
SELECT * FROM tasks
WHERE id = ? AND user_id = ?;

-- name: CreateTask :one
INSERT INTO tasks (user_id, title, description, priority, status, position, due_date)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateTask :one
UPDATE tasks
SET title = ?, description = ?, priority = ?, status = ?, due_date = ?, updated_at = datetime('now')
WHERE id = ? AND user_id = ?
RETURNING *;

-- name: DeleteTask :execresult
DELETE FROM tasks
WHERE id = ? AND user_id = ?;

-- name: UpdateTaskPosition :execresult
UPDATE tasks
SET status = ?, position = ?, updated_at = datetime('now')
WHERE id = ? AND user_id = ?;

-- name: GetMaxPosition :one
SELECT COALESCE(MAX(position), 0) AS max_position
FROM tasks
WHERE user_id = ? AND status = ?;

-- name: GetTaskPosition :one
SELECT position FROM tasks
WHERE id = ? AND user_id = ?;

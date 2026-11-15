-- name: ListTodos :many
SELECT id, title, description, completed, due_date, created_at, updated_at
FROM todos
ORDER BY created_at DESC, id DESC;

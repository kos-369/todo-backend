// Package repository stores todos in MySQL.
package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kos-369/todo-backend/internal/repository/internal/sqlcgen"
	"github.com/kos-369/todo-backend/internal/todo"
)

// TodoRepository reads and writes todos using the sqlc-generated queries.
type TodoRepository struct {
	q *sqlcgen.Queries
}

func NewTodoRepository(db *sql.DB) *TodoRepository {
	return &TodoRepository{q: sqlcgen.New(db)}
}

// List returns all todos, newest first.
func (r *TodoRepository) List(ctx context.Context) ([]todo.Todo, error) {
	rows, err := r.q.ListTodos(ctx)
	if err != nil {
		return nil, fmt.Errorf("list todos: %w", err)
	}
	// Non-nil, so an empty list stays [] rather than null in JSON.
	todos := make([]todo.Todo, 0, len(rows))
	for _, row := range rows {
		todos = append(todos, toTodo(row))
	}
	return todos, nil
}

// toTodo converts a sqlc row into the domain model.
func toTodo(r sqlcgen.Todo) todo.Todo {
	t := todo.Todo{
		ID:        int64(r.ID),
		Title:     r.Title,
		Completed: r.Completed,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
	if r.Description.Valid {
		t.Description = &r.Description.String
	}
	if r.DueDate.Valid {
		t.DueDate = &r.DueDate.Time
	}
	return t
}

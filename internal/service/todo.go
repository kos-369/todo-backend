// Package service holds the use cases. It sits between the handler and the repository.
package service

import (
	"context"

	"github.com/kos-369/todo-backend/internal/todo"
)

// TodoRepository is the storage TodoService needs.
// It is defined here, by the user of the interface, so the service doesn't depend on the repository package.
type TodoRepository interface {
	List(ctx context.Context) ([]todo.Todo, error)
}

// TodoService provides the todo use cases.
type TodoService struct {
	repo TodoRepository
}

// NewTodoService returns a TodoService that stores todos in repo.
func NewTodoService(repo TodoRepository) *TodoService {
	return &TodoService{repo: repo}
}

// List returns all todos, newest first.
func (s *TodoService) List(ctx context.Context) ([]todo.Todo, error) {
	return s.repo.List(ctx)
}

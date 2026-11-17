package handler

import (
	"context"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/kos-369/todo-backend/internal/handler/internal/oapigen"
	"github.com/kos-369/todo-backend/internal/todo"
)

// TodoService is the part of service.TodoService the handler needs.
type TodoService interface {
	List(ctx context.Context) ([]todo.Todo, error)
}

// todoHandler implements the todo operations of oapigen.StrictServerInterface.
type todoHandler struct {
	todos TodoService
}

func (h *todoHandler) ListTodos(ctx context.Context, _ oapigen.ListTodosRequestObject) (oapigen.ListTodosResponseObject, error) {
	todos, err := h.todos.List(ctx)
	if err != nil {
		return nil, err
	}
	// Non-nil, so an empty list stays [] rather than null in JSON.
	res := make([]oapigen.Todo, 0, len(todos))
	for _, t := range todos {
		res = append(res, toAPITodo(t))
	}
	return oapigen.ListTodos200JSONResponse{Todos: res}, nil
}

func toAPITodo(t todo.Todo) oapigen.Todo {
	a := oapigen.Todo{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Completed:   t.Completed,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
	if t.DueDate != nil {
		a.DueDate = &openapi_types.Date{Time: *t.DueDate}
	}
	return a
}

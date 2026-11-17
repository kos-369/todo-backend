package service_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/kos-369/todo-backend/internal/service"
	"github.com/kos-369/todo-backend/internal/todo"
)

// fakeTodoRepository returns fixed results, standing in for the DB.
type fakeTodoRepository struct {
	todos []todo.Todo
	err   error
}

func (f fakeTodoRepository) List(context.Context) ([]todo.Todo, error) {
	return f.todos, f.err
}

func TestTodoServiceList(t *testing.T) {
	t.Run("returns the todos from the repository", func(t *testing.T) {
		want := []todo.Todo{{ID: 2, Title: "new"}, {ID: 1, Title: "old"}}
		svc := service.NewTodoService(fakeTodoRepository{todos: want})

		got, err := svc.List(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("List() = %+v, want %+v", got, want)
		}
	})

	t.Run("returns the repository error", func(t *testing.T) {
		repoErr := errors.New("connection refused")
		svc := service.NewTodoService(fakeTodoRepository{err: repoErr})

		_, err := svc.List(context.Background())
		if !errors.Is(err, repoErr) {
			t.Errorf("err = %v, want %v", err, repoErr)
		}
	})
}

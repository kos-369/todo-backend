package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/kos-369/todo-backend/internal/handler"
	"github.com/kos-369/todo-backend/internal/todo"
)

// fakeTodoService returns fixed results, standing in for the service.
type fakeTodoService struct {
	todos []todo.Todo
	err   error
}

func (f fakeTodoService) List(context.Context) ([]todo.Todo, error) {
	return f.todos, f.err
}

func TestListTodos(t *testing.T) {
	memo := "memo"
	due := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	created := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	tests := []struct {
		name       string
		svc        fakeTodoService
		wantStatus int
		wantBody   string // JSON; empty means don't check
	}{
		{
			name: "todos",
			svc: fakeTodoService{todos: []todo.Todo{
				{ID: 2, Title: "new", Description: &memo, Completed: true, DueDate: &due, CreatedAt: created, UpdatedAt: created},
				{ID: 1, Title: "old", CreatedAt: created, UpdatedAt: created},
			}},
			wantStatus: http.StatusOK,
			wantBody: `{"todos": [
				{"id": 2, "title": "new", "description": "memo", "completed": true, "dueDate": "2026-12-31",
				 "createdAt": "2026-01-02T03:04:05Z", "updatedAt": "2026-01-02T03:04:05Z"},
				{"id": 1, "title": "old", "description": null, "completed": false, "dueDate": null,
				 "createdAt": "2026-01-02T03:04:05Z", "updatedAt": "2026-01-02T03:04:05Z"}
			]}`,
		},
		{
			name:       "no todos",
			svc:        fakeTodoService{todos: []todo.Todo{}},
			wantStatus: http.StatusOK,
			wantBody:   `{"todos": []}`,
		},
		{
			name:       "service error",
			svc:        fakeTodoService{err: errors.New("connection refused")},
			wantStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/todos", nil)
			rec := httptest.NewRecorder()

			// The API doesn't use the Pinger, so it is nil.
			handler.New(slog.New(slog.DiscardHandler), nil, tt.svc).ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if tt.wantBody == "" {
				return
			}
			if got := rec.Header().Get("Content-Type"); got != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", got)
			}
			assertJSONEqual(t, rec.Body.String(), tt.wantBody)
		})
	}
}

// assertJSONEqual compares JSON by value, so key order and spacing don't matter.
func assertJSONEqual(t *testing.T, got, want string) {
	t.Helper()
	var g, w any
	if err := json.Unmarshal([]byte(got), &g); err != nil {
		t.Fatalf("invalid JSON %q: %v", got, err)
	}
	if err := json.Unmarshal([]byte(want), &w); err != nil {
		t.Fatalf("invalid want JSON: %v", err)
	}
	if !reflect.DeepEqual(g, w) {
		t.Errorf("body = %s, want %s", got, want)
	}
}

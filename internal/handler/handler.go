// Package handler serves the HTTP API defined in api/openapi.yaml.
package handler

import (
	"net/http"

	"github.com/kos-369/todo-backend/internal/handler/internal/oapigen"
)

// New returns the HTTP routes for the API.
func New(todos TodoService) http.Handler {
	return oapigen.Handler(oapigen.NewStrictHandler(&todoHandler{todos: todos}, nil))
}

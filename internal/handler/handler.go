// Package handler serves the HTTP API defined in api/openapi.yaml.
package handler

import (
	"log/slog"
	"net/http"

	"github.com/kos-369/todo-backend/internal/handler/internal/oapigen"
)

// New returns all HTTP routes: the health checks and the API.
func New(logger *slog.Logger, db Pinger, todos TodoService) http.Handler {
	mux := http.NewServeMux()
	registerHealth(mux, logger, db)
	return oapigen.HandlerFromMux(oapigen.NewStrictHandler(&todoHandler{todos: todos}, nil), mux)
}

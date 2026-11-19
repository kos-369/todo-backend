package handler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// readyzTimeout bounds the DB ping so /readyz answers quickly even if the DB hangs.
const readyzTimeout = 2 * time.Second

// Pinger is the part of *sql.DB that /readyz needs, so tests can pass a fake.
type Pinger interface {
	PingContext(ctx context.Context) error
}

// registerHealth adds the health checks. They are not in the OpenAPI spec
// because they are for the load balancer, not for API clients.
func registerHealth(mux *http.ServeMux, logger *slog.Logger, db Pinger) {
	// Liveness: the process is up. It doesn't check the DB, so a DB outage
	// doesn't make ECS restart healthy tasks.
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})

	// Readiness: the DB is reachable.
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), readyzTimeout)
		defer cancel()
		if err := db.PingContext(ctx); err != nil {
			// Log the cause; don't send DB details to the client.
			logger.Warn("readyz: DB ping failed", "error", err)
			http.Error(w, "not ready", http.StatusServiceUnavailable)
			return
		}
		fmt.Fprintln(w, "ok")
	})
}

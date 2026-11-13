package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-sql-driver/mysql"
)

// shutdownTimeout limits how long Shutdown waits for in-flight requests.
// Keep it below the ECS stop timeout (30s by default).
const shutdownTimeout = 10 * time.Second

// readyzTimeout bounds the DB ping so /readyz answers quickly even if the DB hangs.
const readyzTimeout = 2 * time.Second

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Cancel ctx on Ctrl+C (SIGINT) or when ECS stops the task (SIGTERM).
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := dbConfigFromEnv()
	if err != nil {
		logger.Error("invalid DB config", "error", err)
		os.Exit(1)
	}
	connector, err := mysql.NewConnector(cfg)
	if err != nil {
		logger.Error("invalid DB config", "error", err)
		os.Exit(1)
	}
	// OpenDB doesn't connect yet, so the server starts even if the DB is down.
	db := sql.OpenDB(connector)
	defer db.Close()

	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		logger.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	if err := run(ctx, logger, ln, newHandler(logger, db)); err != nil {
		logger.Error("server stopped with error", "error", err)
		os.Exit(1)
	}
}

// run serves HTTP until ctx is canceled, then shuts down gracefully.
// Graceful shutdown lets in-flight requests finish, so deploys don't
// cut off clients mid-request.
// It takes a listener so tests can use a random port.
func run(ctx context.Context, logger *slog.Logger, ln net.Listener, handler http.Handler) error {
	srv := &http.Server{Handler: handler}

	// Buffered so the goroutine can exit even if no one receives.
	serveErr := make(chan error, 1)
	go func() {
		logger.Info("server starting", "addr", ln.Addr().String())
		serveErr <- srv.Serve(ln)
	}()

	select {
	case err := <-serveErr:
		return err
	case <-ctx.Done():
	}

	logger.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}
	logger.Info("server stopped")
	return nil
}

// pinger is the part of *sql.DB that /readyz needs, so tests can pass a fake.
type pinger interface {
	PingContext(ctx context.Context) error
}

// newHandler returns the HTTP routes.
func newHandler(logger *slog.Logger, db pinger) http.Handler {
	mux := http.NewServeMux()

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

	return mux
}

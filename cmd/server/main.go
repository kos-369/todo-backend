package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-sql-driver/mysql"

	"github.com/kos-369/todo-backend/internal/handler"
	"github.com/kos-369/todo-backend/internal/repository"
	"github.com/kos-369/todo-backend/internal/service"
)

// shutdownTimeout limits how long Shutdown waits for in-flight requests.
// Keep it below the ECS stop timeout (30s by default).
const shutdownTimeout = 10 * time.Second

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

	routes := handler.New(logger, db, service.NewTodoService(repository.NewTodoRepository(db)))
	if err := run(ctx, logger, ln, routes); err != nil {
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

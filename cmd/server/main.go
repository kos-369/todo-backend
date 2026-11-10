package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	addr := ":8080"
	logger.Info("server starting", "addr", addr)
	if err := http.ListenAndServe(addr, newHandler()); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func newHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})
	return mux
}

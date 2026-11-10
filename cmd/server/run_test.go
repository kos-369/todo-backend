package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"testing"
	"time"
)

// TestRunGracefulShutdown checks that a request in progress at shutdown still completes.
func TestRunGracefulShutdown(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	started := make(chan struct{})
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		time.Sleep(500 * time.Millisecond)
		fmt.Fprintln(w, "done")
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	runErr := make(chan error, 1)
	go func() {
		runErr <- run(ctx, logger, ln, handler)
	}()

	type result struct {
		status int
		body   string
		err    error
	}
	resCh := make(chan result, 1)
	go func() {
		resp, err := http.Get("http://" + ln.Addr().String())
		if err != nil {
			resCh <- result{err: err}
			return
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		resCh <- result{status: resp.StatusCode, body: string(body), err: err}
	}()

	<-started
	cancel()

	res := <-resCh
	if res.err != nil {
		t.Fatalf("request failed: %v", res.err)
	}
	if res.status != http.StatusOK {
		t.Errorf("status = %d, want %d", res.status, http.StatusOK)
	}
	if res.body != "done\n" {
		t.Errorf("body = %q, want %q", res.body, "done\n")
	}

	if err := <-runErr; err != nil {
		t.Errorf("run returned error: %v", err)
	}
}

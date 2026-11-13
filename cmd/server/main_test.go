package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthz(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		wantBody   string
	}{
		{"GET returns ok", http.MethodGet, "/healthz", http.StatusOK, "ok\n"},
		{"POST is not allowed", http.MethodPost, "/healthz", http.StatusMethodNotAllowed, ""},
		{"unknown path", http.MethodGet, "/unknown", http.StatusNotFound, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			newHandler(slog.New(slog.DiscardHandler), fakePinger{}).ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if tt.wantBody != "" && rec.Body.String() != tt.wantBody {
				t.Errorf("body = %q, want %q", rec.Body.String(), tt.wantBody)
			}
		})
	}
}

// fakePinger returns err from PingContext, standing in for the DB.
type fakePinger struct {
	err error
}

func (f fakePinger) PingContext(context.Context) error {
	return f.err
}

func TestReadyz(t *testing.T) {
	tests := []struct {
		name       string
		pingErr    error
		wantStatus int
		wantBody   string
	}{
		{"DB reachable", nil, http.StatusOK, "ok\n"},
		{"DB unreachable", errors.New("connection refused"), http.StatusServiceUnavailable, "not ready\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
			rec := httptest.NewRecorder()

			newHandler(slog.New(slog.DiscardHandler), fakePinger{err: tt.pingErr}).ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if rec.Body.String() != tt.wantBody {
				t.Errorf("body = %q, want %q", rec.Body.String(), tt.wantBody)
			}
		})
	}
}

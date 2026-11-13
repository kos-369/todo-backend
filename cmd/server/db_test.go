package main

import (
	"strings"
	"testing"
)

func TestDBConfigFromEnv(t *testing.T) {
	t.Run("all set", func(t *testing.T) {
		t.Setenv("DB_ADDR", "127.0.0.1:3306")
		t.Setenv("DB_USER", "todo")
		t.Setenv("DB_PASSWORD", "secret")
		t.Setenv("DB_NAME", "todo")

		cfg, err := dbConfigFromEnv()
		if err != nil {
			t.Fatal(err)
		}
		want := "todo:secret@tcp(127.0.0.1:3306)/todo"
		if got := cfg.FormatDSN(); got != want {
			t.Errorf("DSN = %q, want %q", got, want)
		}
	})

	t.Run("missing values", func(t *testing.T) {
		t.Setenv("DB_ADDR", "127.0.0.1:3306")
		t.Setenv("DB_USER", "")
		t.Setenv("DB_PASSWORD", "")
		t.Setenv("DB_NAME", "todo")

		_, err := dbConfigFromEnv()
		if err == nil {
			t.Fatal("err = nil, want an error")
		}
		for _, key := range []string{"DB_USER", "DB_PASSWORD"} {
			if !strings.Contains(err.Error(), key) {
				t.Errorf("err = %q, want it to mention %s", err, key)
			}
		}
	})
}

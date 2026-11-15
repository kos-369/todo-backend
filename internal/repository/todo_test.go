package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/kos-369/todo-backend/internal/repository"
	"github.com/kos-369/todo-backend/internal/testdb"
)

func TestTodoRepositoryList(t *testing.T) {
	db := testdb.New(t)
	repo := repository.NewTodoRepository(db)
	ctx := context.Background()

	got, err := repo.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || len(got) != 0 {
		t.Fatalf("empty table: got %#v, want an empty non-nil slice", got)
	}

	// Set created_at explicitly so the expected order doesn't depend on timing.
	_, err = db.ExecContext(ctx, `
		INSERT INTO todos (title, description, due_date, created_at) VALUES
		('old', NULL, NULL, '2026-01-01 00:00:00'),
		('new', 'memo', '2026-12-31', '2026-01-02 00:00:00')`)
	if err != nil {
		t.Fatal(err)
	}

	got, err = repo.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d todos, want 2", len(got))
	}
	newer, older := got[0], got[1]
	if newer.Title != "new" || older.Title != "old" {
		t.Fatalf("order: got %q, %q; want new, old", newer.Title, older.Title)
	}

	if newer.Description == nil || *newer.Description != "memo" {
		t.Errorf("description: got %v, want memo", newer.Description)
	}
	wantDue := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	if newer.DueDate == nil || !newer.DueDate.Equal(wantDue) {
		t.Errorf("due date: got %v, want %v", newer.DueDate, wantDue)
	}
	if older.Description != nil || older.DueDate != nil {
		t.Errorf("NULL columns: got description %v, due date %v; want nil", older.Description, older.DueDate)
	}
}

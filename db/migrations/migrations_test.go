package migrations_test

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/testcontainers/testcontainers-go"
	tcmysql "github.com/testcontainers/testcontainers-go/modules/mysql"

	"github.com/kos-369/todo-backend/db/migrations"
)

// TestMigrations applies all migrations to a throwaway MySQL, then rolls
// them back and applies them again, so both up and down files are checked.
func TestMigrations(t *testing.T) {
	if testing.Short() {
		t.Skip("needs Docker")
	}
	ctx := context.Background()

	ctr, err := tcmysql.Run(ctx, "mysql:8.4",
		tcmysql.WithDatabase("todo"),
		tcmysql.WithUsername("todo"),
		tcmysql.WithPassword("todo"),
	)
	testcontainers.CleanupContainer(t, ctr)
	if err != nil {
		t.Fatal(err)
	}
	dsn, err := ctr.ConnectionString(ctx)
	if err != nil {
		t.Fatal(err)
	}

	src, err := iofs.New(migrations.FS, ".")
	if err != nil {
		t.Fatal(err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", src, "mysql://"+dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := m.Up(); err != nil {
		t.Fatalf("up: %v", err)
	}
	if !tableExists(t, db, "todos") {
		t.Fatal("todos table missing after up")
	}

	if err := m.Down(); err != nil {
		t.Fatalf("down: %v", err)
	}
	if tableExists(t, db, "todos") {
		t.Fatal("todos table still exists after down")
	}

	if err := m.Up(); err != nil {
		t.Fatalf("up after down: %v", err)
	}
}

// tableExists reports whether the table is in the current database.
func tableExists(t *testing.T, db *sql.DB, name string) bool {
	t.Helper()
	var n int
	err := db.QueryRow(
		"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?",
		name,
	).Scan(&n)
	if err != nil {
		t.Fatal(err)
	}
	return n == 1
}

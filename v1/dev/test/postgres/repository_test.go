package postgres_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"http/v1/dev/internal/config"
	"http/v1/dev/internal/storage/postgres"
	"http/v1/dev/internal/user"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var mockedUser = user.User{
	ID:   1,
	Name: "Test",
	Age:  19,
}

var mockEnvs = map[string]string{
	"DB_HOST":     "localhost",
	"DB_PORT":     "5242",
	"DB_DATABASE": "http",
	"DB_USER":     "test",
	"DB_PASSWORD": "test",
}

func TestCreate(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db := openTestDB(t)
	defer db.Close()

	resetUsersTable(t, db)

	repo := postgres.New(db)

	user, err := repo.Create(ctx, mockedUser)
	if err != nil {
		t.Fatalf("error while creating the user: %v", err)
	}

	if user != mockedUser {
		t.Fatalf("error while creating the user: %v", err)
	}
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dbCtx, dbStop := context.WithTimeout(context.Background(), 3*time.Second)
	defer dbStop()

	for key, value := range mockEnvs {
		t.Setenv(key, value)
	}

	c, err := config.Load()
	if err != nil {
		t.Fatalf("error when loading the config, err: %v", err)
	}

	db, err := sql.Open("pgx", c.DatabaseURL)
	if err != nil {
		t.Fatalf("error when opening the connection, err: %v", err)
	}

	err = db.PingContext(dbCtx)
	if err != nil {
		t.Fatalf("error when pinging the db, err: %v", err)
	}

	return db
}

func resetUsersTable(t *testing.T, db *sql.DB) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	const createUsersTable = `
			CREATE TABLE IF NOT EXISTS users (
                id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
                name TEXT NOT NULL UNIQUE,
                age INTEGER NOT NULL CHECK (age >= 0)
            ;
		`

	if _, err := db.ExecContext(ctx, createUsersTable); err != nil {
		t.Fatalf("create users table: %v", err)
	}

	if _, err := db.ExecContext(ctx, "TRUNCATE users RESTART IDENTITY"); err != nil {
		t.Fatalf("truncate users table: %v", err)
	}
}

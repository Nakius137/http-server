package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"http/v1/dev/internal/config"
	"http/v1/dev/internal/storage/postgres"
	"http/v1/dev/internal/user"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var mockedUser = user.User{
	Name: "Test",
	Age:  19,
}

var secondMockedUser = user.User{
	Name: "Second",
	Age:  25,
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

	created, err := repo.Create(ctx, mockedUser)
	if err != nil {
		t.Fatalf("error while creating the user: %v", err)
	}

	if created.ID <= 0 {
		t.Fatalf("expected generated user ID to be greater than 0, got %d", created.ID)
	}

	if created.Name != mockedUser.Name {
		t.Fatalf("expected user name %q, got %q", mockedUser.Name, created.Name)
	}

	if created.Age != mockedUser.Age {
		t.Fatalf("expected user age %d, got %d", mockedUser.Age, created.Age)
	}
}

func TestCreateAlreadyExists(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	db := openTestDB(t)
	defer db.Close()
	resetUsersTable(t, db)

	repo := postgres.New(db)

	_, err := repo.Create(ctx, mockedUser)
	if err != nil {
		t.Fatalf("error while creating first user: %v", err)
	}

	_, err = repo.Create(ctx, user.User{Name: mockedUser.Name, Age: 30})
	if !errors.Is(err, user.ErrAlreadyExists) {
		t.Fatalf("expected user.ErrAlreadyExists, got %v", err)
	}
}

func TestList(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	db := openTestDB(t)
	defer db.Close()
	resetUsersTable(t, db)

	repo := postgres.New(db)

	firstUser, err := repo.Create(ctx, mockedUser)
	if err != nil {
		t.Fatalf("error while creating first user: %v", err)
	}

	secondUser, err := repo.Create(ctx, secondMockedUser)
	if err != nil {
		t.Fatalf("error while creating second user: %v", err)
	}

	users, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("error while fetching list of users: %v", err)
	}

	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}

	if users[0] != firstUser {
		t.Fatalf("expected first user %+v, got %+v", firstUser, users[0])
	}

	if users[1] != secondUser {
		t.Fatalf("expected second user %+v, got %+v", secondUser, users[1])
	}
}

func TestDeleteByName(t *testing.T) {
	t.Run("deletes existing user", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		db := openTestDB(t)
		defer db.Close()
		resetUsersTable(t, db)

		repo := postgres.New(db)

		created, err := repo.Create(ctx, user.User{Name: "Test", Age: 19})
		if err != nil {
			t.Fatalf("error while creating a user, %v", err)
		}

		deleted, err := repo.DeleteByName(ctx, "Test")
		if err != nil {
			t.Fatalf("error while deleting a user, %v", err)
		}

		if created != deleted {
			t.Fatalf("expected deleted user %+v, got %+v", created, deleted)
		}

		users, err := repo.List(ctx)
		if err != nil {
			t.Fatalf("error while fetching a user, %v", err)
		}

		if len(users) != 0 {
			t.Fatalf("expected 0 users after delete, got %d", len(users))
		}
	})

	t.Run("returns not found for missing user", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		db := openTestDB(t)
		defer db.Close()
		resetUsersTable(t, db)

		repo := postgres.New(db)

		_, err := repo.DeleteByName(ctx, "missing")
		if !errors.Is(err, user.ErrNotFound) {
			t.Fatalf("expected user.ErrNotFound, got %v", err)
		}
	})
}

func TestUpdateByName(t *testing.T) {
	t.Run("updates existing user", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		db := openTestDB(t)
		defer db.Close()
		resetUsersTable(t, db)

		repo := postgres.New(db)

		created, err := repo.Create(ctx, mockedUser)
		if err != nil {
			t.Fatalf("error while creating a user, %v", err)
		}

		update := user.User{Name: "Updated", Age: 30}

		updated, err := repo.UpdateByName(ctx, created.Name, update)
		if err != nil {
			t.Fatalf("error while updating a user, %v", err)
		}

		if updated.ID != created.ID {
			t.Fatalf("expected updated user ID %d, got %d", created.ID, updated.ID)
		}

		if updated.Name != update.Name {
			t.Fatalf("expected updated user name %q, got %q", update.Name, updated.Name)
		}

		if updated.Age != update.Age {
			t.Fatalf("expected updated user age %d, got %d", update.Age, updated.Age)
		}

		users, err := repo.List(ctx)
		if err != nil {
			t.Fatalf("error while fetching users, %v", err)
		}

		if len(users) != 1 {
			t.Fatalf("expected 1 user after update, got %d", len(users))
		}

		if users[0] != updated {
			t.Fatalf("expected persisted user %+v, got %+v", updated, users[0])
		}
	})

	t.Run("returns not found for missing user", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		db := openTestDB(t)
		defer db.Close()
		resetUsersTable(t, db)

		repo := postgres.New(db)

		_, err := repo.UpdateByName(ctx, "missing", user.User{Name: "Updated", Age: 30})
		if !errors.Is(err, user.ErrNotFound) {
			t.Fatalf("expected user.ErrNotFound, got %v", err)
		}
	})

	t.Run("returns already exists for duplicate name", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		db := openTestDB(t)
		defer db.Close()
		resetUsersTable(t, db)

		repo := postgres.New(db)

		firstUser, err := repo.Create(ctx, mockedUser)
		if err != nil {
			t.Fatalf("error while creating first user: %v", err)
		}

		secondUser, err := repo.Create(ctx, secondMockedUser)
		if err != nil {
			t.Fatalf("error while creating second user: %v", err)
		}

		_, err = repo.UpdateByName(ctx, firstUser.Name, user.User{Name: secondUser.Name, Age: 30})
		if !errors.Is(err, user.ErrAlreadyExists) {
			t.Fatalf("expected user.ErrAlreadyExists, got %v", err)
		}
	})
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
		db.Close()
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
            );
		`

	if _, err := db.ExecContext(ctx, createUsersTable); err != nil {
		t.Fatalf("create users table: %v", err)
	}

	if _, err := db.ExecContext(ctx, "TRUNCATE users RESTART IDENTITY"); err != nil {
		t.Fatalf("truncate users table: %v", err)
	}
}

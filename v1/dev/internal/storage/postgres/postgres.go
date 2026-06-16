package postgres

import (
	"context"
	"database/sql"
	"errors"
	"http/v1/dev/internal/user"

	"github.com/jackc/pgx/v5/pgconn"
)

const uniqueViolationCode = "23505"

type UserRepository struct {
	db *sql.DB
}

var _ user.Repository = (*UserRepository)(nil)

func New(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) List(ctx context.Context) ([]user.User, error) {
	var users []user.User

	rows, err := r.db.QueryContext(ctx, "SELECT id, name, age FROM users ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var u user.User

		err = rows.Scan(&u.ID, &u.Name, &u.Age)
		if err != nil {
			return nil, err
		}

		users = append(users, u)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (r *UserRepository) Create(ctx context.Context, u user.User) (user.User, error) {
	var newUser user.User

	row := r.db.QueryRowContext(ctx, "INSERT INTO users (name, age) VALUES ($1, $2) RETURNING id, name, age", u.Name, u.Age)

	err := row.Scan(&newUser.ID, &newUser.Name, &newUser.Age)
	if err != nil {
		if isUniqueViolation(err) {
			return user.User{}, user.ErrAlreadyExists
		}

		return user.User{}, err
	}

	return newUser, nil
}

func (r *UserRepository) DeleteByName(ctx context.Context, name string) (user.User, error) {
	var deletedUser user.User

	row := r.db.QueryRowContext(ctx, "DELETE FROM users WHERE name = $1 RETURNING id, name, age", name)
	err := row.Scan(&deletedUser.ID, &deletedUser.Name, &deletedUser.Age)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return user.User{}, user.ErrNotFound
		}

		return user.User{}, err
	}

	return deletedUser, nil
}

func (r *UserRepository) UpdateByName(ctx context.Context, name string, u user.User) (user.User, error) {
	var updatedUser user.User

	row := r.db.QueryRowContext(ctx, "UPDATE users SET name = $2, age = $3 WHERE name = $1 RETURNING id, name, age", name, u.Name, u.Age)
	err := row.Scan(&updatedUser.ID, &updatedUser.Name, &updatedUser.Age)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return user.User{}, user.ErrNotFound
		}
		if isUniqueViolation(err) {
			return user.User{}, user.ErrAlreadyExists
		}

		return user.User{}, err
	}

	return updatedUser, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode
}

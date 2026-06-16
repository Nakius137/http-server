package user

import (
	"context"
	"errors"
)

type User struct {
	ID   int64
	Name string
	Age  int
}

var (
	ErrNotFound      = errors.New("user not found")
	ErrAlreadyExists = errors.New("user already exists")
)

type Repository interface {
	List(ctx context.Context) ([]User, error)
	Create(ctx context.Context, u User) (User, error)
	DeleteByName(ctx context.Context, name string) (User, error)
	UpdateByName(ctx context.Context, name string, u User) (User, error)
}

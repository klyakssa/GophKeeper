package auth

import "context"

//go:generate mockgen -source=repository.go -destination=mocks/repository.go

// Repository is an interface for auth repository
type Repository interface {
	CreateUser(ctx context.Context, login, password string) (string, error) // return user ID and error
	GetUserByLogin(ctx context.Context, login string) (*User, error)        // return user and error
}

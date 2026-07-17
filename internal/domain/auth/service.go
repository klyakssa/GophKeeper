package auth

import "context"

//go:generate mockgen -source=service.go -destination=mocks/service.go

// Service is an interface for auth service
type Service interface {
	Register(ctx context.Context, username, password string) (string, error) // Register registers a new user and returns jwt token and error
	Login(ctx context.Context, username, password string) (string, error)    // Login returns jwt token and error
}

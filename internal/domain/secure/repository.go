package secure

import "context"

//go:generate mockgen -source=repository.go -destination=mocks/repository.go

// Repository is an interface for secure repository
type Repository interface {
	CreateSecureData(ctx context.Context, req *SecureDataCreate) error
	GetSecureData(ctx context.Context, userID string) ([]SecureData, error)
	DeleteSecureData(ctx context.Context, userID string, id string) error
	UpdateSecureData(ctx context.Context, req *SecureDataUpdate, userid string) error
}

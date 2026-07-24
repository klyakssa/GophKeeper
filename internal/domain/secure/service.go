package secure

import "context"

//go:generate mockgen -source=service.go -destination=mocks/service.go

// Service is an interface for auth service
type Service interface {
	CreateSecureData(ctx context.Context, req *SecureDataCreate) error
	GetSecureData(ctx context.Context, userID string) ([]SecureData, error)
	DeleteSecureData(ctx context.Context, userID string, id string) error
	UpdateSecureData(ctx context.Context, req *SecureDataUpdate, userid string) error
}

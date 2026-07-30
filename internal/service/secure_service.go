package service

import (
	"context"
	"gophkeeper/internal/domain/secure"
)

type SecureService struct {
	repo secure.Repository
}

func NewSecureService(repo secure.Repository) *SecureService {
	return &SecureService{repo: repo}
}

func (s *SecureService) CreateSecureData(ctx context.Context, req *secure.SecureDataCreate) error {
	return s.repo.CreateSecureData(ctx, req)
}

func (s *SecureService) GetSecureData(ctx context.Context, userID string) ([]secure.SecureData, error) {
	return s.repo.GetSecureData(ctx, userID)
}

func (s *SecureService) DeleteSecureData(ctx context.Context, userID string, id int) error {
	return s.repo.DeleteSecureData(ctx, userID, id)
}

func (s *SecureService) UpdateSecureData(ctx context.Context, req *secure.SecureDataUpdate) error {
	return s.repo.UpdateSecureData(ctx, req)
}

package service

import (
	"context"
	"errors"
	"gophkeeper/internal/domain/auth"
	"gophkeeper/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
)

// AuthService
type AuthService struct {
	repo       auth.Repository
	jwtManager *jwt.JWTManager
}

// NewAuthService creates a new instance of AuthService
func NewAuthService(repo auth.Repository, jwtManager *jwt.JWTManager) *AuthService {
	return &AuthService{repo: repo, jwtManager: jwtManager}
}

// Register registers a new user
func (s *AuthService) Register(ctx context.Context, login, password string) (string, error) {
	if len(login) > 255 {
		return "", auth.ErrLoginTooLong
	}

	if _, err := s.repo.GetUserByLogin(ctx, login); err == nil {
		return "", auth.ErrUserAlreadyExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		if errors.Is(err, bcrypt.ErrPasswordTooLong) {
			return "", auth.ErrPasswordTooLong
		}
		return "", err
	}

	userid, err := s.repo.CreateUser(ctx, login, string(hash))
	if err != nil {
		return "", err
	}

	token, err := s.jwtManager.GenerateToken(userid)
	if err != nil {
		return "", err
	}

	return token, nil
}

// Login logs in a user
func (s *AuthService) Login(ctx context.Context, login, password string) (string, error) {
	user, err := s.repo.GetUserByLogin(ctx, login)
	if err != nil {
		return "", auth.ErrUserNotFound
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", auth.ErrInvalidCredentials
	}

	token, err := s.jwtManager.GenerateToken(user.ID)
	if err != nil {
		return "", err
	}

	return token, nil
}

package jwt_test

import (
	"gophkeeper/pkg/jwt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestJWTManager_GenerateToken(t *testing.T) {

	j := jwt.NewJWTManager("secret", time.Hour)

	tests := []struct {
		name   string
		userID string
	}{
		{
			name:   "valid token generation",
			userID: "user123",
		},
		{
			name:   "another user",
			userID: "admin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			token, err := j.GenerateToken(tt.userID)

			assert.NoError(t, err)
			assert.NotEmpty(t, token)

			claims, err := j.VerifyToken(token)

			assert.NoError(t, err)
			assert.Equal(t, tt.userID, claims.UserID)
		})
	}
}

func TestJWTManager_VerifyToken(t *testing.T) {

	secret := "secret"
	j := jwt.NewJWTManager(secret, time.Hour)

	validToken, _ := j.GenerateToken("user123")

	tests := []struct {
		name     string
		tokenStr string
		manager  *jwt.JWTManager
		wantErr  bool
	}{
		{
			name:     "valid token",
			tokenStr: validToken,
			manager:  j,
			wantErr:  false,
		},
		{
			name:     "invalid token string",
			tokenStr: "invalid.token.here",
			manager:  j,
			wantErr:  true,
		},
		{
			name:     "wrong secret",
			tokenStr: validToken,
			manager:  jwt.NewJWTManager("another-secret", time.Hour),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			claims, err := tt.manager.VerifyToken(tt.tokenStr)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, claims)
		})
	}
}

func TestJWTManager_TokenExpired(t *testing.T) {

	j := jwt.NewJWTManager("secret", time.Millisecond)

	token, _ := j.GenerateToken("user")

	time.Sleep(2 * time.Millisecond)

	_, err := j.VerifyToken(token)

	assert.Error(t, err)
}

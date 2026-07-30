package http_test

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"gophkeeper/internal/domain/auth"
	mock_auth "gophkeeper/internal/domain/auth/mocks"
	"gophkeeper/internal/logger"
	httptransport "gophkeeper/internal/transport/http"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestAuthHandler_Register(t *testing.T) {

	type mockBehavior func(s *mock_auth.MockService, login, password string)

	logger := &logger.Logger{Logger: zap.NewNop()}

	tests := []struct {
		name          string
		body          string
		mock          mockBehavior
		statusCode    int
		expectedToken string
	}{
		{
			name: "success register",
			body: `{"login":"user","password":"pass"}`,
			mock: func(s *mock_auth.MockService, login, password string) {
				s.EXPECT().Register(gomock.Any(), login, password).Return("1", nil)
			},
			statusCode:    200,
			expectedToken: "1",
		},
		{
			name: "user exists",
			body: `{"login":"user","password":"pass"}`,
			mock: func(s *mock_auth.MockService, login, password string) {
				s.EXPECT().Register(gomock.Any(), login, password).Return("", auth.ErrUserAlreadyExists)
			},
			statusCode:    409,
			expectedToken: "",
		},
		{
			name:          "invalid json",
			body:          `{}`,
			mock:          func(s *mock_auth.MockService, login, password string) {},
			statusCode:    400,
			expectedToken: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			auth := mock_auth.NewMockService(c)
			tt.mock(auth, "user", "pass")

			handler := httptransport.NewAuthHandler(logger, auth)

			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.POST("/register", handler.Register)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/register", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.statusCode, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tt.expectedToken != "" {
				token, exists := response["token"]
				assert.True(t, exists, "token should be in response")
				assert.Equal(t, tt.expectedToken, token)
			} else {
				token, exists := response["token"]
				if exists {
					assert.Empty(t, token)
				}
			}

			if tt.statusCode >= 400 {
				errorMsg, exists := response["error"]
				if exists {
					assert.NotEmpty(t, errorMsg)
				}
			}
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {

	type mockBehavior func(s *mock_auth.MockService, login, password string)

	logger := &logger.Logger{Logger: zap.NewNop()}

	tests := []struct {
		name          string
		body          string
		mock          mockBehavior
		statusCode    int
		expectedToken string
	}{
		{
			name: "success login",
			body: `{"login":"user","password":"pass"}`,
			mock: func(s *mock_auth.MockService, login, password string) {
				s.EXPECT().Login(gomock.Any(), login, password).Return("1", nil)
			},
			statusCode:    200,
			expectedToken: "1",
		},
		{
			name: "invalid credentials",
			body: `{"login":"user","password":"pass"}`,
			mock: func(s *mock_auth.MockService, login, password string) {
				s.EXPECT().Login(gomock.Any(), login, password).Return("", auth.ErrInvalidCredentials)
			},
			statusCode:    401,
			expectedToken: "",
		},
		{
			name:          "invalid json",
			body:          `{}`,
			mock:          func(s *mock_auth.MockService, login, password string) {},
			statusCode:    400,
			expectedToken: "",
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			c := gomock.NewController(t)
			defer c.Finish()

			auth := mock_auth.NewMockService(c)
			tt.mock(auth, "user", "pass")

			handler := httptransport.NewAuthHandler(logger, auth)

			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.POST("/login", handler.Login)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/login", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.statusCode, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tt.expectedToken != "" {
				token, exists := response["token"]
				assert.True(t, exists, "token should be in response")
				assert.Equal(t, tt.expectedToken, token)
			} else {
				token, exists := response["token"]
				if exists {
					assert.Empty(t, token)
				}
			}

			if tt.statusCode >= 400 {
				errorMsg, exists := response["error"]
				if exists {
					assert.NotEmpty(t, errorMsg)
				}
			}
		})
	}
}

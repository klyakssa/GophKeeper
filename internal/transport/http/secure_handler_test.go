package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"gophkeeper/internal/domain/secure"
	mock_secure "gophkeeper/internal/domain/secure/mocks"
	"gophkeeper/internal/logger"
	client "gophkeeper/internal/transport/websocket"
)

func TestSecureHandler_CreateSecureHandler(t *testing.T) {
	type mockBehavior func(s *mock_secure.MockService, req *secure.SecureDataCreate)

	logger := &logger.Logger{Logger: zap.NewNop()}
	manager := client.NewManager(context.Background(), zap.NewNop())

	tests := []struct {
		name       string
		userID     string
		body       string
		mock       mockBehavior
		statusCode int
	}{
		{
			name:   "success create credentials",
			userID: "1",
			body: `{
				"data_type": "credentials",
				"login": "testuser",
				"password": "dGVzdHBhc3M=",
				"metadata": "test metadata"
			}`,
			mock: func(s *mock_secure.MockService, req *secure.SecureDataCreate) {
				s.EXPECT().
					CreateSecureData(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx interface{}, r *secure.SecureDataCreate) error {
						assert.Equal(t, 1, r.UserID)
						assert.Equal(t, secure.DataTypeCredentials, r.DataType)
						assert.Equal(t, "testuser", r.Login)
						return nil
					}).
					Times(1)
			},
			statusCode: http.StatusAccepted,
		},
		{
			name:   "success create card",
			userID: "2",
			body: `{
				"data_type": "card",
				"card_number": "NDU2Nzg5MDEyMzQ1Ng==",
				"card_holder": "John Doe",
				"card_expiry_month": 12,
				"card_expiry_year": 2026,
				"card_cvv": "MTIz",
				"card_type": "Visa",
				"metadata": "card metadata"
			}`,
			mock: func(s *mock_secure.MockService, req *secure.SecureDataCreate) {
				s.EXPECT().
					CreateSecureData(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx interface{}, r *secure.SecureDataCreate) error {
						assert.Equal(t, 2, r.UserID)
						assert.Equal(t, secure.DataTypeCard, r.DataType)
						assert.Equal(t, "John Doe", r.CardHolder)
						assert.Equal(t, 12, r.CardExpiryMonth)
						return nil
					}).
					Times(1)
			},
			statusCode: http.StatusAccepted,
		},
		{
			name:       "invalid json",
			userID:     "1",
			body:       `{"data_type": "credentials", "login": "test"`,
			mock:       func(s *mock_secure.MockService, req *secure.SecureDataCreate) {},
			statusCode: http.StatusBadRequest,
		},
		{
			name:   "service error",
			userID: "1",
			body: `{
				"data_type": "credentials",
				"login": "testuser",
				"password": "dGVzdHBhc3M="
			}`,
			mock: func(s *mock_secure.MockService, req *secure.SecureDataCreate) {
				s.EXPECT().
					CreateSecureData(gomock.Any(), gomock.Any()).
					Return(errors.New("database error")).
					Times(1)
			},
			statusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			service := mock_secure.NewMockService(ctrl)
			tt.mock(service, nil)

			handler := NewSecureHandler(service, logger, manager)

			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.POST("/secure", func(c *gin.Context) {
				c.Set("user_id", tt.userID)
				c.Next()
			}, handler.CreateSecureHandler)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/secure", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.statusCode, w.Code)
		})
	}
}

func TestSecureHandler_GetSecureHandler(t *testing.T) {
	type mockBehavior func(s *mock_secure.MockService, userID string)

	logger := &logger.Logger{Logger: zap.NewNop()}
	manager := client.NewManager(context.Background(), zap.NewNop())

	// Создаем тестовые данные как []*secure.SecureData для мока
	testDataPtr := []secure.SecureData{
		{
			ID:       1,
			UserID:   1,
			DataType: secure.DataTypeCredentials,
			Login:    "testuser",
			Metadata: "test metadata",
		},
		{
			ID:       2,
			UserID:   1,
			DataType: secure.DataTypeText,
			TextData: "some text",
		},
	}

	// Ожидаемые данные для проверки ответа
	expectedData := []secure.SecureData{
		{
			ID:       1,
			UserID:   1,
			DataType: secure.DataTypeCredentials,
			Login:    "testuser",
			Metadata: "test metadata",
		},
		{
			ID:       2,
			UserID:   1,
			DataType: secure.DataTypeText,
			TextData: "some text",
		},
	}

	tests := []struct {
		name         string
		userID       string
		mock         mockBehavior
		statusCode   int
		expectedData []secure.SecureData
	}{
		{
			name:   "success get data",
			userID: "1",
			mock: func(s *mock_secure.MockService, userID string) {
				s.EXPECT().
					GetSecureData(gomock.Any(), userID).
					Return(testDataPtr, nil).
					Times(1)
			},
			statusCode:   http.StatusOK,
			expectedData: expectedData,
		},
		{
			name:   "no data found",
			userID: "2",
			mock: func(s *mock_secure.MockService, userID string) {
				s.EXPECT().
					GetSecureData(gomock.Any(), userID).
					Return(nil, nil).
					Times(1)
			},
			statusCode: http.StatusNoContent,
		},
		{
			name:   "service error",
			userID: "1",
			mock: func(s *mock_secure.MockService, userID string) {
				s.EXPECT().
					GetSecureData(gomock.Any(), userID).
					Return(nil, errors.New("database error")).
					Times(1)
			},
			statusCode: http.StatusInternalServerError,
		},
		{
			name:   "empty list",
			userID: "3",
			mock: func(s *mock_secure.MockService, userID string) {
				s.EXPECT().
					GetSecureData(gomock.Any(), userID).
					Return([]secure.SecureData{}, nil).
					Times(1)
			},
			statusCode: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			service := mock_secure.NewMockService(ctrl)
			tt.mock(service, tt.userID)

			handler := NewSecureHandler(service, logger, manager)

			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.GET("/secure", func(c *gin.Context) {
				c.Set("user_id", tt.userID)
				c.Next()
			}, handler.GetSecureHandler)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/secure", nil)

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.statusCode, w.Code)

			if tt.statusCode == http.StatusOK {
				var data []secure.SecureData
				err := json.Unmarshal(w.Body.Bytes(), &data)
				require.NoError(t, err)
				assert.Len(t, data, len(tt.expectedData))
				if len(data) > 0 {
					assert.Equal(t, tt.expectedData[0].Login, data[0].Login)
				}
			}
		})
	}
}

func TestSecureHandler_DeleteSecureHandler(t *testing.T) {
	type mockBehavior func(s *mock_secure.MockService, userID string, id int)

	logger := &logger.Logger{Logger: zap.NewNop()}
	manager := client.NewManager(context.Background(), zap.NewNop())

	tests := []struct {
		name       string
		userID     string
		body       string
		mock       mockBehavior
		statusCode int
	}{
		{
			name:   "success delete",
			userID: "1",
			body:   `{"id": 123}`,
			mock: func(s *mock_secure.MockService, userID string, id int) {
				s.EXPECT().
					DeleteSecureData(gomock.Any(), userID, id).
					Return(nil).
					Times(1)
			},
			statusCode: http.StatusOK,
		},
		{
			name:       "invalid json",
			userID:     "1",
			body:       `{"id": }`,
			mock:       func(s *mock_secure.MockService, userID string, id int) {},
			statusCode: http.StatusBadRequest,
		},
		{
			name:   "service error",
			userID: "1",
			body:   `{"id": 123}`,
			mock: func(s *mock_secure.MockService, userID string, id int) {
				s.EXPECT().
					DeleteSecureData(gomock.Any(), userID, id).
					Return(errors.New("database error")).
					Times(1)
			},
			statusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			service := mock_secure.NewMockService(ctrl)

			// Правильно настраиваем мок в зависимости от теста
			if tt.name == "success delete" {
				tt.mock(service, tt.userID, 123)
			} else if tt.name == "service error" {
				tt.mock(service, tt.userID, 123)
			} else {
				tt.mock(service, tt.userID, 0)
			}

			handler := NewSecureHandler(service, logger, manager)

			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.DELETE("/secure", func(c *gin.Context) {
				c.Set("user_id", tt.userID)
				c.Next()
			}, handler.DeleteSecureHandler)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("DELETE", "/secure", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.statusCode, w.Code)

			if tt.statusCode == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "message")
				assert.Equal(t, "Order deleted successfully", response["message"])
			}
		})
	}
}

func TestSecureHandler_UpdateSecureHandler(t *testing.T) {
	type mockBehavior func(s *mock_secure.MockService, req *secure.SecureDataUpdate)

	logger := &logger.Logger{Logger: zap.NewNop()}
	manager := client.NewManager(context.Background(), zap.NewNop())

	tests := []struct {
		name       string
		userID     string
		body       string
		mock       mockBehavior
		statusCode int
	}{
		{
			name:   "success update credentials",
			userID: "1",
			body: `{
				"id": 1,
				"data_type": "credentials",
				"login": "updateduser",
				"password": "bmV3cGFzcw==",
				"metadata": "updated metadata"
			}`,
			mock: func(s *mock_secure.MockService, req *secure.SecureDataUpdate) {
				s.EXPECT().
					UpdateSecureData(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx interface{}, r *secure.SecureDataUpdate) error {
						assert.Equal(t, 1, r.UserID)
						assert.Equal(t, 1, r.ID)
						// Проверяем через указатели
						if r.DataType != nil {
							assert.Equal(t, secure.DataTypeCredentials, *r.DataType)
						}
						if r.Login != nil {
							assert.Equal(t, "updateduser", *r.Login)
						}
						return nil
					}).
					Times(1)
			},
			statusCode: http.StatusOK,
		},
		{
			name:   "success update card",
			userID: "2",
			body: `{
				"id": 2,
				"data_type": "card",
				"card_holder": "Jane Doe",
				"card_expiry_month": 10,
				"card_expiry_year": 2027,
				"card_type": "Mastercard"
			}`,
			mock: func(s *mock_secure.MockService, req *secure.SecureDataUpdate) {
				s.EXPECT().
					UpdateSecureData(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx interface{}, r *secure.SecureDataUpdate) error {
						assert.Equal(t, 2, r.UserID)
						assert.Equal(t, 2, r.ID)
						if r.DataType != nil {
							assert.Equal(t, secure.DataTypeCard, *r.DataType)
						}
						if r.CardHolder != nil {
							assert.Equal(t, "Jane Doe", *r.CardHolder)
						}
						if r.CardExpiryMonth != nil {
							assert.Equal(t, 10, *r.CardExpiryMonth)
						}
						return nil
					}).
					Times(1)
			},
			statusCode: http.StatusOK,
		},
		{
			name:       "invalid json",
			userID:     "1",
			body:       `{"id": 1, "data_type": "credentials"`,
			mock:       func(s *mock_secure.MockService, req *secure.SecureDataUpdate) {},
			statusCode: http.StatusBadRequest,
		},
		{
			name:   "service error",
			userID: "1",
			body: `{
				"id": 1,
				"data_type": "credentials",
				"login": "test"
			}`,
			mock: func(s *mock_secure.MockService, req *secure.SecureDataUpdate) {
				s.EXPECT().
					UpdateSecureData(gomock.Any(), gomock.Any()).
					Return(errors.New("database error")).
					Times(1)
			},
			statusCode: http.StatusInternalServerError,
		},
		{
			name:   "update text data",
			userID: "3",
			body: `{
				"id": 3,
				"data_type": "text",
				"text_data": "new text content"
			}`,
			mock: func(s *mock_secure.MockService, req *secure.SecureDataUpdate) {
				s.EXPECT().
					UpdateSecureData(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx interface{}, r *secure.SecureDataUpdate) error {
						assert.Equal(t, 3, r.UserID)
						assert.Equal(t, 3, r.ID)
						if r.DataType != nil {
							assert.Equal(t, secure.DataTypeText, *r.DataType)
						}
						if r.TextData != nil {
							assert.Equal(t, "new text content", *r.TextData)
						}
						return nil
					}).
					Times(1)
			},
			statusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			service := mock_secure.NewMockService(ctrl)
			tt.mock(service, nil)

			handler := NewSecureHandler(service, logger, manager)

			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.PUT("/secure", func(c *gin.Context) {
				c.Set("user_id", tt.userID)
				c.Next()
			}, handler.UpdateSecureHandler)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("PUT", "/secure", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.statusCode, w.Code)

			if tt.statusCode == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "message")
				assert.Equal(t, "Order updated successfully", response["message"])
			}
		})
	}
}

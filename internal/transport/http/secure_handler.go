package http

import (
	"gophkeeper/internal/domain/secure"
	"gophkeeper/internal/logger"
	client "gophkeeper/internal/transport/websocket"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type SecureHandler struct {
	manager *client.Manager
	service secure.Service
	log     *logger.Logger
}

func NewSecureHandler(service secure.Service, log *logger.Logger, manager *client.Manager) *SecureHandler {
	return &SecureHandler{
		service: service,
		log:     log,
		manager: manager,
	}
}

func (s *SecureHandler) CreateSecureHandler(c *gin.Context) {
	var req secure.SecureDataCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		s.log.Error("Failed to create secure data", zap.Error(err))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	s.log.Debug("CreateSecureHandler", zap.Any("body", req))

	id, _ := strconv.Atoi(c.GetString("user_id"))
	req.UserID = id

	err := s.service.CreateSecureData(c.Request.Context(), &req)
	if err != nil {
		s.log.Error("Failed to create secure data", zap.Error(err))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	go s.manager.BroadcastMessage(c.GetString("user_id"))

	c.JSON(http.StatusAccepted, gin.H{"message": "Order added successfully"})
}

func (s *SecureHandler) GetSecureHandler(c *gin.Context) {

	s.log.Debug("GetSecureHandler")

	data, err := s.service.GetSecureData(c.Request.Context(), c.GetString("user_id"))
	if err != nil {
		s.log.Error("Failed to get secure data", zap.Error(err))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if data == nil {
		c.JSON(http.StatusNoContent, gin.H{"data": nil})
		return
	}

	c.JSON(http.StatusOK, data)
}

type SecureDataDelete struct {
	ID int `json:"id"`
}

func (s *SecureHandler) DeleteSecureHandler(c *gin.Context) {
	var req SecureDataDelete
	if err := c.ShouldBindJSON(&req); err != nil {
		s.log.Error("Failed to delete secure data", zap.Error(err))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	s.log.Debug("DeleteSecureHandler", zap.Any("body", req))

	err := s.service.DeleteSecureData(c.Request.Context(), c.GetString("user_id"), req.ID)
	if err != nil {
		s.log.Error("Failed to delete secure data", zap.Error(err))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order deleted successfully"})
}

func (s *SecureHandler) UpdateSecureHandler(c *gin.Context) {
	var req secure.SecureDataUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		s.log.Error("Failed to update secure data", zap.Error(err))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	id, _ := strconv.Atoi(c.GetString("user_id"))
	req.UserID = id

	s.log.Debug("UpdateSecureHandler", zap.Any("body", req))

	err := s.service.UpdateSecureData(c.Request.Context(), &req)
	if err != nil {
		s.log.Error("Failed to update secure data", zap.Error(err))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	go s.manager.BroadcastMessage(c.GetString("user_id"))

	c.JSON(http.StatusOK, gin.H{"message": "Order updated successfully"})
}

package http

import (
	"gophkeeper/internal/domain/auth"
	"gophkeeper/internal/logger"
	client "gophkeeper/internal/transport/websocket"
	"gophkeeper/internal/utils"
	"net/http"

	"github.com/fasthttp/websocket"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type WebSocketHandler struct {
	manager *client.Manager
	service auth.Service
	log     *logger.Logger
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewWebSocketHandler(log *logger.Logger, service auth.Service, manager *client.Manager) *WebSocketHandler {
	return &WebSocketHandler{log: log, service: service, manager: manager}
}

func (w *WebSocketHandler) WebSocketHandler(c *gin.Context) {

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		w.log.Error("Failed to upgrade connection", zap.Error(err))
		return
	}

	user_id, err := utils.GetUserIDFromContextString(c)
	if err != nil {
		w.log.Error("Failed to get user id", zap.Error(err))
		conn.WriteMessage(websocket.CloseMessage, nil)
		conn.Close()
		return
	}

	user, err := w.service.GetUserByID(c.Request.Context(), user_id)
	if err != nil {
		w.log.Error("Failed to get user", zap.Error(err))
		conn.WriteMessage(websocket.CloseMessage, nil)
		conn.Close()
		return
	}

	cl := client.NewClient(conn, w.log, user, w.manager)
	err = w.manager.AddClient(cl)
	if err != nil {
		w.log.Error("Failed to add client", zap.Error(err))
		conn.WriteMessage(websocket.CloseMessage, nil)
		conn.Close()
		return
	}

	go cl.WriteMessages()
	go cl.ReadMessages()

}

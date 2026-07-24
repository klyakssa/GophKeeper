package http

import (
	"gophkeeper/internal/domain/auth"
	"gophkeeper/internal/logger"
	client "gophkeeper/internal/transport/websocket"
	"log"
	"net/http"

	"github.com/fasthttp/websocket"
	"github.com/gin-gonic/gin"
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
		log.Print("Failed to upgrade connection:", err)
		return
	}

	user, err := w.service.GetUserByID(c.Request.Context(), c.GetString("user_id"))
	if err != nil {
		log.Println(err)
		conn.WriteMessage(websocket.CloseMessage, nil)
		conn.Close()
		return
	}

	cl := client.NewClient(conn, w.log, user, w.manager)
	err = w.manager.AddClient(cl)
	if err != nil {
		log.Println(err)
		conn.WriteMessage(websocket.CloseMessage, nil)
		conn.Close()
		return
	}

	go cl.WriteMessages()
	go cl.ReadMessages()

}

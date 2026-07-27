package client

import (
	"encoding/json"
	"gophkeeper/internal/domain/auth"
	"gophkeeper/internal/logger"
	"time"

	"github.com/fasthttp/websocket"
	guuid "github.com/google/uuid"
	"go.uber.org/zap"
)

// ClientList is a map used to help manage a map of clients
// type ClientList map[*Client]bool
type ClientList map[string]map[*Client]bool

// type ClientList map[*Client]map[string]bool

// Client is a websocket client, basically a frontend visitor
type Client struct {
	// the websocket connection
	connection *websocket.Conn

	uuid string

	done chan *Client

	log *logger.Logger

	userInfo *auth.User

	// manager is the manager used to manage the client
	manager *Manager
	// egress is used to avoid concurrent writes on the WebSocket
	egress chan Event
}

var (
	// pongWait is how long we will await a pong response from client
	pongWait = 10 * time.Second
	// pingInterval has to be less than pongWait, We cant multiply by 0.9 to get 90% of time
	// Because that can make decimals, so instead *9 / 10 to get 90%
	// The reason why it has to be less than PingRequency is becuase otherwise it will send a new Ping before getting response
	pingInterval = (pongWait * 9) / 10
)

// NewClient is used to initialize a new Client with all required values initialized
func NewClient(conn *websocket.Conn, log *logger.Logger, userInfo *auth.User, manager *Manager) *Client {
	return &Client{
		connection: conn,
		uuid:       genUUID(),
		log:        log,
		userInfo:   userInfo,
		manager:    manager,
		done:       make(chan *Client),
		egress:     make(chan Event),
	}
}

func (c *Client) ReadMessages() {
	defer func() {
		// Graceful Close the Connection once this
		// function is done
		c.done <- c
		c.manager.RemoveClient(c)
	}()
	if err := c.connection.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		c.log.Warn("error setting read deadline", zap.Error(err))
		return
	}
	c.connection.SetReadLimit(512)
	c.connection.SetPongHandler(c.pongHandler)
	c.connection.SetCloseHandler(c.closeHandler)
	for {
		// ReadMessage is used to read the next message in queue
		// in the connection
		_, payload, err := c.connection.ReadMessage()

		if err != nil {
			// If Connection is closed, we will Recieve an error here
			// We only want to log Strange errors, but simple Disconnection
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				break
			}
			break // Break the loop to close conn & Cleanup
		}
		c.log.Info("message received: ", zap.String("message", string(payload)))
	}
}

// writeMessages is a process that listens for new messages to output to the Client
func (c *Client) WriteMessages() {
	// Create a ticker that triggers a ping at given interval
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		c.manager.RemoveClient(c)
	}()

	for {
		select {
		case message, ok := <-c.egress:
			// Ok will be false Incase the egress channel is closed
			if !ok {
				// Manager has closed this connection channel, so communicate that to frontend
				if err := c.connection.WriteMessage(websocket.CloseMessage, nil); err != nil {
					// Log that the connection is closed and the reason
					c.log.Warn("error writing message", zap.Error(err))
				}
				// Return to close the goroutine
				return
			}

			data, err := json.Marshal(message)
			if err != nil {
				c.log.Warn("error marshalling message", zap.Error(err))
				return // closes the connection, should we really
			}
			// Write a Regular text message to the connection
			if err := c.connection.WriteMessage(websocket.TextMessage, data); err != nil {
				c.log.Warn("error writing message", zap.Error(err))
			}
			c.log.Info("message sent")
		case <-c.done:
			// Close the connection
			c.log.Info("Closing connection", zap.String("username", c.userInfo.Login))
			return
		case <-ticker.C:
			// Send a Ping
			// Send the Ping
			if err := c.connection.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				c.log.Warn("error sending ping", zap.Error(err))
				return // return to break this goroutine triggeing cleanup
			}
		}

	}
}

func (c *Client) Close() error {
	err := c.connection.WriteMessage(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, "closing"))
	if err != nil {
		c.log.Debug("error writing close message", zap.Error(err))
	}
	return c.connection.Close()
}

func (c *Client) pongHandler(pongMsg string) error {
	// Current time + Pong Wait time
	c.log.Debug("pong received", zap.String("uuid", c.uuid), zap.String("pongMsg", pongMsg), zap.String("login", c.userInfo.Login))
	return c.connection.SetReadDeadline(time.Now().Add(pongWait))
}

func (c *Client) closeHandler(code int, text string) error {
	c.log.Debug("close received", zap.String("uuid", c.uuid), zap.Int("code", code), zap.String("text", text), zap.String("login", c.userInfo.Login))
	c.done <- c
	return c.connection.Close()
}

func genUUID() string {
	id := guuid.New()
	//log.Printf("github.com/google/uuid:         %s\n", id.String())
	return id.String()
}

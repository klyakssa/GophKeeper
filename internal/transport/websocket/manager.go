package client

import (
	"context"
	"gophkeeper/internal/domain/auth"
	"sync"

	"go.uber.org/zap"
)

// Manager is used to hold references to all Clients Registered, and Broadcasting etc
type Manager struct {
	clients ClientList
	log     *zap.Logger

	// Using a syncMutex here to be able to lcok state before editing clients
	// Could also use Channels to block
	sync.RWMutex
	// handlers are functions that are used to handle Events
	handlers map[string]EventHandler
}

// NewManager is used to initalize all the values inside the manager
func NewManager(ctx context.Context, log *zap.Logger) *Manager {
	m := &Manager{
		clients:  make(ClientList),
		log:      log,
		handlers: make(map[string]EventHandler),
	}
	m.setupEventHandlers()
	return m
}

// setupEventHandlers configures and adds all handlers
func (m *Manager) setupEventHandlers() {
	m.handlers[EventSendMessage] = SendMessageHandler
}

// routeEvent is used to make sure the correct event goes into the correct handler
func (m *Manager) routeEvent(event Event, c *Client) error {
	if _, ok := m.clients[c.userInfo.Login]; ok {
		if handler, ok := m.handlers[event.Type]; ok {
			// Execute the handler and return any err
			if err := handler(event, c); err != nil {
				return err
			}
			return nil
		} else {
			return auth.ErrEventNotSupported
		}
	} else {
		return auth.ErrClientNotConnected
	}
}

// addClient will add clients to our clientList
func (m *Manager) AddClient(client *Client) error { //type ClientList map[*models.Client]map[string]*Client
	m.Lock()
	defer m.Unlock()
	if _, ok := m.clients[client.userInfo.ID][client]; ok {
		return auth.ErrClientAlreadyConnected
	}

	// Check if Client exists
	if _, ok := m.clients[client.userInfo.ID]; !ok {
		m.clients[client.userInfo.ID] = make(map[*Client]bool)
		m.clients[client.userInfo.ID][client] = true
	} else {
		m.clients[client.userInfo.ID][client] = true
	}

	return nil
}

// removeClient will remove the client and clean up
func (m *Manager) RemoveClient(client *Client) {
	m.Lock()
	defer m.Unlock()

	// Check if Client exists, then delete it
	if _, ok := m.clients[client.userInfo.ID][client]; ok {
		// close connection
		client.connection.Close()
		// remove
		delete(m.clients[client.userInfo.ID], client)
	}

	if len(m.clients[client.userInfo.ID]) == 0 {
		delete(m.clients, client.userInfo.ID)
	}
}

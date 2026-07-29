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

	closeOnce sync.Once
}

// NewManager is used to initalize all the values inside the manager
func NewManager(ctx context.Context, log *zap.Logger) *Manager {
	return &Manager{
		clients:  make(ClientList),
		log:      log,
		handlers: make(map[string]EventHandler),
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

// BroadcastMessage will send the message to all clients in the chatroom
func (m *Manager) BroadcastMessage(userID string) {
	m.RLock()
	defer m.RUnlock()

	for client := range m.clients[userID] {
		client.egress <- Event{
			Type: EventSendMessageToUpdate,
		}
	}
}

func (m *Manager) Close() error {
	var err error

	m.closeOnce.Do(func() {
		m.Lock()
		defer m.Unlock()
		for _, clients := range m.clients {
			for client := range clients {
				if errC := client.Close(); err != nil {
					m.log.Warn("error closing client", zap.Error(err))
					err = errC
				}
			}
		}
	})

	return err
}

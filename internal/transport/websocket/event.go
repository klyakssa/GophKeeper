package client

import (
	"encoding/json"
	"fmt"
	"time"
)

// Event is the Messages sent over the websocket
// Used to differ between different actions
type Event struct {
	// Type is the message type sent
	Type string `json:"type"`
	// Payload is the data Based on the Type
	Payload json.RawMessage `json:"payload"`
}

// EventHandler is a function signature that is used to affect messages on the socket and triggered
// depending on the type
type EventHandler func(event Event, c *Client) error

const (
	// EventSendNewToken is the event name for new tokens
	EventSendNewToken = "send_new_token"
	// EventSendMessage is the event name for new chat messages sent
	EventSendMessage = "send_message"
	// EventNewMessage is a response to send_message
	EventNewMessage = "new_message"
)

// SendMessageEvent is the payload sent in the
// send_message event
type SendMessageEvent struct {
	Message string `json:"message"`
	From    string `json:"from"`
}

// NewMessageEvent is returned when responding to send_message
type NewMessageEvent struct {
	SendMessageEvent
	Sent string `json:"sent"`
}

// SendMessageHandler will send out a message to all other participants in the chat
func SendMessageHandler(event Event, c *Client) error {
	// Marshal Payload into wanted format
	var chatevent SendMessageEvent
	if err := json.Unmarshal(event.Payload, &chatevent); err != nil {
		return fmt.Errorf("bad payload in request: %v", err)
	}

	// Prepare an Outgoing Message to others
	var broadMessage NewMessageEvent

	broadMessage.Sent = time.Now().Format("2006-01-02 15:04:05")
	broadMessage.Message = chatevent.Message
	broadMessage.From = c.userInfo.Login

	data, err := json.Marshal(broadMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal broadcast message: %v", err)
	}

	// Place payload into an Event
	var outgoingEvent Event
	outgoingEvent.Payload = data
	outgoingEvent.Type = EventNewMessage
	// Broadcast to all other Clients
	// for client := range c.manager.clients {
	// 	// Only send to clients inside the same chatroom
	// 	if client.chatroom == c.chatroom {
	// 		client.egress <- outgoingEvent
	// 	}

	// }
	c.egress <- outgoingEvent
	return nil
}

func SendNewTokenHandler(event Event, c *Client) error {
	var outgoingEvent Event
	outgoingEvent.Payload = event.Payload
	outgoingEvent.Type = EventNewMessage
	c.egress <- outgoingEvent
	return nil
}

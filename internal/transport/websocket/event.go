package client

// Event is the Messages sent over the websocket
// Used to differ between different actions
type Event struct {
	// Type is the message type sent
	Type string `json:"type"`
}

// EventHandler is a function signature that is used to affect messages on the socket and triggered
// depending on the type
type EventHandler func(event Event, c *Client) error

const (
	// EventSendMessage is the event name for new chat messages sent
	EventSendMessageToUpdate = "send_message_to_update"
)

// SendMessageHandler will send out a message to all other participants
func SendMessageToUpdateHandler(c *Client) {
	c.egress <- Event{
		Type: EventSendMessageToUpdate,
	}
}

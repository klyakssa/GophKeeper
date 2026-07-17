package pas_client

import "errors"

var (
	ErrEventNotSupported  = errors.New("this event type is not supported")
	ErrClientNotConnected = errors.New("client is not connected")
)

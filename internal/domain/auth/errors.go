package auth

import "errors"

var (
	ErrUserAlreadyExists      = errors.New("user already exists") // return when user already exists
	ErrInvalidCredentials     = errors.New("invalid login or password")
	ErrUserNotFound           = errors.New("user not found")
	ErrPasswordTooLong        = errors.New("password is too long")
	ErrLoginTooLong           = errors.New("login is too long")
	ErrEventNotSupported      = errors.New("this event type is not supported")
	ErrClientNotConnected     = errors.New("client is not connected")
	ErrClientAlreadyConnected = errors.New("client is already connected")
)

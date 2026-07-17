package auth

import "errors"

var (
	ErrUserAlreadyExists  = errors.New("user already exists")       // return when user already exists
	ErrInvalidCredentials = errors.New("invalid login or password") // return when user already exists
	ErrUserNotFound       = errors.New("user not found")            // return when user already exists
	ErrPasswordTooLong    = errors.New("password is too long")      // return when user already exists
	ErrLoginTooLong       = errors.New("login is too long")         // return when user already exists
)

package utils

import (
	"context"
	"errors"
)

type contextKey string

const (
	UserIDKey contextKey = "user_id"
)

func GetUserIDFromContextString(ctx context.Context) (string, error) {
	val := ctx.Value(UserIDKey)
	if val == nil {
		return "", errors.New("user_id not found in context")
	}

	userID, ok := val.(string)
	if !ok {
		return "", errors.New("user_id not found in context")
	}

	return userID, nil
}

func GetUserIDFromContextInt(ctx context.Context) (int, error) {
	val := ctx.Value(UserIDKey)
	if val == nil {
		return 0, errors.New("user_id not found in context")
	}

	userID, ok := val.(int)
	if !ok {
		return 0, errors.New("user_id not found in context")
	}

	return userID, nil
}

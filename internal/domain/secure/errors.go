package secure

import "errors"

var (
	ErrSecureDataInsert = errors.New("failed to insert secure data")
)

func IsSecureDataInsertError(err error) bool {
	return errors.Is(err, ErrSecureDataInsert)
}

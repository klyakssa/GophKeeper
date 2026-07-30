package secure

import (
	"time"
)

type DataType string

const (
	DataTypeCredentials DataType = "credentials"
	DataTypeText        DataType = "text"
	DataTypeBinary      DataType = "binary"
	DataTypeCard        DataType = "card"
)

type SecureData struct {
	UserID              int        `db:"user_id" json:"-"`
	ID                  int        `db:"id" json:"id"`
	DataType            DataType   `db:"data_type" json:"data_type"`
	Login               string     `db:"login" json:"login,omitempty"`
	PasswordEncrypted   []byte     `db:"password_encrypted" json:"password_encrypted,omitempty"`
	TextData            string     `db:"text_data" json:"text_data,omitempty"`
	BinaryData          []byte     `db:"binary_data" json:"binary_data,omitempty"`
	BinaryMimeType      string     `db:"binary_mime_type" json:"binary_mime_type,omitempty"`
	CardNumberEncrypted []byte     `db:"card_number_encrypted" json:"card_number_encrypted,omitempty"`
	CardHolder          string     `db:"card_holder" json:"card_holder,omitempty"`
	CardExpiryMonth     int        `db:"card_expiry_month" json:"card_expiry_month,omitempty"`
	CardExpiryYear      int        `db:"card_expiry_year" json:"card_expiry_year,omitempty"`
	CardCvvEncrypted    []byte     `db:"card_cvv_encrypted" json:"card_cvv_encrypted,omitempty"`
	CardType            string     `db:"card_type" json:"card_type,omitempty"`
	Metadata            string     `db:"metadata" json:"metadata"`
	CreatedAt           time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt           *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}

type SecureDataCreate struct {
	UserID          int      `db:"user_id"`
	DataType        DataType `json:"data_type" db:"data_type" validate:"required,oneof=credentials text binary card"`
	Login           string   `json:"login,omitempty" db:"login"`
	Password        []byte   `json:"password,omitempty" db:"password_encrypted"`
	TextData        string   `json:"text_data,omitempty" db:"text_data"`
	BinaryData      []byte   `json:"binary_data,omitempty" db:"binary_data"`
	BinaryMimeType  string   `json:"binary_mime_type,omitempty" db:"binary_mime_type"`
	CardNumber      []byte   `json:"card_number,omitempty" db:"card_number_encrypted"`
	CardHolder      string   `json:"card_holder,omitempty" db:"card_holder"`
	CardExpiryMonth int      `json:"card_expiry_month,omitempty" db:"card_expiry_month"`
	CardExpiryYear  int      `json:"card_expiry_year,omitempty" db:"card_expiry_year"`
	CardCvv         []byte   `json:"card_cvv,omitempty" db:"card_cvv_encrypted"`
	CardType        string   `json:"card_type,omitempty" db:"card_type"`
	Metadata        string   `json:"metadata" db:"metadata"`
}

type SecureDataUpdate struct {
	UserID              int       `db:"user_id"`
	ID                  int       `json:"id"`
	DataType            *DataType `json:"data_type,omitempty"`
	Login               *string   `json:"login,omitempty"`
	PasswordEncrypted   []byte    `json:"password_encrypted,omitempty"`
	TextData            *string   `json:"text_data,omitempty"`
	BinaryData          []byte    `json:"binary_data,omitempty"`
	BinaryMimeType      *string   `json:"binary_mime_type,omitempty"`
	CardNumberEncrypted []byte    `json:"card_number_encrypted,omitempty"`
	CardHolder          *string   `json:"card_holder,omitempty"`
	CardExpiryMonth     *int      `json:"card_expiry_month,omitempty"`
	CardExpiryYear      *int      `json:"card_expiry_year,omitempty"`
	CardCvvEncrypted    []byte    `json:"card_cvv_encrypted,omitempty"`
	CardType            *string   `json:"card_type,omitempty"`
	Metadata            *string   `json:"metadata,omitempty"`
}

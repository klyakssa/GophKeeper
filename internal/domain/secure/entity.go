package secure

import (
	"database/sql"
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
	ID                  int64          `db:"id" json:"id"`
	DataType            DataType       `db:"data_type" json:"data_type"`
	Login               sql.NullString `db:"login" json:"login,omitempty"`
	PasswordEncrypted   []byte         `db:"password_encrypted" json:"password_encrypted,omitempty"`
	TextData            sql.NullString `db:"text_data" json:"text_data,omitempty"`
	BinaryData          []byte         `db:"binary_data" json:"binary_data,omitempty"`
	BinaryMimeType      sql.NullString `db:"binary_mime_type" json:"binary_mime_type,omitempty"`
	CardNumberEncrypted []byte         `db:"card_number_encrypted" json:"card_number_encrypted,omitempty"`
	CardHolder          sql.NullString `db:"card_holder" json:"card_holder,omitempty"`
	CardExpiryMonth     sql.NullInt16  `db:"card_expiry_month" json:"card_expiry_month,omitempty"`
	CardExpiryYear      sql.NullInt16  `db:"card_expiry_year" json:"card_expiry_year,omitempty"`
	CardCvvEncrypted    []byte         `db:"card_cvv_encrypted" json:"card_cvv_encrypted,omitempty"`
	CardType            sql.NullString `db:"card_type" json:"card_type,omitempty"`
	Metadata            sql.NullString `db:"metadata" json:"metadata"`
	CreatedAt           time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt           sql.NullTime   `db:"deleted_at" json:"deleted_at,omitempty"`
}

type SecureDataCreate struct {
	DataType        DataType `json:"data_type" validate:"required,oneof=credentials text binary card"`
	Login           string   `json:"login,omitempty"`
	Password        []byte   `json:"password,omitempty"`
	TextData        string   `json:"text_data,omitempty"`
	BinaryData      []byte   `json:"binary_data,omitempty"`
	BinaryMimeType  string   `json:"binary_mime_type,omitempty"`
	CardNumber      []byte   `json:"card_number,omitempty"`
	CardHolder      string   `json:"card_holder,omitempty"`
	CardExpiryMonth int16    `json:"card_expiry_month,omitempty"`
	CardExpiryYear  int16    `json:"card_expiry_year,omitempty"`
	CardCvv         []byte   `json:"card_cvv,omitempty"`
	CardType        string   `json:"card_type,omitempty"`
	Metadata        string   `json:"metadata"`
}

type SecureDataUpdate struct {
	ID                  int64     `json:"id"`
	DataType            *DataType `json:"data_type,omitempty"`
	Login               *string   `json:"login,omitempty"`
	PasswordEncrypted   []byte    `json:"password_encrypted,omitempty"`
	TextData            *string   `json:"text_data,omitempty"`
	BinaryData          []byte    `json:"binary_data,omitempty"`
	BinaryMimeType      *string   `json:"binary_mime_type,omitempty"`
	CardNumberEncrypted []byte    `json:"card_number_encrypted,omitempty"`
	CardHolder          *string   `json:"card_holder,omitempty"`
	CardExpiryMonth     *int16    `json:"card_expiry_month,omitempty"`
	CardExpiryYear      *int16    `json:"card_expiry_year,omitempty"`
	CardCvvEncrypted    []byte    `json:"card_cvv_encrypted,omitempty"`
	CardType            *string   `json:"card_type,omitempty"`
	Metadata            *string   `json:"metadata,omitempty"`
}

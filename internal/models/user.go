package models

import "errors"

type User struct {
	ID           int64
	Login        string
	PasswordHash []byte
	CreatedAt    int64
}

var (
	ErrUserAlreadyExists  = errors.New("user with this login already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

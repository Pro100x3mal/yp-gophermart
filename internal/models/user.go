package models

import (
	"errors"
	"time"
)

type User struct {
	ID           int64     `json:"-"`
	Login        string    `json:"login"`
	PasswordHash []byte    `json:"-"`
	CreatedAt    time.Time `json:"-"`
}

var (
	ErrUserAlreadyExists  = errors.New("user with this login already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

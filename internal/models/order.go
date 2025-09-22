package models

import (
	"errors"
	"time"
)

type Order struct {
	ID         int64     `json:"-"`
	UserID     int64     `json:"-"`
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    float64   `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}

var (
	ErrOrderExists                    = errors.New("order already exists")
	ErrOrderNotFound                  = errors.New("order not found")
	ErrOrderBelongsToAnotherUser      = errors.New("order belongs to another user")
	ErrOrderAlreadyUploadedBySameUser = errors.New("order already uploaded by same user")
)

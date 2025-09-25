package models

import (
	"errors"
	"time"
)

type Creds struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type User struct {
	ID           int64     `json:"-"`
	Login        string    `json:"login"`
	PasswordHash []byte    `json:"-"`
	CreatedAt    time.Time `json:"-"`
}

type Order struct {
	ID         int64     `json:"-"`
	UserID     int64     `json:"-"`
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    float64   `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type AccrualResp struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}

type Withdrawal struct {
	Order       string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type WithdrawReq struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

const (
	StatusRegistered = "REGISTERED"
	StatusInvalid    = "INVALID"
	StatusProcessing = "PROCESSING"
	StatusProcessed  = "PROCESSED"
	StatusNew        = "NEW"
)

var (
	ErrUserAlreadyExists      = errors.New("user with this login already exists")
	ErrUserNotFound           = errors.New("user not found")
	ErrUserInvalidCredentials = errors.New("invalid credentials")

	ErrOrderExists                    = errors.New("order already exists")
	ErrOrderNotFound                  = errors.New("order not found")
	ErrOrderBelongsToAnotherUser      = errors.New("order belongs to another user")
	ErrOrderAlreadyUploadedBySameUser = errors.New("order already uploaded by same user")

	ErrAccrualOrderNotRegistered = errors.New("order not registered in accrual system")
	ErrAccrualOrderTooMany       = errors.New("too many requests to accrual system")

	ErrWithdrawalOrderExists = errors.New("order already exists")
	ErrPaymentRequired       = errors.New("payment required")
)

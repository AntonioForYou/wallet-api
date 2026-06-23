package domain

import (
	"context"

	"github.com/google/uuid"
)

type OperationType string

const (
	Deposit  OperationType = "DEPOSIT"
	Withdraw OperationType = "WITHDRAW"
)

type WalletRepository interface {
	GetBalance(ctx context.Context, walletID uuid.UUID) (int64, error)
	UpdateBalance(ctx context.Context, walletID uuid.UUID, amount int64) (int64, error)
}

type Wallet struct {
	ID      uuid.UUID `json:"walletId"`
	Balance int64     `json:"balance"`
}

type WalletRequest struct {
	WalletID      uuid.UUID     `json:"walletId" binding:"required"`
	OperationType OperationType `json:"operationType" binding:"required,oneof=DEPOSIT WITHDRAW"`
	Amount        int64         `json:"amount" binding:"required,gt=0"`
}

type Job struct {
	Ctx           context.Context
	WalletID      uuid.UUID
	OperationType OperationType
	Amount        int64
	ResultChan    chan Result
}

type Result struct {
	NewBalance int64
	Err        error
}

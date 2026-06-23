package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrInsufficientFunds = errors.New("insufficient funds")

var ErrWalletNotFound = errors.New("wallet not found")

type WalletRepo struct {
	pool *pgxpool.Pool
}

func NewWalletRepo(pool *pgxpool.Pool) *WalletRepo {
	return &WalletRepo{
		pool: pool,
	}
}

func (r *WalletRepo) GetBalance(ctx context.Context, walletID uuid.UUID) (int64, error) {
	var balance int64
	query := `SELECT balance FROM wallets WHERE id = $1`

	err := r.pool.QueryRow(ctx, query, walletID).Scan(&balance)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrWalletNotFound
		}
		return 0, fmt.Errorf("failed to get balance: %w", err)
	}

	return balance, nil
}

func (r *WalletRepo) UpdateBalance(ctx context.Context, walletID uuid.UUID, amount int64) (int64, error) {
	var newBalance int64

	query := `
		UPDATE wallets 
		SET balance = balance + $2, updated_at = NOW() 
		WHERE id = $1 AND (balance + $2 >= 0)
		RETURNING balance
	`

	err := r.pool.QueryRow(ctx, query, walletID, amount).Scan(&newBalance)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			_, checkErr := r.GetBalance(ctx, walletID)
			if errors.Is(checkErr, ErrWalletNotFound) {
				return 0, ErrWalletNotFound
			}
			return 0, ErrInsufficientFunds
		}
		return 0, fmt.Errorf("failed to update balance: %w", err)
	}

	return newBalance, nil
}

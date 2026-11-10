package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrWalletNotFound    = errors.New("кошелек не найден")
	ErrInsufficientFunds = errors.New("недостаточно средств")
)

type WalletRepository struct {
	pool *pgxpool.Pool
}

func NewWalletRepository(pool *pgxpool.Pool) *WalletRepository {
	return &WalletRepository{pool: pool}
}

func (r *WalletRepository) Deposit(ctx context.Context, id string, amount int64) (int64, error) {
	var balance int64
	err := r.pool.QueryRow(ctx,
		`UPDATE wallets SET balance = balance + $2, updated_at = NOW()
		 WHERE wallet_id = $1 RETURNING balance`, id, amount,
	).Scan(&balance)

	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrWalletNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("ошибка базы: %w", err)
	}
	return balance, nil
}

func (r *WalletRepository) Withdraw(ctx context.Context, id string, amount int64) (int64, error) {
	var balance int64
	err := r.pool.QueryRow(ctx,
		`UPDATE wallets SET balance = balance - $2, updated_at = NOW()
		 WHERE wallet_id = $1 AND balance >= $2 RETURNING balance`, id, amount,
	).Scan(&balance)

	if err == nil {
		return balance, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return 0, fmt.Errorf("ошибка базы: %w", err)
	}

	var exists bool
	if err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM wallets WHERE wallet_id = $1)`, id,
	).Scan(&exists); err != nil {
		return 0, fmt.Errorf("ошибка базы: %w", err)
	}
	if !exists {
		return 0, ErrWalletNotFound
	}
	return 0, ErrInsufficientFunds
}

func (r *WalletRepository) GetBalance(ctx context.Context, walletId string) (int64, error) {
	var balance int64
	err := r.pool.QueryRow(ctx, "SELECT balance FROM wallets WHERE wallet_id = $1", walletId).Scan(&balance)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrWalletNotFound
		}

		return 0, fmt.Errorf("ошибка базы: %w", err)
	}
	return balance, nil
}

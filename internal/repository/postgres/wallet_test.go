package postgres

import (
	"context"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := "postgres://postgres:password@localhost:5432/wallet_test_db?sslmode=disable"

	cfg, err := pgxpool.ParseConfig(dsn)
	require.NoError(t, err)
	cfg.MaxConns = 30

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	require.NoError(t, err, "не удалось подключиться к БД")
	t.Cleanup(pool.Close)
	return pool
}

func newWallet(t *testing.T, pool *pgxpool.Pool, balance int64) string {
	t.Helper()

	id := uuid.NewString()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO wallets (wallet_id, balance) VALUES ($1, $2)`, id, balance)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM wallets WHERE wallet_id = $1`, id)
	})
	return id
}

func requireBalance(t *testing.T, repo *WalletRepository, id string, want int64) {
	t.Helper()
	got, err := repo.GetBalance(context.Background(), id)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestWalletRepository_GetBalance(t *testing.T) {
	pool := newPool(t)
	repo := NewWalletRepository(pool)
	ctx := context.Background()

	t.Run("успешное получение", func(t *testing.T) {
		id := newWallet(t, pool, 100)

		balance, err := repo.GetBalance(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, int64(100), balance)
	})

	t.Run("кошелек не найден", func(t *testing.T) {
		_, err := repo.GetBalance(ctx, uuid.NewString())
		assert.ErrorIs(t, err, ErrWalletNotFound)
	})
}

func TestWalletRepository_Deposit(t *testing.T) {
	pool := newPool(t)
	repo := NewWalletRepository(pool)
	ctx := context.Background()

	t.Run("успешное пополнение", func(t *testing.T) {
		id := newWallet(t, pool, 100)

		balance, err := repo.Deposit(ctx, id, 50)
		require.NoError(t, err)
		assert.Equal(t, int64(150), balance)
		requireBalance(t, repo, id, 150)
	})

	t.Run("несуществующий кошелек", func(t *testing.T) {
		_, err := repo.Deposit(ctx, uuid.NewString(), 50)
		assert.ErrorIs(t, err, ErrWalletNotFound)
	})
}

func TestWalletRepository_Withdraw(t *testing.T) {
	pool := newPool(t)
	repo := NewWalletRepository(pool)
	ctx := context.Background()

	t.Run("успешное снятие", func(t *testing.T) {
		id := newWallet(t, pool, 100)

		balance, err := repo.Withdraw(ctx, id, 30)
		require.NoError(t, err)
		assert.Equal(t, int64(70), balance)
		requireBalance(t, repo, id, 70)
	})

	t.Run("недостаточно средств", func(t *testing.T) {
		id := newWallet(t, pool, 100)

		_, err := repo.Withdraw(ctx, id, 101)
		assert.ErrorIs(t, err, ErrInsufficientFunds)
		requireBalance(t, repo, id, 100) // баланс не изменился
	})

	t.Run("несуществующий кошелек", func(t *testing.T) {
		_, err := repo.Withdraw(ctx, uuid.NewString(), 50)
		assert.ErrorIs(t, err, ErrWalletNotFound)
	})
}

// 1000 одновременных пополнений
const workers = 1000

func TestWalletRepository_ConcurrentDeposit(t *testing.T) {
	pool := newPool(t)
	repo := NewWalletRepository(pool)
	id := newWallet(t, pool, 0)

	errs := runConcurrently(workers, func(int) error {
		_, err := repo.Deposit(context.Background(), id, 1)
		return err
	})

	require.Empty(t, errs, "неожиданные ошибки при пополнении")
	requireBalance(t, repo, id, workers)
}

func runConcurrently(n int, fn func(i int) error) []error {
	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		errs []error
	)
	start := make(chan struct{})

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			if err := fn(i); err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
		}(i)
	}
	close(start)
	wg.Wait()
	return errs
}

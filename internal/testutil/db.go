package testutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool подключается к тестовой БД и применяет схему.
// Без TEST_DB_DSN тест пропускается.
func NewPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"))

	ctx := context.Background()
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatalf("parse dsn: %v", err)
	}
	// маленький пул: пакеты тестируются параллельно, а у Postgres лимит 100 соединений
	cfg.MaxConns = 20

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)

	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping: %v", err)
	}
	applySchema(t, pool)
	return pool
}

func applySchema(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()

	_, file, _, _ := runtime.Caller(0)
	path := filepath.Join(filepath.Dir(file), "..", "..", "init-db", "01_create_tables.sql")
	schema, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read schema: %v", err)
	}

	conn, err := pool.Acquire(ctx)
	if err != nil {
		t.Fatalf("acquire: %v", err)
	}
	defer conn.Release()

	// пакеты тестов идут параллельно, а CREATE TABLE IF NOT EXISTS
	// не защищён от гонки, поэтому берём advisory-lock
	if _, err := conn.Exec(ctx, `SELECT pg_advisory_lock(424242)`); err != nil {
		t.Fatalf("lock: %v", err)
	}
	defer conn.Exec(ctx, `SELECT pg_advisory_unlock(424242)`)

	if _, err := conn.Exec(ctx, string(schema)); err != nil {
		t.Fatalf("apply schema: %v", err)
	}
}

// CreateWallet создаёт кошелёк со случайным UUID и удаляет его после теста.
func CreateWallet(t *testing.T, pool *pgxpool.Pool, balance int64) string {
	t.Helper()
	ctx := context.Background()

	id := uuid.NewString()
	if _, err := pool.Exec(ctx,
		`INSERT INTO wallets (wallet_id, balance) VALUES ($1, $2)`, id, balance,
	); err != nil {
		t.Fatalf("create wallet: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM wallets WHERE wallet_id = $1`, id)
	})
	return id
}

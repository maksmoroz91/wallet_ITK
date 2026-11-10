package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/maksmoroz91/wallet_ITK/internal/model"
	"github.com/maksmoroz91/wallet_ITK/internal/repository/postgres"
)

const validID = "11111111-1111-1111-1111-111111111111"

type mockRepo struct {
	balance    int64
	err        error
	calls      []string
	lastID     string
	lastAmount int64
}

func (m *mockRepo) Deposit(_ context.Context, id string, amount int64) (int64, error) {
	m.calls = append(m.calls, "deposit")
	m.lastID, m.lastAmount = id, amount
	return m.balance, m.err
}

func (m *mockRepo) Withdraw(_ context.Context, id string, amount int64) (int64, error) {
	m.calls = append(m.calls, "withdraw")
	m.lastID, m.lastAmount = id, amount
	return m.balance, m.err
}

func (m *mockRepo) GetBalance(_ context.Context, id string) (int64, error) {
	m.calls = append(m.calls, "get")
	m.lastID = id
	return m.balance, m.err
}

func serve(repo WalletRepository, method, path, body string) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	NewHandler(repo).Register(mux)

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	return rr
}

func opBody(id, opType string, amount int64) string {
	return fmt.Sprintf(`{"walletId":%q,"operationType":%q,"amount":%d}`, id, opType, amount)
}

// POST /api/v1/wallet
func TestHandler_OperationWallet(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		repo      mockRepo
		wantCode  int
		wantCalls string
		wantMsg   string
	}{
		{
			name:      "успешное пополнение (DEPOSIT)",
			body:      opBody(validID, "DEPOSIT", 100),
			repo:      mockRepo{balance: 150},
			wantCode:  http.StatusOK,
			wantCalls: "deposit",
		},
		{
			name:      "успешное снятие (WITHDRAW)",
			body:      opBody(validID, "WITHDRAW", 30),
			repo:      mockRepo{balance: 70},
			wantCode:  http.StatusOK,
			wantCalls: "withdraw",
		},
		{
			name:      "кошелек не найден (DEPOSIT)",
			body:      opBody(validID, "DEPOSIT", 10),
			repo:      mockRepo{err: postgres.ErrWalletNotFound},
			wantCode:  http.StatusNotFound,
			wantCalls: "deposit",
		},
		{
			name:      "кошелек не найден (WITHDRAW)",
			body:      opBody(validID, "WITHDRAW", 10),
			repo:      mockRepo{err: postgres.ErrWalletNotFound},
			wantCode:  http.StatusNotFound,
			wantCalls: "withdraw",
		},
		{
			name:      "недостаточно средств (WITHDRAW)",
			body:      opBody(validID, "WITHDRAW", 1000),
			repo:      mockRepo{err: postgres.ErrInsufficientFunds},
			wantCode:  http.StatusUnprocessableEntity,
			wantCalls: "withdraw",
		},
		{
			name:      "ошибка базы 500",
			body:      opBody(validID, "DEPOSIT", 10),
			repo:      mockRepo{err: errors.New("ошибка базы")},
			wantCode:  http.StatusInternalServerError,
			wantCalls: "deposit",
		},

		// Невалидный ввод
		{name: "некорректный JSON", body: `invalid json`, wantCode: 400, wantMsg: "некорректный JSON"},
		{name: "некорректный UUID", body: opBody("123456", "DEPOSIT", 10), wantCode: 400, wantMsg: "некорректный walletId"},
		{name: "нулевая сумма", body: opBody(validID, "DEPOSIT", 0), wantCode: 400, wantMsg: "сумма должна быть больше нуля"},
		{name: "отрицательная сумма (DEPOSIT)", body: opBody(validID, "DEPOSIT", -5), wantCode: 400, wantMsg: "сумма должна быть больше нуля"},
		{name: "отрицательная сумма (WITHDRAW)", body: opBody(validID, "WITHDRAW", -5), wantCode: 400, wantMsg: "сумма должна быть больше нуля"},
		{name: "неизвестная операция", body: opBody(validID, "EXCHANGE", 10), wantCode: 400, wantMsg: "operationType должен быть DEPOSIT или WITHDRAW"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.repo
			rr := serve(&repo, http.MethodPost, "/api/v1/wallet", tt.body)

			require.Equal(t, tt.wantCode, rr.Code, "тело ответа: %s", rr.Body.String())
			assert.Equal(t, tt.wantCalls, strings.Join(repo.calls, ","))

			if tt.wantCode == http.StatusOK {
				var resp model.OperationResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
				assert.Equal(t, "success", resp.Status)
				assert.Equal(t, repo.balance, resp.Data.Balance)
				assert.Equal(t, validID, repo.lastID, "в репозиторий ушёл не тот walletId")
			} else {
				var resp model.ErrorResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
				assert.Equal(t, "error", resp.Status)
				if tt.wantMsg != "" {
					assert.Contains(t, resp.Message, tt.wantMsg)
				}
			}
		})
	}
}

// GET /api/v1/wallets/{walletId}
func TestHandler_GetBalanceWallet(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		repo      mockRepo
		wantCode  int
		wantCalls string
	}{
		{"успешное получение баланса", validID, mockRepo{balance: 250}, http.StatusOK, "get"},
		{"кошелек не найден", validID, mockRepo{err: postgres.ErrWalletNotFound}, http.StatusNotFound, "get"},
		{"ошибка БД -> 500", validID, mockRepo{err: errors.New("boom")}, http.StatusInternalServerError, "get"},
		{"некорректный UUID", "123456", mockRepo{}, http.StatusBadRequest, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.repo
			rr := serve(&repo, http.MethodGet, "/api/v1/wallets/"+tt.id, "")

			require.Equal(t, tt.wantCode, rr.Code, "тело ответа: %s", rr.Body.String())
			assert.Equal(t, tt.wantCalls, strings.Join(repo.calls, ","))

			if tt.wantCode == http.StatusOK {
				var resp model.BalanceResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
				assert.Equal(t, "success", resp.Status)
				assert.Equal(t, int64(250), resp.Data.Balance)
				assert.Equal(t, validID, resp.Data.WalletId)
			}
		})
	}
}

// 1000 HTTP-запросов на один кошелёк
func TestHandler_Concurrent1000_NoServerErrors(t *testing.T) {
	pool := newTestPool(t)

	id := uuid.NewString()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO wallets (wallet_id, balance) VALUES ($1, 0)`, id)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM wallets WHERE wallet_id = $1`, id)
	})

	mux := http.NewServeMux()
	NewHandler(postgres.NewWalletRepository(pool)).Register(mux)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: &http.Transport{MaxIdleConnsPerHost: 100},
	}
	defer client.CloseIdleConnections()

	post := func(opType string, amount int64) int {
		resp, err := client.Post(srv.URL+"/api/v1/wallet", "application/json",
			strings.NewReader(opBody(id, opType, amount)))
		if err != nil {
			return 0
		}
		defer resp.Body.Close()
		_, _ = io.Copy(io.Discard, resp.Body)
		return resp.StatusCode
	}

	getBalance := func() int64 {
		resp, err := client.Get(srv.URL + "/api/v1/wallets/" + id)
		require.NoError(t, err)
		defer resp.Body.Close()

		var out model.BalanceResponse
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&out))
		return out.Data.Balance
	}

	const n = 1000

	// 1000 пополнений по 1: все 200, баланс 1000
	codes := fire(n, 100, func() int { return post("DEPOSIT", 1) })
	require.Equal(t, map[int]int{200: n}, codes, "пополнения: должны быть только 200")
	require.Equal(t, int64(n), getBalance())

	// 1000 списаний по 1: все 200, баланс 0
	codes = fire(n, 100, func() int { return post("WITHDRAW", 1) })
	require.Equal(t, map[int]int{200: n}, codes, "списания: должны быть только 200")
	require.Equal(t, int64(0), getBalance())

	// списание при нулевом балансе: 422
	assert.Equal(t, http.StatusUnprocessableEntity, post("WITHDRAW", 1))
}

func newTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := "postgres://postgres:password@localhost:5432/wallet_test_db?sslmode=disable"

	cfg, err := pgxpool.ParseConfig(dsn)
	require.NoError(t, err)
	cfg.MaxConns = 30

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	require.NoError(t, err, "не удалось подключиться к БД (запущен ли docker compose up -d postgres?)")
	t.Cleanup(pool.Close)
	return pool
}

func fire(n, concurrency int, fn func() int) map[int]int {
	var (
		mu    sync.Mutex
		wg    sync.WaitGroup
		codes = map[int]int{}
		sem   = make(chan struct{}, concurrency)
	)
	for i := 0; i < n; i++ {
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			code := fn()
			mu.Lock()
			codes[code]++
			mu.Unlock()
		}()
	}
	wg.Wait()
	return codes
}

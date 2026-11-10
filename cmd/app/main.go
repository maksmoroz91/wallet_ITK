package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/maksmoroz91/wallet_ITK/internal/handler"
	"github.com/maksmoroz91/wallet_ITK/internal/repository/postgres"
)

func main() {
	port := os.Getenv("APP_PORT")
	certFile := os.Getenv("CERT_FILE")
	keyFile := os.Getenv("KEY_FILE")
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"))

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatalf("Некорректная конфигурация БД: %v", err)
	}
	maxConns := 50
	if value, err := strconv.Atoi(os.Getenv("DB_MAX_CONNS")); err == nil && value > 0 {
		maxConns = value
	}
	poolCfg.MaxConns = int32(maxConns)

	pool, err := pgxpool.NewWithConfig(context.Background(), poolCfg)
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("Не удалось пингануть БД: %v", err)
	}

	h := handler.NewHandler(postgres.NewWalletRepository(pool))

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/wallet", h.OperationWallet)
	mux.HandleFunc("GET /api/v1/wallets/{walletId}", h.GetBalanceWallet)

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("Сервер запущен на порту %s (HTTPS)", port)
	if err := srv.ListenAndServeTLS(certFile, keyFile); err != nil {
		log.Fatalln("Ошибка запуска сервера: ", err)
	}
}

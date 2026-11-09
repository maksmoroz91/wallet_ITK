package main

import (
	"log"
	"net/http"
	"os"

	"github.com/maksmoroz91/wallet_ITK/internal/handler"
)

func main() {
	port := os.Getenv("APP_PORT")
	certFile := os.Getenv("CERT_FILE")
	keyFile := os.Getenv("KEY_FILE")

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/wallet", handler.OperationWallet)
	mux.HandleFunc("GET /api/v1/wallet/{walletId}", handler.GetBalanceWallet)

	log.Println("Сервер запущен на порту 9090")
	if err := http.ListenAndServeTLS(":"+port, certFile, keyFile, mux); err != nil {
		log.Fatalln("Ошибка запуска сервера: ", err)
	}
}

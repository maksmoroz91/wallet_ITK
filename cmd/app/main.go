package main

import (
	"log"
	"net/http"

	"github.com/maksmoroz91/wallet_ITK/internal/handler"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/wallet", handler.OperationWallet)
	mux.HandleFunc("GET /api/v1/wallet/{walletId}", handler.GetBalanceWallet)

	log.Println("Сервер запущен на порту 9090")
	if err := http.ListenAndServeTLS(":9090", "certs/cert.pem", "certs/key.pem", mux); err != nil {
		log.Fatalln("Ошибка запуска сервера: ", err)
	}
}

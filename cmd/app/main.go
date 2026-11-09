package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type OperationRequest struct {
	WalletId      string `json:"walletId"`
	OperationType string `json:"operationType"`
	Amount        int64  `json:"amount"`
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/wallet", handlerOperation)
	mux.HandleFunc("GET /api/v1/wallet/{walletId}", handlerGetBalance)

	log.Println("Сервер запущен на порту 9090")
	if err := http.ListenAndServeTLS(":9090", "certs/cert.pem", "certs/key.pem", mux); err != nil {
		log.Fatalln("Ошибка запуска сервера: ", err)
	}
}

func handlerOperation(w http.ResponseWriter, r *http.Request) {
	var req OperationRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
		return
	}
	log.Printf("Пришел запрос: WalletId=%s, Type=%s, Amount=%d", req.WalletId, req.OperationType, req.Amount)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]any{
		"status":  "success",
		"message": "Запрос принят",
		"data":    req,
	}); err != nil {
		log.Println("Ошибка кодирования JSON: ", err)
	}
}

func handlerGetBalance(w http.ResponseWriter, r *http.Request) {
	walletId := r.PathValue("walletId")
	log.Println("Запрос баланса кошелька: ", walletId)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]any{
		"walletId": walletId,
		"balance":  0,
	}); err != nil {
		log.Println("Ошибка кодирования JSON: ", err)
	}

}

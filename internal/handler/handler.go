package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/maksmoroz91/wallet_ITK/internal/model"
)

func OperationWallet(w http.ResponseWriter, r *http.Request) {
	var req model.OperationRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
		return
	}
	log.Printf("Пришел запрос: WalletId=%s, Type=%s, Amount=%d", req.WalletId, req.OperationType, req.Amount)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := model.OperationResponse{
		Status:  "success",
		Message: "Запрос принят",
		Data:    req,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Ошибка кодирования JSON: ", err)
	}
}

func GetBalanceWallet(w http.ResponseWriter, r *http.Request) {
	walletId := r.PathValue("walletId")
	log.Println("Запрос баланса кошелька: ", walletId)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := model.BalanceResponse{
		WalletId: walletId,
		Balance:  0,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Ошибка кодирования JSON: ", err)
	}

}

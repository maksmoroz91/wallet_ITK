package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/maksmoroz91/wallet_ITK/internal/model"
	"github.com/maksmoroz91/wallet_ITK/internal/repository/postgres"
)

type WalletRepository interface {
	Deposit(ctx context.Context, id string, amount int64) (int64, error)
	Withdraw(ctx context.Context, id string, amount int64) (int64, error)
	GetBalance(ctx context.Context, id string) (int64, error)
}

type Handler struct {
	repo WalletRepository
}

func NewHandler(repo WalletRepository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/wallet", h.OperationWallet)
	mux.HandleFunc("GET /api/v1/wallets/{walletId}", h.GetBalanceWallet)
}

func (h *Handler) OperationWallet(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req model.OperationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}
	if _, err := uuid.Parse(req.WalletId); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный walletId")
		return
	}
	if req.Amount <= 0 {
		writeError(w, http.StatusBadRequest, "сумма должна быть меньше нуля")
		return
	}
	if req.OperationType != "DEPOSIT" && req.OperationType != "WITHDRAW" {
		writeError(w, http.StatusBadRequest, "operationType должен быть DEPOSIT или WITHDRAW")
		return
	}

	var balance int64
	var err error
	switch req.OperationType {
	case "DEPOSIT":
		balance, err = h.repo.Deposit(r.Context(), req.WalletId, req.Amount)
	case "WITHDRAW":
		balance, err = h.repo.Withdraw(r.Context(), req.WalletId, req.Amount)
	default:
		writeError(w, http.StatusBadRequest, "operationType должен быть DEPOSIT или WITHDRAW")
		return
	}
	if err != nil {
		handleRepoError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, model.OperationResponse{
		Status:  "success",
		Message: "операция выполнена",
		Data: model.OperationData{
			WalletId:      req.WalletId,
			OperationType: req.OperationType,
			Amount:        req.Amount,
			Balance:       balance,
		},
	})
}

func (h *Handler) GetBalanceWallet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("walletId")
	if _, err := uuid.Parse(id); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный walletId")
		return
	}

	balance, err := h.repo.GetBalance(r.Context(), id)
	if err != nil {
		handleRepoError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, model.BalanceResponse{
		Status: "success",
		Data:   model.WalletInfo{WalletId: id, Balance: balance},
	})
}

func handleRepoError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, postgres.ErrWalletNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, postgres.ErrInsufficientFunds):
		writeError(w, http.StatusUnprocessableEntity, err.Error())
	default:
		log.Printf("внутренняя ошибка: %v", err)
		writeError(w, http.StatusInternalServerError, "внутренняя ошибка")
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, model.ErrorResponse{Status: "error", Message: msg})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

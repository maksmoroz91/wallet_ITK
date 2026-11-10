package model

type OperationRequest struct {
	WalletId      string `json:"walletId"`
	OperationType string `json:"operationType"`
	Amount        int64  `json:"amount"`
}

type WalletInfo struct {
	WalletId string `json:"walletId"`
	Balance  int64  `json:"balance"`
}
type BalanceResponse struct {
	Status string     `json:"status"`
	Data   WalletInfo `json:"data"`
}

type OperationResponse struct {
	Status  string        `json:"status"`
	Message string        `json:"message"`
	Data    OperationData `json:"data"`
}

type OperationData struct {
	WalletId      string `json:"walletId"`
	OperationType string `json:"operationType"`
	Amount        int64  `json:"amount"`
	Balance       int64  `json:"balance"`
}

type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

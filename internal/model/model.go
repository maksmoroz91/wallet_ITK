package model

type OperationRequest struct {
	WalletId      string `json:"walletId"`
	OperationType string `json:"operationType"`
	Amount        int64  `json:"amount"`
}

type BalanceResponse struct {
	WalletId string `json:"walletId"`
	Balance  int64  `json:"balance"`
}

type OperationResponse struct {
	Status  string           `json:"status"`
	Message string           `json:"message"`
	Data    OperationRequest `json:"data"`
}

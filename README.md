## Сборка и запуск проекта
```bach
docker compose up --build
```
## WalletId тестового кошелька
```text 
11111111-1111-1111-1111-111111111111
```

## Endpoints для тестирования
GET
```text
https://localhost:9090/api/v1/wallets/11111111-1111-1111-1111-111111111111
```
POST
```text
https://localhost:9090/api/v1/wallet
```
Body
```json
{
    "walletId": "11111111-1111-1111-1111-111111111111",
    "operationType": "DEPOSIT",
    "amount": 123
}
```
```json
{
    "walletId": "11111111-1111-1111-1111-111111111111",
    "operationType": "WITHDRAW",
    "amount": 123
}
```

## Запуск всех тестов при работающем приложении
```bash
go test ./... -v -cover 
```


FROM golang:1.26-alpine AS builder
WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /wallet-app ./cmd/app

FROM alpine:3.22
RUN apk --no-cache add ca-certificates openssl

WORKDIR /app

COPY --from=builder /wallet-app .
COPY entrypoint.sh .

RUN chmod +x entrypoint.sh

EXPOSE 9090

ENTRYPOINT ["./entrypoint.sh"]
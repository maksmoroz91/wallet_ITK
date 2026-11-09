FROM golang:1.26-alpine AS builder
WORKDIR /build

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o wallet-app ./cmd/app/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates openssl

WORKDIR /app

COPY --from=builder /build/wallet-app .
COPY entrypoint.sh .

RUN chmod +x entrypoint.sh

EXPOSE 9090

ENTRYPOINT ["./entrypoint.sh"]
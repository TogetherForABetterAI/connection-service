FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
WORKDIR /app/src
RUN go mod tidy

RUN go build -o app-binary main.go

FROM alpine:latest
COPY --from=builder /auth-gateway /auth-gateway

CMD ["/auth-gateway"]
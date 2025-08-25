FROM golang:1.20-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /auth-gateway /app/src

FROM alpine:latest
COPY --from=builder /auth-gateway /auth-gateway

CMD ["/auth-gateway"]
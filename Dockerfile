FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
WORKDIR /app/src
RUN go build -o app-binary main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/src/app-binary /app/connection-service
COPY --from=builder /app/init.sql /app/init.sql

CMD ["/connection-service"]
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
WORKDIR /app/src
RUN go build -o app-binary main.go

FROM alpine:latest
COPY --from=builder /app/src/app-binary /auth-gateway

CMD ["/auth-gateway"]
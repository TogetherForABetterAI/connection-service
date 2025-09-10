FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
WORKDIR /app/src
RUN go build -o app-binary main.go

FROM alpine:latest
COPY --from=builder /app/src/app-binary /auth-gateway

CMD ["/auth-gateway"]
FROM golang:1.24.4-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/chat ./cmd

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/chat /app/chat
COPY --from=builder /app/config/config.yaml /app/config/config.yaml

EXPOSE 8080

CMD ["/app/chat"]
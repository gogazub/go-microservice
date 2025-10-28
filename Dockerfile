FROM golang:1.24.0-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o app ./cmd/main.go


FROM alpine:3.18

WORKDIR /app

COPY --from=builder /app/app .
COPY --from=builder /app/internal/api/web /app/internal/api/web
COPY --from=builder /app/.env /app/.env

EXPOSE 8081

CMD ["./app"]

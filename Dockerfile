FROM golang:1.24.0-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o app ./cmd/main.go

FROM alpine:latest
COPY --from=builder /app/app .
EXPOSE ${SERVER_PORT}

CMD ["./app"]

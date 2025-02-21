# --- Build Stage ---
FROM golang:1.23.1-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /shadowapi ./cmd/shadowapi

# --- Final Stage ---
FROM golang:1.23.1-alpine

WORKDIR /app
COPY --from=builder /shadowapi ./shadowapi

EXPOSE 8080
CMD ["./shadowapi"]

# --- Build Stage ---
FROM golang:1.23.3-alpine AS builder

WORKDIR /app

# Copy mod files from the backend folder:
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy the rest of the backend code:
COPY backend/ ./

# Build the Go binary
RUN go build -o /shadowapi ./cmd/shadowapi

# Final Stage
FROM golang:1.23.3-alpine
WORKDIR /app
COPY --from=builder /shadowapi ./shadowapi

EXPOSE 8080
CMD ["/app/shadowapi", "serve"]



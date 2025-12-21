# --- Build Stage ---
FROM --platform=linux/amd64 golang:1.24-alpine AS builder

WORKDIR /app

# Copy mod files from the backend folder:
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy the rest of the backend code:
COPY backend/ ./

# Build the Go binary for linux/amd64
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /shadowapi ./cmd/shadowapi

# Final Stage - use minimal scratch image since CGO is disabled
FROM --platform=linux/amd64 alpine:3.19
WORKDIR /app

# Add ca-certificates for HTTPS calls
RUN apk --no-cache add ca-certificates

COPY --from=builder /shadowapi ./shadowapi

EXPOSE 8080
CMD ["/app/shadowapi", "serve"]

# --- Build Stage ---
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy mod files from the backend folder:
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy the rest of the backend code:
COPY backend/ ./

# Build the Go binary
RUN go build -o /shadowapi ./cmd/shadowapi

# Final Stage
FROM golang:1.24-alpine
WORKDIR /app
COPY --from=builder /shadowapi ./shadowapi
COPY front/dist ./dist

EXPOSE 8080
CMD ["sh", "-c", "echo SA_CONFIG_PATH=$SA_CONFIG_PATH && cat $SA_CONFIG_PATH && /app/shadowapi serve"]



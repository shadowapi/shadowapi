FROM --platform=linux/amd64 golang:1.24-alpine AS builder

WORKDIR /app

# Install bash in alpine
RUN apk add --no-cache bash

# Cache dependencies
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy source code
COPY backend/ ./

# Build the worker binary for linux/amd64
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/worker ./cmd/worker

FROM --platform=linux/amd64 alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates bash jq

# Create data directory for credential persistence
RUN mkdir -p /data && chmod 700 /data

# Copy the binary from builder
COPY --from=builder /bin/worker /bin/worker

# Copy entrypoint script
COPY devops/docker/worker/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]

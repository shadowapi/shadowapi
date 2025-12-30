FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache bash git

# Cache dependencies
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy source code
COPY backend/ ./

# Build the grpc2nats binary
RUN go build -o /bin/grpc2nats ./cmd/grpc2nats

FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates

# Copy the binary from builder
COPY --from=builder /bin/grpc2nats /bin/grpc2nats

# Default port for gRPC
EXPOSE 9090

ENTRYPOINT ["/bin/grpc2nats"]

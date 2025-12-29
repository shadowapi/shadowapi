FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install bash in alpine
RUN apk add --no-cache bash

# Cache dependencies
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy source code
COPY backend/ ./

# Build the worker binary
RUN go build -o /bin/worker ./cmd/worker

FROM golang:1.24-alpine

WORKDIR /app

RUN apk add --no-cache bash

# Copy the binary from builder
COPY --from=builder /bin/worker /bin/worker

# Default command is connect
CMD ["/bin/worker", "connect"]

FROM golang:1.24-alpine AS builder

WORKDIR /app

# install bash in alpine without update
RUN apk add --no-cache bash

# Cache dependencies
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy source code
COPY backend/ ./
COPY spec/ /spec/

# Build the binary
RUN go build -o /bin/shadowapi ./cmd/shadowapi

FROM golang:1.24-alpine

WORKDIR /app

RUN apk add --no-cache bash

# Copy the binary from builder
COPY --from=builder /bin/shadowapi /bin/shadowapi

CMD ["/bin/shadowapi", "serve"]

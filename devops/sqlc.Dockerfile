# STEP 1: Build sqlc
FROM golang:1.23.3 AS builder

COPY . /workspace
WORKDIR /workspace
RUN GOBIN=/workspace go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.26.0

# STEP 2: Build a tiny image
FROM gcr.io/distroless/base-debian12

COPY --from=builder /workspace/sqlc /workspace/sqlc
ENTRYPOINT ["/workspace/sqlc"]

# Stage 1: Build Go backend
FROM golang:1.26-alpine3.22 AS backend-builder

RUN apk add -U git bash

WORKDIR /src
COPY backend/ ./

RUN go version
RUN go mod tidy
RUN go mod download
RUN go build -v -o /go/bin/shadowapi ./cmd/shadowapi

# Stage 2: Build Node frontend
FROM node:22-alpine AS frontend-builder

WORKDIR /app
COPY frontend/package*.json ./
RUN npm ci --force
COPY frontend/ ./
RUN npm run build

# Stage 3: Final image
FROM golang:1.26-alpine3.22

RUN apk add -U git bash curl
RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=backend-builder /go/bin/shadowapi /go/bin/shadowapi
COPY --from=frontend-builder /app/dist /app/dist

RUN ln -s /go/bin/shadowapi /usr/local/bin/shadowapi

EXPOSE 8090

ENTRYPOINT ["/go/bin/shadowapi", "serve"]

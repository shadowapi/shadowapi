# combine Dockerfile for production with dokku

# docker run --name sa-backend -p 8080:8080 \
#  -e SA_DB_URI="postgres://shadowapi:shadowapi@postgres:5432/shadowapi" \
#  -e SA_QUEUE_URL="nats://nats:4222" \
#  shadowapi-backend:latest

# Stage 1: Build the Go backend
FROM golang:1.24-alpine AS backend-builder
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN go build -o /shadowapi ./cmd/shadowapi

# Stage 2: Build the Node frontend
FROM node:20.10.0-alpine AS frontend-builder
WORKDIR /app
COPY front/package*.json ./
RUN npm ci --force
COPY front .
RUN npm run build

# Stage 3: Final stage - combine both builds into one image
FROM golang:1.24-alpine
WORKDIR /app
COPY --from=backend-builder /shadowapi /app/shadowapi
COPY --from=frontend-builder /app/dist /app/dist
EXPOSE 8080
CMD ["/app/shadowapi", "serve"]

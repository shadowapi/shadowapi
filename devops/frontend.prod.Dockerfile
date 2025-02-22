# --- Build Stage ---
FROM node:20.10.0-alpine AS builder

WORKDIR /app

# Copy package files and install dependencies
COPY front/package*.json ./
RUN npm ci --force

# Copy the rest of the frontend source
COPY front .

# Build the production-ready frontend
RUN npm run build

# --- Final Stage ---
FROM node:20.10.0-alpine

WORKDIR /app

# Install a simple static server
RUN npm install -g serve

# Copy the built files from the builder stage
COPY --from=builder /app/dist ./dist

EXPOSE 3000

# Serve the built frontend
CMD ["serve", "-s", "dist", "--listen", "3000"]

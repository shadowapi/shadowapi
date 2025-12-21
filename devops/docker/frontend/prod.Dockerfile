# --- Build Stage ---
FROM --platform=linux/amd64 node:20.18.0-alpine AS builder

WORKDIR /app

# Copy package files and install dependencies
COPY front/package.json front/package-lock.json ./
RUN npm ci

# Copy the rest of the frontend source (excluding node_modules via .dockerignore if present)
COPY front .

# Remove any accidentally copied node_modules and rebuild
RUN rm -rf node_modules && npm ci

# Build the production-ready frontend
RUN npm run build

# --- Final Stage ---
FROM --platform=linux/amd64 node:20.18.0-alpine

WORKDIR /app

# Install a simple static server
RUN npm install -g serve

# Copy the built files from the builder stage
COPY --from=builder /app/dist ./dist

EXPOSE 5173

# Serve the built frontend
CMD ["serve", "-s", "dist", "--listen", "5173"]

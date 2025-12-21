# --- Build Stage ---
FROM --platform=linux/amd64 node:20.18.0-alpine AS builder

# Build arguments for Vite environment variables (baked in at build time)
ARG VITE_API_BASE_URL
ARG VITE_OIDC_URL
ARG VITE_ROOT_URL
ARG VITE_APP_URL

# Set as environment variables for the build process
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL}
ENV VITE_OIDC_URL=${VITE_OIDC_URL}
ENV VITE_ROOT_URL=${VITE_ROOT_URL}
ENV VITE_APP_URL=${VITE_APP_URL}

WORKDIR /app

# Copy package files and install dependencies
COPY front/package.json front/package-lock.json ./
RUN npm ci

# Copy the rest of the frontend source (excluding node_modules via .dockerignore if present)
COPY front .

# Remove any accidentally copied node_modules and rebuild
RUN rm -rf node_modules && npm ci

# Build the production-ready frontend (VITE_ vars are baked in here)
RUN npm run build

# --- Final Stage ---
FROM --platform=linux/amd64 node:20.18.0-alpine

WORKDIR /app

# Copy package files for production dependencies
COPY front/package.json front/package-lock.json ./
RUN npm ci --omit=dev

# Install tsx for running TypeScript server
RUN npm install tsx

# Copy only the client build (not server build)
COPY --from=builder /app/dist/client ./dist

# Copy the production CSR server
COPY front/server.csr.prod.ts ./server.csr.prod.ts

EXPOSE 5173

# Run the production CSR server
CMD ["npx", "tsx", "server.csr.prod.ts"]

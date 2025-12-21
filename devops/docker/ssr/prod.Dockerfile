# --- Build Stage ---
FROM --platform=linux/amd64 node:20.18.0-alpine AS builder

WORKDIR /app

# Copy package files and install all dependencies (including devDependencies for build)
COPY front/package.json front/package-lock.json ./
RUN npm ci

# Copy the rest of the frontend source
COPY front/ .

# Build both client and server bundles
RUN npm run build

# --- Final Stage ---
FROM --platform=linux/amd64 node:20.18.0-alpine

WORKDIR /app

# Copy package files for production dependencies
COPY front/package.json front/package-lock.json ./
RUN npm ci --omit=dev

# Install tsx for running TypeScript server
RUN npm install tsx

# Copy the built files from the builder stage (client and server builds)
COPY --from=builder /app/dist/client ./dist/client
COPY --from=builder /app/dist/server ./dist/server

# Copy the production server and source files needed by it
COPY front/server.prod.ts ./server.prod.ts
COPY front/src/lib/ ./src/lib/
COPY front/tsconfig.json ./tsconfig.json
COPY front/tsconfig.node.json ./tsconfig.node.json

EXPOSE 3000

# Run the production SSR server
CMD ["npx", "tsx", "server.prod.ts"]

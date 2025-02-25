FROM node:20.10.0-alpine AS base
WORKDIR /app

# Copy only package.json first to leverage caching
COPY package.json package-lock.json ./

# Install dependencies
RUN npm install --force

# Copy everything else (avoid overwriting node_modules)
COPY . .

# Start Vite
CMD ["npm", "run", "dev"]

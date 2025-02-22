# Base stage for shared environment setup
FROM node:20.10.0-alpine AS base

# Set working directory
WORKDIR /app
COPY . .
RUN rm -rf node_modules && npm install --force

# Expose port 3000 for Vite development server
# Command to run the development server
CMD ["npm", "run", "dev"]

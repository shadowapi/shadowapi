# Base stage for shared environment setup
FROM node:20.18.0-alpine AS base

# Set working directory
WORKDIR /app
COPY front/ .
RUN rm -rf node_modules && npm install

# Copy entrypoint script
COPY devops/docker/frontend/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Expose port 3000 for Vite development server
# Command to run the development server
ENTRYPOINT ["/entrypoint.sh"]

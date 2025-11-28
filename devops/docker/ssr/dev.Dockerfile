# SSR Development Dockerfile
FROM node:20.19-alpine AS base

# Set working directory
WORKDIR /app
COPY front/ .
RUN rm -rf node_modules && npm install

# Copy entrypoint script
COPY devops/docker/ssr/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Expose port 3000 for SSR server
EXPOSE 3000

ENTRYPOINT ["/entrypoint.sh"]

#!/bin/bash

set -e  # Exit immediately if a command exits with a non-zero status
set -x  # Print commands before executing

# Define the project directory
PROJECT_DIR="/var/www/shadowapi"

echo "Starting deployment on $(hostname)..."

# Step 1: Navigate to the project directory
cd "$PROJECT_DIR" || exit 1

# Step 2: Pull the latest changes
git pull origin main

# Step 3: Bring down the running containers (to free up any locks)
docker compose down

# Step 4: Build the updated containers (only affected services)
docker compose build frontend backend

# Step 5: Start all services in detached mode
docker compose up -d

# Step 6: Apply database migrations
docker compose run --rm db-migrate

# Step 7: Clean up old images
docker image prune -f

echo "Deployment completed successfully!"

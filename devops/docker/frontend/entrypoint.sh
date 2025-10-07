#!/bin/sh
set -e

# Wait for .env.vite to be created by zitadel-init
echo "Waiting for Zitadel configuration..."
while [ ! -f /secrets/.env.vite ]; do
  sleep 1
done

# Copy environment file to Vite working directory
echo "Copying Vite environment configuration..."
cp /secrets/.env.vite /app/.env.local

echo "Starting Vite development server..."
exec npm run dev

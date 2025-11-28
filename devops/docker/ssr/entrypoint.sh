#!/bin/sh
set -e

cd /app

# Start the SSR development server with tsx
exec npm run dev:ssr

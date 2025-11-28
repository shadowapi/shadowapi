#!/bin/sh
set -e

# Optional Zitadel config for legacy flows
if [ -f /secrets/.env.vite ]; then
  echo "Found Zitadel env, copying to Vite .env.local..."
  cp /secrets/.env.vite /app/.env.local || true
else
  echo "No /secrets/.env.vite found. Proceeding without Zitadel VITE vars."
fi

echo "Starting Vite development server..."
exec npm run dev

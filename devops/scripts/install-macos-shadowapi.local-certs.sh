#!/usr/bin/env bash
set -euo pipefail

# Install local HTTPS certs for shadowapi.local
# Requires: mkcert, nginx (Homebrew)
#
# Usage: bash devops/scripts/install-macos-shadowapi.local-certs.sh

DOMAIN="shadowapi.local"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
MKCERT_DIR="$PROJECT_ROOT/devops/nginx/mkcert"
NGINX_SSL_DIR="/opt/homebrew/etc/nginx/ssl"
NGINX_SERVERS_DIR="/opt/homebrew/etc/nginx/servers"

echo "=== ShadowAPI Local HTTPS Setup ==="
echo ""

# 1. Check mkcert
if ! command -v mkcert &>/dev/null; then
    echo "Installing mkcert..."
    brew install mkcert
fi

# 2. Install mkcert CA (if not already)
echo "Ensuring mkcert CA is installed..."
mkcert -install

# 3. Generate certs
echo "Generating certs for $DOMAIN..."
cd "$MKCERT_DIR"
mkcert "$DOMAIN"
echo "Certs saved to: $MKCERT_DIR"
ls -la "$MKCERT_DIR"

# 4. Copy certs to nginx ssl dir
echo ""
echo "Copying certs to nginx ssl dir..."
sudo mkdir -p "$NGINX_SSL_DIR"
sudo cp "$MKCERT_DIR/${DOMAIN}.pem" "$NGINX_SSL_DIR/${DOMAIN}.crt"
sudo cp "$MKCERT_DIR/${DOMAIN}-key.pem" "$NGINX_SSL_DIR/${DOMAIN}.key"
echo "Installed to: $NGINX_SSL_DIR"

# 5. Symlink nginx conf
echo ""
CONF_SRC="$PROJECT_ROOT/devops/nginx/shadowapi.local.conf"
CONF_DST="$NGINX_SERVERS_DIR/shadowapi.local.conf"
if [ -L "$CONF_DST" ] || [ -f "$CONF_DST" ]; then
    echo "Nginx conf already exists at $CONF_DST — skipping"
else
    echo "Symlinking nginx conf..."
    sudo ln -s "$CONF_SRC" "$CONF_DST"
fi

# 6. Add /etc/hosts entry
echo ""
if grep -q "$DOMAIN" /etc/hosts; then
    echo "/etc/hosts already has $DOMAIN"
else
    echo "Adding $DOMAIN to /etc/hosts..."
    echo "127.0.0.1 $DOMAIN" | sudo tee -a /etc/hosts
fi

# 7. Test & restart nginx
echo ""
echo "Testing nginx config..."
sudo nginx -t

echo "Restarting nginx..."
sudo brew services restart nginx

echo ""
echo "=== Done ==="
echo "Open: https://$DOMAIN"
echo ""

#!/bin/bash
# Creates the OAuth2 SPA client in Hydra for the ShadowAPI frontend
# Run this after Hydra is up and migrations are complete

set -e

CLIENT_ID="${BE_OAUTH2_SPA_CLIENT_ID:-shadowapi-spa}"
REDIRECT_URI="${BE_OAUTH2_REDIRECT_URI:-http://localtest.me/api/v1/auth/oauth2/callback}"
HYDRA_ADMIN_URL="${HYDRA_ADMIN_ENDPOINT:-http://localhost:4445}"

echo "Creating OAuth2 client: $CLIENT_ID"

# Check if client already exists
if docker compose exec hydra hydra get client "$CLIENT_ID" --endpoint "$HYDRA_ADMIN_URL" >/dev/null 2>&1; then
    echo "Client $CLIENT_ID already exists, updating..."
    docker compose exec hydra hydra update client "$CLIENT_ID" \
        --endpoint "$HYDRA_ADMIN_URL" \
        --name "ShadowAPI SPA" \
        --grant-type authorization_code,refresh_token \
        --response-type code \
        --scope openid,offline_access,profile,email \
        --redirect-uri "$REDIRECT_URI" \
        --token-endpoint-auth-method none \
        --format json
else
    echo "Creating new client $CLIENT_ID..."
    docker compose exec hydra hydra create client \
        --endpoint "$HYDRA_ADMIN_URL" \
        --id "$CLIENT_ID" \
        --name "ShadowAPI SPA" \
        --grant-type authorization_code,refresh_token \
        --response-type code \
        --scope openid,offline_access,profile,email \
        --redirect-uri "$REDIRECT_URI" \
        --token-endpoint-auth-method none \
        --format json
fi

echo ""
echo "OAuth2 client created successfully!"
echo "Client ID: $CLIENT_ID"
echo "Redirect URI: $REDIRECT_URI"
echo ""
echo "Add this to your .env file:"
echo "BE_OAUTH2_SPA_CLIENT_ID=$CLIENT_ID"

#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_ROOT"

# Generate random string of specified length
generate_secret() {
    local length=${1:-32}
    openssl rand -base64 48 | tr -dc 'a-zA-Z0-9' | head -c "$length"
}

echo "=== ShadowAPI Bootstrap ==="

# Step 1: Force regenerate .env from template
if [ -f .env ]; then
    echo "WARNING: Removing existing .env file and regenerating from template..."
    echo "         Any custom changes will be lost!"
    rm -f .env
fi

echo "Generating secrets..."
HYDRA_SECRETS_SYSTEM=$(generate_secret 32)
OIDC_PAIRWISE_SALT=$(generate_secret 16)

echo "Creating .env from template..."
sed -e "s/__HYDRA_SECRETS_SYSTEM__/$HYDRA_SECRETS_SYSTEM/" \
    -e "s/__OIDC_PAIRWISE_SALT__/$OIDC_PAIRWISE_SALT/" \
    -e "s/__OAUTH2_CLIENT_ID__/pending-creation/" \
    .env.template > .env
echo ".env created successfully"

# Step 1.5: Generate hydra.yaml from template using envsubst
echo "Generating hydra.yaml from template..."
set -a
source .env
set +a
envsubst < devops/ory/hydra/hydra.template.yaml > devops/ory/hydra/hydra.yaml
echo "hydra.yaml created successfully"

# Step 2: Start database
echo "Starting database..."
docker compose up -d db
echo "Waiting for database to be ready..."
sleep 5

# Step 3: Run database migrations
echo "Running database migrations..."
make sync-db

# Step 4: Start Hydra
echo "Starting Hydra..."
docker compose up -d hydra
echo "Waiting for Hydra..."

# Wait for Hydra to be ready
for i in {1..30}; do
    if docker compose exec -T hydra hydra version >/dev/null 2>&1; then
        break
    fi
    echo "Waiting for Hydra... ($i/30)"
    sleep 2
done

# Step 5: Create OAuth2 client in Hydra (idempotent)
echo "Creating OAuth2 client..."
# Redirect URI uses api subdomain
REDIRECT_URI="${BE_PROTOCOL:-http}://${BE_API_SUBDOMAIN:-api}.${BE_DOMAIN}/api/v1/auth/oauth2/callback"
CLIENT_NAME="ShadowAPI SPA"

# Check if client already exists by name (search in list)
EXISTING_CLIENT=$(docker compose exec -T hydra hydra list oauth2-clients \
    --endpoint http://localhost:4445 \
    --format json 2>/dev/null | jq -r ".items[] | select(.client_name == \"$CLIENT_NAME\") | .client_id" || echo "")

if [ -n "$EXISTING_CLIENT" ]; then
    echo "OAuth2 client already exists: $EXISTING_CLIENT"
    CLIENT_ID="$EXISTING_CLIENT"
else
    # Create new client (ID is auto-generated)
    CLIENT_RESPONSE=$(docker compose exec -T hydra hydra create oauth2-client \
        --endpoint http://localhost:4445 \
        --name "$CLIENT_NAME" \
        --grant-type authorization_code,refresh_token \
        --response-type code \
        --scope openid,offline_access,profile,email \
        --redirect-uri "$REDIRECT_URI" \
        --token-endpoint-auth-method none \
        --format json)

    CLIENT_ID=$(echo "$CLIENT_RESPONSE" | jq -r '.client_id')
    echo "OAuth2 client created: $CLIENT_ID"

    # Update .env with the generated client ID
    sed -i "s/^BE_OAUTH2_SPA_CLIENT_ID=.*/BE_OAUTH2_SPA_CLIENT_ID=$CLIENT_ID/" .env
    echo "Updated .env with client ID"
fi

# Step 6: Start the full stack
# The backend will create the admin user on first startup via ensureInitAdmin
# if BE_INIT_ADMIN_EMAIL and BE_INIT_ADMIN_PASSWORD are set in .env
echo "Starting all services..."
docker compose up -d

# Read test user credentials from .env
TEST_EMAIL=$(grep "^BE_INIT_ADMIN_EMAIL=" .env | cut -d'=' -f2)
TEST_PASSWORD=$(grep "^BE_INIT_ADMIN_PASSWORD=" .env | cut -d'=' -f2)

echo ""
echo "=== Bootstrap Complete ==="
echo ""
echo "Services:"
echo "  - Frontend (SPA):  ${BE_PROTOCOL:-http}://${BE_DOMAIN}"
echo "  - API:             ${BE_PROTOCOL:-http}://${BE_API_SUBDOMAIN:-api}.${BE_DOMAIN}"
echo "  - OIDC:            ${BE_PROTOCOL:-http}://${BE_OIDC_SUBDOMAIN:-oidc}.${BE_DOMAIN}"
echo "  - SSR (www):       ${BE_PROTOCOL:-http}://${BE_SSR_SUBDOMAIN:-www}.${BE_DOMAIN}"
echo ""
echo "Workspaces:"
echo "  - Internal: ${BE_PROTOCOL:-http}://${BE_DOMAIN}/w/internal"
echo "  - Demo:     ${BE_PROTOCOL:-http}://${BE_DOMAIN}/w/demo"
echo ""
echo "Test login:       $TEST_EMAIL / $TEST_PASSWORD"
echo "OAuth2 Client ID: $CLIENT_ID"
echo ""
echo "The admin user has super_admin role and owns 'internal' and 'demo' workspaces."

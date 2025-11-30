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

# Step 1: Generate .env if it doesn't exist
if [ -f .env ]; then
    echo ".env already exists, skipping secret generation..."
else
    echo "Generating secrets..."
    KRATOS_SECRETS_DEFAULT=$(generate_secret 32)
    KRATOS_SECRETS_COOKIE=$(generate_secret 32)
    HYDRA_SECRETS_SYSTEM=$(generate_secret 32)
    OIDC_PAIRWISE_SALT=$(generate_secret 16)

    echo "Creating .env from template..."
    sed -e "s/__KRATOS_SECRETS_DEFAULT__/$KRATOS_SECRETS_DEFAULT/" \
        -e "s/__KRATOS_SECRETS_COOKIE__/$KRATOS_SECRETS_COOKIE/" \
        -e "s/__HYDRA_SECRETS_SYSTEM__/$HYDRA_SECRETS_SYSTEM/" \
        -e "s/__OIDC_PAIRWISE_SALT__/$OIDC_PAIRWISE_SALT/" \
        -e "s/__OAUTH2_CLIENT_ID__/pending-creation/" \
        .env.template > .env
    echo ".env created successfully"
fi

# Step 2: Start database
echo "Starting database..."
docker compose up -d db
echo "Waiting for database to be ready..."
sleep 5

# Step 3: Run database migrations
echo "Running database migrations..."
make sync-db

# Step 4: Start Kratos and Hydra
echo "Starting Kratos and Hydra..."
docker compose up -d kratos hydra
echo "Waiting for auth services..."

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
REDIRECT_URI="http://localtest.me/api/v1/auth/oauth2/callback"
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

# Step 6: Create test user in Kratos (idempotent)
echo "Creating test user..."
TEST_EMAIL="admin@example.com"
TEST_PASSWORD="Admin123!"
KRATOS_ADMIN_URL="http://kratos:4434"

# Wait for Kratos admin API (using docker exec since port not exposed)
for i in {1..30}; do
    if docker compose exec -T kratos wget -q -O /dev/null "$KRATOS_ADMIN_URL/admin/health/ready" 2>/dev/null; then
        break
    fi
    echo "Waiting for Kratos admin... ($i/30)"
    sleep 2
done

# Check if user exists (using docker exec)
EXISTING_USER=$(docker compose exec -T kratos wget -q -O - "$KRATOS_ADMIN_URL/admin/identities" 2>/dev/null | grep -o "$TEST_EMAIL" || echo "")

if [ -n "$EXISTING_USER" ]; then
    echo "Test user already exists, skipping..."
else
    # Create user via Kratos admin API
    docker compose exec -T kratos wget -q -O /dev/null \
        --header="Content-Type: application/json" \
        --post-data='{
            "schema_id": "default",
            "traits": {"email": "'"$TEST_EMAIL"'"},
            "credentials": {"password": {"config": {"password": "'"$TEST_PASSWORD"'"}}}
        }' \
        "$KRATOS_ADMIN_URL/admin/identities" 2>/dev/null && \
    echo "Test user created: $TEST_EMAIL / $TEST_PASSWORD" || \
    echo "Failed to create test user (may already exist)"
fi

# Step 7: Start the full stack
echo "Starting all services..."
docker compose up -d

echo ""
echo "=== Bootstrap Complete ==="
echo "Application: http://localtest.me"
echo "Test login:  $TEST_EMAIL / $TEST_PASSWORD"
echo "OAuth2 Client ID: $CLIENT_ID"

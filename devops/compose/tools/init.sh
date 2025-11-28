#!/bin/sh
set -e

echo "=== ShadowAPI Initialization ==="

# Step 1: Check and copy .env.example to .env
if [ -f .env ]; then
  echo "✓ .env file already exists, skipping copy"
else
  echo "→ Copying .env.example to .env..."
  cp .env.example .env
  echo "✓ .env file created"
fi

# Step 2: Check if Zitadel app already created (backend audience present)
if grep -q "BE_ZITADEL_AUDIENCE=" .env && [ "$(grep "BE_ZITADEL_AUDIENCE=" .env | cut -d'=' -f2)" != "" ]; then
  echo "✓ Zitadel app already configured (BE_ZITADEL_AUDIENCE exists)"
  echo "=== Initialization complete ==="
  exit 0
fi

# Step 3: Install dependencies
echo "→ Installing dependencies..."
apk add --no-cache curl python3 bash >/dev/null 2>&1

# Step 4: Wait for Zitadel to be ready
echo "→ Waiting for Zitadel to be ready..."
max_attempts=30
attempt=0
while [ $attempt -lt $max_attempts ]; do
  if curl -sf http://auth.localtest.me:8080/debug/healthz >/dev/null 2>&1; then
    echo "✓ Zitadel is ready"
    break
  fi
  attempt=$((attempt + 1))
  sleep 2
done

if [ $attempt -eq $max_attempts ]; then
  echo "✗ Zitadel failed to start"
  exit 1
fi

# Step 5: Read PAT from secrets
echo "→ Reading Zitadel admin PAT..."
PAT=$(cat /secrets/shadowapi-admin-service.pat | tr -d '\n')
URL=http://auth.localtest.me:8080
HOST=auth.localtest.me

# Step 6: Configure instance features using v2beta API
echo "→ Configuring Zitadel instance features..."
curl -s -X PUT "$URL/v2/features/instance" \
  -H "Host: $HOST" \
  -H "Authorization: Bearer $PAT" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  --data-raw '{
    "loginDefaultOrg": false,
    "userSchema": true,
    "oidcTokenExchange": true,
    "improvedPerformance": ["IMPROVED_PERFORMANCE_ORG_BY_ID"],
    "debugOidcParentError": true,
    "oidcSingleV1SessionTermination": true,
    "enableBackChannelLogout": true,
    "loginV2": {
      "required": true,
      "baseUri": "http://localtest.me/login"
    },
    "permissionCheckV2": true,
    "consoleUseV2UserApi": false
  }' >/dev/null
echo "✓ Instance features configured"

# Step 6.5: Search for organization
echo "→ Searching for organization..."
ORG=$(curl -s -L "$URL/v2/organizations/_search" \
  -H "Host: $HOST" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -H "Authorization: Bearer $PAT" \
  -d '{
    "query": {
      "offset": "0",
      "limit": 100,
      "asc": true
    },
    "sortingColumn": "ORGANIZATION_FIELD_NAME_UNSPECIFIED",
    "queries": [
      {
        "nameQuery": {
          "name": "ShadowAPI",
          "method": "TEXT_QUERY_METHOD_EQUALS"
        }
      }
    ]
  }')
ORG_ID=$(echo "$ORG" | python3 -c "import sys,json;print(json.load(sys.stdin)['result'][0]['id'])")
echo "✓ Organization found: $ORG_ID"

# Step 7: Create Zitadel project (v2beta API)
echo "→ Creating Zitadel project..."
PROJECT=$(curl -s -X POST "$URL/zitadel.project.v2beta.ProjectService/CreateProject" \
  -H "Host: $HOST" \
  -H "Authorization: Bearer $PAT" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\":\"ShadowAPI\",
    \"organizationId\": \"$ORG_ID\"
  }")
PROJECT_ID=$(echo "$PROJECT" | python3 -c "import sys,json;print(json.load(sys.stdin)['id'])")
echo "✓ Project created: $PROJECT_ID"

# Step 8: Create OIDC app (v2beta API)
echo "→ Creating OIDC application..."
APP=$(curl -s -X POST "$URL/zitadel.app.v2beta.AppService/CreateApplication" \
  -H "Host: $HOST" \
  -H "Authorization: Bearer $PAT" \
  -H "Content-Type: application/json" \
  -d "{
    \"projectId\":\"$PROJECT_ID\",
    \"name\":\"ShadowAPI Frontend\",
    \"oidcRequest\":{
      \"redirectUris\":[\"http://localtest.me/api/v1/auth/callback\",\"http://localhost:5173/api/v1/auth/callback\",\"http://localtest.me/auth/callback\"],
      \"postLogoutRedirectUris\":[\"http://localtest.me/\",\"http://localhost:5173/\"],
      \"responseTypes\":[\"OIDC_RESPONSE_TYPE_CODE\"],
      \"grantTypes\":[\"OIDC_GRANT_TYPE_AUTHORIZATION_CODE\",\"OIDC_GRANT_TYPE_REFRESH_TOKEN\",\"OIDC_GRANT_TYPE_TOKEN_EXCHANGE\"],
      \"appType\":\"OIDC_APP_TYPE_USER_AGENT\",
      \"authMethodType\":\"OIDC_AUTH_METHOD_TYPE_NONE\",
      \"version\":\"OIDC_VERSION_1_0\",
      \"devMode\":true,
      \"accessTokenType\":\"OIDC_TOKEN_TYPE_JWT\",
      \"additionalOrigins\":[\"http://localtest.me\",\"http://localhost:5173\"]
    }
  }")
APP_ID=$(echo "$APP" | python3 -c "import sys,json;print(json.load(sys.stdin)['appId'])")
CLIENT_ID=$(echo "$APP" | python3 -c "import sys,json;print(json.load(sys.stdin)['oidcResponse']['clientId'])")
echo "✓ OIDC app created: $CLIENT_ID"

# Debug - verify CLIENT_ID is not empty
if [ -z "$CLIENT_ID" ]; then
  echo "✗ ERROR: CLIENT_ID is empty!"
  echo "Response was: $APP"
  exit 1
fi

# Step 9: Get application details (v2beta API)
echo "→ Fetching application details..."
APP_DETAILS=$(curl -s -X POST "$URL/zitadel.app.v2beta.AppService/GetApplication" \
  -H "Host: $HOST" \
  -H "Authorization: Bearer $PAT" \
  -H "Content-Type: application/json" \
  -d "{\"id\":\"$APP_ID\"}")

echo "Application details:"
echo "$APP_DETAILS" | python3 -m json.tool

# Step 10: Update .env with generated values for backend
echo "→ Updating .env with backend OIDC values..."

# ensure BE_ZITADEL_AUDIENCE is set
grep -v "^BE_ZITADEL_AUDIENCE=" .env >.env.tmp || true
echo "BE_ZITADEL_AUDIENCE=$CLIENT_ID" >>.env.tmp

# ensure BE_ZITADEL_REDIRECT_URL is set (fallback to default callback)
if ! grep -q "^BE_ZITADEL_REDIRECT_URL=" .env; then
  echo "BE_ZITADEL_REDIRECT_URL=http://localtest.me/api/v1/auth/callback" >>.env.tmp
fi
mv .env.tmp .env

echo "✓ .env updated with BE_ZITADEL_AUDIENCE=$CLIENT_ID"

# Optional: also write VITE_* for legacy flows (not required anymore)
echo "VITE_ZITADEL_CLIENT_ID=$CLIENT_ID" >/secrets/.env.vite
echo "VITE_ZITADEL_URL=$BE_ZITADEL_URL" >>/secrets/.env.vite
echo "VITE_ZITADEL_REDIRECT_URL=$BE_ZITADEL_REDIRECT_URL" >>/secrets/.env.vite

echo "=== Initialization complete ==="
echo ""
echo "Next steps:"
echo "  docker compose up"

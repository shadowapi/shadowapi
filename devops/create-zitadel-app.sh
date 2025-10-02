#!/bin/bash
set -e

PAT=$(cat ./secrets/shadowapi-admin-service.pat | tr -d '\n')
URL=http://auth.localtest.me

# Create project
PROJECT=$(curl -s -X POST "$URL/management/v1/projects" \
  -H "Authorization: Bearer $PAT" \
  -H "Content-Type: application/json" \
  -d '{"name":"ShadowAPI"}')
PROJECT_ID=$(echo "$PROJECT" | python3 -c "import sys,json;print(json.load(sys.stdin)['id'])")

# Create OIDC app
APP=$(curl -s -X POST "$URL/management/v1/projects/$PROJECT_ID/apps/oidc" \
  -H "Authorization: Bearer $PAT" \
  -H "Content-Type: application/json" \
  -d '{
    "name":"ShadowAPI Frontend",
    "redirectUris":["http://localtest.me/login","http://localhost:5173/login"],
    "postLogoutRedirectUris":["http://localtest.me/","http://localhost:5173/"],
    "responseTypes":["OIDC_RESPONSE_TYPE_CODE"],
    "grantTypes":["OIDC_GRANT_TYPE_AUTHORIZATION_CODE","OIDC_GRANT_TYPE_REFRESH_TOKEN"],
    "appType":"OIDC_APP_TYPE_USER_AGENT",
    "authMethodType":"OIDC_AUTH_METHOD_TYPE_NONE",
    "version":"OIDC_VERSION_1_0",
    "devMode":true,
    "accessTokenType":"OIDC_TOKEN_TYPE_JWT",
    "additionalOrigins":["http://localtest.me","http://localhost:5173"]
  }')

CLIENT_ID=$(echo "$APP" | python3 -c "import sys,json;print(json.load(sys.stdin)['clientId'])")

echo "Client ID: $CLIENT_ID"
echo "Update VITE_ZITADEL_CLIENT_ID in front/.env and .env.example"

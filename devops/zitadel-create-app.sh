#!/bin/sh
set -e

if [ -f /secrets/zitadel-app-created ]; then
  echo "Zitadel app already created, skipping..."
  exit 0
fi

apk add --no-cache curl bash

PAT=$(cat /secrets/shadowapi-admin-service.pat | tr -d '\n')
URL=http://zitadel:8080
HOST=auth.localtest.me

# Create project
PROJECT=$(curl -s -X POST "$URL/management/v1/projects" \
  -H "Host: $HOST" \
  -H "Authorization: Bearer $PAT" \
  -H "Content-Type: application/json" \
  -d '{"name":"ShadowAPI"}')
PROJECT_ID=$(echo "$PROJECT" | python3 -c "import sys,json;print(json.load(sys.stdin)['id'])")

# Create OIDC app
APP=$(curl -s -X POST "$URL/management/v1/projects/$PROJECT_ID/apps/oidc" \
  -H "Host: $HOST" \
  -H "Authorization: Bearer $PAT" \
  -H "Content-Type: application/json" \
  -d '{
    "name":"ShadowAPI Frontend",
    "redirectUris":["http://localtest.me/login","http://localhost:5173/login"],
    "postLogoutRedirectUris":["http://localtest.me/","http://localhost:5173/"],
    "responseTypes":["OIDC_RESPONSE_TYPE_CODE"],
    "grantTypes":["OIDC_GRANT_TYPE_AUTHORIZATION_CODE","OIDC_GRANT_TYPE_REFRESH_TOKEN","OIDC_GRANT_TYPE_TOKEN_EXCHANGE"],
    "appType":"OIDC_APP_TYPE_USER_AGENT",
    "authMethodType":"OIDC_AUTH_METHOD_TYPE_NONE",
    "version":"OIDC_VERSION_1_0",
    "devMode":true,
    "accessTokenType":"OIDC_TOKEN_TYPE_JWT",
    "additionalOrigins":["http://localtest.me","http://localhost:5173"]
  }')

CLIENT_ID=$(echo "$APP" | python3 -c "import sys,json;print(json.load(sys.stdin)['clientId'])")

echo "Client ID: $CLIENT_ID"

# Enable token exchange feature (beta feature)
echo "Enabling token exchange feature..."
curl -s -X PUT "$URL/v2/features/instance" \
  -H "Host: $HOST" \
  -H "Authorization: Bearer $PAT" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  --data-raw '{
    "oidcTokenExchange": true
  }' || echo "Token exchange feature may already be enabled or error occurred"

# # Configure organization login policy to allow external login
# echo "Configuring login policy for custom login UI..."
# curl -s -X POST "$URL/management/v1/policies/login" \
#   -H "Host: $HOST" \
#   -H "Authorization: Bearer $PAT" \
#   -H "Content-Type: application/json" \
#   -d '{
#     "allowExternalIdp": true,
#     "allowRegister": true,
#     "allowUsernamePassword": true,
#     "externalLoginCheckAllowed": true,
#     "forceMfa": false,
#     "hidePasswordReset": false,
#     "ignoreUnknownUsernames": false,
#     "passwordlessType": "PASSWORDLESS_TYPE_NOT_ALLOWED"
#   }' || echo "Login policy may already exist or error occurred"

echo "VITE_ZITADEL_CLIENT_ID=$CLIENT_ID" >/app/.env.gen
echo "VITE_ZITADEL_URL=$SA_ZITADEL_URL" >>/app/.env.gen
echo "VITE_ZITADEL_REDIRECT_URL=$SA_ZITADEL_REDIRECT_URL" >>/app/.env.gen
touch /secrets/zitadel-app-created
echo "Zitadel app setup completed"

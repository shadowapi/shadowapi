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

# Configure instance features
echo "Configuring instance features..."
curl -s -X PUT "$URL/v2/features/instance" \
  -H "Host: $HOST" \
  -H "Authorization: Bearer $PAT" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  --data-raw '{
    "loginDefaultOrg": false,
    "userSchema": true,
    "oidcTokenExchange": true,
    "improvedPerformance": [1],
    "debugOidcParentError": true,
    "oidcSingleV1SessionTermination": true,
    "enableBackChannelLogout": true,
    "loginV2": {
      "required": true,
      "baseUri": "http://auth.localtest.me/ui/v2/login"
    },
    "permissionCheckV2": true,
    "consoleUseV2UserApi": false,
    "enableRelationalTables": true
  }' || echo "Instance features configuration may have failed"

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

echo "VITE_ZITADEL_CLIENT_ID=$CLIENT_ID" >/secrets/.env.vite
echo "VITE_ZITADEL_URL=$BE_ZITADEL_URL" >>/secrets/.env.vite
echo "VITE_ZITADEL_REDIRECT_URL=$BE_ZITADEL_REDIRECT_URL" >>/secrets/.env.vite
touch /secrets/zitadel-app-created
echo "Zitadel app setup completed"

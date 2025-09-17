#!/bin/bash

# Zitadel Service User Automation Script
# This script creates a service user and generates keys for ShadowAPI

set -e

# Configuration
ZITADEL_URL="${ZITADEL_URL:-http://auth.localtest.me}"
ADMIN_EMAIL="${ADMIN_EMAIL:-admin@example.com}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-Admin123!}"
SERVICE_USER_NAME="shadowapi-service"
SERVICE_USER_DISPLAY_NAME="ShadowAPI Service User"
OUTPUT_DIR="${OUTPUT_DIR:-./secrets}"
KEY_FILE="${OUTPUT_DIR}/zitadel-service-key.json"

echo "🚀 Setting up Zitadel Service User for ShadowAPI..."

# Wait for Zitadel to be ready
echo "⏳ Waiting for Zitadel to be ready..."
until curl -s "${ZITADEL_URL}/debug/ready" > /dev/null 2>&1; do
  echo "Waiting for Zitadel..."
  sleep 5
done
echo "✅ Zitadel is ready!"

# Create output directory
mkdir -p "${OUTPUT_DIR}"

# Get admin access token
echo "🔑 Getting admin access token..."
ADMIN_TOKEN_RESPONSE=$(curl -s -X POST "${ZITADEL_URL}/oauth/v2/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=password&username=${ADMIN_EMAIL}&password=${ADMIN_PASSWORD}&client_id=zitadel&scope=openid profile email urn:zitadel:iam:org:project:id:zitadel:aud")

if [ $? -ne 0 ]; then
  echo "❌ Failed to get admin token"
  exit 1
fi

ACCESS_TOKEN=$(echo "${ADMIN_TOKEN_RESPONSE}" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)

if [ -z "${ACCESS_TOKEN}" ]; then
  echo "❌ Failed to extract access token"
  echo "Response: ${ADMIN_TOKEN_RESPONSE}"
  exit 1
fi

echo "✅ Got admin access token"

# Create service user
echo "👤 Creating service user..."
CREATE_USER_RESPONSE=$(curl -s -X POST "${ZITADEL_URL}/management/v1/users/machine" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "userName": "'${SERVICE_USER_NAME}'",
    "name": "'${SERVICE_USER_DISPLAY_NAME}'",
    "description": "Service user for ShadowAPI Management API access",
    "accessTokenType": "ACCESS_TOKEN_TYPE_BEARER"
  }')

if [ $? -ne 0 ]; then
  echo "❌ Failed to create service user"
  exit 1
fi

# Extract user ID
USER_ID=$(echo "${CREATE_USER_RESPONSE}" | grep -o '"userId":"[^"]*"' | cut -d'"' -f4)

if [ -z "${USER_ID}" ]; then
  echo "❌ Failed to extract user ID"
  echo "Response: ${CREATE_USER_RESPONSE}"
  exit 1
fi

echo "✅ Created service user with ID: ${USER_ID}"

# Generate private key for service user
echo "🔐 Generating private key for service user..."
KEY_RESPONSE=$(curl -s -X POST "${ZITADEL_URL}/management/v1/users/${USER_ID}/keys" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "KEY_TYPE_JSON",
    "expirationDate": null
  }')

if [ $? -ne 0 ]; then
  echo "❌ Failed to generate private key"
  exit 1
fi

# Extract key data
KEY_ID=$(echo "${KEY_RESPONSE}" | grep -o '"keyId":"[^"]*"' | cut -d'"' -f4)
PRIVATE_KEY=$(echo "${KEY_RESPONSE}" | grep -o '"keyDetails":"[^"]*"' | cut -d'"' -f4)

if [ -z "${KEY_ID}" ] || [ -z "${PRIVATE_KEY}" ]; then
  echo "❌ Failed to extract key information"
  echo "Response: ${KEY_RESPONSE}"
  exit 1
fi

# Decode the base64 private key
DECODED_KEY=$(echo "${PRIVATE_KEY}" | base64 -d)

# Create key file
echo "💾 Creating key file..."
cat > "${KEY_FILE}" << EOF
{
  "type": "serviceaccount",
  "keyId": "${KEY_ID}",
  "key": "${DECODED_KEY}",
  "userId": "${USER_ID}"
}
EOF

echo "✅ Key file created: ${KEY_FILE}"

# Grant necessary permissions to service user
echo "🔒 Granting permissions to service user..."

# Get organization ID
ORG_RESPONSE=$(curl -s -X POST "${ZITADEL_URL}/management/v1/orgs/_search" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"limit": 1}')

ORG_ID=$(echo "${ORG_RESPONSE}" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

if [ -n "${ORG_ID}" ]; then
  # Grant ORG_OWNER role
  curl -s -X POST "${ZITADEL_URL}/management/v1/orgs/${ORG_ID}/members" \
    -H "Authorization: Bearer ${ACCESS_TOKEN}" \
    -H "Content-Type: application/json" \
    -d '{
      "userId": "'${USER_ID}'",
      "roles": ["ORG_OWNER"]
    }' > /dev/null

  echo "✅ Granted ORG_OWNER permissions"
fi

# Output environment variables
echo ""
echo "🎉 Setup completed successfully!"
echo ""
echo "📋 Add these environment variables to your .env file:"
echo "SA_AUTH_USER_MANAGER=zitadel"
echo "SA_AUTH_ZITADEL_SERVICE_USER_ID=${USER_ID}"
echo "SA_AUTH_ZITADEL_SERVICE_USER_KEY_PATH=${KEY_FILE}"
echo ""
echo "🔐 Key file location: ${KEY_FILE}"
echo "⚠️  Keep this key file secure and never commit it to version control!"

# Set secure permissions on key file
chmod 600 "${KEY_FILE}"
echo "🔒 Set secure permissions on key file"
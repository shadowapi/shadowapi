#!/bin/sh
set -e

echo "=== Creating Test User in Zitadel ==="

# Install dependencies
echo "→ Installing dependencies..."
apk add --no-cache curl python3 >/dev/null 2>&1

# Configuration
PAT=$(cat /secrets/shadowapi-admin-service.pat | tr -d '\n')
URL=http://auth.localtest.me:8080
HOST=auth.localtest.me
EMAIL="admin@example.com"
PASSWORD="Admin123!"
FIRST_NAME="Admin"
LAST_NAME="User"

echo "→ Checking if test user already exists..."

# Try to find user by email
SEARCH_RESULT=$(curl -s -X POST "$URL/management/v1/users/_search" \
  -H "Host: $HOST" \
  -H "Authorization: Bearer $PAT" \
  -H "Content-Type: application/json" \
  -d "{
    \"query\": \"$EMAIL\",
    \"queries\": [{
      \"emailQuery\": {
        \"emailAddress\": \"$EMAIL\"
      }
    }]
  }")

# Check if user exists
USER_COUNT=$(echo "$SEARCH_RESULT" | python3 -c "import sys,json;data=json.load(sys.stdin);print(len(data.get('result', [])))" 2>/dev/null || echo "0")

if [ "$USER_COUNT" != "0" ]; then
  echo "✓ Test user already exists: $EMAIL"
  exit 0
fi

echo "→ Creating test user..."

# Create human user with email
USER_RESPONSE=$(curl -s -X POST "$URL/management/v1/users/human/_import" \
  -H "Host: $HOST" \
  -H "Authorization: Bearer $PAT" \
  -H "Content-Type: application/json" \
  -d "{
    \"userName\": \"$EMAIL\",
    \"profile\": {
      \"firstName\": \"$FIRST_NAME\",
      \"lastName\": \"$LAST_NAME\",
      \"displayName\": \"$FIRST_NAME $LAST_NAME\"
    },
    \"email\": {
      \"email\": \"$EMAIL\",
      \"isEmailVerified\": true
    },
    \"password\": \"$PASSWORD\",
    \"passwordChangeRequired\": false
  }")

echo "Response: $USER_RESPONSE"

# Extract user ID
USER_ID=$(echo "$USER_RESPONSE" | python3 -c "import sys,json;data=json.load(sys.stdin);print(data.get('userId', ''))" 2>/dev/null || echo "")

if [ -z "$USER_ID" ]; then
  echo "✗ ERROR: Failed to create user"
  echo "Response was: $USER_RESPONSE"
  exit 1
fi

echo "✓ Test user created successfully"
echo "  Email: $EMAIL"
echo "  Password: $PASSWORD"
echo "  User ID: $USER_ID"

echo "=== Test user setup complete ==="

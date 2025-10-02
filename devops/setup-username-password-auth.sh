#!/bin/bash

# ZITADEL Username/Password Authentication Setup Script
# Creates application for traditional username/password login
set -e

# Configuration
ZITADEL_URL="${ZITADEL_URL:-http://zitadel:8080}"
ZITADEL_EXTERNAL_URL="http://auth.localtest.me"
MACHINE_KEY_FILE="./secrets/shadowapi-admin-service.json"
CONFIG_FILE="./zitadel-auth-config.json"
SETUP_COMPLETE_FILE="./secrets/.zitadel-auth-setup-complete"
PROJECT_NAME="ShadowAPI"
APP_NAME="ShadowAPI Frontend"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if setup is already complete
check_setup_complete() {
    if [ -f "$SETUP_COMPLETE_FILE" ] && [ -f "$CONFIG_FILE" ]; then
        log_info "ZITADEL username/password auth setup already completed. Skipping..."
        if [ -f "$CONFIG_FILE" ]; then
            log_info "Configuration:"
            cat "$CONFIG_FILE"
        fi
        exit 0
    fi
}

# Wait for ZITADEL to be ready
wait_for_zitadel() {
    log_info "Waiting for ZITADEL to be ready..."
    local max_attempts=30
    local attempt=1

    while [ $attempt -le $max_attempts ]; do
        if curl -s --fail "$ZITADEL_URL/debug/healthz" >/dev/null 2>&1; then
            log_success "ZITADEL is ready!"
            return 0
        fi

        log_info "Attempt $attempt/$max_attempts: ZITADEL not ready yet, waiting 10 seconds..."
        sleep 10
        attempt=$((attempt + 1))
    done

    log_error "ZITADEL did not become ready within expected time"
    exit 1
}

# Wait for machine key to be available
wait_for_machine_key() {
    log_info "Waiting for machine key to be available..."
    local max_attempts=10
    local attempt=1

    while [ $attempt -le $max_attempts ]; do
        if [ -f "$MACHINE_KEY_FILE" ]; then
            log_success "Machine key found!"
            return 0
        fi

        log_info "Attempt $attempt/$max_attempts: Machine key not found, waiting 5 seconds..."
        sleep 5
        attempt=$((attempt + 1))
    done

    log_error "Machine key file not found at $MACHINE_KEY_FILE"
    exit 1
}

# Check dependencies
check_dependencies() {
    command -v jq >/dev/null 2>&1 || {
        log_error "jq is required but not installed. Aborting."
        exit 1
    }
    command -v openssl >/dev/null 2>&1 || {
        log_error "openssl is required but not installed. Aborting."
        exit 1
    }
}

# Function to create JWT
create_jwt() {
    local iss="$1"
    local sub="$2"
    local aud="$3"
    local key_id="$4"
    local private_key_file="$5"

    # JWT Header
    header=$(echo -n '{"alg":"RS256","typ":"JWT","kid":"'$key_id'"}' | base64 | tr -d '=' | tr '/+' '_-' | tr -d '\n')

    # JWT Payload
    iat=$(date +%s)
    exp=$((iat + 3600))  # 1 hour expiration

    payload=$(echo -n '{"iss":"'$iss'","sub":"'$sub'","aud":"'$aud'","iat":'$iat',"exp":'$exp'}' | base64 | tr -d '=' | tr '/+' '_-' | tr -d '\n')

    # Create signature
    unsigned_token="$header.$payload"
    signature=$(echo -n "$unsigned_token" | openssl dgst -sha256 -sign "$private_key_file" | base64 | tr -d '=' | tr '/+' '_-' | tr -d '\n')

    echo "$header.$payload.$signature"
}

# Get access token using machine key
get_access_token() {
    log_info "Authenticating with ZITADEL using machine key..."

    # Extract values from machine key
    local key_id=$(jq -r '.keyId' "$MACHINE_KEY_FILE")
    local user_id=$(jq -r '.userId' "$MACHINE_KEY_FILE")
    local private_key=$(jq -r '.key' "$MACHINE_KEY_FILE")

    # Create a temporary private key file
    local temp_key_file=$(mktemp)
    echo "$private_key" > "$temp_key_file"

    # Create JWT for authentication
    local jwt=$(create_jwt "$user_id" "$user_id" "$ZITADEL_EXTERNAL_URL" "$key_id" "$temp_key_file")

    # Get access token
    local token_response=$(curl -s -X POST "$ZITADEL_URL/oauth/v2/token" \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -d "grant_type=urn:ietf:params:oauth:grant-type:jwt-bearer&assertion=$jwt&scope=openid profile urn:zitadel:iam:org:project:id:zitadel:aud")

    local access_token=$(echo "$token_response" | jq -r '.access_token // empty')

    # Cleanup
    rm -f "$temp_key_file"

    if [ -z "$access_token" ] || [ "$access_token" = "null" ]; then
        log_error "Failed to get access token"
        log_error "Response: $token_response"
        exit 1
    fi

    log_success "Access token obtained successfully"
    echo "$access_token"
}

# Function to make API calls
api_call() {
    local method="$1"
    local endpoint="$2"
    local data="$3"
    local access_token="$4"

    if [ -n "$data" ]; then
        curl -s -X "$method" \
            "$ZITADEL_URL$endpoint" \
            -H "Authorization: Bearer $access_token" \
            -H "Content-Type: application/json" \
            -d "$data"
    else
        curl -s -X "$method" \
            "$ZITADEL_URL$endpoint" \
            -H "Authorization: Bearer $access_token"
    fi
}

# Delete existing PKCE application if it exists
delete_existing_app() {
    local access_token="$1"
    local project_id="339013429979250696"  # Our existing project ID
    local app_id="339013484052217864"      # Our existing PKCE app ID

    log_info "Deleting existing PKCE application..."

    local delete_response=$(curl -s -X DELETE \
        "$ZITADEL_URL/management/v1/projects/$project_id/apps/$app_id" \
        -H "Authorization: Bearer $access_token")

    log_info "Existing application deleted"
}

# Create username/password application
create_password_application() {
    local project_id="339013429979250696"  # Use existing project
    local access_token="$1"

    log_info "Creating username/password application '$APP_NAME'..."

    local app_response=$(api_call "POST" "/management/v1/projects/$project_id/apps/oidc" '{
        "name": "'$APP_NAME'",
        "redirectUris": [
            "http://localtest.me/login",
            "http://localhost:5173/login"
        ],
        "postLogoutRedirectUris": [
            "http://localtest.me/",
            "http://localhost:5173/"
        ],
        "responseTypes": ["OIDC_RESPONSE_TYPE_CODE"],
        "grantTypes": [
            "OIDC_GRANT_TYPE_AUTHORIZATION_CODE",
            "OIDC_GRANT_TYPE_REFRESH_TOKEN",
            "OIDC_GRANT_TYPE_PASSWORD"
        ],
        "appType": "OIDC_APP_TYPE_WEB",
        "authMethodType": "OIDC_AUTH_METHOD_TYPE_BASIC",
        "version": "OIDC_VERSION_1_0",
        "devMode": true,
        "accessTokenType": "OIDC_TOKEN_TYPE_JWT",
        "accessTokenRoleAssertion": false,
        "idTokenRoleAssertion": false,
        "idTokenUserinfoAssertion": false,
        "additionalOrigins": [
            "http://localtest.me",
            "http://localhost:5173"
        ]
    }' "$access_token")

    local app_id=$(echo "$app_response" | jq -r '.appId // empty')
    local client_id=$(echo "$app_response" | jq -r '.clientId // empty')
    local client_secret=$(echo "$app_response" | jq -r '.clientSecret // empty')

    if [ -z "$app_id" ] || [ "$app_id" = "null" ]; then
        log_error "Failed to create application"
        log_error "Response: $app_response"
        exit 1
    fi

    log_success "Application created successfully!"
    log_success "Application ID: $app_id"
    log_success "Client ID: $client_id"
    log_success "Client Secret: $client_secret"

    echo "$app_id|$client_id|$client_secret"
}

# Save configuration
save_configuration() {
    local project_id="339013429979250696"
    local app_id="$1"
    local client_id="$2"
    local client_secret="$3"

    log_info "Saving configuration..."

    cat > "$CONFIG_FILE" << EOF
{
    "project_id": "$project_id",
    "project_name": "$PROJECT_NAME",
    "app_id": "$app_id",
    "app_name": "$APP_NAME",
    "client_id": "$client_id",
    "client_secret": "$client_secret",
    "auth_type": "password",
    "token_url": "$ZITADEL_EXTERNAL_URL/oauth/v2/token",
    "userinfo_url": "$ZITADEL_EXTERNAL_URL/oidc/v1/userinfo",
    "scope": "openid profile email"
}
EOF

    # Mark setup as complete
    touch "$SETUP_COMPLETE_FILE"

    log_success "Configuration saved to $CONFIG_FILE"
}

# Main execution
main() {
    log_info "Starting ZITADEL username/password authentication setup..."

    # Check if already completed
    check_setup_complete

    # Check dependencies
    check_dependencies

    # Wait for services to be ready
    wait_for_zitadel
    wait_for_machine_key

    # Get authentication
    local access_token=$(get_access_token)

    # Delete existing PKCE app
    delete_existing_app "$access_token"

    # Create new password-based application
    local app_result=$(create_password_application "$access_token")
    local app_id=$(echo "$app_result" | cut -d'|' -f1)
    local client_id=$(echo "$app_result" | cut -d'|' -f2)
    local client_secret=$(echo "$app_result" | cut -d'|' -f3)

    # Save configuration
    save_configuration "$app_id" "$client_id" "$client_secret"

    log_success "✅ ZITADEL username/password authentication setup complete!"
    log_success "Project: $PROJECT_NAME"
    log_success "Application: $APP_NAME (ID: $app_id)"
    log_success "Client ID: $client_id"
    log_success "Client Secret: $client_secret"
    echo ""
    log_info "You can now use username/password authentication in your frontend application."
}

# Run main function
main "$@"
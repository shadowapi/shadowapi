#!/bin/bash
set -e

CREDENTIALS_FILE="/data/credentials.json"

# If arguments are passed, run the worker binary directly with those args
# This allows running `docker compose run --rm worker /bin/worker enroll ...`
if [ $# -gt 0 ]; then
    exec "$@"
fi

# Load credentials from persistent file
load_credentials() {
    if [ -f "$CREDENTIALS_FILE" ]; then
        # Validate JSON structure
        if jq -e '.worker_id and .worker_secret' "$CREDENTIALS_FILE" > /dev/null 2>&1; then
            export WORKER_ID=$(jq -r '.worker_id' "$CREDENTIALS_FILE")
            export WORKER_SECRET=$(jq -r '.worker_secret' "$CREDENTIALS_FILE")
            echo "Loaded credentials from $CREDENTIALS_FILE"
            return 0
        else
            echo "Warning: $CREDENTIALS_FILE exists but is invalid, ignoring"
            return 1
        fi
    fi
    return 1
}

# Save credentials to persistent file
save_credentials() {
    local worker_id="$1"
    local worker_secret="$2"
    local worker_name="${WORKER_NAME:-unknown}"

    # Ensure /data directory exists (should exist from volume mount)
    mkdir -p /data

    # Create JSON with jq for proper escaping
    jq -n \
        --arg id "$worker_id" \
        --arg secret "$worker_secret" \
        --arg name "$worker_name" \
        --arg time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
        '{worker_id: $id, worker_secret: $secret, worker_name: $name, enrolled_at: $time}' \
        > "$CREDENTIALS_FILE"

    # Restrict permissions (readable only by owner)
    chmod 600 "$CREDENTIALS_FILE"

    echo "Saved credentials to $CREDENTIALS_FILE"
}

# Priority 1: Environment variables take precedence (explicit override)
if [ -n "$WORKER_ID" ] && [ -n "$WORKER_SECRET" ]; then
    echo "Using credentials from environment variables"
# Priority 2: Load from persistent file
elif load_credentials; then
    : # Credentials loaded from file
# Priority 3: Auto-enroll if token provided
elif [ -n "$ENROLLMENT_TOKEN" ]; then
    WORKER_NAME=${WORKER_NAME:-"worker-$(hostname)"}
    echo "Auto-enrolling worker '$WORKER_NAME'..."

    # Run enrollment and capture output
    OUTPUT=$(/bin/worker enroll --token="$ENROLLMENT_TOKEN" --name="$WORKER_NAME" 2>&1) || {
        echo "Enrollment failed: $OUTPUT"
        exit 1
    }

    # Parse WORKER_ID and WORKER_SECRET from output
    export WORKER_ID=$(echo "$OUTPUT" | grep "Worker ID:" | awk '{print $3}')
    export WORKER_SECRET=$(echo "$OUTPUT" | grep "Worker Secret:" | awk '{print $3}')

    if [ -z "$WORKER_ID" ] || [ -z "$WORKER_SECRET" ]; then
        echo "Failed to parse credentials from enrollment output"
        echo "$OUTPUT"
        exit 1
    fi

    echo "Enrolled successfully as $WORKER_ID"

    # Save to persistent storage
    save_credentials "$WORKER_ID" "$WORKER_SECRET"
fi

# Final check: ensure we have credentials
if [ -z "$WORKER_ID" ] || [ -z "$WORKER_SECRET" ]; then
    echo "WORKER_ID and WORKER_SECRET not set. Worker is waiting for enrollment."
    echo "Options:"
    echo "  1. Set ENROLLMENT_TOKEN for auto-enrollment (credentials will persist)"
    echo "  2. Set WORKER_ID and WORKER_SECRET environment variables"
    echo "  3. Manually enroll and restart:"
    echo "     docker compose run --rm worker /bin/worker enroll --token=<token> --name=worker-1"
    # Keep container running but not processing
    sleep infinity
fi

exec /bin/worker connect

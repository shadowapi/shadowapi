#!/bin/bash
set -e

# If arguments are passed, run the worker binary directly with those args
# This allows running `docker compose run --rm grpc-worker /bin/worker enroll ...`
if [ $# -gt 0 ]; then
    exec "$@"
fi

# Auto-enroll if ENROLLMENT_TOKEN is set but credentials are missing
if [ -n "$ENROLLMENT_TOKEN" ] && { [ -z "$WORKER_ID" ] || [ -z "$WORKER_SECRET" ]; }; then
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
fi

# Default behavior: check for credentials and run connect
if [ -z "$WORKER_ID" ] || [ -z "$WORKER_SECRET" ]; then
    echo "WORKER_ID and WORKER_SECRET not set. Worker is waiting for enrollment."
    echo "Set ENROLLMENT_TOKEN for auto-enrollment, or manually enroll:"
    echo "  1. docker compose exec backend shadowapi create-enrollment-token --name worker-1"
    echo "  2. docker compose run --rm grpc-worker /bin/worker enroll --token=<token> --name=worker-1"
    echo "  3. Add WORKER_ID and WORKER_SECRET to .env and restart"
    # Keep container running but not processing
    sleep infinity
fi

exec /bin/worker connect

#!/bin/bash
set -e

# If arguments are passed, run the worker binary directly with those args
# This allows running `docker compose run --rm grpc-worker /bin/worker enroll ...`
if [ $# -gt 0 ]; then
    exec "$@"
fi

# Default behavior: check for credentials and run connect
if [ -z "$WORKER_ID" ] || [ -z "$WORKER_SECRET" ]; then
    echo "WORKER_ID and WORKER_SECRET not set. Worker is waiting for enrollment."
    echo "Run 'make up' to automatically enroll the worker, or manually enroll:"
    echo "  1. docker compose exec backend shadowapi create-enrollment-token --name worker-1"
    echo "  2. docker compose run --rm grpc-worker /bin/worker enroll --token=<token> --name=worker-1"
    echo "  3. Add WORKER_ID and WORKER_SECRET to .env and restart"
    # Keep container running but not processing
    sleep infinity
fi

exec /bin/worker connect

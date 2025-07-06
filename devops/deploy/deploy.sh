cd /opt/shadowapi
git pull
docker compose \
        -f compose.yaml \
        -f devops/demo/compose.yaml \
        --env-file .env \
        --env-file devops/demo/.env \
        up \
        db frontend nats backend \
        --wait --build

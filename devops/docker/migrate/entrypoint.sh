#!/bin/sh
set -e

cat /db/schema.sql /db/tg.sql > /tmp/combined.sql

atlas schema apply \
  --url "$DATABASE_URL" \
  --dev-url "$DEV_DATABASE_URL" \
  --to file:///tmp/combined.sql \
  --auto-approve

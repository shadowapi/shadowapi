#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
	CREATE DATABASE shadowapi_test WITH OWNER shadowapi;
	GRANT ALL PRIVILEGES ON DATABASE shadowapi_test TO shadowapi;

	CREATE DATABASE shadowapi_schema WITH OWNER shadowapi;
	GRANT ALL PRIVILEGES ON DATABASE shadowapi_schema TO shadowapi;
EOSQL

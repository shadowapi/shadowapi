#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
	CREATE DATABASE shadowapi_test WITH OWNER shadowapi;
	GRANT ALL PRIVILEGES ON DATABASE shadowapi_test TO shadowapi;

	CREATE DATABASE shadowapi_schema WITH OWNER shadowapi;
	GRANT ALL PRIVILEGES ON DATABASE shadowapi_schema TO shadowapi;

	CREATE USER zitadel WITH PASSWORD 'zitadel';
	CREATE DATABASE zitadel WITH OWNER zitadel;
	GRANT ALL PRIVILEGES ON DATABASE zitadel TO zitadel;

	-- Ory Hydra database
	CREATE DATABASE hydra WITH OWNER shadowapi;
	GRANT ALL PRIVILEGES ON DATABASE hydra TO shadowapi;
EOSQL

# ShadowAPI Makefile
# Run `make help` to see available targets

.PHONY: help init db-shell sync-db api-gen-backend api-gen-frontend api-gen \
        playwright-run playwright-report zitadel-init playwright-real local-certs

# Default target
.DEFAULT_GOAL := help

##@ General

help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Setup

init: ## Initialize the project (reset containers, copy env, start db, migrate)
	docker compose down -v
	cp -f .env.example .env
	docker compose up -d db
	$(MAKE) sync-db

##@ Database

db-shell: ## Open Postgres shell in the container
	docker compose exec db su - postgres -c "psql -U shadowapi shadowapi"

sync-db: ## Migrate database to latest changes using Atlas
	@cat ./db/schema.sql ./db/tg.sql > ./db/combined.sql
	@docker compose run --rm \
		-v ./db/combined.sql:/schema.sql \
		db-migrate \
		schema apply \
			--url postgres://shadowapi:shadowapi@db:5432/shadowapi?sslmode=disable \
			--dev-url postgres://shadowapi:shadowapi@db:5432/shadowapi_schema?sslmode=disable \
			--to file://schema.sql \
			--auto-approve
	@rm -f ./db/combined.sql

##@ Code Generation

api-gen-backend: ## Generate API specification in backend
	cd ./backend && go generate

api-gen-frontend: ## Generate TypeScript API client in frontend
	cd ./front && npm run generate-api-client

api-gen: api-gen-backend api-gen-frontend ## Generate API specs (backend & frontend)

##@ Testing

playwright-run: ## Run Playwright tests
	cd ./front && npx playwright test

playwright-report: ## Generate and open Playwright test report
	cd ./front && npx playwright show-report

playwright-real: ## Run Playwright tests against real Zitadel (no mocks)
	cd ./front && npx playwright test real-zitadel-login.test.ts --headed

##@ Zitadel

zitadel-init: ## Provision Zitadel OIDC app (optional)
	docker compose --profile init run --rm init

##@ Certificates

local-certs: ## Generate local SSL certificates
	openssl req -x509 -nodes -newkey rsa:2048 -days 365 \
		-keyout localtest.me.key -out localtest.me.crt \
		-subj "/CN=localtest.me"

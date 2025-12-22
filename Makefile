# ShadowAPI Makefile
# Run `make help` to see available targets

.PHONY: help up init db-shell sync-db api-gen-backend api-gen-frontend api-gen deploy deploy-auto deploy-no-migrate

# Default target
.DEFAULT_GOAL := help

##@ General

help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Setup

up: ## Bootstrap and start the full stack (first-time setup)
	./devops/bootstrap.sh

init: ## Initialize the project (reset containers, copy env, start db, migrate)
	docker compose down -v
	cp -f .env.template .env
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

##@ Deployment

deploy: ## Deploy to Uncloud (production) with migration approval
	./devops/uncloud/deploy.sh

deploy-auto: ## Deploy to Uncloud with auto-approval (CI/CD)
	./devops/uncloud/deploy.sh --yes

deploy-no-migrate: ## Deploy to Uncloud without migrations
	./devops/uncloud/deploy.sh --skip-migrations

# ShadowAPI Makefile
# Run `make help` to see available targets

.PHONY: help up init db-shell sync-db api-gen-backend api-gen-frontend api-gen deploy-uncloud

# Default target
.DEFAULT_GOAL := help

##@ General

help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Setup

up: ## Bootstrap and start the full stack (first-time setup)
	python3 ./devops/bootstrap.py

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

proto-gen: ## Generate protobuf Go code using buf
	cd ./backend/proto && buf generate

proto-lint: ## Lint protobuf files
	cd ./backend/proto && buf lint

sqlc-gen: ## Generate SQLC database code
	cd ./db && sqlc generate

##@ Worker

worker-build: ## Build worker binary locally
	cd ./backend && go build -o ./bin/worker ./cmd/worker

worker-enroll: ## Enroll a new worker (requires TOKEN and NAME)
	cd ./backend && ./bin/worker enroll --token=$(TOKEN) --name=$(NAME)

worker-logs: ## View distributed worker logs
	docker compose logs -f worker

##@ Secrets Management

# Age key file location (default: ~/.config/sops/age/keys.txt)
SOPS_AGE_KEY_FILE ?= $(HOME)/.config/sops/age/keys.txt
export SOPS_AGE_KEY_FILE

secrets-decrypt: ## Decrypt production .env.enc -> .env
	@sops --decrypt --input-type dotenv --output-type dotenv \
		devops/uncloud/.env.enc > devops/uncloud/.env
	@echo "Decrypted to devops/uncloud/.env"

secrets-encrypt: ## Encrypt production .env -> .env.enc
	@sops --encrypt --input-type dotenv --output-type dotenv \
		devops/uncloud/.env > devops/uncloud/.env.enc
	@echo "Encrypted to devops/uncloud/.env.enc"

secrets-edit: ## Edit encrypted secrets in-place
	@sops devops/uncloud/.env.enc

secrets-rotate: ## Re-encrypt after adding/removing team members in .sops.yaml
	@sops updatekeys devops/uncloud/.env.enc

##@ Deployment to Uncloud
uncloud-deploy: ## Deploy to Uncloud (production, with migrations)
	./devops/uncloud/migrate.py
	./devops/uncloud/deploy.py

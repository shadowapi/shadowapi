.PHONY: help up down clean build shell db-shell sync-db test-init-test-tables test-destroy-test-tables sqlc sqlc-vet api-gen test

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

# ---------- Dev ----------

up: ## Start dev environment (docker compose watch)
	docker compose watch

down: ## Stop dev environment
	docker compose down

clean: ## Remove all containers, volumes, images
	docker compose down -v --rmi all --remove-orphans

build: ## Rebuild containers without cache
	docker compose build --no-cache

shell: ## Open bash in backend container
	docker compose exec backend bash -l

db-shell: ## Open psql shell
	docker compose exec db psql -U shadowapi shadowapi

nats-run: ## Run NATS JetStream locally (foreground)
	docker run --rm --name sa-nats -p 4222:4222 -p 8222:8222 nats:2.11.1 -js --http_port=8222

nats-ui: ## Open NATS monitoring dashboard
	open http://127.0.0.1:8222

ssl-cert: ## Generate and install local SSL certs for shadowapi.local
	bash devops/scripts/install-macos-shadowapi.local-certs.sh

# ---------- Database ----------

sync-db: ## Apply schema changes via psql
	cat ./db/schema.sql ./db/tg.sql | docker compose exec -T db psql -U shadowapi shadowapi

test-init-test-tables: ## Create test_ tables and register "Test Internal" storage
	cat ./db/test_storage_tables.sql ./db/test_storage_register.sql | docker compose exec -T db psql -U shadowapi shadowapi

test-destroy-test-tables: ## Drop test_ tables and remove "Test Internal" storage
	cat ./db/test_storage_destroy.sql | docker compose exec -T db psql -U shadowapi shadowapi

# ---------- Code Generation ----------

sqlc-gen: ## Regenerate SQLC queries
	cd backend && go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.28.0 && cd ../db && sqlc generate

sqlc-vet: ## Validate SQL queries
	cd db && sqlc vet

api-gen: ## Generate API specs (backend + frontend orval SDK)
	cd backend && go generate
	cd frontend && npm run generate-api

# ---------- Test ----------

test: ## Run Go tests
	cd backend && go test ./...

test-login: ## Run login Playwright test (non-headless)
	node frontend/playwright/test-01-login.cjs

test-crud: ## Run CRUD Playwright test (non-headless)
	node frontend/playwright/test-02-crud.cjs

test-form-load: ## Run form data loading Playwright test (non-headless)
	node frontend/playwright/test-03-form-load.cjs

test-get-10-messages: ## Run Gmail fetch pipeline test (non-headless)
	node frontend/playwright/test-04-get-10-messages.cjs

test-pw: ## Run Playwright tests
	cd frontend && npx playwright test

test-pw-report: ## Open Playwright report
	cd frontend && npx playwright show-report

# ---------- Build ----------

build-binary: ## Build backend binary locally
	cd backend && go build -o ../bin/shadowapi ./cmd/shadowapi

build-prod: ## Build production Docker image
	docker build -t shadowapi -f devops/Dockerfile .

docker-build-push: ## Build, tag and push to GitHub Container Registry
	docker build -t shadowapi -f devops/Dockerfile .
	docker tag shadowapi ghcr.io/reactima/shadowapi:latest
	docker push ghcr.io/reactima/shadowapi:latest

# ---------- Versions ----------

versions: ## Check versions of host packages
	@echo "Go:" && (command -v go >/dev/null 2>&1 && go version || echo "go is not installed")
	@echo "Air:" && (command -v air >/dev/null 2>&1 && air -v || echo "air is not installed")
	@echo "Node:" && (command -v node >/dev/null 2>&1 && node -v || echo "node is not installed")
	@echo "npm:" && (command -v npm >/dev/null 2>&1 && npm -v || echo "npm is not installed")
	@echo "Docker:" && (command -v docker >/dev/null 2>&1 && docker --version || echo "docker is not installed")
	@echo "psql:" && (command -v psql >/dev/null 2>&1 && psql --version || echo "psql is not installed")

# ---------- Includes ----------

include Makefile.check-dev-servers
include Makefile.cli

.DEFAULT_GOAL := help

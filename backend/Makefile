air: ## Run backend with Air hot-reload (local)
	@lsof -i :8090 | awk 'NR>1 {print $$2}' | xargs kill -9 2>/dev/null || true
	SA_ENV=dev SA_CONFIG_PATH=config.local.yaml air

air-upgrade: ## Upgrade Air to latest
	go install github.com/air-verse/air@latest

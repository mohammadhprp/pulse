.PHONY: build start stop restart logs clean help

.DEFAULT_GOAL := help

## Available commands:
help:
	@echo "Available Commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Use 'make <command>' to run a command."

build: ## Build and start containers in detached mode
	@docker compose up -d --build

start: ## Start existing containers
	@docker compose up -d

stop: ## Stop and remove containers
	@docker compose down

restart: ## Restart containers (stop + start)
	@$(MAKE) stop
	@$(MAKE) start

logs: ## Tail container logs (add ctrl+c to stop)
	@docker compose logs -f agent collector

clean: ## Stop containers and remove volumes, images
	@docker compose down -v --rmi all
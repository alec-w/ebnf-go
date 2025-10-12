help:
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\033[36m\033[0m\n"} /^[a-zA-Z_\-/0-9]+:.*?##/ { printf " \033[36m%-40s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

build: ## Builds container for local dev environment
	docker compose build

up: ## Creates local dev environment
	docker compose up -d

shell: ## Opens shell in local dev environment
	docker compose exec dev bash

down: ## Deletes local dev environment
	docker compose down -v

# Local dev Makefile for the GO Feature Flag Management API.
# Run from cmd/management/.

SHELL := /bin/bash
.DEFAULT_GOAL := help

GOFF_MGMT_DB_URL ?= postgres://goff:goff@localhost:5432/goff_mgmt?sslmode=disable
KEYCLOAK_OIDC_URL ?= http://localhost:8081/realms/goff/.well-known/openid-configuration
ENV_FILE := $(if $(wildcard .env),.env,$(if $(wildcard .env.example),.env.example,))

.PHONY: help start stop restart status logs infra-up infra-down infra-restart \
        wait-postgres wait-keycloak \
        migrate-up migrate-down migrate-reset migrate-status \
        run build test swagger clean

help: ## Show this help.
	@awk 'BEGIN {FS = ":.*##"; printf "Targets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

# --- dev lifecycle ------------------------------------------------------------

start: infra-up wait-postgres wait-keycloak migrate-up run ## Bring infra up, migrate, then run the API.

stop: infra-down ## Stop docker infra.

restart: stop start ## Stop + start.

status: ## Show docker compose status.
	docker compose ps

logs: ## Tail docker logs (Ctrl-C to exit).
	docker compose logs -f

# --- infra --------------------------------------------------------------------

infra-up: ## docker compose up -d (postgres + keycloak).
	docker compose up -d

infra-down: ## docker compose down -v (wipes volumes).
	docker compose down -v

infra-restart: infra-down infra-up ## Recreate infra from scratch.

wait-postgres: ## Block until postgres accepts connections.
	@echo "==> Waiting for postgres"
	@for i in $$(seq 1 60); do \
		if docker compose exec -T postgres pg_isready -U goff -d goff_mgmt >/dev/null 2>&1; then \
			echo "    postgres is up"; exit 0; \
		fi; \
		sleep 1; \
	done; \
	echo "    postgres not reachable after 60s" >&2; exit 1

wait-keycloak: ## Block until Keycloak OIDC discovery responds 200.
	@echo "==> Waiting for keycloak ($(KEYCLOAK_OIDC_URL))"
	@for i in $$(seq 1 120); do \
		code=$$(curl -s -o /dev/null -w '%{http_code}' "$(KEYCLOAK_OIDC_URL)" || true); \
		if [ "$$code" = "200" ]; then \
			echo "    keycloak is up"; exit 0; \
		fi; \
		sleep 1; \
	done; \
	echo "    keycloak OIDC discovery not reachable after 120s" >&2; exit 1

# --- migrations ---------------------------------------------------------------

migrate-up: ## Apply all goose migrations.
	goose -dir ./db/migrations postgres "$(GOFF_MGMT_DB_URL)" up

migrate-down: ## Roll back the most recent migration.
	goose -dir ./db/migrations postgres "$(GOFF_MGMT_DB_URL)" down

migrate-reset: ## Roll back all migrations.
	goose -dir ./db/migrations postgres "$(GOFF_MGMT_DB_URL)" reset

migrate-status: ## Show migration status.
	goose -dir ./db/migrations postgres "$(GOFF_MGMT_DB_URL)" status

# --- app ----------------------------------------------------------------------

run: ## Run the API (loads $(ENV_FILE) first if present).
	@if [ -n "$(ENV_FILE)" ]; then \
		echo "==> Loading $(ENV_FILE) and running API"; \
		set -a && source $(ENV_FILE) && set +a && go run .; \
	else \
		go run .; \
	fi

build: ## Build binary into ./out/management.
	mkdir -p out
	CGO_ENABLED=0 go build -o out/management .

test: ## Run unit tests.
	go test ./...

swagger: ## Regenerate Swagger docs (requires `swag`).
	swag init -g main.go -o docs

clean: ## Remove build artifacts.
	rm -rf out docs

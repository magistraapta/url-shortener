# Load variables from .env if present (used for DATABASE_URL, etc.)
ifneq (,$(wildcard .env))
	include .env
	export
endif

MIGRATE        := migrate
MIGRATIONS_DIR := database/migrations

# Connection string used by golang-migrate.
# Defaults to a local, non-SSL Postgres. Override on the command line:
#   make migrate-up DB_URL=postgres://user:pass@host:5432/db?sslmode=disable
DB_URL ?= postgres://postgres:postgres@localhost:5432/url_short?sslmode=disable

.PHONY: help migrate-up migrate-down migrate-down-all migrate-create migrate-force migrate-version migrate-drop

help: ## Show this help
	@grep -hE '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

migrate-up: ## Apply all up migrations
	$(MIGRATE) -database "$(DB_URL)" -path $(MIGRATIONS_DIR) up

migrate-down: ## Roll back the last migration
	$(MIGRATE) -database "$(DB_URL)" -path $(MIGRATIONS_DIR) down 1

migrate-down-all: ## Roll back all migrations
	$(MIGRATE) -database "$(DB_URL)" -path $(MIGRATIONS_DIR) down -all

migrate-create: ## Create a new migration: make migrate-create name=add_users_table
	@test -n "$(name)" || (echo "Usage: make migrate-create name=<migration_name>" && exit 1)
	$(MIGRATE) create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

migrate-force: ## Force schema to a version (fix dirty state): make migrate-force version=1
	@test -n "$(version)" || (echo "Usage: make migrate-force version=<version>" && exit 1)
	$(MIGRATE) -database "$(DB_URL)" -path $(MIGRATIONS_DIR) force $(version)

migrate-version: ## Print the current migration version
	$(MIGRATE) -database "$(DB_URL)" -path $(MIGRATIONS_DIR) version

migrate-drop: ## Drop everything in the database
	$(MIGRATE) -database "$(DB_URL)" -path $(MIGRATIONS_DIR) drop -f

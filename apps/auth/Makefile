# The auth server and its database. `make` alone lists the targets; the root Makefile's db and auth wrap them.

# sql-migrate resolves the dirs in dbconfig.yml from the working directory, so it runs beside that file.
MIGRATE  := cd scripts/db && go tool sql-migrate
ENV      ?= development
SEED_ENV ?= development_seeds
STEPS    ?= 1

.PHONY: help run test migrate migrate-down migrate-status migration seed seed-down seed-status seeder

help:
	@grep -E '^[a-z-]+:.*## ' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  make %-14s %s\n", $$1, $$2}'

run: ## run the auth server on 127.0.0.1:7272 (needs CKA_SIM_AUTH_SECRET)
	go run ./cmd/server

test: ## vet and test
	go vet ./... && go test ./...

migrate: ## apply pending migrations
	$(MIGRATE) up -env=$(ENV)

migrate-down: ## revert the last migration (STEPS=n for more)
	$(MIGRATE) down -env=$(ENV) -limit=$(STEPS)

migrate-status: ## show applied and pending migrations
	$(MIGRATE) status -env=$(ENV)

migration: ## create a migration: make migration name=add_x
	$(MIGRATE) new -env=$(ENV) $(name)

seed: ## apply pending seeds
	$(MIGRATE) up -env=$(SEED_ENV)

seed-down: ## revert the last seed (STEPS=n for more)
	$(MIGRATE) down -env=$(SEED_ENV) -limit=$(STEPS)

seed-status: ## show applied and pending seeds
	$(MIGRATE) status -env=$(SEED_ENV)

seeder: ## create a seed: make seeder name=more_users
	$(MIGRATE) new -env=$(SEED_ENV) $(name)

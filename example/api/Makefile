include .env
export

SHELL := /bin/bash -o pipefail

PROJECT_WORKSPACE ?= blaze-api
PROJECT_NAME ?= blaze-api
DOCKER_COMPOSE := docker compose -p $(PROJECT_WORKSPACE) -f deploy/develop/docker-compose.yml
DOCKER_CONTAINER_IMAGE := github.com/geniusrabbit/${PROJECT_NAME}
DOCKER_CONTAINER_MUGRATE_IMAGE := ${DOCKER_CONTAINER_IMAGE}:migrate-latest
APP_TAGS ?= migrate,postgres,redis,kafka,nats,graphql,grpc,metrics,opentelemetry

include ../../deploy/build.mk


.PHONY: build-gql
build-gql: ## Build graphql server
	cd protocol/graphql && go run github.com/99designs/gqlgen

.PHONY: build
build: ## Build API application
	@echo "Build application"
	@rm -rf .build
	@$(call do_build,"cmd/api/main.go",api)

.PHONY: build-docker-dev
build-docker-dev: build
	@echo "Build develop docker image"
	DOCKER_BUILDKIT=${DOCKER_BUILDKIT} docker build -t ${DOCKER_CONTAINER_IMAGE} -f deploy/develop/api.Dockerfile  ../../

.PHONY: run-api
run-api: build-docker-dev ## Run API service by docker-compose
	@echo "Run API service ${DOCKER_SERVER_LISTEN}"
	$(DOCKER_COMPOSE) up api

.PHONY: announce-api
announce-api: ## Announce API service in internet
# https://theboroer.github.io/localtunnel-www/
	@echo "Test API server"
	curl https://loca.lt/mytunnelpassword
	@echo ""
	lt --port ${DOCKER_SERVER_LISTEN}
	# gotunnelme ${DOCKER_SERVER_LISTEN}

.PHONY: announce-run-api
announce-run-api: run-api announce-api ## Run and announce API service

.PHONY: stop
stop: ## Stop all services
	@echo "Stop all services"
	$(DOCKER_COMPOSE) stop

.PHONY: build-migrate
build-migrate:
	DOCKER_BUILDKIT=${DOCKER_BUILDKIT} docker build -t ${DOCKER_CONTAINER_MUGRATE_IMAGE} -f deploy/develop/migrate.Dockerfile ../../

.PHONY: migrate
migrate: build-migrate ## Migrate migrations and fixtures
	- $(DOCKER_COMPOSE) run --rm migration
	- $(DOCKER_COMPOSE) run --rm migration-fixtures

.PHONY: dev-migrate
dev-migrate: build-migrate ## Migrate development migrations and fixtures
	$(DOCKER_COMPOSE) run --rm dev-migration

.PHONY: db-cli
db-cli: ## Open development database
	$(DOCKER_COMPOSE) exec $(DOCKER_DATABASE_NAME) psql -U $(DATABASE_USER) $(DATABASE_DB)

#############################################################################
# Combined declarative Atlas "app" schema (alternative to the versioned
# per-domain Atlas envs in the root Makefile, and to migrate/migrate above).
# See docs/MIGRATIONS.md.
#############################################################################

ATLAS_DIALECT      ?= postgres
ATLAS_APP_IMAGE     := ${DOCKER_CONTAINER_IMAGE}:migrate-atlas-latest
ATLAS_APP_DIR       := migrations/atlas-app

.PHONY: atlas-app-assemble
atlas-app-assemble: ## Assemble the combined Atlas "app" schema from repository/*/schema.hcl + local schema.hcl
	./$(ATLAS_APP_DIR)/assemble-schema.sh

.PHONY: atlas-app-diff
atlas-app-diff: atlas-app-assemble ## Preview the SQL Atlas would apply for the combined app schema. Usage: make atlas-app-diff URL=postgres://...
	@test -n "$(URL)" || (echo "Usage: make atlas-app-diff URL=postgres://..."; exit 1)
	cd $(ATLAS_APP_DIR) && atlas schema diff --env app --var dialect=$(ATLAS_DIALECT) --to env://src --from "$(URL)"

.PHONY: atlas-app-apply
atlas-app-apply: atlas-app-assemble ## Apply the combined app schema directly to a real DB — no per-domain SQL migration files. Usage: make atlas-app-apply URL=postgres://... [AUTO_APPROVE=1]
	@test -n "$(URL)" || (echo "Usage: make atlas-app-apply URL=postgres://..."; exit 1)
	cd $(ATLAS_APP_DIR) && atlas schema apply --env app --var dialect=$(ATLAS_DIALECT) --url "$(URL)" $(if $(AUTO_APPROVE),--auto-approve,)

.PHONY: build-migrate-atlas
build-migrate-atlas: ## Build the Docker image that applies the combined Atlas app schema
	DOCKER_BUILDKIT=${DOCKER_BUILDKIT} docker build -t ${ATLAS_APP_IMAGE} -f deploy/develop/atlas-apply.Dockerfile ../../

.PHONY: migrate-atlas
migrate-atlas: build-migrate-atlas ## Apply the combined Atlas app schema via docker-compose (real DB, no SQL migration files)
	$(DOCKER_COMPOSE) run --rm migration-atlas-app

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' Makefile | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help

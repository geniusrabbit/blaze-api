include .env
export

include ../../deploy/build.mk

SHELL := /bin/bash -o pipefail

PROJECT_WORKSPACE ?= blaze-api
PROJECT_NAME ?= blaze-api
DOCKER_COMPOSE := docker-compose -p $(PROJECT_WORKSPACE) -f deploy/develop/docker-compose.yml
DOCKER_CONTAINER_IMAGE := github.com/geniusrabbit/${PROJECT_NAME}
DOCKER_CONTAINER_MUGRATE_IMAGE := ${DOCKER_CONTAINER_IMAGE}:migrate-latest


.PHONY: build-gql
build-gql: ## Build graphql server
	cd protocol/graphql && go run github.com/99designs/gqlgen

.PHONY: build-api
build-api: ## Build API application
	@echo "Build application"
	@rm -rf .build
	@$(call do_build,"cmd/api/main.go",api)

.PHONY: build-docker-dev
build-docker-dev: build-api
	echo "Build develop docker image"
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

.PHONY: dbcli
dbcli: ## Open development database
	$(DOCKER_COMPOSE) exec $(DOCKER_DATABASE_NAME) psql -U $(DATABASE_USER) $(DATABASE_DB)

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' Makefile | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help

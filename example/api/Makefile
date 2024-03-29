include .env
export

SHELL := /bin/bash -o pipefail

BUILD_GOOS ?= $(or ${DOCKER_DEFAULT_GOOS},linux)
BUILD_GOARCH ?= $(or ${DOCKER_DEFAULT_GOARCH},arm64)
# https://github.com/golang/go/wiki/MinimumRequirements#amd64
BUILD_GOAMD64 ?= $(or ${DOCKER_DEFAULT_GOAMD64},1)
BUILD_GOAMD64_LIST ?= $(or ${DOCKER_DEFAULT_GOAMD64_LIST},1)
BUILD_GOARM ?= $(or ${DOCKER_DEFAULT_GOARM},7)
BUILD_GOARM_LIST ?= $(or ${DOCKER_DEFAULT_BUILD_GOARM_LIST},${BUILD_GOARM})
BUILD_CGO_ENABLED ?= 0
DOCKER_BUILDKIT ?= 1

LOCAL_TARGETPLATFORM=${BUILD_GOOS}/${BUILD_GOARCH}
ifeq (${BUILD_GOARCH},arm)
	LOCAL_TARGETPLATFORM=${BUILD_GOOS}/${BUILD_GOARCH}/v${BUILD_GOARM}
endif

COMMIT_NUMBER ?= $(or ${DEPLOY_COMMIT_NUMBER},)
ifeq (${COMMIT_NUMBER},)
	COMMIT_NUMBER = $(shell git log -1 --pretty=format:%h)
endif

TAG_VALUE ?= $(or ${DEPLOY_TAG_VALUE},)
ifeq (${TAG_VALUE},)
	TAG_VALUE = $(shell git describe --exact-match --tags `git log -n1 --pretty='%h'`)
endif
ifeq (${TAG_VALUE},)
	TAG_VALUE = commit-${COMMIT_NUMBER}
endif

OS_LIST   ?= $(or ${DEPLOY_OS_LIST},linux darwin)
ARCH_LIST ?= $(or ${DEPLOY_ARCH_LIST},amd64 arm64 arm)
APP_TAGS  ?= $(or ${APP_BUILD_TAGS},postgres jaeger migrate)

define build_platform_list
	for os in $(OS_LIST); do \
		for arch in $(ARCH_LIST); do \
			if [ "$$os/$$arch" != "darwin/arm" ]; then \
				if [ "$$arch" = "arm" ]; then \
					for armv in $(BUILD_GOARM_LIST); do \
						i="$${os}/$${arch}/v$${armv}"; \
						echo -n "$${i},"; \
					done; \
				else \
					if [ "$$arch" = "amd64" ]; then \
						for amd64v in $(BUILD_GOAMD64_LIST); do \
							if [ "$$amd64v" == "1" ]; then \
								i="$${os}/$${arch}"; \
							else \
								i="$${os}/$${arch}/v$${amd64v}"; \
							fi; \
							echo -n "$${i},"; \
						done; \
					else \
						i="$${os}/$${arch}"; \
						echo -n "$${i},"; \
					fi; \
				fi; \
			fi; \
		done; \
	done;
endef

PROJECT_WORKSPACE ?= blaze-api
PROJECT_NAME ?= blaze-api
DOCKER_PLATFORM_LIST := $(shell $(call build_platform_list))
DOCKER_PLATFORM_LIST := $(shell echo $(DOCKER_PLATFORM_LIST) | sed 's/.$$//')
DOCKER_COMPOSE := docker-compose -p $(PROJECT_WORKSPACE) -f deploy/develop/docker-compose.yml
DOCKER_CONTAINER_IMAGE := github.com/geniusrabbit/${PROJECT_NAME}
DOCKER_CONTAINER_MUGRATE_IMAGE := ${DOCKER_CONTAINER_IMAGE}:migrate-latest

# ==============================================================================

define do_build
	@for os in $(OS_LIST); do \
		for arch in $(ARCH_LIST); do \
			if [ "$$os/$$arch" != "darwin/arm" ]; then \
				if [ "$$arch" = "arm" ]; then \
					for armv in $(BUILD_GOARM_LIST); do \
						echo "Build $$os/$$arch/v$$armv"; \
						GOOS=$$os GOARCH=$$arch CGO_ENABLED=${BUILD_CGO_ENABLED} GOARM=$$armv \
							go build \
								-ldflags "-s -w -X main.appVersion=`date -u +%Y%m%d` -X main.buildCommit=${COMMIT_NUMBER} -X main.buildVersion=${TAG_VALUE} -X main.buildDate=`date -u +%Y%m%d.%H%M%S`"  \
								-tags "${APP_TAGS}" -o .build/$$os/$$arch/v$$armv/$(2) $(1); \
					done; \
				else \
					if [ "$$arch" = "amd64" ]; then \
						for amd64v in $(BUILD_GOAMD64_LIST); do \
							if [ "$$amd64v" == "1" ]; then \
								echo "Build $$os/$$arch -> v1"; \
								GOOS=$$os GOARCH=$$arch CGO_ENABLED=${BUILD_CGO_ENABLED} GOAMD64=v$$amd64v \
									go build \
										-ldflags "-s -w -X main.appVersion=`date -u +%Y%m%d` -X main.buildCommit=${COMMIT_NUMBER} -X main.buildVersion=${TAG_VALUE} -X main.buildDate=`date -u +%Y%m%d.%H%M%S`"  \
										-tags "${APP_TAGS}" -o .build/$$os/$$arch/$(2) $(1); \
							else \
								echo "Build $$os/$$arch/v$$amd64v"; \
								GOOS=$$os GOARCH=$$arch CGO_ENABLED=${BUILD_CGO_ENABLED} GOAMD64=v$$amd64v \
									go build \
										-ldflags "-s -w -X main.appVersion=`date -u +%Y%m%d` -X main.buildCommit=${COMMIT_NUMBER} -X main.buildVersion=${TAG_VALUE} -X main.buildDate=`date -u +%Y%m%d.%H%M%S`"  \
										-tags "${APP_TAGS}" -o .build/$$os/$$arch/v$$amd64v/$(2) $(1); \
							fi; \
						done; \
					else \
						echo "Build $$os/$$arch"; \
						GOOS=$$os GOARCH=$$arch CGO_ENABLED=${BUILD_CGO_ENABLED} \
							go build \
								-ldflags "-s -w -X main.appVersion=`date -u +%Y%m%d` -X main.buildCommit=${COMMIT_NUMBER} -X main.buildVersion=${TAG_VALUE} -X main.buildDate=`date -u +%Y%m%d.%H%M%S`"  \
								-tags "${APP_TAGS}" -o .build/$$os/$$arch/$(2) $(1); \
					fi; \
				fi; \
			fi; \
		done; \
	done
endef

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

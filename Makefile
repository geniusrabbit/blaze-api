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

PROJECT_WORKSPACE ?= api-template
PROJECT_NAME ?= api-template-base
DOCKER_PLATFORM_LIST := $(shell $(call build_platform_list))
DOCKER_PLATFORM_LIST := $(shell echo $(DOCKER_PLATFORM_LIST) | sed 's/.$$//')
DOCKER_COMPOSE := docker-compose -p $(PROJECT_WORKSPACE) -f deploy/develop/docker-compose.yml
DOCKER_CONTAINER_IMAGE := github.com/geniusrabbit/${PROJECT_NAME}
DOCKER_CONTAINER_MUGRATE_IMAGE := ${DOCKER_CONTAINER_IMAGE}:migrate-latest

.PHONY: all
all: lint cover

.PHONY: lint
lint: golint

.PHONY: golint
golint:
	# golint -set_exit_status ./...
	golangci-lint run -v ./...

.PHONY: fmt
fmt: ## Run formatting code
	@echo "Fix formatting"
	@gofmt -w ${GO_FMT_FLAGS} $$(go list -f "{{ .Dir }}" ./...); if [ "$${errors}" != "" ]; then echo "$${errors}"; fi

.PHONY: test
test: ## Run unit tests
	go test -v -tags ${APP_TAGS} -race ./...

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: vendor
vendor:
	go mod vendor

.PHONY: cover
cover:
	@mkdir -p $(TMP_ETC)
	@rm -f $(TMP_ETC)/coverage.txt $(TMP_ETC)/coverage.html
	go test -race -coverprofile=$(TMP_ETC)/coverage.txt -coverpkg=./... ./...
	@go tool cover -html=$(TMP_ETC)/coverage.txt -o $(TMP_ETC)/coverage.html
	@echo
	@go tool cover -func=$(TMP_ETC)/coverage.txt | grep total
	@echo
	@echo Open the coverage report:
	@echo open $(TMP_ETC)/coverage.html

.PHONY: __eval_srcs
__eval_srcs:
	$(eval SRCS := $(shell find . -not -path 'bazel-*' -not -path '.tmp*' -name '*.go'))

.PHONY: generate-code
generate-code: ## Run codegeneration procedure
	@echo "Generate code"
	@go generate ./...

.PHONY: build-gql
build-gql: ## Build graphql server
	cd protocol/graphql && go run github.com/99designs/gqlgen
	# cd protocol/graphql && gqlgen

define do_build
	@for os in $(OS_LIST); do \
		for arch in $(ARCH_LIST); do \
			if [ "$$os/$$arch" != "darwin/arm" ]; then \
				if [ "$$arch" = "arm64" ]; then \
					echo "Build $$os/$$arch/v8"; \
					GOOS=$$os GOARCH=$$arch CGO_ENABLED=${BUILD_CGO_ENABLED} \
						go build \
							-ldflags "-s -w -X main.appVersion=`date -u +%Y%m%d` -X main.buildCommit=${COMMIT_NUMBER} -X main.buildVersion=${TAG_VALUE} -X main.buildDate=`date -u +%Y%m%d.%H%M%S`"  \
							-tags "${APP_TAGS}" -o .build/$$os/$$arch/v8/$(2) $(1); \
				else \
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
			fi; \
		done; \
	done
endef

.PHONY: build-api
build-api: ## Build API application
	@echo "Build application"
	@rm -rf .build
	@$(call do_build,"cmd/api/main.go",api)

.PHONY: build-docker-dev
build-docker-dev: build-api
	echo "Build develop docker image"
	DOCKER_BUILDKIT=${DOCKER_BUILDKIT} docker build -t ${DOCKER_CONTAINER_IMAGE} -f deploy/develop/api.Dockerfile .

.PHONY: run-api
run-api: build-docker-dev ## Run API service by docker-compose
	@echo "Run API service ${DOCKER_SERVER_LISTEN}"
	$(DOCKER_COMPOSE) up api

.PHONY: stop
stop: ## Stop all services
	@echo "Stop all services"
	$(DOCKER_COMPOSE) stop

.PHONY: build-migrate
build-migrate:
	DOCKER_BUILDKIT=${DOCKER_BUILDKIT} docker build -t ${DOCKER_CONTAINER_MUGRATE_IMAGE} -f deploy/develop/migrate.Dockerfile .

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
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help

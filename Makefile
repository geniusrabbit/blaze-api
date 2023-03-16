include .env
export

COMMIT_NUMBER ?= $(shell git log -1 --pretty=format:%h)
BUILD_VERSION ?= $(shell git describe --exact-match --tags $(git log -n1 --pretty='%h'))

ifeq ($(BUILD_VERSION),)
	BUILD_VERSION := commit-$(COMMIT_NUMBER)
endif

PROJECT_WORKSPACE := api-template

BUILD_GOOS ?= $(or ${DOCKER_DEFAULT_GOOS},linux)
BUILD_GOARCH ?= $(or ${DOCKER_DEFAULT_GOARCH},amd64)
BUILD_GOARM ?= 7
BUILD_CGO_ENABLED ?= 0
DOCKER_BUILDKIT ?= 1

LOCAL_TARGETPLATFORM=${BUILD_GOOS}/${BUILD_GOARCH}
ifeq (${BUILD_GOARCH},arm)
	LOCAL_TARGETPLATFORM=${BUILD_GOOS}/${BUILD_GOARCH}/v${BUILD_GOARM}
endif

APP_TAGS := "postgres jaeger migrate"

export GO111MODULE := on
export GOSUMDB := off
export GOFLAGS=-mod=mod
# Go 1.13 defaults to TLS 1.3 and requires an opt-out.  Opting out for now until certs can be regenerated before 1.14
# https://golang.org/doc/go1.12#tls_1_3
export GODEBUG := tls13=0

DOCKER_COMPOSE := docker-compose -p $(PROJECT_WORKSPACE) -f deploy/develop/docker-compose.yml
CONTAINER_IMAGE := github.com/geniusrabbit/api-template-base:latest
CONTAINER_MUGRATE_IMAGE := github.com/geniusrabbit/api-template-base:migrate-latest

OS_LIST = linux
ARCH_LIST = amd64 arm64 arm


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
	# cd protocol/graphql && go run github.com/99designs/gqlgen
	cd protocol/graphql && gqlgen

define do_build
	@for os in $(OS_LIST); do \
		for arch in $(ARCH_LIST); do \
			if [ "$$os/$$arch" != "darwin/arm" ]; then \
				echo "Build $$os/$$arch"; \
				GOOS=$$os GOARCH=$$arch CGO_ENABLED=${BUILD_CGO_ENABLED} GOARM=${BUILD_GOARM} \
					go build \
						-ldflags "-s -w -X main.appVersion=`date -u +%Y%m%d` -X main.buildCommit=${COMMIT_NUMBER} -X main.buildVersion=${TAG_VALUE} -X main.buildDate=`date -u +%Y%m%d.%H%M%S`"  \
						-tags ${APP_TAGS} -o .build/$$os/$$arch/$(2) $(1); \
				if [ "$$arch" = "arm" ]; then \
					mkdir -p .build/$$os/$$arch/v${BUILD_GOARM}; \
					mv .build/$$os/$$arch/$(2) .build/$$os/$$arch/v${BUILD_GOARM}/$(2); \
				fi \
			fi \
		done \
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
	DOCKER_BUILDKIT=${DOCKER_BUILDKIT} docker build -t ${CONTAINER_IMAGE} -f deploy/develop/api.Dockerfile .

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
	DOCKER_BUILDKIT=${DOCKER_BUILDKIT} docker build -t ${CONTAINER_MUGRATE_IMAGE} -f deploy/develop/migrate.Dockerfile .

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

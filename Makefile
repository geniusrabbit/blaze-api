APP_TAGS  ?= $(or ${APP_BUILD_TAGS},postgres jaeger migrate)

.DEPS:
	@echo "Install dependencies"
	go install github.com/NoahShen/gotunnelme
	go install go.uber.org/mock/mockgen@latest

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
	go test -v -tags "${APP_TAGS}" -race ./...

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

.PHONY: generate-code
generate-code: ## Run codegeneration procedure
	@echo "Generate code"
	@go generate ./...

.PHONY: build-gql
build-gql: ## Build graphql server
	@cd protocol/graphql && go run github.com/99designs/gqlgen
	@rm -rf server/graphql/generated
	@rm -rf server/graphql/resolvers

.PHONY: run-test-api
run-test-api: ## Run test api server
	@cd example/api && make run-api

.PHONY: build-test-api
build-test-api: ## Build test api server
	@cd example/api && make build-docker-dev

.PHONY: build-test-gql
build-test-gql: ## Build test graphql server
	@cd example/api && make build-gql

.PHONY: announce-test-api
announce-test-api: ## Run test api server with announce via tunnelme
	@cd example/api && make -j2 announce-run-api

#############################################################################
# Atlas migrations (alternative to golang-migrate, see docs/MIGRATIONS.md)
#############################################################################

ATLAS_DIALECT ?= postgres
ATLAS_IMAGE   ?= blaze-api-atlas

# No per-domain atlas.hcl project file is needed: each domain is just a
# "schema.hcl" living inside its own migrations/ directory (see
# docs/MIGRATIONS.md), so all Atlas config lives here instead.
ATLAS_DEV_URL_postgres := docker://postgres/16/dev?search_path=public
ATLAS_DEV_URL_mysql    := docker://mysql/8/dev
ATLAS_DEV_URL_sqlite   := sqlite://dev?mode=memory&_fk=1
ATLAS_DEV_URL           = $(ATLAS_DEV_URL_$(ATLAS_DIALECT))

.PHONY: atlas-diff
atlas-diff: ## Generate a new Atlas migration. Usage: make atlas-diff DIR=repository/rbac
	@test -n "$(DIR)" || (echo "Usage: make atlas-diff DIR=repository/rbac"; exit 1)
	cd $(DIR) && atlas migrate diff --to "file://migrations/schema.hcl" --dir "file://migrations" --dev-url "$(ATLAS_DEV_URL)" --format '{{ sql . "  " }}'

.PHONY: atlas-apply
atlas-apply: ## Apply Atlas migrations to a live DB. Usage: make atlas-apply DIR=repository/rbac URL=postgres://...
	@test -n "$(DIR)" || (echo "Usage: make atlas-apply DIR=repository/rbac URL=..."; exit 1)
	cd $(DIR) && atlas migrate apply --dir "file://migrations" --url "$(URL)"

.PHONY: atlas-status
atlas-status: ## Show Atlas migration status. Usage: make atlas-status DIR=repository/rbac URL=postgres://...
	@test -n "$(DIR)" || (echo "Usage: make atlas-status DIR=repository/rbac URL=..."; exit 1)
	cd $(DIR) && atlas migrate status --dir "file://migrations" --url "$(URL)"

.PHONY: atlas-build-docker
atlas-build-docker: ## Build the Atlas+Docker-CLI image used to run Atlas without local install
	docker build -t $(ATLAS_IMAGE) -f deploy/develop/atlas.Dockerfile .

.PHONY: atlas-docker-diff
atlas-docker-diff: atlas-build-docker ## Generate a migration via Docker. Usage: make atlas-docker-diff DIR=repository/rbac
	@test -n "$(DIR)" || (echo "Usage: make atlas-docker-diff DIR=repository/rbac"; exit 1)
	docker run --rm -v /var/run/docker.sock:/var/run/docker.sock -v "$(CURDIR):/repo" \
		-w /repo/$(DIR) $(ATLAS_IMAGE) migrate diff --to "file://migrations/schema.hcl" --dir "file://migrations" --dev-url "$(ATLAS_DEV_URL)" --format '{{ sql . "  " }}'

.PHONY: atlas-docker-apply
atlas-docker-apply: atlas-build-docker ## Apply migrations via Docker. Usage: make atlas-docker-apply DIR=repository/rbac URL=postgres://...
	@test -n "$(DIR)" || (echo "Usage: make atlas-docker-apply DIR=repository/rbac URL=..."; exit 1)
	docker run --rm -v /var/run/docker.sock:/var/run/docker.sock -v "$(CURDIR):/repo" \
		-w /repo/$(DIR) $(ATLAS_IMAGE) migrate apply --dir "file://migrations" --url "$(URL)"

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' Makefile | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help

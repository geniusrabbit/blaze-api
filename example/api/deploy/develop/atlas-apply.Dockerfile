# Declarative Atlas "app" schema applier for example/api.
#
# Unlike deploy/develop/migrate.Dockerfile (golang-migrate, versioned SQL
# files), this image bundles a single combined Atlas HCL schema — assembled
# at *build* time from every repository/<domain>/migrations/schema.hcl this
# service depends on, plus its own user/account schema.hcl — and applies it
# directly to a live database with `atlas schema apply`. No per-domain SQL
# migration files are generated or shipped in this image. See
# docs/MIGRATIONS.md and example/api/migrations/atlas-app/assemble-schema.sh.
#
# Build context is the repository root (see example/api/Makefile's
# build-migrate-atlas target), so the relative paths assemble-schema.sh
# uses to reach repository/<domain>/migrations/schema.hcl resolve the same
# way here as they do when run directly from a checkout.
FROM arigaio/atlas:latest-alpine

LABEL maintainer="Dmitry Ponomarev <demdxx@gmail.com>"
LABEL service.name=geniusrabbit.api-template.atlas-app

USER root
RUN apk add --no-cache bash docker-cli

WORKDIR /repo

COPY repository/historylog/migrations/schema.hcl repository/historylog/migrations/schema.hcl
COPY repository/option/migrations/schema.hcl repository/option/migrations/schema.hcl
COPY repository/directaccesstoken/migrations/schema.hcl repository/directaccesstoken/migrations/schema.hcl
COPY repository/authclient/migrations/schema.hcl repository/authclient/migrations/schema.hcl
COPY repository/socialaccount/migrations/schema.hcl repository/socialaccount/migrations/schema.hcl
COPY example/api/migrations/atlas/migrations/schema.hcl example/api/migrations/atlas/migrations/schema.hcl
COPY example/api/migrations/atlas-app/ example/api/migrations/atlas-app/

WORKDIR /repo/example/api/migrations/atlas-app
RUN ./assemble-schema.sh

ENTRYPOINT ["atlas"]

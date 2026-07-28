# Atlas CLI + Docker client, for running Atlas without a local install.
#
# The schema is described directly in Atlas HCL (schema.hcl inside each
# domain's migrations/ directory — no per-domain atlas.hcl project file, see
# docs/MIGRATIONS.md), so no Go toolchain is needed here. The Docker client
# (docker-cli) is required so Atlas's `docker://` dev-url driver (used by
# the root Makefile's atlas-diff/atlas-apply targets) can spin up ephemeral
# Postgres/MySQL containers for schema diffing.
#
# This image is a dev/ops tool only: it is never referenced by, or built
# into, the service image (example/api/deploy/develop/api.Dockerfile). The
# repository is bind-mounted in at runtime (see `make atlas-docker-diff` /
# `atlas-docker-apply` in the root Makefile) rather than copied in at build
# time, so the image never needs rebuilding when schema.hcl/migrations
# change. See docs/MIGRATIONS.md for usage.
FROM arigaio/atlas:latest-alpine

LABEL maintainer="Dmitry Ponomarev <demdxx@gmail.com>"
LABEL service.name=geniusrabbit.api-template.atlas

USER root
RUN apk add --no-cache docker-cli

WORKDIR /repo

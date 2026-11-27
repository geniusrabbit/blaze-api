FROM alpine:latest

ARG TARGETPLATFORM

EXPOSE 8080 6060

LABEL maintainer="Dmitry Ponomarev <demdxx@gmail.com>"
LABEL service.name=geniusrabbit.api-template

ENV SERVER_HTTP_LISTEN=:8080
ENV SERVER_PROFILE_MODE=net
# Development image: pprof is published on 6060 by docker-compose.
# In production keep it on localhost or turn SERVER_PROFILE_MODE off.
ENV SERVER_PROFILE_LISTEN=:6060
ENV MIGRATIONS_DIR=/data/migrations

COPY example/api/.build/${TARGETPLATFORM}/api /api
COPY ./deploy/migrations /data/migrations

## Add all traits migrations to the image
RUN rm -rf /data/migrations/traits && mkdir -p /data/migrations/traits
COPY ./deploy/migrations/traits/account.up.sql /data/migrations/traits/001_account.up.sql
COPY ./deploy/migrations/traits/user_email.up.sql /data/migrations/traits/002_user_email.up.sql
COPY ./deploy/migrations/traits/user_password.up.sql /data/migrations/traits/003_user_password.up.sql
# COPY ./deploy/migrations/traits/user_username.up.sql /data/migrations/traits/004_user_username.up.sql

ENTRYPOINT [ "/api" ]

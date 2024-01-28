FROM migrate/migrate:latest

LABEL maintainer="Dmitry Ponomarev <demdxx@gmail.com>"

ADD deploy/migrations /migrations

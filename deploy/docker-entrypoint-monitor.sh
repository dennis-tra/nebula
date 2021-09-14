#!/usr/bin/env sh

# waiting for initial crawler to fill in DB
sleep 10m

set -e

: "${NEBULA_DATABASE_HOST:=0.0.0.0}"
: "${NEBULA_DATABASE_PORT:=5432}"
: "${NEBULA_DATABASE_NAME:=nebula}"
: "${NEBULA_DATABASE_USER:=nebula}"
: "${NEBULA_DATABASE_PASSWORD:=password}"

migrate -database "postgres://$NEBULA_DATABASE_USER:$NEBULA_DATABASE_PASSWORD@$NEBULA_DATABASE_HOST:$NEBULA_DATABASE_PORT/$NEBULA_DATABASE_NAME?sslmode=disable" -path migrations up

exec "$@"

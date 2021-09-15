#!/usr/bin/env sh

set -e

: "${NEBULA_DATABASE_HOST:=0.0.0.0}"
: "${NEBULA_DATABASE_PORT:=5432}"
: "${NEBULA_DATABASE_NAME:=nebula}"
: "${NEBULA_DATABASE_USER:=nebula}"
: "${NEBULA_DATABASE_PASSWORD:=password}"
: "${NEBULA_CRAWL_PERIOD:=1800}"
: "${NEBULA_SAVE_NEIGHBOURS:=1}"
: "${NEBULA_NOT_TRUNCATE_NEIGHBOURS:=0}"
: "${NEBULA_SAVE_CONNECTIONS:=1}"
: "${NEBULA_NOT_TRUNCATE_CONNECTIONS:=0}"

migrate -database "postgres://$NEBULA_DATABASE_USER:$NEBULA_DATABASE_PASSWORD@$NEBULA_DATABASE_HOST:$NEBULA_DATABASE_PORT/$NEBULA_DATABASE_NAME?sslmode=disable" -path migrations up

period=$NEBULA_CRAWL_PERIOD

monitor=0
while true; do
    start=$(date +%s)
    nebula --prom-port=6666 crawl
    if [ $monitor -eq 0 ]
    then
        nebula --prom-port=6667 monitor &
        monitor=1
    fi
    end=$(date +%s)
    echo $(date)
    d=$((end - start))
    sleep $((period - d))
done
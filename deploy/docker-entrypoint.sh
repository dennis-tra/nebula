#!/usr/bin/env sh

set -e

: "${NEBULA_DATABASE_HOST:=0.0.0.0}"
: "${NEBULA_DATABASE_PORT:=5432}"
: "${NEBULA_DATABASE_NAME:=nebula}"
: "${NEBULA_DATABASE_USER:=nebula}"
: "${NEBULA_DATABASE_PASSWORD:=password}"
: "${CRAWL_PERIOD:=1800}"
: "${CRAWL_NEIGHBOUR_SAVE:=1}"
: "${CRAWL_NEIGHBOUR_FREQUENCY:=48}"
: "${CRAWL_NEIGHBOUR_TRUNCATE:=1}"

migrate -database "postgres://$NEBULA_DATABASE_USER:$NEBULA_DATABASE_PASSWORD@$NEBULA_DATABASE_HOST:$NEBULA_DATABASE_PORT/$NEBULA_DATABASE_NAME?sslmode=disable" -path migrations up

period=$CRAWL_PERIOD
save=$CRAWL_NEIGHBOUR_SAVE
freq=$CRAWL_NEIGHBOUR_FREQUENCY
truncate=$CRAWL_NEIGHBOUR_TRUNCATE

monitor=0
counter=$CRAWL_NEIGHBOUR_FREQUENCY
while true; do
    start=$(date +%s)
    if [ $save -eq 1 ]
    then 
        if [ $counter -eq $freq ]
        then
            if [ $truncate -eq 1 ]
            then
                nebula --prom-port=6666 crawl --save-neighbours --truncate-neighbours
            else
                nebula --prom-port=6666 crawl --save-neighbours
            fi
            counter=1
        else
            nebula --prom-port=6666 crawl
            counter=$((counter + 1))
        fi
    else
        nebula --prom-port=6666 crawl
    fi
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
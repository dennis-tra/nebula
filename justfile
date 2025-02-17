# clickhouse variables
clickhouse_container_prefix := "nebula-clickhouse-"
clickhouse_image := "clickhouse/clickhouse-server:24.12"
clickhouse_user_prefix := "nebula_"
clickhouse_dbname_prefix := "nebula_"
clickhouse_pass_prefix := "password_"

# postgres variables
postgres_container_prefix := "nebula-postgres-"
postgres_image := "postgres:14"
postgres_user_prefix := "nebula_"
postgres_dbname_prefix := "nebula_"
postgres_pass_prefix := "password_"

# lists all available recipes
default:
    @just --list --justfile {{ justfile() }}

migrate-up env="local":
    migrate -database 'clickhouse://localhost:{{ if env == "local" { "9000" } else { "9001" } }}?username={{clickhouse_user_prefix}}{{env}}&database={{clickhouse_dbname_prefix}}{{env}}&password={{clickhouse_pass_prefix}}{{env}}&x-multi-statement=true' -path db/migrations/ch up

migrate-down env="local":
    migrate -database 'clickhouse://localhost:{{ if env == "local" { "9000" } else { "9001" } }}?username={{clickhouse_user_prefix}}{{env}}&database={{clickhouse_dbname_prefix}}{{env}}&password={{clickhouse_pass_prefix}}{{env}}&x-multi-statement=true' -path db/migrations/ch down

# start a clickhouse server
start-clickhouse env="local" detached="true":
    @echo "Starting ClickHouse server..."
    docker run --rm {{ if detached == "true" { "-d" } else { "" } }} \
        --name {{clickhouse_container_prefix}}{{env}} \
        -p {{ if env == "local" { "8123" } else { "8124" } }}:8123 \
        -p {{ if env == "local" { "9000" } else { "9001" } }}:9000 \
        -e CLICKHOUSE_DB={{clickhouse_dbname_prefix}}{{env}} \
        -e CLICKHOUSE_USER={{clickhouse_user_prefix}}{{env}} \
        -e CLICKHOUSE_PASSWORD={{clickhouse_pass_prefix}}{{env}} {{clickhouse_image}} > /dev/null 2>&1 || true

    @echo "Waiting for ClickHouse to become ready..."
    @while ! docker exec {{clickhouse_container_prefix}}{{env}} clickhouse-client --host=127.0.0.1 --query="SELECT 1" > /dev/null 2>&1; do sleep 1; done
    @echo "ClickHouse is ready!"

# stop and clean up the clickhouse server
stop-clickhouse env="local":
    @echo "Stopping and removing ClickHouse server container {{clickhouse_container_prefix}}{{env}}..."
    -@docker stop {{clickhouse_container_prefix}}{{env}} >/dev/null 2>&1
    -@docker rm {{clickhouse_container_prefix}}{{env}} >/dev/null 2>&1

# start a postgres server
start-postgres env="local" detached="true":
    @echo "Starting Postgres server..."
    docker run --rm {{ if detached == "true" { "-d" } else { "" } }} \
        --name {{postgres_container_prefix}}{{env}} \
        -p {{ if env == "local" { "5432" } else { "5433" } }}:5432 \
        -e POSTGRES_DB={{postgres_dbname_prefix}}{{env}} \
        -e POSTGRES_USER={{postgres_user_prefix}}{{env}} \
        -e POSTGRES_PASSWORD={{postgres_pass_prefix}}{{env}} {{postgres_image}} > /dev/null 2>&1

    @echo "Waiting for Postgres to become ready..."
    @while ! docker exec {{postgres_container_prefix}}{{env}} pg_isready > /dev/null 2>&1; do sleep 1; done
    @echo "Postgres is ready!"

# stop and clean up the postgres server
stop-postgres env="local":
    @echo "Stopping and removing Postgres server container {{postgres_container_prefix}}{{env}}..."
    -@docker stop {{postgres_container_prefix}}{{env}} >/dev/null 2>&1
    -@docker rm {{postgres_container_prefix}}{{env}} >/dev/null 2>&1

migrate-postgres dir env="local":
	migrate -database 'postgres://{{postgres_user_prefix}}{{env}}:{{postgres_pass_prefix}}{{env}}@localhost:{{ if env == "local" { "5432" } else { "5433" } }}/{{postgres_dbname_prefix}}{{env}}?sslmode=disable' -path db/migrations/pg {{dir}}

migrate-clickhouse dir env="local":
	migrate -database 'clickhouse://localhost:{{ if env == "local" { "9000" } else { "9001" } }}?username={{clickhouse_user_prefix}}{{env}}&database={{clickhouse_dbname_prefix}}{{env}}&password={{clickhouse_pass_prefix}}{{env}}&x-multi-statement=true' -path db/migrations/ch {{dir}}

test:
    #!/usr/bin/env bash
    just stop-postgres test
    just start-postgres test
    just migrate-postgres up test

    just stop-clickhouse test
    just start-clickhouse test
    just migrate-clickhouse up test

    just test-plain
    exit_code=$?

    just stop-postgres test
    just stop-clickhouse test

    exit $exit_code

test-plain:
    go test ./...
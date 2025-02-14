# clickhouse variables
clickhouse_container_prefix := "nebula-clickhouse-"
clickhouse_image := "clickhouse/clickhouse-server:24.12"
clickhouse_user_prefix := "nebula_"
clickhouse_dbname_prefix := "nebula_"
clickhouse_pass_prefix := "password_"

# lists all available recipes
default:
    @just --list --justfile {{ justfile() }}

migrate-up env="local":
    migrate -database 'clickhouse://localhost:{{ if env == "local" { "9000" } else { "9001" } }}?username={{clickhouse_user_prefix}}{{env}}&database={{clickhouse_dbname_prefix}}{{env}}&password={{clickhouse_pass_prefix}}{{env}}&x-multi-statement=true' -path db/migrations/ch up

migrate-down env="local":
    migrate -database 'clickhouse://localhost:{{ if env == "local" { "9000" } else { "9001" } }}?username={{clickhouse_user_prefix}}{{env}}&database={{clickhouse_dbname_prefix}}{{env}}&password={{clickhouse_pass_prefix}}{{env}}&x-multi-statement=true' -path db/migrations/ch down

# start a clickhouse server (user=test, pw=test, db=test, port=8123, secure=false)
start-clickhouse env="local" detached="true":
    @echo "Starting ClickHouse server..."
    docker run --rm {{ if detached == "true" { "-d" } else { "" } }} \
        --name {{clickhouse_container_prefix}}{{env}} \
        -p {{ if env == "local" { "8123" } else { "8124" } }}:8123 \
        -p {{ if env == "local" { "9000" } else { "9001" } }}:9000 \
        -e CLICKHOUSE_DB={{clickhouse_dbname_prefix}}{{env}} \
        -e CLICKHOUSE_USER={{clickhouse_user_prefix}}{{env}} \
        -e CLICKHOUSE_PASSWORD={{clickhouse_pass_prefix}}{{env}} {{clickhouse_image}}

    @echo "Waiting for ClickHouse to become ready..."
    @while ! docker exec {{clickhouse_container_prefix}}{{env}} clickhouse-client --host=127.0.0.1 --query="SELECT 1" >/dev/null 2>&1; do sleep 1; done
    @echo "ClickHouse is ready!"

# stop and clean up the clickhouse server
stop-clickhouse env="local":
    @echo "Stopping and removing ClickHouse server container {{clickhouse_container_prefix}}{{env}}..."
    -@docker stop {{clickhouse_container_prefix}}{{env}} >/dev/null 2>&1
    -@docker rm {{clickhouse_container_prefix}}{{env}} >/dev/null 2>&1

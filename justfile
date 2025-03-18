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

COMMIT := `git rev-parse --short HEAD`
DATE := `date "+%Y-%m-%dT%H:%M:%SZ"`
USER := `id -un`
VERSION := `git describe --tags --abbrev=0 || true`

# lists all available recipes
default:
    @just --list --justfile {{ justfile() }}

# build a nebula executable to the ./dist/ folder
build:
	go build -ldflags "-X main.version={{VERSION}} -X main.commit={{COMMIT}} -X main.date={{DATE}} -X main.builtBy={{USER}}" -o dist/nebula github.com/dennis-tra/nebula-crawler/cmd/nebula

# build a nebula docker image
docker platform="linux/amd64":
	docker build --platform {{platform}} \
	  --build-arg VERSION={{VERSION}} \
	  --build-arg COMMIT={{COMMIT}} \
	  --build-arg BUILT_BY={{USER}} \
	  --build-arg DATE={{DATE}} \
	  -t dennistra/nebula:{{COMMIT}} \
	  .

# push a nebula image to docker hub
docker-push: docker
	docker push dennistra/nebula:{{COMMIT}}

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

# restarts the nebula clickhouse container
restart-clickhouse env="local": (stop-clickhouse env) (start-clickhouse env)

# start a postgres server
start-postgres env="local" detached="true":
    @echo "Starting Postgres server..."
    docker run --rm {{ if detached == "true" { "-d" } else { "" } }} \
        --name {{postgres_container_prefix}}{{env}} \
        -p {{ if env == "local" { "5432" } else { "5433" } }}:5432 \
        -e POSTGRES_DB={{postgres_dbname_prefix}}{{env}} \
        -e POSTGRES_USER={{postgres_user_prefix}}{{env}} \
        -e POSTGRES_PASSWORD={{postgres_pass_prefix}}{{env}} {{postgres_image}} > /dev/null 2>&1 || true

    @echo "Waiting for Postgres to become ready..."
    @while ! docker exec {{postgres_container_prefix}}{{env}} pg_isready > /dev/null 2>&1; do sleep 1; done
    @echo "Postgres is ready!"

# stop and clean up the postgres server
stop-postgres env="local":
    @echo "Stopping and removing Postgres server container {{postgres_container_prefix}}{{env}}..."
    -@docker stop {{postgres_container_prefix}}{{env}} >/dev/null 2>&1
    -@docker rm {{postgres_container_prefix}}{{env}} >/dev/null 2>&1

# restarts the nebula postgres container
restart-postgres env="local": (stop-postgres env) (start-postgres env)

# applies postgres migrations (up or down)
migrate-postgres dir env="local":
	migrate -database 'postgres://{{postgres_user_prefix}}{{env}}:{{postgres_pass_prefix}}{{env}}@localhost:{{ if env == "local" { "5432" } else { "5433" } }}/{{postgres_dbname_prefix}}{{env}}?sslmode=disable' -path db/migrations/pg {{dir}}

# applies clickhouse migrations (up or down)
migrate-clickhouse dir env="local":
	migrate -database 'clickhouse://localhost:{{ if env == "local" { "9000" } else { "9001" } }}?username={{clickhouse_user_prefix}}{{env}}&database={{clickhouse_dbname_prefix}}{{env}}&password={{clickhouse_pass_prefix}}{{env}}' -path db/migrations/chlocal {{dir}}

# generates postgres models with sqlboiler
models: (restart-postgres "test") (migrate-postgres "up" "test")
    sqlboiler --no-tests psql
    just stop-postgres test

clean: stop-clickhouse stop-postgres
    rm -r dist || true

# runs all tests together with required databases
test: (restart-postgres "test") (restart-clickhouse "test")
    #!/usr/bin/env bash
    just test-plain
    exit_code=$?

    just stop-postgres test
    just stop-clickhouse test

    exit $exit_code

# runs all tests assuming the databases are running
test-plain:
	# maxmind excluded because it requires a database
	go test `go list ./... | grep -v maxmind`

# installs all necessary tools for this repository
tools:
	go install -tags 'postgres,clickhouse' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.2
	go install github.com/volatiletech/sqlboiler/v4@v4.18.0
	go install github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-psql@v4.18.0
	go install go.uber.org/mock/mockgen@v0.5.0
	go install mvdan.cc/gofumpt@v0.7.0

# starts a crawl for the given network and stores results in a clickhouse database
crawl-clickhouse network neighbors="true": (start-clickhouse "local")
    go run ./cmd/nebula --db-engine clickhouse --db-password password_local --db-user nebula_local --db-name nebula_local crawl --network {{network}} --neighbors={{neighbors}}

# starts a REPL in a running clickhouse instance
repl-clickhouse env="local": (start-clickhouse env)
    docker exec -it {{clickhouse_container_prefix}}{{env}} clickhouse-client -d nebula_{{env}}

# starts a REPL in a running postgres instance
repl-postgres env="local": (start-postgres env)
     PGPASSWORD=password_{{env}} docker exec -it {{postgres_container_prefix}}{{env}} psql -d nebula_{{env}} -U nebula_{{env}}

# starts a prometheus server that monitors a locally running nebula crawler
start-prometheus detached="true":
    docker run --rm {{ if detached == "true" { "-d" } else { "" } }} --name nebula-prometheus -p 9090:9090 -v $(pwd)/config/prometheus.yml:/etc/prometheus/prometheus.yml prom/prometheus

# stops a locally running prometheus server
stop-prometheus:
    @echo "Stopping and removing prometheus server container nebula-prometheus..."
    -@docker stop nebula-prometheus >/dev/null 2>&1
    -@docker rm nebula-prometheus >/dev/null 2>&1

# restarts a local prometheus server
restart-prometheus: stop-prometheus
    just start-prometheus

# generates clickhouse migrations which work with a local docker deployment
generate-local-clickhouse-migrations:
	#!/usr/bin/env bash
	OUTDIR=db/migrations/chlocal
	mkdir -p $OUTDIR
	for file in $(find db/migrations/chcluster -maxdepth 1 -name "*.sql"); do
	  filename=$(basename $file)
	  echo "Generating $OUTDIR/$filename"

	  # The "Replicated" variants don't work with a singular clickhouse deployment
	  # We're stripping that part from the file
	  sed 's/Replicated//' $file > $OUTDIR/$filename.tmp_0

	  # Enabling the JSON type is also different in both environments
	  # allow_experimental_json_type in ClickHouse Cloud
	  # enable_json_type locally
	  sed 's/allow_experimental_json_type/enable_json_type/' $OUTDIR/$filename.tmp_0 > $OUTDIR/$filename.tmp_1

	  # Add a warning message to the top of the file
	  cat <(echo -e "-- DO NOT EDIT: This file was generated with: just generate-local-clickhouse-migrations\n") $OUTDIR/$filename.tmp_1 > $OUTDIR/$filename
	  rm $OUTDIR/$filename.tmp*
	done

# generates go mocks
mocks:
    mockgen -source=libp2p/driver_crawler.go -destination=libp2p/mock_host_test.go -package=libp2p

# formats the entire repository
format:
	gofumpt -w -l .

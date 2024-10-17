GIT_SHA := $(shell git rev-parse --short HEAD)
DATE := $(shell date "+%Y-%m-%dT%H:%M:%SZ")
USER := $(shell id -un)
VERSION := $(shell git describe --tags --abbrev=0)

all: clean build

test:
	# maxmind excluded because it requires a database
	# discvx excluded because the tests take quite long and are copied from the prysm codebase
	go test `go list ./... | grep -v maxmind | grep -v discvx`

build:
	go build -ldflags "-X main.version=${VERSION} -X main.commit=${GIT_SHA} -X main.date=${DATE} -X main.builtBy=${USER}" -o dist/nebula github.com/dennis-tra/nebula-crawler/cmd/nebula

format:
	gofumpt -w -l .

clean:
	rm -r dist || true

docker:
	docker build -t dennistra/nebula:latest -t dennistra/nebula:${GIT_SHA} .

docker-linux:
	docker build --platform linux/amd64 -t 019120760881.dkr.ecr.us-east-1.amazonaws.com/probelab:nebula-sha${GIT_SHA} .

docker-push: docker-linux
	docker push dennistra/nebula:latest dennistra/nebula:${GIT_SHA}

tools:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.15.2
	go install github.com/volatiletech/sqlboiler/v4@v4.14.1
	go install github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-psql@v4.14.1
	go install go.uber.org/mock/mockgen@v0.3.0

database-reset: database-stop databased migrate-up models

database:
	docker run --rm -p 5432:5432 -e POSTGRES_PASSWORD=password -e POSTGRES_USER=nebula_test -e POSTGRES_DB=nebula_test --name nebula_test_db postgres:14

databased:
	docker run --rm -d -p 5432:5432 -e POSTGRES_PASSWORD=password -e POSTGRES_USER=nebula_test -e POSTGRES_DB=nebula_test --name nebula_test_db postgres:14
	sleep 1

database-stop:
	docker stop nebula_test_db || true

models:
	sqlboiler --no-tests psql

migrate-up:
	migrate -database 'postgres://nebula_test:password@localhost:5432/nebula_test?sslmode=disable' -path db/migrations up

migrate-down:
	migrate -database 'postgres://nebula_test:password@localhost:5432/nebula_test?sslmode=disable' -path db/migrations down

.PHONY: all clean test format tools models migrate-up migrate-down

default: all

all: clean build

test:
	go test `go list ./... | grep -v maxmind`

build:
	go build -ldflags "-X main.RawVersion=`cat version`" -o dist/nebula github.com/dennis-tra/nebula-crawler/cmd/nebula

build-linux:
	GOOS=linux GOARCH=amd64 make build

format:
	gofumpt -w -l .

clean:
	rm -r dist || true

docker:
	docker build -t dennistra/nebula:latest -t dennistra/nebula:`cat version` .

docker-push: docker
	docker push dennistra/nebula:latest dennistra/nebula:`cat version`

tools:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.15.2
	go install github.com/volatiletech/sqlboiler/v4@v4.13.0
	go install github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-psql@v4.13.0

database-reset: database-stop databased migrate-up models

database:
	docker run --rm -p 2345:5432 -e POSTGRES_PASSWORD=password -e POSTGRES_USER=nebula_test -e POSTGRES_DB=nebula_test --name nebula_test_db postgres:14

databased:
	docker run --rm -d -p 2345:5432 -e POSTGRES_PASSWORD=password -e POSTGRES_USER=nebula_test -e POSTGRES_DB=nebula_test --name nebula_test_db postgres:14
	sleep 1

database-stop:
	docker stop nebula_test_db || true

models:
	sqlboiler --no-tests psql

migrate-up:
	migrate -database 'postgres://nebula_test:password@localhost:2345/nebula_test?sslmode=disable' -path pkg/db/migrations up

migrate-down:
	migrate -database 'postgres://nebula_test:password@localhost:2345/nebula_test?sslmode=disable' -path pkg/db/migrations down

.PHONY: all clean test format tools models migrate-up migrate-down

default: all

all: clean build

test:
	go test ./...

build:
	go build -o dist/nebula cmd/nebula/*

format:
	gofumpt -w -l .

clean:
	rm -r dist || true

docker:
	docker build . -t dennis-tra/nebula-crawler:latest

tools:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.14.1
	go install github.com/volatiletech/sqlboiler/v4@v4.6.0
	go install github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-psql@v4.6.0

db-reset: migrate-down migrate-up models

models:
	sqlboiler psql

migrate-up:
	migrate -database 'postgres://nebula:password@localhost:5432/nebula?sslmode=disable' -path migrations up

migrate-down:
	migrate -database 'postgres://nebula:password@localhost:5432/nebula?sslmode=disable' -path migrations down

.PHONY: all clean test format tools models migrate-up migrate-down

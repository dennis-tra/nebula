default: all

all: clean build

test:
	go test ./...

build:
	go build -o dist/crawler cmd/nebula/*

format:
	gofumpt -w -l .

clean:
	rm -r dist

.PHONY: all clean test format

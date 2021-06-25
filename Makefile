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

.PHONY: all clean test format

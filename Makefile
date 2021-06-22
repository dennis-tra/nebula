default: build

all: clean install

test:
	go test ./...

build:
	go build -o dist/crawler cmd/crawler/crawler.go

install:
	go install cmd/crawler/crawler.go

format:
	gofumpt -w -l .

clean:
	rm -r dist

.PHONY: all clean test install format

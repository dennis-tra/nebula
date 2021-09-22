# Build nebula
FROM golang:1.16 AS builder

WORKDIR /build

RUN CGO_ENABLED=0 GOOS=linux go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.14.1

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o nebula cmd/nebula/*

# Create lightweight container to run nebula
FROM alpine:latest

# Create user ot
RUN adduser -D -H nebula
WORKDIR /home/nebula
RUN mkdir .config && chown nebula:nebula .config
USER nebula

COPY --from=builder /build/nebula /usr/local/bin/nebula
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate

# TODO: migrations on application level? https://github.com/golang-migrate/migrate#use-in-your-go-project
COPY --chown=nebula migrations migrations
COPY --chown=nebula deploy/docker-entrypoint.sh .
RUN chmod +x docker-entrypoint.sh

CMD nebula

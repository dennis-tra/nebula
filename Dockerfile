# Build nebula
FROM golang:1.19 AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-X main.RawVersion=`cat version`" -o nebula github.com/dennis-tra/nebula-crawler/cmd/nebula

# Create lightweight container to run nebula
FROM alpine:latest

# Create user nebula
RUN adduser -D -H nebula
WORKDIR /home/nebula
USER nebula

COPY --from=builder /build/nebula /usr/local/bin/nebula

CMD nebula

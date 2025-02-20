# We use the -alpine build image to build Nebula agains musl because we need
# CGO to be enabled and I'd like use an alpine base image for the final image.
FROM golang:1.23-alpine AS builder

ARG VERSION
ARG COMMIT
ARG DATE
ARG BUILT_BY

# Switch to an isolated build directory
WORKDIR /build

# For caching, only copy the dependency-defining files and download depdencies
COPY go.mod go.sum ./
RUN go mod download

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy everything else minus everything that's in .dockerignore
COPY . ./

# Finally build Nebula
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags "-X main.version=$VERSION -X main.commit=$COMMIT -X main.date=$DATE -X main.builtBy=$USER" -o dist/nebula github.com/dennis-tra/nebula-crawler/cmd/nebula

# Create lightweight container to run nebula
FROM alpine:latest

# Create user nebula
RUN adduser -D -H nebula
WORKDIR /home/nebula
USER nebula

COPY --from=builder /build/nebula /usr/local/bin/nebula

HEALTHCHECK --interval=15s --timeout=5s --start-period=10s CMD nebula health

CMD nebula

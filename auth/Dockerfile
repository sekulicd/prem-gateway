# first image used to build the sources
FROM golang:1.20-buster AS builder

ARG VERSION
ARG COMMIT
ARG DATE
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

COPY . .

RUN go mod download

RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-X 'main.Version=${COMMIT}' -X 'main.Commit=${COMMIT}' -X 'main.Date=${COMMIT}'" -o bin/authd cmd/authd/main.go
RUN go build -ldflags="-X 'main.version=${VERSION}' -X 'main.commit=${COMMIT}' -X 'main.date=${DATE}'" -o bin/authd cmd/authd/*

# Second image, running the oceand executable
FROM debian:buster-slim

# $USER name, and data $DIR to be used in the `final` image
ARG USER=authd
ARG DIR=/home/authd

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates

COPY --from=builder /app/bin/* /usr/local/bin/
COPY --from=builder /app/internal/infrastructure/storage/pg/migration/* /
ENV PREM_GATEWAY_AUTH_DB_MIGRATION_PATH=file://

# NOTE: Default GID == UID == 1000
RUN adduser --disabled-password \
            --home "$DIR/" \
            --gecos "" \
            "$USER"
USER $USER

# Prevents `VOLUME $DIR/.authd/` being created as owned by `root`
RUN mkdir -p "$DIR/.authd/"

# Expose volume containing all `authd` data
VOLUME $DIR/.authd/

ENTRYPOINT [ "authd" ]


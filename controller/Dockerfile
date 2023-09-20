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

RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-X 'main.Version=${COMMIT}' -X 'main.Commit=${COMMIT}' -X 'main.Date=${COMMIT}'" -o bin/controllerd cmd/controllerd/main.go
RUN go build -ldflags="-X 'main.version=${VERSION}' -X 'main.commit=${COMMIT}' -X 'main.date=${DATE}'" -o bin/controllerd cmd/controllerd/*

# Second image, running the oceand executable
FROM debian:buster-slim

# $USER name, and data $DIR to be used in the `final` image
ARG USER=controllerd
ARG DIR=/home/controllerd

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates

COPY --from=builder /app/bin/* /usr/local/bin/

# NOTE: Default GID == UID == 1000
RUN adduser --disabled-password \
            --home "$DIR/" \
            --gecos "" \
            "$USER"
USER $USER

# Prevents `VOLUME $DIR/.controllerd/` being created as owned by `root`
RUN mkdir -p "$DIR/.controllerd/"

# Expose volume containing all `controllerd` data
VOLUME $DIR/.controllerd/

ENTRYPOINT [ "controllerd" ]


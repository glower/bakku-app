# Reproducible builder image
FROM golang:1.10.0 as builder
WORKDIR /go/src/github.com/glower/bakku-app
# This expects that the context passed to the docker build command is
# the cluster-api-provider-digitalocean directory.
# e.g. docker build -t <tag> -f <this_Dockerfile> <path_to_cluster-api-digitalocean>
COPY . . 

RUN CGO_ENABLED=0 GOOS=linux go install -a -ldflags '-extldflags "-static"'github.com/glower/bakku-app/cmd/controller

# Final container
FROM debian:stretch-slim

COPY --from=builder /go/bin/cluster-controller .
###
# Build stage #
###############
FROM golang:1.23.9 AS builder

# Get the dependencies
ENV CGO_ENABLED=0
ENV GOOS=linux

## SETUP for the building

ARG ENV_ID_RSA

# Get the source code from a local copy
WORKDIR /go/src/github.com/GoDieNow/TFT_Code/templates/

COPY . ./collector

# Set the workdir
WORKDIR /go/src/github.com/GoDieNow/TFT_Code/templates/collector

RUN go mod download
RUN go build -a -ldflags "-extldflags \"-static\" -X main.version=collector-dirty -X main.collector=collector" -o /collector

###
# Deploy stage #
################
FROM alpine:latest AS deploy

# Update and certs install
RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates

# Set the workdir
WORKDIR /

## Add the wait script to the image
ADD https://github.com/ufoscout/docker-compose-wait/releases/download/2.7.3/wait /wait
RUN chmod +x /wait

# Copying the binary
COPY --from=builder /collector .

# Entrypoint
CMD /wait && /collector -conf=./config

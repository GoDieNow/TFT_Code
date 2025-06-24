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
WORKDIR /go/src/github.com/GoDieNow/TFT_Code/collectors/

COPY . ./servers

# Set the workdir
WORKDIR /go/src/github.com/GoDieNow/TFT_Code/collectors/servers/server

RUN rm -rf go.sum
RUN go mod tidy
RUN go mod download
RUN go build -a -ldflags "-extldflags \"-static\" -X main.version=servers-dirty -X main.service=servers" -o /servers

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
COPY --from=builder /servers .

# Entrypoint
CMD /wait && /servers -conf=./config

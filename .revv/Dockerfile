FROM golang:1.22-alpine

RUN apk add --no-cache make git gcc musl-dev docker-cli bash

WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download

COPY . /workspace

# Pre-build: cache compilation artifacts so tests start instantly
RUN go build -v -o /workspace/bin/revv ./cmd/revv

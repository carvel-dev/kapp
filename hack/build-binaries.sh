#!/bin/bash

set -e -x -u

GOOS=darwin GOARCH=amd64 go build -o kapp-darwin-amd64 ./cmd/...
GOOS=linux GOARCH=amd64 go build -o kapp-linux-amd64 ./cmd/...

shasum -a 256 ./kapp-*-amd64

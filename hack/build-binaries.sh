#!/bin/bash

set -e -x -u

BUILD_VALUES= ./hack/build.sh

GOOS=darwin GOARCH=amd64 go build -o kapp-darwin-amd64 ./cmd/kapp/...
GOOS=linux GOARCH=amd64 go build -o kapp-linux-amd64 ./cmd/kapp/...
GOOS=windows GOARCH=amd64 go build -o kapp-windows-amd64.exe ./cmd/kapp/...

shasum -a 256 ./kapp-*-amd64*

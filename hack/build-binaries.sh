#!/bin/bash

set -e -x -u

BUILD_VALUES= ./hack/build.sh

# makes builds reproducible
export CGO_ENABLED=0
repro_flags="-ldflags=-buildid= -trimpath"

GOOS=darwin GOARCH=amd64 go build $repro_flags -o kapp-darwin-amd64 ./cmd/kapp/...
GOOS=linux GOARCH=amd64 go build $repro_flags -o kapp-linux-amd64 ./cmd/kapp/...
GOOS=windows GOARCH=amd64 go build $repro_flags -o kapp-windows-amd64.exe ./cmd/kapp/...

shasum -a 256 ./kapp-*-amd64*

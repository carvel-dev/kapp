#!/bin/bash

set -e -x -u

# makes builds reproducible
export CGO_ENABLED=0
repro_flags="-ldflags=-buildid= -trimpath"

go fmt ./cmd/... ./pkg/... ./test/...
go mod vendor
go mod tidy

go build $repro_flags -o kapp ./cmd/kapp/...
./kapp version

echo "Success"

#!/bin/bash

set -e -x -u

# makes builds reproducible
export CGO_ENABLED=0
repro_flags="-ldflags=-buildid= -trimpath"

go mod vendor
go mod tidy
go fmt ./cmd/... ./pkg/... ./test/...

go build $repro_flags -o kapp ./cmd/kapp/...
./kapp version

# compile tests, but do not run them: https://github.com/golang/go/issues/15513#issuecomment-839126426
go test --exec=echo ./... >/dev/null

echo "Success"

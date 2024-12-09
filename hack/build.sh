#!/bin/bash

set -e -x -u

# makes builds reproducible
export CGO_ENABLED=0

go mod vendor
go mod tidy
go fmt ./cmd/... ./pkg/... ./test/...

go build -trimpath -o kapp ./cmd/kapp/...
./kapp version

# compile tests, but do not run them: https://github.com/golang/go/issues/15513#issuecomment-839126426
go test --exec=echo ./... >/dev/null

echo "Success"

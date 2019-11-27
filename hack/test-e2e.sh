#!/bin/bash

set -e -x -u

go clean -testcache

export KAPP_BINARY_PATH="${KAPP_BINARY_PATH:-$PWD/kapp}"

go test ./test/e2e/ -timeout 60m -test.v $@

echo E2E SUCCESS

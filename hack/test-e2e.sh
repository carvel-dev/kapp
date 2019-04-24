#!/bin/bash

set -e -x -u

go clean -testcache

go test ./test/e2e/ -timeout 60m -test.v $@

echo E2E SUCCESS

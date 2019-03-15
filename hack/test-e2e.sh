#!/bin/bash

set -e -x -u

GOCACHE=off go test ./test/e2e/ -timeout 60m -test.v $@

echo E2E SUCCESS

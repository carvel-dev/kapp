#!/bin/bash

set -e -x

if [ -z "$GITHUB_ACTION" ]; then
  go clean -testcache
fi

set -u

go test ./pkg/... -test.v $@

echo UNIT SUCCESS

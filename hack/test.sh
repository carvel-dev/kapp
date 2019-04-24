#!/bin/bash

set -e -x -u

go clean -testcache

go test ./pkg/... -test.v $@

echo UNIT SUCCESS

#!/bin/bash

set -e -x -u

GOCACHE=off go test ./pkg/... -test.v $@

echo UNIT SUCCESS

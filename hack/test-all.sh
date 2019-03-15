#!/bin/bash

set -e -x -u

./hack/build.sh
./hack/test.sh
./hack/test-e2e.sh

echo ALL SUCCESS

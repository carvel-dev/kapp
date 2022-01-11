#!/bin/bash

set -e -x -u

./hack/build.sh

export KAPP_BINARY_PATH="$PWD/kapp"

./hack/test.sh
KAPP_E2E_SSA=0 ./hack/test-e2e.sh
KAPP_E2E_SSA=1 ./hack/test-e2e.sh

echo ALL SUCCESS

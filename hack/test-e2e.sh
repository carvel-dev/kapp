#!/bin/bash

set -e -x -u

go clean -testcache

export KAPP_BINARY_PATH="${KAPP_BINARY_PATH:-$PWD/kapp}"

if [ -z "$KAPP_E2E_NAMESPACE" ]; then
    echo "setting e2e namespace to: kapp-test";
    export KAPP_E2E_NAMESPACE="kapp-test"
fi
# create ns if not exists because the `apply -f -` won't complain on a no-op if the ns already exists.
kubectl create ns $KAPP_E2E_NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

go test ./test/e2e/ -timeout 60m -test.v $@

echo E2E SUCCESS

#!/bin/bash

set -e

./hack/build.sh

if ! git diff --exit-code >/dev/null; then
  echo "Error: Running ./hack/build.sh resulted in non zero exit code from git diff. Please run './hack/build.sh' and 'git add' the generated file(s)."
  echo "Showing diff:"
  git diff --exit-code
  exit 1
fi

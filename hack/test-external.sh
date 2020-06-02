#!/bin/bash

set -e -x -u

time kapp deploy -y -a cert-manager -f examples/cert-manager-v0.15.1/
time kapp delete -y -a cert-manager

time kapp deploy -y -a knative -f examples/knative-v0.15.0/
time kapp delete -y -a knative

echo EXTERNAL SUCCESS

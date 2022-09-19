#!/bin/bash

set -e -x -u

minikube tunnel &> /dev/null &

time kapp app-group deploy -y -g gitops -d examples/gitops/
time kapp app-group delete -y -g gitops

# TODO Add new istio version of istio supporting k8s 1.25 when it's available
# time kapp deploy -y -a istio -f examples/istio-v1.12.2/
# time kapp delete -y -a istio

time kapp deploy -y -a cert-manager -f examples/cert-manager-v1.6.1/
time kapp delete -y -a cert-manager

# TODO Add knative - Commenting because it need > resources than what github action provides.
#time kapp deploy -y -a knative -f examples/knative-v1.1.0/
#time kapp delete -y -a knative

# TODO Add cf-for-k8s-v5.4.3

# TODO Add gatekeeper 3.10.0 when it's available
# time kapp deploy -y -a gk -f examples/gatekeeper-v3.7.0/config.yml
# time kapp delete -y -a gk

time kapp deploy -y -a pinniped -f examples/pinniped-v0.13.0/
time kapp delete -y -a pinniped

pkill -9 'minikube'

echo EXTERNAL SUCCESS

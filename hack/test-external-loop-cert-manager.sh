#!/bin/bash

set -e -x -u

while true; do

time kapp deploy -y -a cert-manager -f examples/cert-manager-v0.15.1/
time kapp delete -y -a cert-manager

done

echo EXTERNAL SUCCESS

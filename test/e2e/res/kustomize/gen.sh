#!/usr/bin/env bash
cd $(dirname ${BASH_SOURCE})
for overlay in overlays/*/
do
	kustomize build $overlay -o $overlay/kapp.yml
done

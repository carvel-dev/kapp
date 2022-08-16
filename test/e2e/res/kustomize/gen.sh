#!/usr/bin/env bash
for overlay in overlays/*/
do
	kustomize build $overlay -o $overlay/kapp.yml
done

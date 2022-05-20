#!/usr/bin/env bash

cd $(dirname ${BASH_SOURCE[0]})

action=$1

[ -x "$action.sh" ] && action=hack/$action.sh

podman run --rm -v gopath:/go -v gobuild:/root/.cache/go-build -v $(realpath $PWD/..):/mnt -w /mnt docker.io/golang $action

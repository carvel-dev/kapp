#!/usr/bin/env bash

cd $(dirname ${BASH_SOURCE[0]})

action=$1

if [ "$action" ]; then
    [ -x "$action.sh" ] && action=hack/$action.sh
else
    interactive="-it"
fi

# SYS_PTRACE for dlv
podman run --rm -v gopath:/go -v gobuild:/root/.cache/go-build -v $(realpath $PWD/..):/mnt -w /mnt $interactive --cap-add SYS_PTRACE docker.io/golang $action

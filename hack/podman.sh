#!/usr/bin/env bash

cd $(dirname ${BASH_SOURCE[0]})

action=$1

if [ "$action" ]; then
    [ -x "$action.sh" ] && action=hack/$action.sh
else
    interactive="-it -v $PWD/$histfile:/root/$histfile"

    histfile=.bash_history
    [ -f $histfile ] || >$histfile
fi

# SYS_PTRACE for dlv
# host network for minikube
podman run --rm \
    -v gopath:/go \
    -v gobuild:/root/.cache/go-build \
    -v $HOME/.kube:/root/.kube \
    -v $HOME/.minikube:$HOME/.minikube \
    -v $(realpath $PWD/..):/mnt -w /mnt $interactive \
    --cap-add SYS_PTRACE \
    --network host \
    docker.io/golang $action

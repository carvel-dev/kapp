#!/usr/bin/env bash

cd $(dirname ${BASH_SOURCE[0]})

action=$1

if [ "$action" ]; then
    [ -x "$action.sh" ] && action=hack/$action.sh
else
    histfile=.bash_history
    [ -f $histfile ] || >$histfile

    interactive="-it -v $PWD/$histfile:/root/$histfile"

    dlv_conf=.dlv
    [ -d $dlv_conf ] || mkdir $dlv_conf

    interactive="$interactive -v $PWD/$dlv_conf:/root/.config/dlv"
fi

gopath_vol=${KAPP_GOPATH_VOL:-gopath}
gobuild_vol=${KAPP_GOBUILD_VOL:-gobuild}

# SYS_PTRACE for dlv
# host network for minikube
podman run --rm \
    -v $gopath_vol:/go \
    -v $gobuild_vol:/root/.cache/go-build \
    -v $HOME/.kube:/root/.kube \
    -v $HOME/.minikube:$HOME/.minikube \
    -v $(realpath $PWD/..):/mnt -w /mnt \
    $interactive \
    --cap-add SYS_PTRACE \
    --network host \
    docker.io/golang $action

#!/bin/bash

set -e -x -u

function get_latest_git_tag {
  git describe --tags | grep -Eo '[0-9]+\.[0-9]+\.[0-9]+'
}

VERSION="${1:-`get_latest_git_tag`}"

# makes builds reproducible
export CGO_ENABLED=0
LDFLAGS="-X github.com/vmware-tanzu/carvel-kapp/pkg/kapp/version.Version=$VERSION -buildid="

GOOS=darwin GOARCH=amd64 go build -ldflags="$LDFLAGS" -trimpath -o kapp-darwin-amd64 ./cmd/kapp/...
GOOS=darwin GOARCH=arm64 go build -ldflags="$LDFLAGS" -trimpath -o kapp-darwin-arm64 ./cmd/kapp/...
GOOS=linux GOARCH=amd64 go build -ldflags="$LDFLAGS" -trimpath -o kapp-linux-amd64 ./cmd/kapp/...
GOOS=linux GOARCH=arm64 go build -ldflags="$LDFLAGS" -trimpath -o kapp-linux-arm64 ./cmd/kapp/...
GOOS=windows GOARCH=amd64 go build -ldflags="$LDFLAGS" -trimpath -o kapp-windows-amd64.exe ./cmd/kapp/...

shasum -a 256 ./kapp-darwin-amd64 ./kapp-darwin-arm64 ./kapp-linux-amd64 ./kapp-linux-arm64 ./kapp-windows-amd64.exe

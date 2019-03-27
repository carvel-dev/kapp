#!/bin/bash

if test -z "$BASH_VERSION"; then
  echo "Please run this script using bash, not sh or any other shell." >&2
  exit 1
fi

install() {
	set -euo pipefail

	# Start Kubernetes on Katacoda
	launch.sh

	# Install ytt
	echo "Installing ytt..."
	wget -O- https://github.com/k14s/ytt/releases/download/v0.4.0/ytt-linux-amd64 > /tmp/ytt
	echo "5205c22c5bd5a34c2d6f63e9870d70a7272d33d01dcbace2c5fbefcca85e00aa  /tmp/ytt" | shasum -c -
	mv /tmp/ytt /usr/local/bin/ytt
	chmod +x /usr/local/bin/ytt
	echo "Installed ytt"

	# Install kapp
	echo "Installing kapp..."
	wget -O- https://github.com/k14s/kapp/releases/download/v0.4.0/kapp-linux-amd64 > /tmp/kapp
	echo "7a61cf9523281d94711c752ebd8d890068602f3fe6baf344895fbc59b9d5664e  /tmp/kapp" | shasum -c -
	mv /tmp/kapp /usr/local/bin/kapp
	chmod +x /usr/local/bin/kapp
	echo "Installed kapp"

	git clone https://github.com/k14s/kapp
	echo "Cloned github.com/k14s/kapp for examples"
}

install

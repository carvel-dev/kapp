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
	wget -O- https://github.com/k14s/ytt/releases/download/v0.5.0/ytt-linux-amd64 > /tmp/ytt
	echo "340a3bd30c925f865b53762e3c54b88843b2d0b898fbb58e2deb003ea182df26  /tmp/ytt" | shasum -c -
	mv /tmp/ytt /usr/local/bin/ytt
	chmod +x /usr/local/bin/ytt
	echo "Installed ytt"

	# Install kapp
	echo "Installing kapp..."
	wget -O- https://github.com/k14s/kapp/releases/download/v0.4.0/kapp-linux-amd64 > /tmp/kapp
	echo "c6b603ac7dce5ba7f0679df7b69f39a35c8278f479534c2ea5cda8a83acfc0a1  /tmp/kapp" | shasum -c -
	mv /tmp/kapp /usr/local/bin/kapp
	chmod +x /usr/local/bin/kapp
	echo "Installed kapp"

	git clone https://github.com/k14s/kapp
	echo "Cloned github.com/k14s/kapp for examples"
}

install

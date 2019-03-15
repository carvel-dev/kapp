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
	wget -O- https://github.com/get-ytt/ytt/releases/download/v0.1.0/ytt-linux-amd64 > /tmp/ytt
	echo "08d52157a8a7cea47215f05b5e16f318005430d15729f45697686715cb92b705  /tmp/ytt" | shasum -c -
	mv /tmp/ytt /usr/local/bin/ytt
	chmod +x /usr/local/bin/ytt
	echo "Installed ytt"

	# Install kapp
	wget -O- https://github.com/k14s/kapp/releases/download/v0.1.0/kapp-linux-amd64 > /tmp/kapp
	echo "xxx  /tmp/kapp" | shasum -c -	
	mv /tmp/kapp /usr/local/bin/kapp
	chmod +x /usr/local/bin/kapp
	echo "Installed kapp"

	git clone https://github.com/k14s/kapp
	echo "Cloned github.com/k14s/kapp for examples"
}

install

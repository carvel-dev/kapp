#!/bin/bash

set -e -x -u

go fmt ./cmd/... ./pkg/... ./test/...

(
	# template all playground assets
	# into a single Go file
	cd pkg/kapp/website; 

	ytt version || { echo >&2 "ytt is required for building. Install from https://github.com/k14s/ytt"; exit 1; }
	ytt template -R \
		-f . \
		-f ../../../hack/build-values.yml \
		--file-mark 'generated.go.txt:exclusive-for-output=true' \
		--output ../../../tmp/
)
mv tmp/generated.go.txt pkg/kapp/website/generated.go

go build -o kapp ./cmd/kapp/...
./kapp version

# build aws lambda binary
export GOOS=linux GOARCH=amd64
go build -o ./tmp/main ./cmd/kapp-lambda-website/...
(
	cd tmp
	chmod +x main
	rm -f kapp-lambda-website.zip
	zip kapp-lambda-website.zip main
)

echo "Success"

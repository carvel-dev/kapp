#!/bin/bash

set -e -x -u

go fmt ./cmd/... ./pkg/... ./test/...

(
	# template all playground assets
	# into a single Go file
	cd pkg/kapp/website; 
	ytt template -R \
		-f . \
		-f ../../../hack/build-values.yml \
		--filter-template-file generated.go.txt \
		--filter-template-file build-values.yml \
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

// Copyright 2022 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"bytes"
	"embed"
	"io"
)

//go:embed res
var testResFs embed.FS

func TestResReader(path string) (io.Reader, error) {
	b, err := TestResBytes(path)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b), nil
}

func TestResBytes(path string) ([]byte, error) {
	return testResFs.ReadFile(path)
}

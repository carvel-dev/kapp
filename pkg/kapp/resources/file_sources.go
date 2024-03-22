// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
)

type FileSource interface {
	Description() string
	Bytes() ([]byte, error)
}

type BytesSource struct {
	bytes []byte
}

var _ FileSource = BytesSource{}

func NewBytesSource(bytes []byte) BytesSource { return BytesSource{bytes} }
func (s BytesSource) Description() string     { return "bytes" }
func (s BytesSource) Bytes() ([]byte, error)  { return s.bytes, nil }

type StdinSource struct{}

var _ FileSource = StdinSource{}

func NewStdinSource() StdinSource            { return StdinSource{} }
func (s StdinSource) Description() string    { return "stdin" }
func (s StdinSource) Bytes() ([]byte, error) { return io.ReadAll(os.Stdin) }

type LocalFileSource struct {
	fsys fs.FS
	path string
}

var _ FileSource = LocalFileSource{}

func NewLocalFileSource(fsys fs.FS, path string) LocalFileSource {
	return LocalFileSource{fsys: fsys, path: path}
}
func (s LocalFileSource) Description() string { return fmt.Sprintf("file '%s'", s.path) }
func (s LocalFileSource) Bytes() ([]byte, error) {
	switch t := s.fsys.(type) {
	case fs.ReadFileFS:
		return t.ReadFile(s.path)
	case fs.FS:
		f, err := t.Open(s.path)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		return fs.ReadFile(s.fsys, s.path)
	default:
		return os.ReadFile(s.path)
	}
}

type HTTPFileSource struct {
	url    string
	Client *http.Client
}

var _ FileSource = HTTPFileSource{}

func NewHTTPFileSource(path string) HTTPFileSource { return HTTPFileSource{path, &http.Client{}} }

func (s HTTPFileSource) Description() string {
	return fmt.Sprintf("HTTP URL '%s'", s.url)
}

func (s HTTPFileSource) Bytes() ([]byte, error) {
	resp, err := s.Client.Get(s.url)
	if err != nil {
		return nil, fmt.Errorf("Requesting URL '%s': %w", s.url, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("Requesting URL '%s': %s", s.url, resp.Status)
	}

	result, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Reading URL '%s': %w", s.url, err)
	}

	return result, nil
}

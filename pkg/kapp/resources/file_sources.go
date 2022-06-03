// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"fmt"
	"io/ioutil"
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
func (s StdinSource) Bytes() ([]byte, error) { return ioutil.ReadAll(os.Stdin) }

type LocalFileSource struct {
	path string
}

var _ FileSource = LocalFileSource{}

func NewLocalFileSource(path string) LocalFileSource { return LocalFileSource{path} }
func (s LocalFileSource) Description() string        { return fmt.Sprintf("file '%s'", s.path) }
func (s LocalFileSource) Bytes() ([]byte, error)     { return ioutil.ReadFile(s.path) }

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
		return nil, fmt.Errorf("Requesting URL '%s': %s", s.url, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("Requesting URL '%s': %s", s.url, resp.Status)
	}

	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Reading URL '%s': %s", s.url, err)
	}

	return result, nil
}

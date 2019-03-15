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

type StdinSource struct{}

var _ FileSource = StdinSource{}

func NewStdinSource() StdinSource { return StdinSource{} }

func (s StdinSource) Description() string { return "stdin" }

func (s StdinSource) Bytes() ([]byte, error) {
	return ioutil.ReadAll(os.Stdin)
}

type LocalFileSource struct {
	path string
}

var _ FileSource = LocalFileSource{}

func NewLocalFileSource(path string) LocalFileSource { return LocalFileSource{path} }

func (s LocalFileSource) Description() string {
	return fmt.Sprintf("file '%s'", s.path)
}

func (s LocalFileSource) Bytes() ([]byte, error) {
	return ioutil.ReadFile(s.path)
}

type HTTPFileSource struct {
	url string
}

var _ FileSource = HTTPFileSource{}

func NewHTTPFileSource(path string) HTTPFileSource { return HTTPFileSource{path} }

func (s HTTPFileSource) Description() string {
	return fmt.Sprintf("HTTP URL '%s'", s.url)
}

func (s HTTPFileSource) Bytes() ([]byte, error) {
	resp, err := http.Get(s.url)
	if err != nil {
		return nil, fmt.Errorf("Requesting URL '%s': %s", s.url, err)
	}

	defer resp.Body.Close()

	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Reading URL '%s': %s", s.url, err)
	}

	return result, nil
}

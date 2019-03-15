package resources

import (
	"bufio"
	"bytes"
	"io"

	kyaml "k8s.io/apimachinery/pkg/util/yaml"
)

type YAMLFile struct {
	fileSrc FileSource
}

func NewYAMLFile(fileSrc FileSource) YAMLFile {
	return YAMLFile{fileSrc}
}

func (f YAMLFile) Docs() ([][]byte, error) {
	var docs [][]byte

	fileBytes, err := f.fileSrc.Bytes()
	if err != nil {
		return nil, err
	}

	reader := kyaml.NewYAMLReader(bufio.NewReaderSize(bytes.NewReader(fileBytes), 4096))

	for {
		docBytes, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		docs = append(docs, docBytes)
	}

	return docs, nil
}

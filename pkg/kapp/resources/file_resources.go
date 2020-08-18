// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var (
	fileResourcesAllowedExts = []string{".json", ".yaml", ".yml"} // matches kubectl
)

type FileResource struct {
	fileSrc FileSource
}

func NewFileResources(file string) ([]FileResource, error) {
	var fileRs []FileResource

	switch {
	case file == "-":
		fileRs = append(fileRs, NewFileResource(NewStdinSource()))

	case strings.HasPrefix(file, "http://") || strings.HasPrefix(file, "https://"):
		fileRs = append(fileRs, NewFileResource(NewHTTPFileSource(file)))

	default:
		fileInfo, err := os.Stat(file)
		if err != nil {
			return nil, fmt.Errorf("Checking file '%s'", file)
		}

		if fileInfo.IsDir() {
			var paths []string

			err := filepath.Walk(file, func(path string, fi os.FileInfo, err error) error {
				if err != nil || fi.IsDir() {
					return err
				}
				ext := filepath.Ext(path)
				for _, allowedExt := range fileResourcesAllowedExts {
					if allowedExt == ext {
						paths = append(paths, path)
					}
				}
				return nil
			})
			if err != nil {
				return nil, fmt.Errorf("Listing files '%s'", file)
			}

			sort.Strings(paths)

			for _, path := range paths {
				fileRs = append(fileRs, NewFileResource(NewLocalFileSource(path)))
			}
		} else {
			fileRs = append(fileRs, NewFileResource(NewLocalFileSource(file)))
		}
	}

	return fileRs, nil
}

func NewFileResource(fileSrc FileSource) FileResource { return FileResource{fileSrc} }

func (r FileResource) Description() string { return r.fileSrc.Description() }

func (r FileResource) Resources() ([]Resource, error) {
	docs, err := NewYAMLFile(r.fileSrc).Docs()
	if err != nil {
		return nil, err
	}

	var resources []Resource

	for i, doc := range docs {
		rs, err := NewResourcesFromBytes(doc)
		if err != nil {
			return nil, err
		}

		for _, res := range rs {
			res.SetOrigin(fmt.Sprintf("%s doc %d", r.fileSrc.Description(), i+1))
		}

		resources = append(resources, rs...)
	}

	return resources, nil
}

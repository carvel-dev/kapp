// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"fmt"
	"io/fs"
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

// NewFileResources inspects file and returns a slice of FileResource objects. If file is "-", a FileResource for STDIN
// is returned. If it is prefixed with either http:// or https://, a FileResource that supports an HTTP transport is
// returned. If file is a directory, one FileResource object is returned for each file in the directory with an allowed
// extension (.json, .yml, .yaml). If file is not a directory, a FileResource object is returned for that one file. If
// fsys is nil, NewFileResources uses the OS's file system. Otherwise, it uses the passed in file system.
func NewFileResources(fsys fs.FS, file string) ([]FileResource, error) {
	var fileRs []FileResource

	switch {
	case file == "-":
		fileRs = append(fileRs, NewFileResource(NewStdinSource()))

	case strings.HasPrefix(file, "http://") || strings.HasPrefix(file, "https://"):
		fileRs = append(fileRs, NewFileResource(NewHTTPFileSource(file)))

	default:
		dir, err := isDir(fsys, file)
		if err != nil {
			return nil, err
		}

		if dir {
			// The typical command line invocation won't set fsys. If it comes in nil, create a new DirFS rooted at
			// file, then set file to '.' (current working directory) so the fs.WalkDir call below works correctly.
			if fsys == nil {
				fsys = os.DirFS(file)
				file = "."
			}

			var paths []string
			err := fs.WalkDir(fsys, file, func(path string, d fs.DirEntry, err error) error {
				if err != nil || d.IsDir() {
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
				return nil, fmt.Errorf("error listing file %q", file)
			}

			sort.Strings(paths)

			for _, path := range paths {
				fileRs = append(fileRs, NewFileResource(NewLocalFileSource(fsys, path)))
			}
		} else {
			fileRs = append(fileRs, NewFileResource(NewLocalFileSource(fsys, file)))
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

// isDir returns if path is a directory. If fsys is nil, isDir calls os.Stat(path); otherwise, it checks path inside
// fsys.
func isDir(fsys fs.FS, path string) (bool, error) {
	if fsys == nil {
		fileInfo, err := os.Stat(path)
		if err != nil {
			return false, err
		}
		return fileInfo.IsDir(), nil
	}

	switch t := fsys.(type) {
	case fs.StatFS:
		fileInfo, err := t.Stat(path)
		if err != nil {
			return false, err
		}
		return fileInfo.IsDir(), nil
	case fs.FS:
		f, err := t.Open(path)
		if err != nil {
			return false, fmt.Errorf("error opening file %q: %v", path, err)
		}
		defer f.Close()

		fileInfo, err := f.Stat()
		if err != nil {
			return false, err
		}
		return fileInfo.IsDir(), nil
	default:
		return false, fmt.Errorf("error determining if %q is a directory: unexpected FS type %T", path, fsys)
	}
}

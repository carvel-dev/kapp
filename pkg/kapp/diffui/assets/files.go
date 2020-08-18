// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package assets

type File struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

const (
	IndexHTMLPath = "templates/index.html"
)

var Files = map[string]File{
	"assets/all.js": File{
		Name:    "assets/all.js",
		Content: jqueryJS + mainJS,
	},
	"assets/all.css": File{
		Name:    "assets/all.css",
		Content: mainCSS,
	},
	IndexHTMLPath: File{
		Name:    IndexHTMLPath,
		Content: indexHTML,
	},
}

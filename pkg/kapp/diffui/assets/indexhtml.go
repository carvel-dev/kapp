// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package assets

const (
	IndexHTMLDiffDataJSONMarker = "__diffdata_json__"
	indexHTML                   = `
<html>
  <head>
    <title>kapp - diffui</title>
    <link href="/assets/all.css" media="all" rel="stylesheet" type="text/css" />
    <script src="/assets/all.js"></script>
   	<script>window.diffData = ` + IndexHTMLDiffDataJSONMarker + `;</script>
  </head>
  <body>
  	<h1>Changes</h1>
  	<p>Changes are sorted in order that they will be applied. They will be applied in parallel within their group. Each change lists other changes that it will wait for before being applied.</p>
    <ol id="deps"></ol>
  </body>
</html>
`
)

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

  	<h2>Dependencies</h2>
  	<p>Each change lists other changes that it will wait for before being applied.</p>
    <ul id="deps"></ul>
  </body>
</html>
`
)

// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package diffui

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	ctldgraph "carvel.dev/kapp/pkg/kapp/diffgraph"
	"carvel.dev/kapp/pkg/kapp/diffui/assets"
	"github.com/cppforlife/go-cli-ui/ui"
)

type ServerOpts struct {
	DiffDataFunc func() *ctldgraph.ChangeGraph
}

type Server struct {
	opts ServerOpts
	ui   ui.UI
}

func NewServer(opts ServerOpts, ui ui.UI) *Server {
	return &Server{opts, ui}
}

func (s *Server) Mux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.noCacheHandler(s.mainHandler))
	mux.HandleFunc("/assets/", s.noCacheHandler(s.assetHandler))
	return mux
}

func (s *Server) Run() error {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return err
	}

	s.ui.BeginLinef("Diff UI server: http://%s\n", listener.Addr())

	return (&http.Server{Handler: s.Mux()}).Serve(listener)
}

type diffData struct {
	AllChanges               []diffDataChange `json:"allChanges"`
	LinearizedChangeSections [][]string       `json:"linearizedChangeSections"`
	BlockedChanges           []string         `json:"blockedChanges"`
}

type diffDataChange struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	WaitingForIDs []string `json:"waitingForIDs"`
}

func (s *Server) mainHandler(w http.ResponseWriter, _ *http.Request) {
	changesGraph := s.opts.DiffDataFunc()

	allChanges := changesGraph.All()
	linearizedChangeSections, blockedChanges := changesGraph.Linearized()
	changeID := func(ch *ctldgraph.Change) string { return fmt.Sprintf("ch-%p", ch) }

	diffData := diffData{}

	for _, change := range allChanges {
		ddChange := diffDataChange{ID: changeID(change), Name: change.Description()}
		for _, depChange := range change.WaitingFor {
			ddChange.WaitingForIDs = append(ddChange.WaitingForIDs, changeID(depChange))
		}
		diffData.AllChanges = append(diffData.AllChanges, ddChange)
	}

	for _, section := range linearizedChangeSections {
		var changeIDs []string
		for _, change := range section {
			changeIDs = append(changeIDs, changeID(change))
		}
		diffData.LinearizedChangeSections = append(diffData.LinearizedChangeSections, changeIDs)
	}

	for _, change := range blockedChanges {
		diffData.BlockedChanges = append(diffData.BlockedChanges, changeID(change))
	}

	dataBs, _ := json.Marshal(diffData)

	indexHTML := assets.Files[assets.IndexHTMLPath].Content
	content := strings.ReplaceAll(indexHTML, assets.IndexHTMLDiffDataJSONMarker, string(dataBs))

	s.write(w, []byte(content))
}

func (s *Server) assetHandler(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, ".css") {
		w.Header().Set("Content-Type", "text/css")
	}
	if strings.HasSuffix(r.URL.Path, ".js") {
		w.Header().Set("Content-Type", "application/javascript")
	}
	s.write(w, []byte(assets.Files[strings.TrimPrefix(r.URL.Path, "/")].Content))
}

func (s *Server) write(w http.ResponseWriter, data []byte) {
	w.Write(data) // not fmt.Fprintf!
}

var (
	noCacheHeaders = map[string]string{
		"Expires":         time.Unix(0, 0).Format(time.RFC1123),
		"Cache-Control":   "no-cache, private, max-age=0",
		"Pragma":          "no-cache",
		"X-Accel-Expires": "0",
	}
)

func (s *Server) noCacheHandler(wrappedFunc func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		for k, v := range noCacheHeaders {
			w.Header().Set(k, v)
		}

		wrappedFunc(w, r)
	}
}

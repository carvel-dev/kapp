package diffui

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/cppforlife/go-cli-ui/ui"
	ctldgraph "github.com/k14s/kapp/pkg/kapp/diffgraph"
	"github.com/k14s/kapp/pkg/kapp/diffui/assets"
)

type ServerOpts struct {
	ListenAddr   string
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
	server := &http.Server{
		Addr:    s.opts.ListenAddr,
		Handler: s.Mux(),
	}
	s.ui.BeginLinef("Diff UI server listening on http://%s\n", server.Addr)
	return server.ListenAndServe()
}

type diffData struct {
	Nodes []node `json:"nodes"`
	Links []link `json:"links"`
}

type node struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type link struct {
	Source int `json:"source"`
	Target int `json:"target"`
}

func (s *Server) mainHandler(w http.ResponseWriter, r *http.Request) {
	changes := s.opts.DiffDataFunc().All()
	changeIdx := map[*ctldgraph.Change]int{}

	var nodes []node
	var links []link

	for i, change := range changes {
		nodes = append(nodes, node{ID: change.Change.Resource().Description(), Name: ""})
		changeIdx[change] = i
	}

	for i, change := range changes {
		for _, depChange := range change.WaitingFor {
			links = append(links, link{Source: i, Target: changeIdx[depChange]})
		}
	}

	dataBs, _ := json.Marshal(diffData{Nodes: nodes, Links: links})

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

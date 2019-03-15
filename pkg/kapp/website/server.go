package website

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

type ServerOpts struct {
	ListenAddr string
	ErrorFunc  func(error) ([]byte, error)
}

type Server struct {
	opts ServerOpts
}

func NewServer(opts ServerOpts) *Server {
	return &Server{opts}
}

func (s *Server) Mux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.redirectToHTTPs(s.noCacheHandler(s.mainHandler)))
	mux.HandleFunc("/install-katacoda.sh", s.redirectToHTTPs(s.noCacheHandler(s.installKatacodaHandler)))
	mux.HandleFunc("/js/", s.redirectToHTTPs(s.noCacheHandler(s.assetHandler)))
	mux.HandleFunc("/health", s.healthHandler)
	return mux
}

func (s *Server) Run() error {
	server := &http.Server{
		Addr:    s.opts.ListenAddr,
		Handler: s.Mux(),
	}
	fmt.Printf("Listening on http://%s\n", server.Addr)
	return server.ListenAndServe()
}

func (s *Server) mainHandler(w http.ResponseWriter, r *http.Request) {
	s.write(w, []byte(Files["templates/index.html"].Content))
}

func (s *Server) installKatacodaHandler(w http.ResponseWriter, r *http.Request) {
	s.write(w, []byte(Files["templates/install-katacoda.sh"].Content))
}

func (s *Server) assetHandler(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, ".css") {
		w.Header().Set("Content-Type", "text/css")
	}
	if strings.HasSuffix(r.URL.Path, ".js") {
		w.Header().Set("Content-Type", "application/javascript")
	}
	s.write(w, []byte(Files[strings.TrimPrefix(r.URL.Path, "/")].Content))
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	s.write(w, []byte("ok"))
}

func (s *Server) logError(w http.ResponseWriter, err error) {
	log.Print(err.Error())

	resp, err := s.opts.ErrorFunc(err)
	if err != nil {
		fmt.Fprintf(w, "generation error: %s", err.Error())
		return
	}

	s.write(w, resp)
}

func (s *Server) write(w http.ResponseWriter, data []byte) {
	w.Write(data) // not fmt.Fprintf!
}

func (s *Server) redirectToHTTPs(wrappedFunc func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		checkHTTPs := true
		clientIP, _, err := net.SplitHostPort(r.RemoteAddr)
		if err == nil {
			if clientIP == "127.0.0.1" {
				checkHTTPs = false
			}
		}

		if checkHTTPs && r.Header.Get(http.CanonicalHeaderKey("x-forwarded-proto")) != "https" {
			if r.Method == http.MethodGet || r.Method == http.MethodHead {
				host := r.Header.Get("host")
				if len(host) == 0 {
					s.logError(w, fmt.Errorf("expected non-empty Host header"))
					return
				}

				http.Redirect(w, r, "https://"+host, http.StatusMovedPermanently)
				return
			}

			// Fail if it's not a GET or HEAD since req may have carried body insecurely
			s.logError(w, fmt.Errorf("expected HTTPs connection"))
			return
		}

		wrappedFunc(w, r)
	}
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

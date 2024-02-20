package rest

import (
	"fmt"
	"github/joekhoobyar/epigon/storage"
	"log"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"

	"github.com/julienschmidt/httprouter"
)

var (
	routerParamRegexp = regexp.MustCompile(":[^/]+")
)

type Options struct {
	LogPrefix string
}

type Server struct {
	logPrefix string
	Server    *httptest.Server
	Router    *httprouter.Router
}

func (srv *Server) defaultNotFound(w http.ResponseWriter, rq *http.Request) {
	m := rq.Method
	p := rq.URL.Path
	log.Printf("%s defaultNotFound(%s %s)", srv.logPrefix, m, p)
	if m == "GET" {
		w.WriteHeader(404)
	} else {
		w.WriteHeader(599)
	}
	w.Write([]byte(fmt.Sprint("No route matches", m, p))) // nolint
}

func NewServer(options Options) *Server {
	router := httprouter.New()

	server := &Server{
		logPrefix: options.LogPrefix,
		Server:    httptest.NewServer(router),
		Router:    router,
	}

	router.NotFound = http.HandlerFunc(server.defaultNotFound)

	return server
}

func (srv *Server) BuildService(root string, store storage.RWCache) *ServiceBuilder {
	if !strings.HasSuffix(root, "/") {
		root += "/"
	}
	return &ServiceBuilder{
		store:        store,
		server:       srv,
		root:         root,
		routes:       []serviceRoute{},
		resource:     map[string]ResourceAdapter{},
		defaultError: defaultErrorHandler,
	}
}

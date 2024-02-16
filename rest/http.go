package rest

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/julienschmidt/httprouter"
)

type Options struct {
	LogPrefix   string
	FixturePath string
}

type Server struct {
	logPrefix   string
	fixturePath string
	Server      *httptest.Server
	Router      *httprouter.Router
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

func buildServer(options Options) *Server {
	router := httprouter.New()

	server := &Server{
		logPrefix:   options.LogPrefix,
		fixturePath: options.FixturePath,
		Server:      httptest.NewServer(router),
		Router:      router,
	}

	router.NotFound = http.HandlerFunc(server.defaultNotFound)

	return server
}

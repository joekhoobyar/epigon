package test

import (
	"github/joekhoobyar/epigon/rest"
	"net/http"
	"path"
	"runtime"
)

func FixtureDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return path.Clean(path.Join(path.Dir(filename), "fixtures"))
}

func GET(srv *rest.Server, path string) (*http.Response, error) {
	url := srv.Server.URL + path
	return http.Get(url)
}

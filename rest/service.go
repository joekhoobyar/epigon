package rest

import (
	"fmt"
	"github/joekhoobyar/epigon/storage"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
)

type ErrHandler func(http.ResponseWriter, *http.Request, httprouter.Params, error)

type Service struct {
	resource     map[string]ResourceAdapter
	store        storage.RWCache
	defaultError ErrHandler
}

type ResourceAdapter interface {
	New() any
	Convert(rq http.Request, source any) (id string, target any, err error)
}

func NewService(store storage.RWCache) *Service {
	return &Service{
		store:        store,
		defaultError: defaultErrorHandler,
	}
}

func defaultErrorHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, err error) {
	w.WriteHeader(599)
	w.Write([]byte(err.Error())) // nolint
}

func LocateResource(route string, params httprouter.Params) (location string, err error) {
	route = routerParamRegexp.ReplaceAllStringFunc(route, func(match string) string {
		value := params.ByName(match[1:])
		if value == "" {
			err = fmt.Errorf("LocateResource: %s: no such path parameter", match)
		}
		return value
	})
	if err == nil {
		location = route
	}
	return
}

func (svc *Service) List(route string) httprouter.Handle {
	if !strings.HasSuffix(route, "/") {
		route += "/"
	}

	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		var location string
		var buff []byte
		var err error

		if location, err = LocateResource(route, ps); err == nil {
			if buff, err = svc.store.ReadList(location); err == nil {
				w.WriteHeader(200)
				w.Write(buff) // nolint
				return
			}
		}

		svc.defaultError(w, r, ps, err)
	}
}

func (svc *Service) Get(route, idParam string) httprouter.Handle {
	route, _ = strings.CutSuffix(route, "/")

	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		var location string
		var buff []byte
		var err error

		id := ps.ByName(idParam)

		if location, err = LocateResource(route, ps); err == nil {
			location += "/" + id
			if buff, err = svc.store.Read(location); err == nil {
				w.WriteHeader(200)
				w.Write(buff) // nolint
				return
			}
		}

		svc.defaultError(w, r, ps, err)
	}
}

package rest

import (
	"encoding/json"
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
	Convert(rq *http.Request, source any) (id string, target any, err error)
}

func NewService(store storage.RWCache) *Service {
	return &Service{
		store:        store,
		resource:     map[string]ResourceAdapter{},
		defaultError: defaultErrorHandler,
	}
}

func defaultErrorHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params, err error) {
	var msg string
	w.WriteHeader(599)
	if err != nil {
		msg = err.Error()
	}
	if msg == "" {
		msg = "unexpected error"
	}
	w.Write([]byte(msg)) // nolint
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

func (svc *Service) Adapt(route string, adapter ResourceAdapter) (err error) {
	if _, ok := svc.resource[route]; ok {
		err = fmt.Errorf("%s: resource already configured", route)
	} else {
		svc.resource[route] = adapter
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

func (svc *Service) Write(route string, empty bool) httprouter.Handle {
	route, _ = strings.CutSuffix(route, "/")

	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		var location, id string
		var buff []byte
		var err error
		var in, out any

		resource := svc.resource[route]

		if location, err = LocateResource(route, ps); err == nil {

			in = resource.New()
			if err = unmarshall(r, in); err == nil {
				if id, out, err = resource.Convert(r, in); err == nil {
					location += "/" + id
					if buff, err = json.Marshal(out); err == nil {
						if err = svc.store.Write(location, buff); err == nil {
							w.WriteHeader(200)
							if !empty {
								w.Write(buff) // nolint
							}
							return
						}
					}
				}
			}
		}
	}
}

func (svc *Service) Delete(route, idParam string, empty bool) httprouter.Handle {
	route, _ = strings.CutSuffix(route, "/")

	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		var location string
		var buff []byte
		var err error

		id := ps.ByName(idParam)

		if location, err = LocateResource(route, ps); err == nil {
			location += "/" + id
			if existed := svc.store.Delete(location); existed {
				w.WriteHeader(204)
				w.Write(buff) // nolint
				return
			} else {
				err = fmt.Errorf("%s: not found", location)
			}
		}

		svc.defaultError(w, r, ps, err)
	}
}

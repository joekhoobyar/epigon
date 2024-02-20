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

// LocateResource returns a resource's location in the data store, given a data store locationTemplate and
// any params extracted from an HTTP route.  It "fills in" any parameters in the template using params,
// expecting their names to start with ":".
//
// NOTE:  The locationTemplate refers to a location in the data store, not an HTTP route.
func LocateResource(locationTemplate string, params httprouter.Params) (location string, err error) {
	locationTemplate = routerParamRegexp.ReplaceAllStringFunc(locationTemplate, func(match string) string {
		value := params.ByName(match[1:])
		if value == "" {
			err = fmt.Errorf("LocateResource: %s: no such path parameter", match)
		}
		return value
	})
	if err == nil {
		location = locationTemplate
	}
	return
}

// Adapt registers an adapter for the given data store locationTemplate.  This adapter will be used for
// creating new resources or converting resources whenever a location matches the template.
//
// NOTE:  The locationTemplate refers to a location in the data store, not an HTTP route.
func (svc *Service) Adapt(locationTemplate string, adapter ResourceAdapter) (err error) {
	if _, ok := svc.resource[locationTemplate]; ok {
		err = fmt.Errorf("%s: resource already configured", locationTemplate)
	} else {
		svc.resource[locationTemplate] = adapter
	}
	return
}

// List creates a list handler for the given data store location template.   The list
// handler will respond with a list of all resources matching the data store location.
//
// NOTE:  The locationTemplate refers to a location in the data store, not an HTTP route.
func (svc *Service) List(locationTemplate string) httprouter.Handle {
	if !strings.HasSuffix(locationTemplate, "/") {
		locationTemplate += "/"
	}

	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		var location string
		var buff []byte
		var err error

		if location, err = LocateResource(locationTemplate, ps); err == nil {
			if buff, err = svc.store.ReadList(location); err == nil {
				w.WriteHeader(200)
				w.Write(buff) // nolint
				return
			}
		}

		svc.defaultError(w, r, ps, err)
	}
}

// Get creates a handler for the given data store location template.   The handler will
// respond with the resource that matches the data store location.
//
// NOTE:  The locationTemplate refers to a location in the data store, not an HTTP route.
func (svc *Service) Get(locationTemplate, idParam string) httprouter.Handle {
	locationTemplate, _ = strings.CutSuffix(locationTemplate, "/")

	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		var location string
		var buff []byte
		var err error

		id := ps.ByName(idParam)

		if location, err = LocateResource(locationTemplate, ps); err == nil {
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

// Write creates a handler for the given data store location template.   The handler will
// write a resource to the data store at a corresponding location, after appending the
// id returned by the resource adapter's convert function.
//
// NOTE:  The locationTemplate refers to a location in the data store, not an HTTP route.
func (svc *Service) Write(locationTemplate string, empty bool) httprouter.Handle {
	locationTemplate, _ = strings.CutSuffix(locationTemplate, "/")

	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		var location, id string
		var buff []byte
		var err error
		var in, out any

		resource := svc.resource[locationTemplate]

		if location, err = LocateResource(locationTemplate, ps); err == nil {

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

// Delete creates a handler for the given data store location template.   The handler will
// delete a resource from the data store at a corresponding location, after appending the
// id read from the given idParam.
//
// NOTE:  The locationTemplate refers to a location in the data store, not an HTTP route.
func (svc *Service) Delete(locationTemplate, idParam string, empty bool) httprouter.Handle {
	locationTemplate, _ = strings.CutSuffix(locationTemplate, "/")

	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		var location string
		var buff []byte
		var err error

		id := ps.ByName(idParam)

		if location, err = LocateResource(locationTemplate, ps); err == nil {
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

package rest

import (
	"fmt"
	"github/joekhoobyar/epigon/storage"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

const (
	NONE = iota
	List
	Get
	Write
	Delete
)

type serviceRoute struct {
	method   string
	path     string
	resource string
	idParam  string
	action   int
	empty    bool
}

type ServiceBuilder struct {
	root         string
	server       *Server
	store        storage.RWCache
	routes       []serviceRoute
	resource     map[string]ResourceAdapter
	defaultError ErrHandler
}

type ResourceBuilder struct {
	service  *ServiceBuilder
	resource string
	idParam  string
	adapter  ResourceAdapter
	new      NewResourceFunc
	convert  ConvertResourceFunc
	routes   []serviceRoute
}

type NewResourceFunc func() any
type ConvertResourceFunc func(rq *http.Request, source any) (id string, target any, err error)

type resourceAdapter struct {
	new     NewResourceFunc
	convert ConvertResourceFunc
}

func (r *resourceAdapter) New() any { return r.new() }

func (r *resourceAdapter) Convert(rq *http.Request, source any) (id string, target any, err error) {
	return r.convert(rq, source)
}

func (b *ServiceBuilder) addResource(location string, adapter ResourceAdapter) (err error) {
	if _, ok := b.resource[location]; !ok {
		err = fmt.Errorf("%s: resource adapter already configured", location)
	} else {
		b.resource[location] = adapter
	}
	return
}

func (b *ServiceBuilder) AddResource(location string, adapter ResourceAdapter) error {
	return b.addResource(location, adapter)
}

func (b *ServiceBuilder) AdaptResource(location string, new NewResourceFunc, convert ConvertResourceFunc) error {
	return b.addResource(location, &resourceAdapter{new: new, convert: convert})
}

func (b *ServiceBuilder) End() *Service {
	svc := &Service{
		store:        b.store,
		resource:     b.resource,
		defaultError: b.defaultError,
	}

	for _, r := range b.routes {
		var h httprouter.Handle
		switch r.action {
		case Get:
			h = svc.Get(r.resource, r.idParam)
		case List:
			h = svc.List(r.resource)
		case Write:
			h = svc.Write(r.resource, r.empty)
		case Delete:
			h = svc.Delete(r.resource, r.idParam, r.empty)
		}
		b.server.Router.Handle(r.method, b.root+r.path, h)
	}

	return svc
}

func (b *ServiceBuilder) Resource(location, idParam string) *ResourceBuilder {
	return &ResourceBuilder{service: b, resource: location, idParam: idParam}
}

func (b *ResourceBuilder) New(new NewResourceFunc) *ResourceBuilder {
	b.new = new
	b.adapter = nil
	return b
}

func (b *ResourceBuilder) Convert(convert ConvertResourceFunc) *ResourceBuilder {
	b.convert = convert
	b.adapter = nil
	return b
}

func (b *ResourceBuilder) Adapt(adapter ResourceAdapter) *ResourceBuilder {
	b.adapter = adapter
	b.new = nil
	b.convert = nil
	return b
}

func (b *ResourceBuilder) Route(method, path, idParam string, action int, empty bool) *ResourceBuilder {
	b.routes = append(b.routes, serviceRoute{
		resource: b.resource,
		action:   action,
		method:   method,
		path:     path,
		idParam:  idParam,
		empty:    empty,
	})
	return b
}

func (b *ResourceBuilder) GET(path string, action int) *ResourceBuilder {
	return b.Route("GET", path, b.idParam, action, false)
}

func (b *ResourceBuilder) POST(path string, action int, empty bool) *ResourceBuilder {
	return b.Route("POST", path, b.idParam, action, empty)
}

func (b *ResourceBuilder) PUT(path string, action int, empty bool) *ResourceBuilder {
	return b.Route("PUT", path, b.idParam, action, empty)
}

func (b *ResourceBuilder) DELETE(path string, action int, empty bool) *ResourceBuilder {
	return b.Route("DELETE", path, b.idParam, action, empty)
}

func (b *ResourceBuilder) GETWith(path, idParam string, action int) *ResourceBuilder {
	return b.Route("GET", path, idParam, action, false)
}

func (b *ResourceBuilder) POSTWith(path, idParam string, action int, empty bool) *ResourceBuilder {
	return b.Route("POST", path, idParam, action, empty)
}

func (b *ResourceBuilder) PUTWith(path, idParam string, action int, empty bool) *ResourceBuilder {
	return b.Route("PUT", path, idParam, action, empty)
}

func (b *ResourceBuilder) DELETEWith(path, idParam string, action int, empty bool) *ResourceBuilder {
	return b.Route("DELETE", path, idParam, action, empty)
}

func (b *ResourceBuilder) End() (err error) {
	if b.adapter != nil {
		err = b.service.addResource(b.resource, b.adapter)
	} else if b.new != nil || b.convert != nil {
		if b.new == nil {
			err = fmt.Errorf("%s: missing new() handler", b.resource)
		} else if b.convert == nil {
			err = fmt.Errorf("%s: missing convert() handler", b.resource)
		} else {
			err = b.service.addResource(b.resource, &resourceAdapter{new: b.new, convert: b.convert})
		}
	}

	b.service.routes = append(b.service.routes, b.routes...)
	b.routes = nil // in case somebody tries to reuse this builder (which they shouldn't)
	return
}

package rest_test

import (
	"github.com/julienschmidt/httprouter"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github/joekhoobyar/epigon/rest"
)

var _ = Describe("Http", func() {

	Describe("Locating a resource", func() {

		Context("from a path", func() {
			path := "root/child1"
			ps := httprouter.Params{}

			It("should resolve directly", func() {
				Expect(rest.LocateResource(path, ps)).To(Equal(path))
			})
		})

		Context("from a template", func() {
			t := "root/:childId"
			t2 := "root/:childId/children/:key"
			l := "root/child1"
			l2 := "root/child1/children/arm"
			ps := httprouter.Params{
				httprouter.Param{Key: "childId", Value: "child1"},
				httprouter.Param{Key: "key", Value: "arm"},
			}

			It("should resolve a path parameter", func() {
				Expect(rest.LocateResource(t, ps)).To(Equal(l))
			})
			It("should resolve multiple path parameters", func() {
				Expect(rest.LocateResource(t2, ps)).To(Equal(l2))
			})
			It("should raise an error for an unresolvable parameter", func() {
				Expect(rest.LocateResource("root/:missingId", ps)).Error().To(MatchError(HaveSuffix(" no such path parameter")))
				Expect(rest.LocateResource("root/:childId/children/:missingId", ps)).Error().To(MatchError(HaveSuffix(" no such path parameter")))
			})
		})
	})
})

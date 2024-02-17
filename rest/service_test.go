package rest_test

import (
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/julienschmidt/httprouter"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github/joekhoobyar/epigon/rest"
	"github/joekhoobyar/epigon/storage"
	"github/joekhoobyar/epigon/test"
)

var _ = Describe("Service", func() {

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

	Describe("Serving resources", func() {
		var svc *rest.Service
		var w *httptest.ResponseRecorder
		var rq *http.Request
		var rp *http.Response
		var ps httprouter.Params
		store := storage.NewUnionedCache(test.FixtureDir())

		BeforeEach(func() {
			store.Clear()
			svc = rest.NewService(store)
			w = httptest.NewRecorder()
		})

		Context("List()", func() {
			var hndl httprouter.Handle

			It("should list resources", func() {
				hndl = svc.List("root")
				ps = httprouter.Params{}
				expected, err := store.ReadList("root/")
				Expect(err).NotTo(HaveOccurred())

				rq = httptest.NewRequest("GET", "/root", nil)
				hndl(w, rq, ps)
				rp = w.Result()
				Expect(rp.StatusCode).To(Equal(200))

				actual, err := io.ReadAll(rp.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(actual).To(Equal(expected))
			})

			It("should list nested resources", func() {
				hndl = svc.List("root/:childId/nest")
				ps = httprouter.Params{
					httprouter.Param{Key: "childId", Value: "child1"},
				}
				expected, err := store.ReadList("root/child1/nest/")
				Expect(err).NotTo(HaveOccurred())

				rq = httptest.NewRequest("GET", "/root/child1/nest", nil)
				hndl(w, rq, ps)
				rp = w.Result()
				Expect(rp.StatusCode).To(Equal(200))

				actual, err := io.ReadAll(rp.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(actual).To(Equal(expected))
			})

			It("should list empty resources", func() {
				hndl = svc.List("root/:childId/nest")
				ps = httprouter.Params{
					httprouter.Param{Key: "childId", Value: "child2"},
				}

				rq = httptest.NewRequest("GET", "/root/child2/nest", nil)
				hndl(w, rq, ps)
				rp = w.Result()
				Expect(rp.StatusCode).To(Equal(200))

				actual, err := io.ReadAll(rp.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(actual).To(Equal([]byte("[]")))
			})

			It("should error if resources do not exist", func() {
				hndl = svc.List("root/:childId/nest")
				ps = httprouter.Params{
					httprouter.Param{Key: "childId", Value: "child3"},
				}
				_, err := store.ReadList("root/child3/nest/")
				Expect(err).To(HaveOccurred())
				expected := []byte(err.Error())

				rq = httptest.NewRequest("GET", "/root/child3/nest", nil)
				hndl(w, rq, ps)
				rp = w.Result()
				Expect(rp.StatusCode).To(Equal(599))

				actual, err := io.ReadAll(rp.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(actual).To(Equal(expected))
			})

		})

		Context("Get()", func() {
			var hndl httprouter.Handle

			It("should get resources", func() {
				hndl = svc.Get("root", "id")
				ps = httprouter.Params{
					httprouter.Param{Key: "id", Value: "child1"},
				}
				expected, err := store.Read("root/child1")
				Expect(err).NotTo(HaveOccurred())

				rq = httptest.NewRequest("GET", "/root/child1", nil)
				hndl(w, rq, ps)
				rp = w.Result()
				Expect(rp.StatusCode).To(Equal(200))

				actual, err := io.ReadAll(rp.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(actual).To(Equal(expected))
			})

			It("should get nested resources", func() {
				hndl = svc.Get("root/:childId/nest", "key")
				ps = httprouter.Params{
					httprouter.Param{Key: "childId", Value: "child1"},
					httprouter.Param{Key: "key", Value: "arm"},
				}
				expected, err := store.Read("root/child1/nest/arm")
				Expect(err).NotTo(HaveOccurred())

				rq = httptest.NewRequest("GET", "/root/child1/nest/arm", nil)
				hndl(w, rq, ps)
				rp = w.Result()
				Expect(rp.StatusCode).To(Equal(200))

				actual, err := io.ReadAll(rp.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(actual).To(Equal(expected))
			})

			It("should error if resources do not exist", func() {
				hndl = svc.Get("root/:childId/nest", "key")
				ps = httprouter.Params{
					httprouter.Param{Key: "childId", Value: "child2"},
					httprouter.Param{Key: "key", Value: "leg"},
				}
				_, err := store.Read("root/child2/nest/leg")
				Expect(err).To(HaveOccurred())

				rq = httptest.NewRequest("GET", "/root/child2/nest/leg", nil)
				hndl(w, rq, ps)
				rp = w.Result()
				Expect(rp.StatusCode).To(Equal(599))

				_, err = io.ReadAll(rp.Body)
				Expect(err).NotTo(HaveOccurred())
			})

		})

	})
})

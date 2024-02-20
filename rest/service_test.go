package rest_test

import (
	"bytes"
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

type named struct {
	Name string `json:"name,omitempty"`
}
type namedAdapter struct{}

func (*namedAdapter) New() any { return &named{} }

func (*namedAdapter) Convert(rq *http.Request, source any) (id string, target any, err error) {
	id = source.(*named).Name
	target = &named{Name: id}
	return
}

type limb struct {
	Limb string `json:"limb,omitempty"`
	Side string `json:"side,omitempty"`
}
type limbAdapter struct{}

func (*limbAdapter) New() any { return &limb{} }

func (*limbAdapter) Convert(rq *http.Request, source any) (id string, target any, err error) {
	in := source.(*limb)
	id = in.Limb
	target = &limb{Limb: id, Side: in.Side}
	return
}

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
			w = httptest.NewRecorder()

			store.Clear()

			svc = rest.NewService(store)
			err := svc.Adapt("root", &namedAdapter{})
			Expect(err).NotTo(HaveOccurred())
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

		Context("Write()", func() {
			var hndl httprouter.Handle

			It("should write resources", func() {
				hndl = svc.Write("root", false)
				ps = httprouter.Params{}

				buff := []byte("{\"name\":\"child3\"}")
				body := bytes.NewReader(buff)

				rq = httptest.NewRequest("POST", "/root", body)
				hndl(w, rq, ps)
				rp = w.Result()
				Expect(rp.StatusCode).To(Equal(200))

				stored, err := store.Read("root/child3")
				Expect(err).NotTo(HaveOccurred())
				Expect(stored).To(Equal(buff))

				actual, err := io.ReadAll(rp.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(actual).To(Equal(buff))
			})

		})

		Context("Delete()", func() {
			var hndl httprouter.Handle

			It("should delete resources", func() {
				hndl = svc.Delete("root", "childId", true)
				ps = httprouter.Params{
					httprouter.Param{Key: "childId", Value: "child2"},
				}

				rq = httptest.NewRequest("DELETE", "/root/child2", nil)
				hndl(w, rq, ps)
				rp = w.Result()
				Expect(rp.StatusCode).To(Equal(204))

				_, err := store.Read("root/child2")
				Expect(err).To(HaveOccurred())
			})

		})
	})
})

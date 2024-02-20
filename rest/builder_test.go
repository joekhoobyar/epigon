package rest_test

import (
	"io"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github/joekhoobyar/epigon/rest"
	"github/joekhoobyar/epigon/storage"
	"github/joekhoobyar/epigon/test"
)

var _ = Describe("Builder", func() {

	Describe("Building resources", func() {
		var server *rest.Server
		var rp *http.Response

		store := storage.NewUnionedCache(test.FixtureDir())

		BeforeEach(func() {
			server = rest.NewServer(rest.Options{})
			sb := server.BuildService("/", store)

			err := sb.Resource("root", "childId").
				Adapt(&namedAdapter{}).
				GET("root", rest.List).
				GET("root/:childId", rest.Get).
				End()
			Expect(err).NotTo(HaveOccurred())

			err = sb.Resource("root/:childId/nest", "key").
				Adapt(&namedAdapter{}).
				GET("root/:childId/nest", rest.List).
				GET("root/:childId/nest/:key", rest.Get).
				End()

			sb.End()

			store.Clear()
		})

		Context("List()", func() {
			It("should list resources", func() {
				expected, err := store.ReadList("root/")
				Expect(err).NotTo(HaveOccurred())

				rp, err = test.GET(server, "/root")
				Expect(rp.StatusCode).To(Equal(200))

				actual, err := io.ReadAll(rp.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(actual).To(Equal(expected))
			})

			It("should list nested resources", func() {
				expected, err := store.ReadList("root/child1/nest/")
				Expect(err).NotTo(HaveOccurred())

				rp, err = test.GET(server, "/root/child1/nest")
				Expect(rp.StatusCode).To(Equal(200))

				actual, err := io.ReadAll(rp.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(actual).To(Equal(expected))
			})

			It("should list empty resources", func() {
				rp, err := test.GET(server, "/root/child2/nest")
				Expect(rp.StatusCode).To(Equal(200))

				actual, err := io.ReadAll(rp.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(actual).To(Equal([]byte("[]")))
			})

			It("should error if resources do not exist", func() {
				_, err := store.ReadList("root/child3/nest/")
				Expect(err).To(HaveOccurred())
				expected := []byte(err.Error())

				rp, err := test.GET(server, "/root/child3/nest")
				Expect(rp.StatusCode).To(Equal(599))

				actual, err := io.ReadAll(rp.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(actual).To(Equal(expected))
			})

		})

		Context("Get()", func() {
			It("should get resources", func() {
				expected, err := store.Read("root/child1")
				Expect(err).NotTo(HaveOccurred())

				rp, err := test.GET(server, "/root/child1")
				Expect(rp.StatusCode).To(Equal(200))

				actual, err := io.ReadAll(rp.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(actual).To(Equal(expected))
			})

			It("should get nested resources", func() {
				expected, err := store.Read("root/child1/nest/arm")
				Expect(err).NotTo(HaveOccurred())

				rp, err := test.GET(server, "/root/child1/nest/arm")
				Expect(rp.StatusCode).To(Equal(200))

				actual, err := io.ReadAll(rp.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(actual).To(Equal(expected))
			})

			It("should error if resources do not exist", func() {
				_, err := store.Read("root/child2/nest/leg")
				Expect(err).To(HaveOccurred())

				rp, err := test.GET(server, "/root/child2/nest/arm")
				Expect(rp.StatusCode).To(Equal(599))

				_, err = io.ReadAll(rp.Body)
				Expect(err).NotTo(HaveOccurred())
			})

		})
	})
})

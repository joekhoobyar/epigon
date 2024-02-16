package storage_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github/joekhoobyar/epigon/storage"
)

var _ = Describe("InMemoryCache", func() {
	var root, child1, child2 []byte
	var err error

	f := storage.NewInMemoryCache()
	var c storage.RWCache = f // force breakage if we fail to implement the interface

	BeforeEach(func() {
		c.Clear()
		root = []byte("{\"name\":\"root\"}")
		err = c.Write("root", root)
		Expect(err).NotTo(HaveOccurred())
		child1 = []byte("{\"name\":\"baby\"}")
		err = c.Write("root/child1", child1)
		Expect(err).NotTo(HaveOccurred())
		child2 = []byte("{\"name\":\"kid\"}")
		err = c.Write("root/child2", child2)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Reading locations", func() {

		Context("that are objects", func() {
			It("should load root", func() { Expect(f.Read("root")).To(Equal(root)) })
			It("should load child1", func() { Expect(f.Read("root/child1")).To(Equal(child1)) })
			It("should load child2", func() { Expect(f.Read("root/child2")).To(Equal(child2)) })
		})

		Context("that are lists", func() {
			It("should report an error", func() {
				_, err := f.Read("root/")
				Expect(err).To(MatchError(HaveSuffix(" location does not identify an object")))
			})
		})

		Context("that do not exist", func() {
			It("should report an error", func() {
				_, err = f.Read("missing")
				Expect(err).To(MatchError(HaveSuffix(" no such record")))
			})
		})
	})

	Describe("Reading fixture lists", func() {
		Context("that are objects", func() {
			It("should report an error", func() {
				_, err = f.ReadList("root/child1")
				Expect(err).To(MatchError(HaveSuffix(" location does not identify a collection")))
				_, err = f.ReadList("root")
				Expect(err).To(MatchError(HaveSuffix(" location does not identify a collection")))
			})
		})

		Context("that are lists", func() {
			It("should load as an array of immediate children", func() {
				Expect(f.ReadList("root/")).To(Equal([]byte("[{\"name\":\"baby\"},{\"name\":\"kid\"}]")))
			})
		})
	})

	Describe("Listing fixture keys", func() {
		Context("that are objects", func() {
			It("should report an error", func() {
				_, err = f.List("root/child1")
				Expect(err).To(MatchError(HaveSuffix(" location does not identify a collection")))
				_, err = f.List("root")
				Expect(err).To(MatchError(HaveSuffix(" location does not identify a collection")))
			})
		})

		Context("that are lists", func() {
			It("should list subkeys of immediate children", func() {
				Expect(f.List("root/")).To(Equal([]string{"root/child1", "root/child2"}))
			})
		})
	})
})

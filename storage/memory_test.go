package storage_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github/joekhoobyar/epigon/storage"
)

var _ = Describe("InMemoryCache", func() {
	var root, child1, child2 []byte
	var err error

	m := storage.NewInMemoryCache()
	var c storage.RWCache = m // force breakage if we fail to implement the interface

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
			It("should load root", func() { Expect(m.Read("root")).To(Equal(root)) })
			It("should load child1", func() { Expect(m.Read("root/child1")).To(Equal(child1)) })
			It("should load child2", func() { Expect(m.Read("root/child2")).To(Equal(child2)) })
		})

		Context("that are lists", func() {
			It("should report an error", func() {
				_, err := m.Read("root/")
				Expect(err).To(MatchError(HaveSuffix(" location does not identify an object")))
			})
		})

		Context("that do not exist", func() {
			It("should report an error", func() {
				_, err = m.Read("missing")
				Expect(err).To(MatchError(HaveSuffix(" no such object record")))
			})
		})
	})

	Describe("Reading fixture lists", func() {
		Context("that are objects", func() {
			It("should report an error", func() {
				_, err = m.ReadList("root/child1")
				Expect(err).To(MatchError(HaveSuffix(" location does not identify a collection")))
				_, err = m.ReadList("root")
				Expect(err).To(MatchError(HaveSuffix(" location does not identify a collection")))
			})
		})

		Context("that are lists", func() {
			It("should load as an array of immediate children", func() {
				Expect(m.ReadList("root/")).To(Equal([]byte("[{\"name\":\"baby\"},{\"name\":\"kid\"}]")))
			})
			It("should load empty lists", func() {
				Expect(m.ReadList("root/child2/nest/")).To(Equal([]byte("[]")))
			})
			It("should fail on missing parent records", func() {
				Expect(m.ReadList("root/child2/nester/")).Error().NotTo(HaveOccurred())
				Expect(m.ReadList("root/child3/nest/")).Error().To(HaveOccurred())
			})
		})
	})

	Describe("Listing fixture keys", func() {
		Context("that are objects", func() {
			It("should report an error", func() {
				_, err = m.List("root/child1")
				Expect(err).To(MatchError(HaveSuffix(" location does not identify a collection")))
				_, err = m.List("root")
				Expect(err).To(MatchError(HaveSuffix(" location does not identify a collection")))
			})
		})

		Context("that are lists", func() {
			It("should list subkeys of immediate children", func() {
				Expect(m.List("root/")).To(Equal([]string{"root/child1", "root/child2"}))
			})
			It("should load empty lists", func() {
				Expect(m.List("root/child2/nest/")).To(Equal([]string{}))
			})
			It("should fail on missing parent records", func() {
				Expect(m.List("root/child2/nester/")).Error().NotTo(HaveOccurred())
				Expect(m.List("root/child3/nest/")).Error().To(HaveOccurred())
			})
		})
	})

	Describe("Writing locations", func() {
		Context("that are objects", func() {
			It("should be readable", func() {
				expected := []byte("\"guy\"")
				err := m.Write("other", expected)
				Expect(err).NotTo(HaveOccurred())
				actual, err := m.Read("other")
				Expect(actual).To(Equal(expected))
			})
			It("should be listable", func() {
				err := m.Write("root/other", []byte("\"guy\""))
				Expect(err).NotTo(HaveOccurred())
				Expect(m.List("root/")).To(Equal([]string{"root/child1", "root/child2", "root/other"}))
			})
			It("should be readable as a JSON array", func() {
				err := m.Write("root/other", []byte("\"guy\""))
				Expect(err).NotTo(HaveOccurred())
				Expect(m.ReadList("root/")).To(Equal([]byte("[{\"name\":\"baby\"},{\"name\":\"kid\"},\"guy\"]")))
			})
		})

		Context("that are lists", func() {
			It("should report an error", func() {
				err = m.Write("root/", []byte("\"a\""))
				Expect(err).To(MatchError(HaveSuffix(" location does not identify an object")))
				err = m.Write("other/", []byte("\"a\""))
				Expect(err).To(MatchError(HaveSuffix(" location does not identify an object")))
			})
		})
	})

	Describe("Deleting locations", func() {
		Context("that are objects", func() {
			It("should succeed", func() {
				expected := []byte("\"guy\"")
				err := m.Write("other", expected)
				Expect(err).NotTo(HaveOccurred())
				Expect(m.Delete("other")).To(BeTrue())
			})
		})

		Context("that are missing", func() {
			It("should fail", func() {
				Expect(m.Delete("missing")).To(BeFalse())
			})
		})

		Context("that are lists", func() {
			It("should fail", func() {
				Expect(m.Delete("root/")).To(BeFalse())
			})
		})
	})
})

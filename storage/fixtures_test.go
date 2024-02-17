package storage_test

import (
	"os"
	"path"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github/joekhoobyar/epigon/storage"
	"github/joekhoobyar/epigon/test"
)

var _ = Describe("FixtureStorage", func() {
	var root, child1, child2 []byte
	var err error

	dir := test.FixtureDir()
	f := storage.NewFixtureStorage(dir)
	var c storage.RCache = f // force breakage if we fail to implement the interface

	BeforeEach(func() {
		c.Clear()

		root, err = os.ReadFile(path.Join(dir, "root.json"))
		Expect(err).NotTo(HaveOccurred())
		child1, err = os.ReadFile(path.Join(dir, "root/child1.json"))
		Expect(err).NotTo(HaveOccurred())
		child2, err = os.ReadFile(path.Join(dir, "root/child2.json"))
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Reading fixtures", func() {

		Context("that are objects", func() {
			It("should load root.json", func() { Expect(f.Read("root")).To(Equal(root)) })
			It("should load root/child1.json", func() { Expect(f.Read("root/child1")).To(Equal(child1)) })
			It("should load root/child2.json", func() { Expect(f.Read("root/child2")).To(Equal(child2)) })
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
				Expect(err).To(MatchError(HaveSuffix(" no such file or directory")))
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
			It("should load empty lists", func() {
				Expect(f.ReadList("root/child2/nest/")).To(Equal([]byte("[]")))
			})
			It("should fail on missing directories", func() {
				Expect(f.ReadList("root/child2/nester/")).Error().To(HaveOccurred())
				Expect(f.ReadList("root/child3/nest/")).Error().To(HaveOccurred())
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
			It("should load empty lists", func() {
				Expect(f.List("root/child2/nest/")).To(Equal([]string{}))
			})
			It("should fail on missing directories", func() {
				Expect(f.List("root/child2/nester/")).Error().To(HaveOccurred())
				Expect(f.List("root/child3/nest/")).Error().To(HaveOccurred())
			})
		})
	})
})
